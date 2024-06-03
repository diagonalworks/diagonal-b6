import { getWorldFeatureId } from '@/lib/world';
import { ImmerStateCreator } from '@/lib/zustand';
import { FeatureIDProto } from '@/types/generated/api';
import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';

export interface World {
    id: string;
    featureId: FeatureIDProto;
}

export interface WorldsStore {
    worlds: Record<string, World>;
    actions: {
        createWorld: (world: World) => void;
        removeWorld: (worldId: string) => void;
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

export const useWorldStore = create(immer(createWorldStore));
