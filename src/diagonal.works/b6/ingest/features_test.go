package ingest

import (
	"reflect"
	"testing"

	"diagonal.works/b6"
)

func TestClonePoint(t *testing.T) {
	caravan := osmPoint(2300722786, 51.5357237, -0.1253052)
	caravan.AddTag(b6.Tag{Key: "name", Value: "Caravan"})

	if !reflect.DeepEqual(caravan, caravan.Clone()) {
		t.Errorf("Expected cloned point to be equal")
	}
}

func TestClonePath(t *testing.T) {
	a := osmPoint(7555161584, 51.5345488, -0.1251005)
	b := osmPoint(6384669830, 51.5342291, -0.1262792)

	// Goods Way, South of Granary Square
	ab := osmPath(807924986, []*PointFeature{a, b})
	ab.AddTag(b6.Tag{Key: "highway", Value: "tertiary"})
	ab.AddTag(b6.Tag{Key: "lit", Value: "no"})

	if !reflect.DeepEqual(ab, ab.Clone()) {
		t.Errorf("Expected cloned path to be equal")
	}
}

func TestCloneArea(t *testing.T) {
	lighterman := NewAreaFeature(1)
	lighterman.AreaID = AreaIDFromOSMWayID(427900370)
	lighterman.SetPathIDs(0, []b6.PathID{FromOSMWayID(427900370)})
	lighterman.AddTag(b6.Tag{Key: "name", Value: "The Lighterman"})
	lighterman.AddTag(b6.Tag{Key: "wheelchair", Value: "no"})

	if !reflect.DeepEqual(lighterman, lighterman.Clone()) {
		t.Errorf("Expected cloned area to be equal")
	}
}

func TestCloneRelation(t *testing.T) {
	c6 := NewRelationFeature(2)
	c6.RelationID = FromOSMRelationID(11502000)
	c6.Members[0] = b6.RelationMember{ID: FromOSMWayID(673447480).FeatureID()}
	c6.Members[1] = b6.RelationMember{ID: FromOSMWayID(39035445).FeatureID()}
	c6.Tags = []b6.Tag{{Key: "type", Value: "route"}, {Key: "ref", Value: "C6"}}

	if !reflect.DeepEqual(c6, c6.Clone()) {
		t.Errorf("Expected cloned relation to be equal")
	}
}

func TestMergePoint(t *testing.T) {
	caravan := osmPoint(2300722786, 51.5357237, -0.1253052)
	caravan.AddTag(b6.Tag{Key: "name", Value: "Caravan"})

	lighterman := osmPoint(427900370, 51.5353986, -0.1243711)
	lighterman.AddTag(b6.Tag{Key: "name", Value: "The Lighterman"})
	lighterman.AddTag(b6.Tag{Key: "amenity", Value: "restaurant"})

	m := caravan.Clone()
	m.MergeFrom(lighterman)
	if !reflect.DeepEqual(lighterman, m) {
		t.Errorf("Expected features to be equal after merge")
	}

	m = lighterman.Clone()
	m.MergeFrom(caravan)
	if !reflect.DeepEqual(caravan, m) {
		t.Errorf("Expected features to be equal after merge")
	}
}

func TestMergePath(t *testing.T) {
	a := osmPoint(7555161584, 51.5345488, -0.1251005)
	b := osmPoint(6384669830, 51.5342291, -0.1262792)

	// Goods Way, South of Granary Square
	ab := osmPath(807924986, []*PointFeature{a, b})
	ab.AddTag(b6.Tag{Key: "highway", Value: "tertiary"})
	ab.AddTag(b6.Tag{Key: "lit", Value: "no"})

	c := osmPoint(1715968738, 51.5351015, -0.1248611)
	d := osmPoint(1540349977, 51.5350763, -0.1248251)
	e := osmPoint(1447052073, 51.5350326, -0.1247915)

	// Stable Street, South of Granary Square
	cde := osmPath(642639442, []*PointFeature{c, d, e})
	cde.AddTag(b6.Tag{Key: "highway", Value: "service"})

	m := ab.Clone()
	m.MergeFrom(cde)
	if !reflect.DeepEqual(cde, m) {
		t.Errorf("Expected features to be equal after merge")
	}

	m = cde.Clone()
	m.MergeFrom(ab)
	if !reflect.DeepEqual(ab, m) {
		t.Errorf("Expected features to be equal after merge")
	}
}

