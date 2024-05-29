import { usePersistURL } from '@/hooks/usePersistURL';
import { ImmerStateCreator } from '@/lib/zustand';
import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';

interface WorkspaceStore {
    root?: string;
}

export const createWorkspaceStore: ImmerStateCreator<
    WorkspaceStore,
    WorkspaceStore
> = () => ({
    root: undefined,
});

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

export const useWorkspaceURLStorage = () => {
    return usePersistURL(useWorkspaceStore, encode, decode);
};
