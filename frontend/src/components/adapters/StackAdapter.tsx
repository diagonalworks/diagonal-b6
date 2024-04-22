import { AppStore } from '@/atoms/app';
import { StackContextProvider } from '@/lib/context/stack';
import { useState } from 'react';
import { Stack } from '../system/Stack';
import { SubstackAdapter } from './SubstackAdapter';

export const StackAdapter = ({
    stack,
    docked = false,
    mapId,
}: {
    stack: AppStore['stacks'][string];
    docked?: boolean;
    mapId: string;
}) => {
    const [open, setOpen] = useState(docked ? false : true);
    if (!stack.proto.stack) return null;

    const firstSubstack = stack.proto.stack.substacks[0];
    const otherSubstacks = stack.proto.stack.substacks.slice(1);

    return (
        <div>
            <StackContextProvider stack={stack} mapId={mapId}>
                <Stack collapsible={docked} open={open} onOpenChange={setOpen}>
                    <Stack.Trigger>
                        <SubstackAdapter
                            substack={firstSubstack}
                            collapsible={firstSubstack.collapsable}
                        />
                    </Stack.Trigger>
                    <Stack.Content>
                        {otherSubstacks.map((substack, i) => {
                            return (
                                <SubstackAdapter
                                    key={i}
                                    substack={substack}
                                    collapsible={substack.collapsable}
                                />
                            );
                        })}
                    </Stack.Content>
                </Stack>
            </StackContextProvider>
        </div>
    );
};
