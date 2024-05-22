package b6

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"diagonal.works/b6/geojson"
	"diagonal.works/b6/geometry"
	"diagonal.works/b6/units"
	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
)

// TODO(mari): harmonise Value and Literal in expression.go
type Value interface {
	String() string
	ValueType() ValueType
}

func ValueFromString(s string, t ValueType) Value {
	switch t {
	case ValueTypeString:
		return StringExpression(s)
	case ValueTypeLatLng:
		if ll, err := LatLngFromString(s); err == nil {
			return LatLng(ll)
		}
	case ValueTypeFeatureID:
		if id := FeatureIDFromString(s); id.IsValid() {
			return id
		}
	case ValueTypeValues:
		return ValuesFromString(s)
	}

	panic("not implemented")
}

type LatLng s2.LatLng

func (ll LatLng) String() string {
	return LatLngToString(s2.LatLng(ll))
}

func (ll LatLng) ValueType() ValueType {
	return ValueTypeLatLng
}

type Values []Value

const valuesDelimiter = ";"

func (v Values) String() string {
	s := ""
	for i, x := range v {
		s += x.String()
		if i < len(v)-1 {
			s += valuesDelimiter
		}
	}
	return s
}

func TryValueFromString(s string) Value {
	if strings.Contains(s, valuesDelimiter) {
		return ValuesFromString(s)
	} else if ll, err := LatLngFromString(s); err == nil {
		return LatLng(ll)
	} else if id := FeatureIDFromString(s); id.IsValid() {
		return id
	} else {
		return StringExpression(s)
	}
}

func ValuesFromString(s string) Values {
	parts := strings.Split(s, valuesDelimiter)
	v := Values(make([]Value, 0, len(parts)))
	for _, part := range parts {
		v = append(v, TryValueFromString(part))
	}
	return v
}

func (v Values) ValueType() ValueType {
	return ValueTypeValues
}

func Set(vs Values, v Value, i int) Values {
	l := len(vs)
	if l <= i {
		l = i + 1
	}
	r := make([]Value, l)
	copy(r, vs)
	r[i] = v
	return r
}

type Tag struct {
	Key string
	Value
}

func (t Tag) IsValid() bool {
	return t.Key != "" && t.Value != nil
}

func (t Tag) String() string {
	return escapeTagPart(t.Key) + "=" + escapeTagPart(t.Value.String())
}

func (t *Tag) FromString(s string, typ ValueType) {
	var rest string
	t.Key, rest = consumeTagPart(s)
	value, _ := consumeTagPart(rest)
	t.Value = ValueFromString(value, typ)
}

func (t Tag) Equal(other Tag) bool { // TODO(mari): reuse expression equal
	if t.Key != other.Key {
		return false
	}
	switch v := t.Value.(type) {
	case StringExpression:
		if o, ok := other.Value.(StringExpression); ok {
			return string(v) == string(o)
		}
	case LatLng:
		if o, ok := other.Value.(LatLng); ok {
			return s2.LatLng(v) == s2.LatLng(o)
		}
	case Values:
		return v.String() == other.Value.String()
	}
	return false
}

type tagYAML struct {
	Key   string `yaml:"key,omitempty`
	Value Literal
}

func (t Tag) MarshalYAML() (interface{}, error) {
	if s, ok := t.Value.(StringExpression); ok {
		return escapeTagPart(t.Key) + "=" + escapeTagPart(s.String()), nil
	} else if t.Value == nil {
		return escapeTagPart(t.Key) + "=\"\"", nil
	}
	// TODO(mari): harmonise Value and Literal in expression.go
	literal, err := FromLiteral(t.Value.String())
	if err != nil {
		return nil, err
	}
	return &tagYAML{Key: t.Key, Value: literal}, nil
}

func (t *Tag) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err == nil {
		t.FromString(s, ValueTypeString)
		return nil
	}
	var y tagYAML
	if err := unmarshal(&y); err != nil {
		return err
	}
	t.Key = y.Key
	// TODO: harmonise Value and Literal in expression.go
	switch l := y.Value.AnyLiteral.(type) {
	case *PointExpression:
		t.Value = LatLng(*l)
	case *StringExpression:
		t.Value = TryValueFromString(string(*l))
	default:
		return fmt.Errorf("can't use %T as tag values", l)
	}
	return nil
}

func escapeTagPart(s string) string {
	if tagPartNeedsQuotes(s) {
		escaped := "\""
		for _, r := range s {
			switch r {
			case '\\':
				escaped += "\\\\"
			case '"':
				escaped += "\\\""
			default:
				escaped += string(r)
			}
		}
		return escaped + "\""
	}
	return s
}

func tagPartNeedsQuotes(s string) bool {
	for _, r := range s {
		if unicode.IsSpace(r) || r == '=' || r == '"' {
			return true
		}
	}
	return false
}

func consumeTagPart(s string) (string, string) {
	quoted := false
	backslashed := false
	part := ""
	done := false
	for i, r := range s {
		if i == 0 && r == '"' {
			quoted = true
		} else {
			if quoted {
				if backslashed {
					switch r {
					case '\\':
						part += "\\"
					case '"':
						part += "\""
					}
					backslashed = false
				} else {
					switch r {
					case '\\':
						backslashed = true
					case '"':
						done = true
						quoted = false
					default:
						part += string(r)
					}
				}
			} else {
				if r == '=' {
					return part, s[i+len("="):]
				} else if r == '|' {
					return part, s[i+len("|"):]
				} else if !done && !unicode.IsSpace(r) {
					part += string(r)
				}
			}
		}
	}
	return part, ""
}

func InvalidTag() Tag {
	return Tag{Value: StringExpression("")}
}

type Taggable interface {
	AllTags() Tags
	Get(key string) Tag
}

type Tags []Tag

func (t Tags) AllTags() Tags {
	return t
}

func (t Tags) Get(key string) Tag {
	for _, tag := range t {
		if tag.Key == key {
			return tag
		}
	}
	return InvalidTag()
}

func (t Tags) GetAt(key string, i int) Value {
	if i >= 0 {
		for _, tag := range t {
			if tag.Key == key {
				if values, ok := tag.Value.(Values); ok && len(values) > i {
					return values[i]
				}
			}
		}
	}
	return InvalidTag()
}

