package audit

import (
	"strconv"
)

func channelCreateFromEntry(e *rawEntry) (*ChannelCreate, error) {
	chCreate := &ChannelCreate{
		BaseEntry: baseEntryFromRaw(e),
	}

	var err error
	for _, ch := range e.Changes {
		switch changeKey(ch.Key) {
		case changeKeyName:
			chCreate.Name, err = stringValue(ch.New)
			if err != nil {
				return nil, err
			}

		case changeKeyType:
			chCreate.Type, err = intValue(ch.New)
			if err != nil {
				return nil, err
			}

		case changeKeyRateLimitPerUser:
			chCreate.RateLimitPerUser, err = intValue(ch.New)
			if err != nil {
				return nil, err
			}

		case changeKeyNFSW:
			chCreate.NSFW, err = boolValue(ch.New)
			if err != nil {
				return nil, err
			}

		case changeKeyPermissionOverwrites:
			chCreate.PermissionOverwrites, err = permissionOverwritesValue(ch.New)
			if err != nil {
				return nil, err
			}
		}
	}

	return chCreate, nil
}

func channelUpdateFromEntry(e *rawEntry) (*ChannelUpdate, error) {
	chUpdate := &ChannelUpdate{
		BaseEntry: baseEntryFromRaw(e),
	}

	for _, ch := range e.Changes {
		switch changeKey(ch.Key) {
		case changeKeyName:
			oldValue, newValue, err := stringValues(ch.Old, ch.New)
			if err != nil {
				return nil, err
			}
			chUpdate.Name = &StringValues{Old: oldValue, New: newValue}

		case changeKeyTopic:
			oldValue, newValue, err := stringValues(ch.Old, ch.New)
			if err != nil {
				return nil, err
			}
			chUpdate.Topic = &StringValues{Old: oldValue, New: newValue}

		case changeKeyBitrate:
			oldValue, newValue, err := intValues(ch.Old, ch.New)
			if err != nil {
				return nil, err
			}
			chUpdate.Bitrate = &IntValues{Old: oldValue, New: newValue}

		case changeKeyRateLimitPerUser:
			oldValue, newValue, err := intValues(ch.Old, ch.New)
			if err != nil {
				return nil, err
			}
			chUpdate.RateLimitPerUser = &IntValues{Old: oldValue, New: newValue}

		case changeKeyNFSW:
			oldValue, newValue, err := boolValues(ch.Old, ch.New)
			if err != nil {
				return nil, err
			}
			chUpdate.NSFW = &BoolValues{Old: oldValue, New: newValue}

		case changeKeyApplicationID:
			oldValue, newValue, err := stringValues(ch.Old, ch.New)
			if err != nil {
				return nil, err
			}
			chUpdate.ApplicationID = &StringValues{Old: oldValue, New: newValue}

		case changeKeyPosition:
			oldValue, newValue, err := intValues(ch.Old, ch.New)
			if err != nil {
				return nil, err
			}
			chUpdate.Position = &IntValues{Old: oldValue, New: newValue}
		}
	}

	return chUpdate, nil
}

func channelDeleteFromEntry(e *rawEntry) (*ChannelDelete, error) {
	chDelete := &ChannelDelete{
		BaseEntry: baseEntryFromRaw(e),
	}

	var err error
	for _, ch := range e.Changes {
		switch changeKey(ch.Key) {
		case changeKeyName:
			chDelete.Name, err = stringValue(ch.Old)
			if err != nil {
				return nil, err
			}

		case changeKeyType:
			chDelete.Type, err = intValue(ch.Old)
			if err != nil {
				return nil, err
			}

		case changeKeyRateLimitPerUser:
			chDelete.RateLimitPerUser, err = intValue(ch.Old)
			if err != nil {
				return nil, err
			}

		case changeKeyNFSW:
			chDelete.NSFW, err = boolValue(ch.Old)
			if err != nil {
				return nil, err
			}

		case changeKeyPermissionOverwrites:
			chDelete.PermissionOverwrites, err = permissionOverwritesValue(ch.Old)
			if err != nil {
				return nil, err
			}
		}
	}

	return chDelete, nil
}

func channelOverwriteCreateFromEntry(e *rawEntry) (*ChannelOverwriteCreate, error) {
	overwriteCreate := &ChannelOverwriteCreate{
		BaseEntry: baseEntryFromRaw(e),
		RoleName:  e.Options.RoleName,
	}

	var err error
	for _, ch := range e.Changes {
		switch changeKey(ch.Key) {
		case changeKeyID:
			overwriteCreate.ID, err = stringValue(ch.New)
			if err != nil {
				return nil, err
			}

		case changeKeyType:
			overwriteCreate.Type, err = stringValue(ch.New)
			if err != nil {
				return nil, err
			}

		case changeKeyAllow:
			overwriteCreate.Allow, err = intValue(ch.New)
			if err != nil {
				return nil, err
			}

		case changeKeyDeny:
			overwriteCreate.Deny, err = intValue(ch.New)
			if err != nil {
				return nil, err
			}
		}
	}

	return overwriteCreate, nil
}

func channelOverwriteUpdateFromEntry(e *rawEntry) (*ChannelOverwriteUpdate, error) {
	overwriteUpdate := &ChannelOverwriteUpdate{
		BaseEntry: baseEntryFromRaw(e),
		Type:      e.Options.Type,
		ID:        e.Options.ID,
		RoleName:  e.Options.RoleName,
	}

	for _, ch := range e.Changes {
		switch changeKey(ch.Key) {
		case changeKeyAllow:
			oldValue, newValue, err := intValues(ch.Old, ch.New)
			if err != nil {
				return nil, err
			}
			overwriteUpdate.Allow = &IntValues{Old: oldValue, New: newValue}

		case changeKeyDeny:
			oldValue, newValue, err := intValues(ch.Old, ch.New)
			if err != nil {
				return nil, err
			}
			overwriteUpdate.Deny = &IntValues{Old: oldValue, New: newValue}
		}
	}

	return overwriteUpdate, nil
}

func channelOverwriteDeleteFromEntry(e *rawEntry) (*ChannelOverwriteDelete, error) {
	overwriteDelete := &ChannelOverwriteDelete{
		BaseEntry: baseEntryFromRaw(e),
		RoleName:  e.Options.RoleName,
	}

	var err error
	for _, ch := range e.Changes {
		switch changeKey(ch.Key) {
		case changeKeyID:
			overwriteDelete.ID, err = stringValue(ch.Old)
			if err != nil {
				return nil, err
			}

		case changeKeyType:
			overwriteDelete.Type, err = stringValue(ch.Old)
			if err != nil {
				return nil, err
			}

		case changeKeyAllow:
			overwriteDelete.Allow, err = intValue(ch.Old)
			if err != nil {
				return nil, err
			}

		case changeKeyDeny:
			overwriteDelete.Deny, err = intValue(ch.Old)
			if err != nil {
				return nil, err
			}
		}
	}

	return overwriteDelete, nil
}

func messageDeleteFromEntry(e *rawEntry) (*MessageDelete, error) {
	msgDelete := &MessageDelete{
		BaseEntry: baseEntryFromRaw(e),
		ChannelID: e.Options.ChannelID,
	}

	var err error
	msgDelete.Count, err = strconv.Atoi(e.Options.Count)
	if err != nil {
		return nil, err
	}

	return msgDelete, nil
}
