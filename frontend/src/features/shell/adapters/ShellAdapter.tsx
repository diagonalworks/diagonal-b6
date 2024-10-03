import { useMemo } from 'react';

import { Shell } from '@/features/shell/components/Shell';
import { useMap } from '@/hooks/useMap';
import { useStackContext } from '@/lib/context/stack';
import { useOutlinersStore } from '@/stores/outliners';
import { useWorldStore } from '@/stores/worlds';
import { ShellLineProto } from '@/types/generated/ui';

export const ShellAdapter = ({ shell }: { shell?: ShellLineProto }) => {
    const { evaluateExpressionInOutliner, outliner } = useStackContext();
    const { addToShellHistory } = useOutlinersStore((state) => state.actions);

    const handleSubmit = (e: string) => {
        evaluateExpressionInOutliner(e);
        if (outliner) {
            addToShellHistory(outliner.id, e);
        }
    };

    const functions = useMemo(() => {
        if (!shell) return [];
        return (
            shell.functions?.map((func) => {
                return {
                    id: func,
                };
            }) ?? []
        );
    }, [shell?.functions]);

    return (
        <Shell
            functions={functions}
            onSubmit={handleSubmit}
            history={outliner?.properties.shellHistory}
        />
    );
};

export const WorldShellAdapter = ({ mapId }: { mapId: string }) => {
    const [{ evaluateExpression }] = useMap({ id: mapId });
    const world = useWorldStore((state) => state.worlds[mapId]);
    const { addToShellHistory } = useWorldStore((state) => state.actions);

    const handleSubmit = (e: string) => {
        evaluateExpression(e);
        addToShellHistory(mapId, e);
    };

    return (
        <Shell
            functions={[]}
            onSubmit={handleSubmit}
            history={world.shellHistory}
        />
    );
};
