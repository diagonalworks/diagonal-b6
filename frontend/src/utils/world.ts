import { FeatureIDProto, FeatureType } from '@/types/generated/api';

/**
 * Get the feature ID for a world.
 * @param worldId The string ID of the world.
 * @param namespace The namespace of the feature ID.
 * @param collection The collection of the feature ID.
 * @returns The feature ID.
 * @example
 * getWorldFeatureId('baseline', 'diagonal.works', 'skyline-demo-05-2024') // { type: 'FeatureTypeCollection', namespace: 'diagonal.works', value: 1 }
 */
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
