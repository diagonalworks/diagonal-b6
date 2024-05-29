import { ImmerStateCreator } from '@/lib/zustand';
import { useWorldStore } from '@/stores/worlds';
import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';

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
    tabs: Tab[];
    leftTab: Tab['id'];
    rightTab?: Tab['id'];
    splitScreen: boolean;
    actions: {
        add: (tab: Tab) => void;
        remove: (tabId: Tab['id']) => void;
        rename: (tabId: Tab['id'], name: string) => void;
        setActive: (tabId: Tab['id'], side: Tab['side']) => void;
        setSplitScreen: (splitScreen: boolean) => void;
    };
}

export const createTabsStore: ImmerStateCreator<TabsStore, TabsStore> = (
    set
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
                useWorldStore.getState().actions.createWorld({ id: tab.id });
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

export const useTabsStore = create(immer(createTabsStore));
