package ingest

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sort"
	"strings"
	"sync"

	"diagonal.works/b6"
	"diagonal.works/b6/search"
	"github.com/golang/geo/s2"
	"golang.org/x/sync/errgroup"
)

type MutableWorld interface {
	b6.World

	AddFeature(f Feature) error

	AddTag(id b6.FeatureID, tag b6.Tag) error
	RemoveTag(id b6.FeatureID, key string) error
	EachModifiedFeature(each func(f b6.Feature, goroutine int) error, options *b6.EachFeatureOptions) error
	EachModifiedTag(each func(f ModifiedTag, goroutine int) error, options *b6.EachFeatureOptions) error
}

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

type ReadOnlyWorld struct {
	World b6.World
}

func (r ReadOnlyWorld) FindFeatureByID(id b6.FeatureID) b6.Feature {
	return r.World.FindFeatureByID(id)
}

func (r ReadOnlyWorld) HasFeatureWithID(id b6.FeatureID) bool {
	return r.World.HasFeatureWithID(id)
}

func (r ReadOnlyWorld) FindLocationByID(id b6.FeatureID) (s2.LatLng, error) {
	return r.World.FindLocationByID(id)
}

func (r ReadOnlyWorld) FindFeatures(query b6.Query) b6.Features {
	return r.World.FindFeatures(query)
}

func (r ReadOnlyWorld) FindRelationsByFeature(id b6.FeatureID) b6.RelationFeatures {
	return r.World.FindRelationsByFeature(id)
}

func (r ReadOnlyWorld) FindCollectionsByFeature(id b6.FeatureID) b6.CollectionFeatures {
	return r.World.FindCollectionsByFeature(id)
}

func (r ReadOnlyWorld) FindAreasByPoint(id b6.FeatureID) b6.AreaFeatures {
	return r.World.FindAreasByPoint(id)
}

func (r ReadOnlyWorld) FindReferences(id b6.FeatureID, typed ...b6.FeatureType) b6.Features {
	return r.World.FindReferences(id, typed...)
}

func (r ReadOnlyWorld) Traverse(id b6.FeatureID) b6.Segments {
	return r.World.Traverse(id)
}

func (r ReadOnlyWorld) EachFeature(each func(f b6.Feature, goroutine int) error, options *b6.EachFeatureOptions) error {
	return r.World.EachFeature(each, options)
}

func (r ReadOnlyWorld) EachModifiedFeature(each func(f b6.Feature, goroutine int) error, options *b6.EachFeatureOptions) error {
	return nil
}

func (r ReadOnlyWorld) EachModifiedTag(each func(f ModifiedTag, goroutine int) error, options *b6.EachFeatureOptions) error {
	return nil
}

func (r ReadOnlyWorld) Tokens() []string {
	return r.World.Tokens()
}

func (r ReadOnlyWorld) AddFeature(f Feature) error {
	return errors.New("World is read-only")
}

func (r ReadOnlyWorld) AddTag(id b6.FeatureID, tag b6.Tag) error {
	return errors.New("World is read-only")
}

func (r ReadOnlyWorld) RemoveTag(id b6.FeatureID, key string) error {
	return errors.New("World is read-only")
}

type mutableFeatureIndex struct {
	search.TreeIndex
	features b6.FeaturesByID
}

func newMutableFeatureIndex(features b6.FeaturesByID) *mutableFeatureIndex {
	return &mutableFeatureIndex{TreeIndex: *search.NewTreeIndex(featureValues{}), features: features}
}

func (f *mutableFeatureIndex) Feature(v search.Value) b6.Feature {
	if feature, ok := v.(Feature); ok {
		return WrapFeature(feature, f.features)
	}
	panic("Not a feature")
}

func (f *mutableFeatureIndex) ID(v search.Value) b6.FeatureID {
	switch feature := v.(type) {
	case b6.Identifiable:
		return feature.FeatureID()
	}
	return b6.FeatureIDInvalid
}

type BasicMutableWorld struct {
	features   *FeaturesByID
	references *FeatureReferencesByID
	index      *mutableFeatureIndex
}

