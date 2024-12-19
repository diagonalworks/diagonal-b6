import {
    PropsWithChildren,
    createContext,
    useCallback,
    useContext,
    useMemo,
} from 'react';

import { isUndefined } from 'lodash';
import { useStack } from '@/api/stack';
import { OutlinerSpec, useOutlinersStore } from '@/stores/outliners';
import { useWorldStore } from '@/stores/worlds';
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
    evaluateNode: (node: NodeProto, shouldShowOutliner?: boolean, bustCache?: boolean) => void;
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
    const world = useWorldStore((state) => state.worlds[outliner.world]);

    const data = useStack(outliner.world, outliner.request, outliner.data, outliner.properties.magicNumber);

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
        (node: NodeProto, shouldShowOutliner?: boolean, bustCache?: boolean) => {
            const event: Event = 'oc';
            const root = outliner.request?.root ?? world.featureId;

            // TODO: There are at least two problems with this id:
            //
            //  1. It depends who opened it; it contains the id of the parent.
            //  This is probably not ideal.
            //
            //  2. The stringifyied node is a big JSON blob; we probably don't
            //  want that. It'd be better if there was a simple integer; i.e.
            //  a hash.
            //
            // Instead we might consider:
            //
            //  const id = `${event}-${JSON.stringify(node)}`
            //
            // i.e. just drop the first term relating to the outliner it was
            // opened with.
            const id = `${outliner.id}-${event}-${JSON.stringify(node)}`

            // If someone has asked to bust the (query) cache, then just
            // generate a number that will ensure that happens. This allows,
            // for example, the code in api/stack.ts to re-center the map
            // based on the returned data.
            //
            // TODO: In the future, we would not force a re-query here, we
            // would just be able to recenter the map based on the previous
            // query we obtained.
            const magicNumber = bustCache ? Date.now() : undefined;

            actions.add({
                id: id,
                world: outliner.world,
                properties: {
                    active: true,
                    transient: false,
                    docked: false,
                    type: 'core',
                    show: true,
                    showOutliner: isUndefined(shouldShowOutliner) ? true : shouldShowOutliner,
                    magicNumber: magicNumber,
                },
                request: {
                    root,
                    node,
                    locked: true,
                    logEvent: event,
                    logMapCenter: outliner.request?.logMapCenter,
                    logMapZoom: outliner.request?.logMapZoom,
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

    // If we've asked to not render, then just don't return anything. Note
    // that we must check that is is actually defined; otherwise we just show
    // it as normal.
    if ( !isUndefined(outliner.properties.showOutliner) && !outliner.properties.showOutliner ) {
        return <div />
    }

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
