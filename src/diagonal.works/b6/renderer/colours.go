package renderer

import (
	"fmt"
	"math"
	"sort"
)

type Colour struct {
	Red   float64
	Green float64
	Blue  float64
}

func ColourFromHexString(s string) Colour {
	var red, green, blue int
	fmt.Sscanf(s, "#%02x%02x%02x", &red, &green, &blue)
	return Colour{Red: float64(red) / 255.0, Green: float64(green) / 255.0, Blue: float64(blue) / 255.0}
}

func (c Colour) ToHexString() string {
	return fmt.Sprintf("#%02x%02x%02x", int(math.Round(c.Red*255)), int(math.Round(c.Green*255)), int(math.Round(c.Blue*255)))
}

func (c Colour) Interpolate(other Colour, ratio float64) Colour {
	return Colour{
		Red:   c.Red + ((other.Red - c.Red) * ratio),
		Green: c.Green + ((other.Green - c.Green) * ratio),
		Blue:  c.Blue + ((other.Blue - c.Blue) * ratio),
	}
}

type GradientPoint struct {
	Value float64
	Colour
}

type Gradient []GradientPoint

func (g Gradient) Len() int           { return len(g) }
func (g Gradient) Swap(i, j int)      { g[i], g[j] = g[j], g[i] }
func (g Gradient) Less(i, j int) bool { return g[i].Value < g[j].Value }

func (g Gradient) Interpolate(value float64) Colour {
	i := sort.Search(len(g), func(j int) bool {
		return value < g[j].Value
	})
	if i > 0 && i < len(g) {
		return g[i-1].Interpolate(g[i].Colour, (value-g[i-1].Value)/(g[i].Value-g[i-1].Value))
	} else if i <= 0 {
		return g[0].Colour
	}
	return g[len(g)-1].Colour
}
