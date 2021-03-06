package harmony

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/skwair/harmony/channel"
	"github.com/skwair/harmony/internal/endpoint"
	"github.com/skwair/harmony/invite"
	"github.com/skwair/harmony/permission"
)

// Channel represents a guild or DM channel within Discord.
type Channel struct {
	ID                   string                 `json:"id,omitempty"`
	Type                 channel.Type           `json:"type,omitempty"`
	GuildID              string                 `json:"guild_id,omitempty"`
	Position             int                    `json:"position,omitempty"` // Sorting position of the channel.
	PermissionOverwrites []permission.Overwrite `json:"permission_overwrites,omitempty"`
	Name                 string                 `json:"name,omitempty"`
	Topic                string                 `json:"topic,omitempty"`
	NSFW                 bool                   `json:"nsfw,omitempty"`
	LastMessageID        string                 `json:"last_message_id,omitempty"`

	// For voice channels.
	Bitrate          int `json:"bitrate,omitempty"`
	UserLimit        int `json:"user_limit,omitempty"`
	RateLimitPerUser int `json:"rate_limit_per_user"`

	// For DMs.
	Recipients    []User `json:"recipients,omitempty"`
	Icon          string `json:"icon,omitempty"`
	OwnerID       string `json:"owner_id,omitempty"`
	ApplicationID string `json:"application_id,omitempty"` // Application id of the group DM creator if it is bot-created.

	ParentID         string    `json:"parent_id,omitempty"` // ID of the parent category for a channel.
	LastPinTimestamp time.Time `json:"last_pin_timestamp,omitempty"`
}

// ChannelResource is a resource that allows to perform various actions on a Discord channel.
// Create one with Client.Channel.
type ChannelResource struct {
	channelID string
	client    *Client
}

// Channel returns a new channel resource to manage the channel with the given ID.
func (c *Client) Channel(id string) *ChannelResource {
	return &ChannelResource{channelID: id, client: c}
}

// Get returns the channel.
func (r *ChannelResource) Get(ctx context.Context) (*Channel, error) {
	e := endpoint.GetChannel(r.channelID)
	resp, err := r.client.doReq(ctx, e, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, apiError(resp)
	}

	var ch Channel
	if err = json.NewDecoder(resp.Body).Decode(&ch); err != nil {
		return nil, err
	}
	return &ch, nil
}

// Modify is like ModifyWithReason but with no particular reason.
func (r *ChannelResource) Modify(ctx context.Context, settings *channel.Settings) (*Channel, error) {
	return r.ModifyWithReason(ctx, settings, "")
}

// ModifyWithReason updates the channel's settings. Requires the 'MANAGE_CHANNELS'
// permission for the guild. Fires a Channel Update Gateway event. If modifying
// category, individual Channel Update events will fire for each child channel
// that also changes.
// The given reason will be set in the audit log entry for this action.
func (r *ChannelResource) ModifyWithReason(ctx context.Context, settings *channel.Settings, reason string) (*Channel, error) {
	b, err := json.Marshal(settings)
	if err != nil {
		return nil, err
	}

	e := endpoint.ModifyChannel(r.channelID)
	resp, err := r.client.doReqWithHeader(ctx, e, jsonPayload(b), reasonHeader(reason))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, apiError(resp)
	}

	var ch Channel
	if err = json.NewDecoder(resp.Body).Decode(&ch); err != nil {
		return nil, err
	}
	return &ch, nil
}

// Delete is like DeleteWithReason but with no particular reason.
func (r *ChannelResource) Delete(ctx context.Context) (*Channel, error) {
	return r.DeleteWithReason(ctx, "")
}

// DeleteWithReason deletes the channel, or closes the private message. Requires the 'MANAGE_CHANNELS'
// permission for the guild. Deleting a category does not delete its child channels; they will
// have their parent_id removed and a Channel Update Gateway event will fire for each of them.
// Returns the deleted channel on success. Fires a Channel Delete Gateway event.
// The given reason will be set in the audit log entry for this action.
func (r *ChannelResource) DeleteWithReason(ctx context.Context, reason string) (*Channel, error) {
	e := endpoint.DeleteChannel(r.channelID)
	resp, err := r.client.doReqWithHeader(ctx, e, nil, reasonHeader(reason))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, apiError(resp)
	}

	var ch Channel
	if err = json.NewDecoder(resp.Body).Decode(&ch); err != nil {
		return nil, err
	}
	return &ch, nil
}

// UpdatePermissions is like UpdatePermissionsWithReason but with no particular reason.
func (r *ChannelResource) UpdatePermissions(ctx context.Context, perms permission.Overwrite) error {
	return r.UpdatePermissionsWithReason(ctx, perms, "")
}

// UpdatePermissionsWithReason updates the channel permission overwrites for a user or
// role in the channel.
// If the channel permission overwrites do not not exist, they are created.
// Only usable for guild channels. Requires the 'MANAGE_ROLES' permission.
// The given reason will be set in the audit log entry for this action.
func (r *ChannelResource) UpdatePermissionsWithReason(ctx context.Context, perms permission.Overwrite, reason string) error {
	b, err := json.Marshal(perms)
	if err != nil {
		return err
	}

	e := endpoint.EditChannelPermissions(r.channelID, perms.ID)
	resp, err := r.client.doReqWithHeader(ctx, e, jsonPayload(b), reasonHeader(reason))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return apiError(resp)
	}
	return nil
}

