import { Cross1Icon, Link2Icon, MagnifyingGlassIcon } from '@radix-ui/react-icons';
import * as PopoverPrimitive from '@radix-ui/react-popover';
import { omit } from 'lodash';
import React, { HtmlHTMLAttributes, useEffect } from 'react';
import { twMerge } from 'tailwind-merge';

import { IconButton } from './IconButton';
import { TooltipContent } from './Tooltip';

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
        target?: boolean;
        close?: boolean;
        slotProps?: {
            share?: React.HTMLAttributes<HTMLButtonElement> & {
                popover?: {
                    open: boolean;
                    onOpenChange: (open: boolean) => void;
                    content: string;
                };
            };
            close?: React.HTMLAttributes<HTMLButtonElement>;
            target?: React.HTMLAttributes<HTMLButtonElement>;
        };
    }
>(
    (
        { className, close = false, share = false, target = false, slotProps, ...props },
        forwardedRef
    ) => {
        useEffect(() => {
            if (slotProps?.share?.popover?.open) {
                const timeout = setTimeout(() => {
                    slotProps.share?.popover?.onOpenChange(false);
                }, 1000);
                return () => clearTimeout(timeout);
            }
        }, [slotProps?.share?.popover]);

        return (
            <div
                {...props}
                ref={forwardedRef}
                className={twMerge('flex items-center', className)}
            >
                {share && (
                    <PopoverPrimitive.Root
                        open={slotProps?.share?.popover?.open}
                        onOpenChange={slotProps?.share?.popover?.onOpenChange}
                    >
                        <PopoverPrimitive.Trigger asChild>
                            <IconButton {...omit(slotProps?.share, 'popover')}>
                                <Link2Icon />
                            </IconButton>
                        </PopoverPrimitive.Trigger>
                        <PopoverPrimitive.Content
                            sideOffset={5}
                            side="right"
                            align="center"
                        >
                            <TooltipContent type="popover">
                                {slotProps?.share?.popover?.content}
                            </TooltipContent>
                        </PopoverPrimitive.Content>
                    </PopoverPrimitive.Root>
                )}
                {target && (
                    <IconButton {...slotProps?.target}>
                        <MagnifyingGlassIcon />
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
