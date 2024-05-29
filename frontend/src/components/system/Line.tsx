import React from 'react';
import { twMerge } from 'tailwind-merge';

export interface LineProps extends React.HTMLAttributes<HTMLDivElement> {}

/**
 * Renders a value on the right side of a Line.
 */
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
                    'text-graphite-100 text-base text-right',
                    className
                )}
                ref={forwardedRef}
            >
                {children}
            </span>
        );
    }
);

/**
 * Renders a button, should be used to wrap the Line contents when it should be clickable.
 */
const Button = React.forwardRef<
    HTMLButtonElement,
    React.HTMLAttributes<HTMLButtonElement> & { icon?: React.ReactNode }
>(({ children, icon, className, ...props }, forwardedRef) => {
    return (
        <button
            {...props}
            className={twMerge(
                'line-button flex items-center gap-1 w-full justify-between p-3',
                className
            )}
            ref={forwardedRef}
        >
            {children}
            {icon}
        </button>
    );
});

/**
 * Line component that can be used to render line atoms.
 */
const Root = React.forwardRef<HTMLDivElement, LineProps>(
    ({ children, className, ...props }: LineProps, forwardedRef) => {
        return (
            <div
                {...props}
                className={twMerge(
                    'line hover:has-[.tag]:bg-white has-[.shell]:p-0 p-3 border-b flex text-left items-center gap-1 w-80 min-h-11 box-border border-graphite-30 bg-white hover:bg-graphite-10 focus:bg-graphite-10 cursor-default has-[.line-button]:cursor-pointer has-[.line-button]:p-0',
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
    Button,
});
