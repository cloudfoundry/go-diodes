package diodes_test

import (
	"context"
	"sync"
	"time"

	"code.cloudfoundry.org/go-diodes"

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

	It("cancels Next() with context", func() {
		ctx, cancel := context.WithCancel(context.Background())
		p = diodes.NewPoller(spy, diodes.WithPollingContext(ctx))
		cancel()
		done := make(chan struct{})
		go func() {
			defer close(done)
			p.Next()
		}()

		Eventually(done).Should(BeClosed())
	})
})

type spyDiode struct {
	diodes.Diode
	mu       sync.Mutex
	dataList [][]byte
	called   int
}

func (s *spyDiode) Set(data diodes.GenericDataType) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.dataList = append(s.dataList, *(*[]byte)(data))
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
