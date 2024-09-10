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
	IsTV        bool   `xml:"is-tv"`
	ip          string
	port        int
	httpClient  http.Client
}

// NewDevice returns a new roku device with the default timeout
func NewDevice(rc *RokuConfig) *RokuDevice {
	return &RokuDevice{
		httpClient: http.Client{
			Timeout: time.Duration(time.Second * 1),
		},
		ip:   rc.Ip,
		port: rc.Port,
	}
}

// NewDeviceWithTimeout specifies how long to wait before timing out device commands
func NewDeviceWithTimeout(rc *RokuConfig, timeout time.Duration) *RokuDevice {
	return &RokuDevice{
		httpClient: http.Client{
			Timeout: timeout,
		},
		ip:   rc.Ip,
		port: rc.Port,
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

// PingDevice tests whether or not the given roku device is reachable
func PingDevice(device *RokuDevice) (bool, error) {
	_, err := device.httpClient.Get(fmt.Sprintf("http://%s:%d/query/device-info", device.ip, device.port))
	if err != nil {
		return false, fmt.Errorf("unable to ping %s device: %w", device.Name, err)
	}
	return true, nil
}

// SendKeypress sends a keypress command to the roku device
func SendKeypress(device *RokuDevice, kp Keypress) error {
	_, err := device.httpClient.Post(fmt.Sprintf("http://%s:%d/keypress/%s", device.ip, device.port, kp), "xml", http.NoBody)
	if err != nil {
		return fmt.Errorf("could not send keypress to %s device: %w", device.Name, err)
	}
	return nil
}
