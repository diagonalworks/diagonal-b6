import { ImmerStateCreator } from '@/lib/zustand';
import { ScaleOrdinal } from 'd3-scale';
import { GeoJsonObject } from 'geojson';
import { MapGeoJSONFeature } from 'maplibre-gl';
import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';
import { World } from './worlds';

type HistogramLayerSpec = {
    tiles: string;
    show: boolean;
    selected: string | undefined;
    colorScale?: ScaleOrdinal<string, string, never>;
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
        histogram: Record<
            string,
            {
                spec: HistogramLayerSpec;
                world: World['id'];
            }
        >;
        highlight: Record<
            string,
            {
                features: FeatureHighlight[];
                world: World['id'];
            }
        >;
    };
    actions: {
        setGeoJsonLayer: (
            id: string,
            geojson: {
                features: GeoJsonObject[];
                world: World['id'];
            }
        ) => void;
        removeGeoJsonLayer: (id: string) => void;
        setHistogramLayer: (
            id: string,
            histogram: {
                spec: HistogramLayerSpec;
                world: World['id'];
            }
        ) => void;
        setHistogramBucket: (id: string, bucket: string | undefined) => void;
        setHistogramScale: (
            id: string,
            scale: ScaleOrdinal<string, string, never>
        ) => void;
        removeHistogramLayer: (id: string) => void;
    };
}

export const createMapStore: ImmerStateCreator<MapStore, MapStore> = (set) => ({
    layers: {
        geojson: {},
        histogram: {},
        highlight: {},
    },
    actions: {
        setGeoJsonLayer: (id, geojson) => {
            set((state) => {
                state.layers.geojson[id] = geojson;
            });
        },
        setHistogramLayer: (id, spec) => {
            set((state) => {
                state.layers.histogram[id] = spec;
            });
        },
        setHistogramBucket: (id, bucket) => {
            set((state) => {
                if (state.layers.histogram[id]) {
                    state.layers.histogram[id].spec.selected = bucket;
                }
            });
        },
        setHistogramScale: (id, scale) => {
            set((state) => {
                if (state.layers.histogram[id]) {
                    state.layers.histogram[id].spec.colorScale = scale;
                }
            });
        },
        removeGeoJsonLayer: (id) => {
            set((state) => {
                delete state.layers.geojson[id];
            });
        },
        removeHistogramLayer: (id) => {
            set((state) => {
                delete state.layers.histogram[id];
            });
        },
    },
});

// @ts-expect-error - type instanciation, @TODO: fix
export const useMapStore = create(immer(createMapStore));
