import { useEffect, useMemo } from 'react';
import { useMap as useMapLibre } from 'react-map-gl/maplibre';
import { match } from 'ts-pattern';

import { World } from '@/stores/worlds';
import { FeatureIDsProto } from '@/types/generated/ui';

import { useMap } from './useMap';

/**
 * Highlight features on the map
 * @param world - The world the features are in
 * @param features - The features to highlight
 * @param visible - Whether the features are visible
 * @returns The features that are highlighted
 */
export const useHighlight = ({
    world,
    features,
    visible,
}: {
    world: World['id'];
    features?: FeatureIDsProto;
    visible?: boolean;
}) => {
    const { [world]: map } = useMapLibre();
    const [{ findFeatureInLayer, highlightFeature }] = useMap({ id: world });

    const geoJsonFeatures = useMemo(() => {
        if (!map || !features) return [];
        return (
            features.namespaces?.flatMap((ns, i) => {
                const nsType = ns.match(/(?<=^\/)[a-z]+(?=\/)/)?.[0];
                return match(nsType)
                    .with('path', () => {
                        return (
                            features.ids?.[i].ids?.flatMap((id) => {
                                const f = findFeatureInLayer({
                                    layer: 'road',
                                    filter: ['all'],
                                    id,
                                });
                                return f ? f : [];
                            }) ?? []
                        );
                    })
                    .with('area', () => {
                        return (
                            features.ids?.[i].ids?.flatMap((id) => {
                                const f = findFeatureInLayer({
                                    layer: 'building',
                                    filter: ['all'],
                                    id,
                                });
                                return f ? f : [];
                            }) ?? []
                        );
                    })
                    .otherwise(() => []);
            }) ?? []
        );
    }, [map, features, findFeatureInLayer]);

    useEffect(() => {
        geoJsonFeatures.forEach((feature) => {
            highlightFeature({ ...feature, highlight: !!visible });
        });

        return () => {
            try {
                geoJsonFeatures.forEach((feature) => {
                    highlightFeature({ ...feature, highlight: false });
                });
            } catch (e) {
                console.error(e);
            }
        };
    }, [geoJsonFeatures, highlightFeature, visible]);

    return [geoJsonFeatures];
};
