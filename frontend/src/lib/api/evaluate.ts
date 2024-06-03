import {
    EvaluateRequestProto,
    EvaluateResponseProto,
} from '@/types/generated/api';
import { useQuery } from '@tanstack/react-query';
import { b6 } from './client';

export const getEvaluate = (
    request: EvaluateRequestProto
): Promise<EvaluateResponseProto> => {
    return b6.post('evaluate', request);
};

export const useEvaluate = (request: EvaluateRequestProto) => {
    const query = useQuery({
        queryKey: [
            'evaluate',
            request.version,
            JSON.stringify(request.root),
            JSON.stringify(request.request),
        ],
        queryFn: () => getEvaluate(request),
    });

    return query;
};
