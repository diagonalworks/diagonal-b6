package encoding

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"sync"
	"syscall"
	"time"

	"github.com/apache/beam/sdks/go/pkg/beam/io/filesystem"
)

const MaxUint32 = (1 << 32) - 1

type Mmapped struct {
	Data []byte
}

func Mmap(filename string) (Mmapped, error) {
	var m Mmapped
	f, err := os.Open(filename)
	if err == nil {
		stat, err := f.Stat()
		if err == nil {
			m.Data, err = syscall.Mmap(int(f.Fd()), 0, int(stat.Size()), syscall.PROT_READ, syscall.MAP_SHARED)
		}
	}
	if err != nil {
		err = fmt.Errorf("can't mmap %s: %s", filename, err)
	}
	return m, err
}

func (m Mmapped) Close() error {
	return syscall.Munmap(m.Data)
}

func formatBytes(b int) string {
	suffixes := []struct {
		basis int
		unit  string
	}{{1024 * 1024 * 1024, "Gb"}, {1024 * 1024, "Mb"}, {1024, "Kb"}}
	for _, s := range suffixes {
		if b >= s.basis {
			return fmt.Sprintf("%.2f%s", float64(b)/float64(s.basis), s.unit)
		}
	}
	return fmt.Sprintf("%db", b)
}

const ReadToMmappedBufferReports = 20 * time.Second

// ReadToMmappedBuffer reads a (very large) file to a buffer created via mmap,
// avoiding the Go garbage collector. Reading into a conventional slice causes
// the structure to dominate the size of the heap, reducing the effectiveness of
// garbage collecting shorter lived objects.
func ReadToMmappedBuffer(filename string, fs filesystem.Interface, ctx context.Context, status chan<- string) (Mmapped, error) {
	var m Mmapped
	size, err := fs.Size(ctx, filename)
	start := time.Now()
	reported := start
	if err == nil {
		m.Data, err = syscall.Mmap(-1, 0, int(size), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_ANON|syscall.MAP_PRIVATE)
		if err == nil {
			var r io.ReadCloser
			r, err = fs.OpenRead(ctx, filename)
			if err == nil {
				buffer := m.Data
				for len(buffer) > 0 {
					var n int
					n, err = r.Read(buffer)
					if err != nil {
						break
					}
					buffer = buffer[n:]
					if status != nil {
						if d := time.Now().Sub(reported); d > ReadToMmappedBufferReports {
							s := fmt.Sprintf("%s: %s, %d%%.", filename, formatBytes(len(m.Data)-len(buffer)), ((len(m.Data)-len(buffer))*100.0)/len(m.Data))
							if len(buffer) != len(m.Data) {
								rate := float64(len(m.Data)-len(buffer)) / float64(time.Now().Sub(start))
								remaining := time.Duration(float64(len(buffer)) / rate)
								s += fmt.Sprintf(" %s remaining.", remaining.Truncate(time.Minute).String())
							}
							status <- s
							reported = time.Now()
						}
					}
				}
			}
		}
	}
	if err != nil {
		err = fmt.Errorf("can't read %s to mmapped buffer: %s", filename, err)
		log.Println(err.Error())
	}
	return m, err
}

type Buffer struct {
	buffer []byte
	lock   sync.RWMutex
}

func NewBufferWithData(buffer []byte) *Buffer {
	return &Buffer{buffer: buffer}
}

func (b *Buffer) WriteAt(data []byte, offset int64) (int, error) {
	b.lock.RLock()
	if len(b.buffer) >= int(offset)+len(data) {
		copy(b.buffer[offset:], data)
		b.lock.RUnlock()
	} else {
		b.lock.RUnlock()
		b.lock.Lock()
		if len(b.buffer) < int(offset)+len(data) {
			b.buffer = append(b.buffer, make([]byte, int(offset)+len(data)-len(b.buffer))...)
		}
		b.lock.Unlock()
		b.lock.RLock()
		copy(b.buffer[offset:], data)
		b.lock.RUnlock()
	}
	return len(data), nil
}

func (b *Buffer) ReadAt(data []byte, offset int64) (int, error) {
	return copy(data, b.buffer[offset:]), nil
}

func (b *Buffer) Close() error { return nil }

func (b *Buffer) Bytes() []byte { return b.buffer }

