package gateway

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gorilla/websocket"
)

// userGateway connects with a Discord *user* identify payload (no bot intents).
type userGateway struct {
	token   string
	session *discordgo.Session

	conn *websocket.Conn
	once sync.Once

	sequence   int64
	sessionID  string
	resumeURL  string
	heartbeat  time.Duration
	stopBeat   chan struct{}
	beatDone   chan struct{}
	failCh     chan error

	onReady   func()
	onMessage func(*discordgo.Message)
	onError   func(error)

	readyHandler func(*discordgo.Session, *discordgo.Ready)
	handlers     map[string]func(*discordgo.Session, any)
}

func newUserGateway(token string, session *discordgo.Session, c *Client) *userGateway {
	g := &userGateway{
		token:   token,
		session: session,
		onReady: c.onReady,
		onMessage: func(m *discordgo.Message) {
			if c.onMessage != nil {
				c.onMessage(c.mapMessage(m))
			}
		},
		onError:  c.onError,
		stopBeat: make(chan struct{}),
		beatDone: make(chan struct{}),
		failCh:   make(chan error, 1),
	}
	g.readyHandler = c.handleReady
	g.handlers = map[string]func(*discordgo.Session, any){
		"MESSAGE_CREATE": func(s *discordgo.Session, i any) {
			ev := i.(*discordgo.MessageCreate)
			g.onMessage(ev.Message)
		},
		"MESSAGE_UPDATE": func(s *discordgo.Session, i any) {
			ev := i.(*discordgo.MessageUpdate)
			if c.onMessage != nil {
				msg := c.mapMessage(ev.Message)
				msg.Edited = true
				c.onMessage(msg)
			}
		},
		"THREAD_CREATE": func(s *discordgo.Session, i any) {
			ev := i.(*discordgo.ThreadCreate)
			c.handleThreadCreate(s, ev)
		},
		"MESSAGE_REACTION_ADD": func(s *discordgo.Session, i any) {
			ev := i.(*discordgo.MessageReactionAdd)
			c.handleReactionAdd(s, ev)
		},
		"TYPING_START": func(s *discordgo.Session, i any) {
			ev := i.(*discordgo.TypingStart)
			c.resolveTypingUser(ev.UserID)
			c.typing.handle(ev)
		},
	}
	return g
}

func (g *userGateway) open() error {
	gw, err := g.session.Gateway()
	if err != nil {
		return err
	}
	url := gw + "?v=10&encoding=json"
	header := http.Header{}
	header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) discord/1.0.9204 Chrome/134.0.0.0 Safari/537.36")

	conn, _, err := websocket.DefaultDialer.Dial(url, header)
	if err != nil {
		return err
	}
	g.conn = conn
	go g.readLoop()
	return nil
}

func (g *userGateway) close() {
	g.once.Do(func() {
		close(g.stopBeat)
		if g.conn != nil {
			_ = g.conn.Close()
		}
	})
}

func (g *userGateway) fail(err error) {
	select {
	case g.failCh <- err:
	default:
	}
}

func (g *userGateway) readLoop() {
	defer g.close()
	for {
		_, data, err := g.conn.ReadMessage()
		if err != nil {
			var closeErr *websocket.CloseError
			if errors.As(err, &closeErr) {
				text := closeErr.Text
				if text == "" {
					text = "Unknown error"
				}
				g.fail(fmt.Errorf("websocket: close %d: %s", closeErr.Code, text))
			} else if g.onError != nil {
				g.onError(fmt.Errorf("gateway read: %w", err))
			}
			return
		}
		var payload gatewayPayload
		if err := json.Unmarshal(data, &payload); err != nil {
			continue
		}
		if payload.S > 0 {
			g.sequence = payload.S
		}
		switch payload.Op {
		case 10: // Hello
			var hello struct {
				HeartbeatInterval float64 `json:"heartbeat_interval"`
			}
			_ = json.Unmarshal(payload.D, &hello)
			g.heartbeat = time.Duration(hello.HeartbeatInterval) * time.Millisecond
			go g.heartbeatLoop()
			if err := g.identify(); err != nil {
				g.fail(fmt.Errorf("identify: %w", err))
				return
			}
		case 11: // Heartbeat ACK
		case 0: // Dispatch
			g.dispatch(payload.T, payload.D)
		case 7: // Reconnect
			if g.onError != nil {
				g.onError(fmt.Errorf("gateway requested reconnect"))
			}
			return
		case 9: // Invalid session
			var resumable bool
			_ = json.Unmarshal(payload.D, &resumable)
			if !resumable {
				g.fail(fmt.Errorf("invalid session (not resumable)"))
				return
			}
			time.Sleep(time.Second)
			if err := g.identify(); err != nil {
				g.fail(fmt.Errorf("identify: %w", err))
				return
			}
		}
	}
}

