import { AnimatePresence, motion } from 'framer-motion';
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

    useHotkeys('shift+meta+b, `', () => {
        setShowWorldShell((prev) => !prev);
    });

    return (
        <div
            {...props}
            className={twMerge(
                'h-full  border  border-x-graphite-40 border-t-graphite-20 relative',
                props.className
            )}
        >
            <ScenarioMap id={id} mapStyle={mapStyle}>
                <GlobalShell show={showWorldShell} mapId={id} />
                <OutlinersLayer mapId={id} />
            </ScenarioMap>
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
