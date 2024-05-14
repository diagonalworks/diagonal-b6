import {
    ComparisonRequestProto,
    UIRequestProto,
    UIResponseProto,
} from '@/types/generated/ui';
import { $FixMe } from '@/utils/defs';
import { StyleSpecification } from 'maplibre-gl';
import { LocationState } from './location';

type OutlinerSpec = {
    id: string;
    world: WorldSpec['id'];
    state: {
        docked: boolean;
        transient: boolean;
        active?: boolean;
        coordinates?: { x: number; y: number };
    };
    request?: UIRequestProto | ComparisonRequestProto;
    fallbackData?: UIResponseProto;
};

type WorldSpec = {
    id: string;
    name?: string;
    map: {
        tilesPath: string;
        mapStyle: StyleSpecification;
        layers: $FixMe[];
    };
    outliners: OutlinerSpec[];
};

type AppStore = {
    location: LocationState;
    worlds: WorldSpec[];
};
