import { HeaderAdapter } from '@/components/adapters/HeaderAdapter';
import { LabelledIconAdapter } from '@/components/adapters/LabelledIconAdapter';
import { Line } from '@/components/system/Line';
import { highlighted } from '@/lib/text';
import { World, useWorldStore } from '@/stores/worlds';
import { $FixMe } from '@/utils/defs';
import { Combobox } from '@headlessui/react';
import {
    ChevronDownIcon,
    Cross1Icon,
    TriangleRightIcon,
} from '@radix-ui/react-icons';
import * as Select from '@radix-ui/react-select';
import { isEqual } from 'lodash';
import { QuickScore } from 'quick-score';
import { useCallback, useEffect, useMemo, useState } from 'react';
import { twMerge } from 'tailwind-merge';
import { useChangeFunctions } from '../api/change-functions';
import { useScenario } from '../api/run-scenario';
import useDockedAnalysis from '../hooks/useDockedAnalysis';
import { Change, useChangesStore } from '../stores/changes';

export const ChangePanel = ({
    world,
    id,
}: {
    world: World['id'];
    id: Change['id'];
}) => {
    const change = useChangesStore((state) => state.changes[id]);
    const changeActions = useChangesStore((state) => state.actions);

    const [open, setOpen] = useState(true);

    useEffect(() => {
        if (change.created) {
            setOpen(false);
        }
    }, [change.created]);

    useEffect(() => {
        if (!change?.spec.features) {
            setOpen(true);
        }
    }, [change]);

    return (
        <div className="border  bg-rose-30 p-0.5  border-rose-40  shadow-lg">
            <div className="bg-rose-30 flex flex-col gap-2">
                {change.created && (
                    <div>
                        <button
                            className="px-2 text-sm flex gap-1 items-center py-1  rounded hover:underline  text-rose-80 focus-within:outline-none"
                            onClick={() => setOpen((prev) => !prev)}
                        >
                            Scenario Details
                            <ChevronDownIcon
                                className={twMerge(
                                    'h-4 w-4',
                                    open && 'transform rotate-180'
                                )}
                            />
                        </button>
                    </div>
                )}
                {open && (
                    <div>
                        <div className="flex flex-col gap-2">
                            {change?.spec.features &&
                            change.spec.features.length > 0 ? (
                                change.spec.features.map((feature, i) => (
                                    <Line
                                        className={twMerge(
                                            'text-sm py-1 flex gap-2 items-center justify-between',
                                            change.created &&
                                                'bg-rose-10 hover:bg-rose-10 italic'
                                        )}
                                        key={i}
                                    >
                                        {feature.label ? (
                                            <LabelledIconAdapter
                                                labelledIcon={feature.label}
                                            />
                                        ) : (
                                            <span>{feature.expression}</span>
                                        )}
                                        {!change.created && (
                                            <button
                                                className="text-xs hover:bg-graphite-20 p-1 hover:text-graphite-100 text-graphite-70 rounded-full w-5 h-5 flex items-center justify-center"
                                                onClick={() =>
                                                    changeActions.removeFeature(
                                                        id,
                                                        feature
                                                    )
                                                }
                                            >
                                                <Cross1Icon />
                                            </button>
                                        )}
                                    </Line>
                                ))
                            ) : (
                                <div className=" text-graphite-90 italic text-xs py-2 px-3 ">
                                    Click on a feature to add it to the change
                                </div>
                            )}
                        </div>
                        {change.spec.features &&
                            change.spec.features.length > 0 && (
                                <ChangeCombo worldId={world} id={id} />
                            )}
                    </div>
                )}
            </div>
        </div>
    );
};

