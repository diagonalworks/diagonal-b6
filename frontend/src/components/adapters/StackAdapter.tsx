import { useOutlinerContext } from '@/lib/context/outliner';
import { useScenarioContext } from '@/lib/context/scenario';
import { MinusIcon, PlusIcon } from '@radix-ui/react-icons';
import { useState } from 'react';
import { twMerge } from 'tailwind-merge';
import { Line } from '../system/Line';
import { Stack } from '../system/Stack';
import { SubstackAdapter } from './SubstackAdapter';

export const StackAdapter = () => {
    const { outliner } = useOutlinerContext();
    const { setChange, change } = useScenarioContext();
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

    console.log(change);
    const isInChange = change.features.includes(
        outliner.request?.expression ?? ''
    );

    return (
        <>
            {outliner.properties.changeable && (
                <div className="flex justify-start">
                    <button
                        onClick={() => {
                            const features = isInChange
                                ? change.features.filter(
                                      (f) => f !== outliner.request?.expression
                                  )
                                : [
                                      ...change.features,
                                      outliner.request?.expression ?? '',
                                  ];
                            setChange({
                                ...change,
                                features,
                            });
                        }}
                        className="-mb-[2px] p-2 flex gap-1  items-center text-xs  text-orange-90 rounded-t border-b-0 bg-orange-40 hover:bg-orange-30 border border-orange-50"
                    >
                        {isInChange ? (
                            <>
                                <MinusIcon /> remove from change
                            </>
                        ) : (
                            <>
                                <PlusIcon /> add to change
                            </>
                        )}
                    </button>
                </div>
            )}
            <div
                className={twMerge(
                    'stack-wrapper',
                    outliner.properties.changeable &&
                        ' border border-orange-50 rounded'
                )}
            >
                <Stack
                    collapsible={outliner.properties.docked}
                    open={open}
                    onOpenChange={setOpen}
                    className={twMerge(
                        outliner.properties.changeable &&
                            'border-2 border-orange-40'
                    )}
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
            </div>
        </>
    );
};
