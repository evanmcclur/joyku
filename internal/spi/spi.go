package spi

import (
	"context"
	"encoding/binary"
	"fmt"
	"joyku/internal/report"
	"joyku/internal/subcommand"
	"log"
	"time"

	"github.com/sstallion/go-hid"
)

const (
	MaxFlashReadInBytes byte = 29

	LeftStickFactoryCalibrationSection        uint32 = 0x603D
	AxisMotionSensorFactoryCalibrationSection uint32 = 0x6020
	RightStickFactoryCalibrationSection       uint32 = 0x6046
	BodyColorSection                          uint32 = 0x6050
	ButtonColorSection                        uint32 = 0x6053
	LeftStickDeviceParameters                 uint32 = 0x6086
	RightStickDeviceParameters                uint32 = 0x6098
	LeftStickUserCalibrationSection           uint32 = 0x8012
	RightStickUserCalibrationSection          uint32 = 0x801D
	AxisMotionSensorUserCalibrationSection    uint32 = 0x8028
)

type SPIFlashReadCommand struct {
	Address uint32 // Address of subsection in SPI flash memory
	Size    uint8  // Size of SPI flash read request in bytes - Max size is 29 bytes
}

func (s SPIFlashReadCommand) Data() []byte {
	data := []byte{}
	data = binary.LittleEndian.AppendUint32(data, s.Address)
	data = append(data, s.Size)
	return data
}

// Read reads from joycon SPI flash memory and returns the data or an error if one occurred during reading.
func Read(ctx context.Context, d *hid.Device, sfr SPIFlashReadCommand) ([]byte, error) {
	if sfr.Size > MaxFlashReadInBytes {
		sfr.Size = MaxFlashReadInBytes
	}

	err := subcommand.Send(d, subcommand.SPIFlashRead, sfr.Data())
	if err != nil {
		return nil, err
	}

	ctx, cancelCtx := context.WithTimeout(ctx, time.Second)
	defer cancelCtx()

	resp, err := awaitResponse(ctx, d)
	if err != nil {
		return nil, err
	}

	ack := (resp[13] >> 7) == 1
	if !ack {
		return nil, fmt.Errorf("[spi flash] received nack after reading")
	}
	// Return only the data portion of the report and nothing else
	return resp[20:], nil
}

// awaitResponse reads from device until it finds the expected input report or it times out.
//
// Reports that do not match the expected response are essentially ignored
func awaitResponse(ctx context.Context, d *hid.Device) ([]byte, error) {
	buffer := make([]byte, report.ReportLengthBytes)
	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("received unexpected input report -- report %d, subcommand: %d", buffer[0], buffer[14])
		default:
			_, err := d.Read(buffer)
			if err != nil {
				return nil, err
			}

			if buffer[0] == report.StandardInputReportWithReplies.Byte() && buffer[14] == subcommand.SPIFlashRead.Byte() {
				log.Printf("received expected input report")
				return buffer, nil
			}
		}
	}
}
