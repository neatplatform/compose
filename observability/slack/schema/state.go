package schema

import "encoding/json"

type State struct {
	Values map[string]map[string]json.RawMessage `json:"values"`
}

type StateValueBase struct {
	Type string `json:"type"`
}

type StateValueOption struct {
	StateValueBase
	SelectedOption Option `json:"selected_option"`
}

type StateValueOptions struct {
	StateValueBase
	SelectedOptions []Option `json:"selected_options"`
}

type StateValueDate struct {
	StateValueBase
	SelectedDate string `json:"selected_date"`
}

type StateValueTime struct {
	StateValueBase
	SelectedTime string `json:"selected_time"`
}

type StateValueDateTime struct {
	StateValueBase
	SelectedDateTime int `json:"selected_date_time"`
}
