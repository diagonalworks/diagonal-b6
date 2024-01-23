package ingest

import (
	"fmt"
	"io"
	"strings"

	"diagonal.works/b6"
	"github.com/golang/geo/s2"
	yaml "gopkg.in/yaml.v2"
)

type LatLngYAML struct {
	s2.LatLng
}

func (f LatLngYAML) MarshalYAML() (interface{}, error) {
	return fmt.Sprintf("%f, %f", f.LatLng.Lat.Degrees(), f.LatLng.Lng.Degrees()), nil
}

func (f *LatLngYAML) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}
	return f.UnmarshalYAMLString(s)
}

func (f *LatLngYAML) UnmarshalYAMLString(s string) error {
	var err error
	f.LatLng, err = b6.LatLngFromString(s)
	return err
}

type exportedYAML struct {
	ID     b6.FeatureID
	Add    []b6.Tag `yaml:",omitempty"`
	Remove []string `yaml:",omitempty"`

	Point      *LatLngYAML              `yaml:",omitempty"`
	Path       []interface{}            `yaml:",omitempty"`
	Area       []interface{}            `yaml:",omitempty"`
	Relation   []b6.RelationMember      `yaml:",omitempty"`
	Collection *b6.CollectionExpression `yaml:",omitempty"`
	Expression *b6.Expression           `yaml:",omitempty"`
	Tags       []b6.Tag                 `yaml:",omitempty"`
}

func ExportChangesAsYAML(m MutableWorld, w io.Writer) error {
	encoder := yaml.NewEncoder(w)
	y := exportedYAML{ID: b6.FeatureIDInvalid}
	tags := func(t ModifiedTag, goroutine int) error {
		if y.ID != t.ID {
			if y.ID != b6.FeatureIDInvalid {
				if err := encoder.Encode(y); err != nil {
					return err
				}
			}
			y.ID = t.ID
			y.Add = y.Add[0:0]
			y.Remove = y.Remove[0:0]
		}
		if t.Deleted {
			y.Remove = append(y.Remove, t.Tag.Key)
		} else {
			y.Add = append(y.Add, t.Tag)
		}
		return nil
	}
	if err := m.EachModifiedTag(tags, &b6.EachFeatureOptions{Goroutines: 1}); err != nil {
		return err
	}
	if y.ID != b6.FeatureIDInvalid {
		if err := encoder.Encode(y); err != nil {
			return err
		}
	}

	features := func(f b6.Feature, goroutine int) error {
		return encoder.Encode(NewFeatureFromWorld(f))
	}
	return m.EachModifiedFeature(features, &b6.EachFeatureOptions{Goroutines: 1, FeedReferencesFirst: true})
}

func IngestChangesFromYAML(r io.Reader) Change {
	return &ingestedYAML{r: r}
}

type ingestedYAML struct {
	r io.Reader
}

func (i ingestedYAML) Apply(m MutableWorld) (b6.Collection[b6.FeatureID, b6.FeatureID], error) {
	applied := b6.ArrayCollection[b6.FeatureID, b6.FeatureID]{}
	decoder := yaml.NewDecoder(i.r)
	for {
		var y exportedYAML
		if err := decoder.Decode(&y); err != nil {
			if err == io.EOF {
				break
			}
			return applied.Collection(), err
		}
		var err error
		if y.Path != nil {
			var p *PathFeature
			if p, err = newPathFromYAML(&y); err == nil {
				err = m.AddFeature(p)
			}
		} else if y.Area != nil {
			var a *AreaFeature
			if a, err = newAreaFromYAML(&y); err == nil {
				err = m.AddFeature(a)
			}
		} else if y.Relation != nil {
			var r *RelationFeature
			if r, err = newRelationFromYAML(&y); err == nil {
				err = m.AddFeature(r)
			}
		} else if y.Collection != nil {
			var c *CollectionFeature
			if c, err = newCollectionFeatureFromYAML(&y); err == nil {
				err = m.AddFeature(c)
			}
		} else if y.Expression != nil {
			var e *ExpressionFeature
			if e, err = newExpressionFromYAML(&y); err == nil {
				err = m.AddFeature(e)
			}
		} else if y.Tags != nil {
			err = m.AddFeature(newGenericFeatureFromYAML(&y))
		}
		if err != nil {
			return applied.Collection(), err
		}
		for _, tag := range y.Add {
			m.AddTag(y.ID, tag)
		}
		for _, key := range y.Remove {
			m.RemoveTag(y.ID, key)
		}
		applied.Keys = append(applied.Keys, y.ID)
		applied.Values = append(applied.Values, y.ID)
	}
	return applied.Collection(), nil
}

