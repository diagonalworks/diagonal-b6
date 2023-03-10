package ingest

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	"diagonal.works/b6"
	"diagonal.works/b6/search"
	"github.com/golang/geo/s2"
)

func sortAndDiffTokens(before []string, after []string) ([]string, []string) {
	added := make([]string, 0, len(after))
	removed := make([]string, 0, len(before))
	sort.Strings(before)
	sort.Strings(after)
	a := 0
	b := 0
	for a < len(after) && b < len(before) {
		switch strings.Compare(after[a], before[b]) {
		case 0:
			a++
			b++
		case -1:
			added = append(added, after[a])
			a++
		case 1:
			removed = append(removed, before[b])
			b++
		}
	}
	for a < len(after) {
		added = append(added, after[a])
		a++
	}
	for b < len(before) {
		removed = append(removed, before[b])
		b++
	}
	return added, removed
}

type MutableWorld interface {
	b6.World
	AddSimplePoint(id b6.PointID, ll s2.LatLng) error
	AddPoint(p *PointFeature) error
	AddPath(p *PathFeature) error
	AddArea(a *AreaFeature) error
	AddRelation(a *RelationFeature) error
	AddTag(id b6.FeatureID, tag b6.Tag) error
	RemoveTag(id b6.FeatureID, key string) error
}

type mutableFeatureIndex struct {
	search.TreeIndex
	byID b6.FeaturesByID
}

func newMutableFeatureIndex(byID b6.FeaturesByID) *mutableFeatureIndex {
	return &mutableFeatureIndex{TreeIndex: *search.NewTreeIndex(featureValues{}), byID: byID}
}

func (m *mutableFeatureIndex) RewriteSpatialQuery(query search.Spatial) search.Query {
	return RewriteSpatialQuery(query)
}

func (m *mutableFeatureIndex) RewriteFeatureTypeQuery(q b6.FeatureTypeQuery) search.Query {
	return RewriteFeatureTypeQuery(q)
}

func (f *mutableFeatureIndex) Feature(v search.Value) b6.Feature {
	switch feature := v.(type) {
	case *PointFeature:
		return newPointFeature(feature)
	case *PathFeature:
		return newPathFeature(feature, f.byID)
	case *AreaFeature:
		return newAreaFeature(feature, f.byID)
	case *RelationFeature:
		return newRelationFeature(feature, f.byID)
	}
	return nil
}

func (f *mutableFeatureIndex) ID(v search.Value) b6.FeatureID {
	switch feature := v.(type) {
	case *PointFeature:
		return feature.FeatureID()
	case *PathFeature:
		return feature.FeatureID()
	case *AreaFeature:
		return feature.FeatureID()
	case *RelationFeature:
		return feature.FeatureID()
	}
	return b6.FeatureIDInvalid
}

type BasicMutableWorld struct {
	byID       *FeaturesByID
	references *FeatureReferences
	index      *mutableFeatureIndex
}

func NewBasicMutableWorld() *BasicMutableWorld {
	byID := NewFeaturesByID()
	w := &BasicMutableWorld{
		byID:       byID,
		references: NewFeatureReferences(),
		index:      newMutableFeatureIndex(byID),
	}
	return w
}

func (m *BasicMutableWorld) FindFeatureByID(id b6.FeatureID) b6.Feature {
	return m.byID.FindFeatureByID(id)
}

func (m *BasicMutableWorld) HasFeatureWithID(id b6.FeatureID) bool {
	return m.byID.HasFeatureWithID(id)
}

func (m *BasicMutableWorld) FindLocationByID(id b6.PointID) (s2.LatLng, bool) {
	return m.byID.FindLocationByID(id)
}

func (m *BasicMutableWorld) FindFeatures(q search.Query) b6.Features {
	// TODO: Iterators created here will be invalidated if the search index is modified.
	// We should keep an epoch number to track whether the world has been modified since
	// the iterator was created, and panic when methods are called on it.
	return b6.NewFeatureIterator(q.Compile(m.index), m.index)
}

func (m *BasicMutableWorld) FindRelationsByFeature(id b6.FeatureID) b6.RelationFeatures {
	return findRelationsByFeature(m.references, id, m)
}

func (m *BasicMutableWorld) FindPathsByPoint(id b6.PointID) b6.PathSegments {
	return findPathsByPoint(m, m.references, id, m)
}

func (m *BasicMutableWorld) FindAreasByPoint(id b6.PointID) b6.AreaFeatures {
	return findAreasByPoint(m.references, id, m)
}

func (m *BasicMutableWorld) EachFeature(each func(f b6.Feature, goroutine int) error, options *b6.EachFeatureOptions) error {
	return m.byID.EachFeature(each, options)
}

func (m *BasicMutableWorld) Tokens() []string {
	return search.AllTokens(m.index.Tokens())
}

func (m *BasicMutableWorld) AddSimplePoint(id b6.PointID, ll s2.LatLng) error {
	if !id.IsValid() {
		return fmt.Errorf("%s: invalid ID", id)
	}
	m.byID.AddSimplePoint(id, ll)
	return nil
}

