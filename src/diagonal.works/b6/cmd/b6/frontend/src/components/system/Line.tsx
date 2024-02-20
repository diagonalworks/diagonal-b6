import React from 'react';
import { twMerge } from 'tailwind-merge';

export interface LineProps extends React.HTMLAttributes<HTMLDivElement> {}

export const Line = React.forwardRef<HTMLDivElement, LineProps>(
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
