import * as circleIcons from '@/assets/icons/circle';
import { viewAtom } from '@/atoms/location';
import { colorToRgbArray } from '@/lib/colors';
import { getFeaturePath } from '@/lib/map';
import { MVTLayer } from '@deck.gl/geo-layers/typed';
import {
    MapboxOverlay as DeckOverlay,
    MapboxOverlayProps,
} from '@deck.gl/mapbox';
import { DotIcon, MinusIcon, PlusIcon } from '@radix-ui/react-icons';
import { color } from 'd3-color';
import { useAtom } from 'jotai';
import { Feature, MapLayerMouseEvent } from 'maplibre-gl';
import { PropsWithChildren, useCallback, useMemo, useState } from 'react';
import {
    Map as MapLibre,
    Marker,
    useControl,
    useMap,
} from 'react-map-gl/maplibre';
import { match } from 'ts-pattern';
import { MapControls } from './system/MapControls';

// https://github.com/visgl/react-map-gl/discussions/2216#discussioncomment-7537888
import { b6Path } from '@/lib/b6';
import { useScenarioContext } from '@/lib/context/scenario';
import { GeoJsonLayer } from 'deck.gl/typed';
import 'maplibre-gl/dist/maplibre-gl.css';
import { twMerge } from 'tailwind-merge';

export function DeckGLOverlay(props: MapboxOverlayProps) {
    const overlay = useControl(() => new DeckOverlay(props));
    overlay.setProps(props);
    return null;
}

