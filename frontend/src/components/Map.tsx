import { useChartDimensions } from '@/lib/useChartDimensions';
import type { StyleSpecification } from 'maplibre-gl';
import 'maplibre-gl/dist/maplibre-gl.css';
import { HTMLAttributes } from 'react';
import { Map as MapLibre } from 'react-map-gl/maplibre';
import { twMerge } from 'tailwind-merge';
import diagonalBasemapStyle from './diagonal-map-style.json';

const InitalZoom = 16;
const InitialCenter = { latE7: 515361156, lngE7: -1255161 };

export function Map({
    id,
    ...props
}: { id: string } & HTMLAttributes<HTMLDivElement>) {
    const [ref] = useChartDimensions({});

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
                initialViewState={{
                    longitude: InitialCenter.lngE7 / 1e7,
                    latitude: InitialCenter.latE7 / 1e7,
                    zoom: InitalZoom,
                }}
                attributionControl={false}
                mapStyle={diagonalBasemapStyle as StyleSpecification}
                onSourceData={(e) => console.log(e)}
            />
        </div>
    );
}
