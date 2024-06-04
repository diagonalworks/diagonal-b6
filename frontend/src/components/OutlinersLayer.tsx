import {
    DndContext,
    MouseSensor,
    PointerSensor,
    TouchSensor,
    useDraggable,
    useDroppable,
    useSensor,
    useSensors,
} from '@dnd-kit/core';
import {
    restrictToParentElement,
    restrictToWindowEdges,
} from '@dnd-kit/modifiers';
import { AnimatePresence, motion } from 'framer-motion';
import React, { PropsWithChildren, useMemo } from 'react';
import { twMerge } from 'tailwind-merge';

import { OutlinerSpec, useOutlinersStore } from '@/stores/outliners';
import { World } from '@/stores/worlds';
import { popOpen } from '@/utils/animations';

import Outliner from './Outliner';

const useOutlinerSensors = () => {
    const pointerSensor = useSensor(PointerSensor, {
        activationConstraint: {
            distance: 5,
        },
    });
    const mouseSensor = useSensor(MouseSensor);
    const touchSensor = useSensor(TouchSensor);
    return useSensors(pointerSensor, mouseSensor, touchSensor);
};

/**
 * The layer that renders the outliners in a world.
 * @param world - The id of the world to render outliners for
 */
function OutlinersLayer({ world }: { world: World['id'] }) {
    const outliners = useOutlinersStore((state) =>
        state.actions.getByWorld(world)
    );
    const actions = useOutlinersStore((state) => state.actions);

    const sensors = useOutlinerSensors();

    const [dockedOutliners, draggableOutliners] = useMemo(() => {
        const dockedOutliners: OutlinerSpec[] = [];
        const draggableOutliners: OutlinerSpec[] = [];

        outliners.forEach((outliner) => {
            if (outliner.properties.docked) {
                dockedOutliners.push(outliner);
            } else {
                if (outliner.properties.type === 'core') {
                    draggableOutliners.push(outliner);
                }
            }
        });

        return [dockedOutliners, draggableOutliners];
    }, [outliners]);

    return (
        <div className="h-full w-full">
            <div className="absolute top-16 left-2 flex flex-col gap-1">
                {dockedOutliners.map((outliner) => {
                    return <Outliner key={outliner.id} outliner={outliner} />;
                })}
            </div>
            <DndContext
                modifiers={[restrictToWindowEdges, restrictToParentElement]}
                sensors={sensors}
                onDragStart={({ active }) => {
                    actions.setActive(active.id as string, true);
                    actions.setTransient(active.id as string, false);
                }}
                onDragEnd={({ active, delta }) => {
                    actions.move(active.id as string, delta.x, delta.y);
                    actions.setActive(active.id as string, false);
                }}
            >
                <Droppable world={world}>
                    <AnimatePresence>
                        {draggableOutliners.map((outliner) => (
                            <DraggableOutliner
                                key={outliner.id}
                                outliner={outliner}
                            />
                        ))}
                    </AnimatePresence>
                </Droppable>
            </DndContext>
        </div>
    );
}

const memoizedOutlinersLayer = React.memo(OutlinersLayer);
export default memoizedOutlinersLayer;

const Droppable = ({
    children,
    world,
}: PropsWithChildren & { world: World['id'] }) => {
    const { setNodeRef } = useDroppable({
        id: `droppable-${world}`,
    });

    return (
        <div ref={setNodeRef} className="w-full h-full">
            {children}
        </div>
    );
};

const DraggableOutliner = ({ outliner }: { outliner: OutlinerSpec }) => {
    const { attributes, transform, setNodeRef, listeners } = useDraggable({
        id: outliner.id,
    });

    const style = useMemo(
        () => ({
            transform: `${
                transform
                    ? `translate3d(${transform.x}px, ${transform.y}px, 0)`
                    : ''
            }`,
        }),
        [transform]
    );

    return (
        <div
            id={outliner.id}
            ref={setNodeRef}
            style={{
                ...style,
                ...(outliner.properties.coordinates
                    ? {
                          top: outliner.properties.coordinates.y + 4,
                          left: outliner.properties.coordinates.x + 4,
                      }
                    : {
                          top: 280,
                          left: 8,
                      }),
                position: 'absolute',
            }}
            className={twMerge(
                outliner.properties.active &&
                    '[&_.stack-wrapper]:ring-2 [&_.stack-wrapper]:ring-ultramarine-50 [&_.stack-wrapper]:ring-opacity-40'
            )}
            {...attributes}
            {...listeners}
        >
            <motion.div
                variants={popOpen}
                initial="hidden"
                animate="visible"
                exit="hidden"
                transition={{
                    duration: 0.1,
                }}
            >
                <Outliner outliner={outliner} />
            </motion.div>
        </div>
    );
};
