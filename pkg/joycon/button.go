package joycon

// Button represents a button on the joycon
type Button struct {
	Id      int    // Id is a unique integer assigned to each button
	Label   string // Label is the name associated with this button such as 'A', 'B', etc.
	Pressed bool
}

func (b Button) String() string {
	return b.Label
}

var LeftZRButton Button = Button{Label: "Left ZR", Id: 1}
var LeftSRButton Button = Button{Label: "Left SR", Id: 2}
var LeftSLButton Button = Button{Label: "Left SL", Id: 3}
var MinusButton Button = Button{Label: "Minus", Id: 4}
var LeftStickButton Button = Button{Label: "Left Stick", Id: 5}
var CaptureButton Button = Button{Label: "Capture", Id: 6}
var DPadDownButton Button = Button{Label: "D-pad Down", Id: 7}
var DPadUpButton Button = Button{Label: "D-pad Up", Id: 8}
var DPadRightButton Button = Button{Label: "D-pad Right", Id: 9}
var DPadLeftButton Button = Button{Label: "D-pad Left", Id: 10}
var LButton Button = Button{Label: "L", Id: 11}
var ZLButton Button = Button{Label: "ZL", Id: 12}
var YButton Button = Button{Label: "Y", Id: 13}
var XButton Button = Button{Label: "X", Id: 14}
var BButton Button = Button{Label: "B", Id: 15}
var AButton Button = Button{Label: "A", Id: 16}
var RightSRButton Button = Button{Label: "Right SR", Id: 17}
var RightSLButton Button = Button{Label: "Right SL", Id: 18}
var RButton Button = Button{Label: "R", Id: 19}
var ZRButton Button = Button{Label: "ZR", Id: 20}
var PlusButton Button = Button{Label: "Plus", Id: 21}
var RightStickButton Button = Button{Label: "Right Stick", Id: 22}
var HomeButton Button = Button{Label: "Home", Id: 23}
var ChargingGripButton Button = Button{Label: "Charging Grip", Id: 24}
