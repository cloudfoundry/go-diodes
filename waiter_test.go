package diodes_test

import (
	"context"
	"time"

	"code.cloudfoundry.org/go-diodes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Waiter", func() {
	var (
		spy *spyDiode
		w   *diodes.Waiter
	)

	BeforeEach(func() {
		spy = &spyDiode{}
		w = diodes.NewWaiter(spy)
	})

	It("returns the available result", func() {
		spy.dataList = [][]byte{[]byte("a"), []byte("b")}

		Expect(*(*[]byte)(w.Next())).To(Equal([]byte("a")))
		Expect(*(*[]byte)(w.Next())).To(Equal([]byte("b")))
	})

	It("waits until data is available", func() {
		go func() {
			time.Sleep(250 * time.Millisecond)
			data := []byte("a")
			w.Set(diodes.GenericDataType(&data))
		}()

		Expect(*(*[]byte)(w.Next())).To(Equal([]byte("a")))
	})

	It("cancels Next() with context", func() {
		ctx, cancel := context.WithCancel(context.Background())
		w = diodes.NewWaiter(spy, diodes.WithWaiterContext(ctx))
		cancel()
		done := make(chan struct{})
		go func() {
			defer close(done)
			w.Next()
		}()

		Eventually(done).Should(BeClosed())
	})

	It("cancels current Next() with context", func() {
		ctx, cancel := context.WithCancel(context.Background())
		w = diodes.NewWaiter(spy, diodes.WithWaiterContext(ctx))
		go func() {
			time.Sleep(100 * time.Millisecond)
			cancel()
		}()

		Expect(w.Next() == nil).To(BeTrue())
	})
})
