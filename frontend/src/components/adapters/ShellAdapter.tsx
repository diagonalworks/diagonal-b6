import { appAtom } from '@/atoms/app';
import { Shell } from '@/components/system/Shell';
import { fetchB6 } from '@/lib/b6';
import { useStackContext } from '@/lib/context/stack';
import { ShellLineProto } from '@/types/generated/ui';
import { StackResponse } from '@/types/stack';
import { useQuery } from '@tanstack/react-query';
import { useAtom, useAtomValue, useSetAtom } from 'jotai';
import { useEffect, useMemo, useState } from 'react';
import { Point, useMap } from 'react-map-gl/maplibre';

export const ShellAdapter = ({ shell }: { shell?: ShellLineProto }) => {
    const [expression, setExpression] = useState<string | null>(null);
    const stack = useStackContext();
    const { [stack.state.mapId]: map } = useMap();
    const [app, setApp] = useAtom(appAtom);
    const stackExpression = stack.state.stack?.proto.expression;
    const stackId = stack.state.stack?.id;

    const expressionQuery = useQuery({
        queryKey: ['shell', 'outliner', stackExpression, expression],
        queryFn: () => {
            if (!map || !stack.state.stack?.proto || !app?.startup || !stackId)
                return null;
            const { node, locked } = app.stacks[stackId].proto;
            const session = app.startup?.session;

            if (!expression || !session) return null;
            return fetchB6('stack', {
                expression,
                node,
                root: undefined,
                session,
                locked: locked ?? false,
                logEvent: 'os',
                logMapCenter: {
                    latE7: Math.round(map.getCenter().lat * 1e7),
                    lngE7: Math.round(map.getCenter().lng * 1e7),
                },
                logMapZoom: map.getZoom(),
            }).then((res) => res.json() as Promise<StackResponse>);
        },
        enabled: !!expression,
    });

    useEffect(() => {
        if (!expressionQuery.data) return;
        if (expressionQuery.data.proto && stackId) {
            const proto = expressionQuery.data.proto;

            setApp((draft) => {
                draft.stacks[stackId].proto = {
                    ...app.stacks[stackId].proto,
                    ...proto,
                };

                draft.geojson[stackId] = expressionQuery.data?.geoJSON ?? [];
            });
        }
    }, [expressionQuery.data]);

    const functions = useMemo(() => {
        if (!shell) return [];
        return shell.functions.map((func) => {
            return {
                id: func,
            };
        });
    }, [shell?.functions]);

    return <Shell functions={functions} onSubmit={setExpression} />;
};

export const WorldShellAdapter = ({ mapId }: { mapId: string }) => {
    const [expression, setExpression] = useState<string | null>(null);
    const setApp = useSetAtom(appAtom);
    const { startup } = useAtomValue(appAtom);
    const { [mapId]: map } = useMap();

    const expressionQuery = useQuery({
        queryKey: ['shell', 'world', expression],
        queryFn: () => {
            const session = startup?.session;
            if (!map || !session || !expression) return null;

            return fetchB6('stack', {
                expression,
                node: undefined,
                root: undefined,
                session,
                locked: false,
                logEvent: 'ws',
                logMapCenter: {
                    latE7: Math.round(map.getCenter().lat * 1e7),
                    lngE7: Math.round(map.getCenter().lng * 1e7),
                },
                logMapZoom: map.getZoom(),
            }).then((res) => res.json() as Promise<StackResponse>);
        },
        enabled: !!expression,
    });

    useEffect(() => {
        if (!expressionQuery.data) return;
        const { proto } = expressionQuery.data;
        if (proto) {
            const id = expressionQuery.data.proto.expression;
            setApp((draft) => {
                draft.stacks[id] = {
                    coordinates: { x: 8, y: 60 } as Point,
                    id,
                    proto,
                    docked: false,
                    transient: true,
                };
                draft.geojson[id] = expressionQuery.data?.geoJSON ?? [];
            });
        }
    }, [expressionQuery.data]);

    return <Shell functions={[]} onSubmit={setExpression} />;
};
