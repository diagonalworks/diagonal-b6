package compact

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"sort"
	"sync"
	"unsafe"

	"diagonal.works/b6"
	"diagonal.works/b6/encoding"
	"diagonal.works/b6/ingest"
	"diagonal.works/b6/osm"
	pb "diagonal.works/b6/proto"
	"diagonal.works/b6/search"
	"github.com/golang/geo/s2"

	"google.golang.org/protobuf/proto"
)

// A semver 2.0.0 compliant version for the index format. Indicies generated
// with a different major version will fail to load.
const Version = "4.0.0"

func init() {
	if l := encoding.MarshalledSize(Header{}); l != HeaderLength {
		panic(fmt.Sprintf("Expected HeaderLength=%d, found %d", l, HeaderLength))
	}
	if l := encoding.MarshalledSize(BlockHeader{}); l != BlockHeaderLength {
		panic(fmt.Sprintf("Expected BlockHeaderLength=%d, found %d", l, BlockHeaderLength))
	}
	if l := encoding.MarshalledSize(Namespaces{}); l != NamespacesLength {
		panic(fmt.Sprintf("Expected NamespacesLength=%d, found %d", l, NamespacesLength))
	}
}

type Header struct {
	Magic             uint64
	VersionOffset     encoding.Offset
	HeaderProtoOffset encoding.Offset
	StringsOffset     encoding.Offset
	BlockOffset       encoding.Offset
}

func (h *Header) Marshal(buffer []byte) int {
	return encoding.MarshalStruct(h, buffer)
}

func (h *Header) Unmarshal(buffer []byte) int {
	return encoding.UnmarshalStruct(h, buffer)
}

func (h *Header) UnmarshalVersion(buffer []byte) string {
	v, _ := UnmarshalString(buffer[h.VersionOffset:])
	return v
}

const (
	HeaderMagic  = 0xd05fffce9126772e
	HeaderLength = 40 // Verified in init()
)

type BlockType uint64

const (
	BlockTypeFeatures    BlockType = 0
	BlockTypeSearchIndex BlockType = 1
)

type BlockHeader struct {
	Length uint64 // Length of the data following this block header
	Type   BlockType
}

func (b *BlockHeader) Marshal(buffer []byte) int {
	return encoding.MarshalStruct(b, buffer)
}

func (b *BlockHeader) Unmarshal(buffer []byte) int {
	return encoding.UnmarshalStruct(b, buffer)
}

const (
	BlockHeaderLength = 16 // Verified in init()
)

func MarshalString(s string, buffer []byte) int {
	i := binary.PutUvarint(buffer, uint64(len(s)))
	return i + copy(buffer[i:], s)
}

func UnmarshalString(buffer []byte) (string, int) {
	l, i := binary.Uvarint(buffer)
	return string(buffer[i : i+int(l)]), i + int(l)
}

func MarshalledStringEquals(buffer []byte, s string) bool {
	l, n := binary.Uvarint(buffer)
	if int(l) != len(s) {
		return false
	}
	p := unsafe.Pointer(unsafe.StringData(s))
	for i := 0; i < len(s); i++ {
		if *(*byte)(unsafe.Add(p, i)) != buffer[n+i] {
			return false
		}
	}
	return true
}

func WriteProto(w io.WriterAt, m proto.Message, offset encoding.Offset) (encoding.Offset, error) {
	marshalled, err := proto.Marshal(m)
	if err != nil {
		return 0, err
	}
	var buffer [binary.MaxVarintLen64]byte
	l := binary.PutUvarint(buffer[0:], uint64(len(marshalled)))
	if _, err = w.WriteAt(buffer[0:l], int64(offset)); err != nil {
		return 0, err
	}
	offset = offset.Add(l)
	if _, err = w.WriteAt(marshalled, int64(offset)); err != nil {
		return 0, err
	}
	return offset.Add(len(marshalled)), nil
}

func UnmarshalProto(buffer []byte, m proto.Message) error {
	l, i := binary.Uvarint(buffer)
	return proto.Unmarshal(buffer[i:i+int(l)], m)
}

type Namespace uint16

const NamespaceInvalid Namespace = 0

type NamespaceTable struct {
	ToEncoded   map[b6.Namespace]Namespace
	FromEncoded b6.Namespaces
}

func (n *NamespaceTable) FillFromNamespaces(nss []b6.Namespace) {
	n.ToEncoded = make(map[b6.Namespace]Namespace)
	n.FromEncoded = make(b6.Namespaces, len(nss)+1)
	n.FromEncoded[NamespaceInvalid] = b6.NamespaceInvalid
	for i, ns := range nss {
		n.FromEncoded[i+1] = ns
	}
	// Sort to ensure that Namespace values are ordered the same was as
	// b6.Namespace values.
	sort.Sort(n.FromEncoded)
	for i, ns := range n.FromEncoded {
		n.ToEncoded[ns] = Namespace(i)
	}
}

func (n *NamespaceTable) Encode(ns b6.Namespace) Namespace {
	if e, ok := n.ToEncoded[ns]; ok {
		return e
	}
	panic(fmt.Sprintf("Can't encode %s", ns))
}

func (n *NamespaceTable) MaybeEncode(ns b6.Namespace) (Namespace, bool) {
	ens, ok := n.ToEncoded[ns]
	return ens, ok
}

func (n *NamespaceTable) Decode(e Namespace) b6.Namespace {
	if int(e) >= len(n.FromEncoded) {
		panic(fmt.Sprintf("Can't decode %d", e))
	}
	return n.FromEncoded[e]
}

func (n *NamespaceTable) DecodeID(id FeatureID) b6.FeatureID {
	return b6.FeatureID{Type: id.Type, Namespace: n.Decode(id.Namespace), Value: id.Value}
}

func (n *NamespaceTable) EncodeID(id b6.FeatureID) FeatureID {
	return FeatureID{Type: id.Type, Namespace: n.Encode(id.Namespace), Value: id.Value}
}

func (n *NamespaceTable) FillProto(header *pb.CompactHeaderProto) {
	header.Namespaces = make([]string, len(n.FromEncoded))
	for i, ns := range n.FromEncoded {
		header.Namespaces[i] = string(ns)
	}
}

func (n *NamespaceTable) FillFromProto(header *pb.CompactHeaderProto) {
	n.FromEncoded = make(b6.Namespaces, len(header.Namespaces))
	for i, ns := range header.Namespaces {
		n.FromEncoded[i] = b6.Namespace(ns)
	}
	n.ToEncoded = map[b6.Namespace]Namespace{}
	for i, ns := range n.FromEncoded {
		n.ToEncoded[ns] = Namespace(i)
	}
}

type Namespaces [b6.FeatureTypeEnd]Namespace

func OSMNamespaces(nt *NamespaceTable) Namespaces {
	var nss Namespaces
	nss[b6.FeatureTypePoint] = nt.Encode(b6.NamespaceOSMNode)
	nss[b6.FeatureTypePath] = nt.Encode(b6.NamespaceOSMWay)
	nss[b6.FeatureTypeArea] = nt.Encode(b6.NamespaceOSMWay)
	nss[b6.FeatureTypeRelation] = nt.Encode(b6.NamespaceOSMRelation)
	return nss
}

func (n *Namespaces) Marshal(buffer []byte) int {
	i := 0
	for t := b6.FeatureTypeBegin; t < b6.FeatureTypeEnd; t++ {
		binary.LittleEndian.PutUint16(buffer[i:], uint16((*n)[t]))
		i += 2
	}
	return i
}

func (n *Namespaces) Unmarshal(buffer []byte) int {
	i := 0
	for t := b6.FeatureTypeBegin; t < b6.FeatureTypeEnd; t++ {
		(*n)[t] = Namespace(binary.LittleEndian.Uint16(buffer[i:]))
		i += 2
	}
	return i
}

func (n *Namespaces) ForType(t b6.FeatureType) Namespace {
	return (*n)[t]
}

const (
	NamespacesLength = 8 // Verified in init()
)

type TypeAndNamespace uint16

const TypeAndNamespaceInvalid TypeAndNamespace = 0

func CombineTypeAndNamespace(t b6.FeatureType, ns Namespace) TypeAndNamespace {
	return TypeAndNamespace(t<<13) | TypeAndNamespace(ns)
}

func (t TypeAndNamespace) Split() (b6.FeatureType, Namespace) {
	return b6.FeatureType(t >> 13), Namespace(t & ((1 << 13) - 1))
}

type FeatureID struct {
	Type      b6.FeatureType
	Namespace Namespace // An integer, via NamespaceTable, not a string
	Value     uint64
}

func EncodeFeatureID(id b6.FeatureID, nt *NamespaceTable) FeatureID {
	return FeatureID{Type: id.Type, Namespace: nt.Encode(id.Namespace), Value: id.Value}
}

func (f FeatureID) AddValue(delta int) FeatureID {
	if delta >= 0 {
		return FeatureID{Type: f.Type, Namespace: f.Namespace, Value: f.Value + uint64(delta)}
	} else {
		return FeatureID{Type: f.Type, Namespace: f.Namespace, Value: f.Value - uint64(-delta)}
	}
}

type FeatureIDs struct {
	namespaces []TypeAndNamespace // Parallel slices to avoid wasted space due to alignment
	values     []uint64
	lock       sync.Mutex
}

func (f *FeatureIDs) Append(id FeatureID) {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.namespaces = append(f.namespaces, CombineTypeAndNamespace(id.Type, id.Namespace))
	f.values = append(f.values, id.Value)
}

func (f *FeatureIDs) Len() int {
	return len(f.values)
}

func (f *FeatureIDs) At(i int) FeatureID {
	t, n := f.namespaces[i].Split()
	return FeatureID{Type: t, Namespace: n, Value: f.values[i]}
}

