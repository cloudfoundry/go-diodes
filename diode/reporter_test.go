package diode

import (
	"fmt"
	"testing"
	"time"
)

type fakeReporter struct {
	done chan struct{}

	alerts  int
	dropped uint64
}

func (fr *fakeReporter) Alert(dropped uint64) {
	fr.alerts++
	fr.dropped += dropped
	close(fr.done)
}
func (fr *fakeReporter) Warn(msg string) {}

func TestReporter(t *testing.T) {
	done := make(chan struct{})
	fr := &fakeReporter{done: done}
	d := New(5, WithReporter(fr))

	for i := range [7]int{} {
		s := fmt.Sprintf("test-%d", i)
		d.Set(GenericData(&s))
	}

	_, _ = d.TryNext()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
	}

	if fr.alerts != 1 {
		t.Errorf("Alert() called %d times; want 1 time", fr.alerts)
	}
	if fr.dropped != 5 {
		t.Errorf("Alert(%d) called; want Alert(5)", fr.dropped)
	}
}
