import { PlusIcon } from '@radix-ui/react-icons';
import { AnimatePresence } from 'framer-motion';
import { isEmpty } from 'lodash';
import { customAlphabet } from 'nanoid';
import { useCallback, useEffect, useMemo } from 'react';
import { MapProvider } from 'react-map-gl';

import { useStartup } from '@/api/startup';
import ComparisonCard from '@/features/scenarios/components/ComparisonCard';
import { Tabs } from '@/features/scenarios/components/Tabs';
import { useComparisonsStore } from '@/features/scenarios/stores/comparisons';
import {
    useTabsStore,
    useTabsURLStorage,
} from '@/features/scenarios/stores/tabs';
import { useViewStore, useViewURLStorage } from '@/stores/view';
import { useWorkspaceStore, useWorkspaceURLStorage } from '@/stores/workspace';
import { useWorldStore, useWorldURLStorage } from '@/stores/worlds';

import World from './World';

const generateWorldId = (path?: string) => {
    const nanoid = customAlphabet('1234567890', 6);
    return `${path}${nanoid()}`;
};

/**
 * The workspace on which worlds are rendered. It contains the tabs and the map.
 */
export default function Workspace() {
    useWorkspaceURLStorage();
    useViewURLStorage();
    useWorldURLStorage();
    useTabsURLStorage();

    const worlds = useWorldStore((state) => state.worlds);
    const root = useWorkspaceStore((state) => state.root);
    const rootWorld = worlds?.[root || 'baseline'];

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

    useEffect(() => {
        if (leftTabs.length === 0 && !isEmpty(worlds)) {
            tabActions.add({
                id: root || 'baseline',
                side: 'left',
                index: 0,
                properties: {
                    name: 'Baseline',
                    closable: false,
                    editable: false,
                    persist: true,
                },
            });
        }
    }, [tabs, worlds]);

    const comparators = useComparisonsStore((state) => state.comparisons);

    const comparison = useMemo(() => {
        return Object.values(comparators).find(
            (comparison) =>
                comparison.baseline.id === leftTab &&
                comparison.scenarios.some(
                    (scenario) => scenario.id === rightTab
                )
        );
    }, [comparators, leftTab, rightTab]);

    const handleAddScenario = useCallback(() => {
        const id = generateWorldId(
            `/collection/${rootWorld?.featureId?.namespace}/scenario/`
        );
        tabActions.add({
            id,
            side: 'right',
            index: rightTabs.length,
            properties: {
                name: 'Untitled',
                closable: true,
                editable: true,
                persist: false,
            },
        });
        tabActions.setActive(id, 'right');
    }, [tabActions, rightTabs.length, rootWorld?.featureId?.namespace]);

    return (
        <div className="h-screen max-h-screen flex flex-col relative">
            {/* @TODO: extract tabs menu logic to a separate component. */}
            <Tabs.Menu splitScreen={splitScreen}>
                <div className="flex items-end justify-between gap-1">
                    <div className="flex gap-1">
                        {leftTabs.map((tab) => (
                            <Tabs.Button
                                key={tab.id}
                                tab={tab}
                                active={leftTab === tab.id}
                                onClick={(id) =>
                                    tabActions.setActive(id, 'left')
                                }
                                onClose={tabActions.remove}
                                onValueChange={tabActions.rename}
                            />
                        ))}
                    </div>
                    {rightTabs.length === 0 && (
                        <button
                            onClick={() => {
                                handleAddScenario();
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
                <div className="flex gap-1">
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
                    {rightTabs.length > 0 && (
                        <button
                            className="bg-rose-10 hover:bg-rose-20  border border-b border-b-rose-40 border-rose-30 text-rose-70 hover:text-rose-90 px-2 rounded-t"
                            aria-label="create new scenario"
                            onClick={handleAddScenario}
                        >
                            <PlusIcon />
                        </button>
                    )}
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
            {comparison && (
                <div className="absolute bottom-10 left-1/2 -translate-x-1/2 translate bg-white">
                    <ComparisonCard comparison={comparison} />
                </div>
            )}
        </div>
    );
}
