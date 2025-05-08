package report

const (
	ReportLengthBytes byte = 49
)

// InputReport is an alias for a byte value that corrosponds to a certain input report id
type InputReport byte

const (
	Unknown                        InputReport = 0x00
	StandardInputReportWithReplies InputReport = 0x21
	StandardFullMode               InputReport = 0x30
	NFCIRMode                      InputReport = 0x31
)

func (i InputReport) String() string {
	switch i {
	case StandardInputReportWithReplies:
		return "Standard Input with Replies"
	case StandardFullMode:
		return "Standard Full Mode"
	case NFCIRMode:
		return "NFC/IR MCU Mode"
	default:
		return "Unknown"
	}
}

func (i InputReport) Byte() byte {
	return byte(i)
}

// Supported returns true if support for the given report id is implemented and false otherwise
func Supported(reportID byte) bool {
	return reportID == StandardInputReportWithReplies.Byte() || reportID == StandardFullMode.Byte() || reportID == NFCIRMode.Byte()
}
