package subcommand

import (
	"math"

	"github.com/sstallion/go-hid"
)

func NewRumbleCommand(d *hid.Device, freq float64, amp float64) Subcommand {
	freqHF, freqLF := encodeFreq(freq)
	ampHF, ampLF := encodeAmp(amp)

	data := make([]byte, 8)
	//Byte swapping
	data[0] = byte(freqHF & 0xFF)
	data[1] = byte(ampHF) + byte((freqHF>>8)&0xFF) //Add amp + 1st byte of frequency to amplitude byte

	//Byte swapping
	data[2] = freqLF + ((ampLF >> 8) & 0xFF) //Add freq + 1st byte of LF amplitude to the frequency byte
	data[3] = ampLF & 0xFF

	return Subcommand{
		Data:       make([]byte, 8), // purposely empty since this is rumble only
		RumbleOnly: true,
		device:     d,
	}
}

// encodeFreq encodes the given frequency value and returns the high and low values
func encodeFreq(freq float64) (uint16, uint8) {
	// Clamp freq to prevent going above safe boundries
	freq = math.Min(freq, 1252.0)
	// Encode algorithm for frequency
	encodedFreq := uint8(math.Round(math.Log2(freq/10) * 32.0))
	// Convert to Joy-Con HF range. Range in big-endian: 0x0004-0x01FC with +0x0004 steps.
	hf := uint16((encodedFreq - 0x60) * 4)
	// Convert to Joy-Con LF range. Range: 0x01-0x7F.
	lf := uint8(encodedFreq - 0x40)
	// Return both high and low frequency values
	return hf, lf
}

func encodeAmp(amp float64) (uint16, uint8) {
	// Clamp freq to prevent going above safe boundries
	amp = math.Min(amp, 1.00)
	// Float amplitude to hex conversion
	var encodedAmp uint8
	if amp > 0.23 {
		encodedAmp = uint8(math.Round(math.Log2(amp*8.7) * 32.0))
	} else if amp > 0.12 {
		encodedAmp = uint8(math.Round(math.Log2(amp*17.0) * 16.0))
	}
	hf := uint16(encodedAmp * 2) // encoded_hex_amp<<1;
	lf := encodedAmp/2 + 64      // (encoded_hex_amp>>1)+0x40;
	return hf, lf
}