func (f *FeatureIDs) Swap(i, j int) {
	f.namespaces[i], f.namespaces[j] = f.namespaces[j], f.namespaces[i]
	f.values[i], f.values[j] = f.values[j], f.values[i]
}

func (f *FeatureIDs) Less(i, j int) bool {
	if f.namespaces[i] == f.namespaces[j] {
		return f.values[i] < f.values[j]
	}
	return f.namespaces[i] < f.namespaces[j]
}

type featureIDIterator struct {
	f *FeatureIDs
	i int
}

func (f *featureIDIterator) Next() bool {
	f.i++
	return f.i < len(f.f.values)
}

func (f *featureIDIterator) FeatureID() FeatureID {
	t, n := f.f.namespaces[f.i].Split()
	return FeatureID{Type: t, Namespace: n, Value: f.f.values[f.i]}
}

func (f *FeatureIDs) Begin() FeatureIDIterator {
	return &featureIDIterator{f: f, i: -1}
}

type LatLng struct {
	LatE7 int32
	LngE7 int32
}

func (l *LatLng) FromS2LatLng(ll s2.LatLng) {
	l.LatE7 = ll.Lat.E7()
	l.LngE7 = ll.Lng.E7()
}

func (l *LatLng) ToS2LatLng() s2.LatLng {
	return s2.LatLngFromDegrees(float64(l.LatE7)/1e7, float64(l.LngE7)/1e7)
}

func (l *LatLng) Marshal(_ TypeAndNamespace, buffer []byte) int {
	i := binary.PutUvarint(buffer, EncodeValueType(b6.ValueTypeLatLng, encoding.ZigzagEncode(int64(l.LatE7))))
	binary.LittleEndian.PutUint32(buffer[i:], uint32(l.LngE7))
	return i + 4
}

func (l *LatLng) Unmarshal(_ TypeAndNamespace, buffer []byte) int {
	lat, i := DecodeValue(buffer)
	l.LatE7 = int32(encoding.ZigzagDecode(lat))
	l.LngE7 = int32(binary.LittleEndian.Uint32(buffer[i:]))
	return i + 4
}

func LatLngFromDegrees(lat float64, lng float64) LatLng {
	return LatLng{LatE7: int32(lat * 1e7), LngE7: int32(lng * 1e7)}
}

func LatLngFromS2Point(p s2.Point) LatLng {
	var ll LatLng
	ll.FromS2LatLng(s2.LatLngFromPoint(p))
	return ll
}

type LatLngs []LatLng

func (lls LatLngs) Marshal(_ TypeAndNamespace, buffer []byte) int {
	i := binary.PutUvarint(buffer, EncodeValueType(b6.ValueTypeValues, EncodeGeometry(GeometryEncodingLatLngs, len(lls))))
	return i + lls.MarshalWithoutLength(buffer[i:])
}

func (lls LatLngs) MarshalWithoutLength(buffer []byte) int {
	last := LatLng{LatE7: 0, LngE7: 0}
	i := 0
	for _, p := range lls {
		i += binary.PutVarint(buffer[i:], int64(p.LatE7-last.LatE7))
		i += binary.PutVarint(buffer[i:], int64(p.LngE7-last.LngE7))
		last = p
	}
	return i
}

func (lls *LatLngs) Unmarshal(_ TypeAndNamespace, buffer []byte) int {
	v, i := DecodeValue(buffer)
	return i + lls.UnmarshalWithoutLength(DecodeGeometryLen(v), buffer[i:])
}

func (lls *LatLngs) UnmarshalWithoutLength(l int, buffer []byte) int {
	for len(*lls) < int(l) {
		*lls = append(*lls, LatLng{})
	}
	*lls = (*lls)[0:l]
	last := LatLng{LatE7: 0, LngE7: 0}
	i := 0
	for j := range *lls {
		deltaLat, n := binary.Varint(buffer[i:])
		i += n
		deltaLng, n := binary.Varint(buffer[i:])
		i += n
		(*lls)[j].LatE7 = last.LatE7 + int32(deltaLat)
		(*lls)[j].LngE7 = last.LngE7 + int32(deltaLng)
		last = (*lls)[j]
	}
	return i
}

const ValueTypeBits = 2

type Value interface {
	Marshal(TypeAndNamespace, []byte) int
	Unmarshal(TypeAndNamespace, []byte) int
}

func EncodeValueType(t b6.ValueType, v uint64) uint64 {
	if e := v << ValueTypeBits; e>>ValueTypeBits != v {
		panic("Can't encode value type")
	} else {
		e |= uint64(t)
		return e
	}
}

func DecodeValue(buffer []byte) (uint64, int) {
	v, n := binary.Uvarint(buffer)
	return v >> ValueTypeBits, n
}

func fromCompactValue(v Value, s encoding.Strings, nt *NamespaceTable) b6.Value {
	switch v := v.(type) {
	case *Int:
		return b6.StringExpression(s.Lookup(int(*v)))
	case *LatLng:
		return b6.LatLng(v.ToS2LatLng())
	case *LatLngs:
		vs := b6.Values(make([]b6.Value, 0, len(*v)))
		for _, ll := range *v {
			vs = append(vs, b6.LatLng(ll.ToS2LatLng()))
		}
		return vs
	case *References:
		vs := b6.Values(make([]b6.Value, 0, len(*v)))
		for _, r := range *v {
			typ, ns := r.TypeAndNamespace.Split()
			vs = append(vs, b6.FeatureID{typ, nt.Decode(ns), r.Value})
		}
		return vs
	case *ReferencesAndLatLngs:
		vs := b6.Values(make([]b6.Value, 0, len(*v)))
		for _, r := range *v {
			var rll b6.Value
			if r.Reference != ReferenceInvald {
				typ, ns := r.Reference.TypeAndNamespace.Split()
				rll = b6.FeatureID{typ, nt.Decode(ns), r.Reference.Value}
			} else {
				vs = append(vs, b6.LatLng(r.LatLng.ToS2LatLng()))
			}

			vs = append(vs, rll)
		}
		return vs
	default:
		panic("cannot convert from compact value")
	}
}

func toCompactValue(v b6.Value, s *encoding.StringTableBuilder, nt *NamespaceTable, e GeometryEncoding) Value {
	switch v := v.(type) {
	case b6.StringExpression:
		r := Int(s.Lookup(v.String()))
		return &r
	case b6.LatLng:
		return &LatLng{v.Lat.E7(), v.Lng.E7()}
	case b6.FeatureID:
		return &Reference{CombineTypeAndNamespace(v.Type, nt.Encode(v.Namespace)), v.Value}
	case b6.Values:
		switch e {
		case GeometryEncodingLatLngs:
			lls := LatLngs(make([]LatLng, 0, len(v)))
			for _, x := range v {
				lls = append(lls, *toCompactValue(x, s, nt, e).(*LatLng))
			}
			return &lls
		case GeometryEncodingReferences:
			refs := References(make([]Reference, 0, len(v)))
			for _, x := range v {
				refs = append(refs, *toCompactValue(x, s, nt, e).(*Reference))
			}
			return &refs
		case GeometryEncodingMixed:
			m := ReferencesAndLatLngs(make([]ReferenceAndLatLng, 0, len(v)))
			for _, x := range v {
				c := toCompactValue(x, s, nt, e)
				switch c := c.(type) {
				case *LatLng:
					m = append(m, ReferenceAndLatLng{ReferenceInvald, *c})
				case *Reference:
					m = append(m, ReferenceAndLatLng{*c, LatLng{LatE7: 0, LngE7: 0}})
				default:
					panic("not implemented")
				}
			}
			return &m
		default:
			panic("not implemented")
		}
	}

	panic("cannot convert to compact value")
}

type Tag struct {
	Key   int
	Value Value
}

type Int int

func (i *Int) Marshal(_ TypeAndNamespace, buffer []byte) int {
	return binary.PutUvarint(buffer, EncodeValueType(b6.ValueTypeString, uint64(*i)))
}

func (i *Int) Unmarshal(_ TypeAndNamespace, buffer []byte) int {
	v, n := DecodeValue(buffer)
	*i = Int(v)
	return n
}

func (t *Tag) Marshal(tns TypeAndNamespace, buffer []byte) int {
	i := binary.PutUvarint(buffer, uint64(t.Key))
	return i + t.Value.Marshal(tns, buffer[i:])
}

func inferValueType(buffer []byte) Value {
	v, _ := binary.Uvarint(buffer)
	switch b6.ValueType(v & ((1 << ValueTypeBits) - 1)) {
	case b6.ValueTypeString:
		i := Int(0)
		return &i
	case b6.ValueTypeLatLng:
		return &LatLng{}
	case b6.ValueTypeValues:
		switch DecodeGeometryEncoding(v >> ValueTypeBits) {
		case GeometryEncodingLatLngs:
			return &LatLngs{}
		case GeometryEncodingReferences:
			return &References{}
		case GeometryEncodingMixed:
			return &ReferencesAndLatLngs{}
		default:
			panic("not implemented")
		}
	default:
		panic("not implemented")
	}
}

func (t *Tag) Unmarshal(tns TypeAndNamespace, buffer []byte) int {
	key, i := binary.Uvarint(buffer)
	t.Key = int(key)
	t.Value = inferValueType(buffer[i:])
	return i + t.Value.Unmarshal(tns, buffer[i:])
}

type Tags []Tag

func (t Tags) Marshal(tns TypeAndNamespace, buffer []byte) int {
	i := binary.PutUvarint(buffer, uint64(len(t)))
	for _, tag := range t {
		i += tag.Marshal(tns, buffer[i:])
	}
	return i
}