func (m *BasicMutableWorld) AddPoint(p *PointFeature) error {
	if err := ValidatePoint(p); err != nil {
		return err
	}
	if err := validatePathsWithNewPoint(p, m.byID, m.references, m); err != nil {
		return err
	}
	if err := validateAreasWithNewPoint(p, m.byID, m.references, m); err != nil {
		return err
	}
	features := listFeaturesReferencedByPoint(p, m)
	modified := NewModifiedFeatures(p, features, m.byID, m)
	modified.Update(m.byID, m.references, m.index, m)
	return nil
}

func (m *BasicMutableWorld) AddPath(p *PathFeature) error {
	if err := ValidatePath(p, ClockwisePathsAreInvalid, m); err != nil {
		return err
	}
	if err := validateAreasWithNewPath(p, m.byID, m.references, m); err != nil {
		return err
	}
	features := listFeaturesReferencedByPath(p, m)
	modified := NewModifiedFeatures(p, features, m.byID, m)
	modified.Update(m.byID, m.references, m.index, m)
	return nil
}

func (m *BasicMutableWorld) AddArea(a *AreaFeature) error {
	if err := ValidateArea(a, m); err != nil {
		return err
	}
	modified := NewModifiedFeatures(a, []b6.PhysicalFeature{}, m.byID, m)
	modified.Update(m.byID, m.references, m.index, m)
	return nil
}

func (m *BasicMutableWorld) AddRelation(r *RelationFeature) error {
	if err := ValidateRelation(r); err != nil {
		return err
	}
	modified := NewModifiedFeatures(r, []b6.PhysicalFeature{}, m.byID, m)
	modified.Update(m.byID, m.references, m.index, m)
	return nil
}

func (m *BasicMutableWorld) AddTag(id b6.FeatureID, tag b6.Tag) error {
	tokenAfter, indexedAfter := TokenForTag(tag)
	if f := m.byID.FindMutableFeatureByID(id); f != nil {
		var tokenBefore string
		var indexedBefore bool
		if before := f.Get(tag.Key); before.IsValid() {
			if tokenBefore, indexedBefore = TokenForTag(before); indexedBefore && (!indexedAfter || tokenBefore != tokenAfter) {
				m.index.Remove(f, []string{tokenBefore})
			}
		}
		f.ModifyOrAddTag(tag)
		if indexedAfter && (!indexedBefore || tokenAfter != tokenBefore) {
			m.index.Add(f, []string{tokenAfter})
		}
		return nil
	}
	return fmt.Errorf("No feature with ID %s", id)
}

func (m *BasicMutableWorld) RemoveTag(id b6.FeatureID, key string) error {
	if f := m.byID.FindMutableFeatureByID(id); f != nil {
		if tag := f.Get(key); tag.IsValid() {
			if token, indexed := TokenForTag(tag); indexed {
				m.index.Remove(f, []string{token})
			}
		}
		f.RemoveTag(key)
	}
	return nil
}

func wrapFeature(f Feature, w b6.FeaturesByID) b6.PhysicalFeature {
	switch f := f.(type) {
	case *PointFeature:
		return newPointFeature(f)
	case *PathFeature:
		return newPathFeature(f, w)
	case *AreaFeature:
		return newAreaFeature(f, w)
	case *RelationFeature:
		return newRelationFeature(f, w)
	}
	panic("Invalid feature type")
}

func NewMutableWorldFromSource(source FeatureSource, cores int, validity FeatureValidity) (b6.World, error) {
	w := NewBasicMutableWorld()
	var lock sync.Mutex
	f := func(feature Feature, g int) error {
		feature = feature.Clone()
		lock.Lock()
		switch f := feature.(type) {
		case *PointFeature:
			w.AddPoint(f)
		case *PathFeature:
			w.AddPath(f)
		case *AreaFeature:
			w.AddArea(f)
		case *RelationFeature:
			w.AddRelation(f)
		}
		lock.Unlock()
		return nil
	}
	options := ReadOptions{Parallelism: cores}
	err := source.Read(options, f, context.Background())
	return w, err
}

type modifiedTag struct {
	value   string
	deleted bool
}

func modifyTags(tagged b6.Tagged, modifications map[string]modifiedTag) []b6.Tag {
	// TODO: Consider enforcing AllTags() to return sorted tags, then we could merge
	// sorted lists here.
	original := tagged.AllTags()
	modified := make([]b6.Tag, 0, len(original))
	seen := make(map[string]struct{})
	for _, tag := range original {
		seen[tag.Key] = struct{}{}
		if modification, ok := modifications[tag.Key]; ok {
			if !modification.deleted {
				modified = append(modified, b6.Tag{Key: tag.Key, Value: modification.value})
			}
		} else {
			modified = append(modified, tag)
		}
	}
	for key, modification := range modifications {
		if !modification.deleted {
			if _, ok := seen[key]; !ok {
				modified = append(modified, b6.Tag{Key: key, Value: modification.value})
			}
		}
	}
	return modified

}

