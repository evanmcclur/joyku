package joycon

import (
	"log"
	"math"
)

type StickDirection uint8

const (
	InvalidStickDirection StickDirection = iota
	None
	StickUp
	StickUpperRight
	StickRight
	StickLowerRight
	StickDown
	StickLowerLeft
	StickLeft
	StickUpperLeft
)

func (d StickDirection) String() string {
	switch d {
	case StickUp:
		return "Up"
	case StickUpperRight:
		return "Upper right"
	case StickRight:
		return "Right"
	case StickLowerRight:
		return "Lower right"
	case StickDown:
		return "Down"
	case StickLowerLeft:
		return "Lower left"
	case StickLeft:
		return "Left"
	case StickUpperLeft:
		return "Upper left"
	case None:
		return "Center"
	default:
		return "Invalid"
	}
}

// StickData contains the horizontal and vertical values of the joycon
type StickData struct {
	Horizontal uint16
	Vertical   uint16
	Direction  StickDirection
}

type StickCalibration struct {
	XAxisCenter         uint16
	XAxisMinBelowCenter uint16
	XAxisMaxAboveCenter uint16
	YAxisCenter         uint16
	YAxisMinBelowCenter uint16
	YAxisMaxAboveCenter uint16
	Deadzone            uint8
}

func unmarshalLeftStickCalibration(calibration []byte) StickCalibration {
	sc := StickCalibration{}
	sc.XAxisMaxAboveCenter = ((uint16(calibration[1]) << 8) & 0xF00) | uint16(calibration[0])
	sc.YAxisMaxAboveCenter = (uint16(calibration[2]) << 4) | (uint16(calibration[1]) >> 4)
	sc.XAxisCenter = ((uint16(calibration[4]) << 8) & 0xF00) | uint16(calibration[3])
	sc.YAxisCenter = (uint16(calibration[5]) << 4) | (uint16(calibration[4]) >> 4)
	sc.XAxisMinBelowCenter = ((uint16(calibration[7]) << 8) & 0xF00) | uint16(calibration[6])
	sc.YAxisMinBelowCenter = (uint16(calibration[8]) << 4) | (uint16(calibration[7]) >> 4)
	return sc
}

func unmarshalRightStickCalibration(calibration []byte) StickCalibration {
	sc := StickCalibration{}
	sc.XAxisCenter = ((uint16(calibration[1]) << 8) & 0xF00) | uint16(calibration[0])
	sc.YAxisCenter = (uint16(calibration[2]) << 4) | (uint16(calibration[1]) >> 4)
	sc.XAxisMinBelowCenter = ((uint16(calibration[4]) << 8) & 0xF00) | uint16(calibration[3])
	sc.YAxisMinBelowCenter = (uint16(calibration[5]) << 4) | (uint16(calibration[4]) >> 4)
	sc.XAxisMaxAboveCenter = ((uint16(calibration[7]) << 8) & 0xF00) | uint16(calibration[6])
	sc.YAxisMaxAboveCenter = (uint16(calibration[8]) << 4) | (uint16(calibration[7]) >> 4)
	return sc
}

func calculateStickDirection(sd StickData, sc StickCalibration) StickDirection {
	stick_x_min := sc.XAxisCenter - sc.XAxisMinBelowCenter
	stick_x_max := sc.XAxisCenter + sc.XAxisMaxAboveCenter
	stick_y_min := sc.YAxisCenter - sc.YAxisMinBelowCenter
	stick_y_max := sc.YAxisCenter + sc.YAxisMaxAboveCenter

	stick_x_center := float64((stick_x_min + stick_x_max) / 2)
	stick_y_center := float64((stick_y_min + stick_y_max) / 2)

	if math.Pow((float64(sd.Horizontal)-stick_x_center), 2)+math.Pow((float64(sd.Vertical)-stick_y_center), 2) < math.Pow(float64(sc.Deadzone), 2) {
		// joystick is within deadzone - no direction
		return None
	}

	x := clamp((float64(sd.Horizontal) - stick_x_center) / (stick_x_center - 1))
	y := clamp((float64(sd.Vertical) - stick_y_center) / (stick_y_center - 1))

	stickDegrees := radiansToDegrees(math.Atan2(y, x))
	if stickDegrees < 0 {
		stickDegrees += 360
	}

	if stickDegrees > 60 && stickDegrees < 120 {
		return StickUp
	} else if stickDegrees <= 60 && stickDegrees >= 30 {
		return StickUpperRight
	} else if (stickDegrees >= 0 && stickDegrees < 30) || stickDegrees > 330 {
		return StickRight
	} else if stickDegrees >= 300 && stickDegrees <= 330 {
		return StickLowerRight
	} else if stickDegrees > 240 && stickDegrees < 300 {
		return StickDown
	} else if stickDegrees >= 210 && stickDegrees <= 240 {
		return StickLowerLeft
	} else if stickDegrees > 150 && stickDegrees < 210 {
		return StickLeft
	} else if stickDegrees >= 120 && stickDegrees <= 150 {
		return StickUpperLeft
	} else {
		log.Printf("Unknown direction!!! Angle: %f", stickDegrees)
		return InvalidStickDirection
	}
}

func clamp(v float64) float64 {
	if v > 1.0 {
		return 1.0
	} else if v < -1.0 {
		return -1.0
	}
	return v
}

func radiansToDegrees(radians float64) float64 {
	return (radians * 180) / math.Pi
}
