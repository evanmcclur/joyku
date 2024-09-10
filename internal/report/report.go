package report

// InputReport is an alias
type InputReport byte

const (
	Unknown                        InputReport = 0x00
	StandardInputReportWithReplies InputReport = 0x21
	StandardFullMode               InputReport = 0x30
	NFCIRMode                      InputReport = 0x31
)

func (i InputReport) Byte() byte {
	return byte(i)
}
