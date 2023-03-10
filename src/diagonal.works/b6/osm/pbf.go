package osm

import (
	"bytes"
	"compress/zlib"
	"context"
	"encoding/binary"
	"errors"
	fmt "fmt"
	"io"
	"io/ioutil"
	"os"
	"sync"

	pb "diagonal.works/b6/osm/proto"
	"google.golang.org/protobuf/proto"
)

const blobTypeOSMHeader = "OSMHeader"
const blobTypeOSMData = "OSMData"
const blobTypeDone = "Done"
const elementsPerGroup = 8000

type Emit func(e Element) error
type EmitWithGoroutine func(e Element, goroutine int) error

type blob struct {
	Type string
	Blob pb.Blob
}

func readBlob(r io.Reader) (string, []byte, error) {
	var lengthBuffer [4]byte
	_, err := io.ReadFull(r, lengthBuffer[0:])
	if err != nil {
		return "", nil, err
	}
	length := binary.BigEndian.Uint32(lengthBuffer[0:]) // Network byte order
	headerBuffer := make([]byte, length)
	_, err = io.ReadFull(r, headerBuffer)
	if err != nil {
		return "", nil, err
	}
	var header pb.BlobHeader
	err = proto.Unmarshal(headerBuffer, &header)
	if err != nil {
		return "", nil, err
	}
	blobBuffer := make([]byte, header.GetDatasize())
	var blob pb.Blob
	_, err = io.ReadFull(r, blobBuffer)
	if err != nil {
		return "", nil, err
	}
	err = proto.Unmarshal(blobBuffer, &blob)
	if err != nil {
		return "", nil, err
	}
	if blob.Raw != nil {
		return header.GetType(), blob.Raw, nil
	} else if blob.ZlibData != nil {
		reader, err := zlib.NewReader(bytes.NewBuffer(blob.ZlibData))
		defer reader.Close()
		if err != nil {
			return "", nil, err
		}
		decompressed, err := ioutil.ReadAll(reader)
		if err != nil {
			return "", nil, err
		}
		return header.GetType(), decompressed, nil
	}
	return "", nil, errors.New("Only Raw or ZlibData is supported")
}