func (t Tags) TagOrFallback(key string, fallback string) Tag {
	if value := t.Get(key); value.IsValid() {
		return value
	}
	return Tag{Key: key, Value: StringExpression(fallback)}
}

func (t *Tags) SetTags(tags []Tag) {
	*t = tags
}

func (t *Tags) AddTag(tag Tag) {
	*t = append(*t, tag)
}

// Modifies an existing tag value, or add it if it doesn't exist.
// Returns (true, old value) if it modifies, or (false, undefined) if added.
func (t *Tags) ModifyOrAddTag(tag Tag) (bool, Value) {
	for i := range *t {
		if (*t)[i].Key == tag.Key {
			old := (*t)[i].Value
			(*t)[i].Value = tag.Value
			return true, old
		}
	}
	t.AddTag(tag)
	return false, StringExpression("")
}

func (t *Tags) ModifyOrAddTagAt(tag Tag, index int) (bool, Value) {
	for i := range *t {
		if (*t)[i].Key == tag.Key && (*t)[i].Value.ValueType() == ValueTypeValues {
			old := (*t)[i].Value
			(*t)[i].Value = Set((*t)[i].Value.(Values), tag.Value, index)
			return true, old
		}
	}
	t.AddTag(Tag{tag.Key, Set(make([]Value, 0), tag.Value, index)})
	return false, StringExpression("")
}

func (t *Tags) RemoveTag(key string) {
	for i, tag := range *t {
		if tag.Key == key {
			*t = append((*t)[:i], (*t)[i+1:]...)
		}
	}
}

func (t *Tags) RemoveTags(keys []string) {
	for i, tag := range *t {
		for _, key := range keys {
			if tag.Key == key {
				*t = append((*t)[:i], (*t)[i+1:]...)
			}
		}
	}
}

func (t *Tags) RemoveAllTags() {
	*t = []Tag{}
}

func (t Tags) Clone() Tags {
	clone := make([]Tag, len(t))
	copy(clone, t)
	return clone
}

func (t *Tags) MergeFrom(other Tags) {
	i := copy(*t, other)
	if i < len(other) {
		*t = append(*t, other[i:]...)
	} else {
		*t = (*t)[0:len(other)]
	}
}

type IndexedReference interface {
	Reference
	Index() int
	SetIndex(i int)
}

type IndexedFeatureID struct {
	FeatureID
	index int
}

func (r *IndexedFeatureID) Source() FeatureID {
	return r.FeatureID
}

func (r *IndexedFeatureID) Index() int {
	return r.index
}

func (r *IndexedFeatureID) SetIndex(i int) {
	r.index = i
}

func (t *Tags) References() []Reference {
	var refs []Reference
	if r := t.Get(PathTag); r.IsValid() && r.Value != nil && r.ValueType() == ValueTypeValues {
		for i, v := range r.Value.(Values) {
			if v, ok := v.(FeatureID); ok && v.IsValid() {
				refs = append(refs, &IndexedFeatureID{v, i})
			}
		}
	}

	return refs
}

func (t *Tags) Reference(i int) Reference {
	if id, ok := t.GetAt(PathTag, i).(Reference); ok {
		return id
	}
	return FeatureIDInvalid
}

type FeatureType int

const (
	FeatureTypePoint FeatureType = iota
	FeatureTypePath
	FeatureTypeArea
	FeatureTypeRelation
	FeatureTypeInvalid
	FeatureTypeCollection
	FeatureTypeExpression

	FeatureTypeBegin = FeatureTypePoint
	FeatureTypeEnd   = FeatureTypeInvalid
	FeatureTypeBits  = 2 // Bits necessary to represent all types up to, and excluding, invalid
)

func (f FeatureType) String() string {
	switch f {
	case FeatureTypePoint:
		return "point"
	case FeatureTypePath:
		return "path"
	case FeatureTypeArea:
		return "area"
	case FeatureTypeRelation:
		return "relation"
	case FeatureTypeCollection:
		return "collection"
	case FeatureTypeExpression:
		return "expression"
	default:
		return "invalid"
	}
}

func FeatureTypeFromString(s string) FeatureType {
	for t := FeatureTypeBegin; t <= FeatureTypeExpression; t++ { // TODO(mari): move down invalid index
		if s == t.String() {
			return t
		}
	}
	return FeatureTypeInvalid
}

type Namespace string

func (n Namespace) String() string {
	return string(n)
}

type Namespaces []Namespace

func (n Namespaces) Len() int           { return len(n) }
func (n Namespaces) Swap(i, j int)      { n[i], n[j] = n[j], n[i] }
func (n Namespaces) Less(i, j int) bool { return n[i] < n[j] }

const (
	// Features from OSM
	NamespaceOSMNode     Namespace = "openstreetmap.org/node"
	NamespaceOSMWay      Namespace = "openstreetmap.org/way"
	NamespaceOSMRelation Namespace = "openstreetmap.org/relation"

	// Used internally
	NamespacePrivate      Namespace = "diagonal.works/ns/private"
	NamespaceLatLng       Namespace = "diagonal.works/ns/ll"
	NamespaceMaterialised Namespace = "diagonal.works/ns/m"
	NamespaceUI           Namespace = "diagonal.works/ns/ui"

	// Used when connecting features to the street network
	NamespaceDiagonalEntrances    Namespace = "diagonal.works/ns/entrance"
	NamespaceDiagonalAccessPaths  Namespace = "diagonal.works/ns/access-path"
	NamespaceDiagonalAccessPoints Namespace = "diagonal.works/ns/access-point"

	NamespaceDiagonalUPRNCluster Namespace = "diagonal.works/ns/uprn-cluster"

	NamespaceUKONSBoundaries       Namespace = "statistics.gov.uk/datasets/regions"
	NamespaceGBUPRN                Namespace = "ordnancesurvey.co.uk/uprn"
	NamespaceGBOSTerrain50Contours Namespace = "ordnancesurvey.co.uk/terrain-50/contours"
	NamespaceGBOSOpenRoadsLinks    Namespace = "ordnancesurvey.co.uk/os-open-roads/links"
	NamespaceGBOSOpenRoadsNodes    Namespace = "ordnancesurvey.co.uk/os-open-roads/nodes"
	NamespaceGBOSMapBuildings      Namespace = "www.ordnancesurvey.co.uk/os-open-map-local/buildings"
	NamespaceGBCodePoint           Namespace = "ordnancesurvey.co.uk/code-point"

	// For GTFS transport data.
	NamespaceGTFS Namespace = "diagonal.works/ns/gtfs"

	NamespaceInvalid Namespace = ""
)

