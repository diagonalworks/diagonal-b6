import * as d3 from 'd3';

import Map from 'ol/Map';
import View from 'ol/View';
import Zoom from 'ol/control/Zoom';
import GeoJSONFormat from 'ol/format/GeoJSON';
import MVT from 'ol/format/MVT';
import { defaults as InteractionDefaults } from 'ol/interaction';
import VectorLayer from 'ol/layer/Vector';
import VectorTileLayer from 'ol/layer/VectorTile';
import { fromLonLat, toLonLat } from 'ol/proj';
import VectorSource from 'ol/source/Vector';
import VectorTileSource from 'ol/source/VectorTile';
import Circle from 'ol/style/Circle';
import Fill from 'ol/style/Fill';
import Icon from 'ol/style/Icon';
import Stroke from 'ol/style/Stroke';
import Style from 'ol/style/Style';
import Text from 'ol/style/Text';

const LineHeight = 20 + 2 * 12;
const LineBorderWidth = 1;
const InsetSquishXS = 8;

const InitialCenter = { latE7: 515361156, lngE7: -1255161 };
const InitalZoom = 16;
const RoadWidths = {
    motorway: 36.0,
    trunk: 36.0,
    primary: 32.0,
    secondary: 26.0,
    tertiary: 26.0,
    street: 18.0,
    unclassified: 18.0,
    service: 18.0,
    residential: 18.0,
    cycleway: 8.0,
    footway: 8.0,
    path: 8.0,
};

function withAlpha(colour, alpha) {
    const rgb = d3.rgb(colour);
    return `rgba(${rgb.r}, ${rgb.g}, ${rgb.b}, ${alpha})`;
}
function scaleWidth(width, resolution) {
    return width * (0.3 / resolution);
}

function roadWidth(feature, resolution) {
    if (RoadWidths[feature.get('highway')]) {
        return scaleWidth(RoadWidths[feature.get('highway')], resolution);
    }
    return 0;
}

function waterwayWidth(resolution) {
    return scaleWidth(32.0, resolution);
}

function newGeoJSONStyle(_, styles) {
    const point = styles.lookupCircle('geojson-point');
    const path = styles.lookupStyle('geojson-path');
    const area = styles.lookupStyle('geojson-area');

    return function (feature) {
        var s = point;
        switch (feature.getGeometry().getType()) {
            case 'LineString':
            case 'MultiLineString':
                s = path;
                break;
            case 'Polygon':
            case 'MultiPolygon':
                s = area;
        }
        if (feature.get('-b6-style')) {
            s = styles.lookupStyle(feature.get('-b6-style'));
        } else if (feature.get('-b6-circle')) {
            s = styles.lookupCircle(feature.get('-b6-circle'));
        } else if (feature.get('-b6-icon')) {
            s = styles.lookupIcon(feature.get('-b6-icon'));
        }
        const label = feature.get('name') || feature.get('label');
        if (label) {
            s = s.clone();
            s.setText(
                new Text({
                    text: label,
                    textAlign: 'left',
                    offsetX: 6,
                    offsetY: 1,
                    font: '"Roboto" 12px',
                }),
            );
        }
        return s;
    };
}