func (g *userGateway) heartbeatLoop() {
	ticker := time.NewTicker(g.heartbeat)
	defer ticker.Stop()
	defer close(g.beatDone)
	for {
		select {
		case <-g.stopBeat:
			return
		case <-ticker.C:
			seq := g.sequence
			_ = g.conn.WriteJSON(map[string]interface{}{"op": 1, "d": seq})
		}
	}
}

func (g *userGateway) identify() error {
	payload := map[string]interface{}{
		"op": 2,
		"d": map[string]interface{}{
			"token":        g.token,
			"capabilities": userCapabilities,
			"properties": map[string]interface{}{
				"os":                       "Windows",
				"browser":                  "Discord Client",
				"device":                   "",
				"system_locale":            "en-US",
				"browser_user_agent":       "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) discord/1.0.9204 Chrome/134.0.0.0 Safari/537.36",
				"browser_version":          "134.0.0.0",
				"os_version":               "10",
				"referrer":                 "",
				"referring_domain":         "",
				"referrer_current":         "",
				"referring_domain_current": "",
				"release_channel":          "stable",
				"client_build_number":      clientBuildNumber,
				"client_event_source":      nil,
			},
			"compress":        false,
			"large_threshold": 250,
			"presence": map[string]interface{}{
				"status":     "online",
				"since":      0,
				"activities": []interface{}{},
				"afk":        false,
			},
			"client_state": map[string]interface{}{
				"guild_versions":              map[string]interface{}{},
				"highest_last_message_id":     "0",
				"read_state_version":          0,
				"user_guild_settings_version": -1,
				"user_settings_version":       -1,
				"private_channels_version":    "0",
				"api_code_version":            0,
			},
		},
	}
	return g.conn.WriteJSON(payload)
}

func (g *userGateway) dispatch(event string, raw json.RawMessage) {
	switch event {
	case "READY":
		var ready discordgo.Ready
		if err := json.Unmarshal(raw, &ready); err != nil {
			g.fail(fmt.Errorf("ready decode: %w", err))
			return
		}
		g.sessionID = ready.SessionID
		if g.readyHandler != nil {
			g.readyHandler(g.session, &ready)
		}
	default:
		if fn, ok := g.handlers[event]; ok {
			ev := newEvent(event, raw)
			if ev != nil {
				fn(g.session, ev)
			}
		}
	}
}

type gatewayPayload struct {
	Op int             `json:"op"`
	D  json.RawMessage `json:"d"`
	S  int64           `json:"s"`
	T  string          `json:"t"`
}

func newEvent(name string, raw json.RawMessage) any {
	switch name {
	case "MESSAGE_CREATE":
		ev := &discordgo.MessageCreate{}
		if json.Unmarshal(raw, ev) == nil {
			return ev
		}
	case "MESSAGE_UPDATE":
		ev := &discordgo.MessageUpdate{}
		if json.Unmarshal(raw, ev) == nil {
			return ev
		}
	case "THREAD_CREATE":
		ev := &discordgo.ThreadCreate{}
		if json.Unmarshal(raw, ev) == nil {
			return ev
		}
	case "MESSAGE_REACTION_ADD":
		ev := &discordgo.MessageReactionAdd{}
		if json.Unmarshal(raw, ev) == nil {
			return ev
		}
	case "MESSAGE_REACTION_REMOVE":
		ev := &discordgo.MessageReactionRemove{}
		if json.Unmarshal(raw, ev) == nil {
			return ev
		}
	case "TYPING_START":
		ev := &discordgo.TypingStart{}
		if json.Unmarshal(raw, ev) == nil {
			return ev
		}
	}
	return nil
}
