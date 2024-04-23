import * as circleIcons from '@/assets/icons/circle';
import { AppStore, appAtom } from '@/atoms/app';
import { viewAtom } from '@/atoms/location';
import { MapControls } from '@/components/system/MapControls';
import { fetchB6 } from '@/lib/b6';
import { ChartDimensions, useChartDimensions } from '@/lib/useChartDimensions';
import { StackResponse } from '@/types/stack';
import {
    MapboxOverlay as DeckOverlay,
    MapboxOverlayProps,
} from '@deck.gl/mapbox';
import {
    DndContext,
    KeyboardSensor,
    MouseSensor,
    PointerSensor,
    TouchSensor,
    UniqueIdentifier,
    useDraggable,
    useDroppable,
    useSensor,
    useSensors,
} from '@dnd-kit/core';
import { restrictToWindowEdges } from '@dnd-kit/modifiers';
import { DotIcon, MinusIcon, PlusIcon } from '@radix-ui/react-icons';
import { useQuery } from '@tanstack/react-query';
import { MVTLayer } from 'deck.gl/typed';
import { AnimatePresence, motion } from 'framer-motion';
import { useAtom } from 'jotai';
import { debounce, pickBy } from 'lodash';
import {
    Feature,
    MapLayerMouseEvent,
    Point,
    StyleSpecification,
} from 'maplibre-gl';
import 'maplibre-gl/dist/maplibre-gl.css';
import {
    HTMLAttributes,
    PropsWithChildren,
    useCallback,
    useEffect,
    useMemo,
    useState,
} from 'react';
import {
    Map as MapLibre,
    Marker,
    Source,
    ViewState,
    useControl,
    useMap,
} from 'react-map-gl/maplibre';
import { twMerge } from 'tailwind-merge';
import { match } from 'ts-pattern';
import { StackAdapter } from './adapters/StackAdapter';
import diagonalBasemapStyle from './diagonal-map-style.json';

export function DeckGLOverlay(props: MapboxOverlayProps) {
    const overlay = useControl(() => new DeckOverlay(props));
    overlay.setProps(props);
    return null;
}

