import * as PopoverPrimitive from '@radix-ui/react-popover';
import * as TooltipPrimitive from '@radix-ui/react-tooltip';
import { AnimatePresence, MotionProps, motion } from 'framer-motion';
import React, { useImperativeHandle } from 'react';

import useOverflow from '@/hooks/useOverflow';
import { getNodeText } from '@/utils/text';

/**
 * Wrapping component that will display a tooltip if the content is overflowing.
 */
export const TooltipOverflow = React.forwardRef<
    HTMLDivElement,
    React.HTMLAttributes<HTMLDivElement>
>(({ children, ...props }, forwardedRef) => {
    const innerRef = React.useRef<HTMLDivElement>(null);
    useImperativeHandle(forwardedRef, () => innerRef.current!, []);
    const childrenText = getNodeText(children);
    const isTextOverflowing = useOverflow(innerRef, childrenText);

    return (
        <TooltipPrimitive.Provider delayDuration={400}>
            <TooltipPrimitive.Root>
                <TooltipPrimitive.Trigger asChild>
                    <div
                        {...props}
                        ref={innerRef}
                        className="overflow-hidden overflow-ellipsis"
                    >
                        {children}
                    </div>
                </TooltipPrimitive.Trigger>
                <AnimatePresence>
                    {isTextOverflowing && (
                        <TooltipPrimitive.Portal>
                            <TooltipPrimitive.Content asChild>
                                <TooltipContent>{childrenText}</TooltipContent>
                            </TooltipPrimitive.Content>
                        </TooltipPrimitive.Portal>
                    )}
                </AnimatePresence>
            </TooltipPrimitive.Root>
        </TooltipPrimitive.Provider>
    );
});

export const TooltipContent = React.forwardRef<
    HTMLDivElement,
    React.HTMLAttributes<HTMLDivElement> &
        MotionProps & { type?: 'tooltip' | 'popover' }
>(({ children, type = 'tooltip' }, forwardedRef) => {
    const ArrowComponent =
        type === 'tooltip' ? TooltipPrimitive.Arrow : PopoverPrimitive.Arrow;
    return (
        <motion.div
            ref={forwardedRef}
            initial={{
                y: 3,
                opacity: 0.3,
            }}
            animate={{ y: 0, opacity: 1 }}
            exit={{ y: 3, opacity: 0.3 }}
            transition={{ duration: 0.2 }}
            className="bg-graphite-80 text-graphite-10 px-2 shadow py-1 rounded text-sm"
        >
            {children}
            <ArrowComponent
                width={11}
                height={5}
                className=" fill-graphite-80 stroke stroke-graphite-80"
            />
        </motion.div>
    );
});

export const Tooltip = React.forwardRef<
    HTMLDivElement,
    React.HTMLAttributes<HTMLDivElement> & { content: string }
>(({ children, content, ...props }, forwardedRef) => {
    return (
        <TooltipPrimitive.Provider>
            <TooltipPrimitive.Root>
                <TooltipPrimitive.Trigger asChild>
                    <div
                        ref={forwardedRef}
                        {...props}
                        className="overflow-hidden overflow-ellipsis"
                    >
                        {children}
                    </div>
                </TooltipPrimitive.Trigger>
                <AnimatePresence>
                    <TooltipPrimitive.Portal>
                        <TooltipPrimitive.Content asChild>
                            <TooltipContent>{content}</TooltipContent>
                        </TooltipPrimitive.Content>
                    </TooltipPrimitive.Portal>
                </AnimatePresence>
            </TooltipPrimitive.Root>
        </TooltipPrimitive.Provider>
    );
});
