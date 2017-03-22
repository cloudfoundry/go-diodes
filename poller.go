package diodes

import (
	"time"
)

// Diode is any implementation of a diode
type Diode interface {
	Set(GenericDataType)
	TryNext() (GenericDataType, bool)
}

// Poller will poll a diode until a value is available
type Poller struct {
	Diode
	interval time.Duration
}

// PollerConfigOption can be used to setup the poller
type PollerConfigOption func(*Poller)

// WithPollingInterval sets the interval at which the diode is queried
// for new data. The default is 10ms.
func WithPollingInterval(interval time.Duration) PollerConfigOption {
	return PollerConfigOption(func(c *Poller) {
		c.interval = interval
	})
}

// NewPoller wraps a diode to allow accessing data via polling
func NewPoller(d Diode, opts ...PollerConfigOption) *Poller {
	p := &Poller{
		Diode:    d,
		interval: 10 * time.Millisecond,
	}

	for _, o := range opts {
		o(p)
	}

	return p
}

// Next polls the diode until data is available
func (p *Poller) Next() GenericDataType {
	for {
		data, ok := p.Diode.TryNext()
		if !ok {
			time.Sleep(p.interval)
			continue
		}
		return data
	}
}
