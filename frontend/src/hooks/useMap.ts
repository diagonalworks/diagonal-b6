import {
    FilterSpecification,
    MapGeoJSONFeature,
    MapLayerMouseEvent,
} from 'maplibre-gl';
import { useCallback, useMemo } from 'react';
import { useMap as useMapLibre } from 'react-map-gl/maplibre';

import { OutlinerSpec, useOutlinersStore } from '@/stores/outliners';
import { World, useWorldStore } from '@/stores/worlds';
import { Event } from '@/types/events';
import { getFeaturePath } from '@/utils/map';

/**
 * Hook for interacting with the rendered map. This hook provides functions for evaluating
 * expressions and features on the map.
 * @param id - The id of the world the map is in
 * @returns The actions for interacting with the map
 */
export const useMap = ({ id }: { id: World['id'] }) => {
    const { [id]: maplibre } = useMapLibre();
    const outlinerActions = useOutlinersStore((state) => state.actions);
    const world = useWorldStore((state) => state.worlds[id]);
    const baseline = useWorldStore((state) => state.worlds.baseline);

    const baseRequest: () => Partial<OutlinerSpec['request']> =
        useCallback(() => {
            const mapCenter = maplibre?.getCenter();
            const mapZoom = maplibre?.getZoom();
            return {
                root: world.featureId ?? baseline.featureId,
                ...(mapCenter && {
                    logMapCenter: {
                        latE7: Math.round(mapCenter.lat * 1e7),
                        lngE7: Math.round(mapCenter.lng * 1e7),
                    },
                }),
                ...(mapZoom && { logMapZoom: mapZoom }),
            };
        }, [maplibre, world, baseline]);

    const evaluateLatLng = useCallback(
        ({ e, locked }: { e: MapLayerMouseEvent; locked: boolean }) => {
            const expression = `${e.lngLat.lat.toFixed(
                6
            )}, ${e.lngLat.lng.toFixed(6)}`;
            const event: Event = 'mlc';

            outlinerActions.add({
                id: `${id}-${event}-${expression}`,
                world: id,
                properties: {
                    active: false,
                    docked: false,
                    transient: true,
                    coordinates: e.point,
                    type: 'core',
                    show: true,
                },
                request: {
                    ...baseRequest(),
                    expression,
                    locked,
                    logEvent: event,
                },
            });
        },
        [outlinerActions, baseRequest, id]
    );

    const evaluateFeature = useCallback(
        ({
            e,
            locked,
            feature,
        }: {
            e: MapLayerMouseEvent;
            locked: boolean;
            feature: MapGeoJSONFeature;
        }) => {
            const path = getFeaturePath(feature);
            const expression = `find-feature ${path}`;
            const event: Event = 'mfc';

            outlinerActions.add({
                id: `${id}-${event}-${expression}`,
                world: id,
                properties: {
                    active: false,
                    docked: false,
                    transient: true,
                    coordinates: e.point,
                    type: 'core',
                    show: true,
                },
                request: {
                    ...baseRequest(),
                    expression,
                    logEvent: event,
                    locked,
                },
            });
        },
        [outlinerActions, id, baseRequest]
    );

    const evaluateExpression = useCallback(
        (expression: string) => {
            outlinerActions.add({
                id: `${id}-${expression}`,
                world: id,
                properties: {
                    active: false,
                    docked: false,
                    transient: true,
                    coordinates: { x: 8, y: 60 },
                    type: 'core',
                    show: true,
                },
                request: {
                    ...baseRequest(),
                    expression,
                    locked: false,
                    logEvent: 'ws',
                },
            });
        },
        [outlinerActions, id, baseRequest]
    );

    const findFeatureInLayer = useCallback(
        ({
            layer,
            filter,
            id,
        }: {
            layer: string;
            filter: FilterSpecification;
            id: number;
        }) => {
            if (!maplibre) return;
            const queryFeatures = maplibre.querySourceFeatures('diagonal', {
                sourceLayer: layer,
                filter,
            });
            const feature = queryFeatures?.find(
                (f) => parseInt(f.properties.id, 16) == id
            );
            return feature ? [{ feature, layer }] : undefined;
        },
        [maplibre]
    );

    const highlightFeature = useCallback(
        ({
            feature,
            layer,
            highlight,
        }: {
            feature: MapGeoJSONFeature;
            layer: string;
            highlight: boolean;
        }) => {
            if (!maplibre) return;
            maplibre.setFeatureState(
                {
                    source: 'diagonal',
                    sourceLayer: layer,
                    id: feature.id,
                },
                {
                    highlighted: highlight,
                }
            );
        },
        [maplibre]
    );

    const actions = useMemo(
        () => ({
            evaluateLatLng,
            evaluateFeature,
            findFeatureInLayer,
            highlightFeature,
            evaluateExpression,
        }),
        [
            highlightFeature,
            evaluateLatLng,
            evaluateFeature,
            findFeatureInLayer,
            evaluateExpression,
        ]
    );

    return [actions];
};
