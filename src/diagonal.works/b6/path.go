package b6

import (
	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
)

func SamplePoints(path Path, spacing s1.Angle, points []s2.Point) []s2.Point {
	if path.Len() < 2 {
		return points
	}
	const epsilon s1.Angle = 1.6e-09 // Roughly 1cm
	i := 0
	p := path.Point(0)
	remaining := spacing
	for {
		if i+1 == path.Len() {
			points = append(points, path.Point(i))
			break
		}
		next := path.Point(i + 1)
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
