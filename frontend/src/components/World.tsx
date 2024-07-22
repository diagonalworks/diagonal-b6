import { AnimatePresence, motion } from 'framer-motion';
import { useState } from 'react';
import { useHotkeys } from 'react-hotkeys-hook';

import { Map } from '@/components/Map';
import { ChangePanel } from '@/features/scenarios/components/ChangePanel';
import { useChangesStore } from '@/features/scenarios/stores/changes';
import { World as WorldT, useWorldStore } from '@/stores/worlds';

import GeoJsonLayer from './GeoJsonLayer';
import OutlinersLayer from './OutlinersLayer';
import { WorldShellAdapter } from './adapters/ShellAdapter';

/**
 * A world, renders the respective map and the associated outliners.
 * @param id - The id of the world
 * @param side - The tab side on which the world is rendered
 */
export default function World({
    id,
    side,
}: {
    id: WorldT['id'];
    side: 'left' | 'right';
}) {
    const [showWorldShell, setShowWorldShell] = useState(false);
    useHotkeys('shift+meta+b, `', () => {
        setShowWorldShell((prev) => !prev);
    });

    const world = useWorldStore((state) => state.worlds[id]);
    const change = useChangesStore((state) => state.changes?.[id]);

    if (!world) return null;

    return (
        <div className=" w-full h-full absolute top-0 left-0">
            <Map root={world.tiles} side={side} world={id}>
                <GlobalShell show={showWorldShell} mapId={id} />
                <OutlinersLayer world={id} side={side} />
                <GeoJsonLayer world={id} side={side} />
            </Map>
            <div className="absolute top-0 left-0 ">
                {side === 'right' && change && (
                    <ChangePanel world={id} id={id} key={id} />
                )}
            </div>
        </div>
    );
}

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
