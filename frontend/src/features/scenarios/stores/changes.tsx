import { ImmerStateCreator } from '@/lib/zustand';
import { World } from '@/stores/worlds';
import { FeatureIDProto } from '@/types/generated/api';
import { isEqual } from 'lodash';
import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';
import { ChangeFeature, ChangeFunction, ChangeSpec } from '../types/change';

export interface Change {
    id: string;
    origin: World['id'];
    target: World['id'];
    created: boolean;
    spec: ChangeSpec;
}

interface ChangesStore {
    changes: Record<string, Change>;
    actions: {
        add: (change: Change) => void;
        remove: (change: Change) => void;
        addFeature: (id: Change['id'], feature: ChangeFeature) => void;
        removeFeature: (id: Change['id'], feature: ChangeFeature) => void;
        setFunction: (id: Change['id'], func: ChangeFunction) => void;
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

export const useChangesStore = create(immer(createChangesStore));
