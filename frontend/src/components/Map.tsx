import { viewAtom } from '@/atoms/location';
import { useChartDimensions } from '@/lib/useChartDimensions';
import { useAtom } from 'jotai';
import { debounce } from 'lodash';
import type { StyleSpecification } from 'maplibre-gl';
import 'maplibre-gl/dist/maplibre-gl.css';
import { HTMLAttributes, useCallback, useState } from 'react';
import { Map as MapLibre, ViewState } from 'react-map-gl/maplibre';
import { twMerge } from 'tailwind-merge';
import diagonalBasemapStyle from './diagonal-map-style.json';

export function Map({
    id,
    ...props
}: { id: string } & HTMLAttributes<HTMLDivElement>) {
    const [ref] = useChartDimensions({});

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
                'maplibregl-map h-full border-t border-graphite-20  ',
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
                attributionControl={false}
                mapStyle={diagonalBasemapStyle as StyleSpecification}
                onSourceData={(e) => console.log(e)}
            />
        </div>
    );
}
