package diode

import (
	"fmt"
	"testing"
)

func TestNew(t *testing.T) {
	t.Parallel()

	d := New[string](5)

	if size := len(d.buf); size != 5 {
		t.Errorf("New Diode buffer size = %d; want 5", size)
	}
	if d.readIdx != 0 {
		t.Errorf("New Diode readIdx = %d; want 0", d.readIdx)
	}
	var i uint64
	i = ^i
	if d.writeIdx != i {
		t.Errorf("New Diode writeIdx = %d; want %d", d.writeIdx, i)
	}
}

func TestSet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		d    *Diode[string]
	}{
		{name: "one writer", d: New[string](5)},
		{name: "many writers", d: New[string](5, WithManyWriters())},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var (
				oldReadIdx  = tc.d.readIdx
				oldWriteIdx = tc.d.writeIdx
			)

			data := "test"
			tc.d.Set(&data)

			if b := (*bucket[string])(tc.d.buf[tc.d.writeIdx]); *b.data != data || b.seq != tc.d.writeIdx {
				t.Errorf(`buf[0] = {data: %s, seq: %d}; want {data: %s, seq: %d}`, *b.data, b.seq, data, tc.d.writeIdx)
			}
			if tc.d.readIdx != oldReadIdx {
				t.Errorf("readIdx = %d; want %d", tc.d.readIdx, oldReadIdx)
			}
			if tc.d.writeIdx != oldWriteIdx+1 {
				t.Errorf("writeIdx = %d; want %d", tc.d.writeIdx, oldWriteIdx+1)
			}
		})
	}
}

func TestSet_WithCopy(t *testing.T) {
	t.Parallel()

	d := New[string](5, WithCopy())

	data := "test"
	d.Set(&data)

	if b := (*bucket[string])(d.buf[d.writeIdx]); b.data == &data {
		t.Errorf("data is not copied: buf[0].data == result == %+v", b.data)
	}
}

func TestTryNext(t *testing.T) {
	t.Parallel()

	type test struct {
		name    string
		preFill int
		readIdx uint64
		result  string
		ok      bool
	}

	tests := []test{
		{name: "simple", preFill: 1, readIdx: 1, result: "test-0", ok: true},
		{name: "empty", preFill: 0, readIdx: 0, result: "", ok: false},
		{name: "overwrite", preFill: 7, readIdx: 6, result: "test-5", ok: true},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			d := New[string](5)

			for i := 0; i < tc.preFill; i++ {
				s := fmt.Sprintf("test-%d", i)
				d.Set(&s)
			}

			result, ok := d.TryNext()

			if d.readIdx != tc.readIdx {
				t.Errorf("readIdx = %d; want %d", d.readIdx, tc.readIdx)
			}
			if ok != tc.ok {
				t.Errorf("TryNext() = _, %t; want _, %t", ok, tc.ok)
			}
			if (!ok && result != nil) || (ok && *result != tc.result) {
				t.Errorf("TryNext() = %s, _; want %s, _", *result, tc.result)
			}
		})
	}
}

func TestTryNext_Overwrite(t *testing.T) {
	t.Parallel()

	d := New[string](5)

	data := "test"
	d.Set(&data)

	result, _ := d.TryNext()
	bucket := (*bucket[string])(d.buf[0])
	if bucket != nil && result == bucket.data {
		t.Errorf("old data is not overwritten: result == buf[0].data")
	}
}

func BenchmarkSet(b *testing.B) {
	data := "test"

	tests := []struct {
		name            string
		withManyWriters bool
		withCopy        bool
	}{
		{name: "one writer with copy", withManyWriters: false, withCopy: true},
		{name: "one writer", withManyWriters: false, withCopy: false},
		{name: "many writers with copy", withManyWriters: true, withCopy: true},
		{name: "many writers", withManyWriters: true, withCopy: false},
	}

	for _, tc := range tests {
		tc := tc
		var opts []Option
		if tc.withManyWriters {
			opts = append(opts, WithManyWriters())
		}
		if tc.withCopy {
			opts = append(opts, WithCopy())
		}
		d := New[string](100, opts...)

		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				d.Set(&data)
			}
		})
	}
}

func BenchmarkTryNext(b *testing.B) {
	data := "test"

	for k := 10; k < 10001; k *= 10 {
		for j := 10; j < 10001; j *= 10 {
			b.Run(fmt.Sprintf("size %d pre-insert %d", k, j), func(b *testing.B) {
				d := New[string](k)

				for i := 0; i < j; i++ {
					d.Set(&data)
				}

				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					_, _ = d.TryNext()
				}
			})
		}
	}
}