var StandardNamespaces = []Namespace{
	NamespaceOSMNode,
	NamespaceOSMWay,
	NamespaceOSMRelation,
	NamespaceLatLng,
	NamespaceDiagonalEntrances,
	NamespaceDiagonalAccessPaths,
	NamespaceDiagonalAccessPoints,
}

var OSMNamespaces = []Namespace{
	NamespaceOSMNode,
	NamespaceOSMWay,
	NamespaceOSMRelation,
}

type Identifiable interface {
	FeatureID() FeatureID
}

type FeatureID struct {
	Type      FeatureType
	Namespace Namespace
	Value     uint64
}

func (f FeatureID) IsValid() bool {
	return f.Namespace != NamespaceInvalid && f.Type != FeatureTypeInvalid
}

func (f FeatureID) Less(other FeatureID) bool {
	if f.Type == other.Type {
		if f.Namespace == other.Namespace {
			return f.Value < other.Value
		} else {
			return f.Namespace < other.Namespace
		}
	} else {
		return f.Type < other.Type
	}
}

func (f FeatureID) String() string {
	return fmt.Sprintf("%s/%s/%d", f.Type.String(), f.Namespace, f.Value)
}

func (f FeatureID) ValueType() ValueType {
	return ValueTypeFeatureID
}

func (f FeatureID) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.String())
}

func (f *FeatureID) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	*f = FeatureIDFromString(s)
	return nil
}

func (f FeatureID) MarshalYAML() (interface{}, error) {
	return "/" + f.String(), nil
}

func (f *FeatureID) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}
	if len(s) > 0 {
		*f = FeatureIDFromString(s[1:])
	} else {
		*f = FeatureIDInvalid
	}
	return nil
}

func (f FeatureID) FeatureID() FeatureID {
	return f
}

func (f *FeatureID) SetFeatureID(id FeatureID) {
	*f = id
}

func (f *FeatureID) FromFeatureID(other FeatureID) {
	f.Type = other.Type
	f.Namespace = other.Namespace
	f.Value = other.Value
}

func (f FeatureID) ToAreaID() AreaID {
	if f.Type == FeatureTypeArea || f.Type == FeatureTypeInvalid {
		return AreaID{Namespace: f.Namespace, Value: f.Value}
	}
	panic("Not a area")
}

func (f FeatureID) ToRelationID() RelationID {
	if f.Type == FeatureTypeRelation || f.Type == FeatureTypeInvalid {
		return RelationID{Namespace: f.Namespace, Value: f.Value}
	}
	panic("Not a relation")
}

func (f FeatureID) ToCollectionID() CollectionID {
	if f.Type == FeatureTypeCollection || f.Type == FeatureTypeInvalid {
		return CollectionID{Namespace: f.Namespace, Value: f.Value}
	}
	panic("Not a collection")
}

func (f FeatureID) ToExpressionID() ExpressionID {
	if f.Type == FeatureTypeExpression || f.Type == FeatureTypeInvalid {
		return ExpressionID{Namespace: f.Namespace, Value: f.Value}
	}
	panic("Not an expression")
}

type Reference interface {
	Source() FeatureID
}

func (f FeatureID) Source() FeatureID {
	return f
}

func (f FeatureID) Index() (int, error) {
	return -1, fmt.Errorf("index not available")
}

func (f FeatureID) SetIndex(i int) {}

func FeatureIDFromString(s string) FeatureID {
	if len(s) > 0 && s[0] == '/' {
		s = s[1:]
	}
	i := strings.Index(s, "/")
	j := strings.LastIndex(s, "/")
	if i < 0 || i == j {
		return FeatureIDInvalid
	}
	id := FeatureID{Type: FeatureTypeFromString(s[0:i])}
	if id.Type == FeatureTypeInvalid {
		return FeatureIDInvalid
	}
	id.Namespace = Namespace(s[i+1 : j])
	if v, err := strconv.ParseUint(s[j+1:], 10, 64); err == nil {
		id.Value = v
	} else {
		return FeatureIDInvalid
	}
	return id
}

var (
	FeatureIDInvalid = FeatureID{Type: FeatureTypeInvalid, Namespace: NamespaceInvalid}

	FeatureIDEnd = FeatureID{Type: FeatureTypeEnd, Namespace: NamespaceInvalid} // For sentinels

	FeatureIDPointBegin    = FeatureID{Type: FeatureTypePoint, Namespace: NamespaceInvalid, Value: 0}
	FeatureIDPathBegin     = FeatureID{Type: FeatureTypePath, Namespace: NamespaceInvalid, Value: 0}
	FeatureIDPointEnd      = FeatureIDPathBegin
	FeatureIDAreaBegin     = FeatureID{Type: FeatureTypeArea, Namespace: NamespaceInvalid, Value: 0}
	FeatureIDPathEnd       = FeatureIDAreaBegin
	FeatureIDRelationBegin = FeatureID{Type: FeatureTypeRelation, Namespace: NamespaceInvalid, Value: 0}
	FeatureIDAreaEnd       = FeatureIDRelationBegin
	FeatureIDRelationEnd   = FeatureIDEnd
)

type FeatureIDs []FeatureID

func (f FeatureIDs) Len() int           { return len(f) }
func (f FeatureIDs) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
func (f FeatureIDs) Less(i, j int) bool { return f[i].Less(f[j]) }

type ID = FeatureID

type AreaID struct {
	Namespace Namespace
	Value     uint64
}

func MakeAreaID(ns Namespace, v uint64) AreaID {
	return AreaID{Namespace: ns, Value: v}
}

func (a AreaID) FeatureID() FeatureID {
	return FeatureID{Type: FeatureTypeArea, Namespace: a.Namespace, Value: a.Value}
}

func (a AreaID) IsValid() bool {
	return a.Namespace != NamespaceInvalid
}

func (a AreaID) String() string {
	return a.FeatureID().String()
}

func (a AreaID) Less(other AreaID) bool {
	if a.Namespace == other.Namespace {
		return a.Value < other.Value
	} else {
		return a.Namespace < other.Namespace
	}
}

