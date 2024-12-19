import {
	MapboxOverlay as DeckOverlay,
	MapboxOverlayProps,
} from "@deck.gl/mapbox";
import { useControl } from "react-map-gl/maplibre";

export function DeckGLOverlay(props: MapboxOverlayProps) {
	const overlay = useControl(() => new DeckOverlay(props));
	overlay.setProps(props);
	return null;
}
