package renderer

import (
	"math"
	"testing"

	"github.com/golang/geo/r2"
)

func TestDistanceBetweenPointAndLine(t *testing.T) {
	tests := []struct {
		a        r2.Point
		b        r2.Point
		p        r2.Point
		expected float64
	}{
		{r2.Point{1.0, 1.0}, r2.Point{3.0, 1.0}, r2.Point{2.0, 2.0}, 1.0},
		{r2.Point{1.0, 1.0}, r2.Point{3.0, 1.0}, r2.Point{2.0, -2.0}, 3.0},
		{r2.Point{1.0, 1.0}, r2.Point{3.0, 1.0}, r2.Point{10.0, 2.0}, 1.0},
	}

	for _, test := range tests {
		d := distance(test.a, test.b, test.p)
		if math.Abs((d-test.expected)/test.expected) > 0.0001 {
			t.Errorf("Expected distance(%v,%v,%v) == %f, found %f", test.a, test.b, test.p, test.expected, d)
		}
	}
}

func TestSimplify(t *testing.T) {
	tests := []struct {
		points   []r2.Point
		expected []r2.Point
	}{
		{
			points: []r2.Point{
				{0.0, 0.0},
				{1.0, 0.0},
				{1.0, 1.0},
				{0.0, 1.0}},

			expected: []r2.Point{
				{0.0, 0.0},
				{1.0, 0.0},
				{1.0, 1.0},
				{0.0, 1.0}},
		},
		{
			points: []r2.Point{
				{0.0, 0.0},
				{0.5, 0.0},
				{1.0, 0.0},
				{1.0, 0.5},
				{1.0, 1.0},
				{0.5, 1.0},
				{0.0, 1.0},
				{0.0, 0.5}},
			expected: []r2.Point{
				{0.0, 0.0},
				{1.0, 0.0},
				{1.0, 1.0},
				{0.0, 1.0},
				{0.0, 0.5}},
		},
		{
			points: []r2.Point{
				{10.0, 10.0},
				{10.5, 10.0},
				{11.0, 10.0},
				{11.0, 10.5},
				{11.0, 11.0},
				{10.5, 11.0},
				{10.0, 11.0},
				{10.0, 10.5}},
			expected: []r2.Point{
				{10.0, 10.0},
				{11.0, 10.0},
				{11.0, 11.0},
				{10.0, 11.0},
				{10.0, 10.5}},
		},
	}

	run := func(simplify func([]r2.Point, float64) []r2.Point, t *testing.T) {
		for i, test := range tests {
			simplified := simplify(test.points, 0.1)
			if len(simplified) != len(test.expected) {
				t.Errorf("Expected %d points for case %d, found %d", len(test.expected), i, len(simplified))
			} else {
				for j := 0; j < len(simplified); j++ {
					if simplified[j].Sub(test.expected[j]).Norm() > 0.1 {
						t.Errorf("Expected %v for case %d, point %d, found %v", test.expected[j], i, j, simplified[j])
					}
				}
			}
		}
	}

	algorithms := []struct {
		name string
		f    func([]r2.Point, float64) []r2.Point
	}{
		{"Naive", naiveSimplify},
		{"ReferenceDouglasPecker", referenceDouglasPeuckerSimplify},
		{"DouglasPecker", douglasPeuckerSimplify},
	}
	for _, algorithm := range algorithms {
		t.Run(algorithm.name, func(t *testing.T) { run(algorithm.f, t) })
	}
}
