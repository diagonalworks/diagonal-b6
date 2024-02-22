import React from 'react';
import { twMerge } from 'tailwind-merge';

export interface LineProps extends React.HTMLAttributes<HTMLDivElement> {}

const Value = React.forwardRef<
    HTMLSpanElement,
    React.HTMLAttributes<HTMLSpanElement>
>(
    (
        {
            children,
            className,
            ...props
        }: React.HTMLAttributes<HTMLSpanElement>,
        forwardedRef
    ) => {
        return (
            <span
                {...props}
                className={twMerge(
                    'text-ultramarine-60 text-base text-right',
                    className
                )}
                ref={forwardedRef}
            >
                {children}
            </span>
        );
    }
);

const Root = React.forwardRef<HTMLDivElement, LineProps>(
    ({ children, className, ...props }: LineProps, forwardedRef) => {
        return (
            <div
                {...props}
                className={twMerge(
                    'line p-3 border flex items-center gap-1 max-w-80 min-h-11 box-border border-graphite-30 bg-white hover:bg-graphite-10 cursor-default',
                    className
                )}
                ref={forwardedRef}
            >
                {children}
            </div>
        );
    }
);

export const Line = Object.assign(Root, {
    Value,
});