function setupMap(target, state, styles, mapCenter, mapZoom, uiContext) {
    const zoom = new Zoom({
        zoomInLabel: '',
        zoomOutLabel: '',
    });

    var tileURL = '/tiles/base/{z}/{x}/{y}.mvt';
    if (uiContext) {
        const params = new URLSearchParams({ r: idTokenFromProto(uiContext) });
        tileURL += '?' + params.toString();
    }

    const baseSource = new VectorTileSource({
        format: new MVT(),
        url: tileURL,
        minZoom: 10,
        maxZoom: 16,
    });

    var backgroundFill = new Style({
        fill: new Fill({ color: '#eceff7' }),
    });

    const background = new VectorTileLayer({
        source: baseSource,
        style: function (feature) {
            if (feature.get('layer') == 'background') {
                return backgroundFill;
            }
        },
    });

    const boundaries = new VectorTileLayer({
        source: baseSource,
        style: function (feature) {
            if (feature.get('layer') == 'boundary') {
                if (state.featureColours) {
                    const colour =
                        state.featureColours[idKeyFromFeature(feature)];
                    if (colour) {
                        return new Style({
                            fill: new Fill({ color: colour }),
                            stroke: new Stroke({
                                color: '#4f5a7d',
                                width: 0.3,
                            }),
                        });
                    }
                }
                if (feature.get('natural') == 'coastline') {
                    return styles.lookupStyle('coastline');
                }
            }
        },
    });

    const water = new VectorTileLayer({
        source: baseSource,
        style: function (feature, resolution) {
            if (feature.get('layer') == 'water') {
                if (feature.getType() == 'Polygon') {
                    return styles.lookupStyle('water-area');
                } else {
                    const width = waterwayWidth(resolution);
                    if (width > 0) {
                        return styles.lookupStyleWithStokeWidth(
                            'water-line',
                            width,
                        );
                    }
                }
            }
        },
    });
    water.set('clickable', true);

    const landuse = new VectorTileLayer({
        source: baseSource,
        style: function (feature) {
            const landuse = feature.get('landuse');
            const leisure = feature.get('leisure');
            const natural = feature.get('natural');
            if (feature.get('layer') == 'landuse') {
                if (
                    landuse == 'park' ||
                    landuse == 'grass' ||
                    leisure == 'pitch' ||
                    leisure == 'park' ||
                    leisure == 'garden' ||
                    leisure == 'playground' ||
                    leisure == 'nature_reserve'
                ) {
                    return styles.lookupStyle('landuse-greenspace');
                } else if (landuse == 'forest') {
                    return styles.lookupStyle('landuse-forest');
                } else if (landuse == 'meadow' || natural == 'heath') {
                    return styles.lookupStyle('landuse-nature');
                } else if (
                    landuse == 'commercial' ||
                    landuse == 'residential' ||
                    landuse == 'industrial'
                ) {
                    return styles.lookupStyle('landuse-urban');
                }
            } else if (feature.get('layer') == 'contour') {
                return styles.lookupStyle('contour');
            }
        },
    });
    landuse.set('clickable', true);

    const roadOutlines = new VectorTileLayer({
        source: baseSource,
        style: function (feature, resolution) {
            if (feature.get('layer') == 'road' && feature.get('highway')) {
                const width = roadWidth(feature, resolution);
                if (width > 0) {
                    return styles.lookupStyleWithStokeWidth(
                        'road-outline',
                        width + 2.0,
                    );
                }
            }
        },
    });
    roadOutlines.set('clickable', true);
    roadOutlines.set('position', 'MapLayerPositionRoads');

    const roadFills = new VectorTileLayer({
        source: baseSource,
        style: function (feature, resolution) {
            if (feature.get('layer') == 'road' && feature.get('highway') && feature.getGeometry().getType() == 'LineString') {
                const width = roadWidth(feature, resolution);
                if (width > 0) {
                    const id = idKeyFromFeature(feature);
                    const bucket = state.bucketed[id];

                    if (bucket && Object.keys(state.bucketed).length > 0) {
                        if (state.showBucket < 0 || state.showBucket == bucket) {
                            return styles.lookupStyleWithStokeWidth(`bucketed-road-fill-${bucket}`, width);
                        }
                    }
    
                    if (state.highlighted[id]) {
                        return styles.lookupStyleWithStokeWidth(
                            'highlighted-road-fill',
                            width,
                        );
                    } else {
                        return styles.lookupStyleWithStokeWidth(
                            'road-fill',
                            width,
                        );
                    }
                }
            }
        },
    });

    const roadLabels = new VectorTileLayer({
        source: baseSource,
        declutter: true,
        style: function (feature) {
            if (feature.get('layer') == 'road') {
                if (feature.get('name')) {
                    return styles.lookupLineTextWithText(
                        'road-label',
                        feature.get('name'),
                    );
                }
            }
        },
    });

    const rails = new VectorTileLayer({
        source: baseSource,
        style: function (feature) {
            if (feature.get('layer') == 'road' && feature.get('railway')) {
                const id = idKeyFromFeature(feature);
                if (state.highlighted[id]) {
                    return styles.lookupStyle('highlighted-rail');
                } else {
                    return styles.lookupStyle('rail');
                }
            }
        },
    });

    const buildingFill = new Style({
        fill: new Fill({ color: '#ffffff' }),
        stroke: new Stroke({ color: '#4f5a7d', width: 0.3 }),
    });

    const highlightedBuildingFill = styles.lookupStyle('highlighted-area');

    const bucketedBuildingFill = Array.from(Array(6).keys()).map((b) => {
        return styles.lookupStyle(`bucketed-${b}`);
    });

    const buildings = new VectorTileLayer({
        source: baseSource,
        style: function (feature) {
            if (feature.get('layer') === 'building') {
                const id = idKeyFromFeature(feature);
                const bucket = state.bucketed[id];

                if (bucket && Object.keys(state.bucketed).length > 0) {
                    if (state.showBucket < 0) {
                        return bucketedBuildingFill[bucket];
                    }

                    const color = bucketedBuildingFill[bucket]
                        .getFill()
                        .getColor();

                    if (state.showBucket == bucket && state.showBucket > -1) {
                        return new Style({
                            fill: new Fill({ color }),
                            stroke: new Stroke({ color: '#4f5a7d', width: 1 }),
                        });
                    }

                    return new Style({
                        fill: new Fill({ color: withAlpha(color, 0.42) }),
                        stroke: new Stroke({ color: '#4f5a7d', width: 0.3 }),
                    });
                }

                if (state.highlighted[id]) {
                    return highlightedBuildingFill;
                }
                return buildingFill;
            }
        },
    });
    buildings.set('position', 'MapLayerPositionBuildings');
    buildings.set('clickable', true);

    const bucketedPoint = Array.from(Array(6).keys()).map((b) => {
        return styles.lookupCircle(`bucketed-${b}`);
    });

    const points = new VectorTileLayer({
        source: baseSource,
        style: function (feature) {
            if (feature.get('layer') == 'point') {
                const id = idKeyFromFeature(feature);
                const bucket = state.bucketed[id];
                if (
                    bucket &&
                    (state.showBucket < 0 || state.showBucket == bucket)
                ) {
                    return bucketedPoint[bucket];
                }
                if (state.highlighted[id]) {
                    return styles.lookupStyle('highlighted-point');
                }
                return styles.lookupStyle('point');
            }
        },
    });

    const labels = new VectorTileLayer({
        source: baseSource,
        style: function (feature) {
            if (feature.get('layer') == 'label') {
                return new Style({
                    text: new Text({
                        text: feature.get('name'),
                        textAlign: 'left',
                        offsetX: 6,
                        offsetY: 1,
                        fill: new Fill({
                            color: '#000000',
                        }),
                    }),
                    image: new Circle({
                        radius: 2,
                        fill: new Fill({
                            color: '#000000',
                        }),
                    }),
                });
            }
        },
    });

    const view = new View({
        center: fromLonLat([mapCenter.lngE7 / 1e7, mapCenter.latE7 / 1e7]),
        zoom: mapZoom,
    });

    const map = new Map({
        target: target.node(),
        layers: [
            background,
            boundaries,
            water,
            landuse,
            rails,
            roadOutlines,
            roadFills,
            roadLabels,
            buildings,
            points,
            labels,
        ],
        interactions: InteractionDefaults(),
        controls: [zoom],
        view: view,
    });

    const tilesChanged = () => {
        baseSource.refresh();
        baseSource.changed();
    }

    const highlightChanged = () => {
        boundaries.changed();
        buildings.changed();
        roadFills.changed();
        points.changed();
    };

    return [map, tilesChanged, highlightChanged];
}

function lonLatToLiteral(ll) {
    return `${ll[1].toPrecision(8)}, ${ll[0].toPrecision(8)}`;
}

function showFeature(feature, locked, position, ui, logEvent) {
    const ns = feature.get('ns');
    const id = feature.get('id');
    const types = {
        Point: 'point',
        LineString: 'path',
        Polygon: 'area',
        MultiPolygon: 'area',
    };
    if (ns && id && types[feature.getType()]) {
        const expression = `find-feature /${
            types[feature.getType()]
        }/${ns}/${BigInt('0x' + id)}`;
        ui.evaluateExpressionInNewStack(
            expression,
            null,
            locked,
            position,
            logEvent,
        );
    }
}

const StackOffset = [6, 6]; // Relative coordinates of stacks shown next to the mouse cursor

function elementPosition(element) {
    return [
        +element.style('left').replace('px', ''),
        +element.style('top').replace('px', ''),
    ];
}

class Stack {
    constructor(response, target, ui) {
        this.response = response;
        this.target = target;
        this.ui = ui;
        this.chipValues = {};
        for (const i in response.proto.chipValues) {
            this.chipValues[i] = response.proto.chipValues[i];
        }
        this.setup();
        this.onMap = false;
    }

    setup() {
        this.layers = [];
        this.setupGeoJSON(this.response.proto.geoJSON, this.response.geoJSON);
        const queryLayers = this.response.proto.layers;
        for (const i in queryLayers) {
            this.layers.push(
                this.ui.createQueryLayer(
                    queryLayers[i].query,
                    queryLayers[i].before,
                ),
            );
        }
    }

    render() {
        const substacks = this.target
            .selectAll('.substack')
            .data((d) => d.proto.stack.substacks)
            .join((enter) => {
                const div = enter.append('div').attr('class', 'substack');
                div.append('div').attr('class', 'scrollable');
                return div;
            });
        substacks.classed('collapsable', (d) => d.collapsable);
        this.target.selectAll('.collapsable').on('click', function (e) {
            e.preventDefault();
            const substack = d3.select(this);
            substack.classed(
                'collapsable-open',
                !substack.classed('collapsable-open'),
            );
        });

        const scrollables = substacks.select('.scrollable');
        const lines = scrollables
            .selectAll('.line')
            .data((d) => (d.lines ? d.lines : []))
            .join('div');
        lines.attr('class', 'line');
        renderFromProto(lines, 'line', this);
    }

    remove() {
        this.ui.removeStack(this);
    }

    isLocked() {
        return this.locked;
    }

    getLogDetail() {
        return this.response.proto.logDetail;
    }

    getElement() {
        return this.target.node();
    }

    getBlobURL() {
        return this.blobURL;
    }

    getChipValue(index) {
        return this.chipValues[index] || 0;
    }