func NewBasicMutableWorld() *BasicMutableWorld {
	features := NewFeaturesByID()
	w := &BasicMutableWorld{
		features:   features,
		references: NewFeatureReferences(),
		index:      newMutableFeatureIndex(features),
	}
	return w
}

func (m *BasicMutableWorld) FindFeatureByID(id b6.FeatureID) b6.Feature {
	return m.features.FindFeatureByID(id)
}

func (m *BasicMutableWorld) HasFeatureWithID(id b6.FeatureID) bool {
	return m.features.HasFeatureWithID(id)
}

func (m *BasicMutableWorld) FindLocationByID(id b6.FeatureID) (s2.LatLng, error) {
	return m.features.FindLocationByID(id)
}

func (m *BasicMutableWorld) FindFeatures(q b6.Query) b6.Features {
	// TODO: Iterators created here will be invalidated if the search index is modified.
	// We should keep an epoch number to track whether the world has been modified since
	// the iterator was created, and panic when methods are called on it.
	return b6.NewSearchFeatureIterator(q.Compile(m.index, m), m.index)
}

func (m *BasicMutableWorld) FindRelationsByFeature(id b6.FeatureID) b6.RelationFeatures {
	var features []b6.RelationFeature
	references := m.FindReferences(id, b6.FeatureTypeRelation)
	for references.Next() {
		features = append(features, references.Feature().(b6.RelationFeature))
	}

	return NewRelationFeatureIterator(features)
}

func (m *BasicMutableWorld) FindCollectionsByFeature(id b6.FeatureID) b6.CollectionFeatures {
	var features []b6.CollectionFeature
	references := m.FindReferences(id, b6.FeatureTypeCollection)
	for references.Next() {
		features = append(features, references.Feature().(b6.CollectionFeature))
	}

	return NewCollectionFeatureIterator(features)
}

func (m *BasicMutableWorld) FindAreasByPoint(id b6.FeatureID) b6.AreaFeatures {
	references := m.FindReferences(id, b6.FeatureTypeArea)
	var features []b6.AreaFeature
	for references.Next() {
		features = append(features, references.Feature().(b6.AreaFeature))
	}

	return NewAreaFeatureIterator(features)
}

func (m *BasicMutableWorld) FindReferences(id b6.FeatureID, typed ...b6.FeatureType) b6.Features {
	references := m.references.FindReferences(id, typed...)

	features := make([]b6.Feature, 0, len(references))
	for _, reference := range references {
		if feature := m.FindFeatureByID(reference.Source()); feature != nil {
			if !slices.Contains(features, feature) {
				features = append(features, feature)
			}
		}
	}

	return b6.NewFeatureIterator(features)
}

func (m *BasicMutableWorld) Traverse(origin b6.FeatureID) b6.Segments {
	return NewSegmentIterator(traverse(origin, m, m.references))
}

func (m *BasicMutableWorld) EachFeature(each func(f b6.Feature, goroutine int) error, options *b6.EachFeatureOptions) error {
	return EachFeature(each, m.features, m.references, options)
}

func (m *BasicMutableWorld) EachModifiedFeature(each func(f b6.Feature, goroutine int) error, options *b6.EachFeatureOptions) error {
	return m.EachFeature(each, options)
}

func (m *BasicMutableWorld) EachModifiedTag(each func(f ModifiedTag, goroutine int) error, options *b6.EachFeatureOptions) error {
	return nil // There are no modified tags, as we store every feature.
}

func (m *BasicMutableWorld) Tokens() []string {
	return search.AllTokens(m.index.Tokens())
}

func (m *BasicMutableWorld) AddFeature(f Feature) error {
	if err := ValidateFeature(f, &ValidateOptions{InvertClockwisePaths: false}, m); err != nil {
		return err
	}

	existing := (*m.features)[f.FeatureID()]
	references := allReferences(f, m)
	if existing != nil {
		(*m.features)[f.FeatureID()] = f
		for _, reference := range references {
			if err := ValidateFeature(NewFeatureFromWorld(reference), &ValidateOptions{InvertClockwisePaths: false}, m); err != nil {
				return err
			}
		}
		(*m.features)[f.FeatureID()] = existing
	}

	modified := NewModifiedFeatures(f, references, m.features, m)
	modified.Update(m.features, m.references, m.index, m)
	return nil
}

