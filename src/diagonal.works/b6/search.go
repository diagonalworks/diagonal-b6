package b6

import (
	"fmt"
	"strings"

	pb "diagonal.works/b6/proto"
	"diagonal.works/b6/search"
)

type FeatureValues interface {
	Feature(v search.Value) Feature
	ID(v search.Value) FeatureID
}

type FeatureIndex interface {
	search.Index
	FeatureValues
}

type Query interface {
	Compile(i FeatureIndex, w World) search.Iterator
	Matches(f Feature, w World) bool
	ToProto() (*pb.QueryProto, error)
	String() string
}

type Empty struct {
	search.Empty
}

func (_ Empty) Matches(f Feature, w World) bool {
	return false
}

func (_ Empty) Compile(index FeatureIndex, w World) search.Iterator {
	return search.Empty{}.Compile(index)
}

func (_ Empty) String() string {
	return "(empty)"
}

func (_ Empty) ToProto() (*pb.QueryProto, error) {
	return &pb.QueryProto{
		Query: &pb.QueryProto_Empty{},
	}, nil
}

type All struct{}

func (_ All) Matches(f Feature, w World) bool {
	return true
}

func (_ All) Compile(i FeatureIndex, w World) search.Iterator {
	return search.All{Token: search.AllToken}.Compile(i)
}

func (_ All) String() string {
	return "(all)"
}

func (_ All) ToProto() (*pb.QueryProto, error) {
	return &pb.QueryProto{
		Query: &pb.QueryProto_All{},
	}, nil
}

func TokenForTag(tag Tag) (string, bool) {
	if strings.HasPrefix(tag.Key, "#") {
		return fmt.Sprintf("%s=%s", tag.Key[1:], tag.Value), true
	} else if strings.HasPrefix(tag.Key, "@") {
		return tag.Key[1:], true
	}
	return "", false
}

type Tagged Tag

func (t Tagged) Compile(i FeatureIndex, w World) search.Iterator {
	if strings.HasPrefix(t.Key, "#") {
		return search.All{Token: fmt.Sprintf("%s=%s", t.Key[1:], t.Value)}.Compile(i)
	}
	return search.NewEmptyIterator()
}

func (t Tagged) Matches(f Feature, w World) bool {
	return f.Get(t.Key) == Tag(t)
}

func (t Tagged) String() string {
	return fmt.Sprintf("(key-value %s %s)", t.Key, t.Value)
}

func (t Tagged) ToProto() (*pb.QueryProto, error) {
	return &pb.QueryProto{
		Query: &pb.QueryProto_Tagged{
			Tagged: &pb.TagProto{
				Key:   t.Key,
				Value: t.Value,
			},
		},
	}, nil
}

type Keyed struct {
	Key string
}

func (k Keyed) Compile(i FeatureIndex, w World) search.Iterator {
	if strings.HasPrefix(k.Key, "#") {
		return search.TokenPrefix{Prefix: k.Key[1:] + "="}.Compile(i)
	} else if strings.HasPrefix(k.Key, "@") {
		return search.All{Token: k.Key[1:]}.Compile(i)
	}
	return search.NewEmptyIterator()
}

func (k Keyed) Matches(f Feature, w World) bool {
	return f.Get(k.Key).IsValid()
}

func (k Keyed) String() string {
	return fmt.Sprintf("(key %s)", k.Key)
}

func (k Keyed) ToProto() (*pb.QueryProto, error) {
	return &pb.QueryProto{
		Query: &pb.QueryProto_Keyed{
			Keyed: k.Key,
		},
	}, nil
}

type Typed struct {
	Type  FeatureType
	Query Query
}

type adaptQuery struct {
	Query Query
	World World
}

func (a adaptQuery) Compile(i search.Index) search.Iterator {
	return a.Query.Compile(i.(FeatureIndex), a.World)
}

func (a adaptQuery) String() string {
	return a.Query.String()
}

func (t Typed) Compile(i FeatureIndex, w World) search.Iterator {
	var begin, end FeatureID
	switch t.Type {
	case FeatureTypePoint:
		begin, end = FeatureIDPointBegin, FeatureIDPointEnd
	case FeatureTypePath:
		begin, end = FeatureIDPathBegin, FeatureIDPathEnd
	case FeatureTypeArea:
		begin, end = FeatureIDAreaBegin, FeatureIDAreaEnd
	case FeatureTypeRelation:
		begin, end = FeatureIDRelationBegin, FeatureIDRelationEnd
	default:
		panic("Bad FeatureType")
	}
	return search.KeyRange{Begin: begin, End: end, Query: adaptQuery{Query: t.Query, World: w}}.Compile(i)
}

func (t Typed) Matches(f Feature, w World) bool {
	return f.FeatureID().Type == t.Type
}

func (t Typed) String() string {
	return fmt.Sprintf("(feature-type %s %s)", t.Type.String(), t.Query.String())
}

