import { FeatureIDProto } from '@/types/generated/api';
import { LabelledIconProto } from '@/types/generated/ui';

export type ChangeFeature = {
    id: FeatureIDProto;
    label?: LabelledIconProto;
    expression: string;
};

export type ChangeFunction = {
    label?: string;
    id: FeatureIDProto;
};

export type ChangeSpec = {
    analysis?: FeatureIDProto;
    features: ChangeFeature[];
    changeFunction?: ChangeFunction;
};
