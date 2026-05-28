package schema

import "encoding/json"

type Text struct {
	Type     string `json:"type"`
	Text     string `json:"text"`
	Emoji    bool   `json:"emoji,omitempty"`
	Verbatim bool   `json:"verbatim,omitempty"`
}

type Option struct {
	Value string `json:"value"`
	Text  Text   `json:"text"`
}

type OptionGroup struct {
	Label   Text     `json:"label"`
	Options []Option `json:"options"`
}

type BlockBase struct {
	Type    string `json:"type"`
	BlockID string `json:"block_id"`
}

type BlockSection struct {
	BlockBase
	Text      Text            `json:"text"`
	Accessory json.RawMessage `json:"accessory,omitempty"`
}

type BlockActions struct {
	BlockBase
	Elements []json.RawMessage `json:"elements"`
}

type ElementBase struct {
	Type     string `json:"type"`
	ActionID string `json:"action_id"`
}

type ElementButton struct {
	ElementBase
	Value string `json:"value"`
	Style string `json:"style,omitempty"`
	Text  Text   `json:"text"`
}

type ElementCheckboxes struct {
	ElementBase
	Options []Option `json:"options"`
}

type ElementRadioButtons struct {
	ElementBase
	Options []Option `json:"options"`
}

type ElementSelect struct {
	ElementBase
	Placeholder Text     `json:"placeholder,omitempty"`
	Options     []Option `json:"options"`
}

type ElementMultiSelect struct {
	ElementBase
	Placeholder Text     `json:"placeholder,omitempty"`
	Options     []Option `json:"options"`
}

type ElementOverflow struct {
	ElementBase
	Options []Option `json:"options"`
}

type ElementFeedbackButtons struct {
	ElementBase
	PositiveButton Option `json:"positive_button"`
	NegativeButton Option `json:"negative_button"`
}

type ElementDatePicker struct {
	ElementBase
	InitialDate string `json:"initial_date"`
	Placeholder Text   `json:"placeholder,omitempty"`
}

type ElementTimePicker struct {
	ElementBase
	InitialTime string `json:"initial_time"`
	Placeholder Text   `json:"placeholder,omitempty"`
}

type ElementDateTimePicker struct {
	ElementBase
	InitialDateTime int `json:"initial_date_time"`
}
