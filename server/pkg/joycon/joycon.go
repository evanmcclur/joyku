package joycon

import (
	"fmt"
	"gocon/internal/report"
	"gocon/internal/spi"
	"gocon/internal/subcommand"
	"image/color"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/sstallion/go-hid"
)

const (
	JoyconVendorId       = 0x057E // Same for every joycon
	LeftJoyconProductId  = 0x2006
	RightJoyconProductId = 0x2007
	ProController        = 0x2009 // TODO: maybe support pro controllers in the future??
)

type AxisData struct {
	X int16
	Y int16
	Z int16
}

type JoyconStatus struct {
	BatteryLevel       BatteryLevel
	ConnectionKind     byte
	LeftButtonSR       bool // If the SR button is being pressed
	LeftButtonSL       bool // If the SL button is being pressed
	ButtonMinus        bool // If the minus button is being pressed
	LeftStickPress     bool // If the left stick is being pressed
	ButtonCapture      bool // If the capture button is being pressed
	DPadDown           bool // If the down d-pad button is being pressed
	DPadUp             bool // If the up d-pad button is being pressed
	DPadRight          bool // If the right d-pad button is being pressed
	DPadLeft           bool // If the left d-pad button is being pressed
	ButtonL            bool // If the L button is being pressed
	ButtonZL           bool // If the ZL button is being pressed
	ButtonY            bool // If the Y button is being pressed
	ButtonX            bool // If the X button is being pressed
	ButtonB            bool // If the B button is being pressed
	ButtonA            bool // If the A button is being pressed
	RightButtonSR      bool // If the SR button is being pressed
	RightButtonSL      bool // If the SL button is being pressed
	ButtonR            bool // If the R button is being pressed
	ButtonZR           bool // If the ZR button is being pressed
	ButtonPlus         bool // If the plus button is being pressed
	RightStickPress    bool // If the right stick is being pressed
	ButtonHome         bool // If the home button is being pressed
	ButtonChargingGrip bool // If the charging grip button is being pressed
	JoystickData       StickData
	Acceleration       AxisData
	GyroscopeData      AxisData
}

func (js *JoyconStatus) String() string {
	sb := strings.Builder{}
	sb.WriteString("---- Joycon Status ----\n")
	sb.WriteString(fmt.Sprintf("joycon battery level: %s\n", js.BatteryLevel))
	sb.WriteString(fmt.Sprintf("joycon connection status: %d\n", js.ConnectionKind))
	sb.WriteString("---- Button States ----\n")
	sb.WriteString(fmt.Sprintf("left joycon left SR button pressed: %t\n", js.LeftButtonSR))
	sb.WriteString(fmt.Sprintf("left joycon left SL button pressed: %t\n", js.LeftButtonSL))
	sb.WriteString(fmt.Sprintf("left joycon minus button pressed: %t\n", js.ButtonMinus))
	sb.WriteString(fmt.Sprintf("left joycon left stick pressed: %t\n", js.LeftStickPress))
	sb.WriteString(fmt.Sprintf("left joycon capture button pressed: %t\n", js.ButtonCapture))
	sb.WriteString(fmt.Sprintf("left joycon d-pad down button pressed: %t\n", js.DPadDown))
	sb.WriteString(fmt.Sprintf("left joycon d-pad up button pressed: %t\n", js.DPadUp))
	sb.WriteString(fmt.Sprintf("left joycon d-pad right button pressed: %t\n", js.DPadRight))
	sb.WriteString(fmt.Sprintf("left joycon d-pad left button pressed: %t\n", js.DPadLeft))
	sb.WriteString(fmt.Sprintf("left joycon L button pressed: %t\n", js.ButtonL))
	sb.WriteString(fmt.Sprintf("left joycon ZL button pressed: %t\n", js.ButtonZL))
	sb.WriteString(fmt.Sprintf("right joycon Y button pressed: %t\n", js.ButtonY))
	sb.WriteString(fmt.Sprintf("right joycon X button pressed: %t\n", js.ButtonX))
	sb.WriteString(fmt.Sprintf("right joycon B button pressed: %t\n", js.ButtonB))
	sb.WriteString(fmt.Sprintf("right joycon A button pressed: %t\n", js.ButtonA))
	sb.WriteString(fmt.Sprintf("right joycon right SR button pressed: %t\n", js.RightButtonSR))
	sb.WriteString(fmt.Sprintf("right joycon right SL button pressed: %t\n", js.RightButtonSL))
	sb.WriteString(fmt.Sprintf("right joycon R button pressed: %t\n", js.ButtonR))
	sb.WriteString(fmt.Sprintf("right joycon ZR button pressed: %t\n", js.ButtonZR))
	sb.WriteString(fmt.Sprintf("right joycon plus button pressed: %t\n", js.ButtonPlus))
	sb.WriteString(fmt.Sprintf("right joycon right stick pressed: %t\n", js.RightStickPress))
	sb.WriteString(fmt.Sprintf("right joycon home button pressed: %t\n", js.ButtonHome))
	sb.WriteString(fmt.Sprintf("right joycon charging grip button pressed: %t\n", js.ButtonChargingGrip))
	sb.WriteString("---- Stick Data ----\n")
	sb.WriteString(fmt.Sprintf("joycon stick horizontal: %d\n", js.JoystickData.Horizontal))
	sb.WriteString(fmt.Sprintf("joycon stick vertical: %d\n", js.JoystickData.Vertical))
	sb.WriteString(fmt.Sprintf("joycon stick direction: %s", js.JoystickData.Direction))
	return sb.String()
}

