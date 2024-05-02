package ingest

import (
	"fmt"
	"math"

	"diagonal.works/b6"
	"github.com/golang/geo/s2"
)

// TODO(mari): see if this can be more feature/agnostic; also we prob want to validate our references aren't circular in general
func ValidateFeature(feature Feature, o *ValidateOptions, features b6.FeaturesByID) error {
	if !feature.FeatureID().IsValid() {
		return fmt.Errorf("%s: invalid ID", feature.FeatureID())
	}

	if feature.FeatureID().Type == b6.FeatureTypePath {
		return ValidatePath(feature.(b6.PhysicalFeature), o, features)
	}

	if feature.FeatureID().Type == b6.FeatureTypeArea {
		return ValidateArea(feature.(*AreaFeature), features)
	}

	return nil
}

type ValidateOptions struct {
	// Invert clockwise paths to anticlockwise if true, otherwise, consider them
	// invalid
	InvertClockwisePaths bool
}

func ValidatePath(p b6.PhysicalFeature, o *ValidateOptions, features b6.LocationsByID) error {
	if p == nil {
		return fmt.Errorf("ValidatePath: path is nil")
	}
	if !p.FeatureID().IsValid() {
		return fmt.Errorf("%s: invalid ID", p.FeatureID())
	}
	if p.GeometryLen() < 2 {
		return fmt.Errorf("%s: %d points, expected 2 or more", p.FeatureID(), p.GeometryLen())
	}
	points, err := pathPoints(p, features)
	if err != nil {
		return fmt.Errorf("%s: %s", p.FeatureID(), err)
	}
	if p.AllTags().ClosedPath() {
		loop := s2.LoopFromPoints(points[0 : len(points)-1])
		if err := loop.Validate(); err != nil {
			return fmt.Errorf("%s: invalid loop: %s", p.FeatureID(), err)
		}
		if loop.Area() > 2.0*math.Pi {
			if o.InvertClockwisePaths {
				invertPoints(p.(Feature))
			} else {
				return fmt.Errorf("%s: ordered clockwise", p.FeatureID())
			}
		}
	}
	return nil
}

func invertPoints(f Feature) {
	if refs := f.Get(b6.PathTag).Value; refs != nil {
		if refs, ok := refs.(b6.Values); ok {
			n := len(refs)
			for i := 0; i < n/2; i++ {
				(refs)[i], (refs)[n-i-1] = (refs)[n-i-1], (refs)[i]
			}

			f.ModifyOrAddTag(b6.Tag{b6.PathTag, refs})
		}
	}
}

// pathPoints returns a list of s2.Points for the points along the path,
// or nil if at least one point is missing, together with an error.
func pathPoints(f b6.PhysicalFeature, byID b6.LocationsByID) ([]s2.Point, error) {
	points := make([]s2.Point, f.GeometryLen())
	for i := 0; i < f.GeometryLen(); i++ {
		if point := f.PointAt(i); point.Norm() != 0 {
			points[i] = point
		} else {
			id := f.Reference(i).Source()
			if ll, err := byID.FindLocationByID(id); err == nil {
				points[i] = s2.PointFromLatLng(ll)
			} else {
				return nil, fmt.Errorf("Path %s missing point %s", f.FeatureID(), id)
			}
		}
	}
	return points, nil
}

func ValidatePathForArea(p b6.PhysicalFeature) error {
	if p.GeometryLen() < 3 {
		return fmt.Errorf("%s: %d points, expected 3 or more", p.FeatureID(), p.GeometryLen())
	}
	if p.PointAt(0) != p.PointAt(p.GeometryLen()-1) {
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
				if path := features.FindFeatureByID(id); path != nil {
					if err := ValidatePathForArea(path.(b6.PhysicalFeature)); err != nil {
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

func ValidateCollection(c *CollectionFeature) error {
	if !c.CollectionID.IsValid() {
		return fmt.Errorf("%s: invalid ID", c.CollectionID)
	}
	return nil
}

func ValidateExpression(e *ExpressionFeature) error {
	if !e.ExpressionID.IsValid() {
		return fmt.Errorf("%s: invalid ID", e.ExpressionID)
	}
	return nil
}
