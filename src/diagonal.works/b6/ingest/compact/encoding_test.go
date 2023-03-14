package compact

import (
	"math/rand"
	"reflect"
	"sort"
	"testing"

	"diagonal.works/b6"
	"diagonal.works/b6/encoding"
	"diagonal.works/b6/ingest"
	"github.com/golang/geo/s2"
)

func newOSMNamespaces() (*Namespaces, *NamespaceTable) {
	var nt NamespaceTable
	nt.FillFromNamespaces(append(b6.OSMNamespaces, b6.NamespacePrivate))

	var nss Namespaces
	nss[b6.FeatureTypePoint] = nt.Encode(b6.NamespaceOSMNode)
	nss[b6.FeatureTypePath] = nt.Encode(b6.NamespaceOSMWay)
	nss[b6.FeatureTypeArea] = nt.Encode(b6.NamespaceOSMWay)
	nss[b6.FeatureTypeRelation] = nt.Encode(b6.NamespaceOSMRelation)
	return &nss, &nt
}

func TestLatLngsEncoding(t *testing.T) {
	towpath := []s2.LatLng{
		s2.LatLngFromDegrees(51.5367727, -0.1282827),
		s2.LatLngFromDegrees(51.5357103, -0.1272800),
		s2.LatLngFromDegrees(51.5353655, -0.1236310),
	}
	ls := make(LatLngs, len(towpath))
	for i, l := range towpath {
		ls[i] = LatLng{LatE7: l.Lat.E7(), LngE7: l.Lng.E7()}
	}

	var buffer [128]byte
	n := ls.Marshal(buffer[0:])

	var lls LatLngs
	nn := lls.Unmarshal(buffer[0:])

	if !reflect.DeepEqual(ls, lls) {
		t.Errorf("Expected %+v, found %+v", ls, lls)
	}
	if n != nn {
		t.Errorf("Expected marshalled and unmarshaled lengths to be equal (%d vs %d)", n, nn)
	}

	simple := 0
	for _, l := range ls {
		simple += l.Marshal(buffer[0:])
	}
	if n > simple {
		t.Errorf("Expected delta encoding to be more compact than explicit encoding (%d vs %d)", n, simple)
	}
}

func TestTagEncoding(t *testing.T) {
	s := map[int]string{
		1:    "highway",
		42:   "primary",
		36:   "bicycle",
		2043: "designated",
	}

	ts := Tags{
		{Key: 1, Value: 42},
		{Key: 36, Value: 2043},
	}

	var buffer [128]byte
	n := ts.Marshal(buffer[0:])

	var tts Tags
	nn := tts.Unmarshal(buffer[0:n])

	if !reflect.DeepEqual(ts, tts) {
		t.Errorf("Expected %+v, found %+v", ts, tts)
	}
	if n != nn {
		t.Errorf("Expected marshalled and unmarshaled lengths to be equal (%d vs %d)", n, nn)
	}
	if l := (MarshalledTags{Tags: buffer[0:]}).Length(); l != n {
		t.Errorf("Expected end at %d, found %d", n, l)
	}

	m := MarshalledTags{Tags: buffer[0:n], Strings: encoding.StringMap(s)}
	expected := []b6.Tag{{Key: "highway", Value: "primary"}, {Key: "bicycle", Value: "designated"}}
	if found := m.AllTags(); !reflect.DeepEqual(found, expected) {
		t.Errorf("Expected %+v, found %+v", expected, found)
	}

	if highway := m.Get("highway"); highway.Value != "primary" {
		t.Errorf("Expected to find highway=primary")
	}

	if building := m.Get("building"); building.IsValid() {
		t.Errorf("Didn't expect to find building tag")
	}
}

func TestPointEncoding(t *testing.T) {
	ll := s2.LatLngFromDegrees(51.53532, -0.12521)
	p := Point{
		Location: LatLng{
			LatE7: ll.Lat.E7(),
			LngE7: ll.Lng.E7(),
		},
		Tags: []Tag{
			{Key: 1, Value: 42},
			{Key: 36, Value: 2043},
		},
	}

	var buffer [128]byte
	n := p.Marshal(buffer[0:])

	var pp Point
	nn := pp.Unmarshal(buffer[0:n])

	if !reflect.DeepEqual(p, pp) {
		t.Errorf("Expected %+v, found %+v", p, pp)
	}
	if n != nn {
		t.Errorf("Expected marshalled and unmarshaled lengths to be equal (%d vs %d)", n, nn)
	}

	found := MarshalledPoint(buffer[0:]).Location()
	if s2.PointFromLatLng(found).Distance(s2.PointFromLatLng(ll)) > b6.MetersToAngle(1) {
		t.Errorf("Expected location %s, found %s", ll, found)
	}
}

