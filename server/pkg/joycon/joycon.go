package joycon

import (
	"context"
	"fmt"
	"image/color"
	"joyku/internal/report"
	"joyku/internal/spi"
	"joyku/internal/subcommand"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/sstallion/go-hid"
)

const (
	JoyconVendorID       uint16 = 0x057E // Same for every joycon
	LeftJoyconProductID  uint16 = 0x2006
	RightJoyconProductID uint16 = 0x2007
	ProController        uint16 = 0x2009 // Not supported
)

type AxisData struct {
	X float64
	Y float64
	Z float64
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
	sb.WriteString(fmt.Sprintf("joycon stick direction: %s\n", js.JoystickData.Direction))
	sb.WriteString("---- Accelerometer Data ----\n")
	sb.WriteString(fmt.Sprintf("joycon acceleration x: %f\n", js.Acceleration.X))
	sb.WriteString(fmt.Sprintf("joycon acceleration y: %f\n", js.Acceleration.Y))
	sb.WriteString(fmt.Sprintf("joycon acceleration z: %f", js.Acceleration.Z))
	sb.WriteString("---- Gyroscope Data ----\n")
	sb.WriteString(fmt.Sprintf("joycon gyroscope x: %f\n", js.GyroscopeData.X))
	sb.WriteString(fmt.Sprintf("joycon gyroscope y: %f\n", js.GyroscopeData.Y))
	sb.WriteString(fmt.Sprintf("joycon gyroscope z: %f", js.GyroscopeData.Z))
	return sb.String()
}

// Joycon is the representation of the underlying HID device for a Nintendo Switch Joycon attached to the system
type Joycon struct {
	VendorID         uint16             // This will always be 0x057E
	ProductID        uint16             // This will either be 0x2006 or 0x2007 depending on if its the left or right
	Serial           string             // The serial number of the HID device (This equivalent to the MAC address)
	Name             string             // The name (product str) of the HID device
	BodyColor        color.Color        // The body color of this Joycon
	ButtonColor      color.Color        // The color of this joycons buttons
	StickCalibration StickCalibration   // The stick calibration data for this joycons joystick
	statusC          chan *JoyconStatus // Channel for receiving joycon status updates
	closeC           chan struct{}      // Channel used for notifying when the Joycon was closed
	device           *hid.Device        // The underlying HID device for this joycon - set after calling Connect()
	lock             sync.Mutex         // Internal lock for reading/writing the state of the Joycon
	closed           bool               // If this Joycon is closed and no longer able to provide data - set after calling Disconnect()
}

// Pair represents a Joycon "pair", which consists of a left and right Joycon
type Pair struct {
	Left  *Joycon
	Right *Joycon
}

func (p Pair) Empty() bool {
	return p.Left == nil && p.Right == nil
}

// Map of connected joycons
var connectedJoycons = make(map[string]*Joycon)

// Find attempts to find a Joycon connected to the system with the given serial number
func Find(serial string) *Joycon {
	if j, ok := connectedJoycons[serial]; ok {
		return j
	}

	jc := new(Joycon)
	hid.Enumerate(JoyconVendorID, hid.ProductIDAny, func(info *hid.DeviceInfo) error {
		if jc != nil {
			return nil
		}
		if strings.EqualFold(serial, info.SerialNbr) {
			jc = &Joycon{
				VendorID:  info.VendorID,
				ProductID: info.ProductID,
				Serial:    info.SerialNbr,
				Name:      info.ProductStr,
				device:    nil,
				statusC:   make(chan *JoyconStatus),
				closeC:    make(chan struct{}),
				closed:    false,
			}
			connectedJoycons[info.SerialNbr] = jc
		}
		return nil
	})
	return jc
}

// FindAll finds all joycons connected to this device and returns them
func FindAll() []*Joycon {
	joycons := []*Joycon{}
	hid.Enumerate(JoyconVendorID, hid.ProductIDAny, func(info *hid.DeviceInfo) error {
		if jc, ok := connectedJoycons[info.SerialNbr]; ok {
			joycons = append(joycons, jc)
			return nil
		}

		if info.ProductID == LeftJoyconProductID || info.ProductID == RightJoyconProductID {
			jc := &Joycon{
				VendorID:  info.VendorID,
				ProductID: info.ProductID,
				Serial:    info.SerialNbr,
				Name:      info.ProductStr,
				device:    nil,
				statusC:   make(chan *JoyconStatus),
				closeC:    make(chan struct{}),
				closed:    false,
			}
			joycons = append(joycons, jc)
			connectedJoycons[info.SerialNbr] = jc
		}
		return nil
	})
	return joycons
}