func (t *Tags) Unmarshal(tns TypeAndNamespace, buffer []byte) int {
	l, i := binary.Uvarint(buffer)
	for j := 0; j < int(l); j++ {
		if j >= len(*t) {
			*t = append(*t, Tag{})
		}
		i += (*t)[j].Unmarshal(tns, buffer[i:])
	}
	*t = (*t)[0:l]
	return i
}

func (t *Tags) FromOSM(tags osm.Tags, f ingest.OSMFeature, s *encoding.StringTableBuilder, nt *NamespaceTable) {
	for len(*t) < len(tags) {
		*t = append(*t, Tag{})
	}
	*t = (*t)[0:len(tags)]
	for i, tag := range tags {
		(*t)[i].Key = s.Lookup(ingest.KeyForOSMKey(tag.Key))
		value := Int(s.Lookup(tag.Value))
		(*t)[i].Value = &value
	}

	if f.Node != nil {
		*t = (*t)[0 : len(*t)+1]
		(*t)[len(tags)].Key = s.Lookup(b6.PointTag)
		ll := LatLng{int32(math.Round(f.Node.Location.Lat * 1e7)), int32(math.Round(f.Node.Location.Lng * 1e7))}
		(*t)[len(tags)].Value = &ll
	} else if f.Way != nil {
		*t = (*t)[0 : len(*t)+1]
		(*t)[len(tags)].Key = s.Lookup(b6.PathTag)
		refs := References(make([]Reference, 0, len(f.Way.Nodes)))
		for _, n := range f.Way.Nodes {
			refs = append(refs, Reference{TypeAndNamespace: CombineTypeAndNamespace(b6.FeatureTypePoint, nt.Encode(b6.NamespaceOSMNode)), Value: uint64(n)})
		}
		(*t)[len(tags)].Value = &refs
	}
}

func (t *Tags) FromFeature(f b6.Taggable, s *encoding.StringTableBuilder, nt *NamespaceTable) {
	tags := f.AllTags()
	for len(*t) < len(tags) {
		*t = append(*t, Tag{})
	}

	*t = (*t)[0:len(tags)]
	for i, tag := range tags {
		(*t)[i].Key = s.Lookup(tag.Key)

		e := GeometryEncodingInvalid
		if f, ok := f.(b6.PhysicalFeature); ok && f.Get(b6.PathTag) != b6.InvalidTag() {
			e = GeometryEncodingForPath(f) // TODO(mari): fix all unsafe physical casts
		}

		(*t)[i].Value = toCompactValue(tag.Value, s, nt, e)
	}
}

type MarshalledTags struct {
	Tags    []byte
	Strings encoding.Strings
	Nt      *NamespaceTable
	Tns     TypeAndNamespace
}

func (m MarshalledTags) AllTags() b6.Tags {
	var t Tags
	t.Unmarshal(m.Tns, m.Tags)

	tags := make([]b6.Tag, 0, 2)
	for _, tag := range t {
		tags = append(tags, b6.Tag{Key: m.Strings.Lookup(tag.Key), Value: fromCompactValue(tag.Value, m.Strings, m.Nt)})
	}
	return tags
}

func (m MarshalledTags) Get(key string) b6.Tag {
	l, i := binary.Uvarint(m.Tags)
	var tag Tag
	for j := 0; j < int(l); j++ {
		i += tag.Unmarshal(m.Tns, m.Tags[i:])
		if m.Strings.Equal(tag.Key, key) {
			return b6.Tag{key, fromCompactValue(tag.Value, m.Strings, m.Nt)}
		}
	}
	return b6.InvalidTag()
}

func (m MarshalledTags) GeometryType() b6.GeometryType {
	if m.Get(b6.PointTag) != b6.InvalidTag() {
		return b6.GeometryTypePoint
	} else if m.Get(b6.PathTag) != b6.InvalidTag() {
		return b6.GeometryTypePath
	}

	return b6.GeometryTypeInvalid
}

func (m MarshalledTags) Point() s2.Point {
	if ll := m.Get(b6.PointTag).Value; ll != nil {
		if ll, err := b6.LatLngFromString(ll.String()); err == nil {
			return s2.PointFromLatLng(ll)
		}
	}

	return s2.Point{}
}

func (m MarshalledTags) GeometryLen() int {
	if l := m.Get(b6.PathTag); l != b6.InvalidTag() {
		return len(l.Value.(b6.Values))
	}
	return 0
}

func (m MarshalledTags) PointAt(i int) s2.Point {
	if r := m.Get(b6.PathTag); r.IsValid() && r.Value != nil && r.ValueType() == b6.ValueTypeValues {
		if len(r.Value.(b6.Values)) > i {
			if ll, ok := (r.Value.(b6.Values))[i].(b6.LatLng); ok {
				return s2.PointFromLatLng(s2.LatLng(ll))
			}
		}
	}
	panic("Expected a latlng")
}

func (m MarshalledTags) Polyline() *s2.Polyline {
	polyline := make(s2.Polyline, m.GeometryLen())
	for i := 0; i < m.GeometryLen(); i++ {
		polyline[i] = m.PointAt(i)
	}
	return &polyline
}

func (m MarshalledTags) References() []b6.Reference {
	var refs []b6.Reference
	if r := m.Get(b6.PathTag); r.IsValid() && r.ValueType() == b6.ValueTypeValues {
		for i, v := range r.Value.(b6.Values) {
			if v, ok := v.(b6.FeatureID); ok && v.IsValid() {
				ref := b6.IndexedFeatureID{FeatureID: v}
				ref.SetIndex(i)
				refs = append(refs, &ref)
			}
		}
	}
	return refs
}

func (m MarshalledTags) Reference(i int) b6.Reference {
	if t := m.Get(b6.PathTag); t.IsValid() && t.ValueType() == b6.ValueTypeValues {
		if vs, ok := t.Value.(b6.Values); ok && len(vs) > i {
			if r, ok := vs[i].(b6.Reference); ok {
				return r
			}
		}
	}
	return b6.FeatureIDInvalid
}

type Reference struct {
	TypeAndNamespace TypeAndNamespace
	Value            uint64
}

var ReferenceInvald Reference = Reference{TypeAndNamespace: TypeAndNamespaceInvalid, Value: 0}

func (r *Reference) Marshal(primary TypeAndNamespace, buffer []byte) int {
	if r.TypeAndNamespace != primary || (r.Value&(1<<63)) == (1<<63) {
		i := binary.PutUvarint(buffer, (uint64(r.TypeAndNamespace)<<1)|1)
		return i + binary.PutUvarint(buffer[i:], r.Value)
	} else {
		return binary.PutUvarint(buffer, r.Value<<1)
	}
}

func (r *Reference) Unmarshal(primary TypeAndNamespace, buffer []byte) int {
	v, i := binary.Uvarint(buffer)
	if v&1 == 1 {
		r.TypeAndNamespace = TypeAndNamespace(v >> 1)
		var n int
		r.Value, n = binary.Uvarint(buffer[i:])
		i += n
	} else {
		r.TypeAndNamespace = primary
		r.Value = v >> 1
	}
	return i
}

type MarshalledReference []byte

func (m MarshalledReference) Length() int {
	var r Reference
	return r.Unmarshal(TypeAndNamespaceInvalid, m)
}

type References []Reference

func (rs References) Len() int      { return len(rs) }
func (rs References) Swap(i, j int) { rs[i], rs[j] = rs[j], rs[i] }
func (rs References) Less(i, j int) bool {
	if rs[i].TypeAndNamespace == rs[j].TypeAndNamespace {
		return rs[i].Value < rs[j].Value
	}
	return rs[i].TypeAndNamespace < rs[j].TypeAndNamespace
}

func (rs References) Marshal(primary TypeAndNamespace, buffer []byte) int {
	i := binary.PutUvarint(buffer, EncodeValueType(b6.ValueTypeValues, EncodeGeometry(GeometryEncodingReferences, len(rs))))
	return i + rs.MarshalWithoutLength(primary, buffer[i:])
}

func (rs References) MarshalWithoutLength(primary TypeAndNamespace, buffer []byte) int {
	last := uint64(0)
	i := 0
	for _, r := range rs {
		if r.TypeAndNamespace == primary {
			last, r.Value = r.Value, encoding.ZigzagEncode(int64(r.Value)-int64(last))
		}
		i += r.Marshal(primary, buffer[i:])
	}
	return i
}

func (rs *References) Unmarshal(primary TypeAndNamespace, buffer []byte) int {
	v, i := DecodeValue(buffer)
	return i + rs.UnmarshalWithoutLength(DecodeGeometryLen(v), primary, buffer[i:])
}

func (rs *References) UnmarshalWithoutLength(l int, primary TypeAndNamespace, buffer []byte) int {
	for len(*rs) < int(l) {
		*rs = append(*rs, Reference{})
	}
	*rs = (*rs)[0:l]
	last := uint64(0)
	i := 0
	for j := range *rs {
		i += (*rs)[j].Unmarshal(primary, buffer[i:])
		if (*rs)[j].TypeAndNamespace == primary {
			(*rs)[j].Value = uint64(int64(last) + encoding.ZigzagDecode((*rs)[j].Value))
			last = (*rs)[j].Value
		}
	}
	return i
}

type MarshalledReferences []byte

func (m MarshalledReferences) Len() int {
	l, _ := binary.Uvarint(m)
	return DecodeGeometryLen(l >> ValueTypeBits)
}

type GeometryEncoding uint8

const (
	GeometryEncodingReferences GeometryEncoding = iota
	GeometryEncodingLatLngs
	GeometryEncodingMixed
	GeometryEncodingInvalid
)