func modifyTag(tagged b6.Tagged, key string, modifications map[string]modifiedTag) b6.Tag {
	if modification, ok := modifications[key]; ok {
		if modification.deleted {
			return b6.InvalidTag()
		}
		return b6.Tag{Key: key, Value: modification.value}
	}
	return tagged.Get(key)
}

type modifiedTagsPoint struct {
	b6.PointFeature
	tags map[string]modifiedTag
}

func (m *modifiedTagsPoint) AllTags() []b6.Tag {
	return modifyTags(m.PointFeature, m.tags)
}

func (m *modifiedTagsPoint) Get(key string) b6.Tag {
	return modifyTag(m.PointFeature, key, m.tags)
}

type modifiedTagsPath struct {
	b6.PathFeature
	tags map[string]modifiedTag
}

func (m *modifiedTagsPath) AllTags() []b6.Tag {
	return modifyTags(m.PathFeature, m.tags)
}

func (m *modifiedTagsPath) Get(key string) b6.Tag {
	return modifyTag(m.PathFeature, key, m.tags)
}

type modifiedTagsArea struct {
	b6.AreaFeature
	tags map[string]modifiedTag
}

func (m *modifiedTagsArea) AllTags() []b6.Tag {
	return modifyTags(m.AreaFeature, m.tags)
}

func (m *modifiedTagsArea) Get(key string) b6.Tag {
	return modifyTag(m.AreaFeature, key, m.tags)
}

type modifiedTagsRelation struct {
	b6.RelationFeature
	tags map[string]modifiedTag
}

func (m *modifiedTagsRelation) AllTags() []b6.Tag {
	return modifyTags(m.RelationFeature, m.tags)
}

func (m *modifiedTagsRelation) Get(key string) b6.Tag {
	return modifyTag(m.RelationFeature, key, m.tags)
}

type ModifiedTag struct {
	ID  b6.FeatureID
	Tag b6.Tag
}

type ModifiedTags map[b6.FeatureID]map[string]modifiedTag

func NewModifiedTags() ModifiedTags {
	return make(map[b6.FeatureID]map[string]modifiedTag)
}

func (m ModifiedTags) ModifyOrAddTag(id b6.FeatureID, tag b6.Tag) {
	tags, ok := m[id]
	if !ok {
		tags = make(map[string]modifiedTag)
		m[id] = tags
	}
	tags[tag.Key] = modifiedTag{value: tag.Value, deleted: false}
}

func (m ModifiedTags) RemoveTag(id b6.FeatureID, key string) {
	tags, ok := m[id]
	if !ok {
		tags = make(map[string]modifiedTag)
		m[id] = tags
	}
	tags[key] = modifiedTag{deleted: true}
}

func (m ModifiedTags) WrapFeature(feature b6.Feature) b6.Feature {
	if feature == nil {
		return nil
	}
	if tags, ok := m[feature.FeatureID()]; ok {
		switch f := feature.(type) {
		case b6.PointFeature:
			feature = &modifiedTagsPoint{PointFeature: f, tags: tags}
		case b6.PathFeature:
			feature = &modifiedTagsPath{PathFeature: f, tags: tags}
		case b6.AreaFeature:
			feature = &modifiedTagsArea{AreaFeature: f, tags: tags}
		case b6.RelationFeature:
			feature = &modifiedTagsRelation{RelationFeature: f, tags: tags}
		}
	}
	return feature
}

type modifiedTagsFeatures struct {
	features b6.Features
	m        ModifiedTags
}

func (m *modifiedTagsFeatures) Next() bool {
	return m.features.Next()
}

func (m *modifiedTagsFeatures) Feature() b6.Feature {
	return m.m.WrapFeature(m.features.Feature())
}

func (m *modifiedTagsFeatures) FeatureID() b6.FeatureID {
	return m.features.FeatureID()
}

func (m ModifiedTags) WrapFeatures(features b6.Features) b6.Features {
	return &modifiedTagsFeatures{features: features, m: m}
}

type modifiedTagsRelations struct {
	relations b6.RelationFeatures
	m         ModifiedTags
}

func (m *modifiedTagsRelations) Next() bool {
	return m.relations.Next()
}

func (m *modifiedTagsRelations) Feature() b6.RelationFeature {
	return m.m.WrapFeature(m.relations.Feature()).(b6.RelationFeature)
}

func (m *modifiedTagsRelations) FeatureID() b6.FeatureID {
	return m.relations.FeatureID()
}

func (m *modifiedTagsRelations) RelationID() b6.RelationID {
	return m.relations.RelationID()
}

func (m ModifiedTags) WrapRelations(relations b6.RelationFeatures) b6.RelationFeatures {
	return &modifiedTagsRelations{relations: relations, m: m}
}

type modifiedTagsPathSegments struct {
	segments b6.PathSegments
	m        ModifiedTags
}

func (m *modifiedTagsPathSegments) Next() bool {
	return m.segments.Next()
}

