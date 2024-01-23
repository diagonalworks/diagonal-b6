package ingest

import (
	"reflect"
	"testing"

	"diagonal.works/b6"
)

func TestRemoveTags(t *testing.T) {
	tags := b6.Tags{
		{Key: "startNode", Value: b6.String("23A9FC0E-CBAB-425C-A7C9-B3356F17AF52")},
		{Key: "roadNumber", Value: b6.String("A5202")},
		{Key: "endNode", Value: b6.String("541F7F78-ED83-40D9-9488-3FD36D169B69")},
		{Key: "class", Value: b6.String("A Road")},
	}
	tags.RemoveTags([]string{"startNode", "endNode"})
	expected := b6.Tags{
		{Key: "roadNumber", Value: b6.String("A5202")},
		{Key: "class", Value: b6.String("A Road")},
	}
	if !reflect.DeepEqual(expected, tags) {
		t.Errorf("Expected %+v, found %+v", expected, tags)
	}
}

func TestClonePoint(t *testing.T) {
	caravan := osmPoint(2300722786, 51.5357237, -0.1253052)
	caravan.AddTag(b6.Tag{Key: "name", Value: b6.String("Caravan")})

	if !reflect.DeepEqual(caravan, caravan.Clone()) {
		t.Errorf("Expected cloned point to be equal")
	}
}

func TestClonePath(t *testing.T) {
	a := osmPoint(7555161584, 51.5345488, -0.1251005)
	b := osmPoint(6384669830, 51.5342291, -0.1262792)

	// Goods Way, South of Granary Square
	ab := osmPath(807924986, []Feature{a, b})
	ab.AddTag(b6.Tag{Key: "highway", Value: b6.String("tertiary")})
	ab.AddTag(b6.Tag{Key: "lit", Value: b6.String("no")})

	if !reflect.DeepEqual(ab, ab.Clone()) {
		t.Errorf("Expected cloned path to be equal")
	}
}

func TestCloneArea(t *testing.T) {
	lighterman := NewAreaFeature(1)
	lighterman.AreaID = AreaIDFromOSMWayID(427900370)
	lighterman.SetPathIDs(0, []b6.PathID{FromOSMWayID(427900370)})
	lighterman.AddTag(b6.Tag{Key: "name", Value: b6.String("The Lighterman")})
	lighterman.AddTag(b6.Tag{Key: "wheelchair", Value: b6.String("no")})

	if !reflect.DeepEqual(lighterman, lighterman.Clone()) {
		t.Errorf("Expected cloned area to be equal")
	}
}

func TestCloneRelation(t *testing.T) {
	c6 := NewRelationFeature(2)
	c6.RelationID = FromOSMRelationID(11502000)
	c6.Members[0] = b6.RelationMember{ID: FromOSMWayID(673447480).FeatureID()}
	c6.Members[1] = b6.RelationMember{ID: FromOSMWayID(39035445).FeatureID()}
	c6.Tags = []b6.Tag{{Key: "type", Value: b6.String("route")}, {Key: "ref", Value: b6.String("C6")}}

	if !reflect.DeepEqual(c6, c6.Clone()) {
		t.Errorf("Expected cloned relation to be equal")
	}
}

func TestCloneCollection(t *testing.T) {
	collection := CollectionFeature{
		CollectionID: b6.MakeCollectionID(b6.NamespacePrivate, 1),
		Keys:         []interface{}{b6.PathID{Namespace: b6.NamespaceDiagonalEntrances, Value: 777}},
		Values:       []interface{}{"i wanna be the one to walk in the sun"},
		Tags:         []b6.Tag{{Key: "by", Value: b6.String("chromatics")}},
	}

	clone := collection.Clone()
	if !reflect.DeepEqual(collection.CollectionID, clone.(*CollectionFeature).CollectionID) {
		t.Errorf("Expected cloned ID to be equal")
	}
	if !reflect.DeepEqual(collection.Keys, clone.(*CollectionFeature).Keys) {
		t.Errorf("Expected cloned keys to be equal")
	}
	if !reflect.DeepEqual(collection.Values, clone.(*CollectionFeature).Values) {
		t.Errorf("Expected cloned values to be equal")
	}
	if !reflect.DeepEqual(collection.AllTags(), clone.AllTags()) {
		t.Errorf("Expected cloned tags to be equal")
	}
}

func TestMergePoint(t *testing.T) {
	caravan := osmPoint(2300722786, 51.5357237, -0.1253052)
	caravan.AddTag(b6.Tag{Key: "name", Value: b6.String("Caravan")})

	lighterman := osmPoint(427900370, 51.5353986, -0.1243711)
	lighterman.AddTag(b6.Tag{Key: "name", Value: b6.String("The Lighterman")})
	lighterman.AddTag(b6.Tag{Key: "amenity", Value: b6.String("restaurant")})

	m := caravan.Clone()
	m.MergeFrom(lighterman)
	if !reflect.DeepEqual(lighterman, m) {
		t.Error("Expected features to be equal after merge")
	}

	m = lighterman.Clone()
	m.MergeFrom(caravan)
	if !reflect.DeepEqual(caravan, m) {
		t.Error("Expected features to be equal after merge")
	}
}

