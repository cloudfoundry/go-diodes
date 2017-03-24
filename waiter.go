package diodes

import "sync"

type Waiter struct {
	Diode
	mu sync.Mutex
	c  *sync.Cond
}

func NewWaiter(d Diode) *Waiter {
	w := new(Waiter)
	w.Diode = d
	w.c = sync.NewCond(&w.mu)

	return w
}

func (w *Waiter) Set(data GenericDataType) {
	w.Diode.Set(data)
	w.c.Broadcast()
}

func (w *Waiter) Next() GenericDataType {
	w.mu.Lock()
	defer w.mu.Unlock()

	for {
		data, ok := w.Diode.TryNext()
		if !ok {
			w.c.Wait()
			continue
		}
		return data
	}
}