func (m *modifiedTagsPathSegments) PathSegment() b6.PathSegment {
	pathSegment := m.segments.PathSegment()
	return b6.PathSegment{
		PathFeature: m.m.WrapFeature(pathSegment.PathFeature).(b6.PathFeature),
		First:       pathSegment.First,
		Last:        pathSegment.Last,
	}
}

func (m ModifiedTags) WrapPathSegments(segments b6.PathSegments) b6.PathSegments {
	return &modifiedTagsPathSegments{segments: segments, m: m}
}

type modifiedTagsAreas struct {
	areas b6.AreaFeatures
	m     ModifiedTags
}

func (m *modifiedTagsAreas) Next() bool {
	return m.areas.Next()
}

func (m *modifiedTagsAreas) Feature() b6.AreaFeature {
	return m.m.WrapFeature(m.areas.Feature()).(b6.AreaFeature)
}

func (m *modifiedTagsAreas) FeatureID() b6.FeatureID {
	return m.areas.FeatureID()
}

func (m ModifiedTags) WrapAreas(areas b6.AreaFeatures) b6.AreaFeatures {
	return &modifiedTagsAreas{areas: areas, m: m}
}

type MutableOverlayWorld struct {
	byID       *FeaturesByID
	references *FeatureReferences
	index      *mutableFeatureIndex
	base       b6.World
	tags       ModifiedTags
	epoch      int
}

func NewMutableOverlayWorld(base b6.World) *MutableOverlayWorld {
	w := &MutableOverlayWorld{
		byID:       NewFeaturesByID(),
		references: NewFeatureReferences(),
		base:       base,
		tags:       NewModifiedTags(),
		epoch:      0,
	}
	w.index = newMutableFeatureIndex(w)
	return w
}

type mutableFeatureIterator struct {
	i     b6.Features
	epoch int
	w     *MutableOverlayWorld
}

func (f *mutableFeatureIterator) Next() bool {
	if f.epoch < f.w.epoch {
		panic("World modified during query")
	}
	return f.i.Next()
}

func (f *mutableFeatureIterator) Feature() b6.Feature {
	if f.epoch < f.w.epoch {
		panic("World modified during query")
	}
	return f.i.Feature()
}

func (f *mutableFeatureIterator) FeatureID() b6.FeatureID {
	if f.epoch < f.w.epoch {
		panic("World modified during query")
	}
	return f.i.FeatureID()
}

func (m *MutableOverlayWorld) FindFeatures(q search.Query) b6.Features {
	overlay := b6.NewFeatureIterator(q.Compile(m.index), m.index)
	return &mutableFeatureIterator{
		i:     newOverlayFeatures(m.tags.WrapFeatures(m.base.FindFeatures(q)), overlay, m.byID),
		epoch: m.epoch,
		w:     m,
	}
}

func (m *MutableOverlayWorld) FindFeatureByID(id b6.FeatureID) b6.Feature {
	if feature := m.findFeatureByID(id); feature != nil {
		return feature
	}
	return m.tags.WrapFeature(m.base.FindFeatureByID(id))
}

func (m *MutableOverlayWorld) FindLocationByID(id b6.PointID) (s2.LatLng, bool) {
	if ll, ok := m.byID.FindLocationByID(id); ok {
		return ll, true
	}
	return m.base.FindLocationByID(id)
}

func (m *MutableOverlayWorld) HasFeatureWithID(id b6.FeatureID) bool {
	return m.byID.HasFeatureWithID(id) || m.base.HasFeatureWithID(id)
}

func (m *MutableOverlayWorld) findFeatureByID(id b6.FeatureID) b6.Feature {
	switch id.Type {
	case b6.FeatureTypePoint:
		if p, ok := m.byID.Points[id.ToPointID()]; ok {
			return newPointFeature(p)
		}
	case b6.FeatureTypePath:
		if p, ok := m.byID.Paths[id.ToPathID()]; ok {
			return newPathFeature(p, m)
		}
	case b6.FeatureTypeArea:
		if a, ok := m.byID.Areas[id.ToAreaID()]; ok {
			return newAreaFeature(a, m)
		}
	case b6.FeatureTypeRelation:
		if r, ok := m.byID.Relations[id.ToRelationID()]; ok {
			return newRelationFeature(r, m)
		}
	}
	return nil
}

func (m *MutableOverlayWorld) FindRelationsByFeature(id b6.FeatureID) b6.RelationFeatures {
	result := make([]b6.RelationFeature, 0)
	rs := m.base.FindRelationsByFeature(id)
	for rs.Next() {
		if _, ok := m.byID.Relations[rs.Feature().RelationID()]; !ok {
			result = append(result, rs.Feature())
		}
	}
	if rs, ok := m.references.RelationsByFeature[id]; ok {
		for _, r := range rs {
			result = append(result, relationFeature{r, m})
		}
	}
	return &relationFeatures{relations: result, i: -1}
}

