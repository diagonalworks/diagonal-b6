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
    const { setChange, change, isDefiningChange } = useScenarioContext();
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

    const expression = outliner.data?.proto?.expression;

    const isInChange = change.features.includes(expression ?? '');
    const showChangeElements =
        isDefiningChange && outliner.properties.changeable;

    return (
        <>
            {showChangeElements && (
                <div className="flex justify-start">
                    <button
                        onClick={() => {
                            const features = isInChange
                                ? change.features.filter(
                                      (f) => f !== expression
                                  )
                                : [...change.features, expression ?? ''];
                            setChange({
                                ...change,
                                features,
                            });
                        }}
                        className="-mb-[2px] p-2 flex gap-1  items-center text-xs  text-rose-90 rounded-t border-b-0 bg-rose-40 hover:bg-rose-30 border border-rose-50"
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
                    showChangeElements && ' border border-rose-50 rounded'
                )}
            >
                <Stack
                    collapsible={outliner.properties.docked}
                    open={open}
                    onOpenChange={setOpen}
                    className={twMerge(
                        showChangeElements && 'border-2 border-rose-40'
                    )}
                >
                    {firstSubstack && (
                        <Stack.Trigger>
                            <SubstackAdapter
                                substack={firstSubstack}
                                collapsible={firstSubstack.collapsable}
                                close={!outliner.properties.docked}
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