func (b *Buffer) Len() int { return len(b.buffer) }

type Uint64MapBuilder struct {
	Layout   Uint64MapLayout
	buckets  *ByteArraysBuilder
	offset   int64
	pointers []uint32
	lock     sync.Mutex
}

func NewUint64MapBuilder(bucketBits int, tagBits int) *Uint64MapBuilder {
	return &Uint64MapBuilder{
		Layout: Uint64MapLayout{
			BucketBits: bucketBits,
			TagBits:    tagBits,
		},
		buckets:  NewByteArraysBuilder(1 << bucketBits),
		offset:   -1,
		pointers: make([]uint32, (1<<bucketBits)+1),
	}
}

func (b *Uint64MapBuilder) Reserve(id uint64, tag Tag, length int) {
	if length < 0 {
		panic("Reserve() with negative length")
	}
	var buffer [maxUint64MapBucketHeaderLength]byte
	header := uint64MapBucketHeader{ID: id, Tag: tag, Length: length}
	n := header.Marshal(buffer[0:], &b.Layout) + length
	b.buckets.Reserve(b.Layout.BucketForID(id), n)
}

func (b *Uint64MapBuilder) FinishReservation() {
	b.buckets.FinishReservation()
}

func (b *Uint64MapBuilder) IsEmpty() bool {
	return b.buckets.IsEmpty()
}

func (b *Uint64MapBuilder) WriteHeader(w io.WriterAt, offset Offset) (Offset, error) {
	var buffer [Uint64MapLayoutLength]byte
	n := b.Layout.Marshal(buffer[0:])
	if _, err := w.WriteAt(buffer[0:n], int64(offset)); err != nil {
		return 0, fmt.Errorf("Failed to write Uint64MapLayout: %s", err)
	}
	return b.buckets.WriteHeader(w, offset.Add(n))
}

func (b *Uint64MapBuilder) WriteItem(id uint64, tag Tag, data []byte, w io.WriterAt) error {
	if tag >= 1<<b.Layout.TagBits {
		panic(fmt.Sprintf("Tag %d beyond maximum allowed with %d bits", tag, b.Layout.TagBits))
	}
	var buffer [maxUint64MapBucketHeaderLength]byte
	header := uint64MapBucketHeader{ID: id, Tag: tag, Length: len(data)}
	n := header.Marshal(buffer[0:], &b.Layout)
	return b.buckets.WriteItem(w, b.Layout.BucketForID(id), buffer[0:n], data[0:])
}

func (b *Uint64MapBuilder) Length() int {
	return Uint64MapLayoutLength + b.buckets.Length()
}

type Tag int

const NoTag Tag = 0

type Tagged struct {
	Tag
	Data []byte
}

const Uint64MapLayoutLength = 2

type Uint64MapLayout struct {
	BucketBits int
	TagBits    int
}

func (u *Uint64MapLayout) BucketForID(id uint64) int {
	return int(id & ((1 << u.BucketBits) - 1))
}

func (u *Uint64MapLayout) SentinelBucket() int {
	return int(1 << u.BucketBits)
}

func (u *Uint64MapLayout) Marshal(buffer []byte) int {
	buffer[0] = byte(u.BucketBits)
	buffer[1] = byte(u.TagBits)
	return Uint64MapLayoutLength
}

func (u *Uint64MapLayout) Unmarshal(buffer []byte) int {
	u.BucketBits = int(buffer[0])
	u.TagBits = int(buffer[1])
	return Uint64MapLayoutLength
}

const maxUint64MapBucketHeaderLength = 2 * binary.MaxVarintLen64

type uint64MapBucketHeader struct {
	ID     uint64
	Tag    Tag
	Length int
}

func (u *uint64MapBucketHeader) Unmarshal(buffer []byte, bucket int, layout *Uint64MapLayout) int {
	i := 0
	idAndTag, n := binary.Uvarint(buffer[i:])
	i += n
	u.Tag = Tag(idAndTag & ((1 << layout.TagBits) - 1))
	u.ID = uint64(bucket) | ((idAndTag >> layout.TagBits) << layout.BucketBits)
	var l uint64
	l, n = binary.Uvarint(buffer[i:])
	u.Length = int(l)
	return i + n
}