    setupGeoJSON(indices, data) {
        for (const i in indices) {
            if (!this.stateMatchesCondition(indices[i].condition)) {
                continue;
            }
            const geojson = data[indices[i].index || 0];
            const source = new VectorSource({
                features: [],
            });
            const layer = new VectorLayer({
                source: source,
                style: this.ui.getGeoJSONStyle(),
            });
            const features = new GeoJSONFormat().readFeatures(geojson, {
                dataProjection: 'EPSG:4326',
                featureProjection: this.ui.getProjection(),
            });
            source.addFeatures(features);
            this.layers.push(layer);

            // TODO: We could potentially be creating multiple links
            const blob = new Blob([JSON.stringify(geojson, null, 2)], {
                type: 'application/json',
            });
            this.blobURL = URL.createObjectURL(blob);
        }
    }

    stateMatchesCondition(condition) {
        if (!this.chipValues || !condition) {
            return true;
        }
        for (const i in condition.indices) {
            if (this.chipValues[condition.indices[i]] != condition.values[i]) {
                return false;
            }
        }
        return true;
    }

    redrawHighlights() {
        for (const i in this.layers) {
            this.layers[i].changed();
        }
    }

    _addBucketed() {
        const bucketed = this.response.proto.bucketed;
        for (const i in bucketed) {
            if (!this.stateMatchesCondition(bucketed[i].condition)) {
                continue;
            }
            const buckets = bucketed[i].buckets;
            for (const j in buckets) {
                for (const k in buckets[j].namespaces) {
                    const namespace = buckets[j].namespaces[k];
                    const values = buckets[j].ids[k].ids;
                    for (const l in values) {
                        this.ui.addBucketed(namespace + '/' + values[l], j);
                    }
                }
            }
        }
    }

    _addHighlighted() {
        const highlighted = this.response.proto.highlighted;
        if (highlighted) {
            for (const i in highlighted.namespaces) {
                const namespace = highlighted.namespaces[i];
                const values = highlighted.ids[i].ids;
                for (const j in values) {
                    this.ui.addHighlight(namespace + '/' + values[j]);
                }
            }
        }
    }

    _removeHighlighted() {
        const highlighted = this.response.proto.highlighted;
        if (highlighted) {
            for (const i in highlighted.namespaces) {
                const namespace = highlighted.namespaces[i];
                const values = highlighted.ids[i].ids;
                for (const j in values) {
                    this.ui.removeHighlight(namespace + '/' + values[j]);
                }
            }
        }
    }

    addToMap() {
        this.onMap = true;
        for (const i in this.layers) {
            this.ui.addLayer(this.layers[i]);
        }
        this._addHighlighted();
        this._addBucketed();
        this.ui.basemapHighlightChanged();
    }

    removeFromMap() {
        if (!this.onMap) {
            return;
        }
        this.onMap = false;
        for (const i in this.layers) {
            this.ui.removeLayer(this.layers[i]);
        }
        this._removeHighlighted();
        this.ui.clearBucketed();
        for (const i in this.layers) {
            this.ui.removeLayer(this.layers[i]);
        }
        this.ui.basemapHighlightChanged();
    }

    evaluateNode(node, logEvent) {
        const position = null;
        this.ui.evaluateExpressionInNewStack(
            '',
            node,
            this.response.proto.locked,
            position,
            logEvent,
        );
    }

    evaluateExpressionInContext(expression, logEvent) {
        this.ui.evaluateExpression(
            expression,
            this.response.proto.node,
            this.response.proto.locked,
            logEvent,
            this.target,
        );
    }

    handleDragStart(event, clickAction) {
        this.ui.handleDragStart(event, this.target, clickAction);
    }

    handleChipValueChanged() {
        const chipValues = this.chipValues;
        this.target.selectAll('.atom-chip-clickable').each(function () {
            while (chipValues.length <= this.__chip_index__) {
                chipValues.push(0);
            }
            chipValues[this.__chip_index__] = this.__chip_value__;
        });
        const onMap = this.onMap;
        if (onMap) {
            this.removeFromMap();
        }
        this.render();
        this.setupGeoJSON();
        if (onMap) {
            this.ui.clearBucketed();
            this.addToMap();
        }
    }

    showMessage(message, position) {
        const rect = this.target.node().getBoundingClientRect();
        const x = rect.x + rect.width + InsetSquishXS;
        const y =
            rect.y +
            Math.floor(
                (position[1] - rect.y) / (LineHeight + LineBorderWidth),
            ) *
                LineHeight;
        this.ui.showMessage(message, [x, y]);
    }

    toggleShowBucket(bucket) {
        this.ui.toggleShowBucket(bucket);
    }
}

class ValueAtomRenderer {
    getCSSClass() {
        return 'atom-value';
    }

    enter() {}

    update(atom) {
        atom.text((d) => d.value);
    }
}

class LabelledIconAtomRenderer {
    getCSSClass() {
        return 'atom-labelled-icon';
    }

    enter(atom) {
        atom.append('img');
        atom.append('span');
    }

    update(atom) {
        atom.select('img').attr(
            'src',
            (d) => `/images/${d.labelledIcon.icon}.svg`,
        );
        atom.select('img').attr('class', (d) => `icon-${d.labelledIcon.icon}`);
        atom.select('span').text((d) => d.labelledIcon.label);
    }
}

class DownloadAtomRenderer {
    getCSSClass() {
        return 'atom-download';
    }

    enter(atom) {
        atom.append('a');
    }

    update(atom, stack) {
        const a = atom.select('a');
        a.node().href = stack.getBlobURL();
        a.node().download = 'b6-result.geojson';
        a.text((d) => d.download);
    }
}

class ChipAtomRenderer {
    getCSSClass() {
        return 'atom-chip';
    }

    enter(atom) {
        atom.append('span');
    }

    update(atom, stack) {
        const chip = atom.select('span').classed('atom-chip-clickable', true);
        chip.text((d) => d.chip.labels[stack.getChipValue(d.chip.index || 0)]);
        chip.each(function (d) {
            // Only expects one chip
            this.__chip_index__ = d.chip.index || 0;
            this.__chip_value__ = stack.getChipValue(this.__chip_index__);
        });
        chip.on('click', function (e, d) {
            e.stopPropagation();
            d3.select('.atom-chip-nav').remove();
            chip.classed('open', true);
            const bounds = this.getBoundingClientRect();
            const nav = d3
                .select('body')
                .append('nav')
                .classed('atom-chip-nav', true);
            nav.node().style.top = `${bounds.y + bounds.height}px`;
            nav.node().style.left = `${bounds.x}px`;
            const labels = Array.from(d.chip.labels.entries()).map((l) => {
                return { value: l[0], label: l[1] };
            });
            const chips = nav.selectAll('span').data(labels).join('div');
            chips.classed('atom-chip-clickable', true);
            chips.text((d) => d.label);
            chips.on('click', function (e, d) {
                chip.classed('open', false);
                chip.text(d.label);
                chip.node().__chip_value__ = d.value;
                nav.remove();
                stack.handleChipValueChanged();
            });
        });
    }
}

class ConditionalAtomRenderer {
    getCSSClass() {
        return 'atom-conditional';
    }

    enter() {}

    update(atom, stack) {
        const conditional = atom.datum().conditional;
        for (const i in conditional.conditions) {
            if (stack.stateMatchesCondition(conditional.conditions[i])) {
                const atoms = atom
                    .selectAll('span')
                    .data([conditional.atoms[i]])
                    .join('span');
                renderFromProto(atoms, 'atom', stack);
                return;
            }
        }
    }
}