// FindFirstPair finds the first joycon pair and returns them. A joycon pair consists of one left and one right joycon.
func FindFirstPair() Pair {
	pair := Pair{}

	if len(connectedJoycons) > 0 {
		for _, joycon := range connectedJoycons {
			if joycon.IsLeft() && pair.Left == nil {
				pair.Left = joycon
			} else if joycon.IsRight() && pair.Right == nil {
				pair.Right = joycon
			}
		}
	}

	if pair.Left != nil && pair.Right != nil {
		return pair
	}

	hid.Enumerate(JoyconVendorID, hid.ProductIDAny, func(info *hid.DeviceInfo) error {
		// we already found a pair, skip
		if pair.Left != nil && pair.Right != nil {
			return nil
		}

		// ignore invalid product id values
		if info.ProductID != LeftJoyconProductID && info.ProductID != RightJoyconProductID {
			log.Printf("Received unexpected ProductID value for Joycon: %d, ignoring\n", info.ProductID)
			return nil
		}

		if info.ProductID == LeftJoyconProductID && pair.Left == nil {
			jc := &Joycon{
				VendorID:  info.VendorID,
				ProductID: info.ProductID,
				Serial:    info.SerialNbr,
				Name:      info.ProductStr,
				device:    nil,
				statusC:   make(chan *JoyconStatus),
				closeC:    make(chan struct{}),
				closed:    false,
			}
			connectedJoycons[info.SerialNbr] = jc
			pair.Left = jc
		} else if info.ProductID == RightJoyconProductID && pair.Right == nil {
			jc := &Joycon{
				VendorID:  info.VendorID,
				ProductID: info.ProductID,
				Serial:    info.SerialNbr,
				Name:      info.ProductStr,
				device:    nil,
				statusC:   make(chan *JoyconStatus),
				closeC:    make(chan struct{}),
				closed:    false,
			}
			connectedJoycons[info.SerialNbr] = jc
			pair.Right = jc
		}
		return nil
	})
	return pair
}

// DisconnectAll disconnects all Joycons connected to the system and removes them from internal cache. Each Joycon is
// exposed to the given function after it has been disconnecting, allowing for further cleanup/processing.
//
// This function ignores all errors returned by Joycon.Disconnect()
func DisconnectAll(closeFunc func(jc *Joycon)) {
	for serial, jc := range connectedJoycons {
		jc.Disconnect()
		delete(connectedJoycons, serial)
		if closeFunc != nil {
			closeFunc(jc)
		}
	}
	hid.Exit()
}

// IsLeft returns whether or not this is a left joycon model
func (j *Joycon) IsLeft() bool {
	return j.ProductID == LeftJoyconProductID
}

// IsRight returns whether or not this is a right joycon model
func (j *Joycon) IsRight() bool {
	return j.ProductID == RightJoyconProductID
}

func (j *Joycon) IsConnected() bool {
	return j.device != nil && !j.closed
}

// Status exposes a readonly channel for parsing joycon status packets
func (j *Joycon) Status() <-chan *JoyconStatus {
	return j.statusC
}

