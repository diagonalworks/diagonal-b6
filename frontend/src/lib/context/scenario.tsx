import { StyleSpecification } from 'maplibre-gl';
import {
    PropsWithChildren,
    createContext,
    useCallback,
    useContext,
    useMemo,
} from 'react';

import basemapStyleRose from '@/components/diagonal-map-style-rose.json';
import basemapStyle from '@/components/diagonal-map-style.json';

import { Change, Scenario } from '@/atoms/app';
import {
    EvaluateRequestProto,
    EvaluateResponseProto,
} from '@/types/generated/api';
import { MapLayerProto } from '@/types/generated/ui';
import { $FixMe } from '@/utils/defs';
import { UseQueryResult, useQuery } from '@tanstack/react-query';
import { GeoJsonObject } from 'geojson';
import { isUndefined, pickBy } from 'lodash';
import { MapRef } from 'react-map-gl/maplibre';
import { b6, b6Path } from '../b6';
import { useAppContext } from './app';
import { OutlinerStore } from './outliner';

const ScenarioContext = createContext<{
    scenario: Scenario;
    tab: 'left' | 'right';
    mapStyle: StyleSpecification;
    outliners: Record<string, OutlinerStore>;
    createOutlinerInScenario: (outliner: OutlinerStore) => void;
    draggableOutliners: OutlinerStore[];
    dockedOutliners: OutlinerStore[];
    comparisonOutliners: OutlinerStore[];
    getVisibleMarkers: (map: MapRef) => $FixMe[];
    geoJSON: GeoJsonObject[];
    queryLayers: Array<{
        layer: MapLayerProto;
        histogram: OutlinerStore['histogram'];
    }>;
    isDefiningChange?: boolean;
    setWorldId: (id: string) => void;
    setWorldChange: (change: Change) => void;
    query?: UseQueryResult<EvaluateResponseProto>;
}>({
    tab: 'left',
    scenario: {} as Scenario,
    mapStyle: basemapStyle as StyleSpecification,
    outliners: {},
    draggableOutliners: [],
    dockedOutliners: [],
    comparisonOutliners: [],
    getVisibleMarkers: () => [],
    queryLayers: [],
    geoJSON: [],
    isDefiningChange: false,
    createOutlinerInScenario: () => {},
    setWorldId: () => {},
    setWorldChange: () => {},
});

/**
 * Hook to access the scenario context. Use this hook to access the scenario state and the methods to update it.
 */
export const useScenarioContext = () => {
    return useContext(ScenarioContext);
};

export const ScenarioProvider = ({
    children,
    scenario,
    tab,
}: {
    scenario: Scenario;
    tab: 'left' | 'right';
} & PropsWithChildren) => {
    const {
        app: { outliners },
        createOutliner,
        setApp,
    } = useAppContext();

    const query = useQuery<EvaluateResponseProto, Error>({
        queryKey: ['scenario', scenario.id],
        queryFn: async () => {
            return b6.evaluate({
                root: {
                    type: 'FeatureTypeCollection',
                    namespace: 'diagonal.works/world',
                    value: 0,
                },
                request: {
                    call: {
                        function: {
                            symbol: 'add-world-with-change',
                        },
                        args: [
                            {
                                literal: {
                                    featureIDValue: {
                                        type: 'FeatureTypeCollection',
                                        namespace: 'diagonal.works/world',
                                        value: 1,
                                    },
                                },
                            },
                            {
                                call: {
                                    function: {
                                        symbol: 'add-service',
                                    },
                                    args: [
                                        {
                                            literal: {
                                                featureIDValue: {
                                                    type: 'FeatureTypeArea',
                                                    namespace:
                                                        'openstreetmap.org/way',
                                                    value: 532767912,
                                                },
                                            },
                                        },
                                    ],
                                },
                            },
                        ],
                    },
                },
            } as unknown as EvaluateRequestProto);
        },
        enabled: scenario.id !== 'baseline',
    });

    const isDefiningChange = useMemo(() => {
        return scenario.id !== 'baseline' && isUndefined(scenario.worldId);
    }, [scenario.id, scenario.worldId]);

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

    const createOutlinerInScenario = useCallback(
        (outliner: OutlinerStore) => {
            _removeTransientStacks();
            createOutliner({
                ...outliner,
                properties: {
                    ...outliner.properties,
                    scenario: scenario.id,
                    changeable: isDefiningChange,
                },
            });
        },
        [scenario.id, isDefiningChange, setApp, _removeTransientStacks]
    );

    const scenarioOutliners = useMemo(() => {
        return pickBy(
            outliners,
            (outliner) => outliner.properties.scenario === scenario.id
        );
    }, [outliners, scenario.id]);

    const dockedOutliners = useMemo(() => {
        return Object.values(scenarioOutliners).filter(
            (outliner) => outliner.properties.docked
        );
    }, [scenarioOutliners]);

    const comparisonOutliners = useMemo(() => {
        return Object.values(scenarioOutliners).filter(
            (outliner) => outliner.properties.comparison
        );
    }, [scenarioOutliners]);

    const draggableOutliners = useMemo(() => {
        return Object.values(scenarioOutliners).filter(
            (outliner) =>
                !outliner.properties.docked && !outliner.properties.comparison
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
            tab === 'right'
                ? {
                      ...basemapStyleRose,
                      sources: {
                          ...basemapStyle.sources,
                          diagonal: {
                              ...basemapStyle.sources.diagonal,
                              tiles: [
                                  // so we can set a new basemap for this scenario
                                  `${window.location.origin}${b6Path}tiles/base/{z}/{x}/{y}.mvt`,
                              ],
                          },
                      },
                  }
                : basemapStyle
        ) as StyleSpecification;
    }, [tab]);

    /** temporary while we don't have an API route form making a change to the world */
    const setWorldId = useCallback(
        (id: string) => {
            setApp((draft) => {
                draft.scenarios[scenario.id].worldId = id;
            });
        },
        [setApp, scenario.id]
    );

    const setWorldChange = useCallback(
        (change: Change) => {
            setApp((draft) => {
                draft.scenarios[scenario.id].change = change;
            });
        },
        [setApp, scenario.id]
    );

    const value = useMemo(() => {
        return {
            tab,
            scenario,
            mapStyle,
            outliners: scenarioOutliners,
            draggableOutliners,
            dockedOutliners,
            comparisonOutliners,
            getVisibleMarkers,
            geoJSON,
            queryLayers,
            isDefiningChange,
            createOutlinerInScenario,
            setWorldChange,
            setWorldId,
            query,
        };
    }, [
        scenario,
        scenarioOutliners,
        tab,
        isDefiningChange,
        mapStyle,
        queryLayers,
        geoJSON,
        dockedOutliners,
        draggableOutliners,
        getVisibleMarkers,
        createOutlinerInScenario,
        comparisonOutliners,
        setWorldId,
        setWorldChange,
        query,
    ]);

    return (
        <ScenarioContext.Provider value={value}>
            {children}
        </ScenarioContext.Provider>
    );
};