function atomTextToCopy(atom) {
    if (atom) {
        if (atom.atom.value) {
            return atom.atom.value;
        } else if (atom.atom.labelledIcon) {
            return atom.atom.labelledIcon.label;
        } else if (atom.atom.download) {
            return atom.atom.download;
        }
    }
}

class ValueLineRenderer {
    getCSSClass() {
        return 'line-value';
    }

    enter() {}

    update(line, stack) {
        const atoms = line
            .selectAll('span')
            .data((d) => [d.value.atom])
            .join('span');
        renderFromProto(atoms, 'atom', stack);
        const clickable = line.filter((d) => d.value.clickExpression);
        clickable.classed('clickable', true);
        clickable.on('mousedown', (e, d) => {
            e.stopPropagation();
            const clickHandler = () => {
                stack.evaluateNode(
                    d.value.clickExpression,
                    EventTypeOutlinerClick,
                );
            };
            stack.handleDragStart(e, clickHandler);
        });
        const notClickable = line.filter((d) => !d.value.clickExpression);
        notClickable.on('mousedown', (e, d) => {
            e.stopPropagation();
            const clickHandler = () => {
                const copy = atomTextToCopy(d.value);
                navigator.clipboard.writeText(copy);
                stack.showMessage(
                    `Copied ${copy}`,
                    d3.pointer(e, d3.select('body')),
                );
            };
            stack.handleDragStart(e, clickHandler);
        });
    }
}

class LeftRightValueLineRenderer {
    getCSSClass() {
        return 'line-left-right-value';
    }

    enter() {}

    update(line, stack) {
        const values = [];
        for (const i in line.datum().leftRightValue.left) {
            values.push(line.datum().leftRightValue.left[i]);
        }
        values.push(line.datum().leftRightValue.right);

        let atoms = line
            .selectAll('.line-leftrightvalue-atom')
            .data(values)
            .join('span')
            .attr('class', 'line-leftrightvalue-atom');
        renderFromProto(
            atoms.datum((d) => d.atom),
            'atom',
            stack,
        );

        const clickables = atoms
            .datum((d) => d.clickExpression)
            .filter((d) => d);
        clickables.classed('clickable', true);
        clickables.on('mousedown', (e, d) => {
            e.stopPropagation();
            if (d) {
                const clickHandler = () => {
                    stack.evaluateNode(d, EventTypeOutlinerClick);
                };
                stack.handleDragStart(e, clickHandler);
            }
        });
    }
}

class ExpressionLineRenderer {
    getCSSClass() {
        return 'line-expression';
    }

    enter() {}

    update(line, stack) {
        line.text((d) => d.expression.expression);
        line.on('mousedown', (e, d) => {
            e.stopPropagation();
            const clickHandler = () => {
                navigator.clipboard.writeText(d.expression.expression);
            };
            stack.handleDragStart(e, clickHandler);
        });
    }
}

class TagsLineRenderer {
    getCSSClass() {
        return 'line-tags';
    }

    enter(line) {
        line.append('ul');
    }

    update(line, stack) {
        const formatTags = (t) => [
            { class: 'prefix', text: t.prefix },
            { class: 'key', text: t.key },
            {
                class: 'value',
                text: t.value,
                clickExpression: t.clickExpression,
            },
        ];
        const li = line
            .select('ul')
            .selectAll('li')
            .data((d) => (d.tags.tags ? d.tags.tags.map(formatTags) : []))
            .join('li');
        li.selectAll('span')
            .data((d) => d)
            .join('span')
            .attr('class', (d) => d.class)
            .text((d) => d.text);
        const clickable = li
            .selectAll('.value')
            .filter((d) => d.clickExpression);
        clickable.classed('clickable', true);
        clickable.on('mousedown', (e, d) => {
            e.stopPropagation();
            const clickHandler = () => {
                stack.evaluateNode(d.clickExpression, EventTypeOutlinerClick);
            };
            stack.handleDragStart(e, clickHandler);
        });
    }
}

class HistogramBarLineRenderer {
    getCSSClass() {
        return 'line-histogram-bar';
    }

    enter(line) {
        line.append('div').attr('class', 'range-icon');
        line.append('span').attr('class', 'range');
        line.append('span').attr('class', 'value');
        line.append('span').attr('class', 'total');
        const bar = line.append('div').attr('class', 'value-bar');
        bar.append('div').attr('class', 'fill');
    }

    update(line, stack) {
        line.select('.range-icon').attr(
            'class',
            (d) =>
                `range-icon index-${
                    d.histogramBar.index ? d.histogramBar.index : 0
                }`,
        );
        renderFromProto(
            line.select('.range').datum((d) => d.histogramBar.range),
            'atom',
            stack,
        );
        line.select('.value').text((d) => d.histogramBar.value || '0');
        line.select('.total').text((d) => `/ ${d.histogramBar.total}`);
        line.select('.fill').attr(
            'style',
            (d) =>
                `width: ${
                    ((d.histogramBar.value || 0) / d.histogramBar.total) * 100.0
                }%;`,
        );
    }
}

class SwatchLineRenderer {
    getCSSClass() {
        return 'line-swatch';
    }

    enter(line) {
        const index = line.datum().swatch.index || 0;
        if (index >= 0) {
            line.append('div').attr('class', 'range-icon');
        }
        line.append('span').attr('class', 'label');
    }

    update(line, stack) {
        const index = line.datum().swatch.index || 0;
        if (index >= 0) {
            line.select('.range-icon').attr(
                'class',
                (d) =>
                    `range-icon index-${d.swatch.index ? d.swatch.index : 0}`,
            );

            line.attr('class', `${line.attr('class')} ${'interactive'}`);

            line.on('click', function (e) {
                e.stopPropagation();
                const isSelected = line.classed('selected');

                stack.target
                    .selectAll('.line-swatch')
                    .classed('selected', false);

                if (!isSelected) {
                    line.attr('class', `${line.attr('class')} ${'selected'}`);
                }

                stack.toggleShowBucket(index);
            });
        }
        renderFromProto(
            line.select('.label').datum((d) => d.swatch.label),
            'atom',
            stack,
        );
    }
}

class ShellLineRenderer {
    getCSSClass() {
        return 'line-shell';
    }

    enter(line) {
        const form = line.append('form');
        form.append('div').attr('class', 'prompt').text('b6');
        form.append('input').attr('type', 'text');
        return form;
    }

    update(line, stack) {
        const state = {
            suggestions: line.datum().shell.functions
                ? line.datum().shell.functions.toSorted()
                : [],
            highlighted: 0,
        };
        const form = line.select('form');
        form.select('input').on('focusin', () => {
            form.select('ul').classed('focussed', true);
        });
        form.select('input').on('focusout', () => {
            form.select('ul').classed('focussed', false);
        });
        form.on('keydown', (e) => {
            switch (e.key) {
                case 'Tab': {
                    const node = form.select('input').node();
                    if (
                        state.highlighted >= 0 &&
                        state.filtered[state.highlighted].length >
                            node.value.length
                    ) {
                        node.value = state.filtered[state.highlighted] + ' ';
                    }
                    e.preventDefault();
                    break;
                }
            }
        });
        form.on('keyup', (e) => {
            switch (e.key) {
                case 'ArrowUp':
                    state.highlighted--;
                    e.preventDefault();
                    break;
                case 'ArrowDown':
                    state.highlighted++;
                    e.preventDefault();
                    break;
                default:
                    state.highlighted = 0;
                    break;
            }
            this.updateShellSuggestions(line, state);
        });
        form.on('submit', (e) => {
            e.preventDefault();
            var expression = line.select('input').node().value;
            if (
                state.highlighted >= 0 &&
                state.filtered[state.highlighted].length > expression.length
            ) {
                expression = state.filtered[state.highlighted];
            }
            stack.evaluateExpressionInContext(
                expression,
                EventTypeOutlinerShell,
            );
            return;
        });
    }

