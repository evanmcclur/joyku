package joycon

import (
	"gocon/internal/report"
	"log"
)

func ParseInputReport(joycon *Joycon, reportData []byte) *JoyconStatus {
	reportId := reportData[0]
	if reportId != report.StandardInputReportWithReplies.Byte() && reportId != report.StandardFullMode.Byte() && reportId != report.NFCIRMode.Byte() {
		log.Printf("received unsupported input report: %d", reportId)
		return nil
	}

	joyconStatus := new(JoyconStatus)
	if joycon.IsLeft() {
		joyconStatus = parseLeftJoyconStatus(reportData, joycon.StickCalibration)
	} else if joycon.IsRight() {
		joyconStatus = parseRightJoyconStatus(reportData, joycon.StickCalibration)
	}
	return joyconStatus

	// Axis data is only set for these input reports
	// if reportId == byte(report.StandardFullMode) || reportId == byte(report.NFCIRMode) {

	// }

	// fmt.Println(js)
}

func parseLeftJoyconStatus(report []byte, sc StickCalibration) *JoyconStatus {
	js := new(JoyconStatus)

	batteryAndConnection := report[2]
	js.BatteryLevel = BatteryFromByte((batteryAndConnection >> 4) & 0xF)
	js.ConnectionKind = (batteryAndConnection >> 1) & 0x03

	// Button states
	leftButtons := report[5]
	sharedButtons := report[4]

	js.DPadDown = (leftButtons & 0x01) != 0
	js.DPadUp = ((leftButtons & 0x02) >> 1) != 0
	js.DPadRight = ((leftButtons & 0x04) >> 2) != 0
	js.DPadLeft = ((leftButtons & 0x08) >> 3) != 0
	js.LeftButtonSR = ((leftButtons & 0x10) >> 4) != 0
	js.LeftButtonSL = ((leftButtons & 0x20) >> 5) != 0
	js.ButtonL = ((leftButtons & 0x40) >> 6) != 0
	js.ButtonZL = ((leftButtons & 0x80) >> 7) != 0

	js.ButtonMinus = (sharedButtons & 0x01) != 0
	js.LeftStickPress = ((sharedButtons & 0x08) >> 3) != 0
	js.ButtonCapture = ((sharedButtons & 0x20) >> 5) != 0
	js.ButtonChargingGrip = ((sharedButtons & 0x80) >> 7) != 0

	leftStickData := report[6:9]

	js.JoystickData = StickData{
		Horizontal: uint16(leftStickData[0]) | ((uint16(leftStickData[1] & 0xF)) << 8),
		Vertical:   uint16(leftStickData[1]>>4) | uint16(leftStickData[2])<<4,
	}
	js.JoystickData.Direction = calculateStickDirection(js.JoystickData, sc)

	return js
}

func parseRightJoyconStatus(report []byte, sc StickCalibration) *JoyconStatus {
	js := new(JoyconStatus)
	batteryAndConnection := report[2]
	js.BatteryLevel = BatteryFromByte((batteryAndConnection >> 4) & 0xF)
	js.ConnectionKind = (batteryAndConnection >> 1) & 0x03

	// Button states
	rightButtons := report[3]
	sharedButtons := report[4]

	js.ButtonY = (rightButtons & 0x01) != 0
	js.ButtonX = ((rightButtons & 0x02) >> 1) != 0
	js.ButtonB = ((rightButtons & 0x04) >> 2) != 0
	js.ButtonA = ((rightButtons & 0x08) >> 3) != 0
	js.RightButtonSR = ((rightButtons & 0x10) >> 4) != 0
	js.RightButtonSL = ((rightButtons & 0x20) >> 5) != 0
	js.ButtonR = ((rightButtons & 0x40) >> 6) != 0
	js.ButtonZR = ((rightButtons & 0x80) >> 7) != 0

	js.ButtonPlus = ((sharedButtons & 0x02) >> 1) != 0
	js.RightStickPress = ((sharedButtons & 0x04) >> 2) != 0
	js.ButtonHome = ((sharedButtons & 0x10) >> 4) != 0
	js.ButtonChargingGrip = ((sharedButtons & 0x80) >> 7) != 0

	rightStickData := report[9:12]

	js.JoystickData = StickData{
		Horizontal: uint16(rightStickData[0]) | ((uint16(rightStickData[1] & 0xF)) << 8),
		Vertical:   uint16(rightStickData[1]>>4) | uint16(rightStickData[2])<<4,
	}
	js.JoystickData.Direction = calculateStickDirection(js.JoystickData, sc)

	return js
}