// Connect attempts to initiate a connection via HID to this Joycon device (if one isn't already established).
// Calling this function will populate BodyColor, ButtonColor, and StickCalibration
func (j *Joycon) Connect() error {
	j.lock.Lock()
	// If this joycon has been closed and can no longer be used, return an error
	if j.closed {
		j.lock.Unlock()
		return fmt.Errorf("the connection to this joycon (%s) has been closed", j.Name)
	}
	// If we're already connected to a device
	if j.device != nil {
		j.lock.Unlock()
		return fmt.Errorf("a connection to this joycon (%s) has already been made", j.Name)
	}

	// Open connection to HID device (Joycon)
	d, err := hid.Open(j.VendorID, j.ProductID, j.Serial)
	if err != nil {
		j.lock.Unlock()
		return err
	}
	j.device = d
	j.lock.Unlock()

	ctx := context.TODO()

	// Read static configuration values from Joycon SPI flash memory and update this joycon with the data (if no error)
	err = readColorDataFromSPIFlash(ctx, j)
	if err != nil {
		log.Printf("Error while reading color data from spi flash memory - %s", err.Error())
		return err
	}
	err = readStickCalibrationFromSPIFlash(ctx, j)
	if err != nil {
		log.Printf("Error while reading stick data from spi flash memory - %s", err.Error())
		return err
	}
	// err = readAxisCalibration(ctx, j)
	// if err != nil {
	// 	log.Printf("Error while reading axis data from spi flash memory - %s", err.Error())
	// 	return err
	// }

	// Enable IMU so we can receive accelerometer and gyroscope data
	data := []byte{0x01}
	err = subcommand.Send(j.device, subcommand.EnableIMU, data)
	if err != nil {
		return err
	}

	log.Printf("Enabling IMU..")
	time.Sleep(time.Millisecond * 500)

	// Enable vibration
	data = []byte{0x01}
	err = subcommand.Send(j.device, subcommand.EnableVibration, data)
	if err != nil {
		return err
	}

	log.Printf("Enabling vibration..")
	time.Sleep(time.Millisecond * 500)

	// Configure Joycon to Input Report Mode which outputs its status at 60hz
	data = []byte{0x30}
	err = subcommand.Send(j.device, subcommand.SetInputReportMode, data)
	if err != nil {
		return err
	}

	go j.readStatus()
	return nil
}

// Disconnect closes connection to the Joycon and its connection to the system. Because this causes the Joycon to disconnect
// from the system, this should only be called when you're done with this Joycon. In order to reestablish a connection with this Joycon,
// it must be rediscovered by using the FindJoycons or FindFirstJoyconPair functions.
func (j *Joycon) Disconnect() error {
	j.lock.Lock()
	if j.closed || j.device == nil {
		j.lock.Unlock()
		return fmt.Errorf("the connection to this joycon (%s) has already been closed", j.Name)
	}
	j.closed = true
	j.lock.Unlock()

	// Power the Joy-Con off
	data := []byte{0x00}
	err := subcommand.Send(j.device, subcommand.SetHCIState, data)
	if err != nil {
		return err
	}
	delete(connectedJoycons, j.Serial)

	// This must be done before the HID device is closed to prevent reading after closing
	j.closeC <- struct{}{}
	close(j.closeC)

	err = j.device.Close()
	if err != nil {
		return err
	}
	close(j.statusC)
	return err
}

func (j *Joycon) readStatus() {
	// Make one buffer to be reused for each input report
	buf := make([]byte, report.ReportLengthBytes)
	// Number of consecutive retries before giving up and disconnecting from the device
	retries := 5
	func() {
		log.Println("Starting input report loop")
		for {
			select {
			case <-j.closeC:
				log.Println("Device was closed, stopping report loop")
				return
			default:
				// Read will block if there is no data and timeout if it blocks for too long
				_, err := j.device.ReadWithTimeout(buf, time.Second)
				if err != nil {
					if retries <= 0 {
						log.Printf("Exceeded number of retries while reading from device: %s\n", err)
						return
					}
					retries -= 1
					log.Printf("An error occurred while reading from device: %s, retrying (%d left)", err, retries)
				}
				// Reset number of retries after each successful read
				retries = 5

				js := parseInputReport(j, buf)
				if js != nil {
					j.statusC <- js
				}
			}
		}
	}()
	j.Disconnect()
}

