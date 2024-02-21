import { getNodeText } from '@/lib/text';
import useOverflow from '@/lib/useOverflow';
import * as TooltipPrimitive from '@radix-ui/react-tooltip';
import { AnimatePresence, motion } from 'framer-motion';
import React, { useImperativeHandle } from 'react';

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
                                <motion.div
                                    initial={{ y: 3, opacity: 0.3 }}
                                    animate={{ y: 0, opacity: 1 }}
                                    exit={{ y: 3, opacity: 0.3 }}
                                    transition={{ duration: 0.2 }}
                                    className="bg-graphite-80 text-graphite-10 px-2 shadow py-1 rounded text-sm"
                                >
                                    {childrenText}
                                    <TooltipPrimitive.Arrow
                                        width={11}
                                        height={5}
                                        className=" fill-graphite-80 stroke stroke-graphite-80"
                                    />
                                </motion.div>
                            </TooltipPrimitive.Content>
                        </TooltipPrimitive.Portal>
                    )}
                </AnimatePresence>
            </TooltipPrimitive.Root>
        </TooltipPrimitive.Provider>
    );
});