func GeometryEncodingForPath(f b6.PhysicalFeature) GeometryEncoding {
	references := false
	latlngs := false
	for i := 0; i < f.GeometryLen(); i++ {
		if id := f.Reference(i).Source(); id.IsValid() {
			references = true
			if latlngs {
				return GeometryEncodingMixed
			}
		} else {
			latlngs = true
			if references {
				return GeometryEncodingMixed
			}
		}
	}
	if references {
		return GeometryEncodingReferences
	} else {
		return GeometryEncodingLatLngs
	}
}

func EncodeGeometry(e GeometryEncoding, l int) uint64 {
	var v uint64
	switch e {
	case GeometryEncodingReferences:
		v = uint64(l) << 1
	case GeometryEncodingLatLngs:
		v = (uint64(l) << 2) | 1
	case GeometryEncodingMixed:
		v = (uint64(l) << 2) | 3
	default:
		panic("Unexpected GeometryEncoding")
	}
	return v
}

func MarshalGeometryEncodingAndLength(e GeometryEncoding, l int, buffer []byte) int {
	return binary.PutUvarint(buffer, EncodeGeometry(e, l))
}

func DecodeGeometryLen(v uint64) int {
	if v&1 == 0 {
		return int(v >> 1)
	} else {
		return int(v >> 2)
	}
}

func DecodeGeometryEncoding(v uint64) GeometryEncoding {
	if v&1 == 0 {
		return GeometryEncodingReferences
	} else if v&2 == 0 {
		return GeometryEncodingLatLngs
	} else {
		return GeometryEncodingMixed
	}
}

func UnmarshalGeometryEncodingAndLength(buffer []byte) (GeometryEncoding, int, int) {
	v, i := binary.Uvarint(buffer)
	return DecodeGeometryEncoding(v), DecodeGeometryLen(v), i
}

type Bits []bool

func (b Bits) Marshal(buffer []byte) int {
	i := binary.PutUvarint(buffer, uint64(len(b)))
	bits := byte(0)
	for j, v := range b {
		if v {
			bits |= (1 << (j % 8))
		}
		if j%8 == 7 {
			buffer[i] = bits
			i++
			bits = 0
		}
	}
	if len(b)%8 != 0 {
		buffer[i] = bits
		i++
	}
	return i
}

func (b *Bits) Unmarshal(buffer []byte) int {
	l, i := binary.Uvarint(buffer)
	for len(*b) < int(l) {
		*b = append(*b, false)
	}
	*b = (*b)[0:l]
	j := 0
	for block := 0; block < int(l)/8; block++ {
		for k := uint64(0); k < 8; k++ {
			(*b)[j] = buffer[i]&(1<<k) != 0
			j++
		}
		i++
	}
	if len(*b)%8 != 0 {
		for j < len(*b) {
			(*b)[j] = buffer[i]&(1<<(j%8)) != 0
			j++
		}
		i++
	}
	return i
}

type ReferenceAndLatLng struct {
	Reference Reference
	LatLng    LatLng
}

type ReferencesAndLatLngs []ReferenceAndLatLng

func (g *ReferencesAndLatLngs) Marshal(primary TypeAndNamespace, buffer []byte) int {
	i := binary.PutUvarint(buffer, EncodeValueType(b6.ValueTypeValues, EncodeGeometry(GeometryEncodingMixed, len(*g))))
	references := make(Bits, len(*g))
	for j, r := range *g {
		references[j] = r.Reference != ReferenceInvald
	}
	i += references.Marshal(buffer[i:])
	last := ReferenceAndLatLng{Reference: Reference{Value: 0}, LatLng: LatLng{LatE7: 0, LngE7: 0}}
	for j, r := range *g {
		if references[j] {
			if r.Reference.TypeAndNamespace == primary {
				last.Reference.Value, r.Reference.Value = r.Reference.Value, encoding.ZigzagEncode(int64(r.Reference.Value)-int64(last.Reference.Value))
			}
			i += r.Reference.Marshal(primary, buffer[i:])
		} else {
			i += binary.PutVarint(buffer[i:], int64(r.LatLng.LatE7-last.LatLng.LatE7))
			i += binary.PutVarint(buffer[i:], int64(r.LatLng.LngE7-last.LatLng.LngE7))
			last.LatLng = r.LatLng
		}
	}
	return i
}

func (g *ReferencesAndLatLngs) Unmarshal(primary TypeAndNamespace, buffer []byte) int {
	v, i := DecodeValue(buffer)
	return i + g.UnmarshalWithoutLength(DecodeGeometryLen(v), primary, buffer[i:])
}

func (g *ReferencesAndLatLngs) UnmarshalWithoutLength(l int, primary TypeAndNamespace, buffer []byte) int {
	for len(*g) < l {
		*g = append(*g, ReferenceAndLatLng{})
	}
	*g = (*g)[0:l]
	references := make(Bits, l)
	i := references.Unmarshal(buffer)
	last := ReferenceAndLatLng{Reference: Reference{Value: 0}, LatLng: LatLng{LatE7: 0, LngE7: 0}}
	for j := range *g {
		if references[j] {
			i += (*g)[j].Reference.Unmarshal(primary, buffer[i:])
			if (*g)[j].Reference.TypeAndNamespace == primary {
				(*g)[j].Reference.Value = uint64(int64(last.Reference.Value) + encoding.ZigzagDecode((*g)[j].Reference.Value))
				last.Reference.Value = (*g)[j].Reference.Value
			}
		} else {
			deltaLat, n := binary.Varint(buffer[i:])
			i += n
			deltaLng, n := binary.Varint(buffer[i:])
			i += n
			(*g)[j].LatLng.LatE7 = last.LatLng.LatE7 + int32(deltaLat)
			(*g)[j].LatLng.LngE7 = last.LatLng.LngE7 + int32(deltaLng)
			last.LatLng = (*g)[j].LatLng
		}
	}
	return i
}

type CommonPoint struct {
	Tags
	Path Reference
}

func (c *CommonPoint) Marshal(nss *Namespaces, buffer []byte) int {
	i := c.Tags.Marshal(TypeAndNamespaceInvalid, buffer)
	i += c.Path.Marshal(CombineTypeAndNamespace(b6.FeatureTypePath, nss.ForType(b6.FeatureTypePath)), buffer[i:])
	return i
}

func (c *CommonPoint) Unmarshal(nss *Namespaces, buffer []byte) int {
	i := c.Tags.Unmarshal(TypeAndNamespaceInvalid, buffer)
	i += c.Path.Unmarshal(CombineTypeAndNamespace(b6.FeatureTypePath, nss.ForType(b6.FeatureTypePath)), buffer[i:])
	return i
}

func CombinePointAndPath(point []byte, nss *Namespaces, r Reference, buffer []byte) int {
	i := copy(buffer, point)
	i += r.Marshal(CombineTypeAndNamespace(b6.FeatureTypePath, nss.ForType(b6.FeatureTypePath)), buffer[i:])
	return i
}

type PointReferences struct {
	Paths     References
	Relations References
}

func (p *PointReferences) Marshal(nss *Namespaces, buffer []byte) int {
	sort.Sort(p.Paths)     // Minimise deltas, as order is not important
	sort.Sort(p.Relations) // Minimise deltas, as order is not important
	i := p.Paths.Marshal(CombineTypeAndNamespace(b6.FeatureTypePath, nss.ForType(b6.FeatureTypePath)), buffer)
	i += p.Relations.Marshal(CombineTypeAndNamespace(b6.FeatureTypeRelation, nss.ForType(b6.FeatureTypeRelation)), buffer[i:])
	return i
}

func (p *PointReferences) Unmarshal(nss *Namespaces, buffer []byte) int {
	i := p.Paths.Unmarshal(CombineTypeAndNamespace(b6.FeatureTypePath, nss.ForType(b6.FeatureTypePath)), buffer)
	i += p.Relations.Unmarshal(CombineTypeAndNamespace(b6.FeatureTypeRelation, nss.ForType(b6.FeatureTypeRelation)), buffer[i:])
	return i
}

type FullPoint struct {
	Tags
	PointReferences
}

func (p *FullPoint) Marshal(nss *Namespaces, buffer []byte) int {
	i := p.Tags.Marshal(TypeAndNamespaceInvalid, buffer)
	i += p.PointReferences.Marshal(nss, buffer[i:])
	return i
}

func (p *FullPoint) Unmarshal(nss *Namespaces, buffer []byte) int {
	i := p.Tags.Unmarshal(TypeAndNamespaceInvalid, buffer)
	i += p.PointReferences.Unmarshal(nss, buffer[i:])
	return i
}

func CombinePointAndReferences(point []byte, references PointReferences, nss *Namespaces, buffer []byte) int {
	i := copy(buffer, point)
	i += references.Marshal(nss, buffer[i:])
	return i
}

type Path struct { // TODO(mari): delete once u sort out references
	Tags
	Areas     References
	Relations References
}

func (ts *Tags) PathLen(s *encoding.StringTable) int {
	for _, t := range *ts {
		if s.Equal(t.Key, b6.PathTag) {
			switch v := t.Value.(type) {
			case *LatLngs:
				return len(*v)
			case *References:
				return len(*v)
			case *ReferencesAndLatLngs:
				return len(*v)
			default:
				panic("invalid path tag compact value type")
			}
		}
	}
	return 0
}

func (ts *Tags) Reference(i int, s *encoding.StringTable) (Reference, bool) {
	for _, t := range *ts {
		if s.Equal(t.Key, b6.PathTag) {
			if refs, ok := t.Value.(*References); ok && len(*refs) > i {
				return (*refs)[i], true
			} else if refs, ok := t.Value.(*ReferencesAndLatLngs); ok && len(*refs) > i {
				return (*refs)[i].Reference, (*refs)[i].Reference != ReferenceInvald
			}
		}
	}

	return Reference{}, false
}

