import { HeaderAdapter } from '@/components/adapters/HeaderAdapter';
import { Line } from '@/components/system/Line';
import { useStack } from '@/lib/api/stack';
import { useOutlinersStore } from '@/stores/outliners';
import { useEffect } from 'react';
import { useComparison } from '../api/comparison';
import { Comparison } from '../stores/comparisons';
import ComparisonStack from './ComparisonStack';

export default function ComparisonCard({
    comparison,
}: {
    comparison: Comparison;
}) {
    const query = useComparison(comparison);
    const actions = useOutlinersStore((state) => state.actions);

    const originAnalysisQuery = useStack(comparison.baseline.id, {
        root: comparison.baseline.featureId,
        node: {
            call: {
                function: {
                    symbol: 'find-collection',
                },
                args: [
                    {
                        literal: {
                            featureIDValue: comparison.analysis,
                        },
                    },
                ],
            },
        },
    });

    const analysisTitle =
        originAnalysisQuery.data?.proto.stack?.substacks?.[1]?.lines?.map(
            (l) => l.header
        )[0];

    const targetAnalysisQuery = useStack(comparison.scenarios[0].id, {
        root: comparison.scenarios[0].featureId,
        node: {
            call: {
                function: {
                    symbol: 'find-collection',
                },
                args: [
                    {
                        literal: {
                            featureIDValue: comparison.analysis,
                        },
                    },
                ],
            },
        },
    });

    useEffect(() => {
        if (query.data?.baseline) {
            actions.add({
                id: `${comparison.id}-${comparison.baseline.id}`,
                world: comparison.baseline.id,
                properties: {
                    active: true,
                    transient: false,
                    docked: false,
                    type: 'comparison',
                },
                data: {
                    geoJSON: originAnalysisQuery.data?.geoJSON ?? [],
                    proto: {
                        layers: originAnalysisQuery?.data?.proto.layers ?? [],
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

        query.data?.scenarios?.forEach((scenario, i) => {
            const scenarioWorld = comparison.scenarios[i];

            if (scenarioWorld) {
                actions.add({
                    id: `${comparison.id}-${scenarioWorld.id}`,
                    world: scenarioWorld.id,
                    properties: {
                        active: true,
                        transient: false,
                        docked: false,
                        type: 'comparison',
                    },
                    data: {
                        geoJSON: targetAnalysisQuery.data?.geoJSON ?? [],
                        proto: {
                            layers:
                                targetAnalysisQuery?.data?.proto.layers ?? [],
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
        return () => {
            actions.remove(`${comparison.id}-${comparison.baseline.id}`);
            comparison.scenarios.forEach((scenario) => {
                actions.remove(`${comparison.id}-${scenario.id}`);
            });
        };
    }, [query.data, originAnalysisQuery.data, targetAnalysisQuery.data]);

    return (
        <div>
            <div className="border-t border-x border-graphite-30">
                {analysisTitle && (
                    <Line className="border-b-0">
                        <HeaderAdapter header={analysisTitle} />
                    </Line>
                )}
            </div>
            <div className="flex flex-row ">
                <div className="flex-grow">
                    {query.data?.baseline && (
                        <ComparisonStack
                            id={`${comparison.id}-${comparison.baseline.id}`}
                        />
                    )}
                </div>
                <div className="flex-grow">
                    {query.data?.scenarios?.[0] && comparison.scenarios[0] && (
                        <ComparisonStack
                            id={`${comparison.id}-${comparison.scenarios[0].id}`}
                            origin={`${comparison.id}-${comparison.baseline.id}`}
                        />
                    )}
                </div>
            </div>
        </div>
    );
}
