package joycon

// BatteryLevel is an alias for a byte value that is used to determine battery level of joycon
type BatteryLevel byte

const (
	Invalid  BatteryLevel = 0xF
	Full     BatteryLevel = 0x08
	Medium   BatteryLevel = 0x06
	Low      BatteryLevel = 0x04
	Critical BatteryLevel = 0x02
	Empty    BatteryLevel = 0x00
)

var byteToBatteryMap = map[byte]BatteryLevel{
	0: Empty,
	2: Critical,
	4: Low,
	6: Medium,
	8: Full,
}

func (b BatteryLevel) String() string {
	switch b {
	case Full:
		return "Full"
	case Medium:
		return "Medium"
	case Low:
		return "Low"
	case Critical:
		return "Critical"
	case Empty:
		return "Empty"
	default:
		return "Invalid"
	}
}

func BatteryFromByte(v byte) BatteryLevel {
	b, ok := byteToBatteryMap[v]
	if !ok {
		return Invalid
	}
	return b
}
