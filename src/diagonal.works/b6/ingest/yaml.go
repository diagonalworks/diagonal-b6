package ingest

import (
	"fmt"
	"io"
	"strings"

	"diagonal.works/b6"
	"github.com/golang/geo/s2"
	yaml "gopkg.in/yaml.v2"
)

type FeatureIDYAML struct {
	b6.FeatureID
}

func (f FeatureIDYAML) MarshalYAML() (interface{}, error) {
	return "/" + f.FeatureID.String(), nil
}

func (f *FeatureIDYAML) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}
	return f.UnmarshalYAMLString(s)
}

func (f *FeatureIDYAML) UnmarshalYAMLString(s string) error {
	if len(s) > 0 {
		f.FeatureID = b6.FeatureIDFromString(s[1:])
	} else {
		f.FeatureID = b6.FeatureIDInvalid
	}
	return nil
}

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
	return fmt.Errorf("invalid lat,lng: %s", s)
}

type RelationMemberYAML struct {
	b6.RelationMember
}

func (r RelationMemberYAML) MarshalYAML() (interface{}, error) {
	y := map[string]interface{}{
		"id": FeatureIDYAML{FeatureID: r.RelationMember.ID},
	}
	if r.RelationMember.Role != "" {
		y["role"] = r.RelationMember.Role
	}
	return y, nil
}

func (r *RelationMemberYAML) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var v struct {
		ID   FeatureIDYAML
		Role string
	}
	if err := unmarshal(&v); err != nil {
		return nil
	}
	r.RelationMember.ID = v.ID.FeatureID
	r.RelationMember.Role = v.Role
	return nil
}

type CollectionYAML struct {
	Keys   []InterfaceYAML
	Values []InterfaceYAML
}

type InterfaceYAML struct {
	v interface{}
}

func (i InterfaceYAML) MarshalYAML() (interface{}, error) {
	switch v := i.v.(type) {
	case b6.FeatureID:
		return map[string]interface{}{"v": FeatureIDYAML{FeatureID: v}}, nil
	default:
		return map[string]interface{}{"v": v}, nil
	}
}

func (i *InterfaceYAML) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var m struct {
		V interface{}
	}
	if err := unmarshal(&m); err != nil {
		return err
	}

	if id := b6.FeatureIDFromString(m.V.(string)[1:]); id != b6.FeatureIDInvalid {
		i.v = id
	} else {
		i.v = m.V
	}

	return nil
}

type exportedYAML struct {
	ID     FeatureIDYAML
	Add    []b6.Tag `yaml:",omitempty"`
	Remove []string `yaml:",omitempty"`

	Point      *LatLngYAML          `yaml:",omitempty"`
	Path       []interface{}        `yaml:",omitempty"`
	Area       []interface{}        `yaml:",omitempty"`
	Relation   []RelationMemberYAML `yaml:",omitempty"`
	Collection *CollectionYAML      `yaml:",omitempty"`
	Tags       []b6.Tag             `yaml:",omitempty"`
}

type modifiedFeatureYAML struct {
	Feature Feature
}

func (f modifiedFeatureYAML) MarshalYAML() (interface{}, error) {
	if y, ok := f.Feature.(yaml.Marshaler); ok {
		return y.MarshalYAML()
	}
	return f.Feature, nil
}

