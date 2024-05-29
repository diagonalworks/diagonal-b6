import { INITIAL_CENTER } from '@/atoms/location';
import basemapStyleRose from '@/components/diagonal-map-style-rose.json';
import basemapStyle from '@/components/diagonal-map-style.json';
import { useMap } from '@/hooks/useMap';
import { changeMapStyleSource } from '@/lib/map';
import { useViewStore } from '@/stores/view';
import { World } from '@/stores/worlds';
import { MinusIcon, PlusIcon } from '@radix-ui/react-icons';
import { MapLayerMouseEvent, StyleSpecification } from 'maplibre-gl';
import 'maplibre-gl/dist/maplibre-gl.css';
import { PropsWithChildren, useCallback, useMemo, useState } from 'react';
import { Map as MapLibre, useMap as useMapLibre } from 'react-map-gl/maplibre';
import { twMerge } from 'tailwind-merge';
import { MapControls } from './system/MapControls';

const getTileSource = (root?: string) => {
    return `${window.location.origin}/tiles/base/{z}/{x}/{y}.mvt${
        root ? `?r=${root}` : ''
    }`;
};

export const Map = ({
    children,
    world,
    root,
    side,
}: {
    root: string;
    side: 'left' | 'right';
    world: World['id'];
} & PropsWithChildren) => {
    const view = useViewStore((state) => state.view);
    const { setView } = useViewStore((state) => state.actions);
    const [cursor, setCursor] = useState<'grab' | 'pointer'>('grab');
    const { [world]: maplibre } = useMapLibre();
    const [actions] = useMap({ id: world });

    const mapStyle = useMemo(() => {
        const tileSource = getTileSource(root);
        const map = (
            side === 'left' ? basemapStyle : basemapStyleRose
        ) as StyleSpecification;
        return changeMapStyleSource(map, tileSource);
    }, [root, side]);

    const handleClick = useCallback(
        (e: MapLayerMouseEvent) => {
            // if shift key is pressed, create a new unlocked outliner.
            if (e.originalEvent.shiftKey) {
                actions.evaluateLatLng({ e, locked: false });
            } else {
                const features = maplibre?.queryRenderedFeatures(e.point);
                const feature = features?.[0];
                if (feature) {
                    actions.evaluateFeature({ e, locked: true, feature });
                } else {
                    actions.evaluateLatLng({ e, locked: true });
                }
            }
        },
        [actions, maplibre]
    );

    return (
        <MapLibre
            key={world}
            id={world}
            mapStyle={mapStyle}
            interactive={true}
            interactiveLayerIds={['building', 'road']}
            cursor={cursor}
            {...{
                ...view,
                latitude: view.latitude ?? INITIAL_CENTER.lat,
                longitude: view.longitude ?? INITIAL_CENTER.lng,
                zoom: view.zoom ?? 16,
            }}
            onMove={(evt) => {
                setView(evt.viewState);
            }}
            onClick={handleClick}
            onMouseEnter={() => setCursor('pointer')}
            onMouseLeave={() => setCursor('grab')}
            antialias={true}
            attributionControl={false}
            dragRotate={false}
            boxZoom={false}
        >
            <MapControls
                className={twMerge(side === 'right' && 'right-0 left-auto')}
            >
                <MapControls.Button
                    onClick={() => maplibre?.zoomIn({ duration: 200 })}
                >
                    <PlusIcon />
                </MapControls.Button>
                <MapControls.Button
                    onClick={() => maplibre?.zoomOut({ duration: 200 })}
                >
                    <MinusIcon />
                </MapControls.Button>
            </MapControls>
            {children}
        </MapLibre>
    );
};
