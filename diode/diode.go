package diode

import (
	"context"
	"sync/atomic"
	"time"
	"unsafe"
)

type bucket[T any] struct {
	data *T     // The data stored in this bucket.
	seq  uint64 // The write index at the time of writing.
}

// Diode is a ring buffer.
type Diode[T any] struct {
	// The Diode's buffer; it's length determines size of the diode.
	buf []unsafe.Pointer

	// Indexes to write to or read from next.
	writeIdx uint64
	readIdx  uint64

	// Other options.
	opts options
}

// New creates a diode with the given size and options.
func New[T any](size int, opts ...Option) *Diode[T] {
	d := &Diode[T]{
		buf: make([]unsafe.Pointer, size),
	}

	for _, opt := range opts {
		opt.apply(&d.opts)
	}

	if d.opts.rep == nil {
		d.opts.rep = reporter{}
	}

	// Start write index at the value before 0
	// to allow the first write to use AddUint64
	// and still have a beginning index of 0
	d.writeIdx = ^d.writeIdx

	return d
}

// Set sets the data into the next slot of the ring buffer. This function is
// only thread-safe if the Diode was created WithManyWriters.
func (d *Diode[T]) Set(data *T) {
	setFunc := d.set
	if d.opts.safe {
		setFunc = d.safelySet
	}

	if d.opts.copy {
		copy := *data
		setFunc(&copy)
		return
	}

	setFunc(data)
}

// set sets the data quickly into the next slot of the ring buffer. This is not
// a thread-safe function.
func (d *Diode[T]) set(data *T) {
	d.writeIdx++
	idx := d.writeIdx % uint64(len(d.buf))
	b := &bucket[T]{
		data: data,
		seq:  d.writeIdx,
	}
	atomic.StorePointer(&d.buf[idx], unsafe.Pointer(b))
}

// safelySet will continue to try to set data into the next slot of the ring
// buffer until it succeeds.
func (d *Diode[T]) safelySet(data *T) {
	for {
		writeIdx := atomic.AddUint64(&d.writeIdx, 1)
		idx := writeIdx % uint64(len(d.buf))
		old := atomic.LoadPointer(&d.buf[idx])
		if old != nil &&
			(*bucket[T])(old) != nil &&
			(*bucket[T])(old).seq > writeIdx-uint64(len(d.buf)) {
			go d.opts.rep.Warn("diode set collision (consider using a larger diode)")
			continue
		}

		b := &bucket[T]{
			data: data,
			seq:  writeIdx,
		}
		if !atomic.CompareAndSwapPointer(&d.buf[idx], old, unsafe.Pointer(b)) {
			go d.opts.rep.Warn("diode set collision (consider using a larger diode)")
			continue
		}

		return
	}
}

// TryNext will attempt to read data from the next slot of the ring buffer. This
// method is not thread-safe.
func (d *Diode[T]) TryNext() (data *T, ok bool) {
	// Read a value from the ring buffer based on the readIndex.
	idx := d.readIdx % uint64(len(d.buf))
	result := (*bucket[T])(atomic.SwapPointer(&d.buf[idx], nil))

	// When the result is nil that means the writer has not had the
	// opportunity to write a value into the diode. This value must be ignored
	// and the read head must not increment.
	if result == nil {
		return nil, false
	}

	// When the seq value is less than the current read index that means a
	// value was read from idx that was previously written but has since has
	// been dropped. This value must be ignored and the read head must not
	// increment.
	//
	// The simulation for this scenario assumes the fast forward occurred as
	// detailed below.
	//
	// 5. The reader reads again getting seq 5. It then reads again expecting
	//    seq 6 but gets seq 2. This is a read of a stale value that was
	//    effectively "dropped" so the read fails and the read head stays put.
	//    `| 4 | 5 | 2 | 3 |` r: 7, w: 6
	//
	if result.seq < d.readIdx {
		return nil, false
	}

	// When the seq value is greater than the current read index that means a
	// value was read from idx that overwrote the value that was expected to
	// be at this idx. This happens when the writer has lapped the reader. The
	// reader needs to catch up to the writer so it moves its write head to
	// the new seq, effectively dropping the messages that were not read in
	// between the two values.
	//
	// Here is a simulation of this scenario:
	//
	// 1. Both the read and write heads start at 0.
	//    `| nil | nil | nil | nil |` r: 0, w: 0
	// 2. The writer fills the buffer.
	//    `| 0 | 1 | 2 | 3 |` r: 0, w: 4
	// 3. The writer laps the read head.
	//    `| 4 | 5 | 2 | 3 |` r: 0, w: 6
	// 4. The reader reads the first value, expecting a seq of 0 but reads 4,
	//    this forces the reader to fast forward to 5.
	//    `| 4 | 5 | 2 | 3 |` r: 5, w: 6
	//
	if result.seq > d.readIdx {
		go d.opts.rep.Alert(result.seq - d.readIdx)
		d.readIdx = result.seq
	}

	// Only increment read index if a regular read occurred (where seq was
	// equal to readIdx) or a value was read that caused a fast forward
	// (where seq was greater than readIdx).
	//
	d.readIdx++
	return (*T)(result.data), true
}

// Poll will attempt to read data from the next slot of the ring buffer on a set
// interval until it succeeds, or until the provided context is closed. This
// method is not thread-safe.
func (d *Diode[T]) Poll(ctx context.Context, interval time.Duration) (data *T, ok bool) {
	for {
		select {
		case <-ctx.Done():
			return nil, false
		default:
		}

		data, ok := d.TryNext()
		if !ok {
			time.Sleep(interval)
			continue
		}

		return data, true
	}
}
