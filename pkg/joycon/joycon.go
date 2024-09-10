package joycon

import (
	"fmt"
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
	VendorId         uint16             // This will always be 0x057E
	ProductId        uint16             // This will either be 0x2006 or 0x2007 depending on if its the left or right
	Serial           string             // The serial string of the hid device
	Name             string             // The name (product str) of the hid device
	BodyColor        color.Color        // The body color of this joycon
	ButtonColor      color.Color        // The color of this joycons buttons
	StickCalibration StickCalibration   // The stick calibration data for this joycons joystick
	statusC          chan *JoyconStatus // Channel for receiving joycon status updates
	device           *hid.Device        // The underlying HID device for this joycon - set after calling Open()
	lock             sync.Mutex         // Internal lock for reading/writing from the state of the joycon
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

func (j *Joycon) Connect() error {
	if j.device != nil {
		return nil
	}

	d, err := hid.Open(j.VendorId, j.ProductId, j.Serial)
	if err != nil {
		return err
	}
	j.device = d

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

	data := []byte{0x01}
	err = subcommand.Send(j.device, subcommand.EnableIMU, data)
	if err != nil {
		return err
	}

	log.Printf("enabling IMU..")
	time.Sleep(time.Millisecond * 500)

	data = []byte{0x01}
	err = subcommand.Send(j.device, subcommand.EnableVibration, data)
	if err != nil {
		return err
	}

	log.Printf("enabling vibration..")
	time.Sleep(time.Millisecond * 500)

	data = []byte{0x30}
	err = subcommand.Send(j.device, subcommand.SetInputReportMode, data)
	if err != nil {
		return err
	}

	go j.startStatusUpdates()
	return nil
}

func (j *Joycon) Disconnect() error {
	defer hid.Exit()

	log.Printf("attempting to disconnect joycon..")
	data := []byte{0x00}
	subcommand.Send(j.device, subcommand.SetHCIState, data)
	return j.device.Close()
}

// Status exposes a readonly channel for parsing joycon status packets
func (j *Joycon) Status() <-chan *JoyconStatus {
	return j.statusC
}

func (j *Joycon) startStatusUpdates() {
	buf := make([]byte, subcommand.ReportLengthBytes)

	// it may not really matter if we block or not since this loop is running its own thread
	// for the lifetime of the program
	err := j.device.SetNonblock(true)
	if err != nil {
		log.Fatalf("could not enable non-blocking mode on device %s\n", j.Name)
	}

	log.Println("starting reading input report loop")
	for {
		_, err := j.device.Read(buf)
		// if the error is a timeout, ignore it, otherwise break from loop and log
		if err != nil && err.Error() != "timeout" {
			log.Fatalf("could not read from device %s\n", err.Error())
		}

		// no new data available, skip
		if err != nil {
			continue
		}

		js := ParseInputReport(j, buf)
		j.statusC <- js
	}
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
	// if j.IsLeft() {
	// 	address = spi.LeftStickDeviceParameters
	// } else if j.IsRight() {
	// 	address = spi.RightStickDeviceParameters
	// } else {
	// 	return fmt.Errorf("unknown joycon product id %d", j.ProductId)
	// }

	// sfc = spi.SPIFlashReadCommand{Address: address, Size: 18}
	// data, err = spi.ReadFromSPIFlash(j, p, sfc)
	// if err != nil {
	// 	return err
	// }
	// log.Printf("Stick device parameters: %v", data)

	j.lock.Lock()
	defer j.lock.Unlock()

	if j.IsLeft() {
		lsc := unmarshalLeftStickCalibration(stickCalibration)
		j.StickCalibration = lsc
		// log.Printf("Left Stick Calibration:\nX Axis Max Above Center: %d\nY Axis Max Above Center: %d\nX Axis Center: %d\nY Axis Center: %d\nX Axis Min Above Center: %d\nY Axis Min Above Center: %d\n", lsc.XAxisMaxAboveCenter, lsc.YAxisMaxAboveCenter, lsc.XAxisCenter, lsc.YAxisCenter, lsc.XAxisMinBelowCenter, lsc.YAxisMinBelowCenter)
	} else if j.IsRight() {
		rsc := unmarshalRightStickCalibration(stickCalibration)
		j.StickCalibration = rsc
		// log.Printf("Right Stick Calibration:\nX Axis Center: %d\nY Axis Center: %d\nX Axis Min Center: %d\nY Axis Min Center: %d\nX Axis Max Above Center: %d\nY Axis Max Above Center: %d\n", rsc.XAxisCenter, rsc.YAxisCenter, rsc.XAxisMinBelowCenter, rsc.YAxisMinBelowCenter, rsc.XAxisMaxAboveCenter, rsc.YAxisMaxAboveCenter)
		// log.Printf("Right stick X-axis min/max range: [%d, %d]; Right stick Y-axis min/max range: [%d, %d]", rstick_x_min, rstick_x_max, rstick_y_min, rstick_y_max)
	}

	return nil
}
