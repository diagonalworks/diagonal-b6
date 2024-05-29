import { UIRequestProto } from '@/types/generated/ui';
import { StackResponse } from '@/types/stack';
import { useQueries, useQuery } from '@tanstack/react-query';
import { b6 } from './client';

const getStack = (request: UIRequestProto): Promise<StackResponse> => {
    return b6.post('stack', request);
};

const stackQueryParams = (request: UIRequestProto) => {
    return {
        queryKey: [
            'stack',
            request.expression,
            JSON.stringify(request.root),
            JSON.stringify(request.node),
        ],
        queryFn: () => getStack(request),
    };
};

export const useStack = (request: UIRequestProto) => {
    return useQuery(stackQueryParams(request));
};

export const useStacks = (requests: UIRequestProto[]) => {
    return useQueries({
        queries: requests.map((request) => stackQueryParams(request)),
    });
};
