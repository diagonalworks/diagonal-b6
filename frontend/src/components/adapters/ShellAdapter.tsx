import { Shell } from '@/components/system/Shell';
import { useMap } from '@/hooks/useMap';
import { useStackContext } from '@/lib/context/stack';

import { ShellLineProto } from '@/types/generated/ui';
import { useMemo } from 'react';

export const ShellAdapter = ({ shell }: { shell?: ShellLineProto }) => {
    const { evaluateExpressionInOutliner } = useStackContext();

    const handleSubmit = (e: string) => {
        evaluateExpressionInOutliner(e);
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

    return <Shell functions={functions} onSubmit={handleSubmit} />;
};

export const WorldShellAdapter = ({ mapId }: { mapId: string }) => {
    const [{ evaluateExpression }] = useMap({ id: mapId });

    const handleSubmit = (e: string) => {
        evaluateExpression(e);
    };

    return <Shell functions={[]} onSubmit={handleSubmit} />;
};
