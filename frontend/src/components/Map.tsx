import { MinusIcon, PlusIcon } from "@radix-ui/react-icons";
import { scaleOrdinal } from "@visx/scale";
import { color } from "d3-color";
import { GeoJsonLayer, MVTLayer } from "deck.gl/typed";
import { Feature, MapLayerMouseEvent, StyleSpecification } from "maplibre-gl";
import "maplibre-gl/dist/maplibre-gl.css";
import { PropsWithChildren, useCallback, useMemo, useState } from "react";
import { Map as MapLibre, useMap as useMapLibre } from "react-map-gl/maplibre";
import { twMerge } from "tailwind-merge";
import { P, match } from "ts-pattern";

import basemapStyleRose from "@/assets/map/diagonal-map-style-rose.json";
import basemapStyle from "@/assets/map/diagonal-map-style.json";
import { DeckGLOverlay } from "@/components/DeckGLOverlay";
import { MapControls } from "@/components/system/MapControls";
import { useMap } from "@/hooks/useMap";
import { useMapStore } from "@/stores/map";
import { useViewStore } from "@/stores/view";
import { World } from "@/stores/worlds";
import colors from "@/tokens/colors.json";
import { colorToRgbArray, isColorHex } from "@/utils/colors";
import { changeMapStyleSource, getTileSource } from "@/utils/map";

const INITIAL_CENTER = { lat: 515361156 / 1e7, lng: -1255161 / 1e7 };

const COLLECTION_COLOR_SCALE = scaleOrdinal<string, string>({
	domain: Array.from({ length: 4 }, (_, i) => i.toString()),
	range: [
		colors.ice[30],
		colors.teal[20],
		colors.emerald[20],
		colors.green[10],
	],
});

