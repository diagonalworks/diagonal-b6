import { startupQueryAtom } from '@/atoms/startup';
import { Comparator } from '@/components/Comparator';
import { FeatureIDProto } from '@/types/generated/api';
import { ComparisonLineProto } from '@/types/generated/ui';
import { Docked } from '@/types/startup';
import { $FixMe } from '@/utils/defs';
import { useQuery } from '@tanstack/react-query';
import { useAtomValue } from 'jotai';
import { isEqual } from 'lodash';
import {
    PropsWithChildren,
    createContext,
    useContext,
    useEffect,
    useMemo,
} from 'react';
import { b6 } from '../b6';
import { useAppContext } from './app';

export type Comparator = {
    id: string;
    baseline?: string;
    scenarios?: string[];
    analysis?: FeatureIDProto;
    query?: ReturnType<typeof useQuery<ComparisonLineProto, Error>>;
    data?: ComparisonLineProto;
};

const ComparatorContext = createContext<{
    comparator: Comparator;
    analysis?: Docked;
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
    const {
        app: { scenarios },
    } = useAppContext();
    const startupQuery = useAtomValue(startupQueryAtom);

    const query = useQuery<ComparisonLineProto, Error>({
        queryKey: [
            'comparison',
            comparator?.baseline,
            comparator?.scenarios,
            comparator?.analysis,
        ],
        queryFn: async () => {
            return b6.compare({
                analysis: comparator?.analysis,
                baseline: {
                    ...startupQuery.data?.root,
                    value: 0,
                },
                scenarios: comparator?.scenarios?.flatMap((s) => {
                    const scenarioFeatureId = scenarios[s].featureId;
                    if (!scenarioFeatureId) return [];
                    return scenarioFeatureId;
                }),
            });
        },
    });

    const analysis = useMemo(() => {
        return startupQuery.data?.docked?.find((d) => {
            return isEqual(d.proto.stack?.id, comparator.analysis);
        });
    }, [startupQuery.data?.docked, comparator.analysis]);

    useEffect(() => {
        if (!query.data) return;

        if (query.data.baseline) {
            createOutliner({
                id: `${comparator.id}-baseline`,
                properties: {
                    comparison: true,
                    scenario: comparator?.baseline as $FixMe,
                    docked: false,
                    transient: false,
                    coordinates: { x: 0, y: 0 },
                },
                data: {
                    geoJSON: [],
                    proto: {
                        stack: {
                            substacks: [
                                {
                                    lines:
                                        query.data.baseline?.bars?.map((b) => {
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

        query.data.scenarios?.forEach((scenario, i) => {
            const scenarioId = comparator?.scenarios?.[i];

            if (scenarioId) {
                createOutliner({
                    id: `${comparator.id}-scenario-${i}`,
                    properties: {
                        comparison: true,
                        origin: `${comparator.id}-baseline`,
                        scenario: scenarioId as $FixMe,
                        docked: false,
                        transient: false,
                        coordinates: { x: 0, y: 0 },
                    },
                    data: {
                        geoJSON: [],
                        proto: {
                            layers: analysis ? analysis.proto.layers : [],
                            stack: {
                                substacks: [
                                    {
                                        lines:
                                            scenario?.bars?.map((b) => {
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
    }, [query.data, analysis]);

    const value = {
        analysis,
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
