import { OutlinerStore } from '@/lib/context/outliner';
import { urlSearchParamsStorage } from '@/lib/storage';
import { atomWithImmer } from 'jotai-immer';
import { atomWithStorage } from 'jotai/utils';

export type Scenario = {
    name: string;
    id: string;
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
    scenarios: Scenarios;
    tabs: {
        left: keyof Scenarios;
        right?: keyof Scenarios;
    };
};

export const initialAppStore: AppStore = {
    outliners: {},
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
