import { ImmerStateCreator } from '@/lib/zustand';
import { World } from '@/stores/worlds';
import { FeatureIDProto } from '@/types/generated/api';
import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';

export interface Comparison {
    id: string;
    baseline: World;
    scenarios: World[];
    analysis: FeatureIDProto;
}

export interface ComparisonsStore {
    comparisons: Record<string, Comparison>;
    actions: {
        add: (comparison: Comparison) => void;
        remove: (comparisonId: Comparison['id']) => void;
    };
}

export const createComparisonsStore: ImmerStateCreator<
    ComparisonsStore,
    ComparisonsStore
> = (set) => ({
    comparisons: {},
    actions: {
        add: (comparison) => {
            set((state) => {
                state.comparisons[comparison.id] = comparison;
            });
        },
        remove: (comparisonId) => {
            set((state) => {
                delete state.comparisons[comparisonId];
            });
        },
    },
});

export const useComparisonsStore = create(immer(createComparisonsStore));
