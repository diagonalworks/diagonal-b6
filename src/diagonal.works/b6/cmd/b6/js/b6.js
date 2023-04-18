import * as d3 from "d3";

import {defaults as InteractionDefaults} from "ol/interaction";
import {fromLonLat} from "ol/proj";
import Circle from "ol/style/Circle";
import Fill from "ol/style/Fill";
import GeoJSONFormat from "ol/format/GeoJSON";
import Icon from "ol/style/Icon.js";
import Map from "ol/Map";
import MVT from "ol/format/MVT";
import Stroke from "ol/style/Stroke";
import Style from "ol/style/Style";
import Text from "ol/style/Text";
import VectorLayer from "ol/layer/Vector";
import VectorSource from "ol/source/Vector";
import VectorTileLayer from "ol/layer/VectorTile";
import VectorTileSource from "ol/source/VectorTile";
import View from "ol/View";
import Zoom from "ol/control/Zoom";

const InitialCenter = [-0.1255161, 51.5361156];
const InitalZoom = 16;
const RoadWidths = {
    "motorway": 36.0,
    "trunk": 36.0,
    "primary": 32.0,
    "secondary": 26.0,
    "tertiary": 26.0,
    "street": 18.0,
    "unclassified": 18.0,
    "service": 18.0,
    "residential": 18.0,
    "cycleway": 8.0,
    "footway": 8.0,
    "path": 8.0,
}

function scaleWidth(width, resolution) {
    return width * (0.30 / resolution);
}

function roadWidth(feature, resolution) {
    if (RoadWidths[feature.get("highway")]) {
        return scaleWidth(RoadWidths[feature.get("highway")], resolution);
    }
    return 0;
}

function waterwayWidth(resolution) {
    return scaleWidth(32.0, resolution);
}

function setupMap(state) {
    const zoom = new Zoom({
        zoomInLabel: "",
        zoomOutLabel: "",
    })

    const baseSource = new VectorTileSource({
        format: new MVT(),
        url: "/tiles/base/{z}/{x}/{y}.mvt",
        minZoom: 10,
        maxZoom: 16
    });

    var backgroundFill = new Style({
        fill: new Fill({color: "#eceff7"}),
    });


    const background = new VectorTileLayer({
        source: baseSource,
        style: function (feature, resolution) {
            if (feature.get("layer") == "background") {
                return backgroundFill;
            }
        }
    });

    const waterFill = new Style({
        fill: new Fill({color: "#b2bfe5"}),
    })

    const water = new VectorTileLayer({
        source: baseSource,
        style: function(feature, resolution) {
            if (feature.get("layer") == "water") {
                if (feature.getType() == "Polygon") {
                    return waterFill;
                } else {
                    const width = waterwayWidth(resolution);
                    if (width > 0) {
                        return new Style({
                            stroke: new Stroke({
                                color: "#b2bfe5",
                                width: width
                            })
                        });
                    }
                }
            }
        }
    });

    const parkFill = new Style({
        fill: new Fill({color: "#e1e1ee"}),
    });

    const meadowFill = new Style({
        fill: new Fill({color: "#dbdeeb"}),
    });

    const forestFill = new Style({
        fill: new Fill({color: "#c5cadd"}),
    });

    const contourStroke = new Style({
        stroke: new Stroke({
            color: "#e1e1ed",
            width: 1.0,
        }),
    });

    const landuse = new VectorTileLayer({
        source: baseSource,
        style: function(feature, resolution) {
            const landuse = feature.get("landuse");
            const leisure = feature.get("leisure");
            const natural = feature.get("natural");
            if (feature.get("layer") == "landuse") {
                if (landuse == "park" || landuse == "grass" || leisure == "pitch" || leisure == "park" || leisure == "garden") {
                    return parkFill;
                } else if (landuse == "forest") {
                    return forestFill;
                } else if (landuse == "meadow" || natural == "heath") {
                    return meadowFill;
                }
            } else if (feature.get("layer") == "contour") {
                return contourStroke;
            }
        },
    });

    const roadOutlines = new VectorTileLayer({
        source: baseSource,
        style: function(feature, resolution) {
            if (feature.get("layer") == "road") {
                const width = roadWidth(feature, resolution);
                if (width > 0) {
                    return new Style({
                        stroke: new Stroke({
                            color: "#9aa4cc",
                            width: width + 2.0,
                        })
                    });
                }
            }
        },
    });

    const roadFills = new VectorTileLayer({
        source: baseSource,
        style: function(feature, resolution) {
            if (feature.get("layer") == "road") {
                const width = roadWidth(feature, resolution);
                if (width > 0) {
                    return new Style({
                        stroke: new Stroke({
                            color: "#e1e1ee",
                            width: width
                        })
                    });
                }
            }
        },
    });

    const buildingFill = new Style({
        fill: new Fill({color: "#ffffff"}),
        stroke: new Stroke({color: "#4f5a7d", width: 0.3})
    });

    const buildings = new VectorTileLayer({
        source: baseSource,
        style: function(feature, resolution) {
            if (feature.get("layer") == "building") {
                return buildingFill;
            }
        },
    });

    const labels = new VectorTileLayer({
        source: baseSource,
        style: function(feature, resolution) {
            if (feature.get("layer") == "label") {
                return new Style({
                    text: new Text({
                        text: feature.get("name"),
                        textAlign: "left",
                        offsetX: 6,
                        offsetY: 1,    
                        fill: new Fill({
                            color: "#000000",
                        }),
                    }),
                    image: new Circle({
                        radius: 2,
                        fill: new Fill({
                            color: "#000000",
                        }),
                    }),
                });
            }
        },
    });

    const view = new View({
        center: fromLonLat(InitialCenter),
        zoom: InitalZoom,
    });

    const map = new Map({
        target: "map",
        layers: [background, water, landuse, roadOutlines, roadFills, buildings, labels],
        interactions : InteractionDefaults(),
        controls: [zoom],
        view: view,
    });

    return map;
}

function main() {
    const state = {};
    setupMap(state);
}

main();