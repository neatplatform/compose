package schema

import "encoding/json"

type InteractionPayloadBase struct {
	Type        string `json:"type"`
	User        User   `json:"user"`
	Team        Team   `json:"team"`
	TriggerID   string `json:"trigger_id,omitempty"`
	ResponseURL string `json:"response_url,omitempty"`
}

type InteractionPayloadBlockActions struct {
	InteractionPayloadBase
	APIAppID  string            `json:"api_app_id"`
	Channel   Channel           `json:"channel"`
	Container Container         `json:"container"`
	Actions   []json.RawMessage `json:"actions"`
	State     *State            `json:"state,omitempty"`
	Message   Message           `json:"message"`
}

type InteractionPayloadBlockSuggestion struct {
	InteractionPayloadBase
	APIAppID  string    `json:"api_app_id"`
	Channel   Channel   `json:"channel"`
	Container Container `json:"container"`
	ActionID  string    `json:"action_id"`
	BlockID   string    `json:"block_id"`
	Value     string    `json:"value"`
	Message   Message   `json:"message"`
}

type InteractionPayloadShortcut struct {
	InteractionPayloadBase
	CallbackID string   `json:"callback_id"`
	ActionTS   string   `json:"action_ts"`
	Channel    Channel  `json:"channel,omitempty"`
	MessageTS  string   `json:"message_ts,omitempty"`
	Message    *Message `json:"message,omitempty"`
}

type InteractionPayloadViewSubmission struct {
	InteractionPayloadBase
	APIAppID string       `json:"api_app_id"`
	View     ViewResponse `json:"view"`
}

type InteractionPayloadViewClosed struct {
	InteractionPayloadBase
	IsCleared bool         `json:"is_cleared"`
	View      ViewResponse `json:"view"`
}

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name,omitempty"`
	TeamID   string `json:"team_id,omitempty"`
}

type Team struct {
	ID     string `json:"id"`
	Domain string `json:"domain"`
}

type Channel struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Container struct {
	Type        string `json:"type"`
	ChannelID   string `json:"channel_id"`
	MessageTS   string `json:"message_ts"`
	IsEphemeral bool   `json:"is_ephemeral"`
}
