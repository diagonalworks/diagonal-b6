import { isEqual } from 'lodash';
import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import { immer } from 'zustand/middleware/immer';

import { usePersistURL } from '@/hooks/usePersistURL';
import { ImmerStateCreator } from '@/lib/zustand';
import { World } from '@/stores/worlds';
import { FeatureIDProto } from '@/types/generated/api';

import { ChangeFeature, ChangeFunction, ChangeSpec } from '../types/change';

export interface Change {
    /**
     * The unique identifier for the change
     */
    id: string;
    /**
     * The world that the change is applied to
     */
    origin: World['id'];
    /**
     * The world that results from applying the change
     */
    target: World['id'];
    /**
     * Whether the change has been applied and the target world has been created
     */
    created: boolean;
    /**
     * The specification for the change
     */
    spec: ChangeSpec;
}

/**
 * The store for changes. This is used to store the changes that define a scenario.
 */
interface ChangesStore {
    /**
     * A record of changes, indexed by their unique identifier.
     */
    changes: Record<string, Change>;
    actions: {
        /**
         * Add a change to the store
         * @param change - The change to add
         * @returns void
         */
        add: (change: Change) => void;
        /**
         * Remove a change from the store
         * @param change - The change to remove
         * @returns void
         */
        remove: (change: Change) => void;
        /**
         * Add a feature to a change
         * @param id - The id of the change to add the feature to
         * @param feature - The feature to add
         * @returns void
         */
        addFeature: (id: Change['id'], feature: ChangeFeature) => void;
        /**
         * Remove a feature from a change
         * @param id - The id of the change to remove the feature from
         * @param feature - The feature to remove
         * @returns void
         */
        removeFeature: (id: Change['id'], feature: ChangeFeature) => void;
        /**
         * Set the function that performs the changes
         * @param id - The id of the change to set the function of
         * @param func - The function to set
         * @returns void
         */
        setFunction: (id: Change['id'], func: ChangeFunction) => void;
        /**
         * Set the analysis that should be performed after the change is applied
         * @TODO: Consider decoupling the analysis from the change
         * @param id - The id of the change to set the analysis of
         * @param analysis - The analysis to set
         * @returns void
         */
        setAnalysis: (id: Change['id'], analysis: FeatureIDProto) => void;
        setCreate: (id: Change['id'], created: boolean) => void;
    };
}

export const createChangesStore: ImmerStateCreator<
    ChangesStore,
    ChangesStore
> = (set, get) => ({
    changes: {},
    actions: {
        add: (change) => {
            set((state) => {
                state.changes[change.id] = change;
            });
        },
        remove: (change) => {
            set((state) => {
                delete state.changes[change.id];
            });
        },
        addFeature: (id, feature) => {
            set((state) => {
                state.changes[id].spec.features.push(feature);
            });
        },
        removeFeature: (id, feature) => {
            set((state) => {
                state.changes[id].spec.features = get().changes[
                    id
                ].spec.features.filter((f) => !isEqual(f.id, feature.id));
            });
        },
        setFunction: (id, func) => {
            set((state) => {
                state.changes[id].spec.changeFunction = func;
            });
        },
        setAnalysis: (id, analysis) => {
            set((state) => {
                state.changes[id].spec.analysis = analysis;
            });
        },
        setCreate: (id, created) => {
            set((state) => {
                state.changes[id].created = created;
            });
        },
    },
});

/**
 * The hook to use the changes store. This is used to access and modify the changes that define the scenarios in the application.
 * This is a zustand store that uses immer for immutability.
 * @returns The changes store
 */
export const useChangesStore = create(devtools(immer(createChangesStore)));

type ChangesURLParams = {
    s?: string;
};

const encode = (state: Partial<ChangesStore>): ChangesURLParams => {
    if (!state.changes) {
        return {};
    }
    const createdScenarios = Object.entries(state.changes)
        .filter(([, change]) => change.created)
        .map(([id]) => id);
    return {
        s: createdScenarios.join(','),
    };
};

const decode = (
    params: ChangesURLParams
): ((state: ChangesStore) => ChangesStore) => {
    return (state) => {
        const createdScenarios = params.s?.split(',') ?? [];
        createdScenarios.forEach((id) => {
            if (!state.changes[id]) {
                state.actions.add({
                    id,
                    origin: 'baseline',
                    target: id,
                    created: true,
                    spec: {
                        features: [],
                    },
                });
            }
        });
        return state;
    };
};

export const useChangesURLStorage = () => {
    return usePersistURL(useChangesStore, encode, decode);
};
