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
                    'p-3 border max-w-80 min-h-11 border-graphite-30 hover:bg-graphite-10',
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
