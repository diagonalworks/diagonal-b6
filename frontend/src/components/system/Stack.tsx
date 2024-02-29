import * as CollapsiblePrimitive from '@radix-ui/react-collapsible';
import { AnimatePresence, motion } from 'framer-motion';
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
                {...props}
                ref={forwardedRef}
                className={twMerge(
                    props.open &&
                        'border border-graphite-50 transition-colors w-fit  ',
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
                collapsible && 'cursor-pointer select-none',
                className
            )}
        >
            {children}
        </CollapsiblePrimitive.Trigger>
    );
});

const Content = React.forwardRef<
    HTMLDivElement,
    Omit<CollapsiblePrimitive.CollapsibleContentProps, 'asChild'> &
        React.RefAttributes<HTMLDivElement>
>(({ children, className, ...props }, forwardedRef) => {
    return (
        <AnimatePresence mode="sync">
            <CollapsiblePrimitive.Content {...props} ref={forwardedRef} asChild>
                <motion.div
                    initial={{ height: 0, y: -5 }}
                    animate={{ height: 'fit-content', y: 0 }}
                    exit={{ height: 0, y: -5 }}
                    transition={{ duration: 0.5, type: 'spring' }}
                    className={twMerge(
                        'text-base overflow-hidden [&_.line]:border-t-0 max-h-64 overflow-y-auto',
                        className
                    )}
                >
                    {children}
                </motion.div>
            </CollapsiblePrimitive.Content>
        </AnimatePresence>
    );
});

/**
 * Stack component used to render a (optionally) collapsible stack of Line components.
 */
export const Stack = Object.assign(Root, {
    Trigger,
    Content,
});
