package units

import (
	"github.com/golang/geo/s1"
)

const EarthRadiusMeters = 6371.01 * 1000.0

func AngleToMeters(angle s1.Angle) float64 {
	return EarthRadiusMeters * angle.Radians()
}

func MetersToAngle(meters float64) s1.Angle {
	return s1.Angle(meters/EarthRadiusMeters) * s1.Radian
}

func AreaToMeters2(area float64) float64 {
	return area * EarthRadiusMeters * EarthRadiusMeters
}

func Meters2ToArea(m2 float64) float64 {
	return m2 / (EarthRadiusMeters * EarthRadiusMeters)
}
