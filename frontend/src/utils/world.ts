import { FeatureIDProto, FeatureType } from "@/types/generated/api";

/**
 * Get the feature ID for a world.
 * @param worldId The string ID of the world.
 * @param namespace The namespace of the feature ID.
 * @param collection The collection of the feature ID.
 * @returns The feature ID.
 * @example
 * getWorldFeatureId('baseline', 'diagonal.works', 'skyline-demo-05-2024') // { type: 'FeatureTypeCollection', namespace: 'diagonal.works', value: 'baseline }
 */
export const getWorldFeatureId = ({
	namespace,
	type,
	value,
}: {
	namespace?: string;
	type?: string | FeatureType;
	value?: number;
}): FeatureIDProto => {
	const featureId = {
		type: (type ?? "FeatureTypeCollection") as unknown as FeatureType,
		namespace: namespace ?? "diagonal.works",
		value: value ?? 0,
	};

	return featureId;
};

export const getNamespace = (worldPath: string): string | undefined => {
	const namespace = worldPath.match(/(?<=\/)\w*(?=\/\w*$)/)?.[0];
	return namespace;
};

export const getValue = (worldPath: string): number | undefined => {
	const value = worldPath.match(/(?<=\/)\w*$/)?.[0];
	return value ? +value : undefined;
};

export const getType = (worldPath: string): string | undefined => {
	const type = worldPath.match(/(?<=^)\w*(?=\/)/)?.[0];
	return type;
};

export const getWorldPath = (featureId: FeatureIDProto): string => {
	return `/collection/${featureId.namespace}/${featureId.value}`;
};
