import { Slot } from '@radix-ui/react-slot';
import * as React from 'react';
import { twMerge } from 'tailwind-merge';
import { Spinner } from './Spinner';

export type ButtonProps = React.ButtonHTMLAttributes<HTMLButtonElement> & {
    asChild?: boolean;
    isLoading?: boolean;
    icon?: React.ReactNode;
};

export const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
    (
        { className, asChild = false, children, isLoading, icon, ...props },
        ref
    ) => {
        const Comp = asChild ? Slot : 'button';
        return (
            <Comp
                className={twMerge(
                    'flex items-center bg-graphite-90 justify-center gap-2 px-4 py-2 rounded-sm text-sm font-medium text-white hover:bg-graphite-80 focus:outline-none focus:ring-2 focus:ring-violet-50',
                    className
                )}
                ref={ref}
                {...props}
            >
                {isLoading && <Spinner size="sm" className="text-current" />}
                {!isLoading && icon && <span className="mr-2">{icon}</span>}
                <span className="mx-2">{children}</span>
            </Comp>
        );
    }
);
