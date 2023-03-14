package encoding

import (
	"encoding/binary"
)

func Uint64Length(v uint64) int {
	if v&0xffffffffffffff00 == 0 {
		return 1
	} else if v&0xffffffffffff0000 == 0 {
		return 2
	} else if v&0xffffffffff000000 == 0 {
		return 3
	} else if v&0xffffffff00000000 == 0 {
		return 4
	} else if v&0xffffff0000000000 == 0 {
		return 5
	} else if v&0xffff000000000000 == 0 {
		return 6
	} else if v&0xff00000000000000 == 0 {
		return 7
	}
	return 8
}

func MarshalUint64(v uint64, l int, buffer []byte) {
	for i := 0; i < l; i++ {
		buffer[i] = byte(v & 0xff)
		v >>= 8
	}
}

func UnmarshalUint64(l int, buffer []byte) uint64 {
	var v uint64
	i := l - 1
	for {
		v |= uint64(buffer[i])
		if i == 0 {
			break
		}
		i--
		v <<= 8
	}
	return v
}

type Uint64Slice []uint64

func (u Uint64Slice) Len() int           { return len(u) }
func (u Uint64Slice) Swap(i, j int)      { u[i], u[j] = u[j], u[i] }
func (u Uint64Slice) Less(i, j int) bool { return u[i] < u[j] }

// MarshalDeltaCodedUint64s writes delta coded varints
// to buffer. It does not add a length.
func MarshalDeltaCodedUint64s(vs []uint64, buffer []byte) int {
	i := 0
	last := int64(0)
	for _, v := range vs {
		i += binary.PutUvarint(buffer[i:], ZigzagEncode(int64(v)-last))
		last = int64(v)
	}
	return i
}

// UnmarshalDeltaCodedUint64 fills vs with n delta coded uvarints
// from buffer
func UnmarshalDeltaCodedUint64(vs []uint64, n int, buffer []byte) ([]uint64, int) {
	last := int64(0)
	vs = vs[0:0]
	i := 0
	for j := 0; j < int(n); j++ {
		v, n := binary.Uvarint(buffer[i:])
		i += n
		last += int64(ZigzagDecode(v))
		vs = append(vs, uint64(last))
	}
	return vs, i
}

func ZigzagEncode(value int64) uint64 {
	return uint64(value<<1) ^ uint64(value>>63)
}

func ZigzagDecode(value uint64) int64 {
	return (int64(value) >> 1) ^ (-(int64(value) & 1))
}

// MarshalDeltaCodedInts writes delta, zigzag, coded varints
// to buffer. It does not add a length.
func MarshalDeltaCodedInts(vs []int, buffer []byte) int {
	i := 0
	last := 0
	for _, v := range vs {
		i += binary.PutUvarint(buffer[i:], ZigzagEncode(int64(v-last)))
		last = v
	}
	return i
}

// UnmarshalDeltaCodedInts fills vs with n delta, zigzag, coded uvarints
// from buffer
func UnmarshalDeltaCodedInts(vs []int, n int, buffer []byte) ([]int, int) {
	last := 0
	vs = vs[0:0]
	i := 0
	for j := 0; j < int(n); j++ {
		v, n := binary.Uvarint(buffer[i:])
		i += n
		last += int(ZigzagDecode(v))
		vs = append(vs, last)
	}
	return vs, i
}

func MarshalStruct(s interface{}, buffer []byte) int {
	w := bufferWriter{Buffer: buffer, Pos: 0}
	if err := binary.Write(&w, binary.LittleEndian, s); err != nil {
		panic(err)
	}
	return w.Pos
}

func MarshalledSize(s interface{}) int {
	return binary.Size(s)
}
