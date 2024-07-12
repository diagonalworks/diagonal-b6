import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import { immer } from 'zustand/middleware/immer';

import { usePersistURL } from '@/hooks/usePersistURL';
import { ImmerStateCreator } from '@/lib/zustand';
import { useWorkspaceStore } from '@/stores/workspace';
import { useWorldStore } from '@/stores/worlds';
import { getValue, getWorldFeatureId } from '@/utils/world';

// import { getWorldFeatureId } from '@/utils/world';
import { useChangesStore } from './changes';

export interface Tab {
    id: string;
    index: number;
    side: 'left' | 'right';
    properties: {
        name: string;
        closable: boolean;
        editable: boolean;
        persist: boolean;
    };
}

interface TabsStore {
    /* List of tabs */
    tabs: Tab[];
    /* The id of the active tab on the left side */
    leftTab?: Tab['id'];
    /* The id of the active tab on the right side */
    rightTab?: Tab['id'];
    /* Whether the screen is split */
    splitScreen: boolean;
    actions: {
        /**
         * Add a tab to the store
         * @param tab - The tab to add
         * @returns void
         */
        add: (tab: Tab) => void;
        /**
         * Remove a tab from the store
         * @param tabId - The id of the tab to remove
         * @returns void
         */
        remove: (tabId: Tab['id']) => void;
        /**
         * Rename a tab
         * @param tabId - The id of the tab to rename
         * @param name - The new name for the tab
         * @returns void
         */
        rename: (tabId: Tab['id'], name: string) => void;
        /**
         * Set a tab as active
         * @param tabId - The id of the tab to set as active
         * @param side - The side on which to set the tab as active
         * @returns void
         * */
        setActive: (tabId: Tab['id'], side: Tab['side']) => void;
        /**
         * Set whether the screen is split
         * @param splitScreen - Whether the screen is split
         * @returns void
         */
        setSplitScreen: (splitScreen: boolean) => void;
        setPersist: (tabId: Tab['id'], persist: boolean) => void;
    };
}

export const createTabsStore: ImmerStateCreator<TabsStore, TabsStore> = (
    set,
    get
) => ({
    tabs: [],
    splitScreen: false,
    actions: {
        add: (tab) => {
            set((state) => {
                state.tabs.push(tab);
                if (tab.side === 'left') {
                    state.leftTab = tab.id;
                } else {
                    state.rightTab = tab.id;
                }
                if (tab.side === 'left') return;
                const root = useWorkspaceStore.getState().root ?? 'baseline';

                const rootWorld = useWorldStore.getState().worlds?.[root];

                useWorldStore.getState().actions.createWorld({
                    id: tab.id,
                    featureId: getWorldFeatureId({
                        ...rootWorld?.featureId,
                        namespace: `${rootWorld.featureId.namespace}/scenario`,
                        value: getValue(tab.id),
                    }),
                    tiles: rootWorld.id,
                });
                useChangesStore.getState().actions.add({
                    id: tab.id,
                    origin: root,
                    target: tab.id,
                    created: false,
                    spec: {
                        features: [],
                    },
                });
            });
        },
        remove: (tabId) => {
            set((state) => {
                const index = state.tabs.findIndex((t) => t.id === tabId);
                state.tabs.splice(index, 1);
                if (tabId === state.leftTab) {
                    const nextTab = state.tabs.find((t) => t.side === 'left');
                    state.leftTab = nextTab?.id || 'baseline';
                } else if (tabId === state.rightTab) {
                    const nextTab = state.tabs.find((t) => t.side === 'right');
                    state.rightTab = nextTab?.id;
                    if (!nextTab) {
                        state.splitScreen = false;
                    }
                }
                // remove the world for the tab
                useWorldStore.getState().actions.removeWorld(tabId);
            });
        },
        rename: (tabId, name) => {
            set((state) => {
                const index = state.tabs.findIndex((t) => t.id === tabId);
                state.tabs[index].properties.name = name;
            });
        },
        setActive: (tabId, side) => {
            set((state) => {
                if (side === 'left') {
                    state.leftTab = tabId;
                    useWorkspaceStore.getState().setRoot(tabId);
                } else {
                    state.rightTab = tabId;
                }
            });
        },
        setSplitScreen: (splitScreen) => {
            set({ splitScreen });
        },
        setPersist: (tabId, persist) => {
            set((state) => {
                const index = state.tabs.findIndex((t) => t.id === tabId);
                state.tabs[index].properties.persist = persist;
            });
        },
    },
});

/**
 * The hook to use the tabs store. This is used to access and modify the tabs win the workspace.
 * This is a zustand store that uses immer for immutability.
 * @returns The tabs store
 */
export const useTabsStore = create(devtools(immer(createTabsStore)));

type TabsURLParams = {
    t?: string;
};

const encode = (state: Partial<TabsStore>): TabsURLParams => {
    if (!state.tabs || state.tabs.length === 0) return {};

    return {
        t: state.tabs
            .filter((tab) => tab.properties.persist)
            .map((tab) => {
                return `${tab.id}:${tab.side === 'left' ? 'l' : 'r'}:${
                    tab.index
                }:${tab.properties.name}`;
            })
            .join(','),
    };
};

const decode = (params: TabsURLParams): ((state: TabsStore) => TabsStore) => {
    return (state) => {
        if (!params.t) return state;

        const tabs = params.t.split(',').map((tab) => {
            const [id, side, index, name] = tab.split(':');
            return {
                id,
                index: parseInt(index),
                side: (side === 'l' ? 'left' : 'right') as Tab['side'],
                properties: {
                    name,
                    closable: side === 'right',
                    editable: side === 'right',
                },
            };
        });

        const rightTabs = tabs
            .filter((tab) => tab.side === 'right')
            .sort((a, b) => a.index - b.index);

        return {
            ...state,
            tabs,
            splitScreen: rightTabs.length > 0,
            rightTab: rightTabs?.[0]?.id,
        };
    };
};

export const useTabsURLStorage = () => {
    return usePersistURL(useTabsStore, encode, decode);
};
