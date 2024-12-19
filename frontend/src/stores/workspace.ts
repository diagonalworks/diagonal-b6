import { create } from "zustand";
import { immer } from "zustand/middleware/immer";

import { useTabsStore } from "@/features/scenarios/stores/tabs";
import { usePersistURL } from "@/hooks/usePersistURL";
import { ImmerStateCreator } from "@/lib/zustand";

/**
 * Workspace store that holds data related with the workspace that wraps the tabs and the map.
 */
interface WorkspaceStore {
	root?: string;
	setRoot: (root: string) => void;
}

export const createWorkspaceStore: ImmerStateCreator<
	WorkspaceStore,
	WorkspaceStore
> = (set) => ({
	root: undefined,
	setRoot: (root: string) => {
		set((state) => {
			state.root = root;
		});
	},
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
	r: state.root || "",
});

const decode =
	(params: WorkspaceURLParams): ((state: WorkspaceStore) => WorkspaceStore) =>
	(state) => {
		const root = params.r || state.root;
		if (root) {
			useTabsStore.getState().actions.setActive(root, "left");
		}
		return {
			...state,
			root,
		};
	};

export const useWorkspaceURLStorage = () => {
	return usePersistURL(useWorkspaceStore, encode, decode);
};
