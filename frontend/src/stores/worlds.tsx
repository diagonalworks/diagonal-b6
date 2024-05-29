import { ImmerStateCreator } from '@/lib/zustand';
import { FeatureIDProto } from '@/types/generated/api';
import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';

export interface World {
    id: string;
    featureId?: FeatureIDProto;
}

export interface WorldsStore {
    worlds: Record<string, World>;
    actions: {
        createWorld: (world: World) => void;
        removeWorld: (worldId: string) => void;
    };
}

export const createWorldStore: ImmerStateCreator<WorldsStore, WorldsStore> = (
    set
) => ({
    worlds: {
        baseline: {
            id: 'baseline',
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
    },
});

export const useWorldStore = create(immer(createWorldStore));
