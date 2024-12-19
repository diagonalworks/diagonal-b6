export type BaseFeature = {
	properties: {
		layerName: string;
		id: string;
		ns: string;
	};
};

export type HighwayFeature = BaseFeature & {
	properties: {
		layerName: "road";
		highway?:
			| "motorway"
			| "trunk"
			| "primary"
			| "secondary"
			| "tertiary"
			| "street"
			| "unclassified"
			| "service"
			| "residential"
			| "cycleway"
			| "footway"
			| "path"
			| string;
		railway?: string;
		name?: string;
	};
};

export type BuildingFeature = BaseFeature & {
	properties: {
		layerName: "building";
		building:
			| "yes"
			| "residential"
			| "silo"
			| "apartments"
			| "university"
			| "commercial"
			| "university"
			| "office"
			| "disused_station"
			| "school"
			| "roof"
			| "vent_shaft"
			| "retail"
			| "train_station"
			| "civic"
			| "house"
			| "hotel"
			| "terrace"
			| "greenhouse"
			| "container"
			| "construction"
			| "church"
			| string;
	};
};

export type LanduseFeature = BaseFeature & {
	properties: {
		layerName: "landuse";
		landuse?:
			| "grass"
			| "forest"
			| "meadow"
			| "commercial"
			| "residential"
			| "industrial"
			| string;
		leasure?:
			| "park"
			| "playground"
			| "garden"
			| "pitch"
			| "nature_reserve"
			| string;
	};
};

export type BackgroundFeature = BaseFeature & {
	properties: {
		layerName: "background";
	};
};

export type BoundaryFeature = BaseFeature & {
	properties: {
		layerName: "boundary";
		natural?: "coastline" | "heath" | string;
	};
};

export type WaterFeature = BaseFeature & {
	properties: {
		layerName: "water";
		water?: "canal" | "basin" | "pond" | "waterfall" | string;
		waterway?: "lock_gate" | "water_point" | "sanitary_dump_station" | string;
	};
};

export type ContourFeature = BaseFeature & {
	properties: {
		layerName: "contour";
	};
};

export type LabelFeature = BaseFeature & {
	properties: {
		layerName: "label";
		name: string;
	};
};

export type Feature =
	| HighwayFeature
	| BuildingFeature
	| LanduseFeature
	| BackgroundFeature
	| BoundaryFeature
	| WaterFeature
	| ContourFeature
	| LabelFeature;
