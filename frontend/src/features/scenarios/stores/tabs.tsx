import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import { immer } from 'zustand/middleware/immer';

import { usePersistURL } from '@/hooks/usePersistURL';
import { ImmerStateCreator } from '@/lib/zustand';
import { useWorldStore } from '@/stores/worlds';

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
    };
}

interface TabsStore {
    /* List of tabs */
    tabs: Tab[];
    /* The id of the active tab on the left side */
    leftTab: Tab['id'];
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
    };
}

export const createTabsStore: ImmerStateCreator<TabsStore, TabsStore> = (
    set,
    get
) => ({
    tabs: [
        {
            id: useWorldStore.getState().worlds.baseline.id,
            index: 0,
            side: 'left',
            properties: {
                name: 'Baseline',
                closable: false,
                editable: false,
            },
        },
    ],
    leftTab: 'baseline',
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
                // create a new world for the tab
                useWorldStore.getState().actions.createWorld({
                    id: tab.id,
                    featureId:
                        useWorldStore.getState().worlds.baseline.featureId,
                });
                useChangesStore.getState().actions.add({
                    id: tab.id,
                    origin: get().leftTab,
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
                } else {
                    state.rightTab = tabId;
                }
            });
        },
        setSplitScreen: (splitScreen) => {
            set({ splitScreen });
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
    console.log('HERE');
    console.log(state.tabs);
    if (!state.tabs || state.tabs.length === 0) return {};
    const t = state.tabs
        .map(
            (tab) => `${tab.id}:${tab.side}:${tab.index}:${tab.properties.name}`
        )
        .join(',');
    return {
        t,
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
                side: side as 'left' | 'right',
                properties: {
                    name,
                    closable: side === 'right',
                    editable: side === 'right',
                },
            };
        });

        tabs.forEach(() => {
            // create world for tab if it doesnt exist yet
            // if right tab, create change - ignor for now, different PR.
        });
        return {
            ...state,
            tabs: state.tabs.concat(tabs),
        };
    };
};

export const useTabsURLStorage = () => {
    return usePersistURL(useTabsStore, encode, decode);
};
