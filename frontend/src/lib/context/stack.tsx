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
    close: () => void;
    evaluateNode: (node: NodeProto) => void;
};

const StackContext = createContext<StoreContext>({
    close: () => {},
    evaluateNode: () => {},
});

export const StackContextProvider = ({
    children,
    outliner,
}: { outliner: OutlinerSpec } & PropsWithChildren) => {
    const actions = useOutlinersStore((state) => state.actions);

    const data = useStack(outliner.request);

    const close = useCallback(() => {
        actions.remove(outliner.id);
    }, [actions, outliner.id]);

    const evaluateNode = useCallback(
        (node: NodeProto) => {
            const event: Event = 'oc';

            actions.add({
                id: `${outliner.id}-${event}-${JSON.stringify(node)}`,
                world: outliner.world,
                properties: {
                    active: true,
                    transient: false,
                    docked: false,
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
        [
            actions,
            outliner.id,
            outliner.request.root,
            outliner.request.logMapCenter,
            outliner.request.logMapZoom,
            outliner.world,
        ]
    );

    const value = useMemo(() => {
        return {
            data: data.data,
            close,
            evaluateNode,
        };
    }, [data.data, close, evaluateNode]);

    return (
        <StackContext.Provider value={value}>{children}</StackContext.Provider>
    );
};

export const useStackContext = () => {
    return useContext(StackContext);
};
