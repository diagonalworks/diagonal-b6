import { appAtom } from '@/atoms/app';
import { Shell } from '@/components/system/Shell';
import { fetchB6 } from '@/lib/b6';
import { useStackContext } from '@/lib/context/stack';
import { ShellLineProto } from '@/types/generated/ui';
import { useQuery } from '@tanstack/react-query';
import { useAtom } from 'jotai';
import { useCallback, useEffect, useMemo, useState } from 'react';
import { useMap } from 'react-map-gl/maplibre';

export const ShellAdapter = ({ shell }: { shell: ShellLineProto }) => {
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
            }).then((res) => res.json());
        },
        enabled: !!expression,
    });

    useEffect(() => {
        if (expressionQuery.data && stackId) {
            console.log(expressionQuery.data);
            setApp((draft) => {
                draft.stacks[stackId].proto = {
                    ...app.stacks[stackId].proto,
                    ...expressionQuery.data.proto,
                };
            });
        }
    }, [expressionQuery.data]);

    const functions = useMemo(() => {
        return shell.functions.map((func) => {
            return {
                id: func,
                description: 'No description available',
            };
        });
    }, [shell.functions]);

    const onSubmitHandle = useCallback(
        ({ func, args }: { func: string; args: string }) => {
            setExpression(`${func} ${args}`);
        },
        [setExpression]
    );

    return <Shell functions={functions} onSubmit={onSubmitHandle} />;
};
