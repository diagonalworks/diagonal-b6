package ingest

import (
	"diagonal.works/b6"
	"github.com/golang/geo/s2"
)

type OverlayWorld struct {
	overlay b6.World
	base    b6.World
}

func NewOverlayWorld(overlay b6.World, base b6.World) b6.World {
	return &OverlayWorld{overlay: overlay, base: base}
}

type overlayFeatures struct {
	base      b6.Features
	overlay   b6.Features
	filter    b6.FeaturesByID
	baseID    b6.FeatureID // Memoise IDs, since there are tight loops involving them.
	overlayID b6.FeatureID
	baseOK    bool
	overlayOK bool
	started   bool
}

func newOverlayFeatures(base b6.Features, overlay b6.Features, filter b6.FeaturesByID) *overlayFeatures {
	return &overlayFeatures{
		base:      base,
		overlay:   overlay,
		filter:    filter,
		baseOK:    true,
		overlayOK: true,
		started:   false,
	}
}

func (o *overlayFeatures) Next() bool {
	if !o.started {
		if o.overlayOK = o.overlay.Next(); o.overlayOK {
			o.overlayID = o.overlay.FeatureID()
		}
		o.advanceBase()
		o.started = true
	} else if o.overlayOK {
		if o.baseOK {
			if o.overlayID.Less(o.baseID) {
				if o.overlayOK = o.overlay.Next(); o.overlayOK {
					o.overlayID = o.overlay.FeatureID()
				}
			} else {
				if o.overlayID == o.baseID {
					if o.overlayOK = o.overlay.Next(); o.overlayOK {
						o.overlayID = o.overlay.FeatureID()
					}
				}
				o.advanceBase()
			}
		} else {
			if o.overlayOK = o.overlay.Next(); o.overlayOK {
				o.overlayID = o.overlay.FeatureID()
			}
		}
	} else {
		o.advanceBase()
	}
	return o.overlayOK || o.baseOK
}

func (o *overlayFeatures) advanceBase() {
	if o.baseOK {
		for {
			if o.baseOK = o.base.Next(); o.baseOK {
				o.baseID = o.base.FeatureID()
			}
			if !o.baseOK || !o.filter.HasFeatureWithID(o.baseID) {
				break
			}
		}
	}
}

func (o *overlayFeatures) FeatureID() b6.FeatureID {
	if o.overlayOK {
		if o.baseOK {
			if o.overlayID.Less(o.baseID) {
				return o.overlayID
			}
			return o.baseID
		}
		return o.overlayID
	}
	return o.baseID
}

func (o *overlayFeatures) Feature() b6.Feature {
	if o.overlayOK {
		if o.baseOK {
			if o.overlayID.Less(o.baseID) {
				return o.overlay.Feature()
			}
			return o.base.Feature()
		}
		return o.overlay.Feature()
	}
	return o.base.Feature()
}

func (o *OverlayWorld) FindFeatures(q b6.Query) b6.Features {
	return newOverlayFeatures(o.base.FindFeatures(q), o.overlay.FindFeatures(q), o.overlay)
}

func (o *OverlayWorld) FindFeatureByID(id b6.FeatureID) b6.Feature {
	if feature := o.overlay.FindFeatureByID(id); feature != nil {
		return feature
	}
	return o.base.FindFeatureByID(id)
}

func (o *OverlayWorld) HasFeatureWithID(id b6.FeatureID) bool {
	return o.overlay.HasFeatureWithID(id) || o.base.HasFeatureWithID(id)
}

func (o *OverlayWorld) FindLocationByID(id b6.FeatureID) (s2.LatLng, error) {
	if ll, err := o.overlay.FindLocationByID(id); err == nil {
		return ll, nil
	}
	return o.base.FindLocationByID(id)
}

func (o *OverlayWorld) FindRelationsByFeature(id b6.FeatureID) b6.RelationFeatures {
	byID := make(map[b6.RelationID]b6.RelationFeature)
	for _, w := range []b6.World{o.base, o.overlay} {
		r := w.FindRelationsByFeature(id)
		for r.Next() {
			byID[r.Feature().RelationID()] = r.Feature()
		}
	}
	relations := make([]b6.RelationFeature, 0, len(byID))
	for _, relation := range byID {
		relations = append(relations, relation)
	}
	return &relationFeatures{relations: relations, i: -1}
}

func (o *OverlayWorld) FindCollectionsByFeature(id b6.FeatureID) b6.CollectionFeatures {
	byID := make(map[b6.CollectionID]b6.CollectionFeature)
	for _, w := range []b6.World{o.base, o.overlay} {
		c := w.FindCollectionsByFeature(id)
		for c.Next() {
			byID[c.Feature().CollectionID()] = c.Feature()
		}
	}
	collections := make([]b6.CollectionFeature, 0, len(byID))
	for _, collection := range byID {
		collections = append(collections, collection)
	}
	return &collectionFeatures{collections: collections, i: -1}
}

func (o *OverlayWorld) FindAreasByPoint(p b6.FeatureID) b6.AreaFeatures {
	byID := make(map[b6.AreaID]b6.AreaFeature)
	for _, w := range []b6.World{o.base, o.overlay} {
		areas := w.FindAreasByPoint(p)
		for areas.Next() {
			byID[areas.Feature().AreaID()] = areas.Feature()
		}

	}
	features := make([]b6.AreaFeature, 0, len(byID))
	for _, area := range byID {
		features = append(features, area)
	}
	return &areaFeatures{features: features, i: -1}
}

func (o *OverlayWorld) FindReferences(id b6.FeatureID, typed ...b6.FeatureType) b6.Features {
	byID := make(map[b6.FeatureID]b6.Feature)
	for _, w := range []b6.World{o.base, o.overlay} {
		references := w.FindReferences(id, typed...)
		for references.Next() {
			byID[references.FeatureID()] = references.Feature()
		}

	}

	features := make([]b6.Feature, 0, len(byID))
	for _, feature := range byID {
		features = append(features, feature)
	}

	return b6.NewFeatureIterator(features)
}

func (o *OverlayWorld) Traverse(id b6.FeatureID) b6.Segments {
	byKey := make(map[b6.SegmentKey]b6.Segment)
	for _, w := range []b6.World{o.base, o.overlay} {
		segments := w.Traverse(id)
		for segments.Next() {
			s := segments.Segment()
			byKey[s.ToKey()] = s
		}
	}
	segments := make([]b6.Segment, 0, len(byKey))
	for _, s := range byKey {
		segments = append(segments, s)
	}
	return NewSegmentIterator(segments)
}

func (o *OverlayWorld) EachFeature(each func(f b6.Feature, goroutine int) error, options *b6.EachFeatureOptions) error {
	if err := o.overlay.EachFeature(each, options); err != nil {
		return err
	}
	filter := func(feature b6.Feature, goroutine int) error {
		if !o.overlay.HasFeatureWithID(feature.FeatureID()) {
			return each(feature, goroutine)
		}
		return nil
	}
	return o.base.EachFeature(filter, options)
}

func (o *OverlayWorld) Tokens() []string {
	tokens := make(map[string]struct{})
	for _, token := range o.base.Tokens() {
		tokens[token] = struct{}{}
	}
	for _, token := range o.overlay.Tokens() {
		tokens[token] = struct{}{}
	}
	all := make([]string, 0, len(tokens))
	for token, _ := range tokens {
		all = append(all, token)
	}
	return all
}
