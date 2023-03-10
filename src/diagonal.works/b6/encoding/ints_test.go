package encoding

import (
	"reflect"
	"testing"
)

func TestUint64(t *testing.T) {
	tests := []struct {
		v uint64
		l int
	}{
		{0, 1},
		{255, 1},
		{256, 2},
		{66000, 3},
		{17000000, 4},
		{1 << 63, 8},
	}

	for _, test := range tests {
		l := Uint64Length(test.v)
		if l != test.l {
			t.Errorf("Expected length %d, found %d for %d", test.l, test.v, l)
		}
		buffer := make([]byte, l)
		MarshalUint64(test.v, l, buffer[0:])
		if v := UnmarshalUint64(l, buffer[0:]); v != test.v {
			t.Errorf("Expected to unmarshal %d, found %d", test.v, v)
		}
	}
}

func TestDeltaEncodedUint64s(t *testing.T) {
	e := []uint64{42, 36, 1038, 96, 21}
	var buffer [128]byte
	n := MarshalDeltaCodedUint64s(e, buffer[0:])

	ee := make([]uint64, 0, len(e))
	ee, nn := UnmarshalDeltaCodedUint64(ee, len(e), buffer[0:n])

	if n != nn {
		t.Errorf("Expected marshalled unmarshalled lengths to be equal")
	}

	if !reflect.DeepEqual(e, ee) {
		t.Errorf("Expected value %v, found: %v", e, ee)
	}

	if n >= 2*len(e) {
		t.Errorf("Expected less than 2 bytes per element, found total length: %d", n)
	}
}

func TestDeltaEncodedInts(t *testing.T) {
	examples := [][]int{
		{42, 36, 1038, 96, 21},
		{1, (1 << 62) - 2, 0, 1<<62 - 1},
	}

	for _, e := range examples {
		var buffer [128]byte
		n := MarshalDeltaCodedInts(e, buffer[0:])

		ee := make([]int, 0, len(e))
		ee, nn := UnmarshalDeltaCodedInts(ee, len(e), buffer[0:n])

		if n != nn {
			t.Errorf("Expected marshalled unmarshalled lengths to be equal")
		}

		if !reflect.DeepEqual(e, ee) {
			t.Errorf("Expected value %v, found: %v", e, ee)
		}
	}
}

type exampleHeader struct {
	Magic  int8
	Offset uint64
}

func TestStruct(t *testing.T) {
	h := exampleHeader{Magic: -42, Offset: 36}
	var buffer [16]byte
	n := MarshalStruct(h, buffer[0:])

	var hh exampleHeader
	nn := UnmarshalStruct(&hh, buffer[0:n])

	if !reflect.DeepEqual(h, hh) {
		t.Errorf("Expected value %+v, found: %+v", h, hh)
	}

	if n != nn {
		t.Errorf("Expected marshalled unmarshalled lengths to be equal")
	}
}
