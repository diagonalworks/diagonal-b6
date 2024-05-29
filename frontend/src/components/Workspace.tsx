import { Tabs } from '@/features/scenarios/components/Tabs';
import { useTabsStore } from '@/features/scenarios/stores/tabs';
import { useStartup } from '@/lib/api/startup';
import { useViewStore, useViewURLStorage } from '@/stores/view';
import { useWorkspaceURLStorage } from '@/stores/workspace';
import { PlusIcon } from '@radix-ui/react-icons';
import { AnimatePresence } from 'framer-motion';
import { customAlphabet } from 'nanoid';
import { useEffect, useMemo } from 'react';
import { MapProvider } from 'react-map-gl';
import World from './World';

const generateWorldId = () => {
    const nanoid = customAlphabet('1234567890', 6);
    return nanoid();
};

export default function Workspace() {
    useWorkspaceURLStorage();
    useViewURLStorage();
    const [setView, view] = useViewStore((state) => [
        state.actions.setView,
        state.view,
    ]);

    const startup = useStartup();

    useEffect(() => {
        if (startup.data) {
            setView({
                ...view,
                ...(startup.data.mapCenter && {
                    latitude: startup.data.mapCenter.latE7 / 1e7,
                    longitude: startup.data.mapCenter.lngE7 / 1e7,
                }),
                ...(startup.data.mapZoom && {
                    zoom: startup.data.mapZoom,
                }),
            });
        }
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [startup.data, setView]);

    const {
        splitScreen,
        tabs,
        actions: tabActions,
        leftTab,
        rightTab,
    } = useTabsStore();
    const leftTabs = useMemo(
        () => tabs.filter((tab) => tab.side === 'left'),
        [tabs]
    );
    const rightTabs = useMemo(
        () => tabs.filter((tab) => tab.side === 'right'),
        [tabs]
    );

    return (
        <div className="h-screen max-h-screen flex flex-col relative">
            {/* @TODO: extract tabs menu logic to a separate component. */}
            <Tabs.Menu splitScreen={splitScreen}>
                <div className="flex items-end justify-between gap-1">
                    {leftTabs.map((tab) => (
                        <Tabs.Button
                            key={tab.id}
                            tab={tab}
                            active={leftTab === tab.id}
                            onClick={(id) => tabActions.setActive(id, 'left')}
                            onClose={tabActions.remove}
                            onValueChange={tabActions.rename}
                        />
                    ))}
                    {rightTabs.length === 0 && (
                        <button
                            onClick={() => {
                                const id = generateWorldId();
                                tabActions.add({
                                    id,
                                    side: 'right',
                                    index: rightTabs.length,
                                    properties: {
                                        name: 'Untitled',
                                        closable: true,
                                        editable: true,
                                    },
                                });
                                tabActions.setActive(id, 'right');
                                tabActions.setSplitScreen(true);
                            }}
                            aria-label="add scenario"
                            className="text-sm flex gap-2 mb-[1px] items-center bg-rose-10 rounded w-fit border border-b-0 hover:bg-rose-20 rounded-b-none border-rose-30 text-rose-60 px-2 py-1"
                        >
                            <PlusIcon />
                            scenario
                        </button>
                    )}
                </div>
                <div className="flex items-end justify-between gap-1">
                    {rightTabs.map((tab) => (
                        <Tabs.Button
                            key={tab.id}
                            tab={tab}
                            active={rightTab === tab.id}
                            onClick={(id) => tabActions.setActive(id, 'right')}
                            onClose={tabActions.remove}
                            onValueChange={tabActions.rename}
                            initial={{
                                x: 100,
                            }}
                            animate={{
                                x: 0,
                            }}
                        />
                    ))}
                </div>
            </Tabs.Menu>
            <Tabs.Content>
                {leftTab && (
                    <Tabs.Item side="left" splitScreen={splitScreen}>
                        <MapProvider>
                            <World id={leftTab} side="left" />
                        </MapProvider>
                    </Tabs.Item>
                )}
                <AnimatePresence>
                    {splitScreen && rightTab && (
                        <Tabs.Item
                            side="right"
                            splitScreen
                            transition={{ duration: 0.2 }}
                        >
                            <MapProvider>
                                <World id={rightTab} side="right" />
                            </MapProvider>
                        </Tabs.Item>
                    )}
                </AnimatePresence>
            </Tabs.Content>
        </div>
    );
}
