package diodes_test

import (
	"github.com/cloudfoundry/diodes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ManyToOne", func() {
	var (
		d    *diodes.ManyToOne
		data []byte

		mockAlerter *mockAlerter
	)

	BeforeEach(func() {
		mockAlerter = newMockAlerter()

		d = diodes.NewManyToOne(5, mockAlerter)

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
				Expect(mockAlerter.AlertInput.Missed).To(Receive(Equal(5)))
			})

			It("it updates the read index", func() {
				d.TryNext()
				Expect(mockAlerter.AlertInput.Missed).To(Receive(Equal(5)))

				for i := 0; i < 6; i++ {
					d.Set(diodes.GenericDataType(&secondData))
				}

				d.TryNext()
				Expect(mockAlerter.AlertInput.Missed).To(Receive(Equal(5)))
			})

			Context("read catches up with write", func() {
				BeforeEach(func() {
					d.TryNext()
					<-mockAlerter.AlertInput.Missed
				})

				It("does not alert", func() {
					d.TryNext()
					Expect(mockAlerter.AlertInput.Missed).To(Not(Receive()))
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
					Expect(mockAlerter.AlertInput.Missed).To(Receive(Equal(10)))
				})
			})
		})
	})
})
