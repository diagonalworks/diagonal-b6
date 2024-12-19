import { FeatureCollection } from "geojson";

import { FeatureIDProto } from "./generated/api";
import { UIResponseProto } from "./generated/ui";

export type LatLng = {
	latE7: number;
	lngE7: number;
};

export type Docked = {
	geoJSON: FeatureCollection[];
	proto: UIResponseProto;
};

export type StartupRequest = {
	z?: string;
	ll?: string;
	r?: string;
};

export type StartupResponse = {
	version?: string;
	docked?: Docked[];
	openDockIndex?: number;
	mapCenter?: LatLng;
	mapZoom?: number;
	root?: FeatureIDProto;
	expression?: string;
	error?: string;
	session: number;
	locked?: boolean;
};
