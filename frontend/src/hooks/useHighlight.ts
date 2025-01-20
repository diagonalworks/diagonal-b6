import { useEffect, useMemo } from "react";
import { useMap as useMapLibre } from "react-map-gl/maplibre";
import { match } from "ts-pattern";

import { useMapStore } from "@/stores/map";
import { OutlinerSpec } from "@/stores/outliners";
import { FeatureIDsProto } from "@/types/generated/ui";

import { useMap } from "./useMap";

/**
 * Highlight features on the map
 * @param outliner - The outliner specification
 * @param features - The features to highlight
 * @returns The features that are highlighted
 */
export const useHighlight = ({
	outliner,
	features,
}: {
	outliner: OutlinerSpec;
	features?: FeatureIDsProto;
}) => {
	const { [outliner.world]: map } = useMapLibre();
	const [{ findFeatureInLayer, highlightFeature }] = useMap({
		id: outliner.world,
	});
	const { setHighlightLayer, removeHighlightLayer } = useMapStore(
		(state) => state.actions,
	);

	// These are the layers that we can enable highlighting in. That is, we
	// _must_ have a setting in `diagonal-map-style.json` corresponding to this
	// layer and it's `highlighted` status as well as the layer itself appearing
	// here.
	const highlightableLayers = ["building", "amenity"];

	const geoJsonFeatures = useMemo(() => {
		if (!map || !features) return [];
		return (
			features.namespaces?.flatMap((ns, i) => {
				const nsType = ns.match(/(?<=^\/)[a-z]+(?=\/)/)?.[0];
				return match(nsType)
					.with("path", () => {
						return (
							features.ids?.[i].ids?.flatMap((id) => {
								const f = findFeatureInLayer({
									layer: "road",
									filter: ["all"],
									id,
								});
								return f ? f : [];
							}) ?? []
						);
					})
					.with("area", () => {
						return (
							features.ids?.[i].ids?.flatMap((id) => {
								const fs = highlightableLayers.flatMap((layer) => {
									const f = findFeatureInLayer({ layer, filter: ["all"], id });
									return f ? f : [];
								});
								return fs;
							}) ?? []
						);
					})
					.otherwise(() => []);
			}) ?? []
		);
	}, [map, features, findFeatureInLayer]);

	useEffect(() => {
		geoJsonFeatures.forEach((feature) => {
			highlightFeature({
				...feature,
				highlight: !!outliner.properties.show,
			});
		});

		if (geoJsonFeatures.length > 0) {
			if (outliner.properties.show) {
				setHighlightLayer(outliner.id, {
					features: geoJsonFeatures,
					world: outliner.world,
				});
			} else {
				removeHighlightLayer(outliner.id);
			}
		}

		return () => {
			try {
				geoJsonFeatures.forEach((feature) => {
					highlightFeature({ ...feature, highlight: false });
				});
				removeHighlightLayer(outliner.id);
			} catch (e) {
				console.error(e);
			}
		};
	}, [geoJsonFeatures, highlightFeature, outliner]);

	return [geoJsonFeatures];
};
