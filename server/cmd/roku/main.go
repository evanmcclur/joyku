package main

import (
	"gocon/cmd/roku/config"
	"gocon/pkg/joycon"
	"gocon/pkg/roku"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const DeviceReconnectAttempts int = 3

type DirectionAggregator struct {
	directions   map[joycon.StickDirection]int
	max          int
	maxDirection joycon.StickDirection
	lock         sync.Mutex
}

func NewDirectionAggregator() *DirectionAggregator {
	return &DirectionAggregator{
		directions: make(map[joycon.StickDirection]int, 60),
		max:        0,
	}
}

func (d *DirectionAggregator) Add(dir joycon.StickDirection) {
	d.lock.Lock()
	defer d.lock.Unlock()

	dirCount := 0
	if count, ok := d.directions[dir]; ok {
		d.directions[dir] = count + 1
		dirCount = count + 1
	} else {
		d.directions[dir] = 1
		dirCount = 1
	}

	if dirCount > d.max {
		d.max = dirCount
		d.maxDirection = dir
	}
}

func (d *DirectionAggregator) Count() int {
	d.lock.Lock()
	defer d.lock.Unlock()

	total := 0
	for _, c := range d.directions {
		total += c
	}
	return total
}

func (d *DirectionAggregator) Clear() {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.directions = make(map[joycon.StickDirection]int, 60)
	d.max = 0
}

func main() {
	// quit channel should be used for signaling when we need to terminate the program
	quit := make(chan os.Signal, 1)
	defer close(quit)

	signal.Notify(quit, os.Interrupt, syscall.SIGINT)

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	cfg, err := config.NewRokuConfig()
	if err != nil {
		logger.Error("could not create roku config", "error", err.Error())
		return
	}

	rokuDevice := roku.NewDevice(cfg.Ip, cfg.Port)
	if err := roku.QueryDevice(rokuDevice); err != nil {
		logger.Error("could not connect to roku device", "error", err.Error())
		return
	}

	joycons := joycon.FindFirstJoyconPair()
	if len(joycons) == 0 {
		logger.Error("could not find any joycons")
		return
	}
	logger.Debug("found joycons", "count", len(joycons))

	joyconDevice := joycons[0]
	for _, jc := range joycons {
		err := jc.Connect()
		if err != nil {
			logger.Error("could not connect to joycon", "error", err.Error())
		}
		defer jc.Disconnect()
	}

	// go CheckDeviceConnection(logger, rokuDevice, quit)

	da := NewDirectionAggregator()

	for {
		select {
		case js, ok := <-joyconDevice.Status():
			if !ok {
				quit <- os.Interrupt
				break
			}
			TranslateJoyconStatus(logger, js, da, rokuDevice)
		case <-quit:
			logger.Info("received sigterm - shutting down")
			return
		}
	}
}

// CheckDeviceConnection checks if the given roku device r is reachable. If it cannot be reached within DeviceReconnectAttempts,
func CheckDeviceConnection(logger *slog.Logger, r *roku.RokuDevice, quit chan os.Signal) {
	reconnectAttempts := 0

	ticker := time.NewTicker(time.Duration(5 * time.Second))
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ok, _ := roku.PingDevice(r)
			if ok {
				// logger.Info("device is reachable")
				reconnectAttempts = 0
			} else if reconnectAttempts < DeviceReconnectAttempts {
				reconnectAttempts += 1
				logger.Error("device is not reachable, attempting to reconnect", "attempts", reconnectAttempts)
			} else {
				logger.Error("could not ping device after multiple attempts", "attempts", reconnectAttempts)
				quit <- os.Interrupt
			}
		case <-quit:
			return
		}
	}
}

// Check stick direction and send corresponding keypress
func TranslateJoyconStatus(logger *slog.Logger, js *joycon.JoyconStatus, da *DirectionAggregator, rd *roku.RokuDevice) {
	da.Add(js.JoystickData.Direction)

	if da.Count() == 11 {
		switch da.maxDirection {
		case joycon.None:
			break
		case joycon.StickUp:
			rd.SendKeypress(roku.KeyUp)
		case joycon.StickRight:
			rd.SendKeypress(roku.KeyRight)
		case joycon.StickDown:
			rd.SendKeypress(roku.KeyDown)
		case joycon.StickLeft:
			rd.SendKeypress(roku.KeyLeft)
		default:
			logger.Warn("unsupported direction value", "direction", da.maxDirection.String())
		}
		da.Clear()
	}

	if js.ButtonA {
		rd.SendKeypress(roku.KeySelect)
	} else if js.ButtonB {
		rd.SendKeypress(roku.KeyBack)
	} else if js.ButtonHome {
		rd.SendKeypress(roku.KeyHome)
	} else {
		rd.SendKeypress(roku.None)
	}
}
