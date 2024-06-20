import { useViewStore } from '@/stores/view';
import { World } from '@/stores/worlds';
import { UIRequestProto } from '@/types/generated/ui';
import { StackResponse } from '@/types/stack';
import { useQueries, useQuery } from '@tanstack/react-query';
import { useEffect } from 'react';
import { useMap as useMapLibre } from 'react-map-gl/maplibre';
import { b6 } from './client';

const getStack = (request: UIRequestProto): Promise<StackResponse> => {
    return b6.post('stack', request);
};

const stackQueryParams = (
    request?: UIRequestProto,
    fallback?: StackResponse
) => {
    return {
        queryKey: [
            'stack',
            request ? 'request' : 'fallback',
            request?.expression,
            JSON.stringify(request?.root),
            JSON.stringify(request?.node),
            JSON.stringify(fallback?.proto),
        ],
        queryFn: () =>
            request ? getStack(request) : Promise.resolve(fallback),
    };
};

export const useStack = (
    world: World['id'],
    request?: UIRequestProto,
    fallback?: StackResponse
) => {
    const query = useQuery(stackQueryParams(request, fallback));
    const view = useViewStore((state) => state.view);
    const viewActions = useViewStore((state) => state.actions);
    const { [world]: maplibre } = useMapLibre();

    useEffect(() => {
        if (query.data) {
            const newCenter =
                query.data.proto.mapCenter?.latE7 &&
                query.data.proto.mapCenter?.lngE7
                    ? {
                          lat: query.data.proto.mapCenter.latE7 / 1e7,
                          lng: query.data.proto.mapCenter.lngE7 / 1e7,
                      }
                    : null;

            const currentBounds = maplibre?.getBounds();
            const isOutsideBounds =
                newCenter &&
                currentBounds &&
                !currentBounds.contains(newCenter);

            if (isOutsideBounds) {
                viewActions.setView({
                    ...view,
                    ...(query.data.proto.mapCenter?.latE7 && {
                        latitude: query.data.proto.mapCenter.latE7 / 1e7,
                    }),
                    ...(query.data.proto.mapCenter?.lngE7 && {
                        longitude: query.data.proto.mapCenter.lngE7 / 1e7,
                    }),
                });
            }
        }
    }, [query.data]);
    return query;
};

export const useStacks = (
    queries: { request?: UIRequestProto; fallback?: StackResponse }[]
) => {
    return useQueries({
        queries: queries.map((query) =>
            stackQueryParams(query.request, query.fallback)
        ),
    });
};