func TestCommonPointEncoding(t *testing.T) {
	nss, nt := newOSMNamespaces()
	ll := s2.LatLngFromDegrees(51.53532, -0.12521)
	p := CommonPoint{
		Point: Point{
			Location: LatLng{
				LatE7: ll.Lat.E7(),
				LngE7: ll.Lng.E7(),
			},
			Tags: []Tag{
				{Key: 1, Value: 42},
				{Key: 36, Value: 2043},
			},
		},
		Path: Reference{
			Namespace: nt.Encode(b6.NamespaceOSMNode),
			Value:     544908186,
		},
	}

	var buffer [128]byte
	n := p.Marshal(nss, buffer[0:])

	var pp CommonPoint
	nn := pp.Unmarshal(nss, buffer[0:n])

	if !reflect.DeepEqual(p, pp) {
		t.Errorf("Expected %+v, found %+v", p, pp)
	}
	if n != nn {
		t.Errorf("Expected marshalled and unmarshaled lengths to be equal (%d vs %d)", n, nn)
	}

	found := MarshalledPoint(buffer[0:]).Location()
	if s2.PointFromLatLng(found).Distance(s2.PointFromLatLng(ll)) > b6.MetersToAngle(1) {
		t.Errorf("Expected location %s, found %s", ll, found)
	}
}

func TestReferenceEncoding(t *testing.T) {
	nss, nt := newOSMNamespaces()
	examples := []Reference{
		{Namespace: nt.Encode(b6.NamespaceOSMNode), Value: 5266980038},
		{Namespace: nt.Encode(b6.NamespaceOSMNode), Value: 9223372042121755846},
		{Namespace: nt.Encode(b6.NamespaceOSMWay), Value: 544908186},
		{Namespace: nt.Encode(b6.NamespaceOSMWay), Value: 9223372037399683994},
	}

	var buffer [128]byte
	for _, e := range examples {
		n := e.Marshal(nss.ForType(b6.FeatureTypePoint), buffer[0:])
		var r Reference
		nn := r.Unmarshal(nss.ForType(b6.FeatureTypePoint), buffer[0:n])
		if !reflect.DeepEqual(e, r) {
			t.Errorf("Expected %+v, found %+v", &e, &r)
		}
		if n != nn {
			t.Errorf("Expected marshalled and unmarshaled lengths to be equal (%d vs %d)", n, nn)
		}
	}
}

func TestReferencesEncoding(t *testing.T) {
	nss, nt := newOSMNamespaces()
	r := References{
		{Namespace: nt.Encode(b6.NamespaceOSMNode), Value: 544908185},
		{Namespace: nt.Encode(b6.NamespaceOSMNode), Value: 544908184},
		{Namespace: nt.Encode(b6.NamespaceOSMNode), Value: 544908186},
		{Namespace: nt.Encode(b6.NamespaceOSMNode), Value: 544908182},
	}

	var buffer [128]byte
	n1 := r.Marshal(nss.ForType(b6.FeatureTypePoint), buffer[0:])
	var rr References
	rr.Unmarshal(nss.ForType(b6.FeatureTypePoint), buffer[0:n1])
	if !reflect.DeepEqual(r, rr) {
		t.Errorf("Expected %+v, found %+v", r, rr)
	}
	if n1 >= 8*len(r) {
		t.Errorf("Expected a shorter encoding")
	}
	if l := MarshalledReferences(buffer[0:]).Len(); l != len(r) {
		t.Errorf("Expected length %d, found %d", len(r), l)
	}

	r = References{
		{Namespace: nt.Encode(b6.NamespaceOSMNode), Value: 544908185},
		{Namespace: nt.Encode(b6.NamespaceOSMRelation), Value: 544908184},
		{Namespace: nt.Encode(b6.NamespaceOSMWay), Value: 544908186},
		{Namespace: nt.Encode(b6.NamespaceOSMNode), Value: 544908182},
	}

	n2 := r.Marshal(nss.ForType(b6.FeatureTypePoint), buffer[0:])
	rr.Unmarshal(nss.ForType(b6.FeatureTypePoint), buffer[0:n2])
	if !reflect.DeepEqual(r, rr) {
		t.Errorf("Expected %+v, found %+v", r, rr)
	}

	if n1 >= 8*len(r) {
		t.Errorf("Expected a shorter encoding")
	}
	if n1 >= n2 {
		t.Errorf("Expected primary namespace to have more compact encoding (%d >= %d)", n1, n2)
	}
}

func TestReferencesEncodingWithLargeDeltas(t *testing.T) {
	const maxUint64 = 0xffffffffffffffff
	nss, nt := newOSMNamespaces()
	r := References{
		{Namespace: nt.Encode(b6.NamespaceOSMNode), Value: 1},
		{Namespace: nt.Encode(b6.NamespaceOSMNode), Value: maxUint64 - 1},
		{Namespace: nt.Encode(b6.NamespaceOSMNode), Value: 0},
		{Namespace: nt.Encode(b6.NamespaceOSMNode), Value: maxUint64},
	}

	var buffer [128]byte
	n := r.Marshal(nss.ForType(b6.FeatureTypePoint), buffer[0:])
	var rr References
	rr.Unmarshal(nss.ForType(b6.FeatureTypePoint), buffer[0:n])
	if !reflect.DeepEqual(r, rr) {
		t.Errorf("Expected %+v, found %+v", r, rr)
	}
}

