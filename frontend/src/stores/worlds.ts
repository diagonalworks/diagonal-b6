import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import { immer } from 'zustand/middleware/immer';

import { useTabsStore } from '@/features/scenarios/stores/tabs';
import { usePersistURL } from '@/hooks/usePersistURL';
import { ImmerStateCreator } from '@/lib/zustand';
import { FeatureIDProto } from '@/types/generated/api';
import { getWorldFeatureId } from '@/utils/world';

/**
 * A World represents a single world in the application.
 * It contains an id and a FeatureIDProto that represents the world in b6.
 */
export interface World {
    id: string;
    featureId: FeatureIDProto;
    tiles: string;
    shellHistory?: string[];
}

export interface WorldsStore {
    isShellEnabled: boolean;
    worlds: Record<string, World>;
    actions: {
        /**
         * Set whether or not to show the shell.
         * @param enabled - Should it be enabled everywhere?
         * @returns void
         */
        setShellEnabled: (enabled: boolean) => void;
        /**
         * Create a new world in the store
         * @param world - The world to create
         * @returns void
         */
        createWorld: (world: World) => void;
        /**
         * Remove a world from the store
         * @param worldId - The id of the world to remove
         * @returns void
         */
        removeWorld: (worldId: string) => void;
        /**
         * Set the feature id for a world
         * @param worldId - The id of the world to set the feature id for
         * @param featureId - The feature id to set
         * @returns void
         */
        setFeatureId: (worldId: string, featureId: FeatureIDProto) => void;
        /**
         * Set the tiles for a world
         * @param worldId - The id of the world to set the tiles for
         * @param tiles - The path for the tiles to set
         * @returns void
         */
        setTiles: (worldId: string, tiles: string) => void;
        /**
         * Add a shell command to the history of a world
         * @param worldId - The id of the world to add the command to
         * @param shell - The shell command to add
         * @returns void
         */
        addToShellHistory: (worldId: string, shell: string) => void;
    };
}

export const createWorldStore: ImmerStateCreator<WorldsStore, WorldsStore> = (
    set
) => ({
    isShellEnabled: false,
    worlds: {},
    actions: {
        setShellEnabled: (enabled) => {
          set((state) => {
            state.isShellEnabled = enabled;
          });
        },
        createWorld: (world) => {
            set((state) => {
                state.worlds[world.id] = world;
            });
        },
        removeWorld: (worldId) => {
            set((state) => {
                delete state.worlds[worldId];
            });
        },
        setFeatureId: (worldId, featureId) => {
            set((state) => {
                state.worlds[worldId].featureId = featureId;
            });
        },
        setTiles: (worldId, tiles) => {
            set((state) => {
                state.worlds[worldId].tiles = tiles;
            });
        },
        addToShellHistory: (worldId, shell) => {
            set((state) => {
                state.worlds[worldId].shellHistory = [
                    ...(state.worlds[worldId].shellHistory ?? []),
                    shell,
                ];
            });
        },
    },
});

/**
 * A hook to use the world store that provides access to the worlds in the application.
 * This is a zustand store that uses immer for immutability.
 * @returns The world store
 */
export const useWorldStore = create(devtools(immer(createWorldStore)));

type WorldURLParams = {
    w?: string;
};

const encode = (state: Partial<WorldsStore>): WorldURLParams => {
    if (!state.worlds) {
        return {};
    }
    const persistWorlds = Object.values(state.worlds)
        .filter(
            (w) =>
                useTabsStore.getState().tabs.find((t) => t.id === w.id)
                    ?.properties.persist
        )
        .map((w) => w.id);

    return {
        w: persistWorlds.join(','),
    };
};

const decode = (
    params: WorldURLParams
): ((state: WorldsStore) => WorldsStore) => {
    const worlds: World[] =
        params.w?.split(',').flatMap((ws) => {
            const value = ws.match(/([a-z]|\d)*$/)?.[0];
            const namespace = ws.match(/.*(?=\/([a-z]|\d)*$)/)?.[0];
            if (!value) {
                return [];
            }
            return {
                id: ws,
                featureId: getWorldFeatureId({
                    namespace,
                    value: +value,
                }),
                tiles: ws,
            };
        }) ?? [];

    return (state) => ({
        ...state,
        worlds: worlds.reduce((acc, w) => {
            acc[w.id] = w;
            return acc;
        }, Object.assign({}, state.worlds) as Record<string, World>),
    });
};

export const useWorldURLStorage = () => {
    return usePersistURL(useWorldStore, encode, decode);
};
