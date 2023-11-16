package geometry

import (
	"github.com/golang/geo/s2"
)

type MultiPolygon []*s2.Polygon

func (m MultiPolygon) CapBound() s2.Cap {
	cap := s2.EmptyCap()
	for _, p := range m {
		cap = cap.AddCap(p.CapBound())
	}
	return cap
}

func (m MultiPolygon) RectBound() s2.Rect {
	rect := s2.EmptyRect()
	for _, p := range m {
		bound := p.RectBound()
		rect = s2.Rect{Lat: rect.Lat.Union(bound.Lat), Lng: rect.Lng.Union(bound.Lng)}
	}
	return rect
}

func (m MultiPolygon) ContainsCell(c s2.Cell) bool {
	for _, p := range m {
		if p.ContainsCell(c) {
			return true
		}
	}
	return false
}

func (m MultiPolygon) IntersectsCell(c s2.Cell) bool {
	for _, p := range m {
		if p.IntersectsCell(c) {
			return true
		}
	}
	return false
}

func (m MultiPolygon) ContainsPoint(point s2.Point) bool {
	for _, p := range m {
		if p.ContainsPoint(point) {
			return true
		}
	}
	return false
}

func (m MultiPolygon) CellUnionBound() []s2.CellID {
	union := make(s2.CellUnion, 0, 4)
	for _, p := range m {
		union = s2.CellUnionFromUnion(union, p.CellUnionBound())
	}
	return union
}

type nestedLoop struct {
	Loop     *s2.Loop
	Children []nestedLoop
}

func (l *nestedLoop) Insert(loop *s2.Loop) bool {
	if l.Loop.ContainsNested(loop) {
		for i := 0; i < len(l.Children); i++ {
			if l.Children[i].Insert(loop) {
				return true
			}
		}
		children := make([]nestedLoop, 1, len(l.Children)+1)
		children[0] = nestedLoop{Loop: loop, Children: make([]nestedLoop, 0)}
		for _, child := range l.Children {
			if children[0].Loop.ContainsNested(child.Loop) {
				children[0].Children = append(children[0].Children, child)
			} else {
				children = append(children, child)
			}
		}
		l.Children = children
		return true
	}
	return false
}

func (l *nestedLoop) Log() {
	l.log("  ")
}

func (l *nestedLoop) log(prefix string) {
	for _, child := range l.Children {
		child.log(prefix + "  ")
	}
}

func (l *nestedLoop) Collect(loops []*s2.Loop) []*s2.Loop {
	loops = append(loops, l.Loop)
	for _, child := range l.Children {
		loops = child.Collect(loops)
	}
	return loops
}

// NewMultiPolygonFromLoops takes a list of loops, and returns a list of polygons,
// each with one outer loop, by detmining the nesting hierarchy of loops. Loops
// cannot overlap.
func NewMultiPolygonFromLoops(loops []*s2.Loop) MultiPolygon {
	root := nestedLoop{Loop: s2.FullLoop(), Children: make([]nestedLoop, 0)}
	for _, loop := range loops {
		root.Insert(loop)
	}
	m := make(MultiPolygon, len(root.Children))
	for i, child := range root.Children {
		loops := child.Collect(make([]*s2.Loop, 0, 1))
		m[i] = s2.PolygonFromLoops(loops)
	}
	return m
}

func PolylineEqual(p *s2.Polyline, pp *s2.Polyline) bool {
	if len(*p) != len(*pp) {
		return false
	}
	for i := range *p {
		if (*p)[i] != (*pp)[i] {
			return false
		}
	}
	return true
}

func PolygonEqual(p *s2.Polygon, pp *s2.Polygon) bool {
	if p.NumLoops() != pp.NumLoops() {
		return false
	}
	for i := 0; i < p.NumLoops(); i++ {
		l := p.Loop(i)
		ll := pp.Loop(i)
		if l.NumVertices() != ll.NumVertices() {
			return false
		}
		for j := 0; j < l.NumVertices(); j++ {
			if l.Vertex(j) != ll.Vertex(j) {
				return false
			}
		}
	}
	return true
}

func MultiPolygonEqual(m MultiPolygon, mm MultiPolygon) bool {
	if len(m) != len(mm) {
		return false
	}
	for i := range m {
		if !PolygonEqual(m[i], mm[i]) {
			return false
		}
	}
	return true
}