func parseInputReport(joycon *Joycon, reportData []byte) *JoyconStatus {
	reportId := reportData[0]
	if !report.Supported(reportId) {
		log.Printf("Received unsupported input report: %d", reportId)
		return nil
	}

	var joyconStatus *JoyconStatus
	if joycon.IsLeft() {
		joyconStatus = parseLeftJoyconStatus(reportData, joycon.StickCalibration)
	} else {
		joyconStatus = parseRightJoyconStatus(reportData, joycon.StickCalibration)
	}

	// TODO: Finish implementing gyroscope and accelerometer support
	// 		 - Will probably need some sort of filter here to reduce noise

	// if reportId == report.StandardFullMode.Byte() || reportId == report.NFCIRMode.Byte() {
	// 	// three samples are provided for gyroscope and acceleration
	// 	for i := 0; i < 3; i++ {
	// 		offset := (12 * i)
	// 		// parse sample
	// 		// the constant values are the base values for the first sample
	// 		ax := int16(reportData[13+offset]) | int16(reportData[14+offset])<<8
	// 		ay := int16(reportData[15+offset]) | int16(reportData[16+offset])<<8
	// 		az := int16(reportData[17+offset]) | int16(reportData[18+offset])<<8
	// 		gx := int16(reportData[19+offset]) | int16(reportData[20+offset])<<8
	// 		gy := int16(reportData[21+offset]) | int16(reportData[22+offset])<<8
	// 		gz := int16(reportData[23+offset]) | int16(reportData[24+offset])<<8

	// 		joyconStatus.Acceleration.X = float64(ax)
	// 		joyconStatus.Acceleration.Y = float64(ay)
	// 		joyconStatus.Acceleration.Z = float64(az)
	// 		joyconStatus.GyroscopeData.X = float64(gx)
	// 		joyconStatus.GyroscopeData.Y = float64(gy)
	// 		joyconStatus.GyroscopeData.Z = float64(gz)
	// 	}
	// }
	return joyconStatus
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
func readColorDataFromSPIFlash(ctx context.Context, j *Joycon) error {
	sfc := spi.SPIFlashReadCommand{
		Address: spi.BodyColorSection,
		Size:    6,
	}

	data, err := spi.Read(ctx, j.device, sfc)
	if err != nil {
		return err
	}

	j.BodyColor = color.RGBA{
		R: data[0],
		G: data[1],
		B: data[2],
		A: 100,
	}
	j.ButtonColor = color.RGBA{
		R: data[3],
		G: data[4],
		B: data[5],
		A: 100,
	}
	return nil
}

// TODO: cleanup
func readStickCalibrationFromSPIFlash(ctx context.Context, j *Joycon) error {
	// 1. Read user stick calibration data
	// 2. If user calibration data is all maxed out (i.e. every value is 255), use the factory configuration instead
	var address uint32
	if j.IsLeft() {
		address = spi.LeftStickUserCalibrationSection
	} else if j.IsRight() {
		address = spi.RightStickUserCalibrationSection
	}

	sfc := spi.SPIFlashReadCommand{Address: address, Size: 9}
	data, err := spi.Read(ctx, j.device, sfc)
	if err != nil {
		return err
	}

	rawStickCalibration := data[0:9]
	useUserConfiguration := false
	for _, calibration := range rawStickCalibration {
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
		data, err = spi.Read(ctx, j.device, sfc)
		if err != nil {
			return err
		}

		rawStickCalibration = data[0:9]
	}

	// TODO: cleanup
	if j.IsLeft() {
		address = spi.LeftStickDeviceParameters
	} else if j.IsRight() {
		address = spi.RightStickDeviceParameters
	} else {
		return fmt.Errorf("unknown joycon product id %d", j.ProductID)
	}

	sfc = spi.SPIFlashReadCommand{Address: address, Size: 5}
	data, err = spi.Read(ctx, j.device, sfc)
	if err != nil {
		return err
	}
	deadzone := data[3]

	j.lock.Lock()
	defer j.lock.Unlock()

	if j.IsLeft() {
		lsc := unmarshalLeftStick(rawStickCalibration)
		j.StickCalibration = lsc
		j.StickCalibration.Deadzone = deadzone
	} else if j.IsRight() {
		rsc := unmarshalRightStick(rawStickCalibration)
		j.StickCalibration = rsc
		j.StickCalibration.Deadzone = deadzone
	}

	return nil
}

func readAxisCalibration(ctx context.Context, j *Joycon) error {
	// 1. Read user stick calibration data
	// 2. If user calibration data is all maxed out (i.e. every value is 255), use the factory configuration instead
	sfc := spi.SPIFlashReadCommand{Address: spi.AxisMotionSensorUserCalibrationSection, Size: 24}
	data, err := spi.Read(ctx, j.device, sfc)
	if err != nil {
		return err
	}

	useUserConfiguration := false
	for _, calibration := range data {
		// If the joystick was not calibrated by the user, the entire section will be filled with 255
		if calibration != 255 {
			useUserConfiguration = true
			break
		}
	}

	// Use factory configuration
	if !useUserConfiguration {
		sfc := spi.SPIFlashReadCommand{Address: spi.AxisMotionSensorFactoryCalibrationSection, Size: 24}
		data, err = spi.Read(ctx, j.device, sfc)
		if err != nil {
			return err
		}
	}

	// TODO - do some stuff
	return nil
}