func (a *AreaID) FromFeatureID(other FeatureID) error {
	if other.Type != FeatureTypeArea {
		return errors.New("not a area ID")
	}
	a.Namespace = other.Namespace
	a.Value = other.Value
	return nil
}

func (a AreaID) MarshalYAML() (interface{}, error) {
	return a.FeatureID().MarshalYAML()
}

func (a *AreaID) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var id FeatureID
	err := unmarshal(&id)
	if err == nil {
		err = a.FromFeatureID(id)
	}
	return err
}

var AreaIDInvalid = AreaID{Namespace: NamespaceInvalid}

type RelationID struct {
	Namespace Namespace
	Value     uint64
}

func MakeRelationID(ns Namespace, v uint64) RelationID {
	return RelationID{Namespace: ns, Value: v}
}

func (r RelationID) FeatureID() FeatureID {
	return FeatureID{Type: FeatureTypeRelation, Namespace: r.Namespace, Value: r.Value}
}

func (r RelationID) IsValid() bool {
	return r.Namespace != NamespaceInvalid
}

func (r RelationID) String() string {
	return r.FeatureID().String()
}

func (r RelationID) Less(other RelationID) bool {
	if r.Namespace == other.Namespace {
		return r.Value < other.Value
	} else {
		return r.Namespace < other.Namespace
	}
}

func (r *RelationID) FromFeatureID(other FeatureID) error {
	if other.Type != FeatureTypeRelation {
		return errors.New("Not a relation ID")
	}
	r.Namespace = other.Namespace
	r.Value = other.Value
	return nil
}

func (r RelationID) MarshalYAML() (interface{}, error) {
	return r.FeatureID().MarshalYAML()
}

func (r *RelationID) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var id FeatureID
	err := unmarshal(&id)
	if err == nil {
		err = r.FromFeatureID(id)
	}
	return err
}

var RelationIDInvalid = RelationID{Namespace: NamespaceInvalid}

type CollectionID struct {
	Namespace Namespace
	Value     uint64
}

func MakeCollectionID(ns Namespace, v uint64) CollectionID {
	return CollectionID{Namespace: ns, Value: v}
}

func (c CollectionID) FeatureID() FeatureID {
	return FeatureID{Type: FeatureTypeCollection, Namespace: c.Namespace, Value: c.Value}
}

func (c CollectionID) IsValid() bool {
	return c.Namespace != NamespaceInvalid
}

func (c CollectionID) String() string {
	return c.FeatureID().String()
}

func (c CollectionID) Less(other CollectionID) bool {
	if c.Namespace == other.Namespace {
		return c.Value < other.Value
	} else {
		return c.Namespace < other.Namespace
	}
}

func (c *CollectionID) FromFeatureID(other FeatureID) error {
	if other.Type != FeatureTypeCollection {
		return errors.New("not a collection ID")
	}
	c.Namespace = other.Namespace
	c.Value = other.Value
	return nil
}

func (c CollectionID) MarshalYAML() (interface{}, error) {
	return c.FeatureID().MarshalYAML()
}

func (c *CollectionID) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var id FeatureID
	err := unmarshal(&id)
	if err == nil {
		err = c.FromFeatureID(id)
	}
	return err
}

type ExpressionID struct {
	Namespace Namespace
	Value     uint64
}

func MakeExpressionID(ns Namespace, v uint64) ExpressionID {
	return ExpressionID{Namespace: ns, Value: v}
}

func (e ExpressionID) FeatureID() FeatureID {
	return FeatureID{Type: FeatureTypeExpression, Namespace: e.Namespace, Value: e.Value}
}

func (e ExpressionID) IsValid() bool {
	return e.Namespace != NamespaceInvalid
}

func (e ExpressionID) String() string {
	return e.FeatureID().String()
}

func (e ExpressionID) Less(other CollectionID) bool {
	if e.Namespace == other.Namespace {
		return e.Value < other.Value
	} else {
		return e.Namespace < other.Namespace
	}
}

func (e *ExpressionID) FromFeatureID(other FeatureID) error {
	if other.Type != FeatureTypeExpression {
		return errors.New("not an expression ID")
	}
	e.Namespace = other.Namespace
	e.Value = other.Value
	return nil
}

func (e ExpressionID) MarshalYAML() (interface{}, error) {
	return e.FeatureID().MarshalYAML()
}

func (e *ExpressionID) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var id FeatureID
	err := unmarshal(&id)
	if err == nil {
		err = e.FromFeatureID(id)
	}
	return err
}

type GeometryType int

const (
	GeometryTypeInvalid GeometryType = iota
	GeometryTypePoint
	GeometryTypePath
	GeometryTypeArea
)

type Geometry interface {
	GeometryType() GeometryType

	Point() s2.Point

	GeometryLen() int
	PointAt(i int) s2.Point
	Polyline() *s2.Polyline

	ToGeoJSON() geojson.GeoJSON // TODO(mari): remove from interface / when u do wrapped generic feature
}

type InvalidGeometry struct{}

func (InvalidGeometry) GeometryType() GeometryType {
	return GeometryTypeInvalid
}

func (InvalidGeometry) Point() s2.Point {
	return s2.Point{}
}

func (InvalidGeometry) GeometryLen() int {
	return 0
}

func (InvalidGeometry) PointAt(i int) s2.Point {
	panic("InvalidGeometry has no points")
}

func (InvalidGeometry) Polyline() *s2.Polyline {
	panic("InvalidGeometry has no polyline")
}

func (InvalidGeometry) ToGeoJSON() geojson.GeoJSON {
	return geojson.NewFeatureCollection()
}

var _ Geometry = InvalidGeometry{}

const (
	PointTag = "point"
	PathTag  = "path"
)

func (t *Tags) GeometryType() GeometryType {
	if t.Get(PointTag) != InvalidTag() {
		return GeometryTypePoint
	} else if t.Get(PathTag) != InvalidTag() {
		return GeometryTypePath
	}

	return GeometryTypeInvalid
}

func (t *Tags) Point() s2.Point {
	ll, _ := LatLngFromString(t.Get(PointTag).Value.String())
	return s2.PointFromLatLng(ll)
}

