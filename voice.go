package harmony

import (
	"context"
	"encoding/json"
	"net/http"
	"sync/atomic"

	"github.com/skwair/harmony/internal/endpoint"
)

// VoiceState represents the voice state of a user.
type VoiceState struct {
	GuildID   string `json:"guild_id"`
	ChannelID string `json:"channel_id"`
	UserID    string `json:"user_id"`
	SessionID string `json:"session_id"`
	Deaf      bool   `json:"deaf"`
	Mute      bool   `json:"mute"`
	SelfDeaf  bool   `json:"self_deaf"`
	SelfMute  bool   `json:"self_mute"`
	Suppress  bool   `json:"suppress"` // Whether this user is muted by the current user.
}

// VoiceRegion represents a voice region a guild can use or is using for its voice channels.
type VoiceRegion struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	// Whether this is a vip-only server.
	VIP bool `json:"vip,omitempty"`
	// Whether this is a single server that is closest to the current user's client.
	Optimal bool `json:"optimal,omitempty"`
	// Whether this is a deprecated voice region (avoid switching to these.
	Deprecated bool `json:"deprecated,omitempty"`
	// Whether this is a custom voice region (used for events/etc).
	Custom bool `json:"custom,omitempty"`
}

// VoiceRegions returns a list of available voice regions that can be used when creating
// or updating servers.
func (c *Client) VoiceRegions(ctx context.Context, guildID string) ([]VoiceRegion, error) {
	e := endpoint.GetVoiceRegions()
	resp, err := c.doReq(ctx, e, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, apiError(resp)
	}

	var regions []VoiceRegion
	if err = json.NewDecoder(resp.Body).Decode(&regions); err != nil {
		return nil, err
	}
	return regions, nil
}

// Speaking sends an Opcode 5 Speaking payload. This does nothing
// if the user is already in the given state.
func (vc *VoiceConnection) Speaking(s bool, delay int) error {
	// Return early if the user is already in the asked state.
	prev := atomic.LoadInt32(&vc.speaking)
	if (prev == 1) == s {
		return nil
	}

	if s {
		atomic.StoreInt32(&vc.speaking, 1)
	} else {
		atomic.StoreInt32(&vc.speaking, 0)
	}

	p := struct {
		Speaking bool   `json:"speaking"`
		Delay    int    `json:"delay"`
		SSRC     uint32 `json:"ssrc"`
	}{
		Speaking: s,
		Delay:    delay,
		SSRC:     vc.ssrc,
	}

	if err := vc.sendPayload(voiceOpcodeSpeaking, p); err != nil {
		// If there is an error, reset our internal value to its previous
		// state because the update was not acknowledged by Discord.
		atomic.StoreInt32(&vc.speaking, prev)
		return err
	}

	return nil
}
