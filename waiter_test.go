package diodes_test

import (
	"github.com/cloudfoundry/diodes"
	"time"

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
})
