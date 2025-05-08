package joycon

import (
	"sync"
)

type Multiplexer interface {
	Output() <-chan *JoyconStatus
}

// FOFIMultiplexer implements a Fan-Out, Fan-In pattern that aggregates multiple input streams and returns a single
// output stream. The returned output stream is closed after all inputs have also closed.
type FOFIMultiplexer struct {
	in   chan (<-chan *JoyconStatus)
	out  chan *JoyconStatus
	once sync.Once
	wg   sync.WaitGroup
}

func NewMultiplexer() *FOFIMultiplexer {
	return &FOFIMultiplexer{
		in:  make(chan (<-chan *JoyconStatus)),
		out: make(chan *JoyconStatus),
	}
}

// Join adds a new stream of Joycon status packets to the output stream
func (m *FOFIMultiplexer) Join(jc *Joycon) {
	m.in <- jc.Status()
}

// Output returns the output channel that all input streams have been joined together to output to.
func (m *FOFIMultiplexer) Output() <-chan *JoyconStatus {
	m.once.Do(func() {
		// Start an output goroutine for each input channel in m.streams; output
		// copies values from c to out until c is closed, then calls wg.Done.
		output := func(c <-chan *JoyconStatus) {
			for js := range c {
				m.out <- js
			}
			m.wg.Done()
		}

		// Start a goroutine that listens to new input streams to be joined
		// with the output.
		go func() {
			for c := range m.in {
				m.wg.Add(1)
				go output(c)
			}
		}()
	})
	return m.out
}

// Close closes all channels. This function will block until all input streams have finished.
//
// Once this has been called, no more input streams can be added, so it should only be called
// at the end of the program or when you know you won't need to add any more input streams.
func (m *FOFIMultiplexer) Close() {
	close(m.in)
	m.wg.Wait()
	close(m.out)
}
