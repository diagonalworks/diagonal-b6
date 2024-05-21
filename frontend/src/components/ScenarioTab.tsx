import { useAppContext } from '@/lib/context/app';
import { OutlinerProvider } from '@/lib/context/outliner';
import { useScenarioContext } from '@/lib/context/scenario';
import { $FixMe } from '@/utils/defs';
import { AnimatePresence, motion } from 'framer-motion';
import { HTMLAttributes, useMemo, useState } from 'react';
import { useHotkeys } from 'react-hotkeys-hook';
import { twMerge } from 'tailwind-merge';
import { match } from 'ts-pattern';
import { ChangePanel } from './ChangePanel';
import {
    LeftComparatorTeleporter,
    RightComparatorTeleporter,
} from './Comparator';
import { OutlinersLayer } from './Outliners';
import { ScenarioMap } from './ScenarioMap';
import { WorldShellAdapter } from './adapters/ShellAdapter';
import { StackAdapter } from './adapters/StackAdapter';

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
        activeComparator?.scenarios?.includes(id as $FixMe) ||
        activeComparator?.baseline === (id as $FixMe);

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
