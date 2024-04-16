import { FeatureIDProto } from './generated/api';
import { UIResponseProto } from './generated/ui';

export type LatLng = {
    LatE7: number;
    LngE7: number;
};

export type StartupResponse = {
    version?: string;
    context?: {
        namespace: string;
        type: string;
    };
    docked?: UIResponseProto[];
    openDockIndex?: number;
    mapCenter?: LatLng;
    mapZoom?: number;
    root?: FeatureIDProto;
    expression?: string;
    error?: string;
    session: number;
    locked?: boolean;
};
