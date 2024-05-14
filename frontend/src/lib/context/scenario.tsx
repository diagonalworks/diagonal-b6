import { StyleSpecification } from 'maplibre-gl';
import {
    PropsWithChildren,
    createContext,
    useCallback,
    useContext,
    useEffect,
    useMemo,
} from 'react';

import basemapStyleRose from '@/components/diagonal-map-style-rose.json';
import basemapStyle from '@/components/diagonal-map-style.json';

import { ChangeFeature, ChangeFunction, Scenario } from '@/atoms/app';
import { EvaluateResponseProto, NodeProto } from '@/types/generated/api';
import { MapLayerProto } from '@/types/generated/ui';
import { $FixMe } from '@/utils/defs';
import { UseQueryResult, useQuery } from '@tanstack/react-query';
import { GeoJsonObject } from 'geojson';
import { pickBy } from 'lodash';
import { MapRef } from 'react-map-gl/maplibre';
import { b6Path } from '../b6';
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
        show?: boolean;
    }>;
    isDefiningChange?: boolean;
    addFeatureToChange: (feature: ChangeFeature) => void;
    removeFeatureFromChange: (feature: ChangeFeature) => void;
    setChangeFunction: (changeFunction: ChangeFunction) => void;
    setChangeAnalysis: (analysis: NodeProto) => void;
    queryScenario?: UseQueryResult<EvaluateResponseProto>;
    setSubmitted: (submitted: boolean) => void;
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
    addFeatureToChange: () => {},
    removeFeatureFromChange: () => {},
    setChangeFunction: () => {},
    setChangeAnalysis: () => {},
    setSubmitted: () => {},
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
    //const startupQuery = useAtomValue(startupQueryAtom);

    const setSubmitted = useCallback(
        (submitted: boolean) => {
            setApp((draft) => {
                draft.scenarios[scenario.id].submitted = submitted;
            });
        },
        [scenario.id, setApp]
    );

    const queryScenario = useQuery<EvaluateResponseProto, Error>({
        enabled: false,
        queryKey: ['scenario', scenario.id, JSON.stringify(scenario.change)],
        queryFn: async () => {
            return Promise.reject('not implemented');
            /* return b6.evaluate({
                root: startupQuery.data?.root,
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
                                        namespace: `${
                                            startupQuery.data?.root
                                                ?.namespace ?? 'diagonal.works'
                                        }/scenario`,
                                        value: +scenario.id,
                                    },
                                },
                            },
                            {
                                call: {
                                    function: {
                                        call: {
                                            function: {
                                                symbol: 'evaluate-feature',
                                            },
                                            args: [
                                                {
                                                    literal: {
                                                        featureIDValue: {
                                                            type: 'FeatureTypeExpression',
                                                            namespace:
                                                                'diagonal.works/skyline-demo-05-2024',
                                                            value: 5,
                                                        },
                                                    },
                                                },
                                            ],
                                        },
                                    },
                                    args: [],
                                },
                            },
                        ],
                    },
                },
            } as unknown as EvaluateRequestProto); */
        },
    });

    const addFeatureToChange = useCallback(
        (feature: ChangeFeature) => {
            setApp((draft) => {
                draft.scenarios[scenario.id].change = {
                    ...draft.scenarios[scenario.id].change,
                    features: [
                        ...(draft.scenarios[scenario.id].change?.features ||
                            []),
                        feature,
                    ],
                };
            });
        },
        [scenario.id, setApp]
    );

    const removeFeatureFromChange = useCallback(
        (feature: ChangeFeature) => {
            setApp((draft) => {
                if (!draft.scenarios[scenario.id].change?.features) return;
                draft.scenarios[scenario.id].change = {
                    ...draft.scenarios[scenario.id].change,
                    features:
                        draft.scenarios[scenario.id].change?.features?.filter(
                            (f) => f.expression !== feature.expression
                        ) || [],
                };
            });
        },
        [scenario.id, setApp]
    );

    const setChangeFunction = useCallback(
        (changeFunction: ChangeFunction) => {
            setApp((draft) => {
                draft.scenarios[scenario.id].change = {
                    ...draft.scenarios[scenario.id].change,
                    changeFunction,
                };
            });
        },
        [scenario.id, setApp]
    );

    const setChangeAnalysis = useCallback(
        (analysis: NodeProto) => {
            setApp((draft) => {
                draft.scenarios[scenario.id].change = {
                    ...draft.scenarios[scenario.id].change,
                    analysis,
                };
            });
        },
        [scenario.id, setApp]
    );

    useEffect(() => {
        //console.log(queryScenario);
        return;
        /* setApp((draft) => {
            draft.scenarios[scenario.id].node = queryScenario.data?.result
        }); */
    }, [queryScenario]);

    const isDefiningChange = useMemo(() => {
        return scenario.id !== 'baseline' && !scenario.worldCreated;
    }, [scenario.id, scenario.node]);

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
                    show: outliner.active,
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
            addFeatureToChange,
            removeFeatureFromChange,
            setChangeFunction,
            setChangeAnalysis,
            queryScenario,
            setSubmitted,
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
        addFeatureToChange,
        removeFeatureFromChange,
        setChangeFunction,
        setChangeAnalysis,
        queryScenario,
        setSubmitted,
    ]);

    return (
        <ScenarioContext.Provider value={value}>
            {children}
        </ScenarioContext.Provider>
    );
};