func (t *Tags) GeometryLen() int {
	if v := t.Get(PointTag); v.IsValid() {
		return 1
	} else if v := t.Get(PathTag); v.IsValid() {
		if v, ok := v.Value.(Values); ok {
			return len(v)
		}
	}
	return 0
}

func (t *Tags) PointAt(i int) s2.Point {
	if ll, ok := t.GetAt(PathTag, i).(LatLng); ok {
		return s2.PointFromLatLng(s2.LatLng(ll))
	}

	return s2.Point{}
}

func (t *Tags) Polyline() *s2.Polyline {
	var ps []s2.Point
	for i := 0; i < t.GeometryLen(); i++ {
		ps = append(ps, t.PointAt(i))
	}
	return (*s2.Polyline)(&ps)
}

type Geo struct {
	point s2.Point
	path  []s2.Point
}

func (g Geo) GeometryType() GeometryType {
	if len(g.path) > 0 {
		return GeometryTypePath
	} else {
		return GeometryTypePoint
	}
}

func (g Geo) Point() s2.Point {
	return g.point
}

func (g Geo) PointAt(i int) s2.Point {
	return g.path[i]
}

func (g Geo) Polyline() *s2.Polyline {
	return (*s2.Polyline)(&g.path)
}

func (g Geo) GeometryLen() int {
	return len(g.path)
}

func (g Geo) ToGeoJSON() geojson.GeoJSON {
	return GeometryToGeoJSON(g)
}

func GeometryFromPoint(point s2.Point) Geometry {
	return Geo{point: point}
}

func GeometryFromPoints(points []s2.Point) Geometry {
	return Geo{path: points}
}

func GeometryFromLatLng(ll s2.LatLng) Geometry {
	return Geo{point: s2.PointFromLatLng(ll)}
}

func PhysicalFeatureToGeoJSON(f PhysicalFeature) *geojson.Feature {
	g := GeometryToGeoJSON(f)
	fillPropertiesFromTags(f, g)
	return g
}

func Covering(g Geometry, coverer s2.RegionCoverer) s2.CellUnion {
	switch g := g.(type) {
	case Area:
		covering := make(s2.CellUnion, 0)
		for i := 0; i < g.Len(); i++ {
			covering = s2.CellUnionFromUnion(covering, coverer.Covering(g.Polygon(i)))
		}
		return covering
	default:
		switch g.GeometryType() {
		case GeometryTypePoint:
			return coverer.Covering(g.Point())
		case GeometryTypePath:
			return coverer.Covering(g.Polyline())
		}
	}

	return s2.CellUnion{}
}

func Centroid(geometry Geometry) (s2.Point, bool) {
	switch geometry.GeometryType() {
	case GeometryTypePoint:
		return geometry.Point(), true
	case GeometryTypePath:
		return s2.Point{Vector: geometry.Polyline().Centroid().Normalize()}, true
	case GeometryTypeArea:
		if geometry.(Area).Len() == 1 {
			return s2.Point{Vector: geometry.(Area).Polygon(0).Loop(0).Centroid().Normalize()}, true
		} else {
			query := s2.NewConvexHullQuery()
			for i := 0; i < geometry.(Area).Len(); i++ {
				query.AddPolygon(geometry.(Area).Polygon(i))
			}
			return s2.Point{Vector: query.ConvexHull().Centroid().Normalize()}, true
		}
	}
	return s2.Point{}, false
}

func GeometryToGeoJSON(g Geometry) *geojson.Feature {
	switch g.GeometryType() {
	case GeometryTypePoint:
		return geojson.NewFeatureFromS2Point(g.Point())
	case GeometryTypePath:
		return geojson.NewFeatureWithGeometry(geojson.GeometryFromLineString(geojson.FromPolyline(g.Polyline())))
	default:
		panic("not implemented")
	}
}

type Feature interface {
	Identifiable
	Taggable
	References() []Reference
	Reference(i int) Reference
}

type PhysicalFeature interface {
	Feature
	Geometry
}

type NestedPhysicalFeature interface { // TODO(mari): rethink this
	PhysicalFeature
	Feature(i int) PhysicalFeature
}

type wrappedFeature struct {
	PhysicalFeature
	FeaturesByID
}

func (p wrappedFeature) Feature(i int) PhysicalFeature {
	if id := p.Reference(i).Source(); id.IsValid() {
		if point := p.FindFeatureByID(id); point != nil {
			return point.(PhysicalFeature)
		}
		panic(fmt.Sprintf("No point with ID %s", id))
	}
	return nil
}

func (p wrappedFeature) PointAt(i int) s2.Point {
	if point := p.PhysicalFeature.PointAt(i); point.Norm() != 0 {
		return point
	}

	id := p.Reference(i).Source()
	if id.IsValid() {
		if ll, err := p.FindLocationByID(id); err == nil {
			return s2.PointFromLatLng(ll)
		}
	}

	panic(fmt.Sprintf("No point with ID %s", id))
}

func (p wrappedFeature) Polyline() *s2.Polyline {
	points := make([]s2.Point, 0, p.GeometryLen())
	for i := 0; i < p.GeometryLen(); i++ {
		points = append(points, p.PointAt(i))
	}
	return (*s2.Polyline)(&points)
}

func WrapPhysicalFeature(f PhysicalFeature, features FeaturesByID) NestedPhysicalFeature {
	return wrappedFeature{f, features}
}

func (t Tags) ClosedPath() bool {
	start := t.Reference(0).Source()
	end := t.Reference(len(t.References()) - 1).Source()
	return start == end && start.IsValid()
}

type iterator struct {
	features []Feature
	i        int
}

func NewFeatureIterator(features []Feature) Features {
	return &iterator{features: features}
}

func (i *iterator) Feature() Feature {
	return i.features[i.i-1]
}

func (i *iterator) FeatureID() FeatureID {
	return i.features[i.i-1].FeatureID()
}

func (i *iterator) Next() bool {
	i.i++
	return i.i <= len(i.features)
}

type Area interface {
	Geometry
	Len() int
	Polygon(i int) *s2.Polygon
	MultiPolygon() geometry.MultiPolygon
}

type InvalidArea struct {
	InvalidGeometry
}

func (InvalidArea) Len() int {
	return 0
}

func (InvalidArea) Polygon(i int) *s2.Polygon {
	return s2.PolygonFromLoops([]*s2.Loop{})
}

