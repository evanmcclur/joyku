package bluez

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/godbus/dbus/v5"
)

const (
	bluezService     string          = "org.bluez"
	bluezAdapterPath dbus.ObjectPath = "/org/bluez/hci0"
	bluezAdapterIntf string          = "org.bluez.Adapter1"
	bluezDeviceIntf  string          = "org.bluez.Device1"
	bluezInputIntf   string          = "org.bluez.Input1"
)

// JoyconFilter is a filter that can be used alongside SetDiscoveryFilter to only find devices with names containing
// "Joy-Con". The key is "Pattern" and the value is the string, "Joy-Con", wrapped as a DBus variant value.
var JoyconFilter = map[string]dbus.Variant{
	"Pattern": dbus.MakeVariant("Joy-Con"),
}

// A map where the key is the Dbus Interface (e.g. org.bluez.Device1) and the value is a map of properites
// defined on that interface and their respective values.
type DbusSignalBody = map[string]map[string]dbus.Variant

// Conn represents a connection to the BlueZ Dbus service
type Conn struct {
	conn    *dbus.Conn // Underlying D-Bus connection to the blueZ service
	adapter *Adapter   // Bluetooth adapter
}

// Device represents a bluetooth device
type Device struct {
	Address     string
	AddressType string
	Alias       string
	Blocked     bool
	Bonded      bool
	Paired      bool
	Connected   bool
	Name        string
	Trusted     bool
	path        dbus.ObjectPath
	conn        dbus.BusObject
}

// Connect will connect this bluetooth device to system
func (d *Device) Connect() error {
	if !d.Trusted {
		err := d.conn.Call("org.freedesktop.DBus.Properties.Set", 0, bluezDeviceIntf, "Trusted", dbus.MakeVariant(true)).Err
		if err != nil {
			return fmt.Errorf("could not trust bluetooth device -- %w", err)
		}
		d.Trusted = true
	}

	if !d.Paired {
		err := d.conn.Call(bluezDeviceIntf+".Pair", 0).Err
		if err != nil {
			return fmt.Errorf("could not pair with bluetooth device -- %w", err)
		}
		d.Paired = true
	}

	if !d.Connected {
		err := d.conn.Call(bluezDeviceIntf+".Connect", 0).Err
		if err != nil {
			return fmt.Errorf("could not connect to bluetooth device -- %w", err)
		}
		d.Connected = true
	}

	// Need to make sure the device bonds with the system otherwise it will not be able to establish an HID connection
	bonded := false
	err := d.conn.Call("org.freedesktop.DBus.Properties.Get", 0, bluezDeviceIntf, "Bonded").Store(&bonded)
	if err != nil {
		return fmt.Errorf("could not get property from device -- %w", err)
	}

	if !bonded {
		return fmt.Errorf("%s device could not bond with host", d)
	}

	log.Printf("%s device successfully paired and connected!\n", d)
	return nil
}

// Disconnect will disconnect this bluetooth device from the system; however, it will remain paired and so it will not be
// rediscovered by further scans until RemoveDevice is called on the bluetooth adapter.
func (d *Device) Disconnect() error {
	if !d.Connected {
		log.Printf("Attempted to disconnect an already disconnected device, %s", d)
		return nil
	}

	err := d.conn.Call(bluezDeviceIntf+".Disconnect", 0).Err
	if err != nil {
		return err
	}
	d.Connected = false

	log.Printf("%s device successfully disconnected\n", d)
	return nil
}

func (d *Device) String() string {
	return fmt.Sprintf("%s (%s)", d.Name, d.Address)
}

// Init initializes a connection to the system D-Bus and sets up the Bluetooth service via BlueZ
func Init() (*Conn, error) {
	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		return nil, err
	}

	adpt, err := getAdapter(conn)
	if err != nil {
		return nil, err
	}

	// Initialize the bluetooth adapter by powering it on and setting it to pairable

	if err := adpt.bus.SetProperty(bluezAdapterIntf+".Powered", dbus.MakeVariant(true)); err != nil {
		return nil, fmt.Errorf("could not power bluetooth adapter on: %w", err)
	}

	if err := adpt.bus.SetProperty(bluezAdapterIntf+".Pairable", dbus.MakeVariant(true)); err != nil {
		return nil, fmt.Errorf("could not set bluetooth adapter to pairable: %w", err)
	}

	// New devices found during scanning are sent to the InterfacesAdded signal below. Since thats all we care about,
	// add a match signal so it ignores other signals.

	if err := conn.BusObject().AddMatchSignal("org.freedesktop.DBus.ObjectManager", "InterfacesAdded").Err; err != nil {
		return nil, err
	}

	return &Conn{
		conn:    conn,
		adapter: adpt,
	}, nil
}

func (b *Conn) Adapter() *Adapter {
	return b.adapter
}

// Close closes the underlying D-Bus connection and does any other cleanup
func (b *Conn) Close() error {
	return b.conn.Close()
}

type Adapter struct {
	conn *dbus.Conn
	bus  dbus.BusObject
}

