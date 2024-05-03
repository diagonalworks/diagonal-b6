import { useScenarioContext } from '@/lib/context/scenario';
import { highlighted } from '@/lib/text';
import { Combobox } from '@headlessui/react';
import { AnimatePresence, motion } from 'framer-motion';
import { isUndefined } from 'lodash';
import { StyleSpecification } from 'maplibre-gl';
import { QuickScore } from 'quick-score';
import { HTMLAttributes, useMemo, useState } from 'react';
import { useHotkeys } from 'react-hotkeys-hook';
import { twMerge } from 'tailwind-merge';
import { OutlinersLayer } from './Outliners';
import { ScenarioMap } from './ScenarioMap';
import { WorldShellAdapter } from './adapters/ShellAdapter';

export const ScenarioTab = ({
    id,
    mapStyle,
    tab,
    ...props
}: {
    id: string;
    mapStyle: StyleSpecification;
    tab: 'left' | 'right';
} & HTMLAttributes<HTMLDivElement>) => {
    const [showWorldShell, setShowWorldShell] = useState(false);
    const { change } = useScenarioContext();

    useHotkeys('shift+meta+b, `', () => {
        setShowWorldShell((prev) => !prev);
    });

    return (
        <div
            {...props}
            className={twMerge(
                'h-full border border-x-graphite-40 border-t-graphite-40 border-t bg-graphite-30',
                tab === 'right' &&
                    'border-x-orange-40 border-t-orange-40 bg-orange-30',
                props.className
            )}
        >
            <div
                className={twMerge(
                    'h-full w-full relative border-2 border-graphite-30 rounded',
                    tab === 'right' && 'border-orange-30'
                )}
            >
                <ScenarioMap>
                    <GlobalShell show={showWorldShell} mapId={id} />
                    <OutlinersLayer />
                </ScenarioMap>
                {isUndefined(change) && id !== 'baseline' && (
                    <div className="absolute top-0 left-0 ">
                        <ChangePanel />
                    </div>
                )}
            </div>
        </div>
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
    const [selectedFunction, setSelectedFunction] = useState<
        (typeof CHANGES)[number] | undefined
    >();
    const [search, setSearch] = useState('');

    const functionResults = useMemo(() => {
        if (!search) return [];
        return matcher.search(search);
    }, [search]);

    return (
        <div className="border shadow bg-orange-20 px-0.5 border-orange-30 w-60">
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
            {!isUndefined(selectedFunction) && (
                <form className="flex flex-col gap-4 py-2">
                    <div className="flex flex-col gap-1">
                        <span className="ml-2 text-xs text-orange-90">Add</span>
                        <input className="text-sm" />
                    </div>
                    <div className="flex flex-col gap-2">
                        <span className="ml-2 text-xs text-orange-90">To</span>
                        <input className="text-sm" />
                    </div>
                </form>
            )}
        </div>
    );
};