func TestMergePath(t *testing.T) {
	a := osmPoint(7555161584, 51.5345488, -0.1251005)
	b := osmPoint(6384669830, 51.5342291, -0.1262792)

	// Goods Way, South of Granary Square
	ab := osmPath(807924986, []Feature{a, b})
	ab.AddTag(b6.Tag{Key: "highway", Value: b6.String("tertiary")})
	ab.AddTag(b6.Tag{Key: "lit", Value: b6.String("no")})

	c := osmPoint(1715968738, 51.5351015, -0.1248611)
	d := osmPoint(1540349977, 51.5350763, -0.1248251)
	e := osmPoint(1447052073, 51.5350326, -0.1247915)

	// Stable Street, South of Granary Square
	cde := osmPath(642639442, []Feature{c, d, e})
	cde.AddTag(b6.Tag{Key: "highway", Value: b6.String("service")})

	m := ab.Clone()
	m.MergeFrom(cde)
	if !reflect.DeepEqual(cde, m) {
		t.Error("Expected features to be equal after merge")
	}

	m = cde.Clone()
	m.MergeFrom(ab)
	if !reflect.DeepEqual(ab, m) {
		t.Error("Expected features to be equal after merge")
	}
}

func TestMergeArea(t *testing.T) {
	lighterman := NewAreaFeature(1)
	lighterman.AreaID = AreaIDFromOSMWayID(427900370)
	lighterman.SetPathIDs(0, []b6.PathID{FromOSMWayID(427900370)})
	lighterman.AddTag(b6.Tag{Key: "name", Value: b6.String("The Lighterman")})
	lighterman.AddTag(b6.Tag{Key: "wheelchair", Value: b6.String("no")})

	gasholders := NewAreaFeature(3)
	gasholders.AreaID = AreaIDFromOSMRelationID(7972217)
	gasholders.SetPathIDs(0, []b6.PathID{FromOSMWayID(544908184)})
	gasholders.SetPathIDs(1, []b6.PathID{FromOSMWayID(544908185)})
	gasholders.SetPathIDs(2, []b6.PathID{FromOSMWayID(54490818)})
	gasholders.AddTag(b6.Tag{Key: "name", Value: b6.String("Gasholder Apartments")})

	openreach := NewAreaFeature(1)
	openreach.AreaID = AreaIDFromOSMWayID(11095199)
	openreach.SetPathIDs(0, []b6.PathID{FromOSMWayID(803234786), FromOSMWayID(802221851)})
	openreach.AddTag(b6.Tag{Key: "name", Value: b6.String("BT Openreach")})
	openreach.AddTag(b6.Tag{Key: "building", Value: b6.String("commercial")})
	openreach.AddTag(b6.Tag{Key: "building:levels", Value: b6.String("5")})

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
	c6.Tags = []b6.Tag{{Key: "type", Value: b6.String("route")}, {Key: "ref", Value: b6.String("C6")}}

	cs3 := NewRelationFeature(3)
	cs3.RelationID = FromOSMRelationID(12564854)
	cs3.Members[0] = b6.RelationMember{ID: FromOSMWayID(416112693).FeatureID()}
	cs3.Members[1] = b6.RelationMember{ID: FromOSMWayID(431794370).FeatureID()}
	cs3.Members[2] = b6.RelationMember{ID: FromOSMWayID(416815340).FeatureID()}
	cs3.Tags = []b6.Tag{{Key: "type", Value: b6.String("route")}, {Key: "ref", Value: b6.String("CS3")}, {Key: "name", Value: b6.String("East-West Cycle Superhighway")}}

	m := c6.Clone()
	m.MergeFrom(cs3)
	if !reflect.DeepEqual(cs3, m) {
		t.Error("Expected features to be equal after merge")
	}

	m = cs3.Clone()
	m.MergeFrom(c6)
	if !reflect.DeepEqual(c6, m) {
		t.Error("Expected features to be equal after merge")
	}
}

func TestMergeCollection(t *testing.T) {
	before := CollectionFeature{
		CollectionID: b6.MakeCollectionID(b6.NamespacePrivate, 1),
		Keys:         []interface{}{b6.PathID{Namespace: b6.NamespaceDiagonalEntrances, Value: 11}},
		Values:       []interface{}{"i cant find my chill"},
	}

	after := CollectionFeature{
		CollectionID: b6.MakeCollectionID(b6.NamespacePrivate, 1),
		Keys:         []interface{}{b6.PathID{Namespace: b6.NamespaceDiagonalEntrances, Value: 11}},
		Values:       []interface{}{"i must have lost it"},
		Tags:         []b6.Tag{{Key: "carpenter", Value: b6.String("nonsense")}},
	}

	m := before.Clone()
	m.MergeFrom(&after)

	if !reflect.DeepEqual(after.CollectionID, m.(*CollectionFeature).CollectionID) {
		t.Errorf("Expected collection id to be equal after merge")
	}
	if !reflect.DeepEqual(after.Keys, m.(*CollectionFeature).Keys) {
		t.Errorf("Expected keys to be equal after merge")
	}
	if !reflect.DeepEqual(after.Values, m.(*CollectionFeature).Values) {
		t.Errorf("Expected values to be equal after merge")
	}
	if !reflect.DeepEqual(after.AllTags(), m.AllTags()) {
		t.Errorf("Expected features to be equal after merge")
	}
}
