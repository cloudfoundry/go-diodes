package diode

import (
	"fmt"
	"testing"
)

func TestTryNext(t *testing.T) {
	d := New(5)

	s := "test"
	d.Set(GenericData(&s))

	gd, ok := d.TryNext()
	if !ok {
		t.Errorf("TryNext() = _, %t; want _, true", ok)
	}
	if d := *(*string)(gd); d != s {
		t.Errorf("TryNext() = %v, true; want %v, true", d, s)
	}
}

func TestTryNext_Empty(t *testing.T) {
	d := New(5)

	_, ok := d.TryNext()
	if ok {
		t.Errorf("TryNext() = _, %t; want _, false", ok)
	}
}

func TestTryNext_Overwrite(t *testing.T) {
	d := New(5)

	for i := range [7]int{} {
		s := fmt.Sprintf("test-%d", i)
		d.Set(GenericData(&s))
	}

	gd, ok := d.TryNext()
	if !ok {
		t.Errorf("TryNext() = _, %t; want _, true", ok)
	}
	if d := *(*string)(gd); d != "test-5" {
		t.Errorf("TryNext() = %v, true; want test-5, true", d)
	}
}
