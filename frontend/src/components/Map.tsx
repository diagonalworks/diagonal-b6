import { appAtom } from '@/atoms/app';
import { viewAtom } from '@/atoms/location';
import { MapControls } from '@/components/system/MapControls';
import { fetchB6 } from '@/lib/b6';
import { StackWrapper } from '@/lib/renderer';
import { useChartDimensions } from '@/lib/useChartDimensions';
import { StackResponse } from '@/types/stack';
import {
    DndContext,
    UniqueIdentifier,
    useDraggable,
    useDroppable,
} from '@dnd-kit/core';
import { MinusIcon, PlusIcon } from '@radix-ui/react-icons';
import { useQuery } from '@tanstack/react-query';
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
    const [ref] = useChartDimensions({
        marginTop: 0,
        marginRight: 0,
        marginBottom: 0,
        marginLeft: 0,
    });
    const { [id]: map } = useMap();
    const [activeStackId, setActiveStackId] = useState<UniqueIdentifier | null>(
        null
    );

    const [stacks, setStacks] = useState<{
        [key: UniqueIdentifier]: {
            coordinates: Point;
            expression: string;
        };
    }>({});

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

            setStacks({
                ...stacks,
                [stackId]: {
                    coordinates: evt.point,
                    expression: `${evt.lngLat.lat.toFixed(
                        6
                    )}, ${evt.lngLat.lng.toFixed(6)}`,
                },
            });

            return;
        },
        [stacks]
    );

    console.log(stacks);

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
                    onDragStart={({ active }) => setActiveStackId(active.id)}
                    onDragEnd={({ active, delta }) => {
                        setStacks({
                            ...stacks,
                            [active.id]: {
                                ...stacks[active.id],
                                coordinates: new Point(
                                    stacks[active.id].coordinates.x + delta.x,
                                    stacks[active.id].coordinates.y + delta.y
                                ),
                            },
                        });
                    }}
                >
                    <Droppable>
                        {Object.entries(stacks).map(([stackId, stack]) => (
                            <DraggableStack
                                active={activeStackId === stackId}
                                key={id}
                                id={stackId}
                                mapId={id}
                                expression={stack.expression}
                                top={stack.coordinates.y}
                                left={stack.coordinates.x}
                            />
                        ))}
                    </Droppable>
                </DndContext>
            </MapLibre>
        </div>
    );
}

const Droppable = ({ children }: PropsWithChildren) => {
    const { isOver, setNodeRef } = useDroppable({
        id: 'droppable',
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

    const style = transform
        ? {
              transform: `translate3d(${transform.x}px, ${transform.y}px, 0)`,
          }
        : undefined;

    return (
        <button
            ref={setNodeRef}
            style={{
                ...style,
                top,
                left,
                position: 'absolute',
            }}
            className={twMerge(
                active && 'ring-2 ring-blue-500 ring-opacity-50'
            )}
            {...listeners}
            {...attributes}
        >
            {stackQuery.data && stackQuery.data.proto.stack && (
                <StackWrapper stack={stackQuery.data.proto.stack} />
            )}
        </button>
    );
};
/* fetch(`/api/stack`, {
    method: 'POST',
    headers: {
        'Content-Type': 'application/json',
    },
    body: JSON.stringify({
        node: null,
        locked: true,
        logEvent: 'mlc',
        ...(map?.getCenter() && {
            logMapCenter: {
                lat_e7: (map.getCenter().lat * 1e7).toFixed(0),
                lng_e7: (map.getCenter().lng * 1e7).toFixed(0),
            },
        }),
        context: startup?.context,
        session: startup?.session,
        expression: selected?.expression,
    }),
}).then((res) => res.json() as Promise<StackResponse>), */