func (t Typed) ToProto() (*pb.QueryProto, error) {
	return &pb.QueryProto{
		Query: &pb.QueryProto_Typed{
			Typed: &pb.TypedQueryProto{
				Type: NewProtoFromFeatureType(t.Type),
			},
		},
	}, nil
}

type typedYAML struct {
	Type  FeatureType
	Query queryYAML
}

func (t Typed) MarshalYAML() (interface{}, error) {
	return &typedYAML{Type: t.Type, Query: queryYAML{Query: t.Query}}, nil
}

func (t *Typed) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var y typedYAML
	err := unmarshal(&y)
	if err == nil {
		t.Type = y.Type
		t.Query = y.Query.Query
	}
	return err
}

type Intersection []Query

func (i Intersection) Compile(index FeatureIndex, w World) search.Iterator {
	qs := make(search.Intersection, len(i))
	for ii, q := range i {
		qs[ii] = adaptQuery{Query: q, World: w}
	}
	return qs.Compile(index)
}

func (i Intersection) Matches(f Feature, w World) bool {
	for _, q := range i {
		if !q.Matches(f, w) {
			return false
		}
	}
	return true
}

func (i Intersection) String() string {
	qs := make([]string, len(i))
	for ii, q := range i {
		qs[ii] = q.String()
	}
	return fmt.Sprintf("(intersection %s)", strings.Join(qs, ","))
}

func (i Intersection) ToProto() (*pb.QueryProto, error) {
	qs := make([]*pb.QueryProto, len(i))
	for ii, q := range i {
		var err error
		if qs[ii], err = q.ToProto(); err != nil {
			return nil, err
		}
	}
	return &pb.QueryProto{
		Query: &pb.QueryProto_Intersection{
			Intersection: &pb.QueriesProto{
				Queries: qs,
			},
		},
	}, nil
}

func (i Intersection) MarshalYAML() (interface{}, error) {
	qs := make([]queryYAML, len(i))
	for ii, q := range i {
		qs[ii].Query = q
	}
	return qs, nil
}

func (i *Intersection) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var qs []queryYAML
	err := unmarshal(&qs)
	if err == nil {
		*i = (*i)[0:0]
		for _, q := range qs {
			*i = append(*i, q.Query)
		}
	}
	return err
}

type Union []Query

func (u Union) Compile(index FeatureIndex, w World) search.Iterator {
	qs := make(search.Union, len(u))
	for i, q := range u {
		qs[i] = adaptQuery{Query: q, World: w}
	}
	return qs.Compile(index)
}

func (u Union) Matches(f Feature, w World) bool {
	for _, q := range u {
		if q.Matches(f, w) {
			return true
		}
	}
	return false
}

func (u Union) String() string {
	qs := make([]string, len(u))
	for i, q := range u {
		qs[i] = q.String()
	}
	return fmt.Sprintf("(union %s)", strings.Join(qs, ","))
}

func (u Union) ToProto() (*pb.QueryProto, error) {
	qs := make([]*pb.QueryProto, len(u))
	for i, q := range u {
		var err error
		if qs[i], err = q.ToProto(); err != nil {
			return nil, err
		}
	}
	return &pb.QueryProto{
		Query: &pb.QueryProto_Union{
			Union: &pb.QueriesProto{
				Queries: qs,
			},
		},
	}, nil
}

func (u Union) MarshalYAML() (interface{}, error) {
	qs := make([]queryYAML, len(u))
	for ii, q := range u {
		qs[ii].Query = q
	}
	return qs, nil
}

func (u *Union) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var qs []queryYAML
	err := unmarshal(&qs)
	if err == nil {
		*u = (*u)[0:0]
		for _, q := range qs {
			*u = append(*u, q.Query)
		}
	}
	return err
}

type queryChoices struct {
	Tagged         *Tagged
	Keyed          *Keyed
	Typed          *Typed
	Intersection   *Intersection
	Union          *Union
	MightIntersect *MightIntersect
	Cells          *IntersectsCells
	Cap            *IntersectsCap
	Feature        *IntersectsFeature
	Point          *IntersectsPoint
	Polyline       *IntersectsPolyline
	MultiPolygon   *IntersectsMultiPolygon
}

type queryYAML struct {
	Query
}

func (q queryYAML) MarshalYAML() (interface{}, error) {
	return marshalChoiceYAML(&queryChoices{}, q.Query)
}

func (q *queryYAML) UnmarshalYAML(unmarshal func(interface{}) error) error {
	choice, err := unmarshalChoiceYAML(&queryChoices{}, unmarshal)
	if err == nil {
		switch choice := choice.(type) {
		case *Query:
			q.Query = *choice
		case Query:
			q.Query = choice
		default:
			panic("choice wasn't a query")
		}
	}
	return err
}
