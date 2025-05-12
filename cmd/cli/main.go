package main

import (
	"context"
	"fmt"
	"joyku/internal/bluez"
	"joyku/pkg/joycon"
	"joyku/pkg/roku"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

func main() {
	// The first argument is the program path, grab everything after that
	args := os.Args[1:]
	if len(args) < 2 {
		fmt.Println("Missing command-line arguments")
		printHelp()
		os.Exit(1)
	}

	manualArg := args[0]
	if manualArg != "--manual" && manualArg != "-m" {
		fmt.Println("Missing manual command-line argument")
		printHelp()
		os.Exit(1)
	}

	// Quit channel should be used for signaling when we need to terminate the program
	quit := make(chan os.Signal, 1)
	defer close(quit)

	// Subscribe to interrupt and sigint events so we can shutdown gracefully
	signal.Notify(quit, os.Interrupt, syscall.SIGINT)

	manual := strings.EqualFold(args[1], "true")
	run(manual, quit)
}

// printHelp prints example cli usage string to standard output
func printHelp() {
	fmt.Println("usage: joyku_cli (--manual | -m) <boolean>")
}

func run(manual bool, quit <-chan os.Signal) {
	// Setup Roku device connection
	cfg, err := roku.NewRokuConfig()
	if err != nil {
		log.Fatalf("Could not create roku config. %s\n", err.Error())
	}

	rokuDevice := roku.NewDevice(cfg.Ip, cfg.Port)
	if err := roku.QueryDevice(rokuDevice); err != nil {
		log.Fatalf("Could not connect to roku device: %s\n", err.Error())
	}
	log.Printf("Successfully connected to %s!\n", rokuDevice.Name)

	start := func(search func() []*joycon.Joycon) {
		joycons := search()
		if len(joycons) == 0 {
			fmt.Println("No Joycons were found!")
			return
		}
		fmt.Printf("Found %d Joycons\n", len(joycons))

		mux := joycon.NewMultiplexer()
		for _, joycon := range joycons {
			if err := joycon.Connect(); err != nil {
				fmt.Printf("Failed to connect to %s\n, skipping: %s", joycon.Name, err)
				continue
			}
			mux.Join(joycon)
			defer joycon.Disconnect()
		}

		for {
			select {
			case js, ok := <-mux.Output():
				if !ok {
					log.Println("Joycon status channel closed, shutting down")
					return
				}
				log.Println(js.String())
			case <-quit:
				log.Println("Received SIGINT, shutting down")
				return
			}
		}
	}

	// Find and connect to Joycons
	//
	// MANUAL YES: Look for devices already connected to the system.
	// MANUAL NO: Attempt to find a Joycon using bluetooth and connect it to the system.

	if manual {
		start(manualConnect)
	} else {
		start(wirelessConnect)
	}
}

func manualConnect() []*joycon.Joycon {
	return joycon.FindAll()
}

func wirelessConnect() []*joycon.Joycon {
	joycons := make([]*joycon.Joycon, 0)

	conn, err := bluez.Init()
	if err != nil {
		fmt.Printf("Could not create BlueZ D-Bus connection, err: %s\n", err)
		return joycons
	}
	defer conn.Close()

	// TODO: Look into extending this function so it can accept a context from main
	// TODO: Add CLI argument to set timeout duration
	ctx, cancel := context.WithTimeout(context.TODO(), time.Minute)
	defer cancel()

	conn.Adapter().SetDiscoveryFilter(bluez.JoyconFilter)
	scanC, err := conn.Adapter().Scan(ctx)
	if err != nil {
		fmt.Printf("Could not start bluetooth scan, err: %s\n", err)
		return joycons
	}

	for device := range scanC {
		joycon := joycon.Find(device.Address)
		if joycon != nil {
			joycons = append(joycons, joycon)
		}
	}
	return joycons
}

// TODO: Cleanup below

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

	dirCount := 1
	if count, ok := d.directions[dir]; ok {
		d.directions[dir] = count + 1
		dirCount = count + 1
	} else {
		d.directions[dir] = 1
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

// Check stick direction and send corresponding keypress
func TranslateJoyconStatus(js *joycon.JoyconStatus, da *DirectionAggregator, rd *roku.RokuDevice) {
	if SupportedDirection(js.JoystickData.Direction) {
		da.Add(js.JoystickData.Direction)
	}

	if da.Count() == 11 {
		switch da.maxDirection {
		case joycon.StickUp:
			rd.SendKeypress(roku.KeyUp)
		case joycon.StickRight:
			rd.SendKeypress(roku.KeyRight)
		case joycon.StickDown:
			rd.SendKeypress(roku.KeyDown)
		case joycon.StickLeft:
			rd.SendKeypress(roku.KeyLeft)
		default:
			log.Printf("warn - unsupported direction value: %s\n", da.maxDirection.String())
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

func SupportedDirection(dir joycon.StickDirection) bool {
	return dir == joycon.StickUp || dir == joycon.StickRight || dir == joycon.StickDown || dir == joycon.StickLeft
}