export function Map({
    id,
    ...props
}: { id: string } & HTMLAttributes<HTMLDivElement>) {
    const [ref, dimensions] = useChartDimensions({
        marginTop: 0,
        marginRight: 0,
        marginBottom: 0,
        marginLeft: 0,
    });
    const { [id]: map } = useMap();
    const [cursor, setCursor] = useState<'auto' | 'pointer'>('auto');
    const [activeStackId, setActiveStackId] = useState<UniqueIdentifier | null>(
        null
    );

    const [{ stacks, startup, geojson }, setAppAtom] = useAtom(appAtom);

    const pointerSensor = useSensor(PointerSensor, {
        activationConstraint: {
            distance: 5,
        },
    });
    const keyboardSensor = useSensor(KeyboardSensor);
    const mouseSensor = useSensor(MouseSensor);
    const touchSensor = useSensor(TouchSensor);
    const sensors = useSensors(
        pointerSensor,
        keyboardSensor,
        mouseSensor,
        touchSensor
    );

    const [viewState, setViewState] = useAtom(viewAtom);
    const [mapViewState, setMapViewState] = useState<ViewState>(viewState);

    // Debounce the view state update to avoid updating the URL too often
    // eslint-disable-next-line react-hooks/exhaustive-deps
    const debouncedSetViewState = useCallback(debounce(setViewState, 1000), [
        setViewState,
    ]);

    const [expression, setExpression] = useState<string | null>(null);
    const [coordinates, setCoordinates] = useState<Point>();

    const stackQuery = useQuery({
        queryKey: ['stack', expression],
        queryFn: () => {
            if (
                !expression ||
                !startup?.session ||
                !map?.getCenter() ||
                map?.getZoom() === undefined
            ) {
                return null;
            }

            return fetchB6('stack', {
                expression: expression,
                node: undefined,
                root: undefined,
                locked: true,
                session: startup.session,
                logMapCenter: {
                    latE7: Math.round(map.getCenter().lat * 1e7),
                    lngE7: Math.round(map.getCenter().lng * 1e7),
                },
                logEvent: 'mlc',
                logMapZoom: map.getZoom(),
                context: startup.context,
            }).then((res) => res.json() as Promise<StackResponse>);
        },
        enabled:
            !!expression &&
            !!startup?.session &&
            map?.getCenter() &&
            map?.getZoom() !== undefined,
    });

    useEffect(() => {
        if (!stackQuery.data) return;
        const { proto, geoJSON } = stackQuery.data;
        if (proto && coordinates && expression) {
            setAppAtom((draft) => {
                draft.stacks = pickBy(
                    stacks,
                    (stack) => !stack.transient || stack.docked
                );
                draft.stacks[proto.expression] = {
                    coordinates,
                    docked: false,
                    expression: expression,
                    proto,
                    id: proto.expression,
                    transient: true,
                };
                draft.geojson[proto.expression] = geoJSON;
            });
        }
    }, [stackQuery.data]);

    const handleClick = useCallback(
        (evt: MapLayerMouseEvent) => {
            setCoordinates(evt.point);
            setExpression(
                `${evt.lngLat.lat.toFixed(6)}, ${evt.lngLat.lng.toFixed(6)}`
            );
            /* const feature = evt.features?.[0];
        if (!feature) return;
        const { ns, id } = feature.properties;
        if (!ns || !id) return; */

            /* setAppAtom((draft) => {
                draft.stacks[stackId] = {
                    coordinates: evt.point,
                    expression: `${evt.lngLat.lat.toFixed(
                        6
                    )}, ${evt.lngLat.lng.toFixed(6)}`,
                };
            }); */

            return;
        },
        [setCoordinates, setExpression]
    );

    const draggableStacks = useMemo(() => {
        return pickBy(stacks, (stack) => !stack.docked);
    }, [stacks]);

    const dockedStacks = useMemo(() => {
        return pickBy(stacks, (stack) => stack.docked);
    }, [stacks]);

    const mvt = new MVTLayer({
        data: ['api/tiles/base/{z}/{x}/{y}.mvt'],
        minZoom: 10,
        maxZoom: 16,
        /* getLineColor: (f: Feature) => {
            return 'transparent';
        }, */
        getFillColor: (f: Feature) => {
            if (f.properties.layerName === 'query') {
                //console.log(f);
            }
            // console.log(f);
            return 'transparent';
        },

        //onDataLoad: (data: $FixMe) => console.log(data),
    });

    const layers = [mvt];
    console.log(layers);

    /* 
    not ideal that we're transforming to array and filtering such a large dataset here on every render. @TODO: improve performance 
    */
    const points = useMemo(() => {
        const features = Object.values(geojson)
            .flat()
            .flatMap((f) => f?.features ?? [])
            .filter(
                (f) =>
                    f.geometry.type === 'Point' &&
                    map
                        ?.getBounds()
                        ?.contains(f.geometry.coordinates as [number, number])
            );

        return features;
    }, [geojson, mapViewState]);

    console.log(points);

    return (
        <div
            {...props}
            ref={ref}
            className={twMerge(
                'h-full border-t border-graphite-20 relative',
                props.className
            )}
        >
            <MapLibre
                id={id}
                {...mapViewState}
                onMove={(evt) => {
                    setMapViewState(evt.viewState);
                    debouncedSetViewState(evt.viewState);
                }}
                onClick={handleClick}
                onMouseEnter={() => {
                    setCursor('pointer');
                }}
                onMouseLeave={() => {
                    setCursor('auto');
                }}
                cursor={cursor}
                attributionControl={false}
                mapStyle={diagonalBasemapStyle as StyleSpecification}
                interactive={true}
                interactiveLayerIds={['building', 'road']}
            >
                {points.map((point) => {
                    if (point.geometry.type !== 'Point') return null;
                    return (
                        <Marker
                            className="[&>svg]:fill-graphite-80"
                            latitude={point.geometry.coordinates[1]}
                            longitude={point.geometry.coordinates[0]}
                        >
                            {match(point.properties?.['-b6-icon'])
                                .with('dot', () => {
                                    return (
                                        <DotIcon className="fill-graphite-80" />
                                    );
                                })
                                .otherwise(() => {
                                    const icon = point.properties?.['-b6-icon'];
                                    if (!icon) return <DotIcon />;
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
                                })}
                        </Marker>
                    );
                })}
                <MapControls>
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
                <Source
                    id="geojson"
                    type="geojson"
                    data={Object.values(geojson).flat()}
                >
                    {/* <Layer {...geojsonLayer} /> */}
                </Source>
                {/*                 <DeckGLOverlay layers={layers} />
                 */}
                <div className="absolute top-16 left-2 flex flex-col gap-1">
                    {Object.entries(dockedStacks).map(([stackId, stack]) => {
                        return (
                            <StackAdapter
                                key={stackId}
                                stack={stack}
                                docked={true}
                                mapId={id}
                            />
                        );
                    })}
                </div>
                <DndContext
                    sensors={sensors}
                    modifiers={[restrictToWindowEdges]}
                    onDragStart={({ active }) => {
                        setActiveStackId(active.id);
                        setAppAtom((draft) => {
                            draft.stacks[active.id].transient = false;
                        });
                    }}
                    onDragEnd={({ active, delta }) => {
                        setAppAtom((draft) => {
                            const { coordinates } = draft.stacks[active.id];
                            if (!coordinates) return;
                            draft.stacks[active.id].coordinates = new Point(
                                coordinates.x + delta.x,
                                coordinates.y + delta.y
                            );
                        });
                        setActiveStackId(null);
                    }}
                >
                    <Droppable mapId={id}>
                        <AnimatePresence>
                            {Object.entries(draggableStacks).map(
                                ([stackId, stack]) => (
                                    <DraggableStack
                                        active={activeStackId === stackId}
                                        key={stackId}
                                        id={stackId}
                                        mapId={id}
                                        mapDimensions={dimensions}
                                        stack={stack}
                                    />
                                )
                            )}
                        </AnimatePresence>
                    </Droppable>
                </DndContext>
            </MapLibre>
        </div>
    );
}