func (InvalidArea) MultiPolygon() geometry.MultiPolygon {
	return geometry.MultiPolygon{}
}

type area struct {
	ps []*s2.Polygon
}

func (a area) Len() int {
	return len(a.ps)
}

func (a area) Polygon(i int) *s2.Polygon {
	return a.ps[i]
}

func (a area) MultiPolygon() geometry.MultiPolygon {
	m := make(geometry.MultiPolygon, a.Len())
	for i := 0; i < a.Len(); i++ {
		m[i] = a.Polygon(i)
	}
	return m
}

func (a area) GeometryType() GeometryType {
	return GeometryTypeArea
}

func (a area) Point() s2.Point {
	return s2.Point{}
}

func (a area) PointAt(i int) s2.Point {
	panic("not implemented")
}

func (a area) Polyline() *s2.Polyline {
	panic("not implemented")
}

func (a area) GeometryLen() int {
	panic("not implemented")
}

func (a area) ToGeoJSON() geojson.GeoJSON {
	return AreaToGeoJSON(a)
}

var _ Area = area{}

func AreaFromS2Loop(l *s2.Loop) Area {
	return AreaFromS2Polygon(s2.PolygonFromLoops([]*s2.Loop{l}))
}

func AreaFromS2Polygon(p *s2.Polygon) Area {
	return area{ps: []*s2.Polygon{p}}
}

func AreaFromS2Polygons(ps []*s2.Polygon) Area {
	return area{ps: ps}
}

func AreaToS2Polygons(a Area) []*s2.Polygon {
	ps := make([]*s2.Polygon, a.Len())
	for i := 0; i < a.Len(); i++ {
		ps[i] = a.Polygon(i)
	}
	return ps
}

type AreaFeature interface {
	Feature
	Area
	AreaID() AreaID
	Feature(i int) []NestedPhysicalFeature
}

type RelationMember struct {
	ID   FeatureID
	Role string
}

type RelationFeature interface {
	Feature
	RelationID() RelationID
	Len() int
	Member(i int) RelationMember
}

type CollectionFeature interface {
	Feature
	UntypedCollection
	CollectionID() CollectionID
	IsSortedByKey() bool
	FindValue(key any) (any, bool)
	FindValues(key any, values []any) []any
}

type ExpressionFeature interface {
	Feature
	ExpressionID() ExpressionID
	Expression() Expression
}

type Features interface {
	Feature() Feature
	FeatureID() FeatureID
	Next() bool
}

type SegmentKey struct {
	ID    FeatureID
	First int
	Last  int
}

func (s SegmentKey) ToPathSegment(path PhysicalFeature) Segment {
	return Segment{path, s.First, s.Last}
}

func (s SegmentKey) Less(other SegmentKey) bool {
	if s.ID == other.ID {
		if s.First == other.First {
			return s.Last < other.Last
		}
		return s.First < other.First
	}
	return s.ID.Less(other.ID)
}

type Segment struct {
	Feature PhysicalFeature
	First   int
	Last    int
}

func (s Segment) Len() int {
	if s.First < s.Last {
		return s.Last - s.First + 1
	} else {
		return s.First - s.Last + 1
	}
}

func (s Segment) ToKey() SegmentKey {
	return SegmentKey{ID: s.Feature.FeatureID(), First: s.First, Last: s.Last}
}

func (s Segment) pathIndex(i int) int {
	if s.First < s.Last {
		if s.First+i <= s.Last {
			return s.First + i
		}
	} else if s.First-i >= s.Last {
		return s.First - i
	}
	panic(fmt.Sprintf("Segment point %d out of range (first: %d, last: %d)", i, s.First, s.Last))
}

func (s Segment) SegmentPoint(i int) s2.Point {
	return s.Feature.PointAt(s.pathIndex(i))
}

func (s Segment) SegmentFeatureID(i int) FeatureID {
	return s.Feature.Reference(s.pathIndex(i)).Source()
}

func (s Segment) FirstFeatureID() FeatureID {
	return s.Feature.Reference(s.First).Source()
}

func (s Segment) LastFeatureID() FeatureID {
	return s.Feature.Reference(s.Last).Source()
}

func ToSegment(path PhysicalFeature) Segment {
	return Segment{path, 0, path.GeometryLen() - 1}
}

func (s Segment) Polyline() *s2.Polyline {
	polyline := *(s.Feature.Polyline())
	first, last := s.First, s.Last
	if first > last {
		first, last = last, first
	}
	segment := polyline[first : last+1]
	return &segment
}

var SegmentInvalid = Segment{Feature: nil}

type Segments interface {
	Segment() Segment
	Next() bool
}

type EmptySegments struct{}

func (EmptySegments) Segment() Segment {
	panic("No Segment")
}

func (EmptySegments) Next() bool {
	return false
}

func AllSegments(p Segments) []Segment {
	segments := make([]Segment, 0, 8)
	if p != nil {
		for p.Next() {
			segments = append(segments, p.Segment())
		}
	}
	return segments
}

func FindPathSegmentByKey(key SegmentKey, w World) Segment {
	return Segment{
		Feature: w.FindFeatureByID(key.ID).(PhysicalFeature),
		First:   key.First,
		Last:    key.Last,
	}
}

type Step struct {
	Destination FeatureID
	Via         FeatureID
	Cost        float64
}

func (s *Step) ToSegment(previous FeatureID, w World) Segment {
	// TODO: For long paths, it could be more efficient to use
	// FindPathsByPoint
	path := w.FindFeatureByID(s.Via).(PhysicalFeature)
	if path == nil {
		return SegmentInvalid
	}

	for i := 0; i < path.GeometryLen(); i++ {
		if point := path.Reference(i).Source(); point.IsValid() {
			if point == previous {
				if point == s.Destination {
					return Segment{Feature: path, First: i, Last: i}
				}
				for j := i; j < path.GeometryLen(); j++ {
					if next := w.FindFeatureByID(path.Reference(j).Source()); next != nil && next.FeatureID() == s.Destination {
						return Segment{Feature: path, First: i, Last: j}
					}
				}
			} else if point == s.Destination {
				for j := i; j < path.GeometryLen(); j++ {
					if next := w.FindFeatureByID(path.Reference(j).Source()); next != nil && next.FeatureID() == previous {
						return Segment{Feature: path, First: j, Last: i}
					}
				}
			}
		}
	}

	return SegmentInvalid
}

