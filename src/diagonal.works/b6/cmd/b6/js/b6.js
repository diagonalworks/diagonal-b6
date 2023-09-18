import * as d3 from "d3";

import {defaults as InteractionDefaults} from "ol/interaction";
import {fromLonLat, toLonLat} from "ol/proj";
import Circle from "ol/style/Circle";
import Fill from "ol/style/Fill";
import GeoJSONFormat from "ol/format/GeoJSON";
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
        if (feature.get("-b6-stroke")) {
            if (!cloned) {
                s = s.clone();
                cloned = true;
            }
            if (s.getStroke()) {
                s.getStroke().setColor(parseColour(feature.get("-b6-stroke"), "stroke", styles));
            }
        }
        if (feature.get("-b6-fill")) {
            if (!cloned) {
                s = s.clone();
                cloned = true;
            }
            if (s.getFill()) {
                s.getFill().setColor(parseColour(feature.get("-b6-fill"), "fill", styles));
            }
        }
        if (feature.get("-b6-circle")) {
            if (!cloned) {
                s = s.clone();
                cloned = true;
            }
            s.setImage(new Circle({
                radius: 4,
                fill: new Fill({
                    color: parseColour(feature.get("-b6-circle"), "fill", styles),
                }),
                stroke: new Stroke({
                    color: parseColour(feature.get("-b6-circle"), "stroke", styles),
                    width: 1,
                }),
            }));
        }
        return s;
    }
}

function parseColour(colour, attribute, styles) {
    if (colour) {
        if (colour.startsWith("#")) {
            return colour;
        } else if (styles[colour]) {
            return styles[colour][attribute];
        }
    }
    return "#ff0000";
}

