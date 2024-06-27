import * as DropdownMenu from '@radix-ui/react-dropdown-menu';
import { PlusIcon, ReaderIcon } from '@radix-ui/react-icons';
import React, { HTMLAttributes } from 'react';
import { twMerge } from 'tailwind-merge';

const EXAMPLE_SCENARIOS = [
    {
        name: 'Change bank to doctor',
    },
    {
        name: 'Change bank to nursery',
    },
    {
        name: 'Change bank to restaurant',
    },
];

export default function ScenariosDrawer({
    ...props
}: HTMLAttributes<HTMLDivElement>) {
    return (
        <div {...props} className={twMerge(props.className)}>
            <DropdownMenu.Root>
                <DropdownMenu.Trigger
                    aria-label="view scenarios"
                    className="text-sm flex gap-2 transition-all data-[state=open]:w-60 data-[state=open]:justify-end  mb-[1px] items-center bg-rose-10 rounded w-fit border border-b-0 hover:bg-rose-20 rounded-b-none border-rose-30 text-rose-60 px-2 py-1"
                >
                    scenarios
                </DropdownMenu.Trigger>
                <DropdownMenu.Portal>
                    <DropdownMenu.Content className=" bg-rose-30 p-1 flex flex-col gap-1">
                        {EXAMPLE_SCENARIOS.map((scenario) => (
                            <DropdownMenu.Item asChild key={scenario.name}>
                                <button className=" items-center w-60 flex gap-2 bg-rose-10 px-2 py-1 hover:outline-0 hover:bg-white rounded">
                                    <ReaderIcon />
                                    {scenario.name}
                                </button>
                            </DropdownMenu.Item>
                        ))}
                        <DropdownMenu.Item asChild>
                            <button className=" items-center w-60 flex gap-2 bg-rose-20 px-2 py-1 hover:outline-0 hover:bg-rose-10 rounded">
                                <PlusIcon />
                                New scenario
                            </button>
                        </DropdownMenu.Item>
                    </DropdownMenu.Content>
                </DropdownMenu.Portal>
            </DropdownMenu.Root>
        </div>
    );
}
