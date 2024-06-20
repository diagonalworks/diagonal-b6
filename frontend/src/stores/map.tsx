import { ScaleOrdinal } from 'd3-scale';
import { GeoJsonObject } from 'geojson';
import { MapGeoJSONFeature } from 'maplibre-gl';
import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';

import { ImmerStateCreator } from '@/lib/zustand';
import { World } from '@/stores/worlds';

type HistogramLayerSpec = {
    tiles: string;
    selected: string | undefined;
    colorScale?: ScaleOrdinal<string, string, never>;
    showOnMap?: boolean;
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
         * Set a histogram layer on the map
         * @param id - The unique identifier for the layer
         * @param histogram - An object containing the histogram spec and the world the layer is in
         * @returns void
         */
        setHistogramLayer: (
            id: string,
            histogram: {
                spec: HistogramLayerSpec;
                world: World['id'];
            }
        ) => void;
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
         * Remove a histogram layer from the map
         * @param id - The unique identifier for the layer
         * @returns void
         */
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

/**
 * A hook to access the map store, which contains information about the map layers. This store is used to manage the display of GeoJSON and histogram layers on the map.
 * This is a zustand store that uses immer for immutability.
 * @returns The map store
 */
// @ts-expect-error - type instanciation, @TODO: fix
export const useMapStore = create(immer(createMapStore));
