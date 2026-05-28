package schema

import "encoding/json"

type Message struct {
	Type   string            `json:"type"`
	TS     string            `json:"ts"`
	User   string            `json:"user"`
	Team   string            `json:"team"`
	AppID  string            `json:"app_id"`
	BotID  string            `json:"bot_id"`
	Text   string            `json:"text,omitempty"`
	Blocks []json.RawMessage `json:"blocks,omitempty"`
}