func TestCombineGeometryEncodingAndLength(t *testing.T) {
	tests := []struct {
		e GeometryEncoding
		l int
	}{
		{GeometryEncodingReferences, 42},
		{GeometryEncodingLatLngs, 43},
		{GeometryEncodingMixed, 44},
	}

	var buffer [128]byte
	for _, test := range tests {
		n := MarshalGeometryEncodingAndLength(test.e, test.l, buffer[0:])
		e, l, i := UnmarshalGeometryEncodingAndLength(buffer[0:n])
		if e != test.e {
			t.Errorf("Expected GeometryEncoding %d, found %d", test.e, e)
		}
		if l != test.l {
			t.Errorf("Expected length %d, found %d", test.l, l)
		}
		if i != n {
			t.Errorf("Marshaled and unmarshaled buffer lengths differ (%d vs %d)", i, n)
		}
	}
}

func TestBitsEncoding(t *testing.T) {
	lengths := []int{0, 4, 33, 66, 64, 128}
	r := rand.New(rand.NewSource(42))
	for _, l := range lengths {
		b := make(Bits, l)
		for i := range b {
			b[i] = r.Float32() > 0.5
		}
		var buffer [128]byte
		n := b.Marshal(buffer[0:])
		bb := make(Bits, 0)
		nn := bb.Unmarshal(buffer[0:n])
		if !reflect.DeepEqual(b, bb) {
			t.Errorf("Expected %+v, found %+v, testing length %d", b, bb, l)
		}
		if n != nn {
			t.Errorf("Expected marshalled and unmarshaled lengths to be equal (%d vs %d) testing length %d", n, nn, l)
		}
	}
}

func TestGeometryMixedEncoding(t *testing.T) {
	_, nt := newOSMNamespaces()
	g := PathGeometryMixed{
		Points: []ReferenceAndLatLng{
			ReferenceAndLatLng{
				Reference: Reference{
					Namespace: nt.Encode(b6.NamespaceOSMNode), Value: 5378333638,
				},
			},
			ReferenceAndLatLng{
				LatLng: LatLngFromDegrees(51.5364858, -0.1279054),
			},
			ReferenceAndLatLng{
				Reference: Reference{
					Namespace: nt.Encode(b6.NamespaceOSMNode), Value: 7787634209,
				},
			},
			ReferenceAndLatLng{
				LatLng: LatLngFromDegrees(51.5351683, -0.1268059),
			},
			ReferenceAndLatLng{
				Reference: Reference{
					Namespace: nt.Encode(b6.NamespaceOSMNode), Value: 2512646902,
				},
			},
			ReferenceAndLatLng{
				Reference: Reference{
					Namespace: nt.Encode(b6.NamespacePrivate), Value: 42,
				},
			},
		},
	}

	var buffer [128]byte
	n := g.Marshal(nt.Encode(b6.NamespaceOSMNode), buffer[0:])
	var gg PathGeometryMixed
	nn := gg.Unmarshal(nt.Encode(b6.NamespaceOSMNode), buffer[0:n])
	if len(g.Points) != len(g.Points) {
		t.Errorf("Expected length %d, found %d", len(g.Points), len(gg.Points))
	} else {
		for i := range g.Points {
			if !reflect.DeepEqual(g.Points[i], gg.Points[i]) {
				t.Errorf("Expected %+v, found %+v at index %d", g.Points[i], gg.Points[i], i)
			}
		}
	}
	if n != nn {
		t.Errorf("Expected marshalled and unmarshaled lengths to be equal (%d vs %d)", n, nn)
	}

}

func TestFullPointEncoding(t *testing.T) {
	nss, nt := newOSMNamespaces()
	ll := s2.LatLngFromDegrees(51.53532, -0.12521)
	p := FullPoint{
		Point: Point{
			Location: LatLng{
				LatE7: ll.Lat.E7(),
				LngE7: ll.Lng.E7(),
			},
			Tags: []Tag{
				{Key: 1, Value: 42},
				{Key: 36, Value: 2043},
			},
		},
		PointReferences: PointReferences{
			Paths: References{
				{Namespace: nt.Encode(b6.NamespaceOSMWay), Value: 544908185},
				{Namespace: nt.Encode(b6.NamespaceOSMWay), Value: 544908184},
				{Namespace: nt.Encode(b6.NamespaceOSMWay), Value: 544908186},
			},
			Relations: References{
				{Namespace: nt.Encode(b6.NamespaceOSMRelation), Value: 544908182},
			},
		},
	}

	var buffer [128]byte
	n := p.Marshal(nss, buffer[0:])
	var pp FullPoint
	nn := pp.Unmarshal(nss, buffer[0:n])
	if !reflect.DeepEqual(p, pp) {
		t.Errorf("Expected %+v, found %+v", p, pp)
	}
	if n != nn {
		t.Errorf("Expected marshalled and unmarshaled lengths to be equal (%d vs %d)", n, nn)
	}
}

