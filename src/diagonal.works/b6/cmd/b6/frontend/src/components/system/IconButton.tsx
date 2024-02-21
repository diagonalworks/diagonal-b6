import React from 'react';
import { twMerge } from 'tailwind-merge';

export interface IconButtonProps
    extends React.HTMLAttributes<HTMLButtonElement> {}

export const IconButton = React.forwardRef<HTMLButtonElement, IconButtonProps>(
    ({ children, className, ...props }: IconButtonProps, forwardedRef) => {
        return (
            <button
                {...props}
                className={twMerge(
                    'flex items-center justify-center p-2 rounded-full text-graphite-80  hover:bg-graphite-20',
                    className
                )}
                ref={forwardedRef}
            >
                {children}
            </button>
        );
    }
);