func (m *MutableOverlayWorld) FindPathsByPoint(id b6.PointID) b6.PathSegments {
	result := make([]b6.PathSegment, 0)
	ps := m.base.FindPathsByPoint(id)
	for ps.Next() {
		if _, ok := m.byID.Paths[ps.PathSegment().PathID()]; !ok {
			result = append(result, ps.PathSegment())
		}
	}
	ps = findPathsByPoint(m, m.references, id, m)
	for ps.Next() {
		result = append(result, ps.PathSegment())
	}
	return &pathSegments{pathSegments: result, i: -1}
}

func (m *MutableOverlayWorld) FindAreasByPoint(id b6.PointID) b6.AreaFeatures {
	result := make([]b6.AreaFeature, 0)
	as := m.base.FindAreasByPoint(id)
	for as.Next() {
		if !m.byID.HasFeatureWithID(as.FeatureID()) {
			result = append(result, as.Feature())
		}
	}
	for _, a := range m.references.AreasForPoint(id, m) {
		result = append(result, areaFeature{a, m})
	}
	return &areaFeatures{features: result, i: -1}
}

func (m *MutableOverlayWorld) EachFeature(each func(f b6.Feature, goroutine int) error, options *b6.EachFeatureOptions) error {
	if err := m.byID.EachFeature(each, options); err != nil {
		return err
	}
	filter := func(feature b6.Feature, goroutine int) error {
		if !m.byID.HasFeatureWithID(feature.FeatureID()) {
			return each(feature, goroutine)
		}
		return nil
	}
	return m.base.EachFeature(filter, options)
}

func (m *MutableOverlayWorld) Tokens() []string {
	tokens := make(map[string]struct{})
	for _, token := range m.base.Tokens() {
		tokens[token] = struct{}{}
	}
	overlaid := m.index.Tokens()
	for overlaid.Next() {
		tokens[overlaid.Token()] = struct{}{}
	}
	all := make([]string, 0, len(tokens))
	for token, _ := range tokens {
		all = append(all, token)
	}
	return all
}

func (m *MutableOverlayWorld) AddSimplePoint(id b6.PointID, ll s2.LatLng) error {
	// TODO: This could be improved
	return m.AddPoint(NewPointFeature(id, ll))
}

func (m *MutableOverlayWorld) AddPoint(p *PointFeature) error {
	if err := ValidatePoint(p); err != nil {
		return err
	}
	features := listFeaturesReferencedByPoint(p, m)
	modified := NewModifiedFeaturesWithCopies(p, features, m.byID, m)
	if err := validatePathsWithNewPoint(p, m.byID, m.references, m); err != nil {
		return err
	}
	if err := validateAreasWithNewPoint(p, m.byID, m.references, m); err != nil {
		return err
	}
	modified.Update(m.byID, m.references, m.index, m)
	delete(m.tags, p.FeatureID())
	m.epoch++
	return nil
}

func (m *MutableOverlayWorld) AddPath(p *PathFeature) error {
	if err := ValidatePath(p, ClockwisePathsAreInvalid, m); err != nil {
		return err
	}
	features := listFeaturesReferencedByPath(p, m)
	modified := NewModifiedFeaturesWithCopies(p, features, m.byID, m)
	if err := validateAreasWithNewPath(p, m.byID, m.references, m); err != nil {
		return err
	}
	modified.Update(m.byID, m.references, m.index, m)
	delete(m.tags, p.FeatureID())
	m.epoch++
	return nil
}

func (m *MutableOverlayWorld) AddArea(a *AreaFeature) error {
	if err := ValidateArea(a, m); err != nil {
		return err
	}
	modified := NewModifiedFeatures(a, []b6.PhysicalFeature{}, m.byID, m)
	modified.Update(m.byID, m.references, m.index, m)
	delete(m.tags, a.FeatureID())
	m.epoch++
	return nil
}

func (m *MutableOverlayWorld) AddRelation(r *RelationFeature) error {
	if err := ValidateRelation(r); err != nil {
		return err
	}
	modified := NewModifiedFeatures(r, []b6.PhysicalFeature{}, m.byID, m)
	modified.Update(m.byID, m.references, m.index, m)
	delete(m.tags, r.FeatureID())
	m.epoch++
	return nil
}

func (m *MutableOverlayWorld) AddTag(id b6.FeatureID, tag b6.Tag) error {
	tokenAfter, indexedAfter := TokenForTag(tag)
	if f := m.byID.FindMutableFeatureByID(id); f != nil {
		var tokenBefore string
		var indexedBefore bool
		if before := f.Get(tag.Key); before.IsValid() {
			if tokenBefore, indexedBefore = TokenForTag(before); indexedBefore && (!indexedAfter || tokenBefore != tokenAfter) {
				m.index.Remove(f, []string{tokenBefore})
			}
		}
		f.ModifyOrAddTag(tag)
		if indexedAfter && (!indexedBefore || tokenBefore != tokenAfter) {
			m.index.Add(f, []string{tokenAfter})
		}
	} else {
		base := m.base.FindFeatureByID(id)
		if base == nil {
			return fmt.Errorf("No feature with ID %s", id)
		}
		if indexedAfter {
			f = NewFeatureFromWorld(base)
			f.ModifyOrAddTag(tag)
			m.byID.AddFeature(f)
			m.references.AddFeature(f, m)
			m.index.Add(f, TokensForFeature(wrapFeature(f, m)))
		} else {
			m.tags.ModifyOrAddTag(id, tag)
		}
	}
	return nil
}

