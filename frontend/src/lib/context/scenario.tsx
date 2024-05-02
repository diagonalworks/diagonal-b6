import { StyleSpecification } from 'maplibre-gl';
import {
    PropsWithChildren,
    createContext,
    useCallback,
    useContext,
    useMemo,
} from 'react';

import basemapStyleOrange from '@/components/diagonal-map-style-orange.json';
import basemapStyle from '@/components/diagonal-map-style.json';

import { MapLayerProto } from '@/types/generated/ui';
import { $FixMe } from '@/utils/defs';
import { pickBy } from 'lodash';
import { MapRef } from 'react-map-gl/maplibre';
import { useAppContext } from './app';
import { OutlinerStore } from './outliner';

const ScenarioContext = createContext<{
    id: string;
    change?: string;
    tab: 'left' | 'right';
    mapStyle: StyleSpecification;
    outliners: Record<string, OutlinerStore>;
    draggableOutliners: OutlinerStore[];
    dockedOutliners: OutlinerStore[];
    getVisibleMarkers: (map: MapRef) => $FixMe[];
    queryLayers: Array<{
        layer: MapLayerProto;
        histogram: OutlinerStore['histogram'];
    }>;
}>({
    tab: 'left',
    id: '',
    mapStyle: basemapStyle as StyleSpecification,
    outliners: {},
    draggableOutliners: [],
    dockedOutliners: [],
    getVisibleMarkers: () => [],
    queryLayers: [],
});

export const useScenarioContext = () => {
    return useContext(ScenarioContext);
};

export const ScenarioProvider = ({
    children,
    id,
    tab,
}: {
    id: string;
    tab: 'left' | 'right';
} & PropsWithChildren) => {
    const {
        app: { outliners },
    } = useAppContext();

    const scenarioOutliners = useMemo(() => {
        return pickBy(
            outliners,
            (outliner) => outliner.properties.scenario === id
        );
    }, [outliners, id]);

    const dockedOutliners = useMemo(() => {
        return Object.values(scenarioOutliners).filter(
            (outliner) => outliner.properties.docked
        );
    }, [scenarioOutliners]);

    const draggableOutliners = useMemo(() => {
        return Object.values(scenarioOutliners).filter(
            (outliner) => !outliner.properties.docked
        );
    }, [scenarioOutliners]);

    const queryLayers = useMemo(() => {
        return Object.values(scenarioOutliners).flatMap((outliner) => {
            return (
                outliner.data?.proto.layers?.map((l) => ({
                    layer: l,
                    histogram: outliner.histogram,
                })) || []
            );
        });
    }, [scenarioOutliners]);

    const getVisibleMarkers = useCallback(
        (map: MapRef) => {
            const features = Object.values(scenarioOutliners)
                .flatMap((outliner) => outliner.data?.geoJSON || [])
                .flat()
                .filter((f: $FixMe) => {
                    f.geometry.type === 'Point' &&
                        map
                            ?.getBounds()
                            ?.contains(
                                f.geometry.coordinates as [number, number]
                            );
                    return true;
                });
            return features;
        },
        [scenarioOutliners]
    );

    const mapStyle = useMemo(() => {
        return (
            tab === 'right' ? basemapStyleOrange : basemapStyle
        ) as StyleSpecification;
    }, [tab]);

    const value = useMemo(() => {
        return {
            tab,
            id,
            mapStyle,
            outliners: scenarioOutliners,
            draggableOutliners,
            dockedOutliners,
            getVisibleMarkers,
            queryLayers,
        };
    }, [id, scenarioOutliners]);

    return (
        <ScenarioContext.Provider value={value}>
            {children}
        </ScenarioContext.Provider>
    );
};