func TestPathEncoding(t *testing.T) {
	nss, nt := newOSMNamespaces()
	p := Path{
		Tags: []Tag{
			{Key: 1, Value: 42},
			{Key: 36, Value: 2043},
		},
		Points: &PathGeometryReferences{
			Points: References{
				{Namespace: nt.Encode(b6.NamespaceOSMNode), Value: 5266980038},
				{Namespace: nt.Encode(b6.NamespaceOSMNode), Value: 5266980022},
				{Namespace: nt.Encode(b6.NamespaceOSMNode), Value: 5266980036},
				{Namespace: nt.Encode(b6.NamespaceOSMNode), Value: 5266980031},
				{Namespace: nt.Encode(b6.NamespaceOSMNode), Value: 5266980038},
			},
		},
		Areas: References{
			{Namespace: nt.Encode(b6.NamespaceOSMWay), Value: 544908186},
			{Namespace: nt.Encode(b6.NamespaceOSMRelation), Value: 7972217},
		},
		Relations: References{
			{Namespace: nt.Encode(b6.NamespaceOSMRelation), Value: 7216547},
		},
	}

	var buffer [128]byte
	n := p.Marshal(nss, buffer[0:])
	var pp Path
	nn := pp.Unmarshal(nss, buffer[0:n])
	if !reflect.DeepEqual(p, pp) {
		t.Errorf("Expected %+v, found %+v", p, pp)
	}
	if n != nn {
		t.Errorf("Expected marshalled and unmarshaled lengths to be equal (%d vs %d)", n, nn)
	}

	m := MarshalledPath(buffer[0:])
	if m.Len() != p.Len() {
		t.Errorf("Expected MarshalledPath and Path lengths to be equal (%d vs %d)", m.Len(), p.Len())
	}
}

func TestAreaEncoding(t *testing.T) {
	nss, nt := newOSMNamespaces()
	a := Area{
		Tags: []Tag{
			{Key: 1, Value: 42},
			{Key: 36, Value: 2043},
		},
		Polygons: &AreaGeometryReferences{
			Polygons: []int{1, 2},
			Paths: References{
				{Namespace: nt.Encode(b6.NamespaceOSMWay), Value: 544908185},
				{Namespace: nt.Encode(b6.NamespaceOSMWay), Value: 544908184},
				{Namespace: nt.Encode(b6.NamespaceOSMWay), Value: 544908186},
			},
		},
	}

	var buffer [128]byte
	n := a.Marshal(nss, buffer[0:])
	var aa Area
	nn := aa.Unmarshal(nss, buffer[0:n])
	if !reflect.DeepEqual(a, aa) {
		t.Errorf("Expected %+v, found %+v", a, aa)
	}
	if n != nn {
		t.Errorf("Expected marshalled and unmarshaled lengths to be equal (%d vs %d)", n, nn)
	}
	g := a.Polygons.(*AreaGeometryReferences)
	if l := MarshalledArea(buffer[0:]).Len(); l != len(g.Polygons)+1 {
		t.Errorf("Expected %d polygons, found %d", len(g.Polygons)+1, l)
	}
}

func TestAreaEncodingLatLngs(t *testing.T) {
	nss, _ := newOSMNamespaces()
	a := Area{
		Tags: []Tag{
			{Key: 1, Value: 42},
			{Key: 36, Value: 2043},
		},
		Polygons: &AreaGeometryLatLngs{
			Polygons: []PolygonGeometryLatLngs{
				{
					Loops: []int{4},
					Points: LatLngs{
						LatLngFromDegrees(51.5235396, -0.1251689), // Outer
						LatLngFromDegrees(51.5231971, -0.1249710),
						LatLngFromDegrees(51.5233405, -0.1243188),
						LatLngFromDegrees(51.5236851, -0.1245202),

						LatLngFromDegrees(51.5235250, -0.1246718), // Inner
						LatLngFromDegrees(51.5233982, -0.1245960),
						LatLngFromDegrees(51.5233456, -0.1248273),
						LatLngFromDegrees(51.5234767, -0.1249024),
					},
				},
			},
		},
	}

	var buffer [128]byte
	n := a.Marshal(nss, buffer[0:])
	var aa Area
	nn := aa.Unmarshal(nss, buffer[0:n])
	if !reflect.DeepEqual(a, aa) {
		t.Errorf("Expected %+v, found %+v", a, aa)
	}
	if n != nn {
		t.Errorf("Expected marshalled and unmarshaled lengths to be equal (%d vs %d)", n, nn)
	}
	g := a.Polygons.(*AreaGeometryLatLngs)
	if l := MarshalledArea(buffer[0:]).Len(); l != len(g.Polygons) {
		t.Errorf("Expected %d polygons, found %d", len(g.Polygons), l)
	}
	p, _ := a.Polygons.Polygon(0)
	pp, _ := aa.Polygons.Polygon(0)
	if p.Area() != pp.Area() {
		t.Errorf("Expected marshalled and unmarshaled areas to be equal (%f vs %f)", p.Area(), pp.Area())
	}
	ppp, _ := MarshalledArea(buffer[0:]).UnmarshalPolygons(NamespaceInvalid).Polygon(0)
	if p.Area() != ppp.Area() {
		t.Errorf("Expected marshalled area and area via MarshalledArea to be equal (%f vs %f)", p.Area(), ppp.Area())
	}
}

