package diode

import (
	"fmt"
	"testing"
)

func TestTryNext(t *testing.T) {
	d := New[string](5)

	s := "test"
	d.Set(&s)

	str, ok := d.TryNext()
	if !ok {
		t.Errorf("TryNext() = _, %t; want _, true", ok)
	}
	if *str != s {
		t.Errorf("TryNext() = %v, true; want %v, true", *str, s)
	}
}

func TestTryNext_Empty(t *testing.T) {
	d := New[string](5)

	val, ok := d.TryNext()
	if ok {
		t.Errorf("TryNext() = _, %t; want _, false", ok)
	}
	if val != nil {
		t.Errorf("TryNext() = %+v, false; want nil, false", val)
	}
}

func TestTryNext_Overwrite(t *testing.T) {
	d := New[string](5)

	for i := range [7]int{} {
		s := fmt.Sprintf("test-%d", i)
		d.Set(&s)
	}

	str, ok := d.TryNext()
	if !ok {
		t.Errorf("TryNext() = _, %t; want _, true", ok)
	}
	if *str != "test-5" {
		t.Errorf("TryNext() = %v, true; want test-5, true", *str)
	}
}
