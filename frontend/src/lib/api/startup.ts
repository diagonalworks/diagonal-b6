import { useOutlinersStore } from '@/stores/outliners';
import { useViewStore } from '@/stores/view';
import { useWorkspaceStore } from '@/stores/workspace';
import { Event } from '@/types/events';
import { StartupRequest, StartupResponse } from '@/types/startup';
import { useQuery } from '@tanstack/react-query';
import { useEffect, useMemo } from 'react';
import { b6 } from './client';

const getStartup = (request: StartupRequest): Promise<StartupResponse> => {
    return b6.post('startup', null, {
        params: request,
    });
};

export const useStartup = () => {
    const root = useWorkspaceStore((state) => state?.root);
    const view = useViewStore((state) => state.initialView);
    const actions = useOutlinersStore((state) => state.actions);

    const request = useMemo(() => {
        return {
            ...(root && { r: root }),
            ...(view.latitude &&
                view.longitude && { ll: `${view.latitude},${view.longitude}` }),
            ...(view.zoom && { z: `${view.zoom}` }),
        };
    }, [root, view]);

    const query = useQuery({
        queryKey: ['startup', request.r, request.ll, request.z],
        queryFn: () => getStartup(request),
    });

    useEffect(() => {
        if (query.data) {
            const event: Event = 's';

            query.data.docked?.forEach((d, i) => {
                const id = d.proto.stack?.id;
                actions.add({
                    id: `docked-${i}`,
                    world: 'baseline',
                    properties: {
                        active: false,
                        docked: true,
                        transient: false,
                    },
                    /*  For consistency across how we handle data in outliners, we define the
                    request for the docked outliner instead of using the data directly.
                    */
                    request: {
                        root: query.data.root,
                        node: {
                            call: {
                                function: {
                                    symbol: 'find-collection',
                                },
                                args: [
                                    {
                                        literal: {
                                            featureIDValue: id,
                                        },
                                    },
                                ],
                            },
                        },
                        logEvent: event,
                        locked: true,
                    },
                });
            });
        }
    }, [query.data, actions]);

    return query;
};
