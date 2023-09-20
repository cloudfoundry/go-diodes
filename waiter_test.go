package diodes_test

import (
	"context"
	"time"

	"code.cloudfoundry.org/go-diodes"

	. "github.com/onsi/ginkgo/v2"
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

	Describe("Next", func() {
		BeforeEach(func() {
			spy.dataList = [][]byte{[]byte("a"), []byte("b")}
		})

		It("returns available data points from the wrapped diode", func() {
			Expect(spy.called).To(Equal(0))
			Expect(*(*[]byte)(w.Next())).To(Equal([]byte("a")))
			Expect(spy.called).To(Equal(1))
			Expect(*(*[]byte)(w.Next())).To(Equal([]byte("b")))
			Expect(spy.called).To(Equal(2))
		})

		Context("when there is no new data", func() {
			BeforeEach(func() {
				spy.dataList = nil
			})

			It("waits for Set to be called", func() {
				go func() {
					time.Sleep(250 * time.Millisecond)
					data := []byte("c")
					w.Set(diodes.GenericDataType(&data))
				}()
				Expect(spy.called).To(Equal(0))
				Expect(*(*[]byte)(w.Next())).To(Equal([]byte("c")))
				Expect(spy.called).To(Equal(2)) // Calls TryNext twice during wait loop
			})

			Context("when the context is cancelled", func() {
				var cancel context.CancelFunc

				BeforeEach(func() {
					var ctx context.Context
					ctx, cancel = context.WithCancel(context.Background())
					w = diodes.NewWaiter(spy, diodes.WithWaiterContext(ctx))
				})

				Context("beforehand", func() {
					It("returns nil", func() {
						cancel()
						Expect(spy.called).To(Equal(0))
						Expect(w.Next() == nil).To(BeTrue())
						Expect(spy.called).To(Equal(1))
					})
				})

				Context("while waiting", func() {
					It("returns nil", func() {
						go func() {
							time.Sleep(250 * time.Millisecond)
							cancel()
						}()
						Expect(spy.called).To(Equal(0))
						Expect(w.Next() == nil).To(BeTrue())
						Expect(spy.called).To(Equal(1))
					})
				})
			})
		})
	})
})
