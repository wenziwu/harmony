package discord

import (
	"time"
)

// ChannelType describes the type of a channel. Different fields
// are set or not depending on the channel's type.
type ChannelType int

// Supported channel types:
const (
	ChannelTypeGuildText ChannelType = iota
	ChannelTypeDM
	ChannelTypeGuildVoice
	ChannelTypeGroupDM
	ChannelTypeGuildCategory
	ChannelTypeGuildNews
	ChannelTypeGuildStore
)

// Channel represents a guild or DM channel within Discord.
type Channel struct {
	ID                   string                `json:"id"`
	Type                 ChannelType           `json:"type"`
	GuildID              string                `json:"guild_id"`
	Position             int                   `json:"position"` // Sorting position of the channel.
	PermissionOverwrites []PermissionOverwrite `json:"permission_overwrites"`
	Name                 string                `json:"name"`
	Topic                string                `json:"topic"`
	NSFW                 bool                  `json:"nsfw"`
	LastMessageID        string                `json:"last_message_id"`

	// For voice channels.
	Bitrate          int `json:"bitrate"`
	UserLimit        int `json:"user_limit"`
	RateLimitPerUser int `json:"rate_limit_per_user"`

	// For DMs.
	Recipients    []User `json:"recipients"`
	Icon          string `json:"icon"`
	OwnerID       string `json:"owner_id"`
	ApplicationID string `json:"application_id"` // Application id of the group DM creator if it is bot-created.

	ParentID         string    `json:"parent_id"` // ID of the parent category for a channel.
	LastPinTimestamp time.Time `json:"last_pin_timestamp"`
}

// ChannelMention represents a channel mention.
type ChannelMention struct {
	ID      string      `json:"id"`
	GuildID string      `json:"guild_id"`
	Type    ChannelType `json:"type"`
	Name    string      `json:"name"`
}