type Route struct {
	Origin FeatureID
	Steps  []Step
}

func (r *Route) ToSegments(w World) []Segment {
	segments := make([]Segment, len(r.Steps))
	previous := r.Origin
	for i, step := range r.Steps {
		segments[i] = step.ToSegment(previous, w)
		previous = step.Destination
	}
	return segments
}

type LocationsByID interface {
	FindLocationByID(id FeatureID) (s2.LatLng, error)
}

type FeaturesByID interface {
	LocationsByID
	FindFeatureByID(id FeatureID) Feature
	HasFeatureWithID(id FeatureID) bool
}

type EachFeatureOptions struct {
	SkipPoints      bool
	SkipPaths       bool
	SkipAreas       bool
	SkipRelations   bool
	SkipCollections bool
	SkipExpressions bool

	FeedReferencesFirst bool
	Goroutines          int
}

func (e *EachFeatureOptions) IsSkipped(t FeatureType) bool {
	switch t {
	case FeatureTypePoint:
		return e.SkipPoints
	case FeatureTypePath:
		return e.SkipPaths
	case FeatureTypeArea:
		return e.SkipAreas
	case FeatureTypeRelation:
		return e.SkipRelations
	case FeatureTypeCollection:
		return e.SkipCollections
	}
	return false
}

type World interface {
	// TODO: Include transit once our use of it has stabalised
	FindFeatureByID(id FeatureID) Feature
	HasFeatureWithID(id FeatureID) bool
	FindLocationByID(id FeatureID) (s2.LatLng, error)
	// TODO: make the query type more specific to Features, similar to the level in api.proto
	FindFeatures(query Query) Features
	FindRelationsByFeature(id FeatureID) RelationFeatures
	FindCollectionsByFeature(id FeatureID) CollectionFeatures
	FindAreasByPoint(id FeatureID) AreaFeatures
	FindReferences(id FeatureID, typed ...FeatureType) Features
	Traverse(id FeatureID) Segments
	EachFeature(each func(f Feature, goroutine int) error, options *EachFeatureOptions) error

	// Returns a copy of all tokens known to this world's search index. The
	// order isn't defined.
	Tokens() []string
}

type EmptyWorld struct{}

func (EmptyWorld) FindFeatureByID(id FeatureID) Feature {
	return nil
}

func (EmptyWorld) HasFeatureWithID(id FeatureID) bool {
	return false
}

func (EmptyWorld) FindLocationByID(id FeatureID) (s2.LatLng, error) {
	return s2.LatLng{}, fmt.Errorf("world is empty")
}

func (EmptyWorld) FindFeatures(query Query) Features {
	return EmptyFeatures{}
}

func (EmptyWorld) FindRelationsByFeature(id FeatureID) RelationFeatures {
	return EmptyRelationFeatures{}
}

func (EmptyWorld) FindCollectionsByFeature(id FeatureID) CollectionFeatures {
	return EmptyCollectionFeatures{}
}

func (EmptyWorld) FindAreasByPoint(id FeatureID) AreaFeatures {
	return EmptyAreaFeatures{}
}

func (EmptyWorld) FindReferences(id FeatureID, typed ...FeatureType) Features {
	return EmptyFeatures{}
}

func (EmptyWorld) Traverse(id FeatureID) Segments {
	return EmptySegments{}
}

func (EmptyWorld) EachFeature(each func(f Feature, goroutine int) error, options *EachFeatureOptions) error {
	return nil
}

func (EmptyWorld) Tokens() []string {
	return []string{}
}

type EmptyFeatures struct{}

func (EmptyFeatures) Feature() Feature {
	panic("No Features")
}

func (EmptyFeatures) FeatureID() FeatureID {
	panic("No Features")
}

func (EmptyFeatures) Next() bool {
	return false
}

func AllFeatures(f Features) []Feature {
	features := make([]Feature, 0, 8)
	for f.Next() {
		features = append(features, f.Feature())
	}
	return features
}

func FindAreaByID(id AreaID, features FeaturesByID) AreaFeature {
	if area := features.FindFeatureByID(id.FeatureID()); area != nil {
		return area.(AreaFeature)
	}
	return nil
}

type areaFeatures struct {
	features Features
}

func (a areaFeatures) Next() bool {
	return a.features.Next()
}

func (a areaFeatures) Feature() AreaFeature {
	if area, ok := a.features.Feature().(AreaFeature); ok {
		return area
	}
	panic(fmt.Sprintf("Not an AreaFeature: %T", a.features.Feature()))
}

func (a areaFeatures) FeatureID() FeatureID {
	return a.features.FeatureID()
}

type AreaFeatures interface {
	Feature() AreaFeature
	FeatureID() FeatureID
	Next() bool
}

type EmptyAreaFeatures struct{}

func (EmptyAreaFeatures) Feature() AreaFeature {
	panic("No AreaFeatures")
}

func (EmptyAreaFeatures) FeatureID() FeatureID {
	panic("No AreaFeatures")
}

func (EmptyAreaFeatures) AreaID() AreaID {
	panic("No AreaFeatures")
}

func (EmptyAreaFeatures) Next() bool {
	return false
}

func AllAreas(a AreaFeatures) []AreaFeature {
	features := make([]AreaFeature, 0, 8)
	if a != nil {
		for a.Next() {
			features = append(features, a.Feature())
		}
	}
	return features
}

func NewAreaFeatures(features Features) AreaFeatures {
	return areaFeatures{features: features}
}

func FindRelationByID(id RelationID, features FeaturesByID) RelationFeature {
	if relation := features.FindFeatureByID(id.FeatureID()); relation != nil {
		return relation.(RelationFeature)
	}
	return nil
}

type relationFeatures struct {
	features Features
}

func (r relationFeatures) Next() bool {
	return r.features.Next()
}

type RelationFeatures interface {
	Feature() RelationFeature
	FeatureID() FeatureID
	RelationID() RelationID
	Next() bool
}

func (r relationFeatures) Feature() RelationFeature {
	if relation, ok := r.features.Feature().(RelationFeature); ok {
		return relation
	}
	panic(fmt.Sprintf("Not an RelationFeature: %T", r.features.Feature()))
}

