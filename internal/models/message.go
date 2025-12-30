package models

import (
	"database/sql"
	"encoding/json"
	"time"
)

type Message struct {
	ID            int64          `json:"id" db:"id"`
	MessageID     int            `json:"message_id" db:"message_id"`
	FromID        int64          `json:"from_id" db:"from_id"`
	PeerID        int64          `json:"peer_id" db:"peer_id"`
	PeerType      string         `json:"peer_type" db:"peer_type"`
	Message       sql.NullString `json:"message" db:"message"`
	Date          int            `json:"date" db:"date"`
	RandomID      sql.NullInt64  `json:"random_id" db:"random_id"`
	ReplyToMsgID  sql.NullInt32  `json:"reply_to_msg_id" db:"reply_to_msg_id"`
	FwdFromID     sql.NullInt64  `json:"fwd_from_id" db:"fwd_from_id"`
	FwdDate       sql.NullInt32  `json:"fwd_date" db:"fwd_date"`
	ViaBotID      sql.NullInt64  `json:"via_bot_id" db:"via_bot_id"`
	EditDate      sql.NullInt32  `json:"edit_date" db:"edit_date"`
	MediaType     sql.NullString `json:"media_type" db:"media_type"`
	MediaID       sql.NullInt64  `json:"media_id" db:"media_id"`
	Entities      json.RawMessage `json:"entities" db:"entities"`
	IsOut         bool           `json:"is_out" db:"is_out"`
	IsMentioned   bool           `json:"is_mentioned" db:"is_mentioned"`
	IsMediaUnread bool           `json:"is_media_unread" db:"is_media_unread"`
	IsSilent      bool           `json:"is_silent" db:"is_silent"`
	IsPost        bool           `json:"is_post" db:"is_post"`
	IsPinned      bool           `json:"is_pinned" db:"is_pinned"`
	IsNoforwards  bool           `json:"is_noforwards" db:"is_noforwards"`
	Views         sql.NullInt32  `json:"views" db:"views"`
	Forwards      sql.NullInt32  `json:"forwards" db:"forwards"`
	GroupedID     sql.NullInt64  `json:"grouped_id" db:"grouped_id"`
	CreatedAt     time.Time      `json:"created_at" db:"created_at"`
}

// ToTLMessage converts to TL Message format for API responses
func (m *Message) ToTLMessage() map[string]interface{} {
	result := map[string]interface{}{
		"_":       "message",
		"id":      m.MessageID,
		"from_id": map[string]interface{}{
			"_":       "peerUser",
			"user_id": m.FromID,
		},
		"date": m.Date,
		"out":  m.IsOut,
	}

	// Peer ID
	switch m.PeerType {
	case "user":
		result["peer_id"] = map[string]interface{}{
			"_":       "peerUser",
			"user_id": m.PeerID,
		}
	case "chat":
		result["peer_id"] = map[string]interface{}{
			"_":       "peerChat",
			"chat_id": m.PeerID,
		}
	case "channel":
		result["peer_id"] = map[string]interface{}{
			"_":          "peerChannel",
			"channel_id": m.PeerID,
		}
	}

	// Message content
	if m.Message.Valid {
		result["message"] = m.Message.String
	} else {
		result["message"] = ""
	}

	// Reply
	if m.ReplyToMsgID.Valid {
		result["reply_to"] = map[string]interface{}{
			"_":              "messageReplyHeader",
			"reply_to_msg_id": m.ReplyToMsgID.Int32,
		}
	}

	// Forward
	if m.FwdFromID.Valid {
		result["fwd_from"] = map[string]interface{}{
			"_":       "messageFwdHeader",
			"from_id": map[string]interface{}{
				"_":       "peerUser",
				"user_id": m.FwdFromID.Int64,
			},
			"date": m.FwdDate.Int32,
		}
	}

	// Edit date
	if m.EditDate.Valid {
		result["edit_date"] = m.EditDate.Int32
	}

	// Entities
	if len(m.Entities) > 0 {
		var entities []interface{}
		json.Unmarshal(m.Entities, &entities)
		result["entities"] = entities
	}

	// Flags
	if m.IsMentioned {
		result["mentioned"] = true
	}
	if m.IsMediaUnread {
		result["media_unread"] = true
	}
	if m.IsSilent {
		result["silent"] = true
	}
	if m.IsPost {
		result["post"] = true
	}
	if m.IsPinned {
		result["pinned"] = true
	}
	if m.IsNoforwards {
		result["noforwards"] = true
	}

	// Views and forwards
	if m.Views.Valid {
		result["views"] = m.Views.Int32
	}
	if m.Forwards.Valid {
		result["forwards"] = m.Forwards.Int32
	}

	return result
}

type Dialog struct {
	ID                   int64          `json:"id" db:"id"`
	UserID               int64          `json:"user_id" db:"user_id"`
	PeerID               int64          `json:"peer_id" db:"peer_id"`
	PeerType             string         `json:"peer_type" db:"peer_type"`
	TopMessage           int            `json:"top_message" db:"top_message"`
	ReadInboxMaxID       int            `json:"read_inbox_max_id" db:"read_inbox_max_id"`
	ReadOutboxMaxID      int            `json:"read_outbox_max_id" db:"read_outbox_max_id"`
	UnreadCount          int            `json:"unread_count" db:"unread_count"`
	UnreadMentionsCount  int            `json:"unread_mentions_count" db:"unread_mentions_count"`
	UnreadReactionsCount int            `json:"unread_reactions_count" db:"unread_reactions_count"`
	Pts                  int            `json:"pts" db:"pts"`
	Draft                sql.NullString `json:"draft" db:"draft"`
	FolderID             int            `json:"folder_id" db:"folder_id"`
	IsPinned             bool           `json:"is_pinned" db:"is_pinned"`
	PinnedOrder          int            `json:"pinned_order" db:"pinned_order"`
	MuteUntil            int            `json:"mute_until" db:"mute_until"`
	NotifySound          string         `json:"notify_sound" db:"notify_sound"`
	ShowPreviews         bool           `json:"show_previews" db:"show_previews"`
	UpdatedAt            time.Time      `json:"updated_at" db:"updated_at"`
}

// ToTLDialog converts to TL Dialog format for API responses
func (d *Dialog) ToTLDialog() map[string]interface{} {
	result := map[string]interface{}{
		"_":                      "dialog",
		"top_message":            d.TopMessage,
		"read_inbox_max_id":      d.ReadInboxMaxID,
		"read_outbox_max_id":     d.ReadOutboxMaxID,
		"unread_count":           d.UnreadCount,
		"unread_mentions_count":  d.UnreadMentionsCount,
		"unread_reactions_count": d.UnreadReactionsCount,
	}

	// Peer
	switch d.PeerType {
	case "user":
		result["peer"] = map[string]interface{}{
			"_":       "peerUser",
			"user_id": d.PeerID,
		}
	case "chat":
		result["peer"] = map[string]interface{}{
			"_":       "peerChat",
			"chat_id": d.PeerID,
		}
	case "channel":
		result["peer"] = map[string]interface{}{
			"_":          "peerChannel",
			"channel_id": d.PeerID,
		}
	}

	// Notify settings
	result["notify_settings"] = map[string]interface{}{
		"_":             "peerNotifySettings",
		"mute_until":    d.MuteUntil,
		"sound":         d.NotifySound,
		"show_previews": d.ShowPreviews,
	}

	// Flags
	if d.IsPinned {
		result["pinned"] = true
	}
	if d.FolderID > 0 {
		result["folder_id"] = d.FolderID
	}

	return result
}