func (p *Path) Marshal(nss *Namespaces, buffer []byte) int {
	sort.Sort(p.Areas) // Minimise deltas, as order is not important
	i := p.Tags.Marshal(CombineTypeAndNamespace(b6.FeatureTypePoint, nss[b6.FeatureTypePoint]), buffer)
	i += p.Areas.Marshal(CombineTypeAndNamespace(b6.FeatureTypeArea, nss[b6.FeatureTypeArea]), buffer[i:])
	i += p.Relations.Marshal(CombineTypeAndNamespace(b6.FeatureTypeRelation, nss[b6.FeatureTypeRelation]), buffer[i:])
	return i
}

func (p *Path) Unmarshal(nss *Namespaces, buffer []byte) int {
	i := p.Tags.Unmarshal(CombineTypeAndNamespace(b6.FeatureTypePoint, nss[b6.FeatureTypePoint]), buffer)
	i += p.Areas.Unmarshal(CombineTypeAndNamespace(b6.FeatureTypeArea, nss[b6.FeatureTypeArea]), buffer[i:])
	i += p.Relations.Unmarshal(CombineTypeAndNamespace(b6.FeatureTypeRelation, nss[b6.FeatureTypeRelation]), buffer[i:])
	return i
}

func (p *Path) FromOSM(way *osm.Way, s *encoding.StringTableBuilder, nt *NamespaceTable) {
	p.Tags.FromOSM(way.Tags, ingest.OSMFeature{Way: way}, s, nt)
}

func (p *Path) FromFeature(t b6.Taggable, s *encoding.StringTableBuilder, nt *NamespaceTable) {
	p.Tags.FromFeature(t, s, nt)
}

type AreaGeometry interface {
	Len() int
	PathIDs(i int) (References, bool)
	Polygon(i int) (*s2.Polygon, bool)
	Marshal(primary TypeAndNamespace, buffer []byte) int
	UnmarshalWithoutLength(l int, primary TypeAndNamespace, buffer []byte) int
}

func GeometryEncodingForArea(f *ingest.AreaFeature) GeometryEncoding {
	references := false
	latlngs := false
	for i := 0; i < f.Len(); i++ {
		if _, ok := f.PathIDs(i); ok {
			references = true
			if latlngs {
				return GeometryEncodingMixed
			}
		} else {
			latlngs = true
			if references {
				return GeometryEncodingMixed
			}
		}
	}
	if references {
		return GeometryEncodingReferences
	} else {
		return GeometryEncodingLatLngs
	}
}

type PolygonGeometryReferences struct {
	Paths References
}

func (p *PolygonGeometryReferences) FromPathIDs(paths []b6.FeatureID, nt *NamespaceTable) {
	p.Paths = p.Paths[0:0]
	for i, id := range paths {
		p.Paths[i] = Reference{TypeAndNamespace: CombineTypeAndNamespace(id.Type, nt.Encode(id.Namespace)), Value: id.Value}
	}
}

func (p *PolygonGeometryReferences) Marshal(paths TypeAndNamespace, buffer []byte) int {
	return p.Paths.Marshal(paths, buffer)
}

func (p *PolygonGeometryReferences) Unmarshal(paths TypeAndNamespace, buffer []byte) int {
	return p.Paths.Unmarshal(paths, buffer)
}

type AreaGeometryReferences struct {
	Polygons []int
	Paths    References
}

func (a *AreaGeometryReferences) Len() int {
	return len(a.Polygons) + 1
}

func (a *AreaGeometryReferences) PathIDs(i int) (References, bool) {
	start := 0
	if i > 0 {
		start = a.Polygons[i-1]
	}
	end := len(a.Paths)
	if i < len(a.Polygons) {
		end = a.Polygons[i]
	}
	return a.Paths[start:end], true
}

func (a *AreaGeometryReferences) Polygon(i int) (*s2.Polygon, bool) {
	return nil, false
}

func (a *AreaGeometryReferences) Marshal(paths TypeAndNamespace, buffer []byte) int {
	i := MarshalGeometryEncodingAndLength(GeometryEncodingReferences, len(a.Polygons), buffer)
	i += encoding.MarshalDeltaCodedInts(a.Polygons, buffer[i:])
	return i + a.Paths.Marshal(paths, buffer[i:])
}

func (a *AreaGeometryReferences) Unmarshal(paths TypeAndNamespace, buffer []byte) int {
	_, l, i := UnmarshalGeometryEncodingAndLength(buffer)
	return l + a.UnmarshalWithoutLength(l, paths, buffer[i:])
}

func (a *AreaGeometryReferences) UnmarshalWithoutLength(l int, paths TypeAndNamespace, buffer []byte) int {
	i := 0
	a.Polygons, i = encoding.UnmarshalDeltaCodedInts(a.Polygons[0:0], int(l), buffer[i:])
	return i + a.Paths.Unmarshal(paths, buffer[i:])
}

type MarshalledMultiPolygon []byte

func (m MarshalledMultiPolygon) Len() int {
	l, _ := binary.Uvarint(m)
	return int(l) + 1
}

var _ AreaGeometry = &AreaGeometryReferences{}

type PolygonGeometryLatLngs struct {
	Loops  []int
	Points LatLngs
}

func (p *PolygonGeometryLatLngs) Polygon() *s2.Polygon {
	points := make([]s2.Point, len(p.Points))
	for i, p := range p.Points {
		points[i] = s2.PointFromLatLng(p.ToS2LatLng())
	}
	loops := make([]*s2.Loop, len(p.Loops)+1)
	start := 0
	for loop := 0; loop <= len(p.Loops); loop++ {
		end := len(points)
		if loop < len(p.Loops) {
			end = p.Loops[loop]
		}
		loops[loop] = s2.LoopFromPoints(points[start:end])
		start = end
	}
	return s2.PolygonFromLoops(loops)
}

func (p *PolygonGeometryLatLngs) FromS2Polygon(pp *s2.Polygon) {
	p.Loops = p.Loops[0:0]
	p.Points = p.Points[0:0]
	for i := 0; i < pp.NumLoops(); i++ {
		if i > 0 {
			p.Loops = append(p.Loops, len(p.Points))
		}
		loop := pp.Loop(i)
		for j := 0; j < loop.NumVertices(); j++ {
			p.Points = append(p.Points, LatLngFromS2Point(loop.Vertex(j)))
		}
	}
}

func (p *PolygonGeometryLatLngs) Marshal(buffer []byte) int {
	i := binary.PutUvarint(buffer, uint64(len(p.Loops)))
	i += encoding.MarshalDeltaCodedInts(p.Loops, buffer[i:])
	return i + p.Points.Marshal(TypeAndNamespaceInvalid, buffer[i:])
}

func (p *PolygonGeometryLatLngs) Unmarshal(buffer []byte) int {
	l, i := binary.Uvarint(buffer)
	n := 0
	p.Loops, n = encoding.UnmarshalDeltaCodedInts(p.Loops[0:0], int(l), buffer[i:])
	i += n
	return i + p.Points.Unmarshal(TypeAndNamespaceInvalid, buffer[i:])
}

type AreaGeometryLatLngs struct {
	Polygons []PolygonGeometryLatLngs
}

func (a *AreaGeometryLatLngs) Len() int {
	return len(a.Polygons)
}

func (a *AreaGeometryLatLngs) PathIDs(i int) (References, bool) {
	return nil, false
}

func (a *AreaGeometryLatLngs) Polygon(i int) (*s2.Polygon, bool) {
	return a.Polygons[i].Polygon(), true
}

func (a *AreaGeometryLatLngs) Marshal(paths TypeAndNamespace, buffer []byte) int {
	i := MarshalGeometryEncodingAndLength(GeometryEncodingLatLngs, len(a.Polygons), buffer)
	for _, p := range a.Polygons {
		i += p.Marshal(buffer[i:])
	}
	return i
}

func (a *AreaGeometryLatLngs) Unmarshal(paths TypeAndNamespace, buffer []byte) int {
	_, l, i := UnmarshalGeometryEncodingAndLength(buffer)
	return l + a.UnmarshalWithoutLength(l, paths, buffer[i:])
}

func (a *AreaGeometryLatLngs) UnmarshalWithoutLength(l int, paths TypeAndNamespace, buffer []byte) int {
	for len(a.Polygons) < l {
		a.Polygons = append(a.Polygons, PolygonGeometryLatLngs{})
	}
	a.Polygons = a.Polygons[0:l]
	i := 0
	for j := range a.Polygons {
		i += a.Polygons[j].Unmarshal(buffer[i:])
	}
	return i
}

var _ AreaGeometry = &AreaGeometryLatLngs{}

type PolygonGeometryMixed struct {
	References PolygonGeometryReferences
	LatLngs    PolygonGeometryLatLngs
}

type AreaGeometryMixed struct {
	Polygons []PolygonGeometryMixed
}

func (a *AreaGeometryMixed) Len() int {
	return len(a.Polygons)
}

func (a *AreaGeometryMixed) PathIDs(i int) (References, bool) {
	return a.Polygons[i].References.Paths, len(a.Polygons[i].References.Paths) > 0
}

func (a *AreaGeometryMixed) Polygon(i int) (*s2.Polygon, bool) {
	if len(a.Polygons[i].References.Paths) == 0 {
		return a.Polygons[i].LatLngs.Polygon(), true
	}
	return nil, false
}

func (a *AreaGeometryMixed) Marshal(paths TypeAndNamespace, buffer []byte) int {
	i := MarshalGeometryEncodingAndLength(GeometryEncodingMixed, len(a.Polygons), buffer)
	references := make(Bits, len(a.Polygons))
	for j, p := range a.Polygons {
		references[j] = len(p.References.Paths) > 0
	}
	i += references.Marshal(buffer[i:])
	for j, p := range a.Polygons {
		if references[j] {
			i += p.References.Marshal(paths, buffer[i:])
		} else {
			i += p.LatLngs.Marshal(buffer[i:])
		}
	}
	return i
}

