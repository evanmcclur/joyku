package roku

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"time"
)

type RokuDevice struct {
	Name        string `xml:"friendly-device-name"`
	Serial      string `xml:"serial-number"`
	Vendor      string `xml:"vendor-name"`
	ModelNumber string `xml:"model-number"`
	ModelName   string `xml:"model-name"`
	ip          string
	port        int
	httpClient  http.Client
	keyState    *RokuKeyState
}

// NewDevice returns a new roku device with the default timeout
func NewDevice(ip string, port int) *RokuDevice {
	return &RokuDevice{
		httpClient: http.Client{
			Timeout: time.Duration(time.Second * 1),
		},
		ip:       ip,
		port:     port,
		keyState: NewRokuKeyState(),
	}
}

// NewDeviceWithTimeout specifies how long to wait before timing out device commands
func NewDeviceWithTimeout(ip string, port int, timeout time.Duration) *RokuDevice {
	return &RokuDevice{
		httpClient: http.Client{
			Timeout: timeout,
		},
		ip:       ip,
		port:     port,
		keyState: NewRokuKeyState(),
	}
}

// QueryDevice retrieves device information and updates the given device with the retrieved info
func QueryDevice(device *RokuDevice) error {
	resp, err := device.httpClient.Get(fmt.Sprintf("http://%s:%d/query/device-info", device.ip, device.port))
	if err != nil {
		return fmt.Errorf("could not retrieve device info: %w", err)
	}

	defer resp.Body.Close()
	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("could not read device response: %w", err)
	}

	err = xml.Unmarshal(bytes, &device)
	if err != nil {
		return fmt.Errorf("could not unmarshal device info: %w", err)
	}
	return nil
}

func (r *RokuDevice) SendKeypress(key Keypress) {
	now := time.Now()

	if r.keyState.previousKey == key {
		if !r.keyState.keyTime.IsZero() && now.Sub(r.keyState.keyTime).Milliseconds() > 150 {
			if !r.keyState.holding {
				// we only need to enter a hold event once for a key
				r.keyState.holding = true
				r.sendECPCommand(Hold, key)
			}
		} else {
			// do not update time since this is the same key
			r.sendECPCommand(Press, key)
		}
	} else {
		if r.keyState.holding {
			r.sendECPCommand(Release, r.keyState.previousKey)
		} else {
			r.sendECPCommand(Press, key)
		}
		// reset
		r.keyState.previousKey = key
		r.keyState.keyTime = now
		r.keyState.holding = false
	}
}

// SendKeypress sends a keypress command to the roku device
func (r *RokuDevice) sendECPCommand(ecp ECPCommand, key Keypress) error {
	// Ignore empty key presses
	if key == None {
		return nil
	}

	var err error = nil
	// _, err := r.httpClient.Post(fmt.Sprintf("http://%s:%d/%s/%s", r.ip, r.port, ecp, key), "xml", http.NoBody)
	if err != nil {
		return fmt.Errorf("could not send keypress to %s device: %w", r.Name, err)
	}
	fmt.Printf("Command: %s, Key: %s\n", ecp, key)
	return nil
}