    updateShellSuggestions(line, state) {
        const entered = line.select('input').node().value;
        const filtered = state.suggestions.filter((s) => s.startsWith(entered));
        state.filtered = filtered;

        const suggestions = filtered.slice(0, ShellMaxSuggestions).map((s) => {
            return { text: s, highlighted: false };
        });
        if (state.highlighted < 0) {
            state.highlighted = 0;
        } else if (state.highlighted >= filtered.length) {
            state.highlighted = filtered.length - 1;
        }
        if (state.highlighted >= 0) {
            suggestions[state.highlighted].highlighted = true;
        }

        const substack = d3.select(line.node().parentNode);
        const lines = substack
            .selectAll('.line-suggestion')
            .data(suggestions)
            .join('div');
        lines.attr('class', 'line line-suggestion');
        lines.text((d) => d.text).classed('highlighted', (d) => d.highlighted);
    }
}

class ChoiceLineRenderer {
    getCSSClass() {
        return 'line-choice';
    }

    enter() {}

    update(line, stack) {
        const atoms = line
            .selectAll('span')
            .data((d) => this._mergeAtoms(d))
            .join('span');

        renderFromProto(atoms, 'atom', stack);
    }

    _mergeAtoms(d) {
        if (d.choice.chips) {
            return [d.choice.label].concat(d.choice.chips);
        }
        return [d.choice.label];
    }
}

class HeaderLineRenderer {
    getCSSClass() {
        return 'line-header';
    }

    enter(line, stack) {
        line.append('span');
        if (line.datum().header.close) {
            const close = line.append('img');
            close.attr('class', 'line-header-close');
            close.attr('src', '/images/close.svg');
            close.on('click', function (e) {
                e.stopPropagation();
                stack.remove();
            });
        }
        if (line.datum().header.share) {
            const close = line.append('img');
            close.attr('class', 'line-header-share');
            close.attr('src', '/images/share.svg');
            close.on('click', function (e) {
                e.stopPropagation();
                navigator.clipboard.writeText(window.location.href);
                stack.showMessage(
                    'Copied link to clipboard',
                    d3.pointer(e, d3.select('body')),
                );
            });
        }
        line.on('mousedown', (e) => {
            e.stopPropagation();
            stack.handleDragStart(e);
        });
    }

    update(line, stack) {
        renderFromProto(
            line.select('span').datum((d) => d.header.title),
            'atom',
            stack,
        );
    }
}

class ErrorLineRenderer {
    getCSSClass() {
        return 'line-error';
    }

    enter() {}

    update(line) {
        line.text((d) => d.error.error);
    }
}

const Renderers = {
    atom: {
        value: new ValueAtomRenderer(),
        labelledIcon: new LabelledIconAtomRenderer(),
        download: new DownloadAtomRenderer(),
        chip: new ChipAtomRenderer(),
        conditional: new ConditionalAtomRenderer(),
    },
    line: {
        value: new ValueLineRenderer(),
        leftRightValue: new LeftRightValueLineRenderer(),
        expression: new ExpressionLineRenderer(),
        tags: new TagsLineRenderer(),
        histogramBar: new HistogramBarLineRenderer(),
        swatch: new SwatchLineRenderer(),
        shell: new ShellLineRenderer(),
        choice: new ChoiceLineRenderer(),
        header: new HeaderLineRenderer(),
        error: new ErrorLineRenderer(),
    },
};

function renderFromProto(targets, uiElement, stack) {
    const f = function (d) {
        // If the CSS class of the line's div matches the data bound to it, update() it,
        // otherwise remove all child nodes and enter() the line beforehand.
        const renderers = Renderers[uiElement];
        if (!renderers) {
            throw new Error(`Can't render uiElement ${uiElement}`);
        }
        const uiElementType = Object.getOwnPropertyNames(d)[0];
        const renderer = renderers[uiElementType];
        if (!renderer) {
            throw new Error(
                `Can't render ${uiElement} of type ${uiElementType}`,
            );
        }

        const target = d3.select(this);
        if (!target.classed(renderer.getCSSClass)) {
            while (this.firstChild) {
                this.removeChild(this.firstChild);
            }
            target.classed(uiElement, true);
            target.classed(renderer.getCSSClass(), true);
            renderer.enter(target, stack);
        }
        renderer.update(target, stack);
    };
    for (const e of ['click', 'mousedown', 'focusin', 'focusout']) {
        targets.on(e, null);
    }
    targets.each(f);
}

class UI {
    constructor(
        map,
        dockTarget,
        state,
        queryStyle,
        geojsonStyle,
        tilesChanged,
        highlightChanged,
        session,
        logger,
    ) {
        this.map = map;
        (this.dockTarget = dockTarget), (this.state = state);
        this.queryStyle = queryStyle;
        this.geojsonStyle = geojsonStyle;
        this.basemapTilesChanged = tilesChanged;
        this.basemapHighlightChanged = highlightChanged;
        this.session = session;
        this.logger = logger;
        this.uiContext = null;
        this.dragging = null;
        this.shellHistory = [];
        this.html = d3.select('html');
        this.dragPointerOrigin = [0, 0];
        this.dragElementOrigin = [0, 0];
        this.stacks = [];
        this.needHighlightRedraw = false;
        this.docked = [];
        this.openDockIndex = -1;

        this.map.on('moveend', () => {
            this._updateBrowserState();
        });
    }

    getShellHistory() {
        return this.shellHistory;
    }

    handleStartupResponse(response) {
        this.uiContext = response.context;
        if (response.docked) {
            this._renderDock(response.docked);
        }
        if (response.openDockIndex !== undefined) {
            this.toggleDockedAtIndex(response.openDockIndex);
        }
        if (response.expression) {
            const locked = this.uiContext !== undefined;
            const position = null;
            this.evaluateExpressionInNewStack(
                response.expression,
                null,
                locked,
                position,
                EventTypeStartup,
            );
        }
    }

    _renderDock(docked) {
        const target = this.dockTarget
            .selectAll('.stack')
            .data(docked)
            .join('div');
        target.attr('class', 'stack closed');
        const ui = this;
        target.each(function (response, i) {
            this.__dock_index__ = i;
            ui.docked.push(this);
            ui._renderStack(response, d3.select(this), false, false);
        });

        target.on('click', function (e) {
            e.preventDefault();
            ui.toggleDockedAtIndex(this.__dock_index__);
        });
    }

