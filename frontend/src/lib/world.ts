import { FeatureIDProto, FeatureType } from '@/types/generated/api';

export const getWorldFeatureId = (
    worldId: string,
    namespace?: string,
    collection?: string
): FeatureIDProto => {
    const featureId = {
        type: 'FeatureTypeCollection' as unknown as FeatureType,
        namespace: namespace ?? 'diagonal.works',
        value: 0,
    };

    if (worldId === 'baseline') {
        const baselineValue = collection?.match(/(?<=\/)\d*$/)?.[0];
        featureId.value = parseInt(baselineValue ?? '0');
    } else {
        featureId.namespace = `${featureId.namespace}/scenario`;
        featureId.value = +worldId;
    }
    return featureId;
};
