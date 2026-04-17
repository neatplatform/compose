package schema

import "encoding/json"

type View struct {
	Type         string            `json:"type"`
	CallbackID   string            `json:"callback_id"`
	Title        Text              `json:"title"`
	Submit       Text              `json:"submit"`
	Close        Text              `json:"close"`
	ClearOnClose bool              `json:"clear_on_close,omitempty"`
	Blocks       []json.RawMessage `json:"blocks"`
}

type ViewResponse struct {
	View
	State State `json:"state"`
}