func (u *uint64MapBucketHeader) Marshal(buffer []byte, layout *Uint64MapLayout) int {
	idAndTag := ((u.ID >> uint64(layout.BucketBits)) << uint64(layout.TagBits)) | uint64(u.Tag)
	i := 0
	i += binary.PutUvarint(buffer[i:], idAndTag)
	i += binary.PutUvarint(buffer[i:], uint64(u.Length))
	return i
}

type Uint64MapIterator struct {
	m      *Uint64Map
	bucket int
	start  int
	end    int
	ids    idsAndTags
}

func (u *Uint64MapIterator) Next() bool {
	if u.end >= len(u.ids.IDs) {
		u.end = 0
		u.start = 0
		for {
			if u.bucket+1 == u.m.Layout.SentinelBucket() {
				return false
			}
			u.bucket++
			u.m.fillIDsAndTagged(u.bucket, &u.ids)
			if len(u.ids.IDs) > 0 {
				break
			}
		}
	}
	u.start = u.end
	u.end = u.start + 1
	for u.end < len(u.ids.IDs) && u.ids.IDs[u.end] == u.ids.IDs[u.start] {
		u.end++
	}
	return true
}

func (u *Uint64MapIterator) ID() uint64 {
	return u.ids.IDs[u.start]
}

func (u *Uint64MapIterator) Len() int {
	return u.end - u.start
}

func (u *Uint64MapIterator) Data(i int) []byte {
	return u.ids.Tags[u.start+i].Data
}

func (u *Uint64MapIterator) Tag(i int) Tag {
	return u.ids.Tags[u.start+i].Tag
}

type Uint64Map struct {
	Layout  Uint64MapLayout
	buckets *ByteArrays
}

func NewUint64Map(data []byte) *Uint64Map {
	m := &Uint64Map{}
	i := m.Layout.Unmarshal(data)
	m.buckets = NewByteArrays(data[i:])
	return m
}

func (m *Uint64Map) Length() int {
	return Uint64MapLayoutLength + m.buckets.Length()
}

func (m *Uint64Map) MaxBucketLength() int {
	return m.buckets.MaxItemLength()
}

func (m *Uint64Map) FillTagged(id uint64, tagged []Tagged) []Tagged {
	bucket := m.Layout.BucketForID(id)
	items := m.buckets.Item(bucket)
	var header uint64MapBucketHeader
	for i := 0; i < len(items); {
		hn := header.Unmarshal(items[i:], bucket, &m.Layout)
		if i+hn+header.Length > len(items) {
			panic(fmt.Sprintf("corrupt map: bucket %d pos %d header length %d vs bucket length %d", bucket, i, header.Length, len(items)))
		}
		i += hn
		if header.ID == id {
			tagged = append(tagged, Tagged{Tag: header.Tag, Data: items[i : i+int(header.Length)]})
		}
		i += int(header.Length)
	}
	return tagged
}

func (m *Uint64Map) FindFirstWithTag(id uint64, tag Tag) []byte {
	bucket := m.Layout.BucketForID(id)
	items := m.buckets.Item(bucket)
	var header uint64MapBucketHeader
	for i := 0; i < len(items); {
		hn := header.Unmarshal(items[i:], bucket, &m.Layout)
		if i+hn+header.Length > len(items) {
			panic(fmt.Sprintf("corrupt map: bucket %d pos %d header length %d vs bucket length %d", bucket, i, header.Length, len(items)))
		}
		i += hn
		if header.ID == id && header.Tag == tag {
			return items[i : i+int(header.Length)]
		}
		i += int(header.Length)
	}
	return nil
}

func (m *Uint64Map) FindFirst(id uint64) (Tagged, bool) {
	bucket := m.Layout.BucketForID(id)
	items := m.buckets.Item(bucket)
	var header uint64MapBucketHeader
	for i := 0; i < len(items); {
		hn := header.Unmarshal(items[i:], bucket, &m.Layout)
		if i+hn+header.Length > len(items) {
			panic(fmt.Sprintf("corrupt map: bucket %d pos %d header length %d vs bucket length %d", bucket, i, header.Length, len(items)))
		}
		i += hn
		if header.ID == id {
			return Tagged{Tag: header.Tag, Data: items[i : i+int(header.Length)]}, true
		}
		i += int(header.Length)
	}
	return Tagged{}, false
}

