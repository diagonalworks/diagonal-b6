import { Cross1Icon, Share1Icon } from '@radix-ui/react-icons';
import React, { HtmlHTMLAttributes } from 'react';
import { twMerge } from 'tailwind-merge';
import { IconButton } from './IconButton';

export interface HeaderProps extends React.HTMLAttributes<HTMLDivElement> {}

const Root = React.forwardRef<HTMLDivElement, HeaderProps>(
    ({ children, className, ...props }, forwardedRef) => {
        return (
            <div
                {...props}
                ref={forwardedRef}
                className={twMerge(
                    'flex w-full justify-between items-center gap-4',
                    className
                )}
            >
                {children}
            </div>
        );
    }
);

const Label = React.forwardRef<
    HTMLSpanElement,
    HtmlHTMLAttributes<HTMLSpanElement>
>(({ children, className, ...props }, forwardedRef) => {
    return (
        <span {...props} ref={forwardedRef} className={twMerge('', className)}>
            {children}
        </span>
    );
});

const Actions = React.forwardRef<
    HTMLDivElement,
    Omit<HtmlHTMLAttributes<HTMLDivElement>, 'children'> & {
        share?: boolean;
        close?: boolean;
    }
>(({ className, ...props }, forwardedRef) => {
    return (
        <div
            {...props}
            ref={forwardedRef}
            className={twMerge('flex items-center', className)}
        >
            {props.share && (
                <IconButton>
                    <Share1Icon />
                </IconButton>
            )}
            {props.close && (
                <IconButton>
                    <Cross1Icon />
                </IconButton>
            )}
        </div>
    );
});

export const Header = Object.assign(Root, {
    Label,
    Actions,
});