// Scan starts scanning for bluetooth devices until the context has canceled. All devices found during scanning
// are sent to the returned channel as they're discovered. The returned channel is closed after the scan has completed.
func (a *Adapter) Scan(ctx context.Context) (<-chan *Device, error) {
	log.Printf("Starting bluetooth discovery..")

	err := a.bus.Call(bluezAdapterIntf+".StartDiscovery", 0).Err
	if err != nil {
		return nil, err
	}

	deviceC := make(chan *Device)
	signals := make(chan *dbus.Signal)
	a.conn.Signal(signals)

	go func() {
		defer close(deviceC)
		defer close(signals)

		for {
			select {
			case signal := <-signals:
				dp := signal.Body[0].(dbus.ObjectPath)
				dsb := signal.Body[1].(DbusSignalBody)

				// Device signals must have the device interface specified in the signal body. Ignore all other signals.
				if _, ok := dsb[bluezDeviceIntf]; !ok {
					break
				}

				dev, err := parseDevice(a.conn, dp, dsb)
				if err != nil {
					log.Printf("Could not parse bluetooth device: %s", err)
					break
				}
				deviceC <- dev
			case <-ctx.Done():
				log.Println("Scan canceled, stopping bluetooth discovery..")

				a.conn.RemoveSignal(signals)
				// Stop scan and clean up cache of devices that were found
				if err := a.bus.Call(bluezAdapterIntf+".StopDiscovery", 0).Err; err != nil {
					log.Printf("Could not stop bluetooth discovery: %s\n", err)
				}
				return
			}
		}
	}()
	return deviceC, nil
}

// RemoveDevice exposes the BlueZ API for removing connected bluetooth devices from the system
func (a *Adapter) RemoveDevice(d *Device) error {
	return a.bus.Call(bluezAdapterIntf+".RemoveDevice", 0, d.path).Err
}

// RemoveDevice exposes the BlueZ API for removing connected bluetooth devices from the system
func (a *Adapter) RemoveDeviceWithSerial(serial string) error {
	path := dbus.ObjectPath(fmt.Sprintf("%s/dev_%s", a.bus.Path(), strings.Replace(serial, ":", "_", -1)))
	if !path.IsValid() {
		return fmt.Errorf("no device found with serial: %s", serial)
	}
	return a.bus.Call(bluezAdapterIntf+".RemoveDevice", 0, path).Err
}

// SetDiscoveryFilter exposes the BlueZ API for setting a discovery filter that is used during scanning. The key is the
// filter type and the value should be converted to a DBus variant value.
// See: https://web.git.kernel.org/pub/scm/bluetooth/bluez.git/tree/doc/org.bluez.Adapter.rst
func (a *Adapter) SetDiscoveryFilter(dict map[string]dbus.Variant) error {
	return a.bus.Call(bluezAdapterIntf+".SetDiscoveryFilter", 0, dict).Err
}

func getAdapter(c *dbus.Conn) (*Adapter, error) {
	adptObj := c.Object(bluezService, bluezAdapterPath)
	if !adptObj.Path().IsValid() {
		return nil, fmt.Errorf("blueZ adapter path is invalid: %s", adptObj.Path())
	}

	return &Adapter{
		conn: c,
		bus:  adptObj,
	}, nil
}

// parseDevice parses the given signal body and returns a device from what was created.
// See: https://dbus.freedesktop.org/doc/dbus-specification.html#standard-interfaces-objectmanager for body structure
func parseDevice(c *dbus.Conn, dp dbus.ObjectPath, signalBody DbusSignalBody) (*Device, error) {
	device := new(Device)
	device.path = dp
	device.conn = c.Object(bluezService, dp)

	if address, ok := signalBody[bluezDeviceIntf]["Address"]; ok {
		device.Address = address.Value().(string)
	}
	if addrType, ok := signalBody[bluezDeviceIntf]["AddressType"]; ok {
		device.AddressType = addrType.Value().(string)
	}
	if alias, ok := signalBody[bluezDeviceIntf]["Alias"]; ok {
		device.Alias = alias.Value().(string)
	}
	if blocked, ok := signalBody[bluezDeviceIntf]["Blocked"]; ok {
		device.Blocked = blocked.Value().(bool)
	}
	if bonded, ok := signalBody[bluezDeviceIntf]["Bonded"]; ok {
		device.Bonded = bonded.Value().(bool)
	}
	if paired, ok := signalBody[bluezDeviceIntf]["Paired"]; ok {
		device.Paired = paired.Value().(bool)
	}
	if connected, ok := signalBody[bluezDeviceIntf]["Connected"]; ok {
		device.Connected = connected.Value().(bool)
	}
	if name, ok := signalBody[bluezDeviceIntf]["Name"]; ok {
		device.Name = name.Value().(string)
	}
	if trusted, ok := signalBody[bluezDeviceIntf]["Trusted"]; ok {
		device.Trusted = trusted.Value().(bool)
	}
	return device, nil
}
