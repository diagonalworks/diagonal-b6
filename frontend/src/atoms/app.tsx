import { Comparator } from '@/lib/context/comparator';
import { OutlinerStore } from '@/lib/context/outliner';
import { urlSearchParamsStorage } from '@/lib/storage';
import { FeatureIDProto } from '@/types/generated/api';
import { LabelledIconProto } from '@/types/generated/ui';
import { atomWithImmer } from 'jotai-immer';
import { atomWithStorage } from 'jotai/utils';

export type ChangeFeature = {
    id: FeatureIDProto;
    label?: LabelledIconProto;
    expression: string;
};

export type ChangeFunction = {
    label?: string;
    id: FeatureIDProto;
};

export type Change = {
    analysis?: FeatureIDProto;
    features?: ChangeFeature[];
    changeFunction?: ChangeFunction;
};

export type ScenarioSpec = {
    name: string;
    id: string;
};

export type Scenario = ScenarioSpec & {
    change?: Change;
    featureId?: FeatureIDProto;
};

export type Scenarios = Record<string, Scenario>;

export const collectionAtom = atomWithStorage(
    'r',
    '',
    urlSearchParamsStorage({}),
    {
        getOnInit: true,
    }
);

export type AppStore = {
    outliners: Record<string, OutlinerStore>;
    comparators: Record<string, Comparator>;
    scenarios: Scenarios;
    tabs: {
        left: keyof Scenarios;
        right?: keyof Scenarios;
    };
};

export const initialAppStore: AppStore = {
    outliners: {},
    comparators: {},
    scenarios: {
        baseline: {
            id: 'baseline',
            name: 'Baseline',
        },
    },
    tabs: {
        left: 'baseline',
    },
};

/**
 *
 * The main app atom that stores the global state of the app.
 * Do not use this directly, use the `useAppContext` hook instead.
 */
export const appAtom = atomWithImmer<AppStore>(initialAppStore);