// TODO: find a neat way of moving these functions alongside MarshalYAML
// on the feature implementations themselves.
func newPathFromYAML(y *exportedYAML) (*PathFeature, error) {
	if y.ID.Type != b6.FeatureTypePath {
		return nil, fmt.Errorf("expected a path for %s", y.ID)
	}

	p := NewPathFeature(len(y.Path))
	p.PathID = y.ID.ToPathID()
	for i := range y.Path {
		if s, ok := y.Path[i].(string); ok {
			if strings.HasPrefix(s, "/") {
				p.SetPointID(i, b6.FeatureIDFromString(s[1:]))
			} else {
				ll := LatLngYAML{}
				if err := ll.UnmarshalYAMLString(s); err != nil {
					return nil, err
				}
				p.SetLatLng(i, ll.LatLng)
			}
		} else {
			return nil, fmt.Errorf("expected string, found %T", y.Path[i])
		}
	}
	p.Tags = y.Tags
	return p, nil
}

func newAreaFromYAML(y *exportedYAML) (*AreaFeature, error) {
	if y.ID.Type != b6.FeatureTypeArea {
		return nil, fmt.Errorf("expected an area for %s", y.ID)
	}

	a := NewAreaFeature(len(y.Area))
	a.AreaID = y.ID.ToAreaID()
	for i := range y.Area {
		if loopsYAML, ok := y.Area[i].([]interface{}); ok {
			if len(loopsYAML) > 0 {
				if _, ok := loopsYAML[0].(string); ok {
					pathIDs := make([]b6.PathID, len(loopsYAML))
					for j := range loopsYAML {
						if s, ok := loopsYAML[j].(string); ok {
							pathIDs[j] = b6.FeatureIDFromString(s).ToPathID()
						} else {
							return nil, fmt.Errorf("bad feature ID in polygon loops")
						}
					}
					a.SetPathIDs(i, pathIDs)
				} else if _, ok := loopsYAML[0].([]interface{}); ok {
					loops := make([]*s2.Loop, len(loopsYAML))
					for j := range loopsYAML {
						if pointsYAML, ok := loopsYAML[j].([]interface{}); ok {
							points := make([]s2.Point, len(pointsYAML))
							for k := range pointsYAML {
								if s, ok := pointsYAML[k].(string); ok {
									ll := LatLngYAML{}
									if err := ll.UnmarshalYAMLString(s); err != nil {
										return nil, err
									}
									points[k] = s2.PointFromLatLng(ll.LatLng)
								} else {
									return nil, fmt.Errorf("bad point in polygon loop")
								}
							}
							loops[j] = s2.LoopFromPoints(points)
						} else {
							return nil, fmt.Errorf("bad loop in polygon")
						}
					}
					a.SetPolygon(i, s2.PolygonFromLoops(loops))
				}
			}
		} else {
			return nil, fmt.Errorf("bad polygon loops")
		}
	}
	a.Tags = y.Tags
	return a, nil
}

func newRelationFromYAML(y *exportedYAML) (*RelationFeature, error) {
	if y.ID.Type != b6.FeatureTypeRelation {
		return nil, fmt.Errorf("expected a relation for %s", y.ID)
	}
	r := NewRelationFeature(len(y.Relation))
	r.RelationID = y.ID.ToRelationID()
	r.Members = y.Relation
	r.Tags = y.Tags
	return r, nil
}

func newCollectionFeatureFromYAML(y *exportedYAML) (*CollectionFeature, error) {
	if y.ID.Type != b6.FeatureTypeCollection {
		return nil, fmt.Errorf("expected a collection for %s", y.ID)
	}

	var keys, values []interface{}
	i := y.Collection.BeginUntyped()
	for {
		ok, err := i.Next()
		if err != nil {
			return nil, err
		} else if !ok {
			break
		}
		keys = append(keys, i.Key())
		values = append(values, i.Value())
	}

	return &CollectionFeature{
		CollectionID: y.ID.ToCollectionID(),
		Keys:         keys,
		Values:       values,
		Tags:         y.Tags,
	}, nil
}

func newExpressionFromYAML(y *exportedYAML) (*ExpressionFeature, error) {
	if y.ID.Type != b6.FeatureTypeExpression {
		return nil, fmt.Errorf("expected an expression for %s", y.ID)
	}

	return &ExpressionFeature{
		ExpressionID: y.ID.ToExpressionID(),
		Tags:         y.Tags,
		Expression:   *y.Expression,
	}, nil
}

func newGenericFeatureFromYAML(y *exportedYAML) *GenericFeature {
	return &GenericFeature{
		ID:   y.ID,
		Tags: y.Tags,
	}
}
