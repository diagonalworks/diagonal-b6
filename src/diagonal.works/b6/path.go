package b6

import (
	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
)

func SamplePoints(g Geometry, spacing s1.Angle, points []s2.Point) []s2.Point {
	if g.GeometryLen() < 2 {
		return points
	}
	const epsilon s1.Angle = 1.6e-09 // Roughly 1cm
	i := 0
	p := g.PointAt(0)
	remaining := spacing
	for {
		if i+1 == g.GeometryLen() {
			points = append(points, g.PointAt(i))
			break
		}
		next := g.PointAt(i + 1)
		d := p.Distance(next)
		if d < epsilon {
			p = next
			i++
		} else if d < remaining {
			remaining -= d
			p = next
			i++
		} else {
			between := s2.InterpolateAtDistance(remaining, p, next)
			points = append(points, between)
			p = between
			remaining = spacing
		}
	}
	return points
}
