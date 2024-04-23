import { urlSearchParamsStorage } from '@/lib/storage';
import { UIResponseProto } from '@/types/generated/ui';
import { StartupResponse } from '@/types/startup';
import type { FeatureCollection } from 'geojson';
import { atomWithImmer } from 'jotai-immer';
import { atomWithStorage } from 'jotai/utils';
import { Point } from 'maplibre-gl';

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

export type AppStore = {
    startup?: StartupResponse;
    scenarios: Scenarios;
    tabs: {
        left: keyof Scenarios;
        right?: keyof Scenarios;
    };
    geojson: Record<string, FeatureCollection[]>;
    stacks: Record<
        string,
        {
            id: string;
            expression?: string;
            coordinates?: Point;
            docked: boolean;
            transient?: boolean;
            proto: UIResponseProto;
            tab?: keyof Scenarios;
        }
    >;
};

export const appAtom = atomWithImmer<AppStore>({
    tabs: {
        left: 'baseline',
    },
    scenarios: {
        baseline: {
            name: 'Baseline',
        },
    },
    stacks: {},
    geojson: {},
});
