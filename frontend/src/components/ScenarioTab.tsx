import { useScenarioContext } from '@/lib/context/scenario';
import { AnimatePresence, motion } from 'framer-motion';
import { isUndefined } from 'lodash';
import { StyleSpecification } from 'maplibre-gl';
import { HTMLAttributes, useState } from 'react';
import { useHotkeys } from 'react-hotkeys-hook';
import { twMerge } from 'tailwind-merge';
import { OutlinersLayer } from './Outliners';
import { ScenarioMap } from './ScenarioMap';
import { WorldShellAdapter } from './adapters/ShellAdapter';

export const ScenarioTab = ({
    id,
    mapStyle,
    ...props
}: {
    id: string;
    mapStyle: StyleSpecification;
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
                'h-full border border-x-graphite-40 border-t-graphite-20 relative',
                props.className
            )}
        >
            <ScenarioMap>
                <GlobalShell show={showWorldShell} mapId={id} />
                <OutlinersLayer />
            </ScenarioMap>
            {isUndefined(change) && id !== 'baseline' && (
                <div className="absolute top-0 left-0 border shadow bg-orange-20 px-0.5 border-orange-30 w-56">
                    <form className="flex flex-col gap-4 py-2">
                        <div className="flex flex-col gap-1">
                            <span className="ml-2 text-xs text-orange-90">
                                Add
                            </span>
                            <input className="text-sm" />
                        </div>
                        <div className="flex flex-col gap-2">
                            <span className="ml-2 text-xs text-orange-90">
                                To
                            </span>
                            <input className="text-sm" />
                        </div>
                    </form>
                </div>
            )}
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
