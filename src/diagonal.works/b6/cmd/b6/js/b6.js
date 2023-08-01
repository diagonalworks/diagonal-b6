import * as d3 from "d3";

import Shell from "./shell.js";

import {defaults as InteractionDefaults} from "ol/interaction";
import {fromLonLat, toLonLat} from "ol/proj";
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
const GeoJSONFillColour = "#364153";

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

function newGeoJSONStyle(state, styles) {
    const point = new Style({
        image: new Circle({
            radius: 4,
            stroke: new Stroke({
                color: styles["geojson-point"]["stroke"],
                width: +styles["geojson-point"]["stroke-width"].replace("px", ""),
            }),
            fill: new Fill({
                color: styles["geojson-point"]["fill"],
            }),
        }),
    });

    const path = new Style({
        stroke: new Stroke({
            color: styles["geojson-path"]["stroke"],
            width: +styles["geojson-path"]["stroke-width"].replace("px", ""),
        })
    });

    const area = new Style({
        stroke: new Stroke({
            color: styles["geojson-area"]["stroke"],
            width: +styles["geojson-area"]["stroke-width"].replace("px", ""),
        }),
        fill: new Fill({
            color: styles["geojson-area"]["fill"],
        })
    })

    return function(feature, resolution) {
        var s = point;
        switch (feature.getGeometry().getType()) {
            case "LineString":
            case "MultiLineString":
                s = path;
            case "Polygon":
            case "MultiPolygon":
                s = area;
        }
        var cloned = false;
        const label = feature.get("name") || feature.get("label");
        if (label) {
            s = s.clone();
            cloned = true;
            s.setText(new Text({
                text: label,
                textAlign: "left",
                offsetX: 6,
                offsetY: 1,
                font: '"Roboto" 12px',
            }));
        }
        if (feature.get("-diagonal-stroke")) {
            if (!cloned) {
                s = s.clone();
                cloned = true;
            }
            if (s.getStroke()) {
                s.getStroke().setColor(feature.get("-diagonal-stroke"));
            }
        }
        if (feature.get("-diagonal-fill")) {
            if (!cloned) {
                s = s.clone();
                cloned = true;
            }
            if (s.getFill()) {
                s.getFill().setColor(feature.get("-diagonal-fill"));
            }
        }
        return s;
    }
}

