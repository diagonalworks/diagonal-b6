import { Cross1Icon, Link2Icon } from '@radix-ui/react-icons';
import React, { HtmlHTMLAttributes } from 'react';
import { twMerge } from 'tailwind-merge';
import { IconButton } from './IconButton';

export interface HeaderProps extends React.HTMLAttributes<HTMLDivElement> {}

/**
 * A header component that can be used to display a title and actions. Often wrapped in a Line.
 */
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
        <span
            {...props}
            ref={forwardedRef}
            className={twMerge(' overflow-hidden', className)}
        >
            {children}
        </span>
    );
});

/**
 * An actions component that can be used to display actions in the header.
 * Currently supports share and close actions.
 *
 * @TODO: Discuss how flexible we want these actions to be, consider extracting action into its own
 * ActionItem component, so that it's more flexible.
 */
const Actions = React.forwardRef<
    HTMLDivElement,
    Omit<HtmlHTMLAttributes<HTMLDivElement>, 'children'> & {
        share?: boolean;
        close?: boolean;
        slotProps?: {
            share?: React.HTMLAttributes<HTMLButtonElement>;
            close?: React.HTMLAttributes<HTMLButtonElement>;
        };
    }
>(
    (
        { className, close = false, share = false, slotProps, ...props },
        forwardedRef
    ) => {
        return (
            <div
                {...props}
                ref={forwardedRef}
                className={twMerge('flex items-center', className)}
            >
                {share && (
                    <IconButton {...slotProps?.share}>
                        <Link2Icon />
                    </IconButton>
                )}
                {close && (
                    <IconButton {...slotProps?.close}>
                        <Cross1Icon />
                    </IconButton>
                )}
            </div>
        );
    }
);

export const Header = Object.assign(Root, {
    Label,
    Actions,
});
