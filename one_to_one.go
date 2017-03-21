package diodes

import (
	"sync/atomic"
	"time"
	"unsafe"
)

// GenericDataType is the data type the diodes operate on.
type GenericDataType unsafe.Pointer

// Alerter is used to report how many values were overwritten
// since the last write.
type Alerter interface {
	Alert(missed int)
}

// AlerFunc type is an adapter to allow the use of
// ordinary functions as Alert handlers.
type AlertFunc func(missed int)

// Alert calls f(missed)
func (f AlertFunc) Alert(missed int) {
	f(missed)
}

type bucket struct {
	data GenericDataType
	seq  uint64
}

// OneToOne diode is optimized for a single writer and a single reader.
type OneToOne struct {
	buffer     []unsafe.Pointer
	writeIndex uint64
	readIndex  uint64
	alerter    Alerter
}

// NewOneToOne creates a new diode (ring buffer). The OneToOne diode
// is optimzed for a single writer (on go-routine A) and a single reader
// (on go-routine B).
func NewOneToOne(size int, alerter Alerter) *OneToOne {
	d := &OneToOne{
		buffer:  make([]unsafe.Pointer, size),
		alerter: alerter,
	}

	// Start write index at the value before 0
	// to allow the first write to use AddUint64
	// and still have a beginning index of 0
	d.writeIndex = ^d.writeIndex
	return d
}

// Set sets the data in the next slot of the ring buffer.
func (d *OneToOne) Set(data GenericDataType) {
	writeIndex := atomic.AddUint64(&d.writeIndex, 1)
	idx := writeIndex % uint64(len(d.buffer))
	newBucket := &bucket{
		data: data,
		seq:  writeIndex,
	}

	atomic.StorePointer(&d.buffer[idx], unsafe.Pointer(newBucket))
}

// TryNext will attempt to read from the next slot of the ring buffer.
// If there is not data available, it will return (nil, false).
func (d *OneToOne) TryNext() (data GenericDataType, ok bool) {
	readIndex := atomic.LoadUint64(&d.readIndex)
	idx := readIndex % uint64(len(d.buffer))

	value, ok := d.tryNext(idx)
	if ok {
		atomic.AddUint64(&d.readIndex, 1)
	}
	return value, ok
}

// Next will poll the ring buffer (at 10ms intervals) until data is available.
func (d *OneToOne) Next() GenericDataType {
	readIndex := atomic.LoadUint64(&d.readIndex)
	idx := readIndex % uint64(len(d.buffer))

	result := d.pollBuffer(idx)
	atomic.AddUint64(&d.readIndex, 1)

	return result
}

func (d *OneToOne) tryNext(idx uint64) (GenericDataType, bool) {
	result := (*bucket)(atomic.SwapPointer(&d.buffer[idx], nil))

	if result == nil {
		return nil, false
	}

	if result.seq > d.readIndex {
		d.alerter.Alert(int(result.seq - d.readIndex))
		atomic.StoreUint64(&d.readIndex, result.seq)
	}

	return result.data, true
}

func (d *OneToOne) pollBuffer(idx uint64) GenericDataType {
	for {
		result, ok := d.tryNext(idx)
		if !ok {
			time.Sleep(time.Millisecond * 10)
			continue
		}
		return result
	}
}
