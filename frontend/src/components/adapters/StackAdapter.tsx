import { startupQueryAtom } from '@/atoms/startup';
import { findAtoms } from '@/lib/atoms';
import { useAppContext } from '@/lib/context/app';
import { useOutlinerContext } from '@/lib/context/outliner';
import { useScenarioContext } from '@/lib/context/scenario';
import { MinusIcon, PlusIcon } from '@radix-ui/react-icons';
import { useAtomValue } from 'jotai';
import { isEqual } from 'lodash';
import { useMemo, useState } from 'react';
import { twMerge } from 'tailwind-merge';
import { Line } from '../system/Line';
import { Stack } from '../system/Stack';
import { SubstackAdapter } from './SubstackAdapter';

const getFirstWord = (s?: string) => {
    if (!s) return '';
    return s.match(/^\s*(\w+)/)?.[0]?.toLocaleLowerCase();
};

export const StackAdapter = () => {
    const {
        app: { outliners },
    } = useAppContext();
    const { outliner, setVisible } = useOutlinerContext();
    const {
        scenario: { worldCreated, change },
        addFeatureToChange,
        removeFeatureFromChange,
    } = useScenarioContext();
    const [open, setOpen] = useState(outliner.properties.docked ? false : true);
    const startupQuery = useAtomValue(startupQueryAtom);
    const { activeComparator } = useAppContext();

    /** hack to get the analysis title for the histograms @TODO: get rid of this. */
    const analysis = useMemo(() => {
        return startupQuery.data?.docked?.find((d) => {
            return isEqual(d.proto.stack?.id, activeComparator?.analysis);
        });
    }, [startupQuery.data?.docked, activeComparator?.analysis]);

    let analysisTitle = analysis?.proto.stack?.substacks?.[0]?.lines?.map((l) =>
        getFirstWord(l.header?.title?.value)
    )[0];
    // hack stop

    const handleOpenChange = (open: boolean) => {
        setVisible(open);
        setOpen(open);
    };

    const originOutliner = outliner.properties?.origin
        ? outliners?.[outliner.properties.origin]
        : null;

    if (outliner.query?.isLoading) {
        return (
            <Stack>
                <Stack.Trigger>
                    <Line className="flex flex-nowrap w-80 ">
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

    const featureId = outliner.data?.proto?.stack?.id;

    const isInChange = change?.features?.find((f) => isEqual(f.id, featureId));
    const showChangeElements =
        featureId && !worldCreated && outliner.properties.changeable;

    const labelledIcon =
        outliner.data.proto.stack?.substacks?.[1]?.lines?.flatMap((l) =>
            findAtoms(l, 'labelledIcon')
        )?.[0]?.labelledIcon;

    // hack to get the analysis title for the histograms @TODO: get rid of this.
    const headerTitleString = firstSubstack?.lines?.[0].header?.title?.value;
    analysisTitle = headerTitleString
        ? getFirstWord(headerTitleString)
        : analysisTitle;
    // hack stop

    const queryLayers = outliner.query?.data?.proto.layers;
    const geoJsons =
        outliner.query?.data?.geoJSON || outliner.query?.data?.proto.geoJSON;

    return (
        <>
            {showChangeElements && (
                <div className="flex justify-start">
                    <button
                        onClick={() => {
                            const expression =
                                outliner.request?.expression ?? '';

                            if (isInChange) {
                                removeFeatureFromChange({
                                    expression,
                                    id: featureId,
                                    label: labelledIcon,
                                });
                            }
                            addFeatureToChange({
                                expression,
                                id: featureId,
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
                        <Stack.Trigger className="w-full">
                            <SubstackAdapter
                                substack={firstSubstack}
                                collapsible={firstSubstack.collapsable}
                                close={!outliner.properties.docked}
                                show={
                                    !outliner.properties.docked &&
                                    (!!queryLayers || !!geoJsons)
                                }
                                origin={originFirstSubstack}
                                analysisTitle={analysisTitle}
                            />
                        </Stack.Trigger>
                    )}
                    {otherSubstacks && (
                        <Stack.Content className="w-full">
                            {otherSubstacks.map((substack, i) => {
                                return (
                                    <SubstackAdapter
                                        key={i}
                                        substack={substack}
                                        collapsible={substack.collapsable}
                                        origin={originOtherSubstacks?.[i]}
                                        analysisTitle={analysisTitle}
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
