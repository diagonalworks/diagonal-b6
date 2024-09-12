import { ScaleOrdinal } from 'd3-scale';
import { GeoJsonObject } from 'geojson';
import { MapGeoJSONFeature } from 'maplibre-gl';
import { match } from 'ts-pattern';
import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';

import { ImmerStateCreator } from '@/lib/zustand';
import { World } from '@/stores/worlds';

import { OutlinerSpec } from './outliners';

export type HistogramLayerSpec = {
    tiles: string;
    selected: string | undefined;
    colorScale?: ScaleOrdinal<string, string, never>;
    showOnMap?: boolean;
};

export type Layer = {
    world: World['id'];
    outliner: OutlinerSpec['id'];
    type: 'histogram' | 'collection';
};

export type HistogramLayer = Layer & {
    spec: HistogramLayerSpec;
    type: 'histogram';
};

export type CollectionLayerSpec = {
    tiles: string;
    showOnMap?: boolean;
};

export type CollectionLayer = Layer & {
    spec: CollectionLayerSpec;
    type: 'collection';
};

export type FeatureHighlight = {
    feature: MapGeoJSONFeature;
    layer: string;
};

export interface MapStore {
    layers: {
        geojson: Record<
            string,
            {
                features: GeoJsonObject[];
                world: World['id'];
            }
        >;
        tiles: Record<string, CollectionLayer | HistogramLayer>;
        highlight: Record<
            string,
            {
                features: FeatureHighlight[];
                world: World['id'];
            }
        >;
    };
    actions: {
        /**
         * Set a GeoJSON layer on the map
         * @param id - The unique identifier for the layer
         * @param geojson - An object containing the GeoJSON features and the world the layer is in
         * @returns void
         */
        setGeoJsonLayer: (
            id: string,
            geojson: {
                features: GeoJsonObject[];
                world: World['id'];
            }
        ) => void;
        /**
         * Remove a GeoJSON layer from the map
         * @param id - The unique identifier for the layer
         * @returns void
         */
        removeGeoJsonLayer: (id: string) => void;
        /**
         * Set the selected bucket for a histogram layer
         * @param id - The unique identifier for the layer
         * @param bucket - The selected bucket
         */
        setHistogramBucket: (id: string, bucket: string | undefined) => void;
        /**
         * Set the color scale for a histogram layer
         * @param id - The unique identifier for the layer
         * @param scale - The color scale
         * @returns void
         */
        setHistogramScale: (
            id: string,
            scale: ScaleOrdinal<string, string, never>
        ) => void;
        /**
         * Set a highlight layer on the map
         * @param id - The unique identifier for the layer
         * @param features - An array of features to highlight
         * @returns void
         */
        setHighlightLayer: (
            id: string,
            highlight: {
                features: FeatureHighlight[];
                world: World['id'];
            }
        ) => void;
        /**
         * Remove a highlight layer from the map
         * @param id - The unique identifier for the layer
         * @returns void
         */
        removeHighlightLayer: (id: string) => void;
        /**
         * Set a tile layer on the map
         * @param id - The unique identifier for the layer
         * @param layer - An object containing the tile layer spec and the world the layer is in
         * @returns void
         */
        setTileLayer: (
            id: string,
            layer: CollectionLayer | HistogramLayer
        ) => void;
        /**
         * Remove a tile layer from the map
         * @param id - The unique identifier for the layer
         * @returns void
         */
        removeTileLayer: (id: string) => void;
        /**
         * Remove all tile layers of a specific outliner
         * @param outliner - The outliner ID
         * @returns void
         */
        removeOutlinerLayers: (outliner: string) => void;
        /**
         * Hide all tile layers of a specific outliner
         * @param outliner - The outliner ID
         * @returns void
         */
        hideOutlinerLayers: (outliner: string) => void;
        /**
         * Show all tile layers of a specific outliner
         * @param outliner - The outliner ID
         * @returns void
         */
        showOutlinerLayers: (outliner: string) => void;
    };
}

export const createMapStore: ImmerStateCreator<MapStore, MapStore> = (set) => ({
    layers: {
        geojson: {},
        highlight: {},
        tiles: {},
    },
    actions: {
        setGeoJsonLayer: (id, geojson) => {
            set((state) => {
                state.layers.geojson[id] = geojson;
            });
        },
        setHistogramBucket: (id, bucket) => {
            set((state) => {
                match(state.layers.tiles[id])
                    .with({ type: 'histogram' }, (l) => {
                        l.spec.selected = bucket;
                    })
                    .otherwise(() => {});
            });
        },
        setHistogramScale: (id, scale) => {
            set((state) => {
                match(state.layers.tiles[id])
                    .with({ type: 'histogram' }, (l) => {
                        l.spec.colorScale = scale;
                    })
                    .otherwise(() => {});
            });
        },
        removeGeoJsonLayer: (id) => {
            set((state) => {
                delete state.layers.geojson[id];
            });
        },
        setHighlightLayer: (id, highlight) => {
            set((state) => {
                state.layers.highlight[id] = highlight;
            });
        },
        removeHighlightLayer: (id) => {
            set((state) => {
                delete state.layers.highlight[id];
            });
        },
        setTileLayer: (id, layer) => {
            set((state) => {
                state.layers.tiles[id] = layer;
            });
        },
        removeTileLayer: (id) => {
            set((state) => {
                delete state.layers.tiles[id];
            });
        },
        removeOutlinerLayers: (outliner) => {
            set((state) => {
                for (const [key, value] of Object.entries(state.layers.tiles)) {
                    if (value.outliner === outliner) {
                        delete state.layers.tiles[key];
                    }
                }
            });
        },
        hideOutlinerLayers: (outliner) => {
            set((state) => {
                for (const [key, value] of Object.entries(state.layers.tiles)) {
                    if (value.outliner === outliner) {
                        state.layers.tiles[key].spec.showOnMap = false;
                    }
                }
            });
        },
        showOutlinerLayers: (outliner) => {
            set((state) => {
                for (const [key, value] of Object.entries(state.layers.tiles)) {
                    if (value.outliner === outliner) {
                        state.layers.tiles[key].spec.showOnMap = true;
                    }
                }
            });
        },
    },
});

/**
 * A hook to access the map store, which contains information about the map layers. This store is used to manage the display of GeoJSON and histogram layers on the map.
 * This is a zustand store that uses immer for immutability.
 * @returns The map store
 */
// @ts-expect-error - type instanciation, @TODO: fix
export const useMapStore = create(immer(createMapStore));