func (m *MutableOverlayWorld) RemoveTag(id b6.FeatureID, key string) error {
	if f := m.byID.FindMutableFeatureByID(id); f != nil {
		if tag := f.Get(key); tag.IsValid() {
			if token, indexed := TokenForTag(tag); indexed {
				m.index.Remove(f, []string{token})
			}
		}
		f.RemoveTag(key)
	} else {
		base := m.base.FindFeatureByID(id)
		if base == nil {
			return fmt.Errorf("No feature with ID %s", id)
		}
		if tag := base.Get(key); tag.IsValid() {
			if _, indexed := TokenForTag(tag); indexed {
				f = NewFeatureFromWorld(base)
				f.RemoveTag(key)
				m.byID.AddFeature(f)
				m.references.AddFeature(f, m)
				m.index.Add(f, TokensForFeature(wrapFeature(f, m)))
			} else {
				m.tags.RemoveTag(id, key)
			}
		}
	}
	return nil
}

func (m *MutableOverlayWorld) MergeSource(source FeatureSource) error {
	emit := func(f Feature, goroutine int) error {
		switch f := f.(type) {
		case *PointFeature:
			return m.AddPoint(f)
		case *PathFeature:
			return m.AddPath(f)
		case *AreaFeature:
			return m.AddArea(f)
		case *RelationFeature:
			return m.AddRelation(f)
		}
		return nil
	}
	return source.Read(ReadOptions{}, emit, context.Background())
}

func validatePathsWithNewPoint(new *PointFeature, byID *FeaturesByID, references *FeatureReferences, w b6.World) error {
	existing := byID.Points[new.PointID]
	if existing == nil {
		return nil
	}
	byID.Points[new.PointID] = new
	err := validatePaths(references.PathsByPoint[existing.PointID], w)
	if err == nil {
		err = validateAreas(references.AreasForPoint(existing.PointID, w), w)
	}
	byID.Points[new.PointID] = existing
	return err
}

func validatePaths(paths []PathPosition, w b6.World) error {
	seen := make(map[b6.PathID]struct{})
	for _, path := range paths {
		if _, ok := seen[path.Path.PathID]; !ok {
			if err := ValidatePath(path.Path, ClockwisePathsAreInvalid, w); err != nil {
				return err
			}
			seen[path.Path.PathID] = struct{}{}
		}
	}
	return nil
}

func validateAreasWithNewPoint(new *PointFeature, byID *FeaturesByID, references *FeatureReferences, w b6.World) error {
	existing := byID.Points[new.PointID]
	if existing == nil {
		return nil
	}
	byID.Points[new.PointID] = new
	err := validateAreas(references.AreasForPoint(existing.PointID, w), w)
	byID.Points[new.PointID] = existing
	return err
}

func validateAreasWithNewPath(new *PathFeature, byID *FeaturesByID, references *FeatureReferences, w b6.World) error {
	existing := byID.Paths[new.PathID]
	if existing == nil {
		return nil
	}
	byID.Paths[new.PathID] = new
	err := validateAreas(references.AreasForPath(existing.PathID), w)
	byID.Paths[new.PathID] = existing
	return err
}

func validateAreas(areas []*AreaFeature, w b6.World) error {
	for _, area := range areas {
		if err := ValidateArea(area, w); err != nil {
			return err
		}
	}
	return nil
}

// MergeInto adds the features from this world to other. Merging is not atomic.
// If validation fails (for example, because adding a feature would invalidate
// an existing feature in other), other will be left with only some features
// added.
func (m *MutableOverlayWorld) MergeInto(other MutableWorld) error {
	// TODO: this could (perhaps) be made more efficient if necessary,
	// since the features below have already been validated in the
	// context of this world. Resticting merging to an original 'parent'
	// would, and checking an epoch number for changes, could enable
	// a simple copy.
	for _, point := range m.byID.Points {
		if err := other.AddPoint(point); err != nil {
			return err
		}
	}
	for _, path := range m.byID.Paths {
		if err := other.AddPath(path); err != nil {
			return err
		}
	}
	for _, area := range m.byID.Areas {
		if err := other.AddArea(area); err != nil {
			return err
		}
	}
	for _, relation := range m.byID.Relations {
		if err := other.AddRelation(relation); err != nil {
			return err
		}
	}
	return nil
}

func (m *MutableOverlayWorld) Snapshot() b6.World {
	copy := *m
	m.base = &copy
	m.byID = NewFeaturesByID()
	m.references = NewFeatureReferences()
	m.tags = NewModifiedTags()
	m.index = newMutableFeatureIndex(m)
	return m.base
}

