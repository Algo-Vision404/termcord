package cache

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/termcord/termcord/internal/model"
	_ "modernc.org/sqlite"
)

type Store struct {
	db   *sql.DB
	enc  *encrypter
	encOn bool
}

type Options struct {
	Path     string
	Encrypt  bool
	Key      []byte
}

func Open(path string) (*Store, error) {
	return OpenWith(Options{Path: path})
}

func OpenWith(opts Options) (*Store, error) {
	path := opts.Path
	if err := secureDir(parentDir(path)); err != nil {
		return nil, fmt.Errorf("create cache dir: %w", err)
	}
	db, err := sql.Open("sqlite", path+"?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	s := &Store{db: db, encOn: opts.Encrypt}
	if opts.Encrypt {
		e, err := newEncrypter(opts.Key)
		if err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("cache encryption: %w", err)
		}
		s.enc = e
	}
	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := secureFile(path); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("secure cache file: %w", err)
	}
	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) migrate() error {
	schema := `
CREATE TABLE IF NOT EXISTS messages (
	id TEXT NOT NULL,
	channel_id TEXT NOT NULL,
	author TEXT NOT NULL,
	author_id TEXT NOT NULL,
	content TEXT NOT NULL,
	timestamp INTEGER NOT NULL,
	edited INTEGER NOT NULL DEFAULT 0,
	reply_to_id TEXT NOT NULL DEFAULT '',
	reactions TEXT NOT NULL DEFAULT '[]',
	PRIMARY KEY (id, channel_id)
);
CREATE INDEX IF NOT EXISTS idx_messages_channel_ts ON messages(channel_id, timestamp DESC);
`
	if _, err := s.db.Exec(schema); err != nil {
		return err
	}
	_, _ = s.db.Exec(`ALTER TABLE messages ADD COLUMN reactions TEXT NOT NULL DEFAULT '[]'`)
	if !s.encOn {
		_, _ = s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_messages_content ON messages(content)`)
	}
	return nil
}

func (s *Store) sealContent(content string) (string, error) {
	if !s.encOn || s.enc == nil {
		return content, nil
	}
	return s.enc.encrypt(content)
}

func (s *Store) openContent(stored string) (string, error) {
	if !s.encOn || s.enc == nil {
		return stored, nil
	}
	return s.enc.decrypt(stored)
}

func (s *Store) UpsertMessage(ctx context.Context, msg model.Message) error {
	content, err := s.sealContent(msg.Content)
	if err != nil {
		return err
	}
	reactions, err := json.Marshal(msg.Reactions)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `
INSERT INTO messages (id, channel_id, author, author_id, content, timestamp, edited, reply_to_id, reactions)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(id, channel_id) DO UPDATE SET
	content=excluded.content,
	edited=excluded.edited,
	timestamp=excluded.timestamp,
	reactions=excluded.reactions
`, msg.ID, msg.ChannelID, msg.Author, msg.AuthorID, content, msg.Timestamp.Unix(), boolToInt(msg.Edited), msg.ReplyToID, string(reactions))
	return err
}

func (s *Store) ListMessages(ctx context.Context, channelID string, limit int, beforeID string) ([]model.Message, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `
SELECT id, channel_id, author, author_id, content, timestamp, edited, reply_to_id, reactions
FROM messages
WHERE channel_id = ?
`
	args := []any{channelID}
	if beforeID != "" {
		query += ` AND timestamp < (SELECT timestamp FROM messages WHERE id = ? AND channel_id = ? LIMIT 1)`
		args = append(args, beforeID, channelID)
	}
	query += ` ORDER BY timestamp DESC LIMIT ?`
	args = append(args, limit)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.Message
	for rows.Next() {
		msg, err := s.scanMessage(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, msg)
	}
	reverse(out)
	return out, rows.Err()
}

func (s *Store) Search(ctx context.Context, query string, limit int) ([]model.Message, error) {
	if s.encOn {
		return s.searchDecrypt(ctx, query, limit)
	}
	if limit <= 0 {
		limit = 50
	}
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, nil
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT id, channel_id, author, author_id, content, timestamp, edited, reply_to_id, reactions
FROM messages
WHERE content LIKE ?
ORDER BY timestamp DESC
LIMIT ?
`, "%"+query+"%", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.Message
	for rows.Next() {
		msg, err := s.scanMessage(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, msg)
	}
	return out, rows.Err()
}

func (s *Store) searchDecrypt(ctx context.Context, query string, limit int) ([]model.Message, error) {
	query = strings.TrimSpace(strings.ToLower(query))
	if query == "" {
		return nil, nil
	}
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT id, channel_id, author, author_id, content, timestamp, edited, reply_to_id, reactions
FROM messages
ORDER BY timestamp DESC
LIMIT 500
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.Message
	for rows.Next() {
		msg, err := s.scanMessage(rows)
		if err != nil {
			return nil, err
		}
		if strings.Contains(strings.ToLower(msg.Content), query) {
			out = append(out, msg)
			if len(out) >= limit {
				break
			}
		}
	}
	return out, rows.Err()
}

func (s *Store) Purge(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM messages`)
	return err
}

func (s *Store) scanMessage(rows *sql.Rows) (model.Message, error) {
	var msg model.Message
	var ts int64
	var edited int
	var reactionsJSON string
	var storedContent string
	if err := rows.Scan(&msg.ID, &msg.ChannelID, &msg.Author, &msg.AuthorID, &storedContent, &ts, &edited, &msg.ReplyToID, &reactionsJSON); err != nil {
		return msg, err
	}
	content, err := s.openContent(storedContent)
	if err != nil {
		return msg, err
	}
	msg.Content = content
	msg.Timestamp = time.Unix(ts, 0).UTC()
	msg.Edited = edited == 1
	if reactionsJSON != "" {
		_ = json.Unmarshal([]byte(reactionsJSON), &msg.Reactions)
	}
	return msg, nil
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func reverse[T any](s []T) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}
