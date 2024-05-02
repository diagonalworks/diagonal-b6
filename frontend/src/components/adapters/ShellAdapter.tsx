import { Shell } from '@/components/system/Shell';
import { useAppContext } from '@/lib/context/app';
import { useOutlinerContext } from '@/lib/context/outliner';

import { ShellLineProto } from '@/types/generated/ui';
import { useMemo } from 'react';
import { Point } from 'react-map-gl/maplibre';

export const ShellAdapter = ({ shell }: { shell?: ShellLineProto }) => {
    const { outliner, setRequest } = useOutlinerContext();

    const handleSubmit = (e: string) => {
        if (!outliner.data?.proto) return;
        const { node, locked } = outliner.data?.proto;
        setRequest({
            ...outliner.request,
            node,
            locked: locked ?? false,
            expression: e,
            eventType: 'os',
        });
    };

    const functions = useMemo(() => {
        if (!shell) return [];
        return shell.functions.map((func) => {
            return {
                id: func,
            };
        });
    }, [shell?.functions]);

    return <Shell functions={functions} onSubmit={handleSubmit} />;
};

export const WorldShellAdapter = ({ mapId }: { mapId: string }) => {
    const { createOutliner } = useAppContext();

    const handleSubmit = (e: string) => {
        createOutliner({
            id: `stack_shell_${e}`,
            properties: {
                tab: mapId,
                docked: false,
                transient: true,
                coordinates: { x: 8, y: 60 } as Point,
            },
            request: {
                expression: e,
                locked: false,
                eventType: 'ws',
            },
        });
    };

    return <Shell functions={[]} onSubmit={handleSubmit} />;
};
