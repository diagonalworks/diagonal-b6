import { atomWithImmer } from 'jotai-immer';

type Scenario = {
    name: string;
};

type Scenarios = Record<string, Scenario>;

export const appAtom = atomWithImmer<{
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
