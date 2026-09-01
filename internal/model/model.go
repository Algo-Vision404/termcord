package model

import "time"

type ChannelKind int

const (
	ChannelGuildText ChannelKind = iota
	ChannelDM
	ChannelGroupDM
	ChannelThread
)

type Channel struct {
	ID        string
	Name      string
	GuildID   string
	GuildName string
	Kind      ChannelKind
	Unread    int
	Mention   bool
	ParentID  string
}

type Guild struct {
	ID   string
	Name string
}

type ChannelSection struct {
	Title    string
	GuildID  string
	Channels []Channel
}

type Attachment struct {
	ID          string
	URL         string
	Filename    string
	ContentType string
	Width       int
	Height      int
	Size        int
}

type Message struct {
	ID              string
	ChannelID       string
	Author          string
	AuthorID        string
	Content         string
	Timestamp       time.Time
	Edited          bool
	ReplyToID       string
	ReplyToAuthor   string
	ReplyToContent  string
	MentionEveryone bool
	MentionsMe      bool
	Embeds          []Embed
	Attachments     []Attachment
	Reactions       []Reaction
}

type Embed struct {
	Title       string
	Description string
	URL         string
	Author      string
	Fields      []EmbedField
}

type EmbedField struct {
	Name  string
	Value string
}

type Reaction struct {
	Emoji string
	Count int
	Me    bool
}

type User struct {
	ID       string
	Username string
	Nickname string
}
