package encoding

import (
	"encoding/binary"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
)

type ByteArraysLayout struct {
	Items         int
	OffsetBytes   int
	MaxItemLength int
}

const ByteArraysLayoutLength = 12

func (b *ByteArraysLayout) Marshal(buffer []byte) int {
	binary.LittleEndian.PutUint32(buffer[0:], uint32(b.Items))
	binary.LittleEndian.PutUint32(buffer[4:], uint32(b.OffsetBytes))
	binary.LittleEndian.PutUint32(buffer[8:], uint32(b.MaxItemLength))
	return ByteArraysLayoutLength
}

func (b *ByteArraysLayout) Unmarshal(buffer []byte) int {
	b.Items = int(binary.LittleEndian.Uint32(buffer[0:]))
	b.OffsetBytes = int(binary.LittleEndian.Uint32(buffer[4:]))
	b.MaxItemLength = int(binary.LittleEndian.Uint32(buffer[8:]))
	return ByteArraysLayoutLength
}

func (b *ByteArraysLayout) PointersLength() int {
	return b.OffsetBytes * (b.Items + 1)
}

func (b *ByteArraysLayout) PointerOffset(i int) Offset {
	return Offset(ByteArraysLayoutLength + (b.OffsetBytes * i))
}

func (b *ByteArraysLayout) DataOffset() Offset {
	return Offset(ByteArraysLayoutLength + b.PointersLength())
}

type ByteArraysBuilder struct {
	pointers     []uint64
	layout       ByteArraysLayout
	reservations uint64
	offset       Offset
	lock         sync.Mutex
}

func NewByteArraysBuilder(l int) *ByteArraysBuilder {
	return &ByteArraysBuilder{
		pointers: make([]uint64, l+1),
		layout: ByteArraysLayout{
			Items: -1,
		},
		offset: -1,
	}
}

func (b *ByteArraysBuilder) Reserve(i int, length int) {
	if b.layout.Items >= 0 {
		panic("Reserve() called after FinishReservation")
	}
	if length < 0 {
		panic("Reserve() with negative length")
	}
	if i >= len(b.pointers)-1 {
		panic(fmt.Sprintf("item %d out of range with length %d", i, len(b.pointers)-1))
	}
	atomic.AddUint64(&b.reservations, 1)
	atomic.AddUint64(&b.pointers[i], uint64(length))
}

func (b *ByteArraysBuilder) IsEmpty() bool {
	return b.reservations == 0
}

func (b *ByteArraysBuilder) FinishReservation() {
	b.layout.Items = len(b.pointers) - 1
	pointer := uint64(0)
	for i, reserved := range b.pointers {
		if b.layout.MaxItemLength < int(reserved) {
			b.layout.MaxItemLength = int(reserved)
		}
		b.pointers[i] = pointer
		pointer += reserved
	}
	b.layout.OffsetBytes = Uint64Length(pointer)
}

func (b *ByteArraysBuilder) WriteHeader(w io.WriterAt, offset Offset) (Offset, error) {
	if b.offset >= 0 {
		panic("Can't call WriteHeader() twice")
	}
	b.offset = offset
	if b.layout.Items < 0 {
		b.FinishReservation()
	}

	var buffer [ByteArraysLayoutLength]byte
	n := b.layout.Marshal(buffer[0:])
	if _, err := w.WriteAt(buffer[0:n], int64(offset)); err != nil {
		return 0, fmt.Errorf("Failed to write ByteArraysLayout: %s", err)
	}
	offset = offset.Add(n)

	for _, pointer := range b.pointers {
		MarshalUint64(pointer, b.layout.OffsetBytes, buffer[0:])
		if _, err := w.WriteAt(buffer[0:b.layout.OffsetBytes], int64(offset)); err != nil {
			return 0, fmt.Errorf("Failed to write pointer: %s", err)
		}
		offset = offset.Add(b.layout.OffsetBytes)
	}
	return offset.Add(int(b.pointers[len(b.pointers)-1])), nil
}

func (b *ByteArraysBuilder) WriteItem(w io.WriterAt, i int, buffers ...[]byte) error {
	if b.offset < 0 {
		panic("must call WriteHeader() before WriteItem()")
	}
	if i >= len(b.pointers)-1 {
		panic(fmt.Sprintf("item %d out of range with length %d", i, len(b.pointers)-1))
	}

	length := 0
	for _, buffer := range buffers {
		length += len(buffer)
	}

	b.lock.Lock()
	if b.pointers[i]+uint64(length) > b.pointers[i+1] {
		b.lock.Unlock()
		panic(fmt.Sprintf("Write beyond reserved space, item %d pointer %d, next %d, length %d", i, b.pointers[i], b.pointers[i+1], length))
	}
	offset := b.layout.DataOffset().Add(int(b.pointers[i]))
	b.pointers[i] += uint64(length)
	b.lock.Unlock()
	for _, buffer := range buffers {
		if _, err := w.WriteAt(buffer[0:], int64(b.offset+offset)); err != nil {
			return fmt.Errorf("Failed to write map entry: %w", err)
		}
		offset = offset.Add(len(buffer))
	}
	return nil
}

func (b *ByteArraysBuilder) Length() int {
	if b.layout.Items < 0 {
		panic("must call FinishReservation() or WriteHeader() before Length()")
	}
	return int(b.layout.DataOffset()) + int(b.pointers[len(b.pointers)-1])
}

type ByteArrays struct {
	Layout ByteArraysLayout
	data   []byte
}

func NewByteArrays(data []byte) *ByteArrays {
	b := &ByteArrays{data: data}
	b.Layout.Unmarshal(data)
	return b
}

func (b *ByteArrays) NumItems() int {
	return b.Layout.Items
}

func (b *ByteArrays) MaxItemLength() int {
	return b.Layout.MaxItemLength
}

func (b *ByteArrays) Length() int {
	offset := b.Layout.PointerOffset(b.Layout.Items)
	pointer := UnmarshalUint64(b.Layout.OffsetBytes, b.data[offset:])
	return ByteArraysLayoutLength + b.Layout.PointersLength() + int(pointer)
}

func (b *ByteArrays) Item(i int) []byte {
	if i >= b.Layout.Items {
		panic(fmt.Sprintf("item %d out of range with length %d", i, b.Layout.Items))
	}
	pointer := UnmarshalUint64(b.Layout.OffsetBytes, b.data[b.Layout.PointerOffset(i):])
	next := UnmarshalUint64(b.Layout.OffsetBytes, b.data[b.Layout.PointerOffset(i+1):])
	offset := b.Layout.DataOffset()
	return b.data[offset+Offset(pointer) : offset+Offset(next)]
}
