package diodes_test

import (
	"sync"

	"github.com/cloudfoundry/diodes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("OneToOne", func() {
	var (
		d    *diodes.OneToOne
		data []byte

		mockAlerter *mockAlerter
	)

	Describe("Next()", func() {
		BeforeEach(func() {
			mockAlerter = newMockAlerter()

			d = diodes.NewOneToOne(5, mockAlerter)

			data = []byte("some-data")
			d.Set(diodes.GenericDataType(&data))
		})

		It("returns the next data slice", func() {
			Expect(*(*[]byte)(d.Next())).To(Equal(data))
		})

		Context("multiple data slices", func() {
			var (
				secondData []byte
			)

			BeforeEach(func() {
				secondData = []byte("some-other-data")
				d.Set(diodes.GenericDataType(&secondData))
			})

			It("returns data slices in order", func() {
				Expect(*(*[]byte)(d.Next())).To(Equal(data))
				Expect(*(*[]byte)(d.Next())).To(Equal(secondData))
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

			Context("reads exceed writes", func() {
				var (
					rxCh chan []byte
					wg   sync.WaitGroup
				)

				var waitForNext = func() {
					defer wg.Done()
					rxCh <- *(*[]byte)(d.Next())
				}

				BeforeEach(func() {
					rxCh = make(chan []byte, 100)
					for i := 0; i < 2; i++ {
						d.Next()
					}
					wg.Add(1)
					go waitForNext()
				})

				AfterEach(func() {
					wg.Wait()
				})

				It("blocks until data is available", func() {
					Consistently(rxCh).Should(HaveLen(0))
					d.Set(diodes.GenericDataType(&data))
					Eventually(rxCh).Should(HaveLen(1))
				})
			})

			Context("buffer size exceeded", func() {
				BeforeEach(func() {
					for i := 0; i < 4; i++ {
						d.Set(diodes.GenericDataType(&secondData))
					}
				})

				It("wraps", func() {
					Expect(*(*[]byte)(d.Next())).To(Equal(secondData))
				})

				It("alerts for each dropped point", func() {
					d.Next()
					Expect(mockAlerter.AlertInput.Missed).To(Receive(Equal(5)))
				})

				It("it updates the read index", func() {
					d.Next()
					Expect(mockAlerter.AlertInput.Missed).To(Receive(Equal(5)))

					for i := 0; i < 6; i++ {
						d.Set(diodes.GenericDataType(&secondData))
					}

					d.Next()
					Expect(mockAlerter.AlertInput.Missed).To(Receive(Equal(5)))
				})

				Context("read catches up with write", func() {
					BeforeEach(func() {
						d.Next()
						<-mockAlerter.AlertInput.Missed
					})

					It("does not alert", func() {
						d.Next()
						Expect(mockAlerter.AlertInput.Missed).To(Not(Receive()))
					})
				})

				Context("writer laps reader", func() {
					BeforeEach(func() {
						for i := 0; i < 5; i++ {
							d.Set(diodes.GenericDataType(&secondData))
						}
						d.Next()
					})

					It("sends an alert for each set", func() {
						Expect(mockAlerter.AlertInput.Missed).To(Receive(Equal(10)))
					})
				})
			})
		})
	})
})

type mockAlerter struct {
	AlertCalled chan bool
	AlertInput  struct {
		Missed chan int
	}
}

func newMockAlerter() *mockAlerter {
	m := &mockAlerter{}
	m.AlertCalled = make(chan bool, 100)
	m.AlertInput.Missed = make(chan int, 100)
	return m
}
func (m *mockAlerter) Alert(missed int) {
	m.AlertCalled <- true
	m.AlertInput.Missed <- missed
}
