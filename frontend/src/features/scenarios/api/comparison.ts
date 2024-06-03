import { b6 } from '@/lib/api/client';
import {
    ComparisonLineProto,
    ComparisonRequestProto,
} from '@/types/generated/ui';
import { useQuery } from '@tanstack/react-query';
import { Comparison } from '../stores/comparisons';

const getComparison = (
    request: ComparisonRequestProto
): Promise<ComparisonLineProto> => {
    return b6.post('compare', request);
};

export const useComparison = ({
    baseline,
    scenarios,
    analysis,
}: Comparison) => {
    const query = useQuery({
        queryKey: [
            'comparison',
            JSON.stringify(baseline.featureId),
            JSON.stringify(scenarios.map((s) => s.featureId)),
            JSON.stringify(analysis),
        ],
        queryFn: () =>
            getComparison({
                baseline: baseline.featureId,
                scenarios: scenarios.map((s) => s.featureId),
                analysis,
            }),
    });

    return query;
};