func (a *AreaGeometryMixed) Unmarshal(paths TypeAndNamespace, buffer []byte) int {
	_, l, i := UnmarshalGeometryEncodingAndLength(buffer)
	return i + a.UnmarshalWithoutLength(l, paths, buffer[i:])

}

func (a *AreaGeometryMixed) UnmarshalWithoutLength(l int, paths TypeAndNamespace, buffer []byte) int {
	for len(a.Polygons) < l {
		a.Polygons = append(a.Polygons, PolygonGeometryMixed{})
	}
	a.Polygons = a.Polygons[0:l]
	references := make(Bits, l)
	i := references.Unmarshal(buffer)
	for j := range a.Polygons {
		if references[j] {
			i += a.Polygons[j].References.Unmarshal(paths, buffer[i:])
		} else {
			i += a.Polygons[j].LatLngs.Unmarshal(buffer[i:])
		}
	}
	return i
}

var _ AreaGeometry = &AreaGeometryMixed{}

func UnmarshalAreaGeometry(primary TypeAndNamespace, buffer []byte) (AreaGeometry, int) {
	e, l, i := UnmarshalGeometryEncodingAndLength(buffer)
	var a AreaGeometry
	switch e {
	case GeometryEncodingReferences:
		a = &AreaGeometryReferences{}
	case GeometryEncodingLatLngs:
		a = &AreaGeometryLatLngs{}
	case GeometryEncodingMixed:
		a = &AreaGeometryMixed{}
	}
	i += a.UnmarshalWithoutLength(l, primary, buffer[i:])
	return a, i
}

type Area struct {
	Tags
	Polygons  AreaGeometry
	Relations References
}

func (a *Area) Marshal(nss *Namespaces, buffer []byte) int {
	i := a.Tags.Marshal(TypeAndNamespaceInvalid, buffer)
	i += a.Polygons.Marshal(CombineTypeAndNamespace(b6.FeatureTypePath, nss.ForType(b6.FeatureTypePath)), buffer[i:])
	i += a.Relations.Marshal(CombineTypeAndNamespace(b6.FeatureTypePath, nss.ForType(b6.FeatureTypePath)), buffer[i:])
	return i
}

func (a *Area) Unmarshal(nss *Namespaces, buffer []byte) int {
	i := a.Tags.Unmarshal(TypeAndNamespaceInvalid, buffer)
	n := 0
	a.Polygons, n = UnmarshalAreaGeometry(CombineTypeAndNamespace(b6.FeatureTypePath, nss.ForType(b6.FeatureTypePath)), buffer[i:])
	i += n
	i += a.Relations.Unmarshal(CombineTypeAndNamespace(b6.FeatureTypeRelation, nss.ForType(b6.FeatureTypeRelation)), buffer[i:])
	return i
}

func (a *Area) FromOSMWay(way *osm.Way, s *encoding.StringTableBuilder, nt *NamespaceTable) {
	a.Tags.FromOSM(way.Tags, ingest.OSMFeature{}, s, nt)
	a.Polygons = &AreaGeometryReferences{
		Polygons: []int{},
		Paths:    []Reference{{TypeAndNamespace: CombineTypeAndNamespace(b6.FeatureTypePath, nt.Encode(b6.NamespaceOSMWay)), Value: uint64(way.ID)}},
	}
}

func (a *Area) FromOSMRelation(relation *osm.Relation, s *encoding.StringTableBuilder, nt *NamespaceTable) bool {
	a.Tags.FromOSM(relation.Tags, ingest.OSMFeature{}, s, nt)
	polygons := &AreaGeometryReferences{}
	start := 0
	for _, member := range relation.Members {
		switch member.Type {
		case osm.ElementTypeNode:
			return false
		case osm.ElementTypeWay:
			if member.Role == "outer" {
				start = len(polygons.Paths)
				if start > 0 {
					polygons.Polygons = append(polygons.Polygons, start)
				}
			}
			// TODO: Handle reassembling long ways
			polygons.Paths = append(polygons.Paths, Reference{TypeAndNamespace: CombineTypeAndNamespace(b6.FeatureTypePath, nt.Encode(b6.NamespaceOSMWay)), Value: uint64(member.ID)})
		case osm.ElementTypeRelation:
			return false
		}
	}
	a.Polygons = polygons
	return true
}

func (a *Area) FromFeature(f *ingest.AreaFeature, s *encoding.StringTableBuilder, nt *NamespaceTable) {
	a.Tags.FromFeature(f, s, nt)
	switch GeometryEncodingForArea(f) {
	case GeometryEncodingReferences:
		polygons := &AreaGeometryReferences{}
		start := 0
		for i := 0; i < f.Len(); i++ {
			start = len(polygons.Paths)
			if start > 0 {
				polygons.Polygons = append(polygons.Polygons, start)
			}
			paths, _ := f.PathIDs(i)
			for _, id := range paths {
				polygons.Paths = append(polygons.Paths, Reference{TypeAndNamespace: CombineTypeAndNamespace(id.Type, nt.Encode(id.Namespace)), Value: uint64(id.Value)})
			}
		}
		a.Polygons = polygons
	case GeometryEncodingLatLngs:
		polygons := &AreaGeometryLatLngs{
			Polygons: make([]PolygonGeometryLatLngs, f.Len()),
		}
		for i := 0; i < f.Len(); i++ {
			p, _ := f.Polygon(i)
			polygons.Polygons[i].FromS2Polygon(p)
		}
		a.Polygons = polygons
	case GeometryEncodingMixed:
		polygons := &AreaGeometryMixed{
			Polygons: make([]PolygonGeometryMixed, f.Len()),
		}
		for i := 0; i < f.Len(); i++ {
			if paths, ok := f.PathIDs(i); ok {
				polygons.Polygons[i].References.FromPathIDs(paths, nt)
			} else {
				p, _ := f.Polygon(i)
				polygons.Polygons[i].LatLngs.FromS2Polygon(p)
			}
		}
	}
}

type MarshalledArea []byte

func (m MarshalledArea) Tags(s encoding.Strings) b6.Taggable {
	return MarshalledTags{Tags: m, Strings: s}
}

func (m MarshalledArea) Len() int {
	var t Tags
	i := t.Unmarshal(TypeAndNamespaceInvalid, m)
	g, l, _ := UnmarshalGeometryEncodingAndLength(m[i:])
	switch g {
	case GeometryEncodingReferences:
		return l + 1
	default:
		return l
	}
}

func (m MarshalledArea) UnmarshalPolygons(paths TypeAndNamespace) AreaGeometry {
	var t Tags
	i := t.Unmarshal(paths, m)
	g, _ := UnmarshalAreaGeometry(paths, m[i:])
	return g
}

type Member struct {
	Type b6.FeatureType
	Role int
	ID   Reference
}

type Members []Member

func (m Members) Marshal(primary TypeAndNamespace, buffer []byte) int {
	i := binary.PutUvarint(buffer, uint64(len(m)))
	for _, member := range m {
		role := uint64(member.Role) << b6.FeatureTypeBits
		if int(role>>b6.FeatureTypeBits) != member.Role {
			panic("Can't encode role")
		}
		role |= uint64(member.Type)
		i += binary.PutUvarint(buffer[i:], uint64(role))
		i += member.ID.Marshal(primary, buffer[i:])
	}
	return i
}

func (m *Members) Unmarshal(primary TypeAndNamespace, buffer []byte) int {
	*m = (*m)[0:0]
	l, i := binary.Uvarint(buffer)
	var id Reference
	for j := 0; j < int(l); j++ {
		role, n := binary.Uvarint(buffer[i:])
		i += n
		i += id.Unmarshal(primary, buffer[i:])
		*m = append(*m, Member{
			Type: b6.FeatureType(role & ((1 << b6.FeatureTypeBits) - 1)),
			Role: int(role >> b6.FeatureTypeBits),
			ID:   id,
		})
	}
	return i
}

type MarshalledMembers []byte

func (m MarshalledMembers) Len() int {
	l, _ := binary.Uvarint(m)
	return int(l)
}

type Relation struct {
	Tags
	Members   Members
	Relations References
}

func (r *Relation) Marshal(primary b6.FeatureType, nss *Namespaces, buffer []byte) int {
	i := r.Tags.Marshal(TypeAndNamespaceInvalid, buffer)
	i += r.Members.Marshal(CombineTypeAndNamespace(primary, nss.ForType(primary)), buffer[i:])
	i += r.Relations.Marshal(CombineTypeAndNamespace(b6.FeatureTypeRelation, nss.ForType(b6.FeatureTypeRelation)), buffer[i:])
	return i
}

func (r *Relation) Unmarshal(primary b6.FeatureType, nss *Namespaces, buffer []byte) int {
	i := r.Tags.Unmarshal(TypeAndNamespaceInvalid, buffer)
	i += r.Members.Unmarshal(CombineTypeAndNamespace(primary, nss.ForType(primary)), buffer[i:])
	i += r.Relations.Unmarshal(CombineTypeAndNamespace(b6.FeatureTypeRelation, nss.ForType(b6.FeatureTypeRelation)), buffer[i:])
	return i
}