func TestAreaEncodingMixed(t *testing.T) {
	nss, nt := newOSMNamespaces()
	ids := []uint64{4256245, 804447787}
	a := Area{
		Tags: []Tag{
			{Key: 1, Value: 42},
			{Key: 36, Value: 2043},
		},
		Polygons: &AreaGeometryMixed{
			Polygons: []PolygonGeometryMixed{
				{
					LatLngs: PolygonGeometryLatLngs{
						Loops: []int{4},
						Points: LatLngs{
							LatLngFromDegrees(51.5235396, -0.1251689), // Outer
							LatLngFromDegrees(51.5231971, -0.1249710),
							LatLngFromDegrees(51.5233405, -0.1243188),
							LatLngFromDegrees(51.5236851, -0.1245202),

							LatLngFromDegrees(51.5235250, -0.1246718), // Inner
							LatLngFromDegrees(51.5233982, -0.1245960),
							LatLngFromDegrees(51.5233456, -0.1248273),
							LatLngFromDegrees(51.5234767, -0.1249024),
						},
					},
				},
				{
					References: PolygonGeometryReferences{
						Paths: References{
							{Namespace: nt.Encode(b6.NamespaceOSMWay), Value: ids[0]},
							{Namespace: nt.Encode(b6.NamespaceOSMWay), Value: ids[1]},
						},
					},
				},
			},
		},
	}

	var buffer [128]byte
	n := a.Marshal(nss, buffer[0:])
	var aa Area
	nn := aa.Unmarshal(nss, buffer[0:n])
	if !reflect.DeepEqual(a, aa) {
		t.Errorf("Expected %+v, found %+v", a, aa)
	}
	if n != nn {
		t.Errorf("Expected marshalled and unmarshaled lengths to be equal (%d vs %d)", n, nn)
	}
	g := a.Polygons.(*AreaGeometryMixed)
	if l := MarshalledArea(buffer[0:]).Len(); l != len(g.Polygons) {
		t.Errorf("Expected %d polygons, found %d", len(g.Polygons), l)
	}
	p, _ := a.Polygons.Polygon(0)
	pp, _ := aa.Polygons.Polygon(0)
	if p.Area() != pp.Area() {
		t.Errorf("Expected marshalled and unmarshaled areas to be equal (%f vs %f)", p.Area(), pp.Area())
	}
	if _, ok := aa.Polygons.PathIDs(0); ok {
		t.Errorf("Expected no PathIDs for polygon 0")
	}
	paths, _ := aa.Polygons.PathIDs(1)
	if len(paths) != 2 || paths[0].Value != ids[0] || paths[1].Value != ids[1] {
		t.Errorf("Expected ids %+v, found %+v", ids, paths)
	}
	if _, ok := aa.Polygons.Polygon(1); ok {
		t.Errorf("Expected no Polygon for polygon 1")
	}
}

func TestRelationEncoding(t *testing.T) {
	nss, nt := newOSMNamespaces()
	r := Relation{
		Tags: []Tag{
			{Key: 1, Value: 42},
			{Key: 36, Value: 2043},
		},
		Members: Members{
			{
				Type: b6.FeatureTypePath,
				ID:   Reference{Namespace: nt.Encode(b6.NamespaceOSMWay), Value: 544908185},
				Role: 2,
			},
			{
				Type: b6.FeatureTypeArea,
				ID:   Reference{Namespace: nt.Encode(b6.NamespaceOSMWay), Value: 544908184},
				Role: 6,
			},
			{
				Type: b6.FeatureTypePoint,
				ID:   Reference{Namespace: nt.Encode(b6.NamespaceOSMWay), Value: 544908186},
				Role: 7,
			},
		},
		Relations: References{
			{Namespace: nt.Encode(b6.NamespaceOSMRelation), Value: 7216547},
		},
	}

	var buffer [128]byte
	n := r.Marshal(b6.FeatureTypePath, nss, buffer[0:])
	var rr Relation
	nn := rr.Unmarshal(b6.FeatureTypePath, nss, buffer[0:n])
	if !reflect.DeepEqual(r, rr) {
		t.Errorf("Expected %+v, found %+v", r, rr)
	}
	if n != nn {
		t.Errorf("Expected marshalled and unmarshaled lengths to be equal (%d vs %d)", n, nn)
	}
	if l := MarshalledRelation(buffer[0:]).Len(); l != len(r.Members) {
		t.Errorf("Expected %d members, found %d", len(r.Members), l)
	}
}

