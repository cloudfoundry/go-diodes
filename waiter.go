package diodes

import (
	"context"
)

// Waiter will use a channel signal to alert the reader to when data is
// available.
type Waiter struct {
	Diode
	c   chan struct{}
	ctx context.Context
}

// WaiterConfigOption can be used to setup the waiter.
type WaiterConfigOption func(*Waiter)

// WithWaiterContext sets the context to cancel any retrieval (Next()). It
// will not change any results for adding data (Set()). Default is
// context.Background().
func WithWaiterContext(ctx context.Context) WaiterConfigOption {
	return WaiterConfigOption(func(c *Waiter) {
		c.ctx = ctx
	})
}

// NewWaiter returns a new Waiter that wraps the given diode.
func NewWaiter(d Diode, opts ...WaiterConfigOption) *Waiter {
	w := new(Waiter)
	w.Diode = d
	w.c = make(chan struct{}, 1)
	w.ctx = context.Background()

	for _, opt := range opts {
		opt(w)
	}

	return w
}

// Set invokes the wrapped diode's Set with the given data and uses broadcast
// to wake up any readers.
func (w *Waiter) Set(data GenericDataType) {
	w.Diode.Set(data)
	w.broadcast()
}

// broadcast sends to the channel if it can.
func (w *Waiter) broadcast() {
	select {
	case w.c <- struct{}{}:
	default:
	}
}

// Next returns the next data point on the wrapped diode. If there is no new
// data, it will wait for Set to be called or the context to be done. If the
// context is done, then nil will be returned.
func (w *Waiter) Next() GenericDataType {
	for {
		data, ok := w.Diode.TryNext() // nolint:staticcheck
		if ok {
			return data
		}
		select {
		case <-w.ctx.Done():
			return nil
		case <-w.c:
		}
	}
}
