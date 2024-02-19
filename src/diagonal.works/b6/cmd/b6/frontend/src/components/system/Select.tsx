import { CheckIcon, ChevronDownIcon } from '@radix-ui/react-icons';
import * as RadixSelect from '@radix-ui/react-select';
import React from 'react';
import { twMerge } from 'tailwind-merge';

const SelectButton = React.forwardRef<
    React.ElementRef<typeof RadixSelect.Trigger>,
    React.ComponentPropsWithoutRef<typeof RadixSelect.Trigger>
>(({ children, className, ...props }, forwardedRef) => {
    return (
        <RadixSelect.Trigger
            {...props}
            className={twMerge(
                'flex items-center gap-1 [&_svg]:data-[state=open]:rotate-180 bg-graphite-20 px-1 py-0.5 rounded focus:outline-none focus:ring-2 focus-visible:ring-violet-50 focus:ring-offset-2 ',
                className
            )}
            ref={forwardedRef}
        >
            {children}
            <RadixSelect.Icon>
                <ChevronDownIcon className="transition-transform text-violet-50 " />
            </RadixSelect.Icon>
        </RadixSelect.Trigger>
    );
});

const SelectOptions = React.forwardRef<
    React.ElementRef<typeof RadixSelect.Content>,
    React.ComponentPropsWithoutRef<typeof RadixSelect.Content>
>(({ children, className, ...props }, forwardedRef) => {
    return (
        <RadixSelect.Portal>
            <RadixSelect.Content
                {...props}
                className={twMerge(
                    ' bg-white border border-graphite-20 shadow p-1.5 rounded ',
                    className
                )}
                ref={forwardedRef}
            >
                <RadixSelect.Viewport className="flex flex-col gap-1 ">
                    {children}
                </RadixSelect.Viewport>
            </RadixSelect.Content>
        </RadixSelect.Portal>
    );
});

export const SelectOption = React.forwardRef<
    React.ElementRef<typeof RadixSelect.Item>,
    React.ComponentPropsWithoutRef<typeof RadixSelect.Item>
>(({ children, className, ...props }, forwardedRef) => {
    return (
        <RadixSelect.Item
            {...props}
            className={twMerge(
                'flex flex-row  items-center gap-2 bg-graphite-20 px-1 py-0.5 rounded data-[highlighted]:bg-graphite-30 data-[highlighted]:outline-none cursor-pointer ',
                className
            )}
            ref={forwardedRef}
        >
            <RadixSelect.SelectItemText>{children}</RadixSelect.SelectItemText>
            <RadixSelect.ItemIndicator>
                <CheckIcon className=" text-violet-50" />
            </RadixSelect.ItemIndicator>
        </RadixSelect.Item>
    );
});

export const Select = Object.assign(RadixSelect.Root, {
    Primitive: RadixSelect,
    Option: SelectOption,
    Options: SelectOptions,
    Button: SelectButton,
});
