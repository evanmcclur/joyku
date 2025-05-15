package subcommand

import (
	"fmt"
	"joyku/internal/report"

	"github.com/sstallion/go-hid"
)

const (
	maxPacketNumber = 15
)

// SubcommandID is an alias for a byte value used to identify subcommands
type SubcommandID byte

// Group of subcommands and their ids
const (
	// Subcommand used to set the type of input mode outputted from the device
	SetInputReportMode SubcommandID = 0x03
	// Subcommand used to set state of Host Controller Interface (disconnect/page/pair/turn off)
	SetHCIState SubcommandID = 0x06
	// Subcommand used to read from the SPI flash
	SPIFlashRead SubcommandID = 0x10
	// Subcommand used to enable or disable the IMU
	EnableIMU SubcommandID = 0x40
	// Subcommand used to enable or disable vibration
	EnableVibration SubcommandID = 0x48
)

// Byte returns the underlying byte value of this subcommand identifier
func (s SubcommandID) Byte() byte {
	return byte(s)
}

// 4 bit value - will loop back to 0 every 15 packets
var globalPacketNumber byte = 0

// Default neutral values
var RumbleDefault = []byte{0x00, 0x01, 0x40, 0x40, 0x00, 0x01, 0x40, 0x40}

type Subcommand struct {
	ID         SubcommandID
	RumbleOnly bool
	Rumble     []byte
	Data       []byte
	device     *hid.Device
}

func NewInputReportCommand(d *hid.Device) Subcommand {
	return Subcommand{
		ID:     SetInputReportMode,
		Data:   []byte{0x30},
		device: d,
	}
}

// Sends a subcommand to joycon with the given subcommand id (sid) and data (sd)
// https://github.com/dekuNukem/Nintendo_Switch_Reverse_Engineering/blob/master/bluetooth_hid_notes.md
func Send(d *hid.Device, sid SubcommandID, sd []byte) error {
	buf := make([]byte, report.ReportLengthBytes)
	buf[0] = 1
	buf[1] = globalPacketNumber

	bufferCopy(buf, RumbleDefault, 2)
	// Set subcommand id and data
	buf[10] = sid.Byte()
	bufferCopy(buf, sd, 11)

	_, err := d.Write(buf)
	if err != nil {
		return fmt.Errorf("could not write to device: %s", err.Error())
	}
	globalPacketNumber = (globalPacketNumber + 1) % 15
	return nil
}

// bufferCopy copies all the data from src to dst starting at start (inclusive)
func bufferCopy(dst []byte, src []byte, start int) error {
	dl := len(dst)
	sl := len(src)

	if start >= dl {
		return fmt.Errorf("starting index %d is out of range %d", start, dl)
	}

	for i := range sl {
		dst[start+i] = src[i]
	}
	return nil
}
