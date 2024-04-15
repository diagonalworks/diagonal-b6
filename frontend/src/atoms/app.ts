import { urlSearchParamsStorage } from '@/lib/storage';
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
    session: number | null;
    scenarios: Scenarios;
    tabs: {
        left: keyof Scenarios;
        right?: keyof Scenarios;
    };
}>({
    session: null,
    tabs: {
        left: 'baseline',
    },
    scenarios: {
        baseline: {
            name: 'Baseline',
        },
    },
});
