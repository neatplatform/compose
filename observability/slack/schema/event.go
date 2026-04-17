package schema

import "encoding/json"

type EventEnvelope struct {
	Type               string          `json:"type"`                // url_verification, event_callback
	Challenge          string          `json:"challenge,omitempty"` // Only present for url_verification events
	TeamID             string          `json:"team_id"`
	APIAppID           string          `json:"api_app_id"`
	Authorizations     []Authorization `json:"authorizations"`
	EventContext       string          `json:"event_context"`
	EventID            string          `json:"event_id"`
	EventTime          int64           `json:"event_time"`
	Event              json.RawMessage `json:"event"` // The actual event payload. The exact schema varies by event type.
	IsExtSharedChannel bool            `json:"is_ext_shared_channel"`
}

type Authorization struct {
	TeamID string `json:"team_id"`
	UserID string `json:"user_id"`
	IsBot  bool   `json:"is_bot"`
}

type BaseEvent struct {
	Type    string `json:"type"`
	User    string `json:"user"`
	TS      string `json:"ts"`
	EventTS string `json:"event_ts"`
}

type MessageEvent struct {
	BaseEvent
	BotID        string `json:"bot_id,omitempty"`
	ParentUserID string `json:"parent_user_id,omitempty"`
	Team         string `json:"team"`
	Channel      string `json:"channel"`
	ChannelType  string `json:"channel_type"`
	ThreadTS     string `json:"thread_ts,omitempty"`
	Text         string `json:"text"`
}

type AppMentionEvent struct {
	BaseEvent
	Team    string `json:"team"`
	Channel string `json:"channel"`
	Text    string `json:"text"`
}
