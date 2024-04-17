import { appAtom } from '@/atoms/app';
import { viewAtom } from '@/atoms/location';
import { MapControls } from '@/components/system/MapControls';
import { fetchB6 } from '@/lib/b6';
import { StackWrapper } from '@/lib/renderer';
import { ChartDimensions, useChartDimensions } from '@/lib/useChartDimensions';
import { StackResponse } from '@/types/stack';
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
import { MinusIcon, PlusIcon } from '@radix-ui/react-icons';
import { useQuery } from '@tanstack/react-query';
import { AnimatePresence, motion } from 'framer-motion';
import { useAtom, useAtomValue } from 'jotai';
import { debounce, uniqueId } from 'lodash';
import { MapLayerMouseEvent, Point, StyleSpecification } from 'maplibre-gl';
import 'maplibre-gl/dist/maplibre-gl.css';
import {
    HTMLAttributes,
    PropsWithChildren,
    useCallback,
    useState,
} from 'react';
import { Map as MapLibre, ViewState, useMap } from 'react-map-gl/maplibre';
import { twMerge } from 'tailwind-merge';
import diagonalBasemapStyle from './diagonal-map-style.json';

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
    const [activeStackId, setActiveStackId] = useState<UniqueIdentifier | null>(
        null
    );

    const [{ stacks }, setAppAtom] = useAtom(appAtom);

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

    const handleClick = useCallback(
        (evt: MapLayerMouseEvent) => {
            /* const feature = evt.features?.[0];
        if (!feature) return;
        const { ns, id } = feature.properties;
        if (!ns || !id) return; */

            const stackId = uniqueId('stack_');

            setAppAtom((draft) => {
                draft.stacks[stackId] = {
                    coordinates: evt.point,
                    expression: `${evt.lngLat.lat.toFixed(
                        6
                    )}, ${evt.lngLat.lng.toFixed(6)}`,
                };
            });

            return;
        },
        [stacks, setAppAtom]
    );

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
                attributionControl={false}
                mapStyle={diagonalBasemapStyle as StyleSpecification}
                interactive={true}
                interactiveLayerIds={['building', 'road']}
            >
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
                <DndContext
                    sensors={sensors}
                    modifiers={[restrictToWindowEdges]}
                    onDragStart={({ active }) => setActiveStackId(active.id)}
                    onDragEnd={({ active, delta }) => {
                        setAppAtom((draft) => {
                            draft.stacks[active.id].coordinates = new Point(
                                stacks[active.id].coordinates.x + delta.x,
                                stacks[active.id].coordinates.y + delta.y
                            );
                        });
                        setActiveStackId(null);
                    }}
                >
                    <Droppable mapId={id}>
                        <AnimatePresence>
                            {Object.entries(stacks).map(([stackId, stack]) => (
                                <DraggableStack
                                    active={activeStackId === stackId}
                                    key={stackId}
                                    id={stackId}
                                    mapId={id}
                                    mapDimensions={dimensions}
                                    expression={stack.expression}
                                    top={stack.coordinates.y}
                                    left={stack.coordinates.x}
                                />
                            ))}
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

const DraggableStack = ({
    id,
    top,
    left,
    mapId,
    active = false,
    expression,
}: PropsWithChildren & {
    mapId: string;
    mapDimensions: ChartDimensions & {
        width: number;
        height: number;
    };
    expression: string;
    id: string;
    top: number;
    left: number;
    active?: boolean;
}) => {
    const { attributes, transform, setNodeRef, listeners } = useDraggable({
        id,
    });

    const { startup } = useAtomValue(appAtom);
    const { [mapId]: map } = useMap();

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
                top,
                left,
                position: 'absolute',
            }}
            className={twMerge(
                active && 'ring-2 ring-ultramarine-50 ring-opacity-40'
            )}
            {...listeners}
            {...attributes}
        >
            <motion.div
                initial={{
                    opacity: 0.4,
                    scale: 0.8,
                }}
                animate={{
                    opacity: 1,
                    scale: 1,
                }}
                exit={{
                    opacity: 0.4,
                    scale: 0.8,
                }}
            >
                <div>
                    {stackQuery.data && stackQuery.data.proto.stack && (
                        <StackWrapper
                            id={id}
                            stack={stackQuery.data.proto.stack}
                        />
                    )}
                </div>
            </motion.div>
        </div>
    );
};