func TestNamespaceTableEncoding(t *testing.T) {
	var nt NamespaceTable
	nt.FillFromNamespaces(b6.StandardNamespaces)

	var output encoding.Buffer
	start := encoding.Offset(42)
	offset, err := nt.Write(&output, start)
	if err != nil {
		t.Errorf("Expected no error, found %s", err)
	}

	var ntt NamespaceTable
	buffer := output.Bytes()
	l := ntt.Unmarshal(buffer[start:])
	for _, ns := range b6.StandardNamespaces {
		if nt.Encode(ns) != ntt.Encode(ns) {
			t.Errorf("Expected to encode %s to %d, found %d", ns, nt.Encode(ns), ntt.Encode(ns))
		}
	}
	for i := range b6.StandardNamespaces {
		if nt.Decode(Namespace(i)) != ntt.Decode(Namespace(i)) {
			t.Errorf("Expected to decode %d to %s, found %s", i, nt.Decode(Namespace(i)), ntt.Decode(Namespace(i)))
		}
	}
	if start+encoding.Offset(l) != offset {
		t.Errorf("Expected marshalled and unmarshaled lengths to be equal (%d vs %d)", start+encoding.Offset(l), offset)
	}
	if int(offset-start) != nt.Length() {
		t.Errorf("Expected marshalled and unmarshaled Length() to be equal (%d vs %d)", int(start-offset), nt.Length())
	}

}

func TestNamespaceIndiciesEncoding(t *testing.T) {
	nss, _ := newOSMNamespaces()
	i := NamespaceIndicies{
		NamespaceIndex{TypeAndNamespace: CombineTypeAndNamespace(b6.FeatureTypePoint, nss.ForType(b6.FeatureTypePoint)), Index: 0},
		NamespaceIndex{TypeAndNamespace: CombineTypeAndNamespace(b6.FeatureTypePath, nss.ForType(b6.FeatureTypePath)), Index: 1029},
		NamespaceIndex{TypeAndNamespace: CombineTypeAndNamespace(b6.FeatureTypeRelation, nss.ForType(b6.FeatureTypeRelation)), Index: 4087},
	}

	var buffer [128]byte
	n := i.Marshal(buffer[0:])
	var ii NamespaceIndicies
	nn := ii.Unmarshal(buffer[0:n])
	if !reflect.DeepEqual(i, ii) {
		t.Errorf("Expected %+v, found %+v", i, ii)
	}
	if n != nn {
		t.Errorf("Expected marshalled and unmarshaled lengths to be equal (%d vs %d)", n, nn)
	}
}

func TestPostingListHeaderEncoding(t *testing.T) {
	nss, _ := newOSMNamespaces()
	h := PostingListHeader{
		Token:    "building=yes",
		Features: 36,
		Namespaces: NamespaceIndicies{
			NamespaceIndex{TypeAndNamespace: CombineTypeAndNamespace(b6.FeatureTypePoint, nss.ForType(b6.FeatureTypePoint)), Index: 0},
			NamespaceIndex{TypeAndNamespace: CombineTypeAndNamespace(b6.FeatureTypePath, nss.ForType(b6.FeatureTypePoint)), Index: 42},
		},
	}
	buffer := make([]byte, 1024)
	n := h.Marshal(buffer)

	var hh PostingListHeader
	nn := hh.Unmarshal(buffer)
	if !reflect.DeepEqual(h, hh) {
		t.Errorf("Expected %+v, found %+v", h, hh)
	}
	if n != nn {
		t.Errorf("Expected marshalled and unmarshaled lengths to be equal (%d vs %d)", n, nn)
	}
}

func TestEncodeLargePostingList(t *testing.T) {
	nss, nt := newOSMNamespaces()
	layout := []struct {
		t     b6.FeatureType
		ns    Namespace
		count int
	}{
		{b6.FeatureTypePoint, nss.ForType(b6.FeatureTypePoint), 1000},
		{b6.FeatureTypePath, nss.ForType(b6.FeatureTypePath), 2000},
		{b6.FeatureTypeArea, nss.ForType(b6.FeatureTypeArea), 1000},
	}

	r := rand.New(rand.NewSource(42))
	var ids FeatureIDs
	for _, l := range layout {
		delta := (1 << 32) / l.count
		value := uint64(r.Intn(delta)) + 1
		for i := 0; i < l.count; i++ {
			ids.Append(FeatureID{Type: l.t, Namespace: l.ns, Value: value})
			value += uint64(r.Intn(delta)) + 1
		}
	}
	sort.Sort(&ids)

	buffer := make([]byte, 2048)
	var pl PostingListHeader
	buffer = FillPostingList(ids.Begin(), "highway=primary", buffer, &pl)
	if len(buffer) > 4*ids.Len() {
		t.Errorf("Expected to average better compression than 3 bytes per ID (%d vs %d)", len(buffer), 4*ids.Len())
	}

	tests := []struct {
		name string
		f    func(header *PostingListHeader, ids []byte, expected *FeatureIDs, nt *NamespaceTable, t *testing.T)
	}{
		{"PostingListIteratorNext", ValidatePostingListIteratorNext},
		{"PostingListIteratorAdvance", ValidatePostingListIteratorAdvance},
		{"PostingListIteratorAdvanceBeyondEnd", ValidatePostingListIteratorAdvanceBeyondEnd},
		{"PostingListIteratorAdvanceBeforeCurrent", ValidatePostingListIteratorAdvanceBeforeCurrent},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) { test.f(&pl, buffer, &ids, nt, t) })
	}
}

