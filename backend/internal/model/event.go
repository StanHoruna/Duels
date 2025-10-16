package model

import "github.com/goccy/go-json"

const (
	EventInsufficientFunds uint8 = 1
)

type Event struct {
	Type uint8 `json:"type"`
}

func (e *Event) MarshalBinary() ([]byte, error) {
	return json.Marshal(e)
}

func (e *Event) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, e)
}

func NewEvent(t uint8) *Event {
	return &Event{Type: t}
}
