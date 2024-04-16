import { urlSearchParamsStorage } from '@/lib/storage';
import { StartupResponse } from '@/types/startup';
import { atomWithImmer } from 'jotai-immer';
import { atomWithStorage } from 'jotai/utils';

type Scenario = {
    name: string;
};

type Scenarios = Record<string, Scenario>;

export const collectionAtom = atomWithStorage(
    'r',
    '',
    urlSearchParamsStorage({}),
    {
        getOnInit: true,
    }
);

export const appAtom = atomWithImmer<{
    startup?: StartupResponse;
    scenarios: Scenarios;
    tabs: {
        left: keyof Scenarios;
        right?: keyof Scenarios;
    };
}>({
    tabs: {
        left: 'baseline',
    },
    scenarios: {
        baseline: {
            name: 'Baseline',
        },
    },
});