    toggleDockedAtIndex(index) {
        if (index === undefined || index < 0 || index >= this.docked.length) {
            return;
        }
        const docked = this.docked[index];
        const logOptions = { ld: docked.__stack__.getLogDetail() };
        this._logEvent(EventTypeDockOpen, logOptions);
        if (d3.select(docked).classed('closed')) {
            this.closeAllDocked();
            this.removeFeaturedStack();
            d3.select(docked).classed('closed', false);
            this.openDockIndex = index;
            if (docked.__stack__) {
                this.state.bucketed = {};
                docked.__stack__.toggleShowBucket(-1);
                docked.__stack__.addToMap();
            }
        } else {
            this.closeAllDocked();
        }
        this._updateBrowserState();
    }

    closeAllDocked() {
        const docked = this.dockTarget.selectAll('.stack');
        docked.each(function () {
            if (this.__stack__) {
                this.__stack__.removeFromMap();
            }
        });
        docked.classed('closed', true);
        this.openDockIndex = -1;
    }

    _updateBrowserState() {
        const ll = toLonLat(this.map.getView().getCenter());
        const params = new URLSearchParams({
            ll: `${Number(ll[1].toFixed(7))},${Number(ll[0].toFixed(7))}`,
            z: `${Number(this.map.getView().getZoom().toFixed(2))}`,
        });
        if (this.uiContext) {
            params.set('r', idTokenFromProto(this.uiContext));
        }
        if (this.openDockIndex >= 0) {
            params.set('d', this.openDockIndex);
        }
        if (this.lastExpression) {
            params.set('e', this.lastExpression);
        }
        const query = params
            .toString()
            .replaceAll('%2C', ',')
            .replaceAll('%2F', '/');
        history.replaceState(null, '', '/?' + query);
    }

    evaluateExpressionInNewStack(
        expression,
        context,
        locked,
        position,
        logEvent,
    ) {
        const ui = this;
        this._sendRequest(expression, context, locked, logEvent).then(
            (response) => {
                ui._renderNewStack(response, position);
            },
        );
    }

    evaluateExpression(expression, context, locked, logEvent, target) {
        this._sendRequest(expression, context, locked, logEvent).then(
            (response) => {
                this._renderStack(response, target, true, false);
            },
        );
    }

    _sendRequest(expression, context, locked, logEvent) {
        const ll = toLonLat(this.map.getView().getCenter());
        const request = {
            node: context,
            expression: expression,
            locked: locked,
            logEvent: logEvent,
            logMapCenter: {
                lat_e7: Math.round(ll[1] * 1e7),
                lng_e7: Math.round(ll[0] * 1e7),
            },
            logMapZoom: this.map.getView().getZoom(),
            session: this.session,
        };
        if (this.uiContext) {
            request.context = this.uiContext;
        }
        const body = JSON.stringify(request);
        const post = {
            method: 'POST',
            body: body,
            headers: {
                'Content-type': 'application/json; charset=UTF-8',
            },
        };
        const promise = d3.json('/stack', post);
        promise.catch((error) => {
            showMessage(error.message);
            this._logEvent(EventTypeError, { ld: error.message });
        });
        return promise;
    }

    _renderNewStack(response, position) {
        // Creates a new featured stack if one doesn't exist, positioning
        // under the dock, otherwise reuses the existing featured stack.
        // Remove the existing featured stack from the UI if response is
        // null.
        const ui = this;
        const target = d3
            .select('body')
            .selectAll('.stack-featured')
            .data(response ? [response] : [])
            .join(
                (enter) => {
                    return enter.append('div');
                },
                (update) => update,
                (exit) => {
                    exit.each(function () {
                        if (this.__stack__) {
                            ui.removeStack(this.__stack__, true);
                        }
                    });
                    return exit.remove();
                },
            );
        const dockRect = this.dockTarget.node().getBoundingClientRect();
        target.attr('class', 'stack stack-featured');
        if (position) {
            target.style('left', `${StackOffset[0] + position[0]}px`);
            target.style('top', `${StackOffset[1] + position[1]}px`);
        } else {
            target.style('left', `${dockRect.left}px`);
            target.style('top', `${StackOffset[1] + dockRect.bottom}px`);
        }
        target.each(function (response) {
            ui.closeAllDocked();
            ui._renderStack(response, d3.select(this), true, !position);
        });
    }

    _renderStack(response, target, addToMap, recenterMap) {
        target = target.datum(response);
        if (target.node().__stack__) {
            this.removeStack(target.node().__stack__, true);
        }
        const stack = new Stack(response, target, this);
        target.node().__stack__ = stack;
        this.stacks.push(stack);
        stack.render();

        if (addToMap) {
            stack.addToMap();
        }

        if (response.proto.expression) {
            this.shellHistory.push(response.proto.expression);
            this.lastExpression = response.proto.expression;
        } else {
            this.lastExpression = undefined;
        }

        if (response && recenterMap) {
            const center = response.proto.mapCenter;
            if (center && center.latE7 && center.lngE7) {
                this.map.getView().animate({
                    center: fromLonLat([
                        center.lngE7 / 1e7,
                        center.latE7 / 1e7,
                    ]),
                    duration: 500,
                });
            }
        }

        if (response && response.proto.tilesChanged) {
            this.basemapTilesChanged();
        }

        if (this.needHighlightRedraw) {
            this.redrawHighlights();
            this.needHighlightRedraw = false;
        }

        this._updateBrowserState();
    }

    removeFeaturedStack() {
        this._renderNewStack(null);
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
        for (const i in this.stacks) {
            this.stacks[i].redrawHighlights();
        }
        this.basemapHighlightChanged();
    }

    clearBucketed() {
        this.state.bucketed = {};
    }

    addBucketed(idKey, bucket) {
        this.state.bucketed[idKey] = bucket;
    }

    createQueryLayer(query, before) {
        const params = new URLSearchParams({ q: query });
        if (this.uiContext) {
            params.append('r', idTokenFromProto(this.uiContext));
        }
        const source = new VectorTileSource({
            format: new MVT(),
            url: '/tiles/query/{z}/{x}/{y}.mvt?' + params.toString(),
            minZoom: 14,
        });
        const layer = new VectorTileLayer({
            source: source,
            style: this.queryStyle,
        });
        layer.set('clickable', true);
        if (before) {
            layer.set('before', before);
        }
        return layer;
    }

    getGeoJSONStyle() {
        return this.geojsonStyle;
    }

    getProjection() {
        return this.map.getView().getProjection();
    }

    addLayer(layer) {
        if (layer.get('before')) {
            const layers = this.map.getLayers();
            for (let i = 0; i < layers.getLength(); i++) {
                if (layers.item(i).get('position') == layer.get('before')) {
                    layers.insertAt(i, layer);
                    return;
                }
            }
        }
        this.map.addLayer(layer);
    }

    removeLayer(layer) {
        this.map.removeLayer(layer);
    }

    removeStack(stack, keepElement) {
        stack.removeFromMap();
        this.stacks = this.stacks.filter((r) => r != stack);
        if (!keepElement) {
            stack.getElement().remove();
        }
    }

    handleMapClick(event) {
        const position = d3.pointer(event.originalEvent, d3.select('html'));
        if (event.originalEvent.shiftKey) {
            showFeatureAtPixel(
                event.pixel,
                false,
                position,
                this.map,
                this,
                EventTypeMapFeatureClick,
            );
            event.stopPropagation();
        } else {
            const ll = lonLatToLiteral(
                toLonLat(this.map.getCoordinateFromPixel(event.pixel)),
            );
            this.evaluateExpressionInNewStack(
                ll,
                null,
                true,
                position,
                EventTypeMapLatLngClick,
            );
            event.stopPropagation();
        }
    }

