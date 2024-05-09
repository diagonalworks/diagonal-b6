import { OutlinerStore } from '@/lib/context/outliner';
import { urlSearchParamsStorage } from '@/lib/storage';
import { FeatureIDProto } from '@/types/generated/api';
import {
    ComparisonLineProto,
    ComparisonRequestProto,
} from '@/types/generated/ui';
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

export type Comparator = {
    id: string;
    request: ComparisonRequestProto;
    data: ComparisonLineProto;
};

export type AppStore = {
    outliners: Record<string, OutlinerStore>;
    scenarios: Scenarios;
    tabs: {
        left: keyof Scenarios;
        right?: keyof Scenarios;
    };
    comparators: Record<string, Comparator>;
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
    comparators: {
        test: {
            id: 'test',
            request: {
                baseline: 'baseline' as unknown as FeatureIDProto,
                scenarios: ['scenario-1' as unknown as FeatureIDProto],
                analysis: 'test' as unknown as FeatureIDProto,
            },
            data: {
                baseline: {
                    bars: [
                        {
                            value: 5,
                            total: 20,
                            index: 1,
                            range: { value: 'test-a' },
                        },
                        {
                            value: 15,
                            total: 20,
                            index: 2,
                            range: { value: 'test-b' },
                        },
                    ],
                },
                scenarios: [
                    {
                        bars: [
                            {
                                value: 18,
                                total: 20,
                                index: 1,
                                range: { value: 'test-a' },
                            },
                            {
                                value: 2,
                                total: 20,
                                index: 2,
                                range: { value: 'test-b' },
                            },
                        ],
                    },
                ],
            },
        },
    },
};

/**
 *
 * The main app atom that stores the global state of the app.
 * Do not use this directly, use the `useAppContext` hook instead.
 */
export const appAtom = atomWithImmer<AppStore>(initialAppStore);
