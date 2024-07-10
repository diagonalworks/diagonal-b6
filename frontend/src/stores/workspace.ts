import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';

import { usePersistURL } from '@/hooks/usePersistURL';
import { ImmerStateCreator } from '@/lib/zustand';
import { getNamespace, getValue } from '@/utils/world';

import { useWorldStore } from './worlds';

/**
 * Workspace store that holds data related with the workspace that wraps the tabs and the map.
 */
interface WorkspaceStore {
    root?: string;
}

export const createWorkspaceStore: ImmerStateCreator<
    WorkspaceStore,
    WorkspaceStore
> = () => ({
    root: undefined,
});

/**
 * Hook to use the workspace store, which holds data related with the workspace that wraps the tabs and the map.
 * This is a zustand store that uses immer for immutability.
 * @returns The workspace store
 */
export const useWorkspaceStore = create(immer(createWorkspaceStore));

type WorkspaceURLParams = {
    r?: string;
};

const encode = (state: Partial<WorkspaceStore>): WorkspaceURLParams => ({
    r: state.root || '',
});

const decode =
    (params: WorkspaceURLParams): ((state: WorkspaceStore) => WorkspaceStore) =>
    (state) => {
        const root = params.r || state.root;

        if (root) {
            useWorldStore.getState().actions.setFeatureId('baseline', {
                namespace: getNamespace(root),
                value: getValue(root),
            });
        }
        return {
            ...state,
            root,
        };
    };

export const usePersistWorkspaceURL = () => {
    return usePersistURL(useWorkspaceStore, encode, decode);
};