func TestMergeArea(t *testing.T) {
	lighterman := NewAreaFeature(1)
	lighterman.AreaID = AreaIDFromOSMWayID(427900370)
	lighterman.SetPathIDs(0, []b6.PathID{FromOSMWayID(427900370)})
	lighterman.AddTag(b6.Tag{Key: "name", Value: "The Lighterman"})
	lighterman.AddTag(b6.Tag{Key: "wheelchair", Value: "no"})

	gasholders := NewAreaFeature(3)
	gasholders.AreaID = AreaIDFromOSMRelationID(7972217)
	gasholders.SetPathIDs(0, []b6.PathID{FromOSMWayID(544908184)})
	gasholders.SetPathIDs(1, []b6.PathID{FromOSMWayID(544908185)})
	gasholders.SetPathIDs(2, []b6.PathID{FromOSMWayID(54490818)})
	gasholders.AddTag(b6.Tag{Key: "name", Value: "Gasholder Apartments"})

	openreach := NewAreaFeature(1)
	openreach.AreaID = AreaIDFromOSMWayID(11095199)
	openreach.SetPathIDs(0, []b6.PathID{FromOSMWayID(803234786), FromOSMWayID(802221851)})
	openreach.AddTag(b6.Tag{Key: "name", Value: "BT Openreach"})
	openreach.AddTag(b6.Tag{Key: "building", Value: "commercial"})
	openreach.AddTag(b6.Tag{Key: "building:levels", Value: "5"})

	comparisons := []struct {
		a *AreaFeature
		b *AreaFeature
	}{
		{lighterman, gasholders},
		{gasholders, lighterman},
		{gasholders, openreach},
		{openreach, gasholders},
	}

	for _, c := range comparisons {
		m := c.a.Clone()
		m.MergeFrom(c.b)
		if !reflect.DeepEqual(c.b, m) {
			t.Errorf("Expected features %q to be equal to %q after merge", c.a.TagOrFallback("name", ""), c.b.TagOrFallback("name", ""))
		}
	}
}

func TestMergeRelation(t *testing.T) {
	c6 := NewRelationFeature(2)
	c6.RelationID = FromOSMRelationID(11502000)
	c6.Members[0] = b6.RelationMember{ID: FromOSMWayID(673447480).FeatureID()}
	c6.Members[1] = b6.RelationMember{ID: FromOSMWayID(39035445).FeatureID()}
	c6.Tags = []b6.Tag{{Key: "type", Value: "route"}, {Key: "ref", Value: "C6"}}

	cs3 := NewRelationFeature(3)
	cs3.RelationID = FromOSMRelationID(12564854)
	cs3.Members[0] = b6.RelationMember{ID: FromOSMWayID(416112693).FeatureID()}
	cs3.Members[1] = b6.RelationMember{ID: FromOSMWayID(431794370).FeatureID()}
	cs3.Members[2] = b6.RelationMember{ID: FromOSMWayID(416815340).FeatureID()}
	cs3.Tags = []b6.Tag{{Key: "type", Value: "route"}, {Key: "ref", Value: "CS3"}, {Key: "name", Value: "East-West Cycle Superhighway"}}

	m := c6.Clone()
	m.MergeFrom(cs3)
	if !reflect.DeepEqual(cs3, m) {
		t.Errorf("Expected features to be equal after merge")
	}

	m = cs3.Clone()
	m.MergeFrom(c6)
	if !reflect.DeepEqual(c6, m) {
		t.Errorf("Expected features to be equal after merge")
	}
}
