package schema

import "encoding/json"

type MessageResponse struct {
	ResponseType    string `json:"response_type"`
	Text            string `json:"text"`
	ReplaceOriginal bool   `json:"replace_original"`
	DeleteOriginal  bool   `json:"delete_original"`
}

type BlockSuggestionResponse struct {
	Options      []Option      `json:"options,omitempty"`
	OptionGroups []OptionGroup `json:"option_groups,omitempty"`
}

type PostMessageRequest struct {
	Channel  string `json:"channel"`
	Text     string `json:"text"`
	ThreadTS string `json:"thread_ts,omitempty"`
}

type PostMessageResponse struct {
	OK      bool   `json:"ok"`
	Error   string `json:"error,omitempty"`
	Channel string `json:"channel,omitempty"`
	TS      string `json:"ts,omitempty"`
}

type AddReactionRequest struct {
	Channel   string `json:"channel"`
	Timestamp string `json:"timestamp"`
	Name      string `json:"name"`
}

type AddReactionResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

type OpenViewRequest struct {
	TriggerID string          `json:"trigger_id"`
	View      json.RawMessage `json:"view"`
}

type OpenViewResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}
