import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';

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
export const useWorldStore = create(immer(createWorldStore));
