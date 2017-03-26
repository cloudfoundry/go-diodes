package diodes_test

import (
	"io/ioutil"
	"log"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/cloudfoundry/diodes"
)

func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)

	os.Exit(m.Run())
}

func BenchmarkOneToOnePoller(b *testing.B) {
	d := diodes.NewPoller(diodes.NewOneToOne(b.N, diodes.AlertFunc(func(missed int) {
		panic("Oops...")
	})))

	go func() {
		for i := 0; i < b.N; i++ {
			data := []byte("some-data")
			d.Set(diodes.GenericDataType(&data))
		}
	}()

	for i := 0; i < b.N; i++ {
		d.Next()
	}
}

func BenchmarkOneToOneWaiter(b *testing.B) {
	d := diodes.NewWaiter(diodes.NewOneToOne(b.N, diodes.AlertFunc(func(missed int) {
		panic("Oops...")
	})))

	var wg sync.WaitGroup
	wg.Add(1)
	defer wg.Wait()

	go func() {
		defer wg.Done()
		for i := 0; i < b.N; i++ {
			data := []byte("some-data")
			d.Set(diodes.GenericDataType(&data))
		}
	}()

	for i := 0; i < b.N; i++ {
		d.Next()
	}
}

func BenchmarkManyToOnePoller(b *testing.B) {
	d := diodes.NewPoller(diodes.NewManyToOne(b.N, diodes.AlertFunc(func(missed int) {
		panic("Oops...")
	})))

	go func() {
		for i := 0; i < b.N; i++ {
			data := []byte("some-data")
			d.Set(diodes.GenericDataType(&data))
		}
	}()

	for i := 0; i < b.N; i++ {
		d.Next()
	}
}

func BenchmarkManyToOneWaiter(b *testing.B) {
	d := diodes.NewWaiter(diodes.NewManyToOne(b.N, diodes.AlertFunc(func(missed int) {
		panic("Oops...")
	})))

	var wg sync.WaitGroup
	wg.Add(1)
	defer wg.Wait()

	go func() {
		defer wg.Done()
		for i := 0; i < b.N; i++ {
			data := []byte("some-data")
			d.Set(diodes.GenericDataType(&data))
		}
	}()

	for i := 0; i < b.N; i++ {
		d.Next()
	}
}

func BenchmarkChannel(b *testing.B) {
	c := make(chan []byte, b.N)

	go func() {
		for i := 0; i < b.N; i++ {
			data := []byte("some-data")
			c <- data
		}
	}()

	for i := 0; i < b.N; i++ {
		<-c
	}
}

func BenchmarkOneToOnePollerDrain(b *testing.B) {
	d := diodes.NewPoller(diodes.NewOneToOne(100, diodes.AlertFunc(func(missed int) {
		// NOP
	})))

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		wg.Done()
		for {
			d.Next()
		}
	}()

	wg.Wait()
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		data := []byte("some-data")
		d.Set(diodes.GenericDataType(&data))
	}
}

func BenchmarkOneToOneWaiterDrain(b *testing.B) {
	d := diodes.NewWaiter(diodes.NewOneToOne(100, diodes.AlertFunc(func(missed int) {
		// NOP
	})))

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		wg.Done()
		for {
			d.Next()
		}
	}()

	wg.Wait()
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		data := []byte("some-data")
		d.Set(diodes.GenericDataType(&data))
	}
}

func BenchmarkChannelDrain(b *testing.B) {
	c := make(chan []byte, 100)
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		wg.Done()
		for range c {
		}
	}()

	wg.Wait()
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		data := []byte("some-data")
		select {
		case c <- data:
		default:
			drainChannel(c)
		}
	}
}

func BenchmarkManyWritersDiode(b *testing.B) {
	d := diodes.NewWaiter(diodes.NewManyToOne(10000, diodes.AlertFunc(func(int) {
		// NOP
	})))

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		wg.Done()
		for {
			d.Next()
			time.Sleep(100 * time.Millisecond)
		}
	}()

	wg.Wait()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			data := []byte("some-data")
			d.Set(diodes.GenericDataType(&data))
		}
	})
}

func BenchmarkManyWritersChannel(b *testing.B) {
	c := make(chan []byte, 10000)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		wg.Done()
		for range c {
			time.Sleep(100 * time.Millisecond)
		}
	}()

	wg.Wait()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			data := []byte("some-data")
			select {
			case c <- data:
			default:
				drainChannel(c)
			}
		}
	})
}

func drainChannel(c chan []byte) {
	for {
		select {
		case <-c:
		default:
			return
		}
	}
}
