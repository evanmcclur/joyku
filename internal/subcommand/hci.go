package subcommand

import "github.com/sstallion/go-hid"

const (
	HCIDisconnect         byte = 0x00
	HCIRebootAndReconnect byte = 0x01
	HCIRebootAndPair      byte = 0x02
)

func NewHCIStateCommand(d *hid.Device, state byte) Subcommand {
	return Subcommand{
		ID:     SetHCIState,
		Data:   []byte{byte(state)},
		device: d,
	}
}
