package diodes

import (
	"sync/atomic"
	"unsafe"
)

// GenericDataType is the data type the diodes operate on.
type GenericDataType unsafe.Pointer

// Alerter is used to report how many values were overwritten since the
// last write.
type Alerter interface {
	Alert(missed int)
}

// AlerFunc type is an adapter to allow the use of ordinary functions as
// Alert handlers.
type AlertFunc func(missed int)

// Alert calls f(missed)
func (f AlertFunc) Alert(missed int) {
	f(missed)
}

type bucket struct {
	data GenericDataType
	seq  uint64
}

// OneToOne diode is meant to be used by a single reader and a single writer.
// It is not thread safe if used otherwise.
type OneToOne struct {
	buffer     []unsafe.Pointer
	writeIndex uint64
	readIndex  uint64
	alerter    Alerter
}

// NewOneToOne creates a new diode is meant to be used by a single reader and
// a single writer.
func NewOneToOne(size int, alerter Alerter) *OneToOne {
	return &OneToOne{
		buffer:  make([]unsafe.Pointer, size),
		alerter: alerter,
	}
}

// Set sets the data in the next slot of the ring buffer.
func (d *OneToOne) Set(data GenericDataType) {
	idx := d.writeIndex % uint64(len(d.buffer))

	newBucket := &bucket{
		data: data,
		seq:  d.writeIndex,
	}
	d.writeIndex++

	atomic.StorePointer(&d.buffer[idx], unsafe.Pointer(newBucket))
}

// TryNext will attempt to read from the next slot of the ring buffer.
// If there is no data available, it will return (nil, false).
func (d *OneToOne) TryNext() (data GenericDataType, ok bool) {
	idx := d.readIndex % uint64(len(d.buffer))
	result := (*bucket)(atomic.SwapPointer(&d.buffer[idx], nil))

	if result == nil {
		return nil, false
	}
	if result.seq < d.readIndex {
		return nil, false
	}

	if result.seq > d.readIndex {
		dropped := result.seq - d.readIndex
		d.readIndex = result.seq
		d.alerter.Alert(int(dropped))
	}

	d.readIndex++
	return result.data, true
}
