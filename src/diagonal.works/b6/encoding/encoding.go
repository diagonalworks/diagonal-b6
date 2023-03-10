package encoding

import (
	"encoding/binary"
	"fmt"
)

// Offset represents a relative offset within a file or byte slice. Without
// additional context, it's assumed to be relative to the beginning of the
// file or slice. Used to differentiate between functions that return offsets
// of the end of encoded data, rather than the length of the encoded data.
type Offset int64

func (o Offset) Add(length int) Offset {
	return o + Offset(length)
}

func (o Offset) Difference(start Offset) int {
	return int(o - Offset(start))
}

type bufferWriter struct {
	Buffer []byte
	Pos    int
}

func (b *bufferWriter) Write(data []byte) (int, error) {
	if len(b.Buffer)-b.Pos < len(data) {
		panic(fmt.Sprintf("buffer to small: length %d, need %d", len(b.Buffer), b.Pos+len(data)))
	}
	copy(b.Buffer[b.Pos:], data)
	b.Pos += len(data)
	return len(data), nil
}

type bufferReader struct {
	Buffer []byte
	Pos    int
}

func (b *bufferReader) Read(data []byte) (int, error) {
	n := copy(data, b.Buffer)
	b.Pos += n
	return len(data), nil
}

func UnmarshalStruct(s interface{}, buffer []byte) int {
	r := bufferReader{Buffer: buffer, Pos: 0}
	if err := binary.Read(&r, binary.LittleEndian, s); err != nil {
		panic(err)
	}
	return r.Pos
}