function setupMap(state, styles, mapCenter) {
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

    const bucketedBuildingFill = Array.from(Array(5).keys()).map(b => {
        return new Style({
            fill: new Fill({color: styles[`bucketed-${b}`]["fill"]}),
            stroke: new Stroke({color: "#4f5a7d", width: 0.3})
        });
    });

    const buildings = new VectorTileLayer({
        source: baseSource,
        style: function(feature, resolution) {
            if (feature.get("layer") == "building") {
                const id = idKeyFromFeature(feature);
                if (state.bucketed[id]) {
                    return bucketedBuildingFill[state.bucketed[id]];
                }
                if (state.highlighted[id]) {
                    console.log("highlighted");
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
        center: fromLonLat(mapCenter ? [mapCenter.lngE7 / 1e7, mapCenter.latE7 / 1e7] : InitialCenter),
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

function showFeature(feature, locked, position, ui) {
    const ns = feature.get("ns");
    const id = feature.get("id");
    const types = {"Point": "point", "LineString": "path", "Polygon": "area", "MultiPolygon": "area"};
    if (ns && id && types[feature.getType()]) {
        ui.evaluateExpression(`find-feature /${types[feature.getType()]}/${ns}/${BigInt("0x" + id)}`, locked, position);
    }
}

const StackOffset = [6, 6]; // Relative coordinates of stacks shown next to the mouse cursor

function elementPosition(element) {
    return [+element.style("left").replace("px", ""), +element.style("top").replace("px", "")];
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
    constructor(response, target, ui) {
        this.expressionContext = response.proto.node;
        this.locked = response.proto.locked;
        this.target = target;
        this.layers = []
        if (response.proto.highlighted) {
            this.highlighted = response.proto.highlighted;
        }
        if (response.proto.bucketed) {
            this.bucketed = response.proto.bucketed;
        }
        if (response.geojson) {
            this.initGeoJSON(response.geojson, ui);
        }
        if (response.proto.queryLayers) {
            for (const i in response.proto.queryLayers) {
                this.layers.push(ui.createQueryLayer(response.proto.queryLayers[i]));
            }
        }
    }

    getTarget() {
        return this.target;
    }

    getExpressionContext() {
        return this.expressionContext;
    }

    isLocked() {
        return this.locked;
    }

    getBlobURL() {
        return this.blobURL;
    }

    initGeoJSON(geojson, ui) {
        const source = new VectorSource({
            features: [],
         })
         const layer = new VectorLayer({
             source: source,
             style: ui.getGeoJSONStyle(),
         })
        const features = new GeoJSONFormat().readFeatures(geojson, {
            dataProjection: "EPSG:4326",
            featureProjection: ui.getProjection(),
        });
        source.addFeatures(features);
        this.layers.push(layer);

        const blob = new Blob([JSON.stringify(geojson, null, 2)], {
            type: "application/json",
        });
        this.blobURL = URL.createObjectURL(blob);
    }

    addBucketed(ui) {
        if (this.bucketed) {
            for (const i in this.bucketed) {
                for (const j in this.bucketed[i].namespaces) {
                    const namespace = this.bucketed[i].namespaces[j];
                    const values = this.bucketed[i].ids[j].ids;
                    for (const k in values) {
                        ui.addBucketed(namespace + "/" + values[k], i);
                    }
                }
            }
        }
    }

    redrawHighlights() {
        for (const i in this.layers) {
            this.layers[i].changed();
        }
    }

    addTo(ui) {
        for (const i in this.layers) {
            ui.addLayer(this.layers[i]);
        }

        if (this.highlighted) {
            for (const i in this.highlighted.namespaces) {
                const namespace = this.highlighted.namespaces[i];
                const values = this.highlighted.ids[i].ids;
                for (const j in values) {
                    ui.addHighlight(namespace + "/" + values[j]);
                }
            }
        }
    }

    removeFrom(ui) {
        for (const i in this.layers) {
            ui.removeLayer(this.layers[i]);
        }
        if (this.highlighted) {
            for (const i in this.highlighted.namespaces) {
                const namespace = this.highlighted.namespaces[i];
                const values = this.highlighted.ids[i].ids;
                for (const j in values) {
                    ui.removeHighlight(namespace + "/" + values[j])
                }
            }
        }
        for (const i in this.layers) {
            ui.removeLayer(this.layers[i]);
        }
    }
}

class ValueAtomRenderer {
    getCSSClass() {
        return "atom-value";
    }

    enter() {}

    update(atom) {        
        atom.text(d => d.value);
    }    
}

class LabelledIconAtomRenderer {
    getCSSClass() {
        return "atom-labelled-icon";
    }

    enter(atom) {
        atom.append("img");
        atom.append("span");
    }

    update(atom) {
        atom.select("img").attr("src", d => `/images/${d.labelledIcon.icon}.svg`);
        atom.select("img").attr("class", d => `icon-${d.labelledIcon.icon}`);
        atom.select("span").text(d => d.labelledIcon.label);
    }    
}

class DownloadAtomRenderer {
    getCSSClass() {
        return "atom-download";
    }

    enter(atom) {
        atom.append("a");
    }

    update(atom, renderedResponse, ui) {
        const a = atom.select("a");
        a.node().href = renderedResponse.getBlobURL();
        a.node().download = "b6-result.geojson";
        a.text(d => d.download);
    }    
}

class ValueLineRenderer {
    getCSSClass() {
        return "line-value";
    }

    enter(line) {}

    update(line, renderedResponse, ui) {
        const atoms = line.selectAll("span").data(d => [d.value.atom]).join("span");
        renderFromProto(atoms, "atom", renderedResponse, ui)
        const clickable = line.filter(d => d.value.clickExpression);
        clickable.classed("clickable", true);
        clickable.on("click", (e, d) => {
            e.preventDefault();
            ui.evaluateNode(d.value.clickExpression, renderedResponse.isLocked());
        })
    }    
}

class LeftRightValueLineRenderer {
    getCSSClass() {
        return "line-left-right-value";
    }

    enter(line) {}

    update(line, renderedResponse, ui) {
        const values = [];
        for (const i in line.datum().leftRightValue.left) {
            values.push(line.datum().leftRightValue.left[i]);
        }
        values.push(line.datum().leftRightValue.right);

        let atoms = line.selectAll("span").data(values).join("span");
        renderFromProto(atoms.datum(d => d.atom), "atom", renderedResponse, ui)

        atoms = line.selectAll("span").data(values);
        const clickables = atoms.datum(d => d.clickExpression).filter(d => d);
        clickables.classed("clickable", true);
        clickables.on("click", (e, d) => {
            if (d) {
                e.preventDefault();
                ui.evaluateNode(d);
            }
        })
    }
}

class ExpressionLineRenderer {
    getCSSClass() {
        return "line-expression";
    }

    enter(line) {}

    update(line, renderedResponse, ui) {
        line.text(d => d.expression.expression);
        line.on("mousedown", e => {
            ui.handleDragStart(e, renderedResponse.getTarget());
        })
    }    
}

class TagsLineRenderer {
    getCSSClass() {
        return "line-tags";
    }

    enter(line) {
        line.append("ul");
    }

    update(line, renderedResponse, ui) {
        const formatTags = t => [
            {class: "prefix", text: t.prefix},
            {class: "key", text: t.key},
            {class: "value", text: t.value, clickExpression: t.clickExpression},
        ];
        const li = line.select("ul").selectAll("li").data(d => d.tags.tags ? d.tags.tags.map(formatTags) : []).join("li");
        li.selectAll("span").data(d => d).join("span").attr("class", d => d.class).text(d => d.text);
        const clickable = li.selectAll(".value").filter(d => d.clickExpression);
        clickable.classed("clickable", true);
        clickable.on("click", (e, d) => {
            e.preventDefault();
            ui.evaluateNode(d.clickExpression);
        });
    }
}

class HistogramBarLineRenderer {
    getCSSClass() {
        return "line-histogram-bar";
    }

    enter(line) {
        line.append("div").attr("class", "range-icon");
        line.append("span").attr("class", "range");
        line.append("span").attr("class", "value");
        line.append("span").attr("class", "total");
        const bar = line.append("div").attr("class", "value-bar");
        bar.append("div").attr("class", "fill");
    }

    update(line, renderedResponse, ui) {
        line.select(".range-icon").attr("class", d => `range-icon index-${d.histogramBar.index ? d.histogramBar.index : 0}`);
        renderFromProto(line.select(".range").datum(d => d.histogramBar.range), "atom", renderedResponse, ui);
        line.select(".value").text(d => d.histogramBar.value);
        line.select(".total").text(d => `/ ${d.histogramBar.total}`);
        line.select(".fill").attr("style", d => `width: ${d.histogramBar.value/d.histogramBar.total*100.00}%;`);
    }
}

class ShellLineRenderer {
    getCSSClass() {
        return "line-shell";
    }

    enter(line) {
        const form = line.append("form");
        form.append("div").attr("class", "prompt").text("b6");
        form.append("input").attr("type", "text");
        return form
    }

    update(line, renderedResponse, ui) {
        const state = {suggestions: line.datum().shell.functions ? line.datum().shell.functions.toSorted() : [], highlighted: 0};
        const form = line.select("form");
        form.select("input").on("focusin", e => {
            form.select("ul").classed("focussed", true);
        });
        form.select("input").on("focusout", e => {
            form.select("ul").classed("focussed", false);
        });
        form.on("keydown", e => {
            switch (e.key) {
                case "Tab":
                    const node = form.select("input").node();
                    if (state.highlighted >= 0 && state.filtered[state.highlighted].length > node.value.length) {
                        node.value = state.filtered[state.highlighted] + " ";
                    }
                    e.preventDefault();
                    break;
            }
        });
        form.on("keyup", e => {
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
            this.updateShellSuggestions(line, state);
        });
        form.on("submit", e => {
            e.preventDefault();
            acceptShellSuggestion(line, state, renderedResponse, ui);
            return;
        });    
    }

    updateShellSuggestions(line, state) {
        const entered = line.select("input").node().value;
        const filtered = state.suggestions.filter(s => s.startsWith(entered));
        state.filtered = filtered;

        const suggestions = filtered.slice(0, ShellMaxSuggestions).map(s => {return {text: s, highlighted: false}});
        if (state.highlighted < 0) {
            state.highlighted = 0
        } else if (state.highlighted >= filtered.length) {
            state.highlighted = filtered.length - 1;
        }
        if (state.highlighted >= 0) {
            suggestions[state.highlighted].highlighted = true;
        }
    
        const substack = d3.select(line.node().parentNode);
        const lines = substack.selectAll(".line-suggestion").data(suggestions).join("div");
        lines.attr("class", "line line-suggestion");
        lines.text(d => d.text).classed("highlighted", d => d.highlighted);
    }
}

class QuestionLineRenderer {
    getCSSClass() {
        return "line-question";
    }

    enter(line) {}

    update(line) {
        line.text(d => d.question.question);
    }
}

class ErrorLineRenderer {
    getCSSClass() {
        return "line-error";
    }

    enter(line) {}

    update(line) {
        line.text(d => d.error.error);
    }    
}

const Renderers = {
    "atom": {
        "value": new ValueAtomRenderer(),
        "labelledIcon": new LabelledIconAtomRenderer(),
        "download": new DownloadAtomRenderer(),
    },
    "line": {
        "value": new ValueLineRenderer(),
        "leftRightValue": new LeftRightValueLineRenderer(),
        "expression": new ExpressionLineRenderer(),
        "tags": new TagsLineRenderer(),
        "histogramBar": new HistogramBarLineRenderer(),
        "shell": new ShellLineRenderer(),
        "question": new QuestionLineRenderer(),
        "error": new ErrorLineRenderer(),
    }
}

function renderFromProto(targets, uiElement, renderedResponse, ui) {
    const f = function(d) {
        // If the CSS class of the line's div matches the data bound to it, update() it,
        // otherwise remove all child nodes and enter() the line beforehand.
        const renderers = Renderers[uiElement];
        if (!renderers) {
            throw new Error(`Can't render uiElement ${uiElement}`);
        }
        const uiElementType = Object.getOwnPropertyNames(d)[0];
        const renderer = renderers[uiElementType];
        if (!renderer) {
            throw new Error(`Can't render ${uiElement} of type ${uiElementType}`);
        }

        const target = d3.select(this);
        if (!target.classed(renderer.getCSSClass)) {
            while (this.firstChild) {
                this.removeChild(this.firstChild);
            }
            target.classed(uiElement, true);
            target.classed(renderer.getCSSClass(), true);
            renderer.enter(target);
       }
       renderer.update(target, renderedResponse, ui);
    }
    targets.each(f);
}

class UI {
    constructor(map, state, queryStyle, geojsonStyle, highlightChanged, context) {
        this.map = map;
        this.state = state;
        this.queryStyle = queryStyle;
        this.geojsonStyle = geojsonStyle;
        this.basemapHighlightChanged = highlightChanged;
        this.context = context;
        this.dragging = null;
        this.html = d3.select("html");
        this.dragPointerOrigin = [0,0];
        this.dragElementOrigin = [0,0];
        this.rendered = [];
        this.needHighlightRedraw = false;
    }

    evaluateExpression(expression, locked, position) {
        this.evaluateExpressionInContext(null, expression, locked, position);
    }

    evaluateNode(node, locked) {
        this.evaluateExpressionInContext(node, null, locked);
    }

    evaluateExpressionInContext(node, expression, locked, position) {
        const request = {
            node: node,
            expression: expression,
            locked: locked,
        }
        if (this.context) {
            request.context = this.context;
        }
        const body = JSON.stringify(request);
        const post = {
            method: "POST",
            body: body,
            headers: {
                "Content-type": "application/json; charset=UTF-8"
            }
        }
        d3.json("/ui", post).then(response => {
            this.renderFeaturedUIResponse(response, position);
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

    addBucketed(idKey, bucket) {
        this.state.bucketed[idKey] = bucket;
    }

    createQueryLayer(query) {
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
        return layer;
    }

    getGeoJSONStyle() {
        return this.geojsonStyle
    }

    getProjection() {
        return this.map.getView().getProjection();
    }

    addLayer(layer) {
        this.map.addLayer(layer);
    }

    removeLayer(layer) {
        this.map.removeLayer(layer);
    }

    renderDock(docked) {
        const target = d3.select("#dock").selectAll(".stack").data(docked).join("div");
        target.attr("class", "stack closed");
        this.renderUIResponse(target);
        const ui = this;
        target.on("click", function(e) {
            e.preventDefault();
            target.classed("closed", true);
            target.each(function() {
                if (this.__rendered__) {
                    this.__rendered__.removeFrom(ui);
                }
            });
            ui.removeFeaturedUIResponse();
            d3.select(this).classed("closed", false);
            if (this.__rendered__) {
                ui.state.bucketed = {};
                this.__rendered__.addTo(ui);
                this.__rendered__.addBucketed(ui);
                ui.basemapHighlightChanged();
            }
        });
    }

    removeFeaturedUIResponse() {
        this.renderFeaturedUIResponse(null);
    }

    renderFeaturedUIResponse(response, position) {
        d3.select("#dock").selectAll(".stack").classed("closed", true);
        if (Object.keys(this.state.bucketed).length > 0) {
            this.state.bucketed = {};
            this.basemapHighlightChanged();
        }
        const ui = this;
        const root = d3.select("body").selectAll(".stack-featured").data(response ? [response] : []).join(
            enter => {
                return enter.append("div");
            },
            update => update,
            exit => {
                exit.each(function() {
                    ui.removeRenderedResponse(this);
                });
                return exit.remove();
            },
        );

        const dockRect = d3.select("#dock").node().getBoundingClientRect();
        root.attr("class", "stack stack-featured");
        if (position) {
            root.style("left",  `${StackOffset[0] + position[0]}px`);
            root.style("top", `${StackOffset[1] + position[1]}px`);
        } else {
            root.style("left",  `${dockRect.left}px`);
            root.style("top", `${StackOffset[1] + dockRect.bottom}px`);
        }
        this.renderUIResponse(root, true);
        if (response && !position) {
            const center = response.proto.mapCenter;
            if (center && center.latE7 && center.lngE7) {
                this.map.getView().animate({
                    center: fromLonLat([center.lngE7 / 1e7, center.latE7 / 1e7]),
                    duration: 500,
                });
            }
        }
    }

    renderUIResponse(target, featured) {
        const substacks = target.selectAll(".substack").data(d => d.proto.stack.substacks).join(
            enter => {
                const div = enter.append("div").attr("class", "substack");
                div.append("div").attr("class", "scrollable");
                return div;
            }
        );
        substacks.classed("collapsable", d => d.collapsable);
        target.selectAll(".collapsable").on("click", function(e) {
            e.preventDefault();
            const substack = d3.select(this);
            substack.classed("collapsable-open", !substack.classed("collapsable-open"));
        });
        const scrollables = substacks.select(".scrollable");
        const lines = scrollables.selectAll(".line").data(d => d.lines).join("div");
        lines.attr("class", "line");
        const ui = this;
        const f = function(response) {
            ui.removeRenderedResponse(this);
            this.__rendered__ = new RenderedResponse(response, d3.select(this), ui);
            ui.rendered.push(this.__rendered__);
            renderFromProto(lines, "line", this.__rendered__, ui);
            if (featured) {
                this.__rendered__.addTo(ui);
            }
        }
        target.each(f);
        if (this.needHighlightRedraw) {
            this.redrawHighlights();
            this.needHighlightRedraw = false;            
        }
    }

    removeRenderedResponse(node) {
        if (node.__rendered__) {
            node.__rendered__.removeFrom(this);
            this.rendered = this.rendered.filter(r => r != node.__rendered__);
        }
    }

    handleDragStart(event, root) {
        event.preventDefault();
        if (root.classed("stack-featured")) {
            root.attr("class", "stack stack-floating");
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

const ShellMaxSuggestions = 6;

function acceptShellSuggestion(block, state, renderedResponse, ui) {
    var expression = block.select("input").node().value;
    if (state.highlighted >= 0 && state.filtered[state.highlighted].length > expression.length) {
        expression = state.filtered[state.highlighted];
    }
    ui.evaluateExpressionInContext(renderedResponse.getExpressionContext(), expression, false);
}

function showFeatureAtPixel(pixel, layers, locked, position, ui) {
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
    search(0, f => showFeature(f, locked, position, ui));
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
    return `/${type}/${feature.get("ns")}/${parseInt(feature.get("id"), 16)}`
}

function setupShell(target, ui) {
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
        ui.evaluateExpression(expression, false);
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
    "bucketed-0",
    "bucketed-1",
    "bucketed-2",
    "bucketed-3",
    "bucketed-4",
    "geojson-area",
    "geojson-path",
    "geojson-point",
    "highlighted-area",
    "highlighted-path",
    "highlighted-point",
    "highlighted-road-fill",
    "outliner-blue",
    "outliner-emerald",
    "outliner-rose",
    "outliner-teal",
    "query-area",
    "query-path",
    "query-point",
    "road-fill",
];

function setup(startupResponse) {
    const state = {highlighted: {}, bucketed: {}};
    const styles = lookupStyles(Styles);
    const [map, searchableLayers, highlightChanged] = setupMap(state, styles, startupResponse.mapCenter);
    const queryStyle = newQueryStyle(state, styles);
    const geojsonStyle = newGeoJSONStyle(state, styles);
    const ui = new UI(map, state, queryStyle, geojsonStyle, highlightChanged, startupResponse.context);
    const html = d3.select("html");
    html.on("pointermove", e => {
        ui.handlePointerMove(e);
    });
    html.on("mouseup", e => {
        ui.handleDragEnd(e);
    });

    setupShell(d3.select("#shell"), ui);

    if (startupResponse.docked) {
        ui.renderDock(startupResponse.docked);
    }

    map.on("singleclick", e => {
        const position = d3.pointer(e.originalEvent, html);
        if (e.originalEvent.shiftKey) {
            showFeatureAtPixel(e.pixel, searchableLayers, false, position, ui);
            e.stopPropagation();
        } else {
            ui.evaluateExpression(lonLatToLiteral(toLonLat(map.getCoordinateFromPixel(e.pixel))), true, position);
            e.stopPropagation();
        }
    });
}

function main() {
    const params = new URLSearchParams(window.location.search);
    d3.json("/startup?" + params.toString()).then(response => setup(response));
}

export default main;