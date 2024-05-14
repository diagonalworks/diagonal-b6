import { useAppContext } from '@/lib/context/app';
import { OutlinerProvider } from '@/lib/context/outliner';
import { useScenarioContext } from '@/lib/context/scenario';
import { highlighted } from '@/lib/text';
import { $FixMe } from '@/utils/defs';
import { Combobox } from '@headlessui/react';
import { Cross1Icon, TriangleRightIcon } from '@radix-ui/react-icons';
import { AnimatePresence, motion } from 'framer-motion';
import { QuickScore } from 'quick-score';
import { HTMLAttributes, useCallback, useMemo, useState } from 'react';
import { useHotkeys } from 'react-hotkeys-hook';
import { twMerge } from 'tailwind-merge';
import { match } from 'ts-pattern';
import {
    LeftComparatorTeleporter,
    RightComparatorTeleporter,
} from './Comparator';
import { OutlinersLayer } from './Outliners';
import { ScenarioMap } from './ScenarioMap';
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
    const { isDefiningChange, comparisonOutliners } = useScenarioContext();

    useHotkeys('shift+meta+b, `', () => {
        setShowWorldShell((prev) => !prev);
    });

    const showComparator =
        activeComparator?.request?.scenarios.includes(id as $FixMe) ||
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
                    {isDefiningChange && (
                        <div className="absolute top-0 left-0 ">
                            <ChangePanel />
                        </div>
                    )}
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

const CHANGES = ['add-service', 'change-use'];
const matcher = new QuickScore(CHANGES);

const ChangePanel = () => {
    const {
        removeFeatureFromChange,
        scenario: { change },
    } = useScenarioContext();

    return (
        <div className="border  bg-rose-30 p-0.5  border-rose-40  shadow-lg">
            <div className="bg-rose-30 flex flex-col gap-2">
                <div>
                    {change.features.length > 0 ? (
                        change.features.map((feature, i) => (
                            <Line
                                className="text-sm py-1 flex gap-2 items-center justify-between"
                                key={i}
                            >
                                {feature.label ? (
                                    <LabelledIconAdapter
                                        labelledIcon={feature.label}
                                    />
                                ) : (
                                    <span>{feature.expression}</span>
                                )}
                                <button
                                    className="text-xs hover:bg-graphite-20 p-1 hover:text-graphite-100 text-graphite-70 rounded-full w-5 h-5 flex items-center justify-center"
                                    onClick={() =>
                                        removeFeatureFromChange(feature)
                                    }
                                >
                                    <Cross1Icon />
                                </button>
                            </Line>
                        ))
                    ) : (
                        <div className=" text-graphite-90 italic text-xs py-2 px-3 ">
                            Click on a feature to add it to the change
                        </div>
                    )}
                </div>
                {change.features.length > 0 && <ChangeCombo />}
            </div>
        </div>
    );
};

const ChangeCombo = () => {
    const { addComparator } = useAppContext();
    const {
        scenario: { id },
    } = useScenarioContext();
    const [selectedFunction, setSelectedFunction] = useState<
        (typeof CHANGES)[number] | undefined
    >();
    const [search, setSearch] = useState('');

    const functionResults = useMemo(() => {
        if (!search) return [];
        return matcher.search(search);
    }, [search]);

    const handleClick = useCallback(() => {
        if (!selectedFunction) return;
        addComparator({
            baseline: 'baseline' as $FixMe,
            scenarios: [id] as $FixMe,
            analysis: 'test' as $FixMe,
        });
    }, [selectedFunction, addComparator, id]);

    return (
        <div className="flex flex-col gap-2">
            <span className="ml-2 text-xs text-rose-90">Change</span>
            <Combobox value={selectedFunction} onChange={setSelectedFunction}>
                <div className="w-full text-sm flex gap-2 bg-white hover:bg-ultramarine-10 py-2.5 px-2">
                    <span className="text-ultramarine-70 "> b6</span>

                    <Combobox.Input
                        onChange={(e) => setSearch(e.target.value)}
                        placeholder="define the change"
                        className=" relative flex-grow bg-transparent text-graphite-70 focus:outline-none w-full"
                    />
                </div>
                <Combobox.Options className="max-h-64 overflow-y-auto border-b border-b-graphite-30 ">
                    {functionResults.map((result) => (
                        <Combobox.Option
                            value={result.item}
                            key={result.item}
                            className=" bg-white py-2 px-1 text-sm  ui-active:bg-ultramarine-10 ui-active:border-l ui-active:border-l-ultramarine-60 last:border-b-0 "
                        >
                            {highlighted(result.item, result.matches)}
                        </Combobox.Option>
                    ))}
                </Combobox.Options>
            </Combobox>
            {selectedFunction && (
                <button
                    className="w-full  text-sm flex gap-1 items-center py-2 justify-center rounded hover:bg-rose-10 bg-rose-20  text-rose-80"
                    onClick={handleClick}
                >
                    Apply change
                    <TriangleRightIcon className="h-5 w-5" />
                </button>
            )}
        </div>
    );
};
