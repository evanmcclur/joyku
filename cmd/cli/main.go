package main

import (
	"fmt"
	"joyku/pkg/joycon"
	"joyku/pkg/roku"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	// parse arguments while ignoring program path
	args := os.Args[1:]
	if len(args) < 2 {
		fmt.Println("Missing command-line arguments")
		os.Exit(1)
	}

	manualArg := args[0]
	if manualArg != "--manual" && manualArg != "-m" {
		fmt.Println("Missing manual command-line argument")
		os.Exit(1)
	}

	manual := args[1] == "true"
	run(manual)
}

func run(manual bool) {
	// Quit channel should be used for signaling when we need to terminate the program
	quit := make(chan os.Signal, 1)
	defer close(quit)

	// Subscribe to interrupt and sigint events so we can shutdown
	signal.Notify(quit, os.Interrupt, syscall.SIGINT)

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

	// Find and connect to Joycons
	//
	// MANUAL YES: Look for devices already connected to the system.
	// MANUAL NO: Attempt to find a Joycon using bluetooth and connect it to the system.

	joycons := joycon.FindAll()
	if len(joycons) == 0 {
		log.Fatalln("No joycons paired to system.")
	}
	log.Printf("Found %d joycon(s)!\n", len(joycons))

	joyconDevice := joycons[0]
	for _, jc := range joycons {
		err := jc.Connect()
		if err != nil {
			log.Fatalf("Could not connect to joycon: %s\n", err.Error())
		}
		defer jc.Disconnect()
	}

	for {
		select {
		case js, ok := <-joyconDevice.Status():
			if !ok {
				log.Println("Joycon status channel closed, shutting down")
				return
			}
			log.Println(js.String())
		case <-quit:
			log.Println("Received sigterm, shutting down")
			return
		}
	}
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
