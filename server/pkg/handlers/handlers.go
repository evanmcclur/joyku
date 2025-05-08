package handlers

import (
	"context"
	"joyku/internal/bluez"
	"joyku/pkg/components"
	"joyku/pkg/joycon"
	"log"
	"net/http"
	"strconv"
	"time"
)

func Home(w http.ResponseWriter, r *http.Request) {
	pair := joycon.FindFirstPair()
	components.Dashboard(pair).Render(r.Context(), w)
}

// Device Searching
// Manual -> Search all HID devices and return the first left and right joycons
// Bluetooth -> Start a scan which will search for Joycons and connect them to the system

func Search(adpt *bluez.Adapter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bfv := r.PostFormValue("bluetooth")
		if bfv == "" {
			w.Header().Set("x-missing-field", "bluetooth")
			http.Error(w, "Missing 'bluetooth' field in request", http.StatusBadRequest)
			return
		}

		useBluetooth, err := strconv.ParseBool(bfv)
		if err != nil {
			log.Printf("Failed to convert to boolean: %s\n", bfv)
			http.Error(w, "Provided 'bluetooth' field is invalid", http.StatusBadRequest)
			return
		}

		var pair joycon.Pair
		if useBluetooth {
			if err = adpt.SetDiscoveryFilter(bluez.JoyconFilter); err != nil {
				log.Printf("Could not set discovery filter, err: %s\n", err)
				http.Error(w, "Failed to start bluetooth discovery", http.StatusInternalServerError)
				return
			}

			ctx, cancel := context.WithTimeout(r.Context(), time.Second*10)
			defer cancel()

			deviceC, err := adpt.Scan(ctx)
			if err != nil {
				log.Printf("Could not start bluetooth scan, err: %s\n", err)
				http.Error(w, "Failed to start bluetooth discovery", http.StatusInternalServerError)
				return
			}

			for device := range deviceC {
				if pair.Left != nil && pair.Right != nil {
					break
				}

				if err := device.Connect(); err != nil {
					log.Printf("Could not connect to Joycon with address: %s, skipping\n", err)
					continue
				}

				jc := joycon.Find(device.Address)
				if jc == nil {
					log.Printf("Could not find Joycon with address: %s, skipping\n", device.Address)
					continue
				}

				if jc.IsLeft() && pair.Left == nil {
					pair.Left = jc
				} else if jc.IsRight() && pair.Right == nil {
					pair.Right = jc
				}
			}
		} else {
			pair = joycon.FindFirstPair()
		}
		components.RenderJoycons(pair).Render(r.Context(), w)
	}
}

func Connect(mux *joycon.FOFIMultiplexer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		serial := r.PostFormValue("joycon")
		if serial == "" {
			w.Header().Set("x-missing-field", "joycon")
			http.Error(w, "Missing 'joycon' field in request", http.StatusBadRequest)
			return
		}

		jc := joycon.Find(serial)
		if jc == nil {
			log.Printf("Could not find Joycon with serial: %s\n", serial)
			http.Error(w, "Could not find Joycon with provided serial number", http.StatusNotFound)
			return
		}

		if err := jc.Connect(); err != nil {
			log.Printf("Failed to connect to %s: %s\n", serial, err)
			http.Error(w, "Failed to connect to Joycon", http.StatusInternalServerError)
			return
		}
		// Add Joycon to multiplexer for event streaming
		mux.Join(jc)
		components.RenderJoycon(jc).Render(r.Context(), w)
	}
}

func Disconnect(adpt *bluez.Adapter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		serial := r.PostFormValue("joycon")
		if serial == "" {
			w.Header().Set("x-missing-field", "joycon")
			http.Error(w, "Missing 'joycon' field in request", http.StatusBadRequest)
			return
		}

		jc := joycon.Find(serial)
		if jc == nil {
			log.Printf("Could not find Joycon with serial: %s\n", serial)
			http.Error(w, "Could not find Joycon with provided serial number", http.StatusNotFound)
			return
		}

		if err := jc.Disconnect(); err != nil {
			log.Printf("Failed to disconnect from %s: %s\n", serial, err)
			http.Error(w, "Failed to disconnect from Joycon", http.StatusInternalServerError)
			return
		}

		if err := adpt.RemoveDeviceWithSerial(jc.Serial); err != nil {
			log.Printf("Could not remove Joycon from bluetooth adapter: %s\n", err)
		}

		// TODO: add a disconnect joycon component and write that to response instead?
		pair := joycon.FindFirstPair()
		components.RenderJoycons(pair).Render(r.Context(), w)
	}
}

func Events(mux *joycon.FOFIMultiplexer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
			return
		}

		ctx := r.Context()

		packets := 0
		for {
			packets += 1
			select {
			case <-ctx.Done():
				log.Println("Client disconnected")
				return
			case _, ok := <-mux.Output():
				if !ok {
					log.Println("Stream closed")
					return
				}
				w.Write([]byte("event: test\n"))
				w.Write([]byte("data: "))
				components.RenderEvent(packets).Render(ctx, w)
				w.Write([]byte("\n"))
				w.Write([]byte("\n\n"))
				flusher.Flush()
			}
		}
	}
}
