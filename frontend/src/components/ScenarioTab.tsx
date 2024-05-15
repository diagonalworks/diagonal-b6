import { startupQueryAtom } from '@/atoms/startup';
import { useAppContext } from '@/lib/context/app';
import { OutlinerProvider } from '@/lib/context/outliner';
import { useScenarioContext } from '@/lib/context/scenario';
import { highlighted } from '@/lib/text';
import { $FixMe } from '@/utils/defs';
import { Combobox } from '@headlessui/react';
import {
    ChevronDownIcon,
    Cross1Icon,
    TriangleRightIcon,
} from '@radix-ui/react-icons';
import * as Select from '@radix-ui/react-select';
import { AnimatePresence, motion } from 'framer-motion';
import { useAtomValue } from 'jotai';
import { isEqual } from 'lodash';
import { QuickScore } from 'quick-score';
import {
    HTMLAttributes,
    useCallback,
    useEffect,
    useMemo,
    useState,
} from 'react';
import { useHotkeys } from 'react-hotkeys-hook';
import { twMerge } from 'tailwind-merge';
import { match } from 'ts-pattern';
import {
    LeftComparatorTeleporter,
    RightComparatorTeleporter,
} from './Comparator';
import { OutlinersLayer } from './Outliners';
import { ScenarioMap } from './ScenarioMap';
import { HeaderAdapter } from './adapters/HeaderAdapter';
import { LabelledIconAdapter } from './adapters/LabelledIconAdapter';
import { WorldShellAdapter } from './adapters/ShellAdapter';
import { StackAdapter } from './adapters/StackAdapter';
import { Line } from './system/Line';

export const ScenarioTab = ({
    id,
    tab,
    ...props
}: {
    id: string;
    tab: 'left' | 'right';
} & HTMLAttributes<HTMLDivElement>) => {
    const { activeComparator } = useAppContext();
    const [showWorldShell, setShowWorldShell] = useState(false);
    const { comparisonOutliners } = useScenarioContext();

    useHotkeys('shift+meta+b, `', () => {
        setShowWorldShell((prev) => !prev);
    });

    const showComparator =
        activeComparator?.request?.scenarios?.includes(id as $FixMe) ||
        activeComparator?.request?.baseline === (id as $FixMe);

    const Teleporter = useMemo(() => {
        return match(tab)
            .with('left', () => LeftComparatorTeleporter)
            .with('right', () => RightComparatorTeleporter)
            .exhaustive();
    }, [tab]);

    return (
        <>
            <div
                {...props}
                className={twMerge(
                    'h-full border border-x-graphite-40 border-t-graphite-40 border-t bg-graphite-30',
                    tab === 'right' &&
                        'border-x-rose-40 border-t-rose-40 bg-rose-30',
                    props.className
                )}
            >
                <div
                    className={twMerge(
                        'h-full w-full relative border-2 border-graphite-30 rounded-lg',
                        tab === 'right' && 'border-rose-30'
                    )}
                >
                    <ScenarioMap>
                        <GlobalShell show={showWorldShell} mapId={id} />
                        <OutlinersLayer />
                    </ScenarioMap>

                    <div className="absolute top-0 left-0 ">
                        {id !== 'baseline' && <ChangePanel />}
                    </div>
                </div>
            </div>
            {showComparator && (
                <Teleporter.Source>
                    {comparisonOutliners.map((outliner) => (
                        <OutlinerProvider key={outliner.id} outliner={outliner}>
                            <StackAdapter />
                        </OutlinerProvider>
                    ))}
                </Teleporter.Source>
            )}
        </>
    );
};

const GlobalShell = ({ show, mapId }: { show: boolean; mapId: string }) => {
    return (
        <AnimatePresence>
            {show && (
                <motion.div
                    initial={{
                        translateX: -100,
                    }}
                    animate={{
                        translateX: 0,
                    }}
                    className="absolute top-2 left-10 w-[95%] z-20 "
                >
                    <WorldShellAdapter mapId={mapId} />
                </motion.div>
            )}
        </AnimatePresence>
    );
};