func ValidatePostingListIteratorNext(header *PostingListHeader, ids []byte, expected *FeatureIDs, nt *NamespaceTable, t *testing.T) {
	found := make(b6.FeatureIDs, 0, expected.Len())
	i := NewIterator(header, ids, nt)
	seen := make(map[b6.FeatureID]struct{})
	for i.Next() {
		if _, ok := seen[i.FeatureID()]; ok {
			t.Errorf("Duplicate ID %s", i.FeatureID())
			return
		} else {
			seen[i.FeatureID()] = struct{}{}
		}
		found = append(found, i.FeatureID())
	}

	if len(found) != header.Features {
		t.Errorf("Expected number of features in list to match count in header")
	}

	if len(found) != expected.Len() {
		t.Errorf("Wrong numer of IDs from iterator: found %d, expected %d)", len(found), expected.Len())
	}

	for i := 0; i < expected.Len() && i < len(found); i++ {
		expected := nt.DecodeID(expected.At(i))
		if found[i] != expected {
			t.Errorf("Difference at index %d: found %s, expected %s", i, found[i], expected)
			return
		}
	}
}

func ValidatePostingListIteratorAdvance(header *PostingListHeader, ids []byte, expected *FeatureIDs, nt *NamespaceTable, t *testing.T) {
	cases := []struct {
		target   FeatureID
		delta    int
		expected FeatureID
	}{
		{expected.At(500), 0, expected.At(500)},    // Middle of current namespace
		{expected.At(1000), 0, expected.At(1000)},  // Start of second namespace
		{expected.At(1500), 0, expected.At(1500)},  // Middle of second namespace
		{expected.At(3500), 0, expected.At(3500)},  // Middle of last namespace
		{expected.At(1000), -1, expected.At(1000)}, // Just before start of second namespace
		{expected.At(999), 0, expected.At(999)},    // In the last block of first namespace
		{expected.At(3999), 0, expected.At(3999)},  // In the last block of last namespace
	}

	for ci, c := range cases {
		i := NewIterator(header, ids, nt)
		target := c.target.AddValue(c.delta)
		if !i.Advance(nt.DecodeID(target)) {
			t.Errorf("Expected advance to return true for case %d", ci)
		}
		expected := nt.DecodeID(c.expected)
		if i.FeatureID() != expected {
			t.Errorf("Expected %s, found %s for case %d", expected, i.Value(), ci)
		}
	}
}

func ValidatePostingListIteratorAdvanceBeyondEnd(header *PostingListHeader, ids []byte, expected *FeatureIDs, nt *NamespaceTable, t *testing.T) {
	i := NewIterator(header, ids, nt)
	for j := 0; j < 42; j++ {
		if !i.Next() {
			t.Errorf("Expected iterator to advance")
			return
		}
	}
	v := i.FeatureID()
	if i.Advance(nt.DecodeID(expected.At(expected.Len() - 1).AddValue(1))) {
		t.Errorf("Expected Advance() to return false when advancing beyond end")
	}
	if i.FeatureID() != v {
		t.Errorf("Expected iterator to not move following a failed Advance()")
	}
}

func ValidatePostingListIteratorAdvanceBeforeCurrent(header *PostingListHeader, ids []byte, expected *FeatureIDs, nt *NamespaceTable, t *testing.T) {
	i := NewIterator(header, ids, nt)
	if !i.Advance(nt.DecodeID(expected.At(2500))) {
		t.Errorf("Expected Advance() to return true")
		return
	}

	if !i.Advance(nt.DecodeID(expected.At(500))) {
		t.Errorf("Expected Advance() to a previoud id to return true")
	}

	if i.FeatureID() != nt.DecodeID(expected.At(2500)) {
		t.Errorf("Expected iterator to not move following Advance() to previous id")
	}
}

func TestPostingListIteratorAdvanceToCurrentAtEndOfBlock(t *testing.T) {
	// Recreate a bug seen in which matching features were dropped from
	// the results of an intersection query, ultimately because Advance()
	// on an iterator at the end of a block incorrectly moved forward when
	// given the ID at the iterator's current position - it should stay
	// where it is. This test recreates that situation.
	nss, nt := newOSMNamespaces()
	r := rand.New(rand.NewSource(42))
	var ids FeatureIDs
	max := 100000
	for i := 0; i < 1000; i++ { // Enough IDs to fill a number of blocks
		ids.Append(FeatureID{Type: b6.FeatureTypePoint, Namespace: nss.ForType(b6.FeatureTypePoint), Value: uint64(r.Intn(max))})
	}
	sort.Sort(&ids)

	var header PostingListHeader
	buffer := make([]byte, 0)
	encoder := NewPostingListEncoder(buffer, &header)
	blockStart := -1
	for i := 0; i < ids.Len(); i++ {
		encoder.Append(ids.At(i))
		if len(encoder.Buffer)%PostingListBlockSize == 2 {
			blockStart = i
		}

	}

	// Move to the end of a block
	i := NewIterator(&header, encoder.Buffer, nt)
	for {
		if !i.Next() {
			t.Errorf("Expected next to return true")
			return
		}
		if i.FeatureID() == nt.DecodeID(ids.At(blockStart-1)) {
			break
		}
	}

	if !i.Advance(nt.DecodeID(ids.At(blockStart - 1))) {
		t.Errorf("Expected advance to return true")
	}
	if i.FeatureID() != nt.DecodeID(ids.At(blockStart-1)) {
		t.Errorf("Expected iterator position to be unchanged")
	}
}