const Droppable = ({
    children,
    mapId,
}: PropsWithChildren & { mapId: string }) => {
    const { isOver, setNodeRef } = useDroppable({
        id: `droppable-${mapId}`,
    });
    const style = {
        color: isOver ? 'green' : undefined,
    };
    return (
        <div ref={setNodeRef} style={style}>
            {children}
        </div>
    );
};

const variants = {
    hidden: {
        opacity: 0,
        scale: 0,
    },
    visible: {
        opacity: 1,
        scale: 1,
    },
};

const DraggableStack = ({
    id,
    mapId,
    active = false,
    stack,
}: PropsWithChildren & {
    mapId: string;
    mapDimensions: ChartDimensions & {
        width: number;
        height: number;
    };
    id: string;
    stack: AppStore['stacks'][string];
    active?: boolean;
}) => {
    const { attributes, transform, setNodeRef, listeners } = useDraggable({
        id,
    });

    const style = {
        transform: `${
            transform
                ? `translate3d(${transform.x}px, ${transform.y}px, 0)`
                : ''
        }`,
    };

    return (
        <div
            id={id}
            ref={setNodeRef}
            style={{
                ...style,
                top: stack.coordinates?.y,
                left: stack.coordinates?.x,
                position: 'absolute',
            }}
            className={twMerge(
                active && 'ring-2 ring-ultramarine-50 ring-opacity-40'
            )}
            {...listeners}
            {...attributes}
        >
            <motion.div
                variants={variants}
                initial="hidden"
                animate="visible"
                exit="hidden"
                transition={{
                    duration: 0.1,
                }}
            >
                <div>
                    <StackAdapter stack={stack} mapId={mapId} />
                </div>
            </motion.div>
        </div>
    );
};