function setupMap(state, styles) {
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

    const boundaries = new VectorTileLayer({
        source: baseSource,
        style: function(feature, resolution) {
            if (feature.get("layer") == "boundary") {
                if (state.featureColours) {
                    const colour = state.featureColours[idKeyFromFeature(feature)];
                    if (colour) {
                        return new Style({
                            fill: new Fill({color: colour}),
                            stroke: new Stroke({color: "#4f5a7d", width: 0.3})
                        });
                    }
                }
            }
        },
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
                    const id = idKeyFromFeature(feature);
                    if (state.featureColours) {
                        const colour = state.featureColours[id];
                        if (colour) {
                            return new Style({
                                stroke: new Stroke({
                                    color: colour,
                                    width: width
                                })
                            });
                        }
                    }
                    if (state.highlighted[id]) {
                        return new Style({
                            stroke: new Stroke({
                                color: styles["highlighted-road-fill"]["stroke"],
                                width: width
                            })
                        });
                    } else {
                        return new Style({
                            stroke: new Stroke({
                                color: styles["road-fill"]["stroke"],
                                width: width
                            })
                        });
                    }
                }
            }
        },
    });

    const buildingFill = new Style({
        fill: new Fill({color: "#ffffff"}),
        stroke: new Stroke({color: "#4f5a7d", width: 0.3})
    });

    const highlightedBuildingFill = new Style({
        fill: new Fill({color: styles["highlighted-area"]["fill"]}),
        stroke: new Stroke({color: styles["highlighted-area"]["stroke"], width: 0.3})
    });

    const buildings = new VectorTileLayer({
        source: baseSource,
        style: function(feature, resolution) {
            if (feature.get("layer") == "building") {
                const id = idKeyFromFeature(feature);
                if (state.highlighted[id]) {
                    return highlightedBuildingFill;
                }
                return buildingFill;
            }
        },
    });

    const points = new VectorTileLayer({
        source: baseSource,
        style: function(feature, resolution) {
            if (feature.get("layer") == "point") {
                if (state.featureColours) {
                    const colour = state.featureColours[idKeyFromFeature(feature)];
                    if (colour) {
                        return new Style({
                            image: new Circle({
                                radius: 2,
                                fill: new Fill({
                                    color: colour,
                                }),
                            }),
                        });
                    }
                }
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
        layers: [background, boundaries, water, landuse, roadOutlines, roadFills, buildings, points, labels],
        interactions: InteractionDefaults(),
        controls: [zoom],
        view: view,
    });

    const searchableLayers = [buildings, roadOutlines, landuse, water];

    const highlightChanged = () => {
        boundaries.changed();
        buildings.changed();
        roadFills.changed();
        points.changed();
    };

    return [map, searchableLayers, highlightChanged];
}

function lonLatToLiteral(ll) {
    return `${ll[1].toPrecision(8)}, ${ll[0].toPrecision(8)}`
}

function showFeature(feature, blocks) {
    const ns = feature.get("ns");
    const id = feature.get("id");
    const types = {"Point": "point", "LineString": "path", "Polygon": "area", "MultiPolygon": "area"};
    if (ns && id && types[feature.getType()]) {
        const request = {
            method: "POST",
            body: JSON.stringify({Expression: `find-feature /${types[feature.getType()]}/${ns}/${BigInt("0x" + id)}`}),
            headers: {
                "Content-type": "application/json; charset=UTF-8"
            }
        }
        d3.json("/block", request).then(response => {
            blocks.renderBlocks(response);
        });
    }
}

const BlockRenderers = {
    "pipeline-stage": renderPipelineStageBlock,
    "feature": renderFeatureBlock,
    "int": renderIntBlock,
    "float": renderFloatBlock,
    "string": renderStringBlock,
    "int-result": renderIntResultBlock,
    "float-result": renderFloatResultBlock,
    "string-result": renderStringResultBlock,
    "area": renderAreaBlock,
    "path": renderPathBlock,
    "geojson-feature-collection": renderGeoJSONFeatureCollectionBlock,
    "geojson-feature": renderGeoJSONFeatureBlock,
    "title-count": renderTitleCountBlock,
    "collection": renderCollectionBlock,
    "collection-feature": renderCollectionFeatureBlock,
    "collection-key-value": renderCollectionKeyValueBlock,
    "collection-feature-key": renderCollectionFeatureKeyBlock,
    "collection-key-or-value": renderCollectionKeyOrValueBlock,
    "shell": renderShellBlock,
    "error": renderStringBlock,
    "placeholder": renderPlaceholderBlock,
}

const BlocksOrigin = [10, 100];

function elementPosition(element) {
    return [+element.style("left").replace("px", ""), +element.style("top").replace("px", "")];
}

function lookupGeoJSONStyles() {
    const palette = d3.select("body").selectAll(".geojson-palette").data([1]).join(
        enter => {
            const palette = enter.append("svg").classed("geojson-palette", true);
            palette.append("g").classed("geojson-palette-point", true);
            palette.append("g").classed("geojson-palette-path", true);
            palette.append("g").classed("geojson-palette-area", true);
            return palette;
        },
    );
    const pointStyle = getComputedStyle(palette.select(".geojson-palette-point").node());
    const pathStyle = getComputedStyle(palette.select(".geojson-palette-path").node());
    const areaStyle = getComputedStyle(palette.select(".geojson-palette-area").node());
    return [pointStyle, pathStyle, areaStyle];
}

function lookupStyles(names) {
    const palette = d3.select("body").selectAll(".palette").data([1]).join("g");
    palette.classed("palette", true);
    const items = palette.selectAll("g").data(names).join("g");
    items.attr("class", d => d);
    const styles = {};
    for (const i in names) {
        styles[names[i]] = getComputedStyle(palette.select("." + names[i]).node());
    }
    return styles;
}

class RenderedResponse {
    constructor(response, blocks) {
        if (response.Highlighted) {
            this.highlighted = response.Highlighted;
            for (const i in this.highlighted) {
                const values = this.highlighted[i];
                for (const j in values) {
                    blocks.addHighlight(i + "/" + values[j])
                }
            }
        }
        this.layers = []
        if (response.QueryLayers) {
            for (const i in response.QueryLayers) {
                this.layers.push(blocks.addQueryLayer(response.QueryLayers[i]));
            }
        }
    }

    redrawHighlights() {
        for (const i in this.layers) {
            this.layers[i].changed();
        }
    }

    remove(blocks) {        
        for (const i in this.layers) {
            blocks.removeLayer(this.layers[i]);
        }
        if (this.highlighted) {
            for (const i in this.highlighted) {
                const values = this.highlighted[i];
                for (const j in values) {
                    blocks.removeHighlight(i + "/" + values[j])
                }
            }
        }
        for (const i in this.layers) {
            blocks.removeLayer(this.layers[i]);
        }
    }
}

class RenderedBlock {
    constructor(block, map, geojsonStyle) {
        if (block.GeoJSON) {
            const source = new VectorSource({
                features: [],
             })
             const layer = new VectorLayer({
                 source: source,
                 style: geojsonStyle,
             })
            const features = new GeoJSONFormat().readFeatures(block.GeoJSON, {
                dataProjection: "EPSG:4326",
                featureProjection: map.getView().getProjection(),
            });
            source.addFeatures(features);
            this.layer = layer;
            map.addLayer(this.layer);

            const blob = new Blob([JSON.stringify(block.GeoJSON, null, 2)], {
                type: "application/json",
            });
            this.blobURL = URL.createObjectURL(blob);
        }
    }

    remove(map) {
        if (this.layer) {
            map.removeLayer(this.layer);
        }
        if (this.blobURL) {
            URL.revokeObjectURL(this.blobURL);
        }
    }
}

class Blocks {
    constructor(map, state, queryStyle, geojsonStyle, highlightChanged) {
        this.map = map;
        this.state = state;
        this.queryStyle = queryStyle;
        this.geojsonStyle = geojsonStyle;
        this.basemapHighlightChanged = highlightChanged;
        this.dragging = null;
        this.html = d3.select("html");
        this.dragPointerOrigin = [0,0];
        this.dragElementOrigin = [0,0];
        this.rendered = [];
        this.needHighlightRedraw = false;
    }

    evaluateExpression(expression) {
        this.evaluateExpressionInContext(null, expression);
    }

    evaluateNode(node) {
        this.evaluateExpressionInContext(node, null);
    }

    evaluateExpressionInContext(node, expression) {
        const body = JSON.stringify({Node: node, Expression: expression});
        const request = {
            method: "POST",
            body: body,
            headers: {
                "Content-type": "application/json; charset=UTF-8"
            }
        }
        d3.json("/block", request).then(response => {
            this.renderBlocks(response);
        });
    }

    addHighlight(idKey) {
        if (this.state.highlighted[idKey]) {
            this.state.highlighted[idKey]++;
        } else {
            this.state.highlighted[idKey] = 1;
        }
        this.needHighlightRedraw = true;
    }

    removeHighlight(idKey) {
        if (--this.state.highlighted[idKey] == 0) {
            delete this.state.highlighted[idKey];
        }
        this.needHighlightRedraw = true;
    }

    redrawHighlights() {
        for (const i in this.rendered) {
            this.rendered[i].redrawHighlights();
        }
        this.basemapHighlightChanged();
    }

    addQueryLayer(query) {
        const params = new URLSearchParams({"q": query});
        const source = new VectorTileSource({
            format: new MVT(),
            url: "/tiles/query/{z}/{x}/{y}.mvt?" + params.toString(),
            minZoom: 14,
        });
        const layer = new VectorTileLayer({
            source: source,
            style: this.queryStyle,
        });
        this.map.addLayer(layer);
        return layer;
    }

    removeLayer(layer) {
        this.map.removeLayer(layer);
    }

    renderBlocks(response) {
        response.Blocks.push({Type: "shell"});
        const root = d3.select("body").selectAll(".featured-blocks").data([1]).join("div");        
        root.attr("class", "featured-blocks blocks");
        root.style("left",  `${BlocksOrigin[0]}px`);
        root.style("top", `${BlocksOrigin[1]}px`);
        const blocks = this;
        const f = function(d) {
            if (this.__rendered__) {
                this.__rendered__.remove(blocks);
                blocks.rendered = blocks.rendered.filter(r => r != this.__rendered__);
            }
            this.__rendered__ = new RenderedResponse(response, blocks);
            blocks.rendered.push(this.__rendered__);
        }
        root.each(f);
        const divs = root.selectAll(".block").data(response.Blocks).join("div");
        divs.attr("class", "block");
        this.renderBlock(divs, root, response);
        if (this.needHighlightRedraw) {
            this.redrawHighlights();
            this.needHighlightRedraw = false;            
        }
    }

    renderBlock(target, root, response) {
        const blocks = this;
        const divs = target.selectAll(".block-container").data(d => [d], d => d.Type).join(
            enter => {
                const div = enter.append("div");
                div.attr("class", d => `block-container ${d.Type}`);
                return div;
            },
            update => {
                return update;
            },
            exit => {
                exit.each(function() {
                    if (this.__rendered__) {
                        this.__rendered__.remove(blocks.map);
                        delete this.__rendered__;
                    }
                });
                exit.remove();
            }
        )
        const f = function(d) {
            if (this.__rendered__) {
                this.__rendered__.remove(blocks.map);
            }
            const rendered = new RenderedBlock(d, blocks.map, blocks.geojsonStyle);
            this.__rendered__ = rendered;
            if (d.MapCenter) {
                blocks.map.getView().animate({
                    center: fromLonLat([parseFloat(d.MapCenter[0]), parseFloat(d.MapCenter[1])]),
                    duration: 500,
                });
            }
            if (BlockRenderers[d.Type]) {
                BlockRenderers[d.Type].apply(null, [d3.select(this), root, response, blocks, rendered]);
            } else {
                throw new Error(`No renderer for block ${d.Type}`);
            }
        }
        divs.each(f);
    }

    handleDragStart(event, root) {
        event.preventDefault();
        if (root.classed("featured-blocks")) {
            root.attr("class", "blocks");
        }
        this.dragging = root;
        this.dragging.classed("dragging", true);
        this.dragPointerOrigin = d3.pointer(event, this.html);
        this.dragElementOrigin = elementPosition(root);
    }

    handlePointerMove(event) {
        if (this.dragging) {
            event.preventDefault();
            const pointer = d3.pointer(event, this.html);
            const delta = [pointer[0]-this.dragPointerOrigin[0], pointer[1]-this.dragPointerOrigin[1]];
            const left = Math.round(this.dragElementOrigin[0]+delta[0]);
            const top = Math.round(this.dragElementOrigin[1]+delta[1]);
            this.dragging.style("left", `${left}px`);
            this.dragging.style("top", `${top}px`);
        }
    }

    handleDragEnd(event) {
        if (this.dragging) {
            event.preventDefault();
            this.dragging.classed("dragging", false);
            this.dragging = null;
        }
    }
}

function renderPipelineStageBlock(block, root, response, blocks) {
    block.classed("top", true);
    block.text(d => `${d.Expression}`);
    block.on("mousedown", e => {
        blocks.handleDragStart(e, root);
    });
}

function renderPlaceholderBlock(block, root, response, blocks) {
    block.classed("top", true);
    block.text(d => `${d.RawValue}`);
}

function renderFeatureBlock(block, root, response, blocks) {
    const ul = block.selectAll(".tags").data(d => [d]).join("ul").attr("class", "tags top");
    const formatTags = t => [
        {class: "prefix", text: t.Prefix},
        {class: "key", text: t.Key, prefix: t.Prefix, key: t.Key, value: t.Value},
        {class: "value", text: t.Value},
    ];
    const li = ul.selectAll("li").data(d => d.Tags.map(formatTags)).join("li");
    li.selectAll("span").data(d => d).join("span").attr("class", d => d.class).text(d => d.text);
    const clickableKeys = ul.selectAll(".key").data(d => d.Tags).filter(d => d.KeyExpression);
    clickableKeys.classed("clickable", true).on("click", (e, d) => {
        e.preventDefault();
        blocks.evaluateNode(d.KeyExpression);
    });
    const clickableValues = ul.selectAll(".value").data(d => d.Tags).filter(d => d.ValueExpression);
    clickableValues.classed("clickable", true).on("click", (e, d) => {
        e.preventDefault();
        blocks.evaluateNode(d.ValueExpression);
    });

    const lastPoint = renderFeatureBlockPoints(block, root, response, blocks);
    if (!lastPoint.empty()) {
        block.select(".points").classed("points-last", true);
        lastPoint.classed("top-last", true);
    } else {
        ul.classed("top-last", true);
    }
}

function renderFeatureBlockPoints(block, root, response, blocks) {
    const points = block.selectAll(".points").data(d => d.Points && d.Points.length > 0 ? [d] : []).join(
        enter => {
            const div = enter.append("div");
            div.attr("class", "points");
            div.append("div").classed("title", true);
            div.append("ul");
            return div
        }
    );
    const title = points.select(".title").datum(d => {return {Type: "title-count", Title: "Points", Count: d.Points.length};});
    blocks.renderBlock(title, root, response);
    title.on("click", e => {
        e.preventDefault();
        points.classed("open", !points.classed("open"));
    });
    const li = points.select("ul").selectAll("li").data(d => d.Points).join("li");
    blocks.renderBlock(li, root, response);
    return li.filter((d, i, l) => i == l.length - 1).select(".top");
}

function renderIntBlock(block, root, response, blocks) {
    block.classed("top", true);
    block.text(d => `${d.Value}`);
}

function renderIntResultBlock(block, root, response, blocks) {
    renderIntBlock(block, root, response, blocks)
    block.classed("top-last", true);
}

function renderFloatBlock(block, root, response, blocks) {
    block.classed("top", true);
    block.text(d => `${d3.format(".2f")(d.Value)}`);
}

function renderFloatResultBlock(block, root, response, blocks) {
    renderFloatBlock(block, root, response, blocks)
    block.classed("top-last", true);
}

function renderStringBlock(block, root, response, blocks) {
    block.classed("top", true);
    block.classed("top-last", true);
    block.text(d => `${d.Value}`);
}

function renderStringResultBlock(block, root, response, blocks) {
    renderStringBlock(block, root, response, blocks)
    block.classed("top-last", true);
    block.on("mousedown", e => {
        blocks.handleDragStart(e, root);
    });
}

function renderAreaBlock(block, root, response, blocks) {
    renderGeometryBlock("area", "m²", block, root, response, blocks);
}

function renderPathBlock(block, root, response, blocks) {
    renderGeometryBlock("path", "m", block, root, response, blocks);
}

function renderGeometryBlock(geometry, units, block, root, response, blocks) {
    block.classed("top", true);
    block.classed("top-last", true);
    block.classed("geometry", true);
    const spans = block.selectAll("span").data(d => [
        {class: `icon icon-${geometry}`, text: ""},
        {class: "", text: `${d3.format(".2f")(d.Dimension)}${units}`},
    ]);
    spans.join("span").attr("class", d => d.class).text(d => d.text);
}

function renderGeoJSONFeatureCollectionBlock(block, root, response, blocks, rendered) {
    block.classed("top", true);
    block.classed("top-last", true);
    block.classed("geometry", true);
    const spans = block.selectAll("span").data(d => [
        {class: `icon icon-area`},
        {class: "link"},
    ]);
    spans.join("span").attr("class", d => d.class);
    const a = renderGeoJSONBlobLink(block.select(".link"), rendered);
    a.text(d => `${d3.format("i")(d.Dimension)} GeoJSON ${d.Dimension == 1 ? "feature" : "features"}`);
}

function renderGeoJSONFeatureBlock(block, root, response, blocks, rendered) {
    block.classed("top", true);
    block.classed("top-last", true);
    block.classed("geometry", true);
    const spans = block.selectAll("span").data(d => [
        {class: `icon icon-area`},
        {class: "link"},
    ]);
    spans.join("span").attr("class", d => d.class).text(d => d.text);
    const a = renderGeoJSONBlobLink(block.select(".link"), rendered);
    a.text("GeoJSON feature");
}

function renderGeoJSONBlobLink(target, rendered) {
    const a = target.selectAll("a").data(d => [d]).join("a");
    a.node().href = rendered.blobURL;
    a.node().download = "b6-result.geojson";
    return a;
}

function renderTitleCountBlock(block, root, response, blocks) {
    block.classed("top", true);
    const spans = block.selectAll("span").data(d => [
        {class: "title", text: d.Title},
        {class: "count", text: d.Count >= 0 ? `${d.Count}` : ""},
    ]);
    spans.join("span").attr("class", d => d.class).text(d => d.text);
}

function renderCollectionBlock(block, root, response, blocks) {
    const title = block.selectAll(".title").data(d => [d.Title]).join("div").classed("title", true);
    blocks.renderBlock(title, root, response);
    const ul = block.selectAll("ul").data(d => [d]).join("ul");
    const li = ul.selectAll("li").data(d => d.Items).join("li");
    blocks.renderBlock(li, root, response);
    li.filter((d, i, l) => i == l.length - 1).select(".top").classed("top-last", true);
}

function renderCollectionFeatureBlock(block, root, response, blocks) {
    block.classed("top", true);
    const spans = block.selectAll("span").data(d => [
        {class: `icon icon-${d.Icon}`, text: ""},
        {class: "label", text: d.Label},
        {class: "namespace", text: d.Namespace},
        {class: "id", text: d.ID},
    ]);
    spans.join("span").attr("class", d => d.class).text(d => d.text);
    block.filter(d => d.Expression).classed("clickable", true).on("click", (e, d) => {
        e.preventDefault();
        blocks.evaluateNode(d.Expression);
    });
}

function renderCollectionKeyValueBlock(block, root, response, blocks) {
    const spans = block.selectAll("span").data(d => [d.Key, d.Value]).join("span");
    spans.attr("class", (d, i) => ["key", "value"][i]);
    blocks.renderBlock(spans);
}

function renderCollectionFeatureKeyBlock(block, root, response, blocks) {
    block.classed("top", true);
    const spans = block.selectAll("span").data(d => [
        {class: `icon icon-${d.Icon}`, text: ""},
        {class: "namespace", text: d.Namespace},
        {class: "id", text: d.ID},
    ]);
    spans.join("span").attr("class", d => d.class).text(d => d.text);
    block.filter(d => d.Expression).classed("clickable", true).on("click", (e, d) => {
        e.preventDefault();
        blocks.evaluateNode(d.Expression);
    });
}

function renderCollectionKeyOrValueBlock(block, root, response, blocks) {
    block.classed("top", true);
    block.text(d => d.Value);
    block.filter(d => d.Expression).classed("clickable", true).on("click", (e, d) => {
        e.preventDefault();
        blocks.evaluateNode(d.Expression);
    });
}

const TerminalMaxSuggestions = 6;

function renderShellBlock(block, root, response, blocks) {
    block.classed("top", true);
    const input = block.selectAll(".shell").data([1]).join(
        enter => {
            const form = enter.append("form").attr("class", "shell");
            form.append("div").attr("class", "prompt").text("b6");
            form.append("input").attr("type", "text");
            return form
        },
        update => update,
        exit => exit.remove(),
    );
    const state = {suggestions: response.Functions ? response.Functions.toSorted() : [], highlighted: 0};
    input.select("input").on("focusin", e => {
        block.select("ul").classed("focussed", true);
    });
    input.select("input").on("focusout", e => {
        block.select("ul").classed("focussed", false);
    });
    input.on("keydown", e => {
        switch (e.key) {
            case "Tab":
                const node = input.select("input").node();
                if (state.highlighted >= 0 && state.filtered[state.highlighted].length > node.value.length) {
                    node.value = state.filtered[state.highlighted] + " ";
                }
                e.preventDefault();
                break;
        }
    });
    input.on("keyup", e => {
        switch (e.key) {
            case "ArrowUp":
                state.highlighted--;
                e.preventDefault();
                break;
            case "ArrowDown":
                state.highlighted++;
                e.preventDefault();
                break;
            default:
                state.highlighted = 0;
                break;
        }
        renderShellSuggestions(block, state);
    });
    input.on("submit", e => {
        e.preventDefault();
        acceptTerminalSuggestion(block, state, response, blocks);
        return;
    });
    renderShellSuggestions(block, state);
}

function renderShellSuggestions(block, state) {
    const entered = block.select("input").node().value;
    const filtered = state.suggestions.filter(s => s.startsWith(entered));
    state.filtered = filtered;
    const suggestions = filtered.slice(0, TerminalMaxSuggestions).map(s => {return {text: s, highlighted: false}});
    if (state.highlighted < 0) {
        state.highlighted = 0
    } else if (state.highlighted >= filtered.length) {
        state.highlighted = filtered.length - 1;
    }
    if (state.highlighted >= 0) {
        suggestions[state.highlighted].highlighted = true;
    }
    const ul = block.selectAll("ul").data([1]).join("ul");
    const li = ul.selectAll("li").data(suggestions).join("li");
    li.text(d => d.text).classed("highlighted", d => d.highlighted);
}

function acceptTerminalSuggestion(block, state, response, blocks) {
    var expression = block.select("input").node().value;
    if (state.highlighted >= 0 && state.filtered[state.highlighted].length > expression.length) {
        expression = state.filtered[state.highlighted];
    }
    blocks.evaluateExpressionInContext(response.Node, expression);
}

function showFeatureAtPixel(pixel, layers, blocks) {
    const search = (i, found) => {
        if (i < layers.length) {
            if (layers[i].getVisible()) {
                layers[i].getFeatures(pixel).then(features => {
                    if (features.length > 0) {
                        found(features[0]);
                        return
                    } else {
                        search(i + 1, found);
                    }
                });
            } else {
                search(i + 1, found);
            }
        }
    };
    search(0, f => showFeature(f, blocks));
}

function idKey(id) {
    return `/${id[0]}/${id[1]}/${id[2]}`;
}

const idGeometryTypes = {
    "Point": "point",
    "LineString": "path",
    "Polygon": "area",
    "MultiPolygon": "area",
}

function idKeyFromFeature(feature) {
    const type = idGeometryTypes[feature.getGeometry().getType()] || "invalid";
    return `/${type}/${feature.get("ns")}/${feature.get("id")}`
}

function setupShell(target, blocks) {
    target.selectAll("form").data([1]).join(
        enter => {
            const form = enter.append("form").attr("class", "shell");
            form.append("div").attr("class", "prompt").text("b6");
            form.append("input").attr("type", "text");
            return form;
        },
        update => {
            return update;
        },
    );
    const state = {history: [], index: 0};
    d3.select("body").on("keydown", (e) => {
        if (e.key == "`" || e.key == "~") {
            e.preventDefault();
            target.classed("closed", !target.classed("closed"));
            target.select("input").node().focus();
        } else if (e.key == "ArrowUp") {
            e.preventDefault();
            if (state.index < state.history.length) {
                state.index++;
                target.select("input").node().value = state.history[state.history.length - state.index];
            }
        } else if (e.key == "ArrowDown") {
            e.preventDefault();
            if (state.index > 0) {
                state.index--;
                if (state.index == 0) {
                    target.select("input").node().value = "";
                } else {
                    target.select("input").node().value = state.history[state.history.length - state.index];
                }
            }
        }
    });
    target.select("form").on("submit", (e) => {
        e.preventDefault();
        target.classed("closed", true);
        const expression = target.select("input").node().value;
        state.history.push(expression);
        state.index = 0;
        blocks.evaluateExpression(expression);
        target.select("input").node().value = "";
    })
}

function newQueryStyle(state, styles) {
    const point = new Style({
        image: new Circle({
            radius: 4,
            stroke: new Stroke({
                color: styles["query-point"]["stroke"],
                width: +styles["query-point"]["stroke-width"].replace("px", ""),
            }),
        }),
    });

    const highlightedPoint = new Style({
        image: new Circle({
            radius: 4,
            stroke: new Stroke({
                color: styles["highlighted-point"]["stroke"],
                width: +styles["highlighted-point"]["stroke-width"].replace("px", ""),
            }),
            fill: new Fill({
                color: styles["highlighted-point"]["fill"],
            }),
        }),
    });

    const path = new Style({
        stroke: new Stroke({
            color: styles["query-path"]["stroke"],
            width: +styles["query-path"]["stroke-width"].replace("px", ""),
        })
    });

    const highlightedPath = new Style({
        stroke: new Stroke({
            color: styles["highlighted-path"]["stroke"],
            width: +styles["highlighted-path"]["stroke-width"].replace("px", ""),
        })
    });

    const area = new Style({
        stroke: new Stroke({
            color: styles["query-area"]["stroke"],
            width: +styles["query-area"]["stroke-width"].replace("px", ""),
        }),
        fill: new Fill({
            color: styles["query-area"]["fill"],
        })
    })

    const highlightedArea = new Style({
        stroke: new Stroke({
            color: styles["highlighted-area"]["stroke"],
            width: +styles["highlighted-area"]["stroke-width"].replace("px", ""),
        }),
        fill: new Fill({
            color: styles["highlighted-area"]["fill"],
        })
    })

    const boundary = new Style({
        stroke: new Stroke({
            color: styles["query-area"]["stroke"],
            width: +styles["query-area"]["stroke-width"].replace("px", ""),
        }),
    })

    const highlightedBoundary = new Style({
        stroke: new Stroke({
            color: styles["highlighted-area"]["stroke"],
            width: +styles["highlighted-area"]["stroke-width"].replace("px", ""),
        }),
    })


    return function(feature, resolution) {
        if (feature.get("layer") != "background") {
            const id = idKeyFromFeature(feature);
            const highlighted = state.highlighted[id];
            switch (feature.getGeometry().getType()) {
                case "Point":
                    return highlighted ? highlightedPoint : point;
                case "LineString":
                    return highlighted ? highlightedPath : path;
                case "MultiLineString":
                    return highlighted ? highlightedPath : path;
                case "Polygon":
                    if (feature.get("boundary")) {
                        return highlighted ? highlightedBoundary : boundary;
                    } else {
                        return highlighted ? highlightedArea : area;
                    }
                case "MultiPolygon":
                    if (feature.get("boundary")) {
                        return highlighted ? highlightedBoundary : boundary;
                    } else {
                        return highlighted ? highlightedArea : area;
                    }
            }
        }
    }
}

const Styles = [
    "road-fill",
    "highlighted-road-fill",
    "highlighted-point",
    "highlighted-path",
    "highlighted-area",
    "geojson-point",
    "geojson-path",
    "geojson-area",
    "query-point",
    "query-path",
    "query-area",
];

function setup(bootstrapResponse) {
    const state = {highlighted: {}};
    const styles = lookupStyles(Styles);
    const [map, searchableLayers, highlightChanged] = setupMap(state, styles);
    const queryStyle = newQueryStyle(state, styles);
    const geojsonStyle = newGeoJSONStyle(state, styles);
    const blocks = new Blocks(map, state, queryStyle, geojsonStyle, highlightChanged);
    const html = d3.select("html");
    html.on("pointermove", e => {
        blocks.handlePointerMove(e);
    });
    html.on("mouseup", e => {
        blocks.handleDragEnd(e);
    });

    setupShell(d3.select("#shell"), blocks);

    map.on("singleclick", e => {
        if (e.originalEvent.shiftKey) {
            showFeatureAtPixel(e.pixel, searchableLayers, blocks);
            e.stopPropagation();
        } else {
            blocks.evaluateExpression(lonLatToLiteral(toLonLat(map.getCoordinateFromPixel(e.pixel))));
            e.stopPropagation();
        }
    });
}

function main() {
    d3.json("/bootstrap").then(response => setup(response));
}

export default main;