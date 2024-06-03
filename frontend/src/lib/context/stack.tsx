import { OutlinerSpec, useOutlinersStore } from '@/stores/outliners';
import { Event } from '@/types/events';
import { NodeProto } from '@/types/generated/api';
import { UIResponseProto } from '@/types/generated/ui';
import {
    PropsWithChildren,
    createContext,
    useCallback,
    useContext,
    useMemo,
} from 'react';
import { useStack } from '../api/stack';

type StoreContext = {
    data?: { proto: UIResponseProto };
    outliner?: OutlinerSpec;
    origin?: OutlinerSpec;
    close: () => void;
    evaluateNode: (node: NodeProto) => void;
    evaluateExpressionInOutliner: (expression: string) => void;
};

const StackContext = createContext<StoreContext>({
    close: () => {},
    evaluateNode: () => {},
    evaluateExpressionInOutliner: () => {},
});

export const StackContextProvider = ({
    children,
    outliner,
    origin,
}: { outliner: OutlinerSpec; origin?: OutlinerSpec } & PropsWithChildren) => {
    const actions = useOutlinersStore((state) => state.actions);

    const data = useStack(outliner.world, outliner.request, outliner.data);

    const close = useCallback(() => {
        actions.remove(outliner.id);
    }, [actions, outliner.id]);

    const evaluateExpressionInOutliner = useCallback(
        (expression: string) => {
            const event: Event = 'os';

            actions.setRequest(outliner.id, {
                ...outliner.request,
                node: data.data?.proto.node,
                expression,
                locked: false,
                logEvent: event,
            });
        },
        [actions, outliner.id, outliner.request, data.data?.proto.node]
    );

    const evaluateNode = useCallback(
        (node: NodeProto) => {
            if (!outliner.request) return;
            const event: Event = 'oc';

            actions.add({
                id: `${outliner.id}-${event}-${JSON.stringify(node)}`,
                world: outliner.world,
                properties: {
                    active: true,
                    transient: false,
                    docked: false,
                    type: 'core',
                },
                request: {
                    root: outliner.request.root,
                    node,
                    locked: true,
                    logEvent: event,
                    logMapCenter: outliner.request.logMapCenter,
                    logMapZoom: outliner.request.logMapZoom,
                    expression: '',
                },
            });
        },
        [actions, outliner.id, outliner.request, outliner.world]
    );

    const value = useMemo(() => {
        return {
            data: data.data,
            outliner,
            origin,
            close,
            evaluateNode,
            evaluateExpressionInOutliner,
        };
    }, [
        data.data,
        outliner,
        origin,
        close,
        evaluateNode,
        evaluateExpressionInOutliner,
    ]);

    return (
        <StackContext.Provider value={value}>{children}</StackContext.Provider>
    );
};

export const useStackContext = () => {
    return useContext(StackContext);
};