func (r *Relation) FromOSM(relation *osm.Relation, wayAreas *ingest.IDSet, relationAreas *ingest.IDSet, s *encoding.StringTableBuilder, nt *NamespaceTable) {
	r.Tags.FromOSM(relation.Tags, ingest.OSMFeature{}, s, nt)
	r.Members = r.Members[0:0]
	var m Member
	for _, member := range relation.Members {
		m.ID.Value = uint64(member.ID)
		m.Role = s.Lookup(member.Role)
		switch member.Type {
		case osm.ElementTypeNode:
			m.Type = b6.FeatureTypePoint
			m.ID.TypeAndNamespace = CombineTypeAndNamespace(b6.FeatureTypePoint, nt.Encode(b6.NamespaceOSMNode))
		case osm.ElementTypeWay:
			var typ b6.FeatureType
			if wayAreas.Has(uint64(member.ID)) {
				typ = b6.FeatureTypeArea
			} else {
				typ = b6.FeatureTypePath
			}
			m.Type = typ
			m.ID.TypeAndNamespace = CombineTypeAndNamespace(typ, nt.Encode(b6.NamespaceOSMWay))

		case osm.ElementTypeRelation:
			m.ID.TypeAndNamespace = CombineTypeAndNamespace(b6.FeatureTypeRelation, nt.Encode(b6.NamespaceOSMRelation))
			if relationAreas.Has(uint64(member.ID)) {
				m.Type = b6.FeatureTypeArea
			} else {
				m.Type = b6.FeatureTypeRelation
			}
		}
		r.Members = append(r.Members, m)
	}
}

func (r *Relation) FromFeature(f *ingest.RelationFeature, s *encoding.StringTableBuilder, nt *NamespaceTable) {
	r.Tags.FromFeature(f, s, nt)
	r.Members = r.Members[0:0]
	for _, member := range f.Members {
		id := Reference{TypeAndNamespace: CombineTypeAndNamespace(member.ID.Type, nt.Encode(member.ID.Namespace)), Value: member.ID.Value}
		r.Members = append(r.Members, Member{Type: member.ID.Type, ID: id, Role: s.Lookup(member.Role)})
	}
}

type MarshalledRelation []byte

func (m MarshalledRelation) Tags(s encoding.Strings) b6.Taggable {
	return MarshalledTags{Tags: m, Strings: s}
}

func (m MarshalledRelation) Len() int {
	var t Tags
	i := t.Unmarshal(TypeAndNamespaceInvalid, m)
	return MarshalledMembers(m[i:]).Len()
}

func (m MarshalledRelation) UnmarshalMembers(primary b6.FeatureType, nss *Namespaces, members *Members) {
	var t Tags
	i := t.Unmarshal(CombineTypeAndNamespace(primary, nss.ForType(primary)), m)
	members.Unmarshal(CombineTypeAndNamespace(primary, nss.ForType(primary)), m[i:])
}

type NamespaceIndex struct {
	TypeAndNamespace TypeAndNamespace
	Index            int
}

func (n *NamespaceIndex) Marshal(buffer []byte) int {
	i := binary.PutUvarint(buffer, uint64(n.TypeAndNamespace))
	i += binary.PutUvarint(buffer[i:], uint64(n.Index))
	return i
}

func (n *NamespaceIndex) Unmarshal(buffer []byte) int {
	tn, i := binary.Uvarint(buffer)
	index, l := binary.Uvarint(buffer[i:])
	n.TypeAndNamespace = TypeAndNamespace(tn)
	n.Index = int(index)
	return i + l
}

const TokenMapMaxLoadFactor = 0.6

type TokenMapEncoder struct {
	tokens  [][]string
	indices [][]int
	n       int
	b       *encoding.ByteArraysBuilder
}

func NewTokenMapEncoder() *TokenMapEncoder {
	return &TokenMapEncoder{
		tokens:  make([][]string, 1),
		indices: make([][]int, 1),
	}
}

func (t *TokenMapEncoder) Add(token string, index int) {
	if (float64(t.n+1) / float64(len(t.tokens))) > TokenMapMaxLoadFactor {
		tokens := t.tokens
		indices := t.indices
		t.tokens = make([][]string, len(t.tokens)*2)
		t.indices = make([][]int, len(t.indices)*2)
		t.n = 0
		for bucket := range tokens {
			for i := range tokens[bucket] {
				t.add(tokens[bucket][i], indices[bucket][i])
			}
		}
	}
	t.add(token, index)
}

func (t *TokenMapEncoder) add(token string, index int) {
	bucket := int(encoding.HashString(token) % uint64(len(t.tokens)))
	t.tokens[bucket] = append(t.tokens[bucket], token)
	t.indices[bucket] = append(t.indices[bucket], index)
	t.n++
}

func (t *TokenMapEncoder) FinishAdds() {
	t.b = encoding.NewByteArraysBuilder(len(t.tokens))
	f := func(bucket int, buffer []byte) error {
		t.b.Reserve(bucket, len(buffer))
		return nil
	}
	t.eachIndex(f)
	t.b.FinishReservation()
}

func (t *TokenMapEncoder) Length() int {
	if t.b == nil {
		t.FinishAdds()
	}
	return t.b.Length()
}

func (t *TokenMapEncoder) eachIndex(f func(bucket int, buffer []byte) error) error {
	var buffer [binary.MaxVarintLen64]byte
	for bucket := range t.indices {
		for _, index := range t.indices[bucket] {
			n := binary.PutUvarint(buffer[0:], uint64(index))
			if err := f(bucket, buffer[0:n]); err != nil {
				return err
			}
		}
	}
	return nil
}

func (t *TokenMapEncoder) Write(w io.WriterAt, offset encoding.Offset) (encoding.Offset, error) {
	if t.b == nil {
		t.FinishAdds()
	}
	offset, err := t.b.WriteHeader(w, offset)
	if err != nil {
		return 0, err
	}
	f := func(bucket int, buffer []byte) error {
		return t.b.WriteItem(w, bucket, buffer)
	}
	t.eachIndex(f)
	return offset, err
}

type TokenMapIterator struct {
	bucket []byte
}

func (t *TokenMapIterator) Next() (int, bool) {
	if len(t.bucket) == 0 {
		return 0, false
	}
	v, n := binary.Uvarint(t.bucket)
	t.bucket = t.bucket[n:]
	return int(v), true
}

type TokenMap struct {
	b *encoding.ByteArrays
}

func (t *TokenMap) Unmarshal(buffer []byte) int {
	t.b = encoding.NewByteArrays(buffer)
	return t.b.Length()
}

func (t *TokenMap) FindPossibleIndices(token string) TokenMapIterator {
	bucket := int(encoding.HashString(token) % uint64(t.b.NumItems()))
	return TokenMapIterator{bucket: t.b.Item(bucket)}
}

type NamespaceIndicies []NamespaceIndex

func (n *NamespaceIndicies) Marshal(buffer []byte) int {
	i := binary.PutUvarint(buffer, uint64(len(*n)))
	for _, v := range *n {
		i += v.Marshal(buffer[i:])
	}
	return i
}

func (n *NamespaceIndicies) Unmarshal(buffer []byte) int {
	*n = (*n)[0:0]
	l, i := binary.Uvarint(buffer)
	var v NamespaceIndex
	for j := 0; j < int(l); j++ {
		i += v.Unmarshal(buffer[i:])
		*n = append(*n, v)
	}
	return i
}

// PostingListHeader is the header for an encoded posting list.
// Posting lists are encoded as a series of 64 byte blocks,
// into which as many varint encoded IDs are stuffed, represented as
// the delta from the previous item in the block (or 0, for the
// first element). Namespaces are factored out by storing the index
// at which the IDs for a given namespace start. Namespaces end
// at block boundaries, padding as appropriate.
// Search happens on the posting list by first using the
// namespace start indicies to determine the range to search,
// before using binary search to determine the block in which
// the given item should fall, before linearly scanning the
// delta encoded block.
type PostingListHeader struct {
	Token      string
	Features   int
	Namespaces NamespaceIndicies
}

const PostingListHeaderMaxLength = 1024 // Empirical, for buffers

func (p *PostingListHeader) Marshal(buffer []byte) int {
	i := MarshalString(p.Token, buffer)
	i += binary.PutUvarint(buffer[i:], uint64(p.Features))
	i += p.Namespaces.Marshal(buffer[i:])
	return i
}

func (p *PostingListHeader) Unmarshal(buffer []byte) int {
	var i int
	p.Token, i = UnmarshalString(buffer)
	l, n := binary.Uvarint(buffer[i:])
	p.Features = int(l)
	i += n
	i += p.Namespaces.Unmarshal(buffer[i:])
	return i
}

func PostingListHeaderToken(buffer []byte) string {
	s, _ := UnmarshalString(buffer)
	return s
}

func PostingListHeaderTokenEquals(buffer []byte, token string) bool {
	return MarshalledStringEquals(buffer, token)
}

const PostingListBlockSize = 64 // Determined experimentally for best compression
const Padding = byte(128)

func padIDBlock(ids []byte) []byte {
	// We pad a block with bytes with the MSB set, so the varint decoder detects it as
	// a broken varint
	if len(ids)%PostingListBlockSize != 0 {
		padding := PostingListBlockSize - (len(ids) % PostingListBlockSize)
		for j := 0; j < padding; j++ {
			ids = append(ids, Padding)
		}
	}
	return ids
}

type FeatureIDIterator interface {
	Next() bool
	FeatureID() FeatureID
}

type PostingListEncoder struct {
	PostingList *PostingList

	tn       TypeAndNamespace
	ni       int
	start    int
	previous uint64
}

func NewPostingListEncoder(pl *PostingList) *PostingListEncoder {
	pl.IDs = pl.IDs[0:0]
	return &PostingListEncoder{
		PostingList: pl,
		tn:          TypeAndNamespaceInvalid,
		ni:          -1,
		start:       0,
		previous:    0,
	}
}

