package diodes

import (
	"sync/atomic"
	"unsafe"
)

type bucketFreeList struct {
	head unsafe.Pointer
}

func (l *bucketFreeList) Add(b *bucket) {
	for {
		head := (*bucket)(atomic.LoadPointer(&l.head))
		b.next = head
		if atomic.CompareAndSwapPointer(&l.head, (unsafe.Pointer)(head), (unsafe.Pointer)(b)) {
			return
		}
	}
}

func (l *bucketFreeList) Get() *bucket {
	for {
		head := (*bucket)(atomic.LoadPointer(&l.head))
		if head == nil {
			return &bucket{}
		}
		if atomic.CompareAndSwapPointer(&l.head, (unsafe.Pointer)(head), (unsafe.Pointer)(head.next)) {
			return head
		}
	}
}
