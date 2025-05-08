package roku

import (
	"time"
)

type Keypress string

const (
	None             Keypress = "None" // Represents no key being pressed
	KeyHome          Keypress = "Home" // Home takes you back to home tab on Roku device
	KeyRev           Keypress = "Rev"
	KeyFwd           Keypress = "Fwd"
	KeyPlay          Keypress = "Play"
	KeySelect        Keypress = "Select" // Select is the same as the "Ok" button on the remote
	KeyLeft          Keypress = "Left"
	KeyRight         Keypress = "Right"
	KeyDown          Keypress = "Down"
	KeyUp            Keypress = "Up"
	KeyBack          Keypress = "Back"
	KeyInstantReplay Keypress = "InstantReplay"
	KeyInfo          Keypress = "Info"
	KeyBackspace     Keypress = "Backspace"
	KeySearch        Keypress = "Search"
	KeyEnter         Keypress = "Enter"
	KeyVolumeUp      Keypress = "VolumeUp"   // Increases the volume - only supported on Roku TVs
	KeyVolumeDown    Keypress = "VolumeDown" // Decreases the volume - only supported on Roku TVs
	KeyVolumeMute    Keypress = "VolumeMute" // Mutes the volume - only supported on Roku TVs
	KeyInputTuner    Keypress = "InputTuner" // Used for Live TV - only supported on Roku TVs
	KeyInputHDMI1    Keypress = "InputHDMI1" // Used to switch to HDMI input one  - only supported on Roku TVs
	KeyInputHDMI2    Keypress = "InputHDMI2" // Used to switch to HDMI input two - only supported on Roku TVs
	KeyInputHDMI3    Keypress = "InputHDMI3" // Used to switch to HDMI input three - only supported on Roku TVs
	KeyInputHDMI4    Keypress = "InputHDMI4" // Used to switch to HDMI input four - only supported on certain Roku TVs
	KeyInputAV1      Keypress = "InputAV1"   // Used to switch to AV one - only supported on Roku TVs
	KeyPowerOn       Keypress = "PowerOn"    // May not be supported by all devices
	KeyPowerOff      Keypress = "PowerOff"   // May not be supported by all devices
)

func (k Keypress) String() string {
	return string(k)
}

type ECPCommand string

const (
	Press   ECPCommand = "keypress"
	Release ECPCommand = "keyup"
	Hold    ECPCommand = "keydown"
)

type RokuKeyState struct {
	previousKey Keypress
	keyTime     time.Time
	holding     bool
}

func NewRokuKeyState() *RokuKeyState {
	rka := &RokuKeyState{}
	rka.previousKey = None
	rka.holding = false
	return rka
}
