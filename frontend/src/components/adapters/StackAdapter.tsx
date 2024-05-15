import { findAtoms } from '@/lib/atoms';
import { useAppContext } from '@/lib/context/app';
import { useOutlinerContext } from '@/lib/context/outliner';
import { useScenarioContext } from '@/lib/context/scenario';
import { MinusIcon, PlusIcon } from '@radix-ui/react-icons';
import { isEqual } from 'lodash';
import { useState } from 'react';
import { twMerge } from 'tailwind-merge';
import { Line } from '../system/Line';
import { Stack } from '../system/Stack';
import { SubstackAdapter } from './SubstackAdapter';

export const StackAdapter = () => {
    const {
        app: { outliners },
        setActiveOutliner,
    } = useAppContext();
    const { outliner } = useOutlinerContext();
    const {
        scenario: { change, submitted },
        addFeatureToChange,
        removeFeatureFromChange,
    } = useScenarioContext();
    const [open, setOpen] = useState(outliner.properties.docked ? false : true);

    const handleOpenChange = (open: boolean) => {
        setActiveOutliner(outliner.id, open);

        setOpen(open);
    };

    const originOutliner = outliner.properties?.origin
        ? outliners?.[outliner.properties.origin]
        : null;

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

    const firstSubstack = outliner.data.proto.stack?.substacks?.[0];
    const otherSubstacks = outliner.data.proto.stack?.substacks?.slice(1);

    const originFirstSubstack =
        originOutliner?.data?.proto.stack?.substacks?.[0];
    const originOtherSubstacks =
        originOutliner?.data?.proto.stack?.substacks?.slice(1);

    const featureNode = outliner.data?.proto?.node;

    const isInChange = change?.features?.find((f) => isEqual(f, featureNode));
    const showChangeElements = !submitted && outliner.properties.changeable;

    const labelledIcon =
        outliner.data.proto.stack?.substacks?.[1]?.lines?.flatMap((l) =>
            findAtoms(l, 'labelledIcon')
        )?.[0]?.labelledIcon;

    return (
        <>
            {showChangeElements && (
                <div className="flex justify-start">
                    <button
                        onClick={() => {
                            const node = outliner.data?.proto.node;
                            const expression =
                                outliner.request?.expression ?? '';
                            if (!node) return;

                            if (isInChange) {
                                removeFeatureFromChange({
                                    expression,
                                    node,
                                    label: labelledIcon,
                                });
                            }
                            addFeatureToChange({
                                expression,
                                node,
                                label: labelledIcon,
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
                    onOpenChange={handleOpenChange}
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
                                origin={originFirstSubstack}
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
                                        origin={originOtherSubstacks?.[i]}
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