export const Map = ({
	children,
	world,
	root,
	side,
}: {
	root: string;
	side: "left" | "right";
	world: World["id"];
} & PropsWithChildren) => {
	const view = useViewStore((state) => state.view);
	const { setView } = useViewStore((state) => state.actions);
	const [cursor, setCursor] = useState<"grab" | "pointer">("grab");
	const { [world]: maplibre } = useMapLibre();
	const [actions] = useMap({ id: world });
	const { geojson, highlight, tiles } = useMapStore((state) => state.layers);

	const mapStyle = useMemo(() => {
		const tileSource = getTileSource(root);
		const map = (
			side === "left" ? basemapStyle : basemapStyleRose
		) as StyleSpecification;
		return changeMapStyleSource(map, tileSource);
	}, [root, side]);

	const handleClick = useCallback(
		(e: MapLayerMouseEvent) => {
			// if shift key is pressed, create a new unlocked outliner.
			if (e.originalEvent.shiftKey) {
				actions.evaluateLatLng({ e, locked: false });
			} else {
				const features = maplibre?.queryRenderedFeatures(e.point);
				const feature = features?.[0];
				if (feature) {
					actions.evaluateFeature({ e, locked: true, feature });
				} else {
					actions.evaluateLatLng({ e, locked: true });
				}
			}
		},
		[actions, maplibre],
	);

	const geojsonGL = useMemo(() => {
		const data = Object.values(geojson)
			.filter((g) => g.world === world)
			.map((g) => g.features)
			.flat();
		return new GeoJsonLayer({
			data,
			id: "geojson",
			getFillColor: colorToRgbArray(side === "left" ? "#b1c5fd" : "#E2B79F"),
			getLineWidth: 0.5,
			getLineColor: colorToRgbArray(side === "left" ? "#37589f" : "#A66B4D"),
		});
	}, [geojson, side, world]);

	const collectionData = useMemo(() => {
		return Object.values(tiles)
			.filter((c) => c.world === world && c.type == "collection")
			.map((c) => c.spec);
	}, [tiles, world]);

	const collectionGL = useMemo(() => {
		return collectionData.flatMap((collection) => {
			if (!collection.showOnMap) return [];
			return [
				new MVTLayer({
					data: [collection.tiles],
					beforeId: "contour",
					id: `${world}-${collection.tiles}-bg`,
					getFillColor: (f: Feature) => {
						if (f.properties["b6:colour"]) {
							if (isColorHex(f.properties["b6:colour"])) {
								return colorToRgbArray(f.properties["b6:colour"]);
							}
							return colorToRgbArray(
								COLLECTION_COLOR_SCALE(f.properties["b6:colour"]),
							);
						}
						return [0, 0, 0, 0];
					},
					getLineColor: [0, 0, 0, 0],
					updateTriggers: {
						getFillColor: [collection.tiles],
					},
				}),
				new MVTLayer({
					data: [collection.tiles],
					beforeId: "road-label",
					id: `${world}-${collection.tiles}-stroke`,
					getFillColor: [0, 0, 0, 0],
					getLineWidth: 4,
					getLineColor: (f: Feature) => {
						if (f.properties["b6:colour"]) {
							if (isColorHex(f.properties["b6:colour"])) {
								return colorToRgbArray(f.properties["b6:colour"]);
							}
							return colorToRgbArray(
								COLLECTION_COLOR_SCALE(f.properties["b6:colour"]),
							);
						}
						return [0, 0, 0, 0];
					},
					updateTriggers: {
						getLineColor: [collection.tiles],
					},
				}),
			];
		});
	}, [collectionData, world]);

	const histogramData = useMemo(() => {
		const histogramLayers = Object.values(tiles).filter(
			(h) => h.world === world && h.type == "histogram",
		);
		return match(histogramLayers)
			.with(P.array({ type: "histogram" as const }), (hl) =>
				hl.map((h) => h.spec),
			)
			.otherwise(() => []);
	}, [tiles, world]);

	const histogramGL = useMemo(() => {
		const highlighted = Object.values(highlight)
			.filter((h) => h.world === world)
			.flatMap((h) => h.features);

		return histogramData.flatMap((hist) => {
			if (!hist.showOnMap) return [];
			return new MVTLayer({
				data: [hist.tiles],
				beforeId: "road-label",
				id: `${world}-${hist.tiles}`,
				getFillColor: (f: Feature) => {
					if (f.properties?.layerName === "background") {
						return [0, 0, 0, 0];
					}
					if (f.properties?.layerName === "histogram") {
						const c = hist.colorScale?.(f.properties.bucket);
						if (!c) {
							return [0, 0, 0, 0];
						}

						const isSelected =
							hist.selected && hist.selected.toString() === f.properties.bucket;

						return colorToRgbArray(
							c,
							hist?.selected ? (isSelected ? 255 : 155) : 255,
						);
					}
					return [0, 0, 0, 0];
				},
				getLineWidth: (f: Feature) => {
					if (f.properties?.layerName === "histogram") {
						/**
						 * skipping road highlighing for performance reasons
						 */
						/* const queryFeatures = map.querySourceFeatures(
                            'diagonal',
                            {
                                sourceLayer: 'road',
                                filter: ['all', ['==', 'id', f.properties.id]],
                            }
                        );

                        const feature = queryFeatures?.[0];

                        if (feature) {
                            return (
                                getRoadWidth(feature.properties?.highway) * 1.5
                            );
                        }
 */
						const isSelected =
							hist.selected && hist.selected.toString() === f.properties.bucket;

						return hist.selected ? (isSelected ? 0.8 : 0.2) : 0.5;
					}
					return 0;
				},
				getLineColor: (f: Feature) => {
					const isHighlighted = highlighted?.find((hf) => {
						return (
							hf.feature.properties.id &&
							hf.feature.properties.id.toString() === f.properties.id
						);
					});
					if (isHighlighted) {
						return [0, 0, 0, 255];
					}
					if (f.properties?.layerName === "histogram") {
						const c = hist.colorScale?.(f.properties?.bucket);
						if (!c) {
							return [0, 0, 0, 0];
						}
						const isSelected =
							hist.selected && hist.selected.toString() === f.properties.bucket;

						const darken = color(c)
							?.darker(isSelected ? 2 : 0.5)
							.formatRgb();

						return colorToRgbArray(
							darken ?? c,
							hist?.selected ? (isSelected ? 255 : 155) : 255,
						);
					}
					return [0, 0, 0, 0];
				},
				updateTriggers: {
					getLineColor: [hist.colorScale, hist.selected, highlight],
					getFillColor: [hist.colorScale, hist.selected, highlight],
					getLineWidth: [hist.selected],
				},
			});
		});
	}, [histogramData, highlight, world]);

	return (
		<MapLibre
			key={world}
			id={world}
			mapStyle={mapStyle}
			interactive={true}
			interactiveLayerIds={["building", "road"]}
			cursor={cursor}
			{...{
				...view,
				latitude: view.latitude ?? INITIAL_CENTER.lat,
				longitude: view.longitude ?? INITIAL_CENTER.lng,
				zoom: view.zoom ?? 16,
			}}
			onMove={(evt) => {
				setView(evt.viewState);
			}}
			onClick={handleClick}
			onMouseEnter={() => setCursor("pointer")}
			onMouseLeave={() => setCursor("grab")}
			antialias={true}
			attributionControl={false}
			dragRotate={false}
			boxZoom={false}
			style={{ width: "100%", height: "100%" }}
		>
			<DeckGLOverlay
				layers={[geojsonGL, histogramGL, collectionGL]}
				interleaved
			/>

			<MapControls className={twMerge(side === "right" && "right-0 left-auto")}>
				<MapControls.Button onClick={() => maplibre?.zoomIn({ duration: 200 })}>
					<PlusIcon />
				</MapControls.Button>
				<MapControls.Button
					onClick={() => maplibre?.zoomOut({ duration: 200 })}
				>
					<MinusIcon />
				</MapControls.Button>
			</MapControls>
			{children}
		</MapLibre>
	);
};
