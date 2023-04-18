package renderer

import (
	"math"

	"github.com/golang/geo/r2"
)

// naiveSimplify greedily removes points following the first point, until the
// change in area exceeds tolerance, before repreating the the following
// point, until the last point is reached. A naive approach, to be
// replaced with something else later.
func naiveSimplify(points []r2.Point, tolerance float64) []r2.Point {
	start := 0
	simplified := make([]r2.Point, 0, len(points))
	for start < len(points)-2 {
		simplified = append(simplified, points[start])
		totalArea := 0.0
		end := start + 2
		for end < len(points) {
			area := math.Abs(points[end].Sub(points[start]).Cross(points[end-1].Sub(points[start])) / 2.0)
			totalArea += area
			if totalArea > tolerance {
				start = end - 1
				break
			} else if end == len(points)-1 {
				start = end
				break
			} else {
				end++
			}
		}
	}
	for i := start; i < len(points); i++ {
		simplified = append(simplified, points[i])
	}
	return simplified
}

// distance returns the perpendicular distance between the point p, and the
// line passing through ab.
func distance(a r2.Point, b r2.Point, p r2.Point) float64 {
	n := b.Sub(a).Normalize()
	return a.Sub(p).Sub(n.Mul(a.Sub(p).Dot(n))).Norm()
}

// referenceDouglasPeuckerSimplify is a reference implementation of the
// recursive Douglas Peucker curve simplification algorithm. See
// https://en.wikipedia.org/wiki/Ramer%E2%80%93Douglas%E2%80%93Peucker_algorithm
func referenceDouglasPeuckerSimplify(points []r2.Point, epsilon float64) []r2.Point {
	max := 0.0
	maxi := 0
	for i := 1; i < len(points)-1; i++ {
		if d := distance(points[0], points[len(points)-1], points[i]); d > max {
			max = d
			maxi = i
		}
	}

	if max > epsilon {
		left := referenceDouglasPeuckerSimplify(points[0:maxi], epsilon)
		right := referenceDouglasPeuckerSimplify(points[maxi:], epsilon)
		return append(append(make([]r2.Point, 0, len(left)+len(right)), left[0:len(left)-1]...), right...)
	} else {
		return []r2.Point{points[0], points[len(points)-1]}
	}

}

type interval struct {
	begin int
	end   int
}

// douglasPeuckerSimplify is an implementation of the Douglas Peucker curve
// simplification algorithm, using an explicit stack rather than recursion.
func douglasPeuckerSimplify(points []r2.Point, epsilon float64) []r2.Point {
	simplified := make([]r2.Point, 0, len(points))
	stack := make([]interval, 1, len(points))
	stack[0] = interval{begin: 0, end: len(points)}
	for len(stack) > 0 {
		top := stack[len(stack)-1]
		stack = stack[0 : len(stack)-1]

		max := 0.0
		maxi := 0
		for i := top.begin + 1; i < top.end-1; i++ {
			if d := distance(points[top.begin], points[top.end-1], points[i]); d > max {
				max = d
				maxi = i
			}
		}

		if max > epsilon {
			stack = append(stack, interval{begin: maxi, end: top.end}, interval{begin: top.begin, end: maxi})
		} else {
			simplified = append(simplified, points[top.begin])
		}
	}
	simplified = append(simplified, points[len(points)-1])
	return simplified
}

// Simplify uses the Douglas Peucker simplifcation algorithm to return a
// new simplified curve with those points falling closer than epsilon to
// simplified line removed.
func Simplify(points []r2.Point, epsilon float64) []r2.Point {
	if len(points) < 2 {
		return []r2.Point{points[0]}
	}
	return douglasPeuckerSimplify(points, epsilon)
}
