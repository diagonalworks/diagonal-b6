import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';

import { usePersistURL } from '@/hooks/usePersistURL';
import { ImmerStateCreator } from '@/lib/zustand';

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
    (state) => ({
        ...state,
        root: params.r || state.root,
    });

/**
 * Hook to use URL persistence for the workspace store.
 */
export const useWorkspaceURLStorage = () => {
    return usePersistURL(useWorkspaceStore, encode, decode);
};
