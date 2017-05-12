package diodes

import (
	"log"
	"sync/atomic"
	"unsafe"
)

// ManyToOne diode is optimal for many writers (go-routines B-n) and a single
// reader (go-routine A). It is not thread safe for multiple readers.
type ManyToOne struct {
	buffer     []unsafe.Pointer
	writeIndex uint64
	readIndex  uint64
	alerter    Alerter
}

// NewManyToOne creates a new diode (ring buffer). The ManyToOne diode
// is optimzed for many writers (on go-routines B-n) and a single reader
// (on go-routine A).
func NewManyToOne(size int, alerter Alerter) *ManyToOne {
	d := &ManyToOne{
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
func (d *ManyToOne) Set(data GenericDataType) {
	for {
		writeIndex := atomic.AddUint64(&d.writeIndex, 1)
		idx := writeIndex % uint64(len(d.buffer))
		old := atomic.LoadPointer(&d.buffer[idx])

		if old != nil &&
			(*bucket)(old) != nil &&
			(*bucket)(old).seq > writeIndex-uint64(len(d.buffer)) {
			log.Println("Diode set collision: consider using a larger diode")
			continue
		}

		newBucket := &bucket{
			data: data,
			seq:  writeIndex,
		}

		if !atomic.CompareAndSwapPointer(&d.buffer[idx], old, unsafe.Pointer(newBucket)) {
			log.Println("Diode set collision: consider using a larger diode")
			continue
		}

		return
	}
}

// TryNext will attempt to read from the next slot of the ring buffer.
// If there is not data available, it will return (nil, false).
func (d *ManyToOne) TryNext() (data GenericDataType, ok bool) {
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