// DeletePermission is like DeletePermissionWithReason but with no particular reason.
func (r *ChannelResource) DeletePermission(ctx context.Context, channelID, targetID string) error {
	return r.DeletePermissionWithReason(ctx, channelID, targetID, "")
}

// DeletePermissionWithReason deletes the channel permission overwrite for a user or
// role in a channel. Only usable for guild channels. Requires the 'MANAGE_ROLES'
// permission.
// The given reason will be set in the audit log entry for this action.
func (r *ChannelResource) DeletePermissionWithReason(ctx context.Context, channelID, targetID, reason string) error {
	e := endpoint.DeleteChannelPermission(channelID, targetID)
	resp, err := r.client.doReqWithHeader(ctx, e, nil, reasonHeader(reason))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return apiError(resp)
	}
	return nil
}

// Invites returns a list of invites (with invite metadata) for the channel.
// Only usable for guild channels. Requires the 'MANAGE_CHANNELS' permission.
func (r *ChannelResource) Invites(ctx context.Context) ([]Invite, error) {
	e := endpoint.GetChannelInvites(r.channelID)
	resp, err := r.client.doReq(ctx, e, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, apiError(resp)
	}

	var invites []Invite
	if err = json.NewDecoder(resp.Body).Decode(&invites); err != nil {
		return nil, err
	}
	return invites, nil
}

// NewInvite is like NewInviteWithReason but with no particular reason.
func (r *ChannelResource) NewInvite(ctx context.Context, settings *invite.Settings) (*Invite, error) {
	return r.NewInviteWithReason(ctx, settings, "")
}

// NewInviteWithReason creates a new invite for the channel. Only usable
// for guild channels. Requires the CREATE_INSTANT_INVITE permission.
// The given reason will be set in the audit log entry for this action.
func (r *ChannelResource) NewInviteWithReason(ctx context.Context, settings *invite.Settings, reason string) (*Invite, error) {
	b, err := json.Marshal(settings)
	if err != nil {
		return nil, err
	}

	e := endpoint.CreateChannelInvite(r.channelID)
	resp, err := r.client.doReqWithHeader(ctx, e, jsonPayload(b), reasonHeader(reason))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, apiError(resp)
	}

	var i Invite
	if err = json.NewDecoder(resp.Body).Decode(&i); err != nil {
		return nil, err
	}
	return &i, nil
}

// AddRecipient adds a recipient to the existing Group DM or to a
// DM channel, creating a new Group DM channel.
// Groups have a limit of 10 recipients, including the current user.
func (r *ChannelResource) AddRecipient(ctx context.Context, channelID, recipientID string) error {
	e := endpoint.GroupDMAddRecipient(channelID, recipientID)
	resp, err := r.client.doReq(ctx, e, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	//
	if resp.StatusCode != http.StatusNoContent {
		return apiError(resp)
	}
	return nil
}

// RemoveRecipient removes a recipient from the Group DM.
func (r *ChannelResource) RemoveRecipient(ctx context.Context, recipientID string) error {
	e := endpoint.GroupDMRemoveRecipient(r.channelID, recipientID)
	resp, err := r.client.doReq(ctx, e, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return apiError(resp)
	}
	return nil
}

// TriggerTyping triggers a typing indicator for the channel.
// Generally bots should not implement this route. However, if a bot is
// responding to a command and expects the computation to take a few
// seconds, this endpoint may be called to let the user know that the
// bot is processing their message. Fires a Typing Start Gateway event.
func (r *ChannelResource) TriggerTyping(ctx context.Context) error {
	e := endpoint.TriggerTypingIndicator(r.channelID)
	resp, err := r.client.doReq(ctx, e, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return apiError(resp)
	}
	return nil
}

// Webhooks returns webhooks for the channel.
func (r *ChannelResource) Webhooks(ctx context.Context) ([]Webhook, error) {
	e := endpoint.GetChannelWebhooks(r.channelID)
	return r.client.webhooks(ctx, e)
}

// NewWebhook is like NewWebhookWithReason but with no particular reason.
func (r *ChannelResource) NewWebhook(ctx context.Context, name, avatar string) (*Webhook, error) {
	return r.NewWebhookWithReason(ctx, name, avatar, "")
}

// NewWebhookWithReason creates a new webhook for the channel. Requires the 'MANAGE_WEBHOOKS'
// permission.
// name must contain between 2 and 32 characters. avatar is an avatar data string,
// see https://discord.com/developers/docs/resources/user#avatar-data for more info.
// It can be left empty to have the default avatar.
// The given reason will be set in the audit log entry for this action.
func (r *ChannelResource) NewWebhookWithReason(ctx context.Context, name, avatar, reason string) (*Webhook, error) {
	st := struct {
		Name   string `json:"name,omitempty"`
		Avatar string `json:"avatar,omitempty"`
	}{
		Name:   name,
		Avatar: avatar,
	}
	b, err := json.Marshal(st)
	if err != nil {
		return nil, err
	}

	e := endpoint.CreateWebhook(r.channelID)
	resp, err := r.client.doReqWithHeader(ctx, e, jsonPayload(b), reasonHeader(reason))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, apiError(resp)
	}

	var w Webhook
	if err = json.NewDecoder(resp.Body).Decode(&w); err != nil {
		return nil, err
	}
	return &w, nil
}