func (r relationFeatures) FeatureID() FeatureID {
	return r.features.FeatureID()
}

func (r relationFeatures) RelationID() RelationID {
	return r.FeatureID().ToRelationID()
}

type EmptyRelationFeatures struct{}

func (EmptyRelationFeatures) Feature() RelationFeature {
	panic("No RelationFeatures")
}

func (EmptyRelationFeatures) FeatureID() FeatureID {
	panic("No RelationFeatures")
}

func (EmptyRelationFeatures) RelationID() RelationID {
	panic("No RelationFeatures")
}

func (EmptyRelationFeatures) Next() bool {
	return false
}

func AllRelations(r RelationFeatures) []RelationFeature {
	features := make([]RelationFeature, 0, 8)
	if r != nil {
		for r.Next() {
			features = append(features, r.Feature())
		}
	}
	return features
}

func NewRelationFeatures(features Features) RelationFeatures {
	return relationFeatures{features: features}
}

type EmptyCollectionFeatures struct{}

func (EmptyCollectionFeatures) Feature() CollectionFeature {
	panic("No CollectionFeatures")
}

func (EmptyCollectionFeatures) FeatureID() FeatureID {
	panic("No CollectionFeatures")
}

func (EmptyCollectionFeatures) CollectionID() CollectionID {
	panic("No CollectionFeatures")
}

func (EmptyCollectionFeatures) Next() bool {
	return false
}

func FindCollectionByID(id CollectionID, features FeaturesByID) CollectionFeature {
	if collection := features.FindFeatureByID(id.FeatureID()); collection != nil {
		return collection.(CollectionFeature)
	}
	return nil
}

func FindExpressionByID(id ExpressionID, features FeaturesByID) ExpressionFeature {
	if expression := features.FindFeatureByID(id.FeatureID()); expression != nil {
		return expression.(ExpressionFeature)
	}
	return nil
}

func FindAreas(q Query, w World) AreaFeatures {
	q = Typed{Type: FeatureTypeArea, Query: q}
	return NewAreaFeatures(w.FindFeatures(q))
}

func FindRelations(q Query, w World) RelationFeatures {
	q = Typed{Type: FeatureTypeRelation, Query: q}
	return NewRelationFeatures(w.FindFeatures(q))
}

type CollectionFeatures interface {
	Feature() CollectionFeature
	FeatureID() FeatureID
	CollectionID() CollectionID
	Next() bool
}

type collectionFeatures struct {
	features Features
}

func (c collectionFeatures) Feature() CollectionFeature {
	if collection, ok := c.features.Feature().(CollectionFeature); ok {
		return collection
	}
	panic(fmt.Sprintf("Not an CollectionFeature: %T", c.features.Feature()))
}

func (c collectionFeatures) FeatureID() FeatureID {
	return c.features.FeatureID()
}

func (c collectionFeatures) CollectionID() CollectionID {
	return c.FeatureID().ToCollectionID()
}

func (c collectionFeatures) Next() bool {
	return c.features.Next()
}

func NewCollectionFeatures(features Features) CollectionFeatures {
	return collectionFeatures{features: features}
}

func FindCollections(q Query, w World) CollectionFeatures {
	q = Typed{Type: FeatureTypeCollection, Query: q}
	return NewCollectionFeatures(w.FindFeatures(q))
}

func fillPropertiesFromTags(t Taggable, feature *geojson.Feature) {
	for _, tag := range t.AllTags() {
		feature.Properties[tag.Key] = tag.Value.String()
	}
}

func AreaToGeoJSON(area Area) *geojson.Feature {
	coordinates := make([][][]geojson.Coordinate, area.Len())
	for i := 0; i < area.Len(); i++ {
		polygon := area.Polygon(i)
		coordinates[i] = geojson.FromPolygon(polygon)
	}
	var geometry geojson.Geometry
	if len(coordinates) == 1 {
		geometry = geojson.GeometryFromPolygon(coordinates[0])
	} else {
		geometry = geojson.GeometryFromMultiPolygon(coordinates)
	}
	return geojson.NewFeatureWithGeometry(geometry)
}

func AreaFeatureToGeoJSON(area AreaFeature) *geojson.Feature {
	g := AreaToGeoJSON(area)
	fillPropertiesFromTags(area, g)
	return g
}

func RelationFeatureToGeoJSON(relation RelationFeature, byID FeaturesByID) *geojson.FeatureCollection {
	collection := geojson.NewFeatureCollection()
	for i := 0; i < relation.Len(); i++ {
		if f := byID.FindFeatureByID(relation.Member(i).ID); f != nil {
			if r, ok := f.(Geometry); ok {
				collection.Add(r.ToGeoJSON())
			}
		}
	}
	return collection
}

func CollectionFeatureToGeoJSON(collection CollectionFeature, byID FeaturesByID) *geojson.FeatureCollection {
	geojson := geojson.NewFeatureCollection()

	i := collection.BeginUntyped()
	for {
		ok, err := i.Next()
		if !ok || err != nil {
			break
		}

		if id, ok := i.Key().(Identifiable); ok {
			if f := byID.FindFeatureByID(id.FeatureID()); f != nil {
				if r, ok := f.(Geometry); ok {
					geojson.Add(r.ToGeoJSON())
				}
			}
		}
	}

	return geojson
}

func AngleToMeters(angle s1.Angle) float64 {
	return units.AngleToMeters(angle)
}

func MetersToAngle(meters float64) s1.Angle {
	return units.MetersToAngle(meters)
}

func AreaToMeters2(area float64) float64 {
	return units.AreaToMeters2(area)
}

func Meters2ToArea(m2 float64) float64 {
	return units.Meters2ToArea(m2)
}

func LatLngFromString(s string) (s2.LatLng, error) {
	parts := strings.SplitN(s, ",", 2)
	if len(parts) == 2 {
		lat, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
		if err == nil {
			lng, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
			if err == nil {
				return s2.LatLngFromDegrees(lat, lng), nil
			}
		}
	}
	return s2.LatLng{}, fmt.Errorf("invalid lat,lng: %s", s)
}

func LatLngToString(ll s2.LatLng) string {
	return strconv.FormatFloat(ll.Lat.Degrees(), 'f', -1, 64) + "," + strconv.FormatFloat(ll.Lng.Degrees(), 'f', -1, 64)
}