// Joycon is the representation of the underlying HID device for a Nintendo Switch Joycon attached to the system
type Joycon struct {
	VendorId         uint16              // This will always be 0x057E
	ProductId        uint16              // This will either be 0x2006 or 0x2007 depending on if its the left or right
	Serial           string              // The serial string of the hid device
	Name             string              // The name (product str) of the hid device
	BodyColor        color.Color         // The body color of this joycon
	ButtonColor      color.Color         // The color of this joycons buttons
	StickCalibration StickCalibration    // The stick calibration data for this joycons joystick
	statusC          chan *JoyconStatus  // Channel for receiving joycon status updates
	directionC       chan StickDirection // Channel for receiving joycon stick direction updates
	buttonListeners  []ButtonListener    // Array of button listeners
	device           *hid.Device         // The underlying HID device for this joycon - set after calling Open()
	lock             sync.Mutex          // Internal lock for reading/writing from the state of the joycon
	closed           bool                // If this joycon was closed
}

// Map of connected joycons
var connectedJoycons = make(map[string]bool)

// FindJoycons finds all joycons connected to this device and returns them as an array
func FindJoycons() []*Joycon {
	jc := []*Joycon{}
	hid.Enumerate(JoyconVendorId, hid.ProductIDAny, func(info *hid.DeviceInfo) error {
		if _, ok := connectedJoycons[info.SerialNbr]; ok {
			return nil
		}

		if info.ProductID == LeftJoyconProductId || info.ProductID == RightJoyconProductId {
			jc = append(jc, &Joycon{
				VendorId:  info.VendorID,
				ProductId: info.ProductID,
				Serial:    info.SerialNbr,
				Name:      info.ProductStr,
				device:    nil,
				statusC:   make(chan *JoyconStatus),
			})
			connectedJoycons[info.SerialNbr] = true
		}
		return nil
	})
	return jc
}

// FindFirstJoyconPair finds the first joycon pair and returns them. A joycon pair consists of one left and one right joycon. This function
// will return at most two joycons (one left and one right)
func FindFirstJoyconPair() []*Joycon {
	jc := []*Joycon{}
	hid.Enumerate(JoyconVendorId, hid.ProductIDAny, func(info *hid.DeviceInfo) error {
		if _, ok := connectedJoycons[info.SerialNbr]; ok {
			return nil
		}

		if len(jc) >= 2 {
			return nil
		}

		if info.ProductID != LeftJoyconProductId && info.ProductID != RightJoyconProductId {
			return nil
		}

		if len(jc) == 0 || ((len(jc) == 1 && jc[0].IsLeft() && info.ProductID == RightJoyconProductId) || (len(jc) == 1 && jc[0].IsRight() && info.ProductID == LeftJoyconProductId)) {
			jc = append(jc, &Joycon{
				VendorId:  info.VendorID,
				ProductId: info.ProductID,
				Serial:    info.SerialNbr,
				Name:      info.ProductStr,
				statusC:   make(chan *JoyconStatus),
				device:    nil,
			})
			connectedJoycons[info.SerialNbr] = true
		}
		return nil
	})
	return jc
}

// IsLeft returns whether or not this is a left joycon model
func (j *Joycon) IsLeft() bool {
	return j.ProductId == LeftJoyconProductId
}

// IsRight returns whether or not this is a right joycon model
func (j *Joycon) IsRight() bool {
	return j.ProductId == RightJoyconProductId
}

// Status exposes a readonly channel for parsing joycon status packets
func (j *Joycon) Status() <-chan *JoyconStatus {
	return j.statusC
}

func (j *Joycon) DirectionStatus() <-chan StickDirection {
	return j.directionC
}

func (j *Joycon) RegisterButtonListener(listener ButtonListener) {
	j.buttonListeners = append(j.buttonListeners, listener)
}

func (j *Joycon) notifyButtonListeners(button *Button) {
	for _, listener := range j.buttonListeners {
		for _, handler := range listener.Handlers() {
			handler(button)
		}
		listener.Receive(button)
	}
}

