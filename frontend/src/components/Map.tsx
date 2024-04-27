import * as circleIcons from '@/assets/icons/circle';
import { AppStore, appAtom } from '@/atoms/app';
import { viewAtom } from '@/atoms/location';
import { MapControls } from '@/components/system/MapControls';
import { fetchB6 } from '@/lib/b6';
import { isSamePositionPoints } from '@/lib/map';
import { ChartDimensions, useChartDimensions } from '@/lib/useChartDimensions';
import { Event } from '@/types/events';
import { StackResponse } from '@/types/stack';
import {
    MapboxOverlay as DeckOverlay,
    MapboxOverlayProps,
} from '@deck.gl/mapbox';
import {
    DndContext,
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
import { AnimatePresence, motion } from 'framer-motion';
import { useAtom } from 'jotai';
import { debounce, isUndefined, pickBy, uniqWith } from 'lodash';
import { MapLayerMouseEvent, Point, StyleSpecification } from 'maplibre-gl';
import 'maplibre-gl/dist/maplibre-gl.css';
import {
    HTMLAttributes,
    PropsWithChildren,
    useCallback,
    useEffect,
    useMemo,
    useState,
} from 'react';
import { useHotkeys } from 'react-hotkeys-hook';
import {
    Map as MapLibre,
    Marker,
    ViewState,
    useControl,
    useMap,
} from 'react-map-gl/maplibre';
import { twMerge } from 'tailwind-merge';
import { match } from 'ts-pattern';
import { WorldShellAdapter } from './adapters/ShellAdapter';
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
    const mouseSensor = useSensor(MouseSensor);
    const touchSensor = useSensor(TouchSensor);
    const sensors = useSensors(pointerSensor, mouseSensor, touchSensor);

    const [viewState, setViewState] = useAtom(viewAtom);
    const [mapViewState, setMapViewState] = useState<ViewState>(viewState);

    // Debounce the view state update to avoid updating the URL too often
    // eslint-disable-next-line react-hooks/exhaustive-deps
    const debouncedSetViewState = useCallback(debounce(setViewState, 1000), [
        setViewState,
    ]);

    const [expression, setExpression] = useState<string | null>(null);
    const [coordinates, setCoordinates] = useState<Point>();
    const [eventType, setEventType] = useState<Event | null>(null);
    const [locked, setLocked] = useState(false);
    const [showWorldShell, setShowWorldShell] = useState(false);

    useHotkeys('shift+meta+b, `', () => {
        console.log('here');
        setShowWorldShell((prev) => !prev);
    });

    const stackQuery = useQuery({
        queryKey: ['stack', expression, locked, eventType],
        queryFn: () => {
            if (
                !expression ||
                !eventType ||
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
                locked,
                session: startup.session,
                logMapCenter: {
                    latE7: Math.round(map.getCenter().lat * 1e7),
                    lngE7: Math.round(map.getCenter().lng * 1e7),
                },
                logEvent: eventType,
                logMapZoom: map.getZoom(),
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
                const stacksToRemove = Object.keys(stacks).filter(
                    (stackId) =>
                        stacks[stackId].transient || !stacks[stackId].docked
                );

                stacksToRemove.forEach((stackId) => {
                    delete draft.stacks[stackId];
                    delete draft.geojson[stackId];
                });

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
            const evaluateLatLon = () => {
                setEventType('mlc');
                setExpression(
                    `${evt.lngLat.lat.toFixed(6)}, ${evt.lngLat.lng.toFixed(6)}`
                );
            };

            setCoordinates(evt.point);

            if (evt.originalEvent.shiftKey) {
                setLocked(false);
                evaluateLatLon();
            } else {
                setLocked(true);
                const features = map?.queryRenderedFeatures(evt.point);
                const feature = features?.[0];

                if (feature) {
                    setEventType('mfc');
                    const { ns, id } = feature.properties;
                    const type = match(feature.geometry.type)
                        .with('Point', () => 'point')
                        .with('LineString', () => 'path')
                        .with('Polygon', () => 'area')
                        .with('MultiPolygon', () => 'area')
                        .otherwise(() => null);

                    if (ns && id && type) {
                        setExpression(
                            `find-feature /${type}/${ns}/${BigInt(`0x${id}`)}`
                        );
                    }
                } else {
                    evaluateLatLon();
                }
            }

            return;
        },
        [setCoordinates, setExpression, map]
    );

    const draggableStacks = useMemo(() => {
        return pickBy(stacks, (stack) => !stack.docked);
    }, [stacks]);

    const dockedStacks = useMemo(() => {
        return pickBy(stacks, (stack) => stack.docked);
    }, [stacks]);

    //const layers = [mvt];

    /* 
    not ideal that we're transforming to array and filtering such a large dataset here on every render. @TODO: improve performance 
    */
    const points = useMemo(() => {
        const features = uniqWith(
            Object.values(geojson)
                .flat()
                .flatMap((f) => (isUndefined(f) ? [] : f?.features ?? [f]))
                .filter(
                    (f) =>
                        f.geometry.type === 'Point' &&
                        map
                            ?.getBounds()
                            ?.contains(
                                f.geometry.coordinates as [number, number]
                            )
                ),
            isSamePositionPoints
        );

        return features;
    }, [geojson, mapViewState]);

    const layers = useMemo(() => {
        const layers = Object.values(stacks).flatMap(
            (stack) => stack.proto.layers ?? []
        );
        return layers;
    }, [stacks]);

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
                interactive={true}
                interactiveLayerIds={['building', 'road']}
                dragRotate={false}
                mapStyle={diagonalBasemapStyle as StyleSpecification}
                boxZoom={false} // https://github.com/mapbox/mapbox-gl-js/issues/6971s
            >
                <GlobalShell show={showWorldShell} mapId={id} />

                {/*  <Source
                    id="diagonal"
                    type="vector"
                    minzoom={10}
                    maxzoom={16}
                    url={'http://localhost:5173/tiles/base/{z}/{x}/{y}.mvt'}
                    //tiles={['http://localhost:5173/tiles/base/{z}/{x}/{y}.mvt']}
                >
                    {diagonalBasemapStyle.layers.map((layer) => {
                        return (
                            <Layer {...(layer as LayerProps)} key={layer.id} />
                        );
                    })}
                </Source> */}
                {points.map((point, i) => {
                    if (point.geometry.type !== 'Point') return null;
                    return (
                        <Marker
                            key={i}
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
                                    if (!icon)
                                        return (
                                            <div className="w-2 h-2 rounded-full bg-ultramarine-50 border border-ultramarine-80" />
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

const GlobalShell = ({ show, mapId }: { show: boolean; mapId: string }) => {
    return (
        <AnimatePresence>
            {show && (
                <motion.div
                    initial={{
                        translateX: -100,
                    }}
                    animate={{
                        translateX: 0,
                    }}
                    className="absolute top-2 left-10 w-[95%] z-20 "
                >
                    <WorldShellAdapter mapId={mapId} />
                </motion.div>
            )}
        </AnimatePresence>
    );
};
