import { $FixMe } from '@/utils/defs';
import { FeatureIDProto } from './generated/api';
import { UIResponseProto } from './generated/ui';

export type LatLng = {
    latE7: number;
    lngE7: number;
};

export type StartupResponse = {
    version?: string;
    context?: {
        namespace: string;
        type: string;
    };
    docked?: {
        geoJson: $FixMe;
        proto: UIResponseProto;
    }[];
    openDockIndex?: number;
    mapCenter?: LatLng;
    mapZoom?: number;
    root?: FeatureIDProto;
    expression?: string;
    error?: string;
    session: number;
    locked?: boolean;
};
