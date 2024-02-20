import React from 'react';
import { twMerge } from 'tailwind-merge';

export interface LabelledIconProps
    extends React.HTMLAttributes<HTMLDivElement> {}

const Root = React.forwardRef<HTMLDivElement, LabelledIconProps>(
    ({ children, className, ...props }: LabelledIconProps, forwardedRef) => {
        return (
            <div
                {...props}
                className={twMerge(
                    'flex gap-2 items-center text-graphite-100 w-fit ',
                    className
                )}
                ref={forwardedRef}
            >
                {children}
            </div>
        );
    }
);

const Icon = React.forwardRef<
    HTMLSpanElement,
    React.HTMLAttributes<HTMLSpanElement> & { children: React.ReactNode }
>(({ className, children, ...props }, forwardedRef) => {
    return (
        <span
            {...props}
            className={twMerge('[&>svg]:fill-graphite-80', className)}
            ref={forwardedRef}
        >
            {children}
        </span>
    );
});

export const Label = React.forwardRef<
    HTMLSpanElement,
    React.HTMLAttributes<HTMLSpanElement>
>(({ children, className, ...props }, forwardedRef) => {
    return (
        <span
            {...props}
            ref={forwardedRef}
            className={twMerge('text-graphite-100 text-base', className)}
        >
            {children}
        </span>
    );
});

export const LabelledIcon = Object.assign(Root, {
    Icon,
    Label,
});