func (m *BasicMutableWorld) AddTag(id b6.FeatureID, tag b6.Tag) error {
	tokenAfter, indexedAfter := b6.TokenForTag(tag)
	if f := m.features.FindMutableFeatureByID(id); f != nil {
		var tokenBefore string
		var indexedBefore bool
		if before := f.Get(tag.Key); before.IsValid() {
			if tokenBefore, indexedBefore = b6.TokenForTag(before); indexedBefore && (!indexedAfter || tokenBefore != tokenAfter) {
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
	if f := m.features.FindMutableFeatureByID(id); f != nil {
		if tag := f.Get(key); tag.IsValid() {
			if token, indexed := b6.TokenForTag(tag); indexed {
				m.index.Remove(f, []string{token})
			}
		}
		f.RemoveTag(key)
	}
	return nil
}

func NewMutableWorldFromSource(o *BuildOptions, source FeatureSource) (b6.World, error) {
	w := NewBasicMutableWorld()
	var lock sync.Mutex
	f := func(feature Feature, g int) error {
		feature = feature.Clone()
		lock.Lock()
		w.AddFeature(feature)
		lock.Unlock()
		return nil
	}
	options := ReadOptions{Goroutines: o.Cores}
	err := source.Read(options, f, context.Background())
	return w, err
}

type modifiedTag struct {
	value   string
	deleted bool
}

func modifyTags(t b6.Taggable, modifications map[string]modifiedTag) []b6.Tag {
	// TODO: Consider enforcing AllTags() to return sorted tags, then we could merge
	// sorted lists here.
	original := t.AllTags()
	modified := make([]b6.Tag, 0, len(original))
	seen := make(map[string]struct{})
	for _, tag := range original {
		seen[tag.Key] = struct{}{}
		if modifications != nil {
			if modification, ok := modifications[tag.Key]; ok {
				if !modification.deleted {
					modified = append(modified, b6.Tag{Key: tag.Key, Value: b6.StringExpression(modification.value)})
				}
			} else {
				modified = append(modified, tag)
			}
		} else {
			modified = append(modified, tag)
		}
	}

	for key, modification := range modifications {
		if !modification.deleted {
			if _, ok := seen[key]; !ok {
				modified = append(modified, b6.Tag{Key: key, Value: b6.StringExpression(modification.value)})
			}
		}
	}
	return modified

}

func modifyTag(t b6.Taggable, key string, modifications map[string]modifiedTag) b6.Tag {
	if modifications != nil {
		if modification, ok := modifications[key]; ok {
			if modification.deleted {
				return b6.InvalidTag()
			}
			return b6.Tag{Key: key, Value: b6.StringExpression(modification.value)}
		}
	}

	return t.Get(key)
}

type modifiedPhysicalFeature struct {
	b6.PhysicalFeature
	tags ModifiedTags
}

func (m *modifiedPhysicalFeature) AllTags() b6.Tags {
	return modifyTags(m.PhysicalFeature, m.tags[m.FeatureID()])
}

func (m *modifiedPhysicalFeature) Get(key string) b6.Tag {
	return modifyTag(m.PhysicalFeature, key, m.tags[m.FeatureID()])
}

type modifiedTagsArea struct {
	b6.AreaFeature
	tags ModifiedTags
}

func (m *modifiedTagsArea) AllTags() b6.Tags {
	return modifyTags(m.AreaFeature, m.tags[m.AreaFeature.FeatureID()])
}

func (m *modifiedTagsArea) Get(key string) b6.Tag {
	return modifyTag(m.AreaFeature, key, m.tags[m.AreaFeature.FeatureID()])
}

func (m *modifiedTagsArea) Feature(i int) []b6.PhysicalFeature {
	if f := m.AreaFeature.Feature(i); f != nil {
		wrapped := make([]b6.PhysicalFeature, len(f))
		for j, p := range f {
			wrapped[j] = &modifiedPhysicalFeature{p, m.tags}
		}
		return wrapped
	}
	return nil
}

type modifiedTagsRelation struct {
	b6.RelationFeature
	tags ModifiedTags
}

func (m *modifiedTagsRelation) AllTags() b6.Tags {
	return modifyTags(m.RelationFeature, m.tags[m.RelationFeature.FeatureID()])
}

func (m *modifiedTagsRelation) Get(key string) b6.Tag {
	return modifyTag(m.RelationFeature, key, m.tags[m.RelationFeature.FeatureID()])
}

type modifiedTagsCollection struct {
	b6.CollectionFeature
	tags ModifiedTags
}

func (m *modifiedTagsCollection) AllTags() b6.Tags {
	return modifyTags(m.CollectionFeature, m.tags[m.CollectionFeature.FeatureID()])
}

func (m *modifiedTagsCollection) Get(key string) b6.Tag {
	return modifyTag(m.CollectionFeature, key, m.tags[m.CollectionFeature.FeatureID()])
}

type modifiedTagsExpression struct {
	b6.ExpressionFeature
	tags ModifiedTags
}

func (m *modifiedTagsExpression) AllTags() b6.Tags {
	return modifyTags(m.ExpressionFeature, m.tags[m.ExpressionFeature.FeatureID()])
}

func (m *modifiedTagsExpression) Get(key string) b6.Tag {
	return modifyTag(m.ExpressionFeature, key, m.tags[m.ExpressionFeature.FeatureID()])
}

type ModifiedTag struct {
	ID      b6.FeatureID
	Tag     b6.Tag
	Deleted bool
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
	tags[tag.Key] = modifiedTag{value: tag.Value.String(), deleted: false}
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
	switch f := feature.(type) {
	case b6.AreaFeature:
		return m.WrapAreaFeature(f)
	case b6.RelationFeature:
		return m.WrapRelationFeature(f)
	case b6.CollectionFeature:
		return m.WrapCollectionFeature(f)
	case b6.ExpressionFeature:
		return m.WrapExpressionFeature(f)
	case b6.PhysicalFeature:
		return m.WrapPhysicalFeature(f)
	}
	panic(fmt.Sprintf("Can't wrap %T", feature))
}

func (m ModifiedTags) WrapPhysicalFeature(f b6.PhysicalFeature) b6.PhysicalFeature {
	return &modifiedPhysicalFeature{PhysicalFeature: f, tags: m}
}

func (m ModifiedTags) WrapAreaFeature(f b6.AreaFeature) b6.AreaFeature {
	return &modifiedTagsArea{AreaFeature: f, tags: m}
}

func (m ModifiedTags) WrapRelationFeature(f b6.RelationFeature) b6.RelationFeature {
	return &modifiedTagsRelation{RelationFeature: f, tags: m}
}

func (m ModifiedTags) WrapCollectionFeature(f b6.CollectionFeature) b6.CollectionFeature {
	return &modifiedTagsCollection{CollectionFeature: f, tags: m}
}

func (m ModifiedTags) WrapExpressionFeature(f b6.ExpressionFeature) b6.ExpressionFeature {
	return &modifiedTagsExpression{ExpressionFeature: f, tags: m}
}

func (m ModifiedTags) WrapSegment(segment b6.Segment) b6.Segment {
	return b6.Segment{
		Feature: m.WrapPhysicalFeature(segment.Feature),
		First:   segment.First,
		Last:    segment.Last,
	}
}

func (m ModifiedTags) WrapFeatures(features b6.Features) b6.Features {
	return &modifiedTagsFeatures{features: features, m: m}
}

func (m ModifiedTags) WrapAreas(areas b6.AreaFeatures) b6.AreaFeatures {
	return &modifiedTagsAreas{areas: areas, m: m}
}

func (m ModifiedTags) WrapRelations(relations b6.RelationFeatures) b6.RelationFeatures {
	return &modifiedTagsRelations{relations: relations, m: m}
}

func (m ModifiedTags) WrapCollections(collections b6.CollectionFeatures) b6.CollectionFeatures {
	return &modifiedTagsCollections{collections: collections, m: m}
}

func (m ModifiedTags) WrapSegments(segments b6.Segments) b6.Segments {
	return &modifiedTagsSegments{segments: segments, m: m}
}

func (m ModifiedTags) EachModifiedTag(each func(f ModifiedTag, goroutine int) error, options *b6.EachFeatureOptions) error {
	goroutines := options.Goroutines
	if goroutines < 1 {
		goroutines = 1
	}

	c := make(chan ModifiedTag, goroutines)
	g, gc := errgroup.WithContext(context.Background())
	for i := 0; i < goroutines; i++ {
		g.Go(func() error {
			for tag := range c {
				if err := each(tag, i); err != nil {
					return err
				}
			}
			return nil
		})
	}

done:
	for id, tags := range m {
		for key, value := range tags {
			if !options.IsSkipped(id.Type) {
				select {
				case <-gc.Done():
					break done
				case c <- ModifiedTag{ID: id, Tag: b6.Tag{Key: key, Value: b6.StringExpression(value.value)}, Deleted: value.deleted}:
				}
			}
		}
	}
	close(c)
	return g.Wait()
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

type modifiedTagsAreas struct {
	areas b6.AreaFeatures
	m     ModifiedTags
}

func (m *modifiedTagsAreas) Next() bool {
	return m.areas.Next()
}

func (m *modifiedTagsAreas) Feature() b6.AreaFeature {
	return m.m.WrapAreaFeature(m.areas.Feature())
}

func (m *modifiedTagsAreas) FeatureID() b6.FeatureID {
	return m.areas.FeatureID()
}

type modifiedTagsRelations struct {
	relations b6.RelationFeatures
	m         ModifiedTags
}

func (m *modifiedTagsRelations) Next() bool {
	return m.relations.Next()
}

func (m *modifiedTagsRelations) Feature() b6.RelationFeature {
	return m.m.WrapRelationFeature(m.relations.Feature())
}

func (m *modifiedTagsRelations) FeatureID() b6.FeatureID {
	return m.relations.FeatureID()
}

func (m *modifiedTagsRelations) RelationID() b6.RelationID {
	return m.relations.RelationID()
}

type modifiedTagsCollections struct {
	collections b6.CollectionFeatures
	m           ModifiedTags
}

func (m *modifiedTagsCollections) Next() bool {
	return m.collections.Next()
}

func (m *modifiedTagsCollections) Feature() b6.CollectionFeature {
	return m.m.WrapCollectionFeature(m.collections.Feature())
}

func (m *modifiedTagsCollections) FeatureID() b6.FeatureID {
	return m.collections.FeatureID()
}

func (m *modifiedTagsCollections) CollectionID() b6.CollectionID {
	return m.collections.CollectionID()
}

type modifiedTagsSegments struct {
	segments b6.Segments
	m        ModifiedTags
}

func (m *modifiedTagsSegments) Segment() b6.Segment {
	return m.m.WrapSegment(m.segments.Segment())
}

func (m *modifiedTagsSegments) Next() bool {
	return m.segments.Next()
}

type MutableOverlayWorld struct {
	features   *FeaturesByID
	references *FeatureReferencesByID
	index      *mutableFeatureIndex
	base       b6.World
	tags       ModifiedTags
	epoch      int
}

func NewMutableOverlayWorld(base b6.World) *MutableOverlayWorld {
	w := &MutableOverlayWorld{
		features:   NewFeaturesByID(),
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

func (m *MutableOverlayWorld) FindFeatures(q b6.Query) b6.Features {
	overlay := b6.NewSearchFeatureIterator(q.Compile(m.index, m), m.index)
	return &mutableFeatureIterator{
		i:     newOverlayFeatures(m.tags.WrapFeatures(m.base.FindFeatures(q)), overlay, m.features),
		epoch: m.epoch,
		w:     m,
	}
}

func (m *MutableOverlayWorld) FindFeatureByID(id b6.FeatureID) b6.Feature {
	if feature, ok := (*m.features)[id]; ok {
		return WrapFeature(feature, m)
	}

	return m.tags.WrapFeature(m.base.FindFeatureByID(id))
}

func (m *MutableOverlayWorld) FindLocationByID(id b6.FeatureID) (s2.LatLng, error) {
	if ll, err := m.features.FindLocationByID(id); err == nil {
		return ll, nil
	}
	return m.base.FindLocationByID(id)
}

func (m *MutableOverlayWorld) HasFeatureWithID(id b6.FeatureID) bool {
	return m.features.HasFeatureWithID(id) || m.base.HasFeatureWithID(id)
}

func (m *MutableOverlayWorld) FindRelationsByFeature(id b6.FeatureID) b6.RelationFeatures {
	features := make([]b6.RelationFeature, 0)
	references := m.FindReferences(id, b6.FeatureTypeRelation)
	for references.Next() {
		features = append(features, references.Feature().(b6.RelationFeature))
	}

	return NewRelationFeatureIterator(features)
}

func (m *MutableOverlayWorld) FindCollectionsByFeature(id b6.FeatureID) b6.CollectionFeatures {
	var features []b6.CollectionFeature
	references := m.FindReferences(id, b6.FeatureTypeCollection)
	for references.Next() {
		features = append(features, references.Feature().(b6.CollectionFeature))
	}

	return NewCollectionFeatureIterator(features)
}

func (m *MutableOverlayWorld) FindAreasByPoint(id b6.FeatureID) b6.AreaFeatures {
	var features []b6.AreaFeature
	references := m.FindReferences(id, b6.FeatureTypeArea)
	for references.Next() {
		features = append(features, references.Feature().(b6.AreaFeature))
	}

	return NewAreaFeatureIterator(features)
}

func (m *MutableOverlayWorld) FindReferences(id b6.FeatureID, typed ...b6.FeatureType) b6.Features {
	references := make(map[b6.FeatureID]bool)

	baseReferences := m.base.FindReferences(id) // Not limiting by type in base search.
	for baseReferences.Next() {
		references[baseReferences.FeatureID()] = true
		for _, reference := range m.references.FindReferences(baseReferences.FeatureID(), typed...) {
			references[reference.Source()] = true
		}
	}

	for _, reference := range m.references.FindReferences(id, typed...) {
		references[reference.Source()] = true
	}

	features := make([]b6.Feature, 0, len(references))
	for reference := range references {
		if feature := m.FindFeatureByID(reference); feature != nil {
			if !slices.Contains(features, feature) && (len(typed) == 0 || slices.Contains(typed, feature.FeatureID().Type)) {
				features = append(features, feature)
			}
		}
	}

	return b6.NewFeatureIterator(features)
}

func (m *MutableOverlayWorld) Traverse(id b6.FeatureID) b6.Segments {
	segments := make([]b6.Segment, 0)
	ss := m.base.Traverse(id)
	for ss.Next() {
		s := ss.Segment()
		if feature := m.features.FindFeatureByID(s.Feature.FeatureID()); feature == nil {
			segments = append(segments, s)
		}
	}
	segments = append(segments, traverse(id, m, m.references)...)

	return NewSegmentIterator(segments)
}

func (m *MutableOverlayWorld) EachFeature(each func(f b6.Feature, goroutine int) error, options *b6.EachFeatureOptions) error {
	wrap := func(feature Feature, goroutine int) error {
		return each(WrapFeature(feature, m), goroutine)
	}
	if err := eachIngestFeature(wrap, m.features, m.references, options); err != nil {
		return err
	}
	filter := func(feature b6.Feature, goroutine int) error {
		if !m.features.HasFeatureWithID(feature.FeatureID()) {
			return each(m.tags.WrapFeature(feature), goroutine)
		}
		return nil
	}

	return m.base.EachFeature(filter, options)
}

func (m *MutableOverlayWorld) EachModifiedFeature(each func(f b6.Feature, goroutine int) error, options *b6.EachFeatureOptions) error {
	wrap := func(feature Feature, goroutine int) error {
		return each(WrapFeature(feature, m), goroutine)
	}
	return eachIngestFeature(wrap, m.features, m.references, options)
}

func (m *MutableOverlayWorld) EachModifiedTag(each func(f ModifiedTag, goroutine int) error, options *b6.EachFeatureOptions) error {
	return m.tags.EachModifiedTag(each, options)
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

func (m *MutableOverlayWorld) AddFeature(f Feature) error {
	if err := ValidateFeature(f, &ValidateOptions{InvertClockwisePaths: false}, m); err != nil {
		return err
	}

	existing := (*m.features)[f.FeatureID()]
	references := allReferences(f, m)
	if existing != nil {
		(*m.features)[f.FeatureID()] = f

		for _, reference := range references {
			if err := ValidateFeature(NewFeatureFromWorld(reference), &ValidateOptions{InvertClockwisePaths: false}, m); err != nil {
				return err
			}
		}

		(*m.features)[f.FeatureID()] = existing
	}

	modified := NewModifiedFeaturesWithCopies(f, references, m.features, m)
	modified.Update(m.features, m.references, m.index, m)
	delete(m.tags, f.FeatureID())
	m.epoch++
	return nil
}

func (m *MutableOverlayWorld) AddTag(id b6.FeatureID, tag b6.Tag) error {
	tokenAfter, indexedAfter := b6.TokenForTag(tag)
	if f := m.features.FindMutableFeatureByID(id); f != nil {
		var tokenBefore string
		var indexedBefore bool
		if before := f.Get(tag.Key); before.IsValid() {
			if tokenBefore, indexedBefore = b6.TokenForTag(before); indexedBefore && (!indexedAfter || tokenBefore != tokenAfter) {
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
			m.features.AddFeature(f)
			m.references.AddFeature(f)
			m.index.Add(f, TokensForFeature(WrapFeature(f, m)))
		} else {
			m.tags.ModifyOrAddTag(id, tag)
		}
	}
	return nil
}

func (m *MutableOverlayWorld) RemoveTag(id b6.FeatureID, key string) error {
	if f := m.features.FindMutableFeatureByID(id); f != nil {
		if tag := f.Get(key); tag.IsValid() {
			if token, indexed := b6.TokenForTag(tag); indexed {
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
			if _, indexed := b6.TokenForTag(tag); indexed {
				f = NewFeatureFromWorld(base)
				f.RemoveTag(key)
				m.features.AddFeature(f)
				m.references.AddFeature(f)
				m.index.Add(f, TokensForFeature(WrapFeature(f, m)))
			} else {
				m.tags.RemoveTag(id, key)
			}
		}
	}
	return nil
}

func (m *MutableOverlayWorld) MergeSource(source FeatureSource) error {
	emit := func(f Feature, goroutine int) error {
		m.AddFeature(f)
		return nil
	}
	return source.Read(ReadOptions{}, emit, context.Background())
}

// MergeInto adds the features from this world to other. Merging is not atomic.
// If validation fails (for example, because adding a feature would invalidate
// an existing feature in other), other will be left with only some features
// added.
func (m *MutableOverlayWorld) MergeInto(other MutableWorld) error {
	for _, feature := range *m.features {
		if err := other.AddFeature(feature); err != nil {
			return err
		}
	}
	// TODO: this could (perhaps) be made more efficient if necessary,
	// since the features below have already been validated in the
	// context of this world. Restricting merging to an original 'parent'
	// would, and checking an epoch number for changes, could enable
	// a simple copy.
	return nil
}

func (m *MutableOverlayWorld) Snapshot() b6.World {
	copy := *m
	m.base = &copy
	m.features = NewFeaturesByID()
	m.references = NewFeatureReferences()
	m.tags = NewModifiedTags()
	m.index = newMutableFeatureIndex(m)
	return m.base
}

func allReferences(f Feature, w b6.World) []b6.Feature {
	var features []b6.Feature
	references := w.FindReferences(f.FeatureID())
	for references.Next() {
		features = append(features, references.Feature())
	}

	return features
}

type ModifiedFeatures struct {
	features []Feature
	tokens   [][]string
	copied   []bool
	existing Feature
}

func NewModifiedFeatures(new Feature, features []b6.Feature, byID *FeaturesByID, w b6.World) *ModifiedFeatures {
	m := &ModifiedFeatures{
		features: make([]Feature, 0, len(features)+1),
		tokens:   make([][]string, 0, len(features)+1),
		copied:   make([]bool, 0, len(features)+1),
	}
	m.features = append(m.features, new)
	if m.existing = byID.FindMutableFeatureByID(new.FeatureID()); m.existing != nil {
		m.tokens = append(m.tokens, TokensForFeature(WrapFeature(m.existing, w)))
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

func NewModifiedFeaturesWithCopies(new Feature, features []b6.Feature, byID *FeaturesByID, w b6.World) *ModifiedFeatures {
	m := &ModifiedFeatures{
		features: make([]Feature, 0, len(features)+1),
		tokens:   make([][]string, 0, len(features)+1),
		copied:   make([]bool, 0, len(features)+1),
	}
	m.features = append(m.features, new)
	if m.existing = byID.FindMutableFeatureByID(new.FeatureID()); m.existing != nil {
		m.tokens = append(m.tokens, TokensForFeature(WrapFeature(m.existing, w)))
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

func (m *ModifiedFeatures) RemoveReferences(references *FeatureReferencesByID) {
	if m.existing != nil {
		references.RemoveFeature(m.existing)
	}
	for i, f := range m.features[1:] {
		if !m.copied[i+1] {
			references.RemoveFeature(f)
		}
	}
}

func (m *ModifiedFeatures) Update(features *FeaturesByID, references *FeatureReferencesByID, index *mutableFeatureIndex, byID b6.FeaturesByID) {
	m.RemoveReferences(references)
	var new Feature
	if m.existing != nil {
		switch existing := m.existing.(type) {
		case *AreaFeature:
			existing.MergeFrom(m.features[0].(*AreaFeature))
		case *RelationFeature:
			existing.MergeFrom(m.features[0].(*RelationFeature))
		case *CollectionFeature:
			existing.MergeFrom(m.features[0].(*CollectionFeature))
		case *ExpressionFeature:
			existing.MergeFrom(m.features[0].(*ExpressionFeature))
		default:
			m.existing.MergeFrom(m.features[0])
		}
		new = m.existing
	} else {
		new = m.features[0].Clone()
		features.AddFeature(new)
	}
	m.features[0] = new
	m.AddReferences(references)
	m.UpdateIndex(index, byID)
}

func (m *ModifiedFeatures) AddReferences(references *FeatureReferencesByID) {
	for _, f := range m.features {
		references.AddFeature(f)
	}
}

func (m *ModifiedFeatures) UpdateIndex(index *mutableFeatureIndex, byID b6.FeaturesByID) {
	for i, f := range m.features {
		added, removed := sortAndDiffTokens(m.tokens[i], TokensForFeature(WrapFeature(f, byID)))
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
	tags[tag.Key] = modifiedTag{value: tag.Value.String(), deleted: false}
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

func (m *MutableTagsOverlayWorld) FindLocationByID(id b6.FeatureID) (s2.LatLng, error) {
	return m.base.FindLocationByID(id)
}

func (m *MutableTagsOverlayWorld) FindFeatures(query b6.Query) b6.Features {
	return m.tags.WrapFeatures(m.base.FindFeatures(query))
}

func (m *MutableTagsOverlayWorld) FindRelationsByFeature(id b6.FeatureID) b6.RelationFeatures {
	return m.tags.WrapRelations(m.base.FindRelationsByFeature(id))
}

func (m *MutableTagsOverlayWorld) FindCollectionsByFeature(id b6.FeatureID) b6.CollectionFeatures {
	return m.tags.WrapCollections(m.base.FindCollectionsByFeature(id))
}

func (m *MutableTagsOverlayWorld) FindAreasByPoint(id b6.FeatureID) b6.AreaFeatures {
	return m.tags.WrapAreas(m.base.FindAreasByPoint(id))
}

func (m *MutableTagsOverlayWorld) FindReferences(id b6.FeatureID, typed ...b6.FeatureType) b6.Features {
	return m.tags.WrapFeatures(m.base.FindReferences(id, typed...))
}

func (m *MutableTagsOverlayWorld) Traverse(id b6.FeatureID) b6.Segments {
	return m.tags.WrapSegments(m.base.Traverse(id))
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