func readBlobs(r io.Reader, blobs chan<- *blob, ctx context.Context) error {
	for {
		var lengthBuffer [4]byte
		_, err := io.ReadFull(r, lengthBuffer[0:])
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		length := binary.BigEndian.Uint32(lengthBuffer[0:]) // Network byte order
		headerBuffer := make([]byte, length)
		_, err = io.ReadFull(r, headerBuffer)
		if err != nil {
			return err
		}
		var header pb.BlobHeader
		err = proto.Unmarshal(headerBuffer, &header)
		if err != nil {
			return err
		}
		blobBuffer := make([]byte, header.GetDatasize())
		_, err = io.ReadFull(r, blobBuffer)
		if err != nil {
			return err
		}
		b := &blob{Type: header.GetType()}
		err = proto.Unmarshal(blobBuffer, &b.Blob)
		if err != nil {
			return err
		}
		select {
		case blobs <- b:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

type ReadOptions struct {
	SkipTags      bool
	SkipNodes     bool
	SkipWays      bool
	SkipRelations bool
	Parallelism   int
}

func ReadPBF(r io.Reader, emit Emit) error {
	f := func(e Element, _ int) error {
		return emit(e)
	}
	return ReadPBFWithOptions(r, f, ReadOptions{})
}

func ReadPBFWithOptions(r io.Reader, emit EmitWithGoroutine, options ReadOptions) error {
	p := options.Parallelism
	if p == 0 {
		p = 1
	}
	c := make(chan *blob, p)
	ctx, cancel := context.WithCancel(context.Background())
	var readBlobErr error
	wg := sync.WaitGroup{}
	wg.Add(p + 1)
	go func() {
		readBlobErr = readBlobs(r, c, ctx)
		for i := 0; i < p; i++ {
			c <- &blob{Type: blobTypeDone}
		}
		wg.Done()
	}()
	var readOSMDataErr error
	for i := 0; i < p; i++ {
		go func(goroutine int) {
			defer wg.Done()
			for {
				f := func(e Element) error {
					return emit(e, goroutine)
				}
				select {
				case <-ctx.Done():
					return
				case b := <-c:
					if b.Type == blobTypeOSMData {
						if err := readOSMDataBlob(b, f, options); err != nil {
							readOSMDataErr = err
							cancel()
						}
					} else if b.Type == blobTypeDone {
						return
					}
				}
			}
		}(i)
	}
	wg.Wait()
	cancel()
	close(c)
	if readBlobErr != nil && readBlobErr != context.Canceled {
		return readBlobErr
	}
	return readOSMDataErr
}

func ReadWholePBF(filename string) ([]Node, []Way, []Relation, error) {
	nodes := make([]Node, 0)
	ways := make([]Way, 0)
	relations := make([]Relation, 0)
	emit := func(element Element) error {
		switch e := element.(type) {
		case *Node:
			nodes = append(nodes, e.Clone())
		case *Way:
			ways = append(ways, e.Clone())
		case *Relation:
			relations = append(relations, e.Clone())
		}
		return nil
	}
	input, err := os.Open(filename)
	if err != nil {
		return nil, nil, nil, err
	}
	defer input.Close()
	err = ReadPBF(input, emit)
	return nodes, ways, relations, err
}

type blockHeader struct {
	LatOffset   int64
	LonOffset   int64
	Granularity int32
	Strings     [][]byte
}

func readOSMDataBlob(blob *blob, emit Emit, options ReadOptions) error {
	var raw []byte
	if blob.Blob.Raw != nil {
		raw = blob.Blob.Raw
	} else if blob.Blob.ZlibData != nil {
		reader, err := zlib.NewReader(bytes.NewBuffer(blob.Blob.ZlibData))
		defer reader.Close()
		if err != nil {
			return err
		}
		decompressed, err := ioutil.ReadAll(reader)
		if err != nil {
			return nil
		}
		raw = decompressed
	}
	return readRawOSMDataBlob(raw, emit, options)
}

func readRawOSMDataBlob(blob []byte, emit Emit, options ReadOptions) error {
	var block pb.PrimitiveBlock
	err := proto.Unmarshal(blob, &block)
	if err != nil {
		return err
	}
	header := blockHeader{
		LatOffset:   block.GetLatOffset(),
		LonOffset:   block.GetLonOffset(),
		Granularity: block.GetGranularity(),
		Strings:     block.Stringtable.GetS(),
	}
	for _, group := range block.GetPrimitivegroup() {
		err = readPrimitiveGroup(group, &header, emit, options)
		if err != nil {
			return err
		}
	}
	return nil
}

func readPrimitiveGroup(group *pb.PrimitiveGroup, header *blockHeader, emit Emit, options ReadOptions) error {
	var node Node
	var way Way
	var relation Relation
	if !options.SkipNodes {
		for _, n := range group.GetNodes() {
			if err := fillNode(n, header, &node, options); err != nil {
				return err
			}
			if err := emit(&node); err != nil {
				return err
			}
		}
		if group.Dense != nil {
			if err := readDenseNodes(group.Dense, header, emit, options); err != nil {
				return err
			}
		}
	}
	if !options.SkipWays {
		for _, w := range group.GetWays() {
			if err := fillWay(w, header, &way, options); err != nil {
				return err
			}
			if err := emit(&way); err != nil {
				return err
			}
		}
	}
	if !options.SkipRelations {
		for _, r := range group.GetRelations() {
			if err := fillRelation(r, header, &relation, options); err != nil {
				return err
			}
			if err := emit(&relation); err != nil {
				return err
			}
		}
	}
	return nil
}

func decodeAngle(angle int64, offset int64, granularity int32) float64 {
	return .000000001 * float64(offset+(int64(granularity)*angle))
}

func encodeAngle(angle float64, offset int64, granularity int32) int64 {
	return (int64(angle/.000000001) - offset) / int64(granularity)
}

func decodeDeltaEncodedAngle(angle int64, last int64, offset int64, granularity int32) float64 {
	return decodeAngle(angle+last, offset, granularity)
}

func encodeDeltaEncodedAngle(angle float64, last int64, offset int64, granularity int32) int64 {
	return encodeAngle(angle, offset, granularity) - last
}

func decodeLat(lat int64, header *blockHeader) float64 {
	return decodeAngle(lat, header.LatOffset, header.Granularity)
}

func decodeLon(lon int64, header *blockHeader) float64 {
	return decodeAngle(lon, header.LonOffset, header.Granularity)
}

func decodeDeltaEncodedLat(lat int64, last int64, header *blockHeader) float64 {
	return decodeDeltaEncodedAngle(lat, last, header.LatOffset, header.Granularity)
}

func decodeDeltaEncodedLon(lon int64, last int64, header *blockHeader) float64 {
	return decodeDeltaEncodedAngle(lon, last, header.LonOffset, header.Granularity)
}

func fillTags(keys []uint32, values []uint32, header *blockHeader, tags []Tag) ([]Tag, error) {
	if tags != nil {
		tags = tags[0:0]
	} else {
		tags = make([]Tag, 0, len(keys))
	}
	if len(keys) != len(values) {
		return nil, fmt.Errorf("Invalid tag encoding: unmatched keys and values")
	}
	for i := range keys {
		if int(keys[i]) >= len(header.Strings) {
			return nil, fmt.Errorf("Invalid tag key encoding: string table index: %d, size: %d", keys[i], len(header.Strings))
		}
		if int(values[i]) >= len(header.Strings) {
			return nil, fmt.Errorf("Invalid tag value encoding: string table index: %d, size: %d", values[i], len(header.Strings))
		}
		key := header.Strings[keys[i]]
		value := header.Strings[values[i]]
		tags = append(tags, Tag{Key: string(key), Value: string(value)})
	}
	return tags, nil
}

func fillNode(node *pb.Node, header *blockHeader, output *Node, options ReadOptions) error {
	var err error
	if !options.SkipTags {
		if output.Tags, err = fillTags(node.GetKeys(), node.GetVals(), header, output.Tags); err != nil {
			return fmt.Errorf("Failed to read tags for node %d: %s", node.GetId(), err)
		}
	}
	output.ID = NodeID(node.GetId())
	output.Location = LatLng{decodeLat(node.GetLat(), header), decodeLon(node.GetLon(), header)}
	return nil
}

func readDenseNodes(dense *pb.DenseNodes, header *blockHeader, emit Emit, options ReadOptions) error {
	var node Node
	node.Tags = make([]Tag, 0, 64)
	lastID, lastLat, lastLon := int64(0), int64(0), int64(0)
	j := 0
	keysVals := dense.GetKeysVals()
	for i, id := range dense.GetId() {
		node.Tags = node.Tags[0:0]
		id = id + lastID
		lat := decodeDeltaEncodedLat(dense.Lat[i], lastLat, header)
		lon := decodeDeltaEncodedLon(dense.Lon[i], lastLon, header)
		lastID, lastLat, lastLon = id, dense.Lat[i]+lastLat, dense.Lon[i]+lastLon
		if !options.SkipTags {
			for j < len(keysVals) {
				if keysVals[j] == 0 {
					j++
					break
				}
				if j+1 >= len(keysVals) {
					return fmt.Errorf("Invalid DenseNodes tag encoding")
				}
				key := header.Strings[keysVals[j]]
				value := header.Strings[keysVals[j+1]]
				node.Tags = append(node.Tags, Tag{Key: string(key), Value: string(value)})
				j += 2
			}
		}
		node.ID = NodeID(id)
		node.Location = LatLng{lat, lon}
		if err := emit(&node); err != nil {
			return err
		}
	}
	return nil
}

func fillWay(way *pb.Way, header *blockHeader, output *Way, options ReadOptions) error {
	var err error
	if !options.SkipTags {
		if output.Tags, err = fillTags(way.GetKeys(), way.GetVals(), header, output.Tags); err != nil {
			return fmt.Errorf("Failed to read tags for way %d: %s", way.GetId(), err)
		}
	}
	if output.Nodes != nil {
		output.Nodes = output.Nodes[0:0]
	} else {
		output.Nodes = make([]NodeID, 0, len(way.GetRefs()))
	}
	lastID := int64(0)
	for _, id := range way.GetRefs() {
		id = id + lastID
		output.Nodes = append(output.Nodes, NodeID(id))
		lastID = id
	}
	output.ID = WayID(way.GetId())
	return nil
}

func fillRelation(relation *pb.Relation, header *blockHeader, output *Relation, options ReadOptions) error {
	var err error
	if !options.SkipTags {
		if output.Tags, err = fillTags(relation.GetKeys(), relation.GetVals(), header, output.Tags); err != nil {
			return fmt.Errorf("Failed to read tags for relation %d: %s", relation.GetId(), err)
		}
	}
	if len(relation.GetMemids()) != len(relation.GetRolesSid()) {
		return fmt.Errorf("Invalid relation encoding: unmatched IDs and roles")
	}
	if len(relation.GetMemids()) != len(relation.GetTypes()) {
		return fmt.Errorf("Invalid relation encoding: unmatched IDs and types")
	}
	if output.Members != nil {
		output.Members = output.Members[0:0]
	} else {
		output.Members = make([]Member, 0, len(relation.GetMemids()))
	}
	roles := relation.GetRolesSid()
	lastID := int64(0)
	types := relation.GetTypes()
	for i, id := range relation.GetMemids() {
		id = id + lastID
		var t ElementType
		switch types[i] {
		case pb.Relation_NODE:
			t = ElementTypeNode
		case pb.Relation_WAY:
			t = ElementTypeWay
		case pb.Relation_RELATION:
			t = ElementTypeRelation
		default:
			return fmt.Errorf("Invalid relation member type: %s", relation.GetTypes()[i])
		}
		output.Members = append(output.Members, Member{ID: AnyID(id), Type: t, Role: string(header.Strings[roles[i]])})
		lastID = id
	}
	output.ID = RelationID(relation.GetId())
	return nil
}

type writerState int

const (
	writerStateNew = iota
	writerStateDenseNodes
	writerStateWays
	writerStateRelations
)

type Writer struct {
	f          io.Writer
	state      writerState
	zlib       *zlib.Writer
	zlibBuffer *bytes.Buffer
	block      pb.PrimitiveBlock
	group      pb.PrimitiveGroup
	dense      pb.DenseNodes
	ways       []*pb.Way
	relations  []*pb.Relation
	strings    map[string]int32
	lastID     int64
	lastLat    int64
	lastLon    int64
}

type BoundingBox struct {
	TopLeft     LatLng
	BottomRight LatLng
}

func (b *BoundingBox) Include(ll LatLng) {
	if ll.Lat > b.TopLeft.Lat {
		b.TopLeft.Lat = ll.Lat
	}
	if ll.Lng < b.TopLeft.Lng {
		b.TopLeft.Lng = ll.Lng
	}
	if ll.Lat < b.BottomRight.Lat {
		b.BottomRight.Lat = ll.Lat
	}
	if ll.Lng > b.BottomRight.Lng {
		b.BottomRight.Lng = ll.Lng
	}
}

var EmptyBoundingBox BoundingBox = BoundingBox{
	TopLeft:     LatLng{Lat: -91.0, Lng: 91.0},
	BottomRight: LatLng{Lat: 91.0, Lng: -91.0},
}

type WriterOptions struct {
	*BoundingBox
}

func NewWriter(f io.Writer) (*Writer, error) {
	return NewWriterWithOptions(f, &WriterOptions{})
}

func NewWriterWithOptions(f io.Writer, options *WriterOptions) (*Writer, error) {
	zlibBuffer := new(bytes.Buffer)
	z, err := zlib.NewWriterLevel(zlibBuffer, zlib.BestSpeed)
	if err != nil {
		return nil, err
	}
	w := &Writer{
		f:          f,
		zlib:       z,
		zlibBuffer: zlibBuffer,
	}
	var header pb.HeaderBlock
	header.RequiredFeatures = []string{"DenseNodes"}
	header.OptionalFeatures = []string{}
	header.Source = proto.String("http://diagonal.works")
	header.Writingprogram = proto.String("diagonal-platform")
	if options.BoundingBox != nil {
		header.Bbox = &pb.HeaderBBox{
			Top:    proto.Int64(int64(options.BoundingBox.TopLeft.Lat * 1e9)),
			Left:   proto.Int64(int64(options.BoundingBox.TopLeft.Lng * 1e9)),
			Bottom: proto.Int64(int64(options.BoundingBox.BottomRight.Lat * 1e9)),
			Right:  proto.Int64(int64(options.BoundingBox.BottomRight.Lng * 1e9)),
		}
	}
	marshalledHeader, err := proto.Marshal(&header)
	if err != nil {
		return nil, err
	}
	err = w.writeBlob(blobTypeOSMHeader, marshalledHeader)
	if err != nil {
		return nil, err
	}
	w.state = writerStateNew
	w.block.Primitivegroup = []*pb.PrimitiveGroup{&w.group}
	w.block.Stringtable = &pb.StringTable{
		S: make([][]byte, 1, elementsPerGroup), // TODO: Find a good estimate
	}
	w.block.Stringtable.S[0] = []byte{}
	w.dense.Id = make([]int64, 0, elementsPerGroup)
	w.dense.Lat = make([]int64, 0, elementsPerGroup)
	w.dense.Lon = make([]int64, 0, elementsPerGroup)
	w.dense.KeysVals = make([]int32, 0, 3*elementsPerGroup) // TODO: Find a good estimate
	return w, nil
}

func (w *Writer) WriteElement(e Element) error {
	switch e := e.(type) {
	case *Node:
		return w.WriteNode(e)
	case *Way:
		return w.WriteWay(e)
	case *Relation:
		return w.WriteRelation(e)
	}
	return nil
}

func (w *Writer) WriteNode(node *Node) error {
	if w.state != writerStateDenseNodes {
		if err := w.Flush(); err != nil {
			return err
		}
		w.state = writerStateDenseNodes
		w.group.Dense = &w.dense
		w.dense.Id = w.dense.Id[0:0]
		w.dense.Lat = w.dense.Lat[0:0]
		w.dense.Lon = w.dense.Lon[0:0]
		w.dense.KeysVals = w.dense.KeysVals[0:0]
		w.lastID = 0
		w.lastLat = 0
		w.lastLon = 0
	}
	w.dense.Id = append(w.dense.Id, int64(node.ID)-w.lastID)
	w.lastID = int64(node.ID)
	deltaLat := encodeDeltaEncodedAngle(node.Location.Lat, w.lastLat, w.block.GetLatOffset(), w.block.GetGranularity())
	w.lastLat += deltaLat
	w.dense.Lat = append(w.dense.Lat, deltaLat)
	deltaLon := encodeDeltaEncodedAngle(node.Location.Lng, w.lastLon, w.block.GetLatOffset(), w.block.GetGranularity())
	w.lastLon += deltaLon
	w.dense.Lon = append(w.dense.Lon, deltaLon)
	for _, tag := range node.Tags {
		w.dense.KeysVals = append(w.dense.KeysVals, w.lookupString(tag.Key))
		w.dense.KeysVals = append(w.dense.KeysVals, w.lookupString(tag.Value))
	}
	w.dense.KeysVals = append(w.dense.KeysVals, 0)
	if len(w.dense.Id) >= elementsPerGroup {
		return w.Flush()
	}
	return nil
}

func (w *Writer) WriteWay(way *Way) error {
	if w.state != writerStateWays {
		if err := w.Flush(); err != nil {
			return err
		}
		w.state = writerStateWays
		w.group.Ways = w.ways[0:0]
	}
	encodedWay := &pb.Way{
		Id:   proto.Int64(int64(way.ID)),
		Refs: make([]int64, len(way.Nodes)),
		Keys: make([]uint32, len(way.Tags)),
		Vals: make([]uint32, len(way.Tags)),
	}
	lastID := int64(0)
	for i, id := range way.Nodes {
		encodedWay.Refs[i] = int64(id) - lastID
		lastID = int64(id)
	}
	for i, tag := range way.Tags {
		encodedWay.Keys[i] = uint32(w.lookupString(tag.Key))
		encodedWay.Vals[i] = uint32(w.lookupString(tag.Value))
	}
	w.group.Ways = append(w.group.Ways, encodedWay)
	if len(w.group.Ways) >= elementsPerGroup {
		return w.Flush()
	}
	return nil
}

func (w *Writer) WriteRelation(relation *Relation) error {
	if w.state != writerStateRelations {
		if err := w.Flush(); err != nil {
			return err
		}
		w.state = writerStateRelations
		w.group.Relations = w.relations[0:0]
	}
	encodedRelation := &pb.Relation{
		Id:       proto.Int64(int64(relation.ID)),
		Memids:   make([]int64, len(relation.Members)),
		Types:    make([]pb.Relation_MemberType, len(relation.Members)),
		RolesSid: make([]int32, len(relation.Members)),
		Keys:     make([]uint32, len(relation.Tags)),
		Vals:     make([]uint32, len(relation.Tags)),
	}
	lastID := int64(0)
	for i, member := range relation.Members {
		id := int64(member.ID)
		encodedRelation.Memids[i] = id - lastID
		lastID = id
		switch member.Type {
		case ElementTypeNode:
			encodedRelation.Types[i] = pb.Relation_NODE
		case ElementTypeWay:
			encodedRelation.Types[i] = pb.Relation_WAY
		case ElementTypeRelation:
			encodedRelation.Types[i] = pb.Relation_RELATION
		}
		encodedRelation.RolesSid[i] = w.lookupString(member.Role)
	}
	for i, tag := range relation.Tags {
		encodedRelation.Keys[i] = uint32(w.lookupString(tag.Key))
		encodedRelation.Vals[i] = uint32(w.lookupString(tag.Value))
	}
	w.group.Relations = append(w.group.Relations, encodedRelation)
	if len(w.group.Relations) >= elementsPerGroup {
		return w.Flush()
	}
	return nil
}

func (w *Writer) lookupString(s string) int32 {
	var i int32
	var ok bool
	if i, ok = w.strings[s]; !ok {
		i = int32(len(w.block.Stringtable.S))
		w.block.Stringtable.S = append(w.block.Stringtable.S, []byte(s))
		w.strings[s] = i
	}
	return i
}

func (w *Writer) resetBlock() {
	w.group.Dense = nil
	w.group.Ways = nil
	w.group.Relations = nil
	w.strings = make(map[string]int32)
	w.block.Stringtable.S = w.block.Stringtable.S[0:1]
}

func (w *Writer) Flush() error {
	if w.state == writerStateNew {
		w.resetBlock()
		return nil
	}
	block, err := proto.Marshal(&w.block)
	if err != nil {
		return err
	}
	w.state = writerStateNew
	w.resetBlock()
	return w.writeBlob(blobTypeOSMData, block)
}

const compression = true

func (w *Writer) writeBlob(blobType string, data []byte) error {
	var blob pb.Blob

	if compression {
		w.zlibBuffer.Reset()
		w.zlib.Reset(w.zlibBuffer)
		w.zlib.Write(data)
		w.zlib.Close()
		blob.ZlibData = w.zlibBuffer.Bytes()
		blob.RawSize = proto.Int32(int32(len(data)))
	} else {
		blob.Raw = data
	}
	blobBuffer, err := proto.Marshal(&blob)
	if err != nil {
		return err
	}

	var header pb.BlobHeader
	header.Type = proto.String(blobType)
	header.Datasize = proto.Int32(int32(len(blobBuffer)))
	headerBuffer, err := proto.Marshal(&header)
	if err != nil {
		return err
	}

	var lengthBuffer [4]byte
	binary.BigEndian.PutUint32(lengthBuffer[0:], uint32(len(headerBuffer)))

	_, err = w.f.Write(lengthBuffer[0:])
	if err != nil {
		return err
	}
	_, err = w.f.Write(headerBuffer)
	if err != nil {
		return err
	}
	_, err = w.f.Write(blobBuffer)
	return err
}