const ChangePanel = () => {
    const {
        removeFeatureFromChange,
        queryScenario,
        scenario: { change },
    } = useScenarioContext();

    const [open, setOpen] = useState(true);

    const worldCreated = useMemo(() => {
        return queryScenario?.isSuccess;
    }, [queryScenario?.isSuccess]);

    useEffect(() => {
        if (worldCreated) {
            setOpen(false);
        }
    }, [worldCreated]);

    return (
        <div className="border  bg-rose-30 p-0.5  border-rose-40  shadow-lg">
            <div className="bg-rose-30 flex flex-col gap-2">
                {worldCreated && (
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
                            {change?.features && change.features.length > 0 ? (
                                change.features.map((feature, i) => (
                                    <Line
                                        className={twMerge(
                                            'text-sm py-1 flex gap-2 items-center justify-between',
                                            worldCreated &&
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
                                        {!worldCreated && (
                                            <button
                                                className="text-xs hover:bg-graphite-20 p-1 hover:text-graphite-100 text-graphite-70 rounded-full w-5 h-5 flex items-center justify-center"
                                                onClick={() =>
                                                    removeFeatureFromChange(
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
                        {change?.features && change.features.length > 0 && (
                            <ChangeCombo />
                        )}
                    </div>
                )}
            </div>
        </div>
    );
};

const ChangeCombo = () => {
    const { addComparator } = useAppContext();
    const { changes } = useAppContext();
    const {
        scenario: { id, change },
        setChangeAnalysis,
        setChangeFunction,
        queryScenario,
    } = useScenarioContext();
    const startupQuery = useAtomValue(startupQueryAtom);

    const [search, setSearch] = useState('');

    const worldCreated = useMemo(() => {
        return queryScenario?.isSuccess;
    }, [queryScenario?.isSuccess]);

    const matcher = useMemo(() => {
        return new QuickScore(changes, ['label']);
    }, [changes]);

    const functionResults = useMemo(() => {
        if (!search) return [];
        return matcher.search(search);
    }, [search, matcher]);

    const handleClick = useCallback(() => {
        if (!change) return;
        queryScenario?.refetch();
        addComparator({
            baseline: 'baseline' as $FixMe,
            scenarios: [id] as $FixMe,
            analysis: 'test' as $FixMe,
        });
    }, [change?.analysis, addComparator, id]);

    const analysisOptions = useMemo(() => {
        const dockedAnalysis = startupQuery.data?.docked;
        return (
            dockedAnalysis?.flatMap((analysis: $FixMe) => {
                const label = analysis.proto.stack?.substacks[0].lines.map(
                    (l: $FixMe) => l.header
                )[0];

                return {
                    node: analysis.proto.node,
                    label,
                };
            }) ?? []
        );
    }, [startupQuery.data?.docked]);

    const selectedAnalysis = useMemo(() => {
        return analysisOptions.find((analysis: $FixMe) =>
            isEqual(change?.analysis, analysis.node)
        );
    }, [change?.analysis, analysisOptions]);
    const selectedAnalysisLabel = selectedAnalysis?.label?.title?.value ?? '';

    return (
        <div className="flex flex-col gap-2 ">
            <div>
                <span className="ml-2 text-xs text-rose-90">Change</span>
                <Combobox
                    disabled={worldCreated}
                    value={change?.changeFunction?.label}
                    onChange={(v) => {
                        const option = functionResults.find(
                            (f) => f.item.label === v
                        );
                        if (!option) return;
                        setChangeFunction(option.item);
                    }}
                >
                    <div
                        className={twMerge(
                            'w-full text-sm flex gap-2 bg-white focus-within:outline-none focus-within:ring-2 focus-within:ring-rose-60/40 hover:bg-rose-10 py-2.5 px-2',
                            worldCreated && 'italic bg-rose-10'
                        )}
                    >
                        <span className={twMerge('text-rose-80')}> b6</span>

                        <Combobox.Input
                            onChange={(e) => setSearch(e.target.value)}
                            placeholder="define the change"
                            className={twMerge(
                                'relative flex-grow bg-transparent text-graphite-70 focus:outline-none w-full',
                                worldCreated && 'italic text-graphite-100'
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
            {change?.changeFunction && (
                <div>
                    <span className="ml-2 text-xs text-rose-90">Analysis</span>

                    <Select.Root
                        disabled={worldCreated}
                        value={selectedAnalysisLabel}
                        onValueChange={(v) => {
                            const option = analysisOptions.find(
                                (analysis: $FixMe) =>
                                    analysis.label?.title?.value === v
                            );
                            if (!option?.node) return;
                            setChangeAnalysis(option.node);
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
                            {!worldCreated && (
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
                                    (analysis: $FixMe, i: number) => (
                                        <Select.Item
                                            key={i}
                                            value={
                                                analysis.label?.title?.value ??
                                                ''
                                            }
                                            className=" cursor-pointer data-[state=checked]:bg-rose-10  data-[highlighted]:bg-rose-10 text-sm py-3 px-2 border-x border-b border-graphite-20 first:border-t items-center focus:outline-none "
                                        >
                                            {analysis?.label?.title?.value}
                                        </Select.Item>
                                    )
                                )}
                            </Select.Viewport>
                        </Select.Content>
                    </Select.Root>
                </div>
            )}
            {change?.changeFunction && !worldCreated && (
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