func (p *PostingListEncoder) Append(id FeatureID) {
	tnn := CombineTypeAndNamespace(id.Type, id.Namespace)
	if tnn != p.tn {
		p.tn = tnn
		p.ni++
		p.PostingList.Header.Namespaces = append(p.PostingList.Header.Namespaces, NamespaceIndex{TypeAndNamespace: tnn, Index: len(p.PostingList.IDs)})
		p.PostingList.IDs = padIDBlock(p.PostingList.IDs)
		p.PostingList.Header.Namespaces[p.ni].Index = len(p.PostingList.IDs)
	}
	if len(p.PostingList.IDs)%PostingListBlockSize == 0 {
		p.start = len(p.PostingList.IDs)
		p.previous = 0
	}
	var varint [binary.MaxVarintLen64]byte
	added := binary.PutUvarint(varint[0:], id.Value-p.previous)
	if (len(p.PostingList.IDs)-p.start)+added > PostingListBlockSize {
		p.PostingList.IDs = padIDBlock(p.PostingList.IDs)
		p.start = len(p.PostingList.IDs)
		added = binary.PutUvarint(varint[0:], id.Value)
	}
	for j := 0; j < added; j++ {
		p.PostingList.IDs = append(p.PostingList.IDs, varint[j])
	}
	p.previous = id.Value
}

type Iterator struct {
	header PostingListHeader
	ids    []byte
	ns     int // The namespace of the current value
	i      int // The index of the next varint id to be read within the ids slice
	value  uint64
	nt     *NamespaceTable
}

func NewIterator(buffer []byte, nt *NamespaceTable) *Iterator {
	i := &Iterator{ns: 0, i: 0, nt: nt}
	start := i.header.Unmarshal(buffer)
	i.ids = buffer[start:]
	return i
}

func (i *Iterator) Next() bool {
	if i.i >= len(i.ids) {
		return false
	}
	// If i.i (the next id to read) is within the block's padding, skip
	// to the next block.
	end := ((i.i / PostingListBlockSize) + 1) * PostingListBlockSize
	if end < len(i.ids) {
		for i.ids[end-1] == Padding {
			end--
		}
	}
	if i.i == end {
		i.i = ((i.i / PostingListBlockSize) + 1) * PostingListBlockSize
		if i.i >= len(i.ids) {
			return false
		}
	}
	for i.ns+1 < len(i.header.Namespaces) && i.i >= i.header.Namespaces[i.ns+1].Index {
		i.ns++
	}
	var n int
	if i.i%PostingListBlockSize == 0 {
		i.value, n = binary.Uvarint(i.ids[i.i:])
	} else {
		var delta uint64
		delta, n = binary.Uvarint(i.ids[i.i:])
		i.value += delta
	}
	i.i += n
	return true
}

func (i *Iterator) Advance(key search.Key) bool {
	id := key.(b6.FeatureID)
	if i.i == 0 {
		if !i.Next() {
			return false
		}
	}

	if current := i.FeatureID(); id.Less(current) || id == current {
		return true
	}

	ns := i.ns
	ii := i.i
	nn := CombineTypeAndNamespace(id.Type, i.nt.Encode(id.Namespace))
	for ns < len(i.header.Namespaces) && i.header.Namespaces[ns].TypeAndNamespace < nn {
		ns++
	}
	if ns == len(i.header.Namespaces) {
		return false
	}
	if ns != i.ns {
		ii = i.header.Namespaces[ns].Index
	}
	if i.header.Namespaces[ns].TypeAndNamespace > nn {
		i.ns = ns
		i.i = i.header.Namespaces[i.ns].Index
		i.value, _ = binary.Uvarint(i.ids[i.i:])
		return true
	}

	// The iterator may have started the function in the previous block
	// in the case ii is the first element of the next block. Either this value is
	// greater than or equal to the target, in which we will have returned above,
	// or the target will be the first element of the next block or above.
	start := ii / PostingListBlockSize
	var end int
	if ns+1 < len(i.header.Namespaces) {
		end = i.header.Namespaces[ns+1].Index / PostingListBlockSize
	} else {
		end = ((len(i.ids) - 1) / PostingListBlockSize) + 1
	}

	j := sort.Search(end-start, func(block int) bool {
		v, _ := binary.Uvarint(i.ids[(block+start)*PostingListBlockSize:])
		return v >= id.Value
	})
	if j > 0 { // Search the block before, if there is one
		ii = (j + start - 1) * PostingListBlockSize
	} else {
		ii = (j + start) * PostingListBlockSize
	}
	ons, oi, ovalue := i.ns, i.i, i.value
	i.i = ii
	for i.Next() {
		if !i.FeatureID().Less(id) {
			return true
		}
	}
	// Reset the iterator, since Advance doesn't move it when failing
	i.ns, i.i, i.value = ons, oi, ovalue
	return false
}

func (i *Iterator) Value() search.Value {
	return i.FeatureID()
}

func (i *Iterator) FeatureID() b6.FeatureID {
	t, n := i.header.Namespaces[i.ns].TypeAndNamespace.Split()
	return b6.FeatureID{Type: t, Namespace: i.nt.Decode(n), Value: i.value}
}

func (i *Iterator) EstimateLength() int {
	return (len(i.ids) - i.i) / 3
}

type PostingList struct {
	Header PostingListHeader
	IDs    []byte
}

func (p *PostingList) Fill(token string, i FeatureIDIterator) {
	p.Header.Token = token
	p.Header.Namespaces = p.Header.Namespaces[0:0]
	encoder := NewPostingListEncoder(p)
	features := 0
	for i.Next() {
		features++
		encoder.Append(i.FeatureID())
	}
	p.Header.Features = features
}

func (p *PostingList) Marshal(buffer []byte) int {
	n := p.Header.Marshal(buffer)
	return n + copy(buffer[n:], p.IDs)
}

type FeatureBlockHeader struct {
	b6.FeatureType
	Namespaces
}

const FeatureBlockHeaderLength = 1 + NamespacesLength

func (f *FeatureBlockHeader) Marshal(buffer []byte) int {
	buffer[0] = uint8(f.FeatureType)
	return 1 + f.Namespaces.Marshal(buffer[1:])
}

func (f *FeatureBlockHeader) Unmarshal(buffer []byte) int {
	f.FeatureType = b6.FeatureType(buffer[0])
	return 1 + f.Namespaces.Unmarshal(buffer[1:])
}

type FeatureBlock struct {
	FeatureBlockHeader
	Map *encoding.Uint64Map
}

func (f *FeatureBlock) Unmarshal(buffer []byte) int {
	i := f.FeatureBlockHeader.Unmarshal(buffer)
	f.Map = encoding.NewUint64Map(buffer[i:])
	return i + f.Map.Length()
}

type FeatureBlockBuilder struct {
	Header FeatureBlockHeader
	Map    *encoding.Uint64MapBuilder
}

func (f *FeatureBlockBuilder) WriteHeader(w io.WriterAt, offset encoding.Offset) (encoding.Offset, error) {
	f.Map.FinishReservation()
	var buffer [BlockHeaderLength + FeatureBlockHeaderLength]byte
	header := BlockHeader{Type: BlockTypeFeatures, Length: uint64(FeatureBlockHeaderLength + f.Map.Length())}
	i := header.Marshal(buffer[0:])
	i += f.Header.Marshal(buffer[i:])
	_, err := w.WriteAt(buffer[0:i], int64(offset))
	if err != nil {
		return offset, err
	}
	offset = offset.Add(i)
	return f.Map.WriteHeader(w, offset)
}

type NamespacedFeatureType struct {
	Namespace   Namespace
	FeatureType b6.FeatureType
}

type NamespacedFeatureTypes []NamespacedFeatureType

func (n NamespacedFeatureTypes) Len() int      { return len(n) }
func (n NamespacedFeatureTypes) Swap(i, j int) { n[i], n[j] = n[j], n[i] }
func (n NamespacedFeatureTypes) Less(i, j int) bool {
	if n[i].Namespace == n[j].Namespace {
		return n[i].FeatureType < n[j].FeatureType
	}
	return n[i].Namespace < n[j].Namespace
}

type FeatureBlockBuilders map[NamespacedFeatureType]*FeatureBlockBuilder

func (f FeatureBlockBuilders) WriteHeaders(w io.WriterAt, offset encoding.Offset) (encoding.Offset, error) {
	keys := make(NamespacedFeatureTypes, 0, len(f))
	for ns, _ := range f {
		keys = append(keys, ns)
	}
	sort.Sort(keys)
	var err error
	for _, key := range keys {
		if b := f[key]; !b.Map.IsEmpty() {
			offset, err = b.WriteHeader(w, offset)
			if err != nil {
				break
			}
		}
	}
	return offset, err
}

func (f FeatureBlockBuilders) Reserve(id FeatureID, tag encoding.Tag, length int) {
	key := NamespacedFeatureType{Namespace: id.Namespace, FeatureType: id.Type}
	if b, ok := f[key]; ok {
		b.Map.Reserve(id.Value, tag, length)
		return
	}
	panic(fmt.Sprintf("No builder for type %s in namespace %d", key.FeatureType, key.Namespace))
}

func (f FeatureBlockBuilders) WriteItem(id FeatureID, tag encoding.Tag, data []byte, w io.WriterAt) error {
	key := NamespacedFeatureType{Namespace: id.Namespace, FeatureType: id.Type}
	if b, ok := f[key]; ok {
		return b.Map.WriteItem(id.Value, tag, data, w)
	}
	panic(fmt.Sprintf("No builder for type %s in namespace %d", key.FeatureType, key.Namespace))
}

type FeatureBlocks []FeatureBlock

func (f *FeatureBlocks) Unmarshal(buffer []byte) int {
	var header BlockHeader
	var block FeatureBlock
	i := 0
	*f = (*f)[0:0]
	for i < len(buffer) {
		i += header.Unmarshal(buffer[i:])
		if header.Type == BlockTypeFeatures {
			block.Unmarshal(buffer[i:])
			*f = append(*f, block)
		}
		i += int(header.Length)
	}
	return i
}
