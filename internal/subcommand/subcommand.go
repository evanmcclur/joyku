package subcommand

import (
	"fmt"
	"log"

	"github.com/sstallion/go-hid"
)

const (
	ReportLengthBytes = 49
	maxPacketNumber   = 15
)

// SubcommandId is an alias for a byte value used to identify subcommands
type SubcommandId byte

// Group of subcommands and their ids
const (
	// Subcommand used to set the type of input mode outputted from the device
	SetInputReportMode SubcommandId = 0x03
	// Subcommand used to set state of Host Controller Interface (disconnect/page/pair/turn off)
	SetHCIState SubcommandId = 0x06
	// Subcommand used to read from the SPI flash
	SPIFlashRead SubcommandId = 0x10
	// Subcommand used to enable or disable the IMU
	EnableIMU SubcommandId = 0x40
	// Subcommand used to enable or disable vibration
	EnableVibration SubcommandId = 0x48
)

// Byte returns the underlying byte value of this subcommand identifier
func (s SubcommandId) Byte() byte {
	return byte(s)
}

// Range of 0-15 - will loop back to 0 every 15 packets
var globalPacketNumber byte = 0

// Default neutral values
var RumbleData = []byte{0x00, 0x01, 0x40, 0x40, 0x00, 0x01, 0x40, 0x40}

// Sends a subcommand to joycon with the given subcommand id (sid) and data (sd)
// https://github.com/dekuNukem/Nintendo_Switch_Reverse_Engineering/blob/master/bluetooth_hid_notes.md
func Send(d *hid.Device, sid SubcommandId, sd []byte) error {
	buf := make([]byte, ReportLengthBytes)
	buf[0] = 1
	buf[1] = globalPacketNumber

	if globalPacketNumber >= maxPacketNumber {
		globalPacketNumber = 0
	} else {
		globalPacketNumber++
	}

	arrayCopy(buf, RumbleData, 2)
	// Set subcommand id and data
	buf[10] = sid.Byte()
	arrayCopy(buf, sd, 11)

	br, err := d.Write(buf)
	if err != nil {
		return fmt.Errorf("could not write to device: %s", err.Error())
	}
	log.Printf("wrote %v (%d) bytes to device", buf, br)
	return nil
}

// arrayCopy copies all the data from src to dst starting at start
func arrayCopy(dst []byte, src []byte, start int) error {
	dl := len(dst)
	sl := len(src)

	if start >= dl {
		return fmt.Errorf("starting index %d is out of range %d", start, dl)
	}

	for i := 0; i < sl; i++ {
		dst[start+i] = src[i]
	}
	return nil
}
