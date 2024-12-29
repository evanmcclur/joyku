package joycon

import "time"

type ButtonState uint8

const (
	Pressed  ButtonState = 0
	Held     ButtonState = 1
	Released ButtonState = 2
)

// Joycon Button Ids
const (
	ButtonA = 9000
	ButtonB = 9001
	ButtonY = 9002
	ButtonX = 9003
)

type Button struct {
	Id        uint16
	Label     string
	Timestamp time.Time
	State     ButtonState
	mask      int
	shift     int
}

type ButtonHandler func(button *Button) error

type ButtonListener interface {
	// Handlers returns a list of button handlers for this listener
	Handlers() []ButtonHandler
	// Receive returns a read-only channel with button events.
	// Note: if any handlers are registered, the button events will go through them before being
	// output from this channel
	Receive(button *Button)
}

// Factory method
func NewButton(id uint16) *Button {
	button := &Button{}
	button.Id = id
	button.State = Pressed
	return button
}
