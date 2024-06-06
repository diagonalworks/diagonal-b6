import * as CollapsiblePrimitive from '@radix-ui/react-collapsible';
import { AnimatePresence, motion } from 'framer-motion';
import { omit } from 'lodash';
import React from 'react';
import { twMerge } from 'tailwind-merge';

const StackContext = React.createContext<{
    collapsible: boolean;
}>({
    collapsible: true,
});

const useStackContext = () => {
    return React.useContext(StackContext);
};

const Root = React.forwardRef<
    HTMLDivElement,
    CollapsiblePrimitive.CollapsibleProps &
        React.HTMLAttributes<HTMLDivElement> & { collapsible?: boolean }
>(({ children, className, ...props }, forwardedRef) => {
    return (
        <StackContext.Provider
            value={{ collapsible: props.collapsible ?? false }}
        >
            <CollapsiblePrimitive.Root
                {...omit(props, 'collapsible')}
                ref={forwardedRef}
                className={twMerge(
                    'border box-border border-graphite-30  w-80',
                    props.open &&
                        props.collapsible &&
                        'border border-graphite-50 transition-colors  ',
                    'stack ',
                    '[&_.line]:border-t-0 [&_.stack]:border-0',
                    className
                )}
                open={props.collapsible ? props.open : true}
            >
                {children}
            </CollapsiblePrimitive.Root>
        </StackContext.Provider>
    );
});

/**
 * If the Stack is collapsible, this component will be used to trigger the collapse/expand action.
 * Depends on the stack context, so it should be used inside a Stack component.
 */
const Trigger = React.forwardRef<
    HTMLButtonElement,
    CollapsiblePrimitive.CollapsibleTriggerProps &
        React.RefAttributes<HTMLButtonElement>
>(({ children, className, ...props }, forwardedRef) => {
    const { collapsible } = useStackContext();

    return (
        <CollapsiblePrimitive.Trigger
            {...props}
            ref={forwardedRef}
            className={twMerge(
                collapsible &&
                    'cursor-pointer overflow-hidden select-none [&_.line]:data-[state=closed]:border-b-0',
                className
            )}
        >
            {children}
        </CollapsiblePrimitive.Trigger>
    );
});

const variants = {
    open: { height: 'fit-content', y: 0 },
    collapsed: { height: 0, y: -5 },
};

const Content = React.forwardRef<
    HTMLDivElement,
    Omit<CollapsiblePrimitive.CollapsibleContentProps, 'asChild'> &
        React.RefAttributes<HTMLDivElement> & {
            collapsible?: boolean;
            header?: boolean;
        }
>(
    (
        { children, collapsible, header = true, className, ...props },
        forwardedRef
    ) => {
        return (
            <AnimatePresence mode="sync">
                <CollapsiblePrimitive.Content
                    {...props}
                    ref={forwardedRef}
                    asChild
                >
                    <motion.div
                        variants={variants}
                        initial={collapsible ? 'collapsed' : 'open'}
                        animate="open"
                        exit={collapsible ? 'collapsed' : 'open'}
                        transition={{ duration: 0.5, type: 'spring' }}
                        className={twMerge(
                            'text-base overflow-hidden overflow-y-auto',
                            header &&
                                'group-[&_.line]:border-t-0 group-[&_.stack]:border-t-0',
                            className
                        )}
                    >
                        {children}
                    </motion.div>
                </CollapsiblePrimitive.Content>
            </AnimatePresence>
        );
    }
);

/**
 * Stack component used to render a (optionally) collapsible stack of Line components.
 */
export const Stack = Object.assign(Root, {
    Trigger,
    Content,
});
