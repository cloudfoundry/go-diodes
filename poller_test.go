package diodes_test

import (
	"sync"
	"time"

	"github.com/cloudfoundry/diodes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Poller", func() {
	var (
		spy *spyDiode
		p   *diodes.Poller
	)

	BeforeEach(func() {
		spy = new(spyDiode)
		p = diodes.NewPoller(spy, diodes.WithPollingInterval(time.Millisecond))
	})

	It("returns the available result", func() {
		spy.dataList = [][]byte{[]byte("a"), []byte("b")}

		Expect(*(*[]byte)(p.Next())).To(Equal([]byte("a")))
		Expect(*(*[]byte)(p.Next())).To(Equal([]byte("b")))
	})

	It("polls the given diode until data is available", func() {
		go func() {
			time.Sleep(250 * time.Millisecond)
			spy.mu.Lock()
			defer spy.mu.Unlock()
			spy.dataList = [][]byte{[]byte("a")}
		}()

		Expect(*(*[]byte)(p.Next())).To(Equal([]byte("a")))
	})
})

type spyDiode struct {
	diodes.Diode
	mu       sync.Mutex
	dataList [][]byte
	called   int
}

func (s *spyDiode) TryNext() (diodes.GenericDataType, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.called++
	if len(s.dataList) == 0 {
		return nil, false
	}

	next := s.dataList[0]
	s.dataList = s.dataList[1:]
	return diodes.GenericDataType(&next), true
}