func listFeaturesReferencedByPoint(p *PointFeature, w b6.World) []b6.PhysicalFeature {
	features := make([]b6.PhysicalFeature, 0)
	ps := w.FindPathsByPoint(p.PointID)
	// Closed paths, and points in the middle of paths, will generate two segments
	// for a single point, corresponding to the two different traversal directions,
	// therefore we need to deduplicate the PathFeatures in those segments.
	seen := make(map[b6.PathID]struct{})
	for ps.Next() {
		segment := ps.PathSegment()
		if _, ok := seen[segment.PathID()]; !ok {
			features = append(features, segment.PathFeature)
			seen[segment.PathID()] = struct{}{}
		}
	}
	as := w.FindAreasByPoint(p.PointID)
	for as.Next() {
		features = append(features, as.Feature())
	}
	return features
}

func listFeaturesReferencedByPath(p *PathFeature, w b6.World) []b6.PhysicalFeature {
	features := make([]b6.PhysicalFeature, 0)
	if existing := b6.FindPathByID(p.PathID, w); existing != nil {
		for i := 0; i < existing.Len(); i++ {
			if point := existing.Feature(i); point != nil {
				as := w.FindAreasByPoint(point.PointID())
				for as.Next() {
					features = append(features, as.Feature())
				}
				break
			}
		}
	}
	return features
}

type ModifiedFeatures struct {
	features []Feature
	tokens   [][]string
	copied   []bool
	existing Feature
}

func NewModifiedFeatures(new Feature, features []b6.PhysicalFeature, byID *FeaturesByID, w b6.World) *ModifiedFeatures {
	m := &ModifiedFeatures{
		features: make([]Feature, 0, len(features)+1),
		tokens:   make([][]string, 0, len(features)+1),
		copied:   make([]bool, 0, len(features)+1),
	}
	m.features = append(m.features, new)
	if m.existing = byID.FindMutableFeatureByID(new.FeatureID()); m.existing != nil {
		m.tokens = append(m.tokens, TokensForFeature(wrapFeature(m.existing, w)))
	} else {
		m.tokens = append(m.tokens, []string{})
	}
	m.copied = append(m.copied, false)
	for _, f := range features {
		if existing := byID.FindMutableFeatureByID(f.FeatureID()); existing != nil {
			m.features = append(m.features, existing)
			m.tokens = append(m.tokens, TokensForFeature(f))
			m.copied = append(m.copied, false)
		}
	}
	return m
}

func NewModifiedFeaturesWithCopies(new Feature, features []b6.PhysicalFeature, byID *FeaturesByID, w b6.World) *ModifiedFeatures {
	m := &ModifiedFeatures{
		features: make([]Feature, 0, len(features)+1),
		tokens:   make([][]string, 0, len(features)+1),
		copied:   make([]bool, 0, len(features)+1),
	}
	m.features = append(m.features, new)
	if m.existing = byID.FindMutableFeatureByID(new.FeatureID()); m.existing != nil {
		m.tokens = append(m.tokens, TokensForFeature(wrapFeature(m.existing, w)))
	} else {
		m.tokens = append(m.tokens, []string{})
	}
	m.copied = append(m.copied, false)
	for _, f := range features {
		if existing := byID.FindMutableFeatureByID(f.FeatureID()); existing != nil {
			m.features = append(m.features, existing)
			m.tokens = append(m.tokens, TokensForFeature(f))
			m.copied = append(m.copied, false)
		} else {
			copy := NewFeatureFromWorld(f)
			byID.AddFeature(copy)
			m.features = append(m.features, copy)
			m.tokens = append(m.tokens, []string{})
			m.copied = append(m.copied, true)
		}
	}
	return m
}

func (m *ModifiedFeatures) RemoveReferences(byID b6.FeaturesByID, references *FeatureReferences) {
	if m.existing != nil {
		references.RemoveFeature(m.existing, byID)
	}
	for i, f := range m.features[1:] {
		if !m.copied[i+1] {
			references.RemoveFeature(f, byID)
		}
	}
}

func (m *ModifiedFeatures) Update(features *FeaturesByID, references *FeatureReferences, index *mutableFeatureIndex, byID b6.FeaturesByID) {
	m.RemoveReferences(byID, references)
	var new Feature
	if m.existing != nil {
		switch existing := m.existing.(type) {
		case *PointFeature:
			existing.MergeFrom(m.features[0].(*PointFeature))
		case *PathFeature:
			existing.MergeFrom(m.features[0].(*PathFeature))
		case *AreaFeature:
			existing.MergeFrom(m.features[0].(*AreaFeature))
		case *RelationFeature:
			existing.MergeFrom(m.features[0].(*RelationFeature))
		}
		new = m.existing
	} else {
		new = m.features[0].Clone()
		features.AddFeature(new)
	}
	m.features[0] = new
	m.AddReferences(byID, references)
	m.UpdateIndex(index, byID)
}

