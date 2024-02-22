import * as CollapsiblePrimitive from '@radix-ui/react-collapsible';
import { AnimatePresence, motion } from 'framer-motion';
import React from 'react';
import { twMerge } from 'tailwind-merge';

const Trigger = React.forwardRef<
    HTMLButtonElement,
    CollapsiblePrimitive.CollapsibleTriggerProps &
        React.RefAttributes<HTMLButtonElement>
>(({ children, className, ...props }, forwardedRef) => {
    return (
        <CollapsiblePrimitive.Trigger
            {...props}
            ref={forwardedRef}
            className={twMerge('cursor-pointer select-none', className)}
        >
            {children}
        </CollapsiblePrimitive.Trigger>
    );
});

/* [&+.line-stack]:border-t-0
 */

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
                        'text-base overflow-hidden [&_.line]:border-t-0',
                        className
                    )}
                >
                    {children}
                </motion.div>
            </CollapsiblePrimitive.Content>
        </AnimatePresence>
    );
});

export const Stack = Object.assign(CollapsiblePrimitive.Root, {
    Trigger,
    Content,
});