const ChangeCombo = ({
    worldId,
    id,
}: {
    worldId: World['id'];
    id: Change['id'];
}) => {
    const change = useChangesStore((state) => state.changes[id]);
    const world = useWorldStore((state) => state.worlds[worldId]);
    const changeActions = useChangesStore((state) => state.actions);

    const baseline = useWorldStore((state) => state.worlds.baseline);
    const changeFunctions = useChangeFunctions({
        origin: baseline,
    });

    const queryScenario = useScenario(baseline, world, change.spec);

    const analysisOptions = useDockedAnalysis();

    const [search, setSearch] = useState('');

    const matcher = useMemo(() => {
        return new QuickScore(changeFunctions, ['label']);
    }, [changeFunctions]);

    const functionResults = useMemo(() => {
        if (!search) return [];
        return matcher.search(search);
    }, [search, matcher]);

    const handleClick = useCallback(() => {
        if (!change) return;
        queryScenario.refetch();
    }, [change, queryScenario]);

    const selectedAnalysis = useMemo(() => {
        return analysisOptions.find((analysis) =>
            isEqual(change?.spec.analysis, analysis.id)
        );
    }, [change.spec.analysis, analysisOptions]);

    return (
        <div className="flex flex-col gap-2 ">
            <div>
                <span className="ml-2 text-xs text-rose-90">Change</span>
                <Combobox
                    disabled={change.created}
                    value={change?.spec.changeFunction?.label}
                    onChange={(v) => {
                        const option = functionResults.find(
                            (f) => f.item.label === v
                        );
                        if (!option) return;
                        changeActions.setFunction(id, option.item);
                    }}
                >
                    <div
                        className={twMerge(
                            'w-full text-sm flex gap-2 bg-white focus-within:outline-none focus-within:ring-2 focus-within:ring-rose-60/40 hover:bg-rose-10 py-2.5 px-2',
                            change.created && 'italic bg-rose-10'
                        )}
                    >
                        <span className={twMerge('text-rose-80')}> b6</span>

                        <Combobox.Input
                            onChange={(e) => setSearch(e.target.value)}
                            placeholder="define the change"
                            className={twMerge(
                                'relative flex-grow bg-transparent text-graphite-70 focus:outline-none w-full',
                                change.created && 'italic text-graphite-100'
                            )}
                        />
                    </div>
                    <Combobox.Options className="max-h-64 overflow-y-auto border border-graphite-20 ">
                        {functionResults.map((result) => (
                            <Combobox.Option
                                value={result.item.label}
                                key={result.item.id.value}
                                className=" bg-white py-3 px-2 text-sm border-graphite-20 border-x border-b first:border-t  ui-active:bg-rose-10 last:border-b-0 "
                            >
                                {highlighted(
                                    result.item?.label ?? '',
                                    result.matches.label
                                )}
                            </Combobox.Option>
                        ))}
                    </Combobox.Options>
                </Combobox>
            </div>
            {change?.spec.changeFunction && (
                <div>
                    <span className="ml-2 text-xs text-rose-90">Analysis</span>

                    <Select.Root
                        disabled={change.created}
                        value={change.spec.analysis?.value?.toString()}
                        onValueChange={(v) => {
                            const option = analysisOptions.find(
                                (analysis) =>
                                    analysis.id.value?.toString() === v
                            );
                            if (!option?.id) return;
                            changeActions.setAnalysis(id, option.id);
                        }}
                    >
                        <Select.Trigger className=" bg-white text-graphite-70 h-10 py-2 px-2 text-sm inline-flex items-center justify-between w-full focus-within:outline-none focus-within:ring-2 focus-within:ring-rose-60/40 data-[disabled]:bg-rose-10 data-[disabled]:italic">
                            <Select.Value placeholder="Select an analysis...">
                                {selectedAnalysis?.label && (
                                    <HeaderAdapter
                                        header={selectedAnalysis.label}
                                    />
                                )}
                            </Select.Value>
                            {!change.created && (
                                <Select.Icon>
                                    <ChevronDownIcon />
                                </Select.Icon>
                            )}
                        </Select.Trigger>

                        <Select.Content
                            position="popper"
                            className=" bg-white rounded z-90"
                            style={{
                                width: 'var(--radix-select-trigger-width)',
                            }}
                        >
                            <Select.Viewport>
                                {analysisOptions.map(
                                    (analysis: $FixMe, i: number) => {
                                        // @TODO: temporary hack to remove bold formatting markers in for the select options. This bold formatting should be handled further up.
                                        const label =
                                            analysis.label?.title?.value;
                                        if (!label) return null;
                                        const boldWorld =
                                            label.match(/_((?:[a-zA-Z]|\d)*)_/);
                                        const labelNoBoldMarkers =
                                            label.replaceAll(
                                                boldWorld[0],
                                                boldWorld[1]
                                            );

                                        return (
                                            <Select.Item
                                                key={i}
                                                value={analysis.id.value?.toString()}
                                                className=" cursor-pointer data-[state=checked]:bg-rose-10  data-[highlighted]:bg-rose-10 text-sm py-3 px-2 border-x border-b border-graphite-20 first:border-t items-center focus:outline-none "
                                            >
                                                {labelNoBoldMarkers}
                                            </Select.Item>
                                        );
                                    }
                                )}
                            </Select.Viewport>
                        </Select.Content>
                    </Select.Root>
                </div>
            )}
            {change?.spec.changeFunction && !change.created && (
                <button
                    disabled={queryScenario?.isFetching}
                    className="w-full  text-sm flex gap-1 items-center py-2 justify-center rounded hover:bg-rose-10 bg-rose-20  text-rose-80 focus-within:outline-none focus-within:ring-2 focus-within:ring-rose-60/40  "
                    onClick={handleClick}
                >
                    {queryScenario?.isFetching ? (
                        <div className="loader-scenario w-5 h-5" />
                    ) : (
                        <>
                            Run Scenario
                            <TriangleRightIcon className="h-5 w-5" />
                        </>
                    )}
                </button>
            )}
        </div>
    );
};
