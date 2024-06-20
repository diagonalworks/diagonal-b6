import {
    PropsWithChildren,
    createContext,
    useCallback,
    useContext,
    useMemo,
} from 'react';

import { useStack } from '@/api/stack';
import { OutlinerSpec, useOutlinersStore } from '@/stores/outliners';
import { Event } from '@/types/events';
import { NodeProto } from '@/types/generated/api';
import { UIResponseProto } from '@/types/generated/ui';

type StoreContext = {
    /* The outliner data */
    data?: { proto: UIResponseProto };
    /* The outliner specification */
    outliner?: OutlinerSpec;
    /**
     * The origin outliner specification, with which the current outliner is compared.
     * This property only exists if the stack corresponds to a comparison outliner.
     */
    origin?: OutlinerSpec;
    /**
     * Close the outliner, this will remove it from the outliners store.
     * @returns void
     */
    close: () => void;
    /**
     * Toggle the visibility of the layers in the outliner.
     * @returns void
     */
    toggleVisibility: () => void;
    /**
     * Evaluate a node in a new outliner.
     * @param node - The node to evaluate
     * @returns void
     */
    evaluateNode: (node: NodeProto) => void;
    /**
     * Evaluate an expression in the outliner.
     * @param expression - The expression to evaluate
     * @returns void
     */
    evaluateExpressionInOutliner: (expression: string) => void;
};

/**
 * Context for the stack.
 * This is used to provide the outliner data to the stack children, and for easy access to functions that manipulate the stack.
 */
const StackContext = createContext<StoreContext>({
    close: () => {},
    toggleVisibility: () => {},
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

    const toggleVisibility = useCallback(() => {
        actions.setVisibility(outliner.id, !outliner.properties.show);
    }, [actions, outliner.id, outliner.properties.show]);

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
                    show: true,
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
            toggleVisibility,
            evaluateNode,
            evaluateExpressionInOutliner,
        };
    }, [
        data.data,
        outliner,
        origin,
        close,
        toggleVisibility,
        evaluateNode,
        evaluateExpressionInOutliner,
    ]);

    return (
        <StackContext.Provider value={value}>{children}</StackContext.Provider>
    );
};

/**
 * The hook to use the stack context, which provides the outliner data and functions to manipulate the stack.
 * @returns The stack context
 */
export const useStackContext = () => {
    return useContext(StackContext);
};
