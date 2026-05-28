package schema

type ActionBase struct {
	Type     string `json:"type"`
	ActionTS string `json:"action_ts"`
	ActionID string `json:"action_id"`
	BlockID  string `json:"block_id"`
}

type ActionButton struct {
	ActionBase
	Value string `json:"value"`
	Style string `json:"style,omitempty"`
	Text  Text   `json:"text"`
}

type ActionSelectOption struct {
	ActionBase
	Placeholder    Text   `json:"placeholder,omitempty"`
	SelectedOption Option `json:"selected_option"`
}

type ActionSelectOptions struct {
	ActionBase
	Placeholder     Text     `json:"placeholder,omitempty"`
	SelectedOptions []Option `json:"selected_options"`
}

type ActionPickDate struct {
	ActionBase
	InitialDate  string `json:"initial_date"`
	SelectedDate string `json:"selected_date"`
}

type ActionPickTime struct {
	ActionBase
	InitialTime  string `json:"initial_time"`
	SelectedTime string `json:"selected_time"`
}

type ActionPickDateTime struct {
	ActionBase
	InitialDateTime  int `json:"initial_date_time"`
	SelectedDateTime int `json:"selected_date_time"`
}
