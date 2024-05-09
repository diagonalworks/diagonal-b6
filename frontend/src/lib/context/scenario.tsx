import { StyleSpecification } from 'maplibre-gl';
import {
    PropsWithChildren,
    createContext,
    useCallback,
    useContext,
    useMemo,
    useState,
} from 'react';

import basemapStyleOrange from '@/components/diagonal-map-style-orange.json';
import basemapStyle from '@/components/diagonal-map-style.json';

import { MapLayerProto } from '@/types/generated/ui';
import { $FixMe } from '@/utils/defs';
import { GeoJsonObject } from 'geojson';
import { isUndefined, pickBy } from 'lodash';
import { MapRef } from 'react-map-gl/maplibre';
import { useAppContext } from './app';
import { OutlinerStore } from './outliner';

export type Change = {
    features: string[];
    function: string;
};

const ScenarioContext = createContext<{
    id: string;
    change: Change;
    setChange: (change: Change) => void;
    worldId?: string;
    tab: 'left' | 'right';
    mapStyle: StyleSpecification;
    outliners: Record<string, OutlinerStore>;
    createOutliner: (outliner: OutlinerStore) => void;
    draggableOutliners: OutlinerStore[];
    dockedOutliners: OutlinerStore[];
    getVisibleMarkers: (map: MapRef) => $FixMe[];
    geoJSON: GeoJsonObject[];
    queryLayers: Array<{
        layer: MapLayerProto;
        histogram: OutlinerStore['histogram'];
    }>;
    isDefiningChange?: boolean;
}>({
    tab: 'left',
    id: '',
    mapStyle: basemapStyle as StyleSpecification,
    outliners: {},
    draggableOutliners: [],
    dockedOutliners: [],
    getVisibleMarkers: () => [],
    queryLayers: [],
    geoJSON: [],
    isDefiningChange: false,
    change: {
        features: [],
        function: '',
    },
    setChange: () => {},
    createOutliner: () => {},
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
        setApp,
    } = useAppContext();

    const [change, setChange] = useState<Change>({
        features: [],
        function: '',
    });
    const [worldId] = useState<string>();

    const isDefiningChange = useMemo(() => {
        return id !== 'baseline' && isUndefined(worldId);
    }, [id, change]);

    const _removeTransientStacks = useCallback(() => {
        setApp((draft) => {
            for (const id in draft.outliners) {
                if (
                    draft.outliners[id].properties.transient &&
                    !draft.outliners[id].properties.docked
                ) {
                    delete draft.outliners[id];
                }
            }
        });
    }, [setApp]);

    const createOutliner = useCallback(
        (outliner: OutlinerStore) => {
            _removeTransientStacks();
            setApp((draft) => {
                draft.outliners[outliner.id] = {
                    ...outliner,
                    properties: {
                        ...outliner.properties,
                        scenario: id,
                        changeable: isDefiningChange,
                    },
                };
            });
        },
        [id, isDefiningChange, setApp, _removeTransientStacks]
    );

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

    const geoJSON = useMemo(() => {
        return Object.values(scenarioOutliners)
            .flatMap((outliner) => outliner.data?.geoJSON || [])
            .flat();
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
            geoJSON,
            queryLayers,
            change,
            setChange,
            isDefiningChange,
            createOutliner,
        };
    }, [
        id,
        scenarioOutliners,
        tab,
        change,
        setChange,
        isDefiningChange,
        mapStyle,
        queryLayers,
        geoJSON,
        dockedOutliners,
        draggableOutliners,
        getVisibleMarkers,
        createOutliner,
    ]);

    return (
        <ScenarioContext.Provider value={value}>
            {children}
        </ScenarioContext.Provider>
    );
};