    // Start dragging root. If the mouse button is raised without the
    // cursor moving, call clickHandler.
    handleDragStart(event, root, clickHandler) {
        event.preventDefault();
        this.dragging = root;
        this.dragging.classed('dragging', true);
        this.dragClickHandler = clickHandler;
        this.dragPointerOrigin = d3.pointer(event, this.html);
        this.dragElementOrigin = elementPosition(root);
    }

    handlePointerMove(event) {
        if (this.dragging) {
            event.preventDefault();
            const pointer = d3.pointer(event, this.html);
            const delta = [
                pointer[0] - this.dragPointerOrigin[0],
                pointer[1] - this.dragPointerOrigin[1],
            ];
            if (delta[0] != 0 || delta[1] != 0) {
                if (this.dragging.classed('stack-featured')) {
                    this.dragging.attr('class', 'stack stack-floating');
                }
            }
            const left = Math.round(this.dragElementOrigin[0] + delta[0]);
            const top = Math.round(this.dragElementOrigin[1] + delta[1]);
            this.dragging.style('left', `${left}px`);
            this.dragging.style('top', `${top}px`);
        }
    }

    handleDragEnd(event) {
        if (this.dragging) {
            event.preventDefault();
            const pointer = d3.pointer(event, this.html);
            const delta = [
                pointer[0] - this.dragPointerOrigin[0],
                pointer[1] - this.dragPointerOrigin[1],
            ];
            this.dragging.classed('dragging', false);
            if (delta[0] == 0 && delta[1] == 0 && this.dragClickHandler) {
                this.dragClickHandler();
            }
            this.dragging = null;
        }
    }

    showMessage(message, position) {
        showMessage(message, position);
    }

    toggleShowBucket(bucket) {
        if (this.state.showBucket == bucket) {
            this.state.showBucket = -1;
        } else {
            this.state.showBucket = bucket;
        }
        this.basemapHighlightChanged();
    }

    _logEvent(event, options) {
        if (this.uiContext) {
            options.r = idTokenFromProto(this.uiContext);
        }
        const ll = toLonLat(this.map.getView().getCenter());
        options.lc = `${Number(ll[1].toFixed(7))},${Number(ll[0].toFixed(7))}`;
        options.lz = d3.format('.3f')(this.map.getView().getZoom());
        options.ls = this.session;
        this.logger.logEvent(event, options);
    }
}

function showMessage(message, position) {
    d3.select('.message').remove();
    const div = d3.select('body').append('div').classed('message', true);
    if (position) {
        div.node().style.top = `${position[1]}px`;
        div.node().style.left = `${position[0]}px`;
    }
    div.text(message);
    const exit = () => {
        div.classed('exiting', true);
        setTimeout(() => div.remove(), 700);
    };
    setTimeout(exit, 700);
}

const ShellMaxSuggestions = 6;

function showFeatureAtPixel(pixel, locked, position, map, ui, logEvent) {
    const layers = map.getLayers();
    const search = (i, found) => {
        if (i >= 0) {
            const layer = layers.item(i);
            if (layer.getVisible() && layer.get('clickable')) {
                layer.getFeatures(pixel).then((features) => {
                    if (features.length > 0) {
                        found(features[0]);
                        return;
                    } else {
                        search(i - 1, found);
                    }
                });
            } else {
                search(i - 1, found);
            }
        } else {
            const ll = lonLatToLiteral(
                toLonLat(map.getCoordinateFromPixel(pixel)),
            );
            ui.evaluateExpressionInNewStack(
                ll,
                null,
                locked,
                position,
                logEvent,
            );
        }
    };
    search(layers.getLength() - 1, (f) =>
        showFeature(f, locked, position, ui, logEvent),
    );
}

const idGeometryTypes = {
    Point: 'point',
    LineString: 'path',
    Polygon: 'area',
    MultiPolygon: 'area',
};

function idKeyFromFeature(feature) {
    const type = idGeometryTypes[feature.getGeometry().getType()] || 'invalid';
    return `/${type}/${feature.get('ns')}/${parseInt(feature.get('id'), 16)}`;
}

const featureTypes = {
    FeatureTypePoint: 'point',
    FeatureTypePath: 'path',
    FeatureTypeArea: 'area',
    FeatureTypeRelation: 'relation',
    FeatureTypeCollection: 'collection',
    FeatureTypeExpression: 'expression',
};

function idTokenFromProto(p) {
    const type = featureTypes[p.type];
    return `/${type}/${p.namespace}/${p.value || 0}`;
}

function setupShell(target, ui) {
    target
        .selectAll('form')
        .data([1])
        .join(
            (enter) => {
                const form = enter.append('form').attr('class', 'shell-form');
                form.append('div').attr('class', 'prompt').text('b6');
                form.append('input').attr('type', 'text');
                return form;
            },
            (update) => {
                return update;
            },
        );
    const state = { index: 0 };
    d3.select('body').on('keydown', (e) => {
        const history = ui.getShellHistory();
        if (e.key == '`' || e.key == '~') {
            e.preventDefault();
            target.classed('closed', !target.classed('closed'));
            target.select('input').node().focus();
        } else if (e.key == 'ArrowUp') {
            e.preventDefault();
            if (state.index < history.length) {
                state.index++;
                target.select('input').node().value =
                    history[history.length - state.index];
            }
        } else if (e.key == 'ArrowDown') {
            e.preventDefault();
            if (state.index > 0) {
                state.index--;
                if (state.index == 0) {
                    target.select('input').node().value = '';
                } else {
                    target.select('input').node().value =
                        history[history.length - state.index];
                }
            }
        }
    });
    target.select('form').on('submit', (e) => {
        e.preventDefault();
        target.classed('closed', true);
        const expression = target.select('input').node().value;
        state.index = 0;
        const locked = false;
        const position = null;
        ui.evaluateExpressionInNewStack(
            expression,
            null,
            locked,
            position,
            EventTypeWorldShell,
        );
        target.select('input').node().value = '';
    });
}

function newQueryStyle(state, styles) {
    const point = styles.lookupCircle('query-point');
    const highlightedPoint = styles.lookupCircle('highlighted-point');
    const path = styles.lookupStyle('query-path');
    const highlightedPath = styles.lookupStyle('highlighted-path');
    const area = styles.lookupStyle('query-area');
    const highlightedArea = styles.lookupStyle('highlighted-area');
    const boundary = styles.lookupStyle('query-boundary');
    const highlightedBoundary = styles.lookupStyle('highlighted-boundary');

    const bucketedArea = Array.from(Array(6).keys()).map((b) => {
        return styles.lookupStyle(`bucketed-${b}`);
    });

    return function (feature) {
        if (feature.get('layer') != 'background') {
            const id = idKeyFromFeature(feature);
            const highlighted = state.highlighted[id];
            switch (feature.getGeometry().getType()) {
                case 'Point':
                    return highlighted ? highlightedPoint : point;
                case 'LineString':
                    return highlighted ? highlightedPath : path;
                case 'MultiLineString':
                    return highlighted ? highlightedPath : path;
                case 'Polygon':
                    if (state.bucketed[id]) {
                        return bucketedArea[state.bucketed[id]];
                    }
                    if (feature.get('boundary')) {
                        return highlighted ? highlightedBoundary : boundary;
                    } else {
                        return highlighted ? highlightedArea : area;
                    }
                case 'MultiPolygon':
                    if (state.bucketed[id]) {
                        return bucketedArea[state.bucketed[id]];
                    }
                    if (feature.get('boundary')) {
                        return highlighted ? highlightedBoundary : boundary;
                    } else {
                        return highlighted ? highlightedArea : area;
                    }
            }
        }
    };
}