func ExportChangesAsYAML(m MutableWorld, w io.Writer) error {
	encoder := yaml.NewEncoder(w)
	y := exportedYAML{ID: FeatureIDYAML{FeatureID: b6.FeatureIDInvalid}}
	tags := func(t ModifiedTag, goroutine int) error {
		if y.ID.FeatureID != t.ID {
			if y.ID.FeatureID != b6.FeatureIDInvalid {
				if err := encoder.Encode(y); err != nil {
					return err
				}
			}
			y.ID.FeatureID = t.ID
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
	if y.ID.FeatureID != b6.FeatureIDInvalid {
		if err := encoder.Encode(y); err != nil {
			return err
		}
	}

	features := func(f b6.Feature, goroutine int) error {
		return encoder.Encode(modifiedFeatureYAML{Feature: NewFeatureFromWorld(f)})
	}
	m.EachModifiedFeature(features, &b6.EachFeatureOptions{Goroutines: 1})
	return nil
}

func IngestChangesFromYAML(r io.Reader) Change {
	return &ingestedYAML{r: r}
}

type ingestedYAML struct {
	r io.Reader
}

func (i ingestedYAML) Apply(m MutableWorld) (AppliedChange, error) {
	applied := make(map[b6.FeatureID]b6.FeatureID)
	decoder := yaml.NewDecoder(i.r)
	for {
		var y exportedYAML
		if err := decoder.Decode(&y); err != nil {
			if err == io.EOF {
				break
			}
			return applied, err
		}
		var err error
		if y.Point != nil {
			var p *PointFeature
			if p, err = newPointFromYAML(&y); err == nil {
				err = m.AddPoint(p)
			}
		} else if y.Path != nil {
			var p *PathFeature
			if p, err = newPathFromYAML(&y); err == nil {
				err = m.AddPath(p)
			}
		} else if y.Area != nil {
			var a *AreaFeature
			if a, err = newAreaFromYAML(&y); err == nil {
				err = m.AddArea(a)
			}
		} else if y.Relation != nil {
			var r *RelationFeature
			if r, err = newRelationFromYAML(&y); err == nil {
				err = m.AddRelation(r)
			}
		} else if y.Collection != nil {
			var c *CollectionFeature
			if c, err = newCollectionFeatureFromYAML(&y); err == nil {
				err = m.AddCollection(c)
			}
		}
		if err != nil {
			return applied, err
		}
		for _, tag := range y.Add {
			m.AddTag(y.ID.FeatureID, tag)
		}
		for _, key := range y.Remove {
			m.RemoveTag(y.ID.FeatureID, key)
		}
		applied[y.ID.FeatureID] = y.ID.FeatureID
	}
	return applied, nil
}

// TODO: find a neat way of moving these functions alongside MarshalYAML
// on the feature implementations themselves.
func newPointFromYAML(y *exportedYAML) (*PointFeature, error) {
	p := NewPointFeature(y.ID.ToPointID(), y.Point.LatLng)
	p.Tags = y.Tags
	return p, nil
}

func newPathFromYAML(y *exportedYAML) (*PathFeature, error) {
	p := NewPathFeature(len(y.Path))
	p.PathID = y.ID.ToPathID()
	for i := range y.Path {
		if s, ok := y.Path[i].(string); ok {
			if strings.HasPrefix(s, "/") {
				id := FeatureIDYAML{}
				id.UnmarshalYAMLString(s)
				p.SetPointID(i, id.ToPointID())
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
	a := NewAreaFeature(len(y.Area))
	a.AreaID = y.ID.ToAreaID()
	for i := range y.Area {
		if loopsYAML, ok := y.Area[i].([]interface{}); ok {
			if len(loopsYAML) > 0 {
				if _, ok := loopsYAML[0].(string); ok {
					pathIDs := make([]b6.PathID, len(loopsYAML))
					for j := range loopsYAML {
						if s, ok := loopsYAML[j].(string); ok {
							id := FeatureIDYAML{}
							id.UnmarshalYAMLString(s)
							pathIDs[j] = id.FeatureID.ToPathID()
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
	r := NewRelationFeature(len(y.Relation))
	r.RelationID = y.ID.ToRelationID()
	r.Members = make([]b6.RelationMember, len(y.Relation))
	for i := range y.Relation {
		r.Members[i] = y.Relation[i].RelationMember
	}
	r.Tags = y.Tags
	return r, nil
}

func newCollectionFeatureFromYAML(y *exportedYAML) (*CollectionFeature, error) {
	var keys, values []interface{}
	for _, key := range y.Collection.Keys {
		keys = append(keys, key.v)
	}

	for _, value := range y.Collection.Values {
		values = append(values, value.v)
	}

	return &CollectionFeature{
		CollectionID: y.ID.ToCollectionID(),
		Keys:         keys,
		Values:       values,
		Tags:         y.Tags,
	}, nil
}
