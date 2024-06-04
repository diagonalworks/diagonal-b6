import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';

import { ImmerStateCreator } from '@/lib/zustand';
import { World } from '@/stores/worlds';
import { FeatureIDProto } from '@/types/generated/api';

export interface Comparison {
    id: string;
    /**
     * The baseline world, which is the world that the scenarios are compared to.
     */
    baseline: World;
    /**
     * The scenarios that are compared to the baseline world.
     */
    scenarios: World[];
    /**
     * The analysis that is compared between the baseline and scenario worlds.
     */
    analysis: FeatureIDProto;
}

export interface ComparisonsStore {
    /* A record of comparisons, indexed by their unique identifier. */
    comparisons: Record<string, Comparison>;
    actions: {
        /**
         * Add a comparison to the store
         * @param comparison - The comparison to add
         */
        add: (comparison: Comparison) => void;
        /**
         * Remove a comparison from the store
         * @param comparisonId - The id of the comparison to remove
         */
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

/**
 * The store for comparisons. This is used to store the comparisons that define a scenario.
 * This is a zustand store that uses immer for immutability.
 * @returns The store for comparisons
 */
export const useComparisonsStore = create(immer(createComparisonsStore));