type idsAndTags struct {
	IDs  []uint64
	Tags []Tagged
}

func (ids *idsAndTags) Len() int { return len(ids.IDs) }
func (ids *idsAndTags) Swap(i, j int) {
	ids.IDs[i], ids.IDs[j] = ids.IDs[j], ids.IDs[i]
	ids.Tags[i], ids.Tags[j] = ids.Tags[j], ids.Tags[i]
}
func (ids *idsAndTags) Less(i, j int) bool { return ids.IDs[i] < ids.IDs[j] }

func (m *Uint64Map) EachItem(f func(id uint64, tagged []Tagged, goroutine int) error, goroutines int) error {
	var cause error
	var lock sync.Mutex
	buckets := make(chan int)
	cancel := make(chan struct{}, goroutines)
	var wg sync.WaitGroup
	readBuckets := func(goroutine int) {
		ids := idsAndTags{
			IDs:  make([]uint64, 0, m.MaxBucketLength()/8), // Estimates
			Tags: make([]Tagged, 0, m.MaxBucketLength()/8),
		}
		var err error
		for bucket := range buckets {
			m.fillIDsAndTagged(bucket, &ids)
			if len(ids.IDs) > 0 {
				start := 0
				for i := 1; i < len(ids.IDs); i++ {
					if ids.IDs[i] != ids.IDs[start] {
						if err = f(ids.IDs[start], ids.Tags[start:i], goroutine); err != nil {
							break
						}
						start = i
					}
				}
				if err = f(ids.IDs[start], ids.Tags[start:], goroutine); err != nil {
					break
				}
			}
		}
		if err != nil {
			lock.Lock()
			cause = err
			cancel <- struct{}{}
			lock.Unlock()
		}
		wg.Done()
	}

	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go readBuckets(i)
	}
	for bucket := 0; bucket < m.Layout.SentinelBucket(); bucket++ {
		select {
		case buckets <- bucket:
		case <-cancel:
			break
		}
	}
	close(buckets)
	wg.Wait()
	close(cancel)
	return cause
}

func (m *Uint64Map) Begin() *Uint64MapIterator {
	return &Uint64MapIterator{
		m:      m,
		bucket: -1,
		start:  1,
		end:    1,
		ids: idsAndTags{
			IDs:  make([]uint64, 0, m.MaxBucketLength()/8), // Estimates
			Tags: make([]Tagged, 0, m.MaxBucketLength()/8),
		},
	}
}

func (m *Uint64Map) fillIDsAndTagged(bucket int, ids *idsAndTags) {
	buffer := m.buckets.Item(bucket)
	ids.Tags = ids.Tags[0:0]
	ids.IDs = ids.IDs[0:0]
	var header uint64MapBucketHeader
	for i := 0; i < len(buffer); {
		hn := header.Unmarshal(buffer[i:], bucket, &m.Layout)
		if i+hn+header.Length > len(buffer) {
			panic(fmt.Sprintf("corrupt map: bucket %d pod %d header length %d vs bucket length %d", bucket, i, header.Length, len(buffer)))
		}
		i += hn
		ids.IDs = append(ids.IDs, header.ID)
		ids.Tags = append(ids.Tags, Tagged{Tag: header.Tag, Data: buffer[i : i+int(header.Length)]})
		i += int(header.Length)
	}
	sort.Sort(ids)
}

// ComputeHistogram fills histogram with the frequency of items
// per bucket (ie histogram[4] will have a count of the number
// buckets with 4 items). If a bucket has more items than
// histogram can represent, it's added to the last element.
func (m *Uint64Map) ComputeHistogram(histogram []int) error {
	for i := range histogram {
		histogram[i] = 0
	}
	for bucket := 0; bucket < m.Layout.SentinelBucket(); bucket++ {
		buffer := m.buckets.Item(bucket)
		var header uint64MapBucketHeader
		items := 0
		for i := 0; i < len(buffer); {
			items++
			hn := header.Unmarshal(buffer[i:], bucket, &m.Layout)
			i += hn + header.Length
		}
		if items < len(histogram) {
			histogram[items]++
		} else {
			histogram[len(histogram)-1]++
		}
	}
	return nil
}
