import { Comparator } from '@/components/Comparator';
import { ComparisonRequestProto } from '@/types/generated/ui';
import { StackResponse } from '@/types/stack';
import { $FixMe } from '@/utils/defs';
import { useQuery } from '@tanstack/react-query';
import { FeatureCollection, GeoJsonProperties, Geometry } from 'geojson';
import { PropsWithChildren, createContext, useContext, useEffect } from 'react';
import { useAppContext } from './app';

export type Comparator = {
    id: string;
    request?: ComparisonRequestProto;
    query?: ReturnType<typeof useQuery<StackResponse, Error>>;
    data?: StackResponse;
};

const testDataComparison = {
    baseline: {
        bars: [
            {
                value: 5,
                total: 20,
                index: 1,
                range: { value: 'test-a' },
            },
            {
                value: 15,
                total: 20,
                index: 2,
                range: { value: 'test-b' },
            },
        ],
    },
    scenarios: [
        {
            bars: [
                {
                    value: 18,
                    total: 20,
                    index: 1,
                    range: { value: 'test-a' },
                },
                {
                    value: 2,
                    total: 20,
                    index: 2,
                    range: { value: 'test-b' },
                },
            ],
        },
    ],
};

const ComparatorContext = createContext<{
    comparator: Comparator;
}>({
    comparator: {} as Comparator,
});

export const useComparatorContext = () => {
    return useContext(ComparatorContext);
};

export const ComparatorProvider = ({
    children,
    comparator,
}: PropsWithChildren & { comparator: Comparator }) => {
    const { createOutliner } = useAppContext();
    const query = useQuery<StackResponse, Error>({
        queryKey: [
            'comparison',
            comparator.request?.baseline,
            comparator.request?.scenarios,
            comparator.request?.analysis,
        ],
        queryFn: async () => {
            console.warn(
                'comparison request not yet available, returning template data.'
            );
            return new Promise((resolve) => {
                resolve({
                    geoJSON: [] as unknown as FeatureCollection<
                        Geometry,
                        GeoJsonProperties
                    >[],
                    proto: {
                        node: undefined,
                        geoJSON: [] as $FixMe,
                        layers: [],
                        mapCenter: { latE7: 0, lngE7: 0 },
                        locked: false,
                        chipValues: [],
                        logDetail: '',
                        tilesChanged: false,
                        expression: '',
                        highlighted: { namespaces: [], ids: [] },
                        stack: {
                            substacks: [
                                {
                                    lines: [{ comparison: testDataComparison }],
                                    collapsable: false,
                                },
                            ],
                        },
                    },
                });
            });
        },
    });

    useEffect(() => {
        if (!query.data) return;

        query.data.proto.stack?.substacks[0].lines.forEach((line) => {
            if (line.comparison) {
                createOutliner({
                    id: `comparison-baseline`,
                    properties: {
                        comparison: true,
                        scenario: comparator.request?.baseline as $FixMe,
                        docked: false,
                        transient: false,
                        coordinates: { x: 0, y: 0 },
                    },
                    data: {
                        ...query.data,
                        proto: {
                            ...query.data.proto,
                            stack: {
                                substacks: [
                                    {
                                        lines:
                                            line.comparison.baseline?.bars.map(
                                                (b) => {
                                                    return {
                                                        histogramBar: b,
                                                    };
                                                }
                                            ) ?? [],
                                        collapsable: false,
                                    },
                                ],
                            },
                        },
                    },
                });

                line.comparison?.scenarios.forEach((scenario, i) => {
                    const scenarioId = comparator.request?.scenarios[i];

                    if (scenarioId) {
                        createOutliner({
                            id: `comparison-scenario-${i}`,
                            properties: {
                                comparison: true,
                                scenario: scenarioId as $FixMe,
                                docked: false,
                                transient: false,
                                coordinates: { x: 0, y: 0 },
                            },
                            data: {
                                ...query.data,
                                proto: {
                                    ...query.data.proto,
                                    stack: {
                                        substacks: [
                                            {
                                                lines:
                                                    scenario.bars.map((b) => {
                                                        return {
                                                            histogramBar: b,
                                                        };
                                                    }) ?? [],
                                                collapsable: false,
                                            },
                                        ],
                                    },
                                },
                            },
                        });
                    }
                });
            }
        });
    }, [query.data]);

    const value = {
        comparator: {
            ...comparator,
            query,
            data: query.data || comparator.data,
        },
    };

    return (
        <ComparatorContext.Provider value={value}>
            {children}
        </ComparatorContext.Provider>
    );
};
