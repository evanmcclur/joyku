package spi

import (
	"encoding/binary"
	"fmt"
	"gocon/internal/report"
	"gocon/internal/subcommand"
	"log"

	"github.com/sstallion/go-hid"
)

const (
	MaxFlashReadInBytes byte = 29

	LeftStickFactoryCalibrationSection  uint32 = 0x603D
	LeftStickDeviceParameters           uint32 = 0x6086
	LeftStickUserCalibrationSection     uint32 = 0x8012
	RightStickFactoryCalibrationSection uint32 = 0x6046
	RightStickDeviceParameters          uint32 = 0x6098
	RightStickUserCalibrationSection    uint32 = 0x801D
)

type SPIFlashReadCommand struct {
	Address uint32 // Address of subsection in SPI flash memory
	Size    byte   // Size of SPI flash read request in bytes - Max size is 29 bytes
}

func (s SPIFlashReadCommand) Data() []byte {
	data := []byte{}
	data = binary.LittleEndian.AppendUint32(data, s.Address)
	data = append(data, s.Size)
	return data
}

// ReadFromSPIFlash reads from joycon SPI flash memory and returns the data or an error if one occurred during reading
func ReadFromSPIFlash(d *hid.Device, sfr SPIFlashReadCommand) ([]byte, error) {
	if sfr.Size > MaxFlashReadInBytes {
		sfr.Size = MaxFlashReadInBytes
	}

	subcommand.Send(d, subcommand.SPIFlashRead, sfr.Data())

	reportBuf := make([]byte, subcommand.ReportLengthBytes)
	_, err := d.Read(reportBuf)
	if err != nil {
		return nil, err
	}

	// Check if the data in this packet does not much what we're expecting
	if reportBuf[0] == report.StandardInputReportWithReplies.Byte() && reportBuf[14] == subcommand.SPIFlashRead.Byte() {
		log.Printf("received expected input report")
	} else {
		log.Printf("received unexpected input report")
	}

	ack := (reportBuf[13] >> 7) == 1
	if !ack {
		return nil, fmt.Errorf("[spi flash] received nack after reading")
	}
	// Return only the data portion of the report and nothing else
	return reportBuf[20:], nil
}