// Connect creates a connection to a Joycon device connected to this system (if one isn't already established)
func (j *Joycon) Connect() error {
	// If we're already connected to a device, return no error
	if j.device != nil {
		return nil
	}

	// Open connection to HID device (Joycon)
	d, err := hid.Open(j.VendorId, j.ProductId, j.Serial)
	if err != nil {
		return err
	}
	j.device = d

	// Read static configuration values from Joycon SPI flash
	err = readColorDataFromSPIFlash(j)
	if err != nil {
		log.Printf("error while reading from spi flash - %s", err.Error())
		return err
	}
	err = readStickCalibrationFromSPIFlash(j)
	if err != nil {
		log.Printf("error while reading from spi flash - %s", err.Error())
		return err
	}

	// Enable IMU so we can receive accelerometer and gyroscope data
	data := []byte{0x01}
	err = subcommand.Send(j.device, subcommand.EnableIMU, data)
	if err != nil {
		return err
	}

	log.Printf("enabling IMU..")
	time.Sleep(time.Millisecond * 500)

	// Enable vibration
	data = []byte{0x01}
	err = subcommand.Send(j.device, subcommand.EnableVibration, data)
	if err != nil {
		return err
	}

	log.Printf("enabling vibration..")
	time.Sleep(time.Millisecond * 500)

	// Configure Joycon to Input Report Mode which outputs its status at 60hz
	data = []byte{0x30}
	err = subcommand.Send(j.device, subcommand.SetInputReportMode, data)
	if err != nil {
		return err
	}

	// Read from joycon in a separate thread
	go j.readStatus()

	return nil
}

// Disconnect closes connection to the Joycon (Note: This function will disconnect joycon from the system)
func (j *Joycon) Disconnect() error {
	j.lock.Lock()
	defer hid.Exit()
	var err error = nil
	if !j.closed {
		j.closed = true
		// disconnect joycon from system
		data := []byte{0x00} // turn off
		subcommand.Send(j.device, subcommand.SetHCIState, data)
		err = j.device.Close()
	}
	j.lock.Unlock()
	return err
}

func (j *Joycon) readStatus() {
	buf := make([]byte, subcommand.ReportLengthBytes)

	// TODO: add a channel to stop this loop if we disconnect
	log.Println("starting reading input report loop")
	for {
		_, err := j.device.Read(buf)
		if err != nil {
			log.Printf("could not read from device %s\n", err.Error())
			break
		}

		js := parseInputReport(j, buf)
		j.statusC <- js
	}

	close(j.statusC)
	j.Disconnect()
}

func parseInputReport(joycon *Joycon, reportData []byte) *JoyconStatus {
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

// ReadColorDataFromSPIFlash reads SPI flash memory on j and stores it in j
func readColorDataFromSPIFlash(j *Joycon) error {
	sfc := spi.SPIFlashReadCommand{
		Address: 0x6050,
		Size:    6,
	}

	data, err := spi.ReadFromSPIFlash(j.device, sfc)
	if err != nil {
		return err
	}

	j.lock.Lock()
	defer j.lock.Unlock()

	j.BodyColor = color.RGBA{
		R: data[0],
		G: data[1],
		B: data[2],
	}
	j.ButtonColor = color.RGBA{
		R: data[3],
		G: data[4],
		B: data[5],
	}

	return nil
}

// TODO: cleanup
func readStickCalibrationFromSPIFlash(j *Joycon) error {
	// 1. Read user stick calibration data
	// 2. If user calibration data is all maxed out (i.e. every value is 255), use the factory configuration instead
	var address uint32
	if j.IsLeft() {
		address = spi.LeftStickUserCalibrationSection
	} else if j.IsRight() {
		address = spi.RightStickUserCalibrationSection
	} else {
		return fmt.Errorf("unknown joycon product id %d", j.ProductId)
	}

	sfc := spi.SPIFlashReadCommand{Address: address, Size: 9}
	data, err := spi.ReadFromSPIFlash(j.device, sfc)
	if err != nil {
		return err
	}

	stickCalibration := data[0:9]
	useUserConfiguration := false
	for _, calibration := range stickCalibration {
		// If the joystick was not calibrated by the user, the entire section will be filled with 255
		if calibration != 255 {
			useUserConfiguration = true
			break
		}
	}

	// Use factory configuration
	if !useUserConfiguration {
		if j.IsLeft() {
			address = spi.LeftStickFactoryCalibrationSection
		} else if j.IsRight() {
			address = spi.RightStickFactoryCalibrationSection
		}

		sfc := spi.SPIFlashReadCommand{Address: address, Size: 9}
		data, err = spi.ReadFromSPIFlash(j.device, sfc)
		if err != nil {
			return err
		}

		stickCalibration = data[0:9]
	}

	// TODO: cleanup
	if j.IsLeft() {
		address = spi.LeftStickDeviceParameters
	} else if j.IsRight() {
		address = spi.RightStickDeviceParameters
	} else {
		return fmt.Errorf("unknown joycon product id %d", j.ProductId)
	}

	sfc = spi.SPIFlashReadCommand{Address: address, Size: 5}
	data, err = spi.ReadFromSPIFlash(j.device, sfc)
	if err != nil {
		return err
	}
	deadzone := uint8(data[3])

	j.lock.Lock()
	defer j.lock.Unlock()

	if j.IsLeft() {
		lsc := unmarshalLeftStickCalibration(stickCalibration)
		j.StickCalibration = lsc
		j.StickCalibration.Deadzone = deadzone
	} else if j.IsRight() {
		rsc := unmarshalRightStickCalibration(stickCalibration)
		j.StickCalibration = rsc
		j.StickCalibration.Deadzone = deadzone
	}

	return nil
}