func (m *ModifiedFeatures) AddReferences(byID b6.FeaturesByID, references *FeatureReferences) {
	for _, f := range m.features {
		references.AddFeature(f, byID)
	}
}

func (m *ModifiedFeatures) UpdateIndex(index *mutableFeatureIndex, byID b6.FeaturesByID) {
	for i, f := range m.features {
		added, removed := sortAndDiffTokens(m.tokens[i], TokensForFeature(wrapFeature(f, byID)))
		index.Remove(f, removed)
		index.Add(f, added)
	}
}

type watcher struct {
	c    chan ModifiedTag
	done bool
	lock sync.Mutex
}

// An implementation of b6.World that allows the tags of features in a base
// world to be changed.
// TODO: This doesn't currently update the search index. Consider whether
// that's OK or not. We would need to separate out tokens that are generated
// from geometry, and tokens that are generated from tags. Maybe the best
// approach is to only use a simple overlay for tags that don't affect the
// search index, and fall back to MutableWorld if they do? That may lead
// to merging the two implementations.
type MutableTagsOverlayWorld struct {
	tags     ModifiedTags
	base     b6.World
	watchers []*watcher
}

func NewMutableTagsOverlayWorld(base b6.World) *MutableTagsOverlayWorld {
	return &MutableTagsOverlayWorld{
		tags:     NewModifiedTags(),
		base:     base,
		watchers: make([]*watcher, 0),
	}
}

func (m *MutableTagsOverlayWorld) AddTag(id b6.FeatureID, tag b6.Tag) {
	var tags map[string]modifiedTag
	var ok bool
	if tags, ok = m.tags[id]; !ok {
		tags = make(map[string]modifiedTag)
		m.tags[id] = tags
	}
	tags[tag.Key] = modifiedTag{value: tag.Value, deleted: false}
	m.notifyWatchers(id, tag)
}

func (m *MutableTagsOverlayWorld) notifyWatchers(id b6.FeatureID, tag b6.Tag) {
	modification := ModifiedTag{ID: id, Tag: tag}
	notified := make([]*watcher, 0, len(m.watchers))
	for _, w := range m.watchers {
		w.lock.Lock()
		if !w.done {
			select {
			case w.c <- modification:
			default:
			}
			notified = append(notified, w)
		} else {
			close(w.c)
		}
		w.lock.Unlock()
	}
	m.watchers = notified
}

func (m *MutableTagsOverlayWorld) Watch() (<-chan ModifiedTag, context.CancelFunc) {
	w := &watcher{c: make(chan ModifiedTag, 2048), done: false}
	m.watchers = append(m.watchers, w)
	cancel := context.CancelFunc(func() {
		w.lock.Lock()
		defer w.lock.Unlock()
		w.done = true
	})
	return w.c, cancel
}

// Snapshot returns a readonly snapshot of the current state of the world, unaffected
// by any future changes.
// TODO: This currently isn't particularly efficent, as we don't compact unused
// snapshots once we've finished with them. We should do this (by returning a done()
// callback along with the world?) if the pattern proves to be useful.
func (m *MutableTagsOverlayWorld) Snapshot() b6.World {
	m.base = &MutableTagsOverlayWorld{
		tags:     m.tags,
		base:     m.base,
		watchers: make([]*watcher, 0),
	}
	m.tags = NewModifiedTags()
	return m.base
}

func (m *MutableTagsOverlayWorld) FindFeatureByID(id b6.FeatureID) b6.Feature {
	return m.tags.WrapFeature(m.base.FindFeatureByID(id))
}

func (m *MutableTagsOverlayWorld) HasFeatureWithID(id b6.FeatureID) bool {
	return m.base.HasFeatureWithID(id)
}

func (m *MutableTagsOverlayWorld) FindLocationByID(id b6.PointID) (s2.LatLng, bool) {
	return m.base.FindLocationByID(id)
}

func (m *MutableTagsOverlayWorld) FindFeatures(query search.Query) b6.Features {
	return m.tags.WrapFeatures(m.base.FindFeatures(query))
}

func (m *MutableTagsOverlayWorld) FindRelationsByFeature(id b6.FeatureID) b6.RelationFeatures {
	return m.tags.WrapRelations(m.base.FindRelationsByFeature(id))
}

func (m *MutableTagsOverlayWorld) FindPathsByPoint(id b6.PointID) b6.PathSegments {
	return m.tags.WrapPathSegments(m.base.FindPathsByPoint(id))
}

func (m *MutableTagsOverlayWorld) FindAreasByPoint(id b6.PointID) b6.AreaFeatures {
	return m.tags.WrapAreas(m.base.FindAreasByPoint(id))
}

func (m *MutableTagsOverlayWorld) EachFeature(each func(f b6.Feature, goroutine int) error, options *b6.EachFeatureOptions) error {
	wrap := func(f b6.Feature, goroutine int) error {
		return each(m.tags.WrapFeature(f), goroutine)
	}
	return m.base.EachFeature(wrap, options)
}

func (m *MutableTagsOverlayWorld) Tokens() []string {
	return m.base.Tokens()
}
