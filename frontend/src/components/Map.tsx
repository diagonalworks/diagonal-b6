import { viewAtom } from '@/atoms/location';
import { MapControls } from '@/components/system/MapControls';
import { useChartDimensions } from '@/lib/useChartDimensions';
import { MinusIcon, PlusIcon } from '@radix-ui/react-icons';
import { useAtom } from 'jotai';
import { debounce } from 'lodash';
import type { StyleSpecification } from 'maplibre-gl';
import 'maplibre-gl/dist/maplibre-gl.css';
import { HTMLAttributes, useCallback, useRef, useState } from 'react';
import { Map as MapLibre, MapRef, ViewState } from 'react-map-gl/maplibre';
import { twMerge } from 'tailwind-merge';
import diagonalBasemapStyle from './diagonal-map-style.json';

export function Map({
    id,
    ...props
}: { id: string } & HTMLAttributes<HTMLDivElement>) {
    const [ref] = useChartDimensions({});
    const mapRef = useRef<MapRef>(null);

    const [viewState, setViewState] = useAtom(viewAtom);
    const [mapViewState, setMapViewState] = useState<ViewState>(viewState);

    // Debounce the view state update to avoid updating the URL too often
    // eslint-disable-next-line react-hooks/exhaustive-deps
    const debouncedSetViewState = useCallback(debounce(setViewState, 1000), [
        setViewState,
    ]);

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
                ref={mapRef}
                id={id}
                {...mapViewState}
                onMove={(evt) => {
                    setMapViewState(evt.viewState);
                    debouncedSetViewState(evt.viewState);
                }}
                attributionControl={false}
                mapStyle={diagonalBasemapStyle as StyleSpecification}
            >
                <MapControls>
                    <MapControls.Button
                        onClick={() =>
                            mapRef.current?.zoomIn({ duration: 200 })
                        }
                    >
                        <PlusIcon />
                    </MapControls.Button>
                    <MapControls.Button
                        onClick={() =>
                            mapRef.current?.zoomOut({ duration: 200 })
                        }
                    >
                        <MinusIcon />
                    </MapControls.Button>
                </MapControls>
            </MapLibre>
        </div>
    );
}
