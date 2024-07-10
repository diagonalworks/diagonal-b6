import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import { immer } from 'zustand/middleware/immer';

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
}

export interface WorldsStore {
    worlds: Record<string, World>;
    actions: {
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
    };
}

export const createWorldStore: ImmerStateCreator<WorldsStore, WorldsStore> = (
    set
) => ({
    worlds: {
        baseline: {
            id: 'baseline',
            featureId: getWorldFeatureId('baseline'),
        },
    },
    actions: {
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
    console.log(state.worlds);
    if (!state.worlds) {
        return {};
    }
    return {
        w: Object.values(state.worlds)
            .map((w) => `${w.featureId.namespace}/${w.featureId.value}`)
            .join(','),
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
                id: value,
                featureId: getWorldFeatureId(value, namespace),
            };
        }) ?? [];

    return (state) => ({
        ...state,
        worlds: worlds.reduce((acc, w) => {
            acc[w.id] = w;
            return acc;
        }, state.worlds ?? {}),
    });
};

export const useWorldURLStorage = () => {
    return usePersistURL(useWorldStore, encode, decode);
};