func TestPostingListIteratorAdvanceWithinLastBlockExactMultiple(t *testing.T) {
	// Recreate a bug where Advance would drop elements when the iterator
	// was positioned within the last block, and the length of the encoded
	// IDs was an exact multiple of PostingListBlockSize.
	nss, nt := newOSMNamespaces()
	var ids FeatureIDs
	var header PostingListHeader
	buffer := make([]byte, 0)
	encoder := NewPostingListEncoder(buffer, &header)
	blocks := make([]int, 0)
	v := 0
	for {
		id := FeatureID{Type: b6.FeatureTypePoint, Namespace: nss.ForType(b6.FeatureTypePoint), Value: uint64(v)}
		ids.Append(id)
		encoder.Append(id)
		if len(encoder.Buffer)%PostingListBlockSize == 0 {
			blocks = append(blocks, v)
			if len(blocks) > 2 {
				break
			}
		}
		v++
	}

	i := NewIterator(&header, encoder.Buffer, nt)
	target := nt.DecodeID(ids.At(blocks[len(blocks)-2]))
	for {
		if !i.Next() {
			t.Errorf("Expected next to return true")
			return
		}
		if i.FeatureID() == target {
			break
		}
	}

	target = nt.DecodeID(ids.At(blocks[len(blocks)-1]))
	if !i.Advance(target) {
		t.Errorf("Expected advance to return true")
	}
	if i.FeatureID() != target {
		t.Errorf("Expected %s, found %s", target, i.FeatureID())
	}
}

func TestPostingListIteratorAdvanceWithinLastBlock(t *testing.T) {
	// Recreate a bug where Advance would drop elements when the iterator
	// was positioned within the last block, and the length of the encoded
	// IDs was not an exact multiple of PostingListBlockSize.
	nss, nt := newOSMNamespaces()
	var ids FeatureIDs
	var header PostingListHeader
	buffer := make([]byte, 0)
	encoder := NewPostingListEncoder(buffer, &header)
	blocks := make([]int, 0)
	v := 0
	for {
		id := FeatureID{Type: b6.FeatureTypePoint, Namespace: nss.ForType(b6.FeatureTypePoint), Value: uint64(v)}
		ids.Append(id)
		encoder.Append(id)
		if len(encoder.Buffer)%PostingListBlockSize == 0 {
			blocks = append(blocks, v)
		}
		if len(blocks) > 2 && len(encoder.Buffer)%PostingListBlockSize == 3 { // 3 is arbitrary
			break
		}
		v++
	}

	i := NewIterator(&header, encoder.Buffer, nt)
	target := nt.DecodeID(ids.At(blocks[len(blocks)-2]))
	for {
		if !i.Next() {
			t.Errorf("Expected next to return true")
			return
		}
		if i.FeatureID() == target {
			break
		}
	}

	target = nt.DecodeID(ids.At(ids.Len() - 2))
	if !i.Advance(target) {
		t.Errorf("Expected advance to return true")
	}
	if i.FeatureID() != target {
		t.Errorf("Expected %s, found %s", target, i.FeatureID())
	}
}

func TestEncodePostingListWithShortBlock(t *testing.T) {
	// A posting list with less than a full block's worth of IDs
	nss, nt := newOSMNamespaces()

	ids := []FeatureID{
		{Type: b6.FeatureTypePoint, Namespace: nss.ForType(b6.FeatureTypePoint), Value: 2309943848},
		{Type: b6.FeatureTypePoint, Namespace: nss.ForType(b6.FeatureTypePoint), Value: 7787634265},
		{Type: b6.FeatureTypePoint, Namespace: nss.ForType(b6.FeatureTypePoint), Value: 7788210781},
	}

	for l := 0; l < len(ids); l++ {
		var subset FeatureIDs
		for i := 0; i < l; i++ {
			subset.Append(ids[i])
		}
		buffer := make([]byte, PostingListBlockSize-1) // The buffer size is less than a full block
		var pl PostingListHeader
		buffer = FillPostingList(subset.Begin(), "highway=primary", buffer, &pl)

		found := make(b6.FeatureIDs, 0, len(ids))
		i := NewIterator(&pl, buffer, nt)
		for i.Next() {
			found = append(found, i.FeatureID())
		}

		if len(found) != l {
			t.Errorf("Wrong numer of IDs from iterator: found %d, expected %d (length %d)", len(found), l, l)
		}

		for i := 0; i < len(ids) && i < len(found); i++ {
			expected := b6.FeatureID{Type: ids[i].Type, Namespace: nt.Decode(ids[i].Namespace), Value: ids[i].Value}
			if found[i] != expected {
				t.Errorf("Difference at index %d: found %v, expected %v (length %d)", i, found[i], expected, l)
				return
			}
		}

		if i.Advance(ingest.FromOSMWayID(140633010).FeatureID()) {
			t.Errorf("Didn't expect to be able to advance iterator")
		}
	}
}
