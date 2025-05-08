package main

import (
	"joyku/internal/bluez"
	"joyku/pkg/handlers"
	"joyku/pkg/joycon"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	quit := make(chan os.Signal, 1)
	defer close(quit)

	signal.Notify(quit, os.Interrupt, syscall.SIGINT)

	conn, err := bluez.Init()
	if err != nil {
		log.Fatalf("Could not initialize connection to Bluetooth adapter, err: %s\n", err)
	}
	adpt := conn.Adapter()

	mux := joycon.NewMultiplexer()

	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))

	http.HandleFunc("/", handlers.Home)
	http.HandleFunc("/search", handlers.Search(adpt))
	http.HandleFunc("/connect", handlers.Connect(mux))
	http.HandleFunc("/disconnect", handlers.Disconnect(adpt))
	http.HandleFunc("/events", handlers.Events(mux))

	go func() {
		log.Println("Running server on localhost:3000")
		http.ListenAndServe(":3000", nil)
	}()

	<-quit
	log.Println("Received SIGTERM, shutting down")
	// cleanup
	joycon.DisconnectAll(func(jc *joycon.Joycon) {
		log.Printf("Disconnecting: %s\n", jc.Name)
		adpt.RemoveDeviceWithSerial(jc.Serial)
	})
	conn.Close()
	mux.Close()
}