const StyleClasses = [
    'bucketed-0',
    'bucketed-1',
    'bucketed-2',
    'bucketed-3',
    'bucketed-4',
    'bucketed-5',
    'bucketed-road-fill-0',
    'bucketed-road-fill-1',
    'bucketed-road-fill-2',
    'bucketed-road-fill-3',
    'bucketed-road-fill-4',
    'bucketed-road-fill-5',
    'geojson-area',
    'geojson-path',
    'geojson-point',
    'highlighted-area',
    'highlighted-path',
    'highlighted-point',
    'highlighted-rail',
    'highlighted-road-fill',
    'highlighted-boundary',
    'outliner-blue',
    'outliner-emerald',
    'outliner-rose',
    'outliner-teal',
    'outliner-stone',
    'outliner-cyan',
    'outliner-violet',
    'outliner-yellow',
    'query-area',
    'query-path',
    'query-point',
    'query-boundary',
    'rail',
    'road-fill',
    'road-outline',
    'road-label',
    'landuse-urban', // #landuse=commercial, #landuse=residential etc
    'landuse-greenspace', // #landuse=park etc
    'landuse-nature', // #natural=heath, #landuse=meadow etc
    'landuse-forest', // #landuse=forest etc
    'contour',
    'water-area',
    'water-line',
    'coastline',
];

class Styles {
    constructor(classes) {
        const palette = d3
            .select('body')
            .selectAll('.palette')
            .data([1])
            .join('g');
        palette.classed('palette', true);
        const items = palette.selectAll('g').data(classes).join('g');
        items.attr('class', (d) => d);
        this.css = {};
        for (const i in classes) {
            this.css[classes[i]] = getComputedStyle(
                palette.select('.' + classes[i]).node(),
            );
        }
        this.styles = {};
        this.strokes = {};
        this.fills = {};
        this.circles = {};
        this.icons = {};
        this.texts = {};

        this.missingStroke = new Stroke({ color: '#ff0000', width: 1 });
        this.missingFill = new Fill({ color: '#ff0000' });
    }

    lookupStyle(name) {
        if (!this.styles[name]) {
            const options = {};
            const stroke = this.lookupStroke(name);
            if (stroke) {
                options['stroke'] = stroke;
            }
            const fill = this.lookupFill(name);
            if (fill) {
                options['fill'] = fill;
            }
            this.styles[name] = new Style(options);
        }
        return this.styles[name];
    }

    lookupStyleWithStokeWidth(name, width) {
        const key = `${name}-width${width}`;
        if (!this.styles[key]) {
            const s = this.lookupStyle(name).clone();
            s.getStroke().setWidth(width);
            this.styles[key] = s;
        }
        return this.styles[key];
    }

    lookupStroke(name) {
        if (!this.strokes[name]) {
            if (this.css[name]) {
                if (this.css[name]['stroke'] != 'none') {
                    this.strokes[name] = new Stroke({
                        color: this.css[name]['stroke'],
                        width: +this.css[name]['stroke-width'].replace(
                            'px',
                            '',
                        ),
                    });
                } else {
                    this.strokes[name] = null;
                }
            } else {
                this.strokes[name] = this.missingStroke;
            }
        }
        return this.strokes[name];
    }

    lookupFill(name) {
        if (!this.fills[name]) {
            if (this.css[name]) {
                if (this.css[name]['fill-opacity'] != 0) {
                    this.fills[name] = new Fill({
                        color: this.css[name]['fill'],
                    });
                } else {
                    this.fills[name] = null;
                }
            } else {
                this.fills[name] = this.missingFill;
            }
        }
        return this.fills[name];
    }

    lookupCircle(name) {
        if (!this.circles[name]) {
            this.circles[name] = new Style({
                image: new Circle({
                    radius: 4,
                    stroke: this.lookupStroke(name),
                    fill: this.lookupFill(name),
                }),
            });
        }
        return this.circles[name];
    }

    lookupIcon(name) {
        if (!this.icons[name]) {
            this.icons[name] = new Style({
                image: new Icon({
                    src: `/images/${name}.svg`,
                }),
            });
        }
        return this.icons[name];
    }

    lookupLineTextWithText(name, text) {
        if (!this.texts[name]) {
            const options = {
                font: '11px sans-serif',
                placement: 'line',
            };
            if (this.css[name]) {
                if (this.css[name].font) {
                    options.font = this.css[name].font;
                }
                if (this.css[name].fill) {
                    options.fill = this.lookupFill(name);
                }
                if (this.css[name].stroke) {
                    options.stroke = this.lookupStroke(name);
                }
            }
            this.texts[name] = new Style({
                text: new Text(options),
            });
        }
        const style = this.texts[name].clone();
        style.getText().setText(text);
        return style;
    }
}

const EventTypeStartup = 's';
const EventTypeDockOpen = 'do';
const EventTypeMapLatLngClick = 'mlc';
const EventTypeMapFeatureClick = 'mfc';
const EventTypeOutlinerClick = 'oc';
const EventTypeOutlinerShell = 'os';
const EventTypeWorldShell = 'ws';
const EventTypeError = 'err';

class NullEventLogger {
    logEvent() {}
}

function setup(selector, startupResponse, logger) {
    const target = d3.select(selector).classed('b6', true);
    const mapTarget = target.append('div').classed('map', true);
    const shellTarget = target
        .append('div')
        .classed('shell', true)
        .classed('closed', true);
    const dockTarget = target.append('div').classed('dock', true);
    if (startupResponse.error) {
        showMessage(startupResponse.error);
        return;
    }
    const state = { highlighted: {}, bucketed: {}, showBucket: -1 };
    const styles = new Styles(StyleClasses);
    const mapCenter = startupResponse.mapCenter || InitialCenter;
    const mapZoom = startupResponse.mapZoom || InitalZoom;
    const [map, tilesChanged, highlightChanged] = setupMap(
        mapTarget,
        state,
        styles,
        mapCenter,
        mapZoom,
        startupResponse.context,
    );
    const queryStyle = newQueryStyle(state, styles);
    const geojsonStyle = newGeoJSONStyle(state, styles);
    const ui = new UI(
        map,
        dockTarget,
        state,
        queryStyle,
        geojsonStyle,
        tilesChanged,
        highlightChanged,
        startupResponse.session,
        logger,
    );
    const html = d3.select('html');
    html.on('pointermove', (e) => {
        ui.handlePointerMove(e);
    });
    html.on('mouseup', (e) => {
        ui.handleDragEnd(e);
    });

    setupShell(shellTarget, ui);
    ui.handleStartupResponse(startupResponse);

    map.on('singleclick', (e) => {
        ui.handleMapClick(e);
    });
}

function main(selector, logger) {
    if (!logger) {
        logger = new NullEventLogger();
    }
    const params = new URLSearchParams(window.location.search);
    d3.json('/startup?' + params.toString()).then((response) =>
        setup(selector, response, logger),
    );
}

export default main;
