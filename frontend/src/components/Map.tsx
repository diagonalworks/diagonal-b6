import { appAtom } from '@/atoms/app';
import { viewAtom } from '@/atoms/location';
import { MapControls } from '@/components/system/MapControls';
import { fetchB6 } from '@/lib/b6';
import { useChartDimensions } from '@/lib/useChartDimensions';
import { StackResponse } from '@/types/stack';
import { FrameIcon, MinusIcon, PlusIcon } from '@radix-ui/react-icons';
import { useQuery } from '@tanstack/react-query';
import { useAtom, useAtomValue } from 'jotai';
import { debounce } from 'lodash';
import type {
    MapLayerMouseEvent,
    Point,
    StyleSpecification,
} from 'maplibre-gl';
import 'maplibre-gl/dist/maplibre-gl.css';
import { HTMLAttributes, useCallback, useState } from 'react';
import { Map as MapLibre, ViewState, useMap } from 'react-map-gl/maplibre';
import { twMerge } from 'tailwind-merge';
import { match } from 'ts-pattern';
import diagonalBasemapStyle from './diagonal-map-style.json';
import { Header } from './system/Header';
import { LabelledIcon } from './system/LabelledIcon';
import { Line } from './system/Line';
import { Stack } from './system/Stack';

export function Map({
    id,
    ...props
}: { id: string } & HTMLAttributes<HTMLDivElement>) {
    const [ref] = useChartDimensions({});
    const { [id]: map } = useMap();
    const [selected, setSelected] = useState<{
        expression: string;
        coordinates: Point;
    }>();
    const { startup } = useAtomValue(appAtom);
    console.log({ startup });
    const stackQuery = useQuery({
        queryKey: ['stack', selected?.expression],
        queryFn: () => {
            if (
                !selected?.expression ||
                !startup?.session ||
                !map?.getCenter() ||
                map?.getZoom() === undefined
            ) {
                return null;
            }

            return fetchB6('stack', {
                expression: selected.expression,
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
            !!selected?.expression &&
            !!startup?.session &&
            map?.getCenter() &&
            map?.getZoom() !== undefined,
    });

    const [viewState, setViewState] = useAtom(viewAtom);
    const [mapViewState, setMapViewState] = useState<ViewState>(viewState);

    // Debounce the view state update to avoid updating the URL too often
    // eslint-disable-next-line react-hooks/exhaustive-deps
    const debouncedSetViewState = useCallback(debounce(setViewState, 1000), [
        setViewState,
    ]);

    const handleClick = useCallback((evt: MapLayerMouseEvent) => {
        const feature = evt.features?.[0];
        if (!feature) return;
        const { ns, id } = feature.properties;
        if (!ns || !id) return;
        console.log(evt);

        setSelected({
            expression: `${evt.lngLat.lat.toFixed(6)}, ${evt.lngLat.lng.toFixed(
                6
            )}`,
            coordinates: evt.point,
        });
        return;
    }, []);

    console.log({
        selectedFeatureStack: stackQuery.data,
        selected,
    });

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
                {stackQuery.data && (
                    <Stack
                        className="absolute"
                        style={{
                            left: selected?.coordinates.x,
                            top: selected?.coordinates.y,
                        }}
                    >
                        {stackQuery.data.proto.stack?.substacks.map((s) => {
                            return Object.entries(s).map(([k, v]) => {
                                return match(k)
                                    .with('lines', () => {
                                        return v.map((l) => {
                                            console.log(l);
                                            return (
                                                <Line>
                                                    {Object.entries(l).map(
                                                        ([lk, lv]) => {
                                                            console.log(lk);
                                                            return match(lk)
                                                                .with(
                                                                    'header',
                                                                    () => {
                                                                        console.log(
                                                                            lv
                                                                        );
                                                                        return (
                                                                            <Header>
                                                                                <LabelledIcon>
                                                                                    <LabelledIcon.Icon>
                                                                                        <FrameIcon />
                                                                                    </LabelledIcon.Icon>
                                                                                    <LabelledIcon.Label>
                                                                                        {
                                                                                            lv
                                                                                                .title
                                                                                                .labelledIcon
                                                                                                .label
                                                                                        }
                                                                                    </LabelledIcon.Label>
                                                                                </LabelledIcon>
                                                                            </Header>
                                                                        );
                                                                    }
                                                                )
                                                                .otherwise(
                                                                    () => (
                                                                        <span>
                                                                            @TODO
                                                                        </span>
                                                                    )
                                                                );
                                                        }
                                                    )}
                                                </Line>
                                            );
                                        });
                                    })
                                    .otherwise(() => null);
                            });
                        })}
                    </Stack>
                )}
            </MapLibre>
        </div>
    );
}

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
