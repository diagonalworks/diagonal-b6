import { useOutlinerContext } from '@/lib/context/outliner';
import { useState } from 'react';
import { Line } from '../system/Line';
import { Stack } from '../system/Stack';
import { SubstackAdapter } from './SubstackAdapter';

export const StackAdapter = () => {
    const { outliner } = useOutlinerContext();
    const [open, setOpen] = useState(outliner.properties.docked ? false : true);

    if (outliner.query?.isLoading) {
        return (
            <Stack>
                <Stack.Trigger>
                    <Line className="flex flex-nowrap ">
                        <div className="loader shrink-0" />
                        <div className="text-graphite-60 italic text-nowrap overflow-hidden overflow-ellipsis">
                            {outliner.request?.expression}
                        </div>
                    </Line>
                </Stack.Trigger>
            </Stack>
        );
    }

    if (!outliner.data) return null;

    const firstSubstack = outliner.data.proto.stack?.substacks[0];
    const otherSubstacks = outliner.data.proto.stack?.substacks.slice(1);

    return (
        <Stack
            collapsible={outliner.properties.docked}
            open={open}
            onOpenChange={setOpen}
        >
            {firstSubstack && (
                <Stack.Trigger>
                    <SubstackAdapter
                        substack={firstSubstack}
                        collapsible={firstSubstack.collapsable}
                    />
                </Stack.Trigger>
            )}
            {otherSubstacks && (
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
            )}
        </Stack>
    );
};