export const ScenarioMap = ({ children }: PropsWithChildren) => {
    const { createOutlinerInScenario } = useScenarioContext();
    const {
        getVisibleMarkers,
        queryLayers,
        geoJSON,
        scenario: { id, featureId, worldCreated },
        mapStyle,
        tab,
    } = useScenarioContext();
    const { [id]: map } = useMap();
    const [viewState, setViewState] = useAtom(viewAtom);
    const [cursor, setCursor] = useState<'auto' | 'pointer'>('auto');

    const isBaseline = id === 'baseline';
    const geoJsonLayerGL = useMemo(() => {
        if (!map) return null;
        return new GeoJsonLayer({
            data: geoJSON,
            id: 'geojson',
            getFillColor: colorToRgbArray(isBaseline ? '#b1c5fd' : '#E2B79F'),
            getLineWidth: 1,
            getLineColor: colorToRgbArray(isBaseline ? '#37589f' : '#A66B4D'),
        });
    }, [geoJSON, map]);

    const queryLayersGL = useMemo(() => {
        if (!map) return null;
        return queryLayers.map((ql) => {
            const histogram = ql.histogram;

            if (!histogram || !ql.show) return null;
            return new MVTLayer({
                data: [
                    `${b6Path}tiles/${ql.layer.path}/{z}/{x}/{y}.mvt?q=${
                        ql.layer.q
                    }${
                        featureId && featureId?.value
                            ? `&r=collection/${featureId.namespace}/${featureId.value}`
                            : ''
                    }`,
                ],
                beforeId: 'road-label',
                id: `${ql.layer.path}+${ql.layer.q}`,
                getFillColor: (f: Feature) => {
                    if (f.properties?.layerName === 'background') {
                        return [0, 0, 0, 0];
                    }
                    if (f.properties?.layerName === ql.layer.path) {
                        const c = histogram.colorScale?.(f.properties.bucket);
                        if (!c) {
                            return [0, 0, 0, 0];
                        }

                        const isSelected =
                            histogram?.selected &&
                            histogram.selected.toString() ===
                                f.properties.bucket;

                        return colorToRgbArray(
                            c,
                            histogram?.selected ? (isSelected ? 255 : 155) : 255
                        );
                    }
                    return [0, 0, 0, 0];
                },
                getLineWidth: (f: Feature) => {
                    if (f.properties?.layerName === ql.layer.path) {
                        /**
                         * skipping road highlighing for performance reasons
                         */
                        /* const queryFeatures = map.querySourceFeatures(
                            'diagonal',
                            {
                                sourceLayer: 'road',
                                filter: ['all', ['==', 'id', f.properties.id]],
                            }
                        );

                        const feature = queryFeatures?.[0];

                        if (feature) {
                            return (
                                getRoadWidth(feature.properties?.highway) * 1.5
                            );
                        }
 */
                        const isSelected =
                            histogram?.selected &&
                            histogram.selected.toString() ===
                                f.properties.bucket;

                        return histogram?.selected
                            ? isSelected
                                ? 0.8
                                : 0.2
                            : 0.5;
                    }
                    return 0;
                },
                getLineColor: (f: Feature) => {
                    if (f.properties?.layerName === ql.layer.path) {
                        const c = histogram?.colorScale?.(f.properties?.bucket);
                        if (!c) {
                            return [0, 0, 0, 0];
                        }
                        const isSelected =
                            histogram.selected &&
                            histogram.selected.toString() ===
                                f.properties.bucket;

                        const darken = color(c)
                            ?.darker(isSelected ? 2 : 0.5)
                            .formatRgb();

                        return colorToRgbArray(
                            darken ?? c,
                            histogram?.selected ? (isSelected ? 255 : 155) : 255
                        );
                    }
                    return [0, 0, 0, 0];
                },
                updateTriggers: {
                    getLineColor: [histogram.colorScale, histogram.selected],
                    getFillColor: [histogram.colorScale, histogram.selected],
                    getLineWidth: [histogram.selected],
                },
            });
        });
    }, [queryLayers]);

    const Markers = useMemo(() => {
        if (!map) return null;
        const markers = getVisibleMarkers(map);
        return markers.map((marker, i) => {
            if (marker.geometry.type !== 'Point') return null;

            const icon = match(marker.properties?.['-b6-icon'])
                .with('dot', () => {
                    return (
                        <DotIcon
                            className={twMerge(
                                'fill-graphite-80',
                                tab === 'right' && 'fill-rose-80'
                            )}
                        />
                    );
                })
                .otherwise(() => {
                    const icon = marker.properties?.['-b6-icon'];
                    if (!icon)
                        return (
                            <div
                                className={twMerge(
                                    'w-2 h-2 rounded-full bg-ultramarine-50 border border-ultramarine-80',
                                    tab === 'right' &&
                                        'bg-rose-50 border-rose-80'
                                )}
                            />
                        );
                    const iconComponentName = `${icon
                        .charAt(0)
                        .toUpperCase()}${icon.slice(1)}`;
                    if (
                        circleIcons[
                            iconComponentName as keyof typeof circleIcons
                        ]
                    ) {
                        const Icon =
                            circleIcons[
                                iconComponentName as keyof typeof circleIcons
                            ];
                        return <Icon />;
                    }
                    return <DotIcon />;
                });
            return (
                <Marker
                    key={i}
                    longitude={marker.geometry.coordinates[0]}
                    latitude={marker.geometry.coordinates[1]}
                    className={twMerge(
                        '[&>svg]:fill-graphite-80',
                        tab === 'right' && '[&>svg]:fill-rose-80'
                    )}
                >
                    {icon}
                </Marker>
            );
        });
    }, [map, getVisibleMarkers]);

    const handleMapClick = useCallback(
        (e: MapLayerMouseEvent) => {
            const outlinerProperties = {
                scenario: id,
                docked: false,
                transient: true,
                coordinates: e.point,
            };

            const evaluateLatLon = ({ locked }: { locked: boolean }) => {
                const expression = `${e.lngLat.lat.toFixed(
                    6
                )}, ${e.lngLat.lng.toFixed(6)}`;

                createOutlinerInScenario({
                    id: `stack_mlc_${expression}`,
                    properties: outlinerProperties,
                    request: {
                        eventType: 'mlc',
                        locked,
                        expression,
                    },
                });
            };

            // if shift key is pressed, create a new unlocked outliner.
            if (e.originalEvent.shiftKey) {
                evaluateLatLon({ locked: false });
            } else {
                const features = map?.queryRenderedFeatures(e.point);
                const feature = features?.[0];
                if (feature) {
                    const path = getFeaturePath(feature);
                    const expression = `find-feature ${path}`;
                    createOutlinerInScenario({
                        id: `stack_mfc_${expression}`,
                        properties: outlinerProperties,
                        request: {
                            eventType: 'mfc',
                            locked: false,
                            expression,
                        },
                    });
                } else {
                    evaluateLatLon({ locked: true });
                }
            }
        },
        [map, createOutlinerInScenario, id]
    );

    return (
        <MapLibre
            key={`${id}-${worldCreated ? '-world' : ''}`}
            id={id}
            {...viewState}
            onMove={(evt) => {
                setViewState(evt.viewState);
            }}
            onMouseEnter={() => {
                setCursor('pointer');
            }}
            onMouseLeave={() => {
                setCursor('auto');
            }}
            onClick={handleMapClick}
            cursor={cursor}
            attributionControl={false}
            interactive={true}
            interactiveLayerIds={['building', 'road']}
            dragRotate={false}
            mapStyle={mapStyle}
            boxZoom={false} // https://github.com/mapbox/mapbox-gl-js/issues/6971s
        >
            <DeckGLOverlay
                layers={[queryLayersGL, geoJsonLayerGL]}
                interleaved
            />

            <MapControls
                className={twMerge(tab === 'right' && 'right-0 left-auto')}
            >
                <MapControls.Button
                    onClick={() => map?.zoomIn({ duration: 200 })}
                >
                    <PlusIcon />
                </MapControls.Button>
                <MapControls.Button
                    onClick={() => map?.zoomOut({ duration: 200 })}
                >
                    <MinusIcon />
                </MapControls.Button>
            </MapControls>

            {Markers}
            {children}
        </MapLibre>
    );
};
