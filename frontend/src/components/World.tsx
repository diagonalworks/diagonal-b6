import { Map } from '@/components/Map';
import { ChangePanel } from '@/features/scenarios/components/ChangePanel';
import { World as WorldT, useWorldStore } from '@/stores/worlds';
import { AnimatePresence, motion } from 'framer-motion';
import { useMemo, useState } from 'react';
import { useHotkeys } from 'react-hotkeys-hook';
import GeoJsonLayer from './GeoJsonLayer';
import OutlinersLayer from './OutlinersLayer';
import { WorldShellAdapter } from './adapters/ShellAdapter';

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

    const mapRoot = useMemo(() => {
        if (!world?.featureId) return '';
        return `collection/${world.featureId.namespace}/${world.featureId.value}`;
    }, [world?.featureId]);

    if (!world) return null;

    return (
        <div className=" w-full h-full absolute top-0 left-0">
            <Map root={mapRoot} side={side} world={id}>
                <GlobalShell show={showWorldShell} mapId={id} />
                <OutlinersLayer world={id} />
                <GeoJsonLayer world={id} side={side} />
            </Map>
            <div className="absolute top-0 left-0 ">
                {id !== 'baseline' && <ChangePanel world={id} id={id} />}
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
