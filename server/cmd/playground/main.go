package main

import (
	"fmt"
	"gocon/pkg/joycon"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// quit channel should be used for signaling when we need to terminate the program
	quit := make(chan os.Signal, 1)
	defer close(quit)

	signal.Notify(quit, os.Interrupt, syscall.SIGINT)

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

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

	packetNumber := 0
	for {
		select {
		case js := <-joyconDevice.Status():
			packetNumber += 1
			go ReceivedStatusUpdate(logger, packetNumber, js)
		case <-quit:
			logger.Info("received sigterm - shutting down")
			return
		}
	}
}

func ReceivedStatusUpdate(logger *slog.Logger, packet int, js *joycon.JoyconStatus) {
	logger.Info(fmt.Sprintf("Packet %d", packet))
	logger.Info(js.String())
}
