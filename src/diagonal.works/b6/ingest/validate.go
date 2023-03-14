package ingest

import (
	"fmt"
	"math"

	"diagonal.works/b6"
	"github.com/golang/geo/s2"
)

func ValidatePoint(p *PointFeature) error {
	if !p.PointID.IsValid() {
		return fmt.Errorf("%s: invalid ID", p.PointID)
	}
	return nil
}

type ClockwisePaths int

const (
	ClockwisePathsAreInvalid   ClockwisePaths = 0
	ClockwisePathsAreCorrected                = 1
)

func ValidatePath(p *PathFeature, clockwise ClockwisePaths, features b6.LocationsByID) error {
	if p == nil {
		return fmt.Errorf("ValidatePath: path is nil")
	}
	if !p.PathID.IsValid() {
		return fmt.Errorf("%s: invalid ID", p.PathID)
	}
	if p.Len() < 2 {
		return fmt.Errorf("%s: %d points, expected 2 or more", p.FeatureID(), p.Len())
	}
	points, err := p.AllPoints(features)
	if err != nil {
		return fmt.Errorf("%s: %s", p.PathID, err)
	}
	if p.IsClosed() {
		loop := s2.LoopFromPoints(points[0 : len(points)-1])
		if err := loop.Validate(); err != nil {
			return fmt.Errorf("%s: invalid loop: %s", p.FeatureID(), err)
		}
		if loop.Area() > 2.0*math.Pi {
			switch clockwise {
			case ClockwisePathsAreInvalid:
				return fmt.Errorf("%s: ordered clockwise", p.FeatureID())
			case ClockwisePathsAreCorrected:
				p.Invert()
			}
		}
	}
	return nil
}

func ValidatePathForArea(p b6.PathFeature) error {
	if p.Len() < 3 {
		return fmt.Errorf("%s: %d points, expected 3 or more", p.FeatureID(), p.Len())
	}
	if p.Point(0) != p.Point(p.Len()-1) {
		return fmt.Errorf("%s: not closed", p.FeatureID())
	}
	// ValidatePath will have already ensured that closed paths are clockwise
	return nil
}

func ValidateArea(a *AreaFeature, features b6.FeaturesByID) error {
	if !a.AreaID.IsValid() {
		return fmt.Errorf("%s: invalid ID", a.AreaID)
	}
	for i := 0; i < a.Len(); i++ {
		if ids, ok := a.PathIDs(i); ok {
			for _, id := range ids {
				if path := b6.FindPathByID(id, features); path != nil {
					if err := ValidatePathForArea(path); err != nil {
						return err
					}
				} else {
					return fmt.Errorf("%s: non-existant path %s", a.AreaID, id)
				}
			}
		}
	}
	return nil
}

func ValidateRelation(r *RelationFeature) error {
	if !r.RelationID.IsValid() {
		return fmt.Errorf("%s: invalid ID", r.RelationID)
	}
	return nil
}
