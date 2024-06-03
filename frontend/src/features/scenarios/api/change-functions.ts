import { useEvaluate } from '@/lib/api/evaluate';
import { World } from '@/stores/worlds';
import { FeatureType } from '@/types/generated/api';
import { useMemo } from 'react';

export const useChangeFunctions = ({ origin }: { origin: World }) => {
    const query = useEvaluate({
        root: origin.featureId,
        request: {
            call: {
                function: {
                    symbol: 'list-feature',
                },
                args: [
                    {
                        literal: {
                            featureIDValue: {
                                type: 'FeatureTypeCollection' as unknown as FeatureType,
                                // @TODO: this is currently hardcoded, but should be dynamic.
                                namespace:
                                    'diagonal.works/skyline-demo-05-2024',
                                value: 1,
                            },
                        },
                    },
                ],
            },
        },
    });

    const changes = useMemo(() => {
        const changes = query.data?.result?.literal?.collectionValue;
        if (!changes) return [];
        return (
            changes.values?.flatMap((v, i) => {
                if (!v.featureIDValue || !changes.keys?.[i].stringValue)
                    return [];
                return {
                    label: changes.keys?.[i].stringValue,
                    id: v.featureIDValue,
                };
            }) ?? []
        );
    }, [query.data]);

    return changes;
};
