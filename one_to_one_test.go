package diodes_test

import (
	"code.cloudfoundry.org/go-diodes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("OneToOne", func() {
	var (
		d    *diodes.OneToOne
		data []byte

		spy *spyAlerter
	)

	BeforeEach(func() {
		spy = newSpyAlerter()

		d = diodes.NewOneToOne(5, spy)

		data = []byte("some-data")
		d.Set(diodes.GenericDataType(&data))
	})

	Context("multiple data slices", func() {
		var (
			secondData []byte
		)

		BeforeEach(func() {
			secondData = []byte("some-other-data")
			d.Set(diodes.GenericDataType(&secondData))
		})

		Describe("TryNext()", func() {
			It("returns true", func() {
				_, ok := d.TryNext()

				Expect(ok).To(BeTrue())
			})

			Context("reads exceed writes", func() {
				BeforeEach(func() {
					d.TryNext()
					d.TryNext()
				})

				It("returns false", func() {
					_, ok := d.TryNext()

					Expect(ok).To(BeFalse())
				})
			})
		})

		Context("buffer size exceeded", func() {
			BeforeEach(func() {
				for i := 0; i < 4; i++ {
					d.Set(diodes.GenericDataType(&secondData))
				}
			})

			It("wraps", func() {
				data, _ := d.TryNext()
				Expect(*(*[]byte)(data)).To(Equal(secondData))
			})

			It("alerts for each dropped point", func() {
				d.TryNext()
				Expect(spy.AlertInput.Missed).To(Receive(Equal(5)))
			})

			It("it updates the read index", func() {
				d.TryNext()
				Expect(spy.AlertInput.Missed).To(Receive(Equal(5)))

				for i := 0; i < 6; i++ {
					d.Set(diodes.GenericDataType(&secondData))
				}

				d.TryNext()
				Expect(spy.AlertInput.Missed).To(Receive(Equal(5)))
			})

			Context("read catches up with write", func() {
				BeforeEach(func() {
					d.TryNext()
					<-spy.AlertInput.Missed
				})

				It("does not alert", func() {
					d.TryNext()
					Expect(spy.AlertInput.Missed).To(Not(Receive()))
				})
			})

			Context("writer laps reader", func() {
				BeforeEach(func() {
					for i := 0; i < 5; i++ {
						d.Set(diodes.GenericDataType(&secondData))
					}
					d.TryNext()
				})

				It("sends an alert for each set", func() {
					Expect(spy.AlertInput.Missed).To(Receive(Equal(10)))
				})
			})

			Context("writer laps reader with nil alerter", func() {
				It("drops the alert", func() {
					d = diodes.NewOneToOne(5, nil)
					for i := 0; i < 10; i++ {
						d.Set(diodes.GenericDataType(&secondData))
					}

					Expect(func() {
						d.TryNext()
					}).ToNot(Panic())
				})
			})
		})
	})

})

var _ = Describe("reader ahead of writer", func() {
	It("must not occur after alerting", func() {
		length := 4
		spy := newSpyAlerter()
		d := diodes.NewOneToOne(length, spy)
		data := []byte("some-data")
		genData := diodes.GenericDataType(&data)

		By("filling up the buffer")
		for i := 0; i < length; i++ {
			d.Set(genData)
		}

		By("overwriting the first index")
		d.Set(genData)

		By("having reader fast forward")
		Expect(spy.AlertInput.Missed).To(BeEmpty())
		_, ok := d.TryNext()
		Expect(ok).To(BeTrue())
		Expect(spy.AlertInput.Missed).To(Receive(Equal(4)))

		By("failing reads until the writer writes over skipped values")
		_, ok = d.TryNext()
		Expect(ok).To(BeFalse())
		d.Set(genData)
		_, ok = d.TryNext()
		Expect(ok).To(BeTrue())
	})
})

type spyAlerter struct {
	AlertCalled chan bool
	AlertInput  struct {
		Missed chan int
	}
}

func newSpyAlerter() *spyAlerter {
	m := &spyAlerter{}
	m.AlertCalled = make(chan bool, 100)
	m.AlertInput.Missed = make(chan int, 100)
	return m
}
func (m *spyAlerter) Alert(missed int) {
	m.AlertCalled <- true
	m.AlertInput.Missed <- missed
}
