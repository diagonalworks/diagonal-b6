/* eslint-disable */
import Long from "long";
import _m0 from "protobufjs/minimal";
import { FeatureIDProto, NodeProto } from "./api";
import { PointProto } from "./geometry";

export const protobufPackage = "ui";

export enum MapLayerPosition {
  MapLayerPositionEnd = 0,
  MapLayerPositionRoads = 1,
  MapLayerPositionBuildings = 2,
  UNRECOGNIZED = -1,
}

export function mapLayerPositionFromJSON(object: any): MapLayerPosition {
  switch (object) {
    case 0:
    case "MapLayerPositionEnd":
      return MapLayerPosition.MapLayerPositionEnd;
    case 1:
    case "MapLayerPositionRoads":
      return MapLayerPosition.MapLayerPositionRoads;
    case 2:
    case "MapLayerPositionBuildings":
      return MapLayerPosition.MapLayerPositionBuildings;
    case -1:
    case "UNRECOGNIZED":
    default:
      return MapLayerPosition.UNRECOGNIZED;
  }
}

export function mapLayerPositionToJSON(object: MapLayerPosition): string {
  switch (object) {
    case MapLayerPosition.MapLayerPositionEnd:
      return "MapLayerPositionEnd";
    case MapLayerPosition.MapLayerPositionRoads:
      return "MapLayerPositionRoads";
    case MapLayerPosition.MapLayerPositionBuildings:
      return "MapLayerPositionBuildings";
    case MapLayerPosition.UNRECOGNIZED:
    default:
      return "UNRECOGNIZED";
  }
}

export interface UIRequestProto {
  node?: NodeProto | undefined;
  expression?: string | undefined;
  root?: FeatureIDProto | undefined;
  locked?: boolean | undefined;
  logEvent?: string | undefined;
  logMapCenter?: PointProto | undefined;
  logMapZoom?: number | undefined;
  session?: number | undefined;
}

export interface UIResponseProto {
  stack?: StackProto | undefined;
  node?: NodeProto | undefined;
  expression?: string | undefined;
  highlighted?:
    | FeatureIDsProto
    | undefined;
  /** References geojson array in response */
  geoJSON?: GeoJSONProto[] | undefined;
  layers?: MapLayerProto[] | undefined;
  mapCenter?: PointProto | undefined;
  locked?: boolean | undefined;
  chipValues?: number[] | undefined;
  logDetail?: string | undefined;
  tilesChanged?: boolean | undefined;
}

export interface MapLayerProto {
  path?: string | undefined;
  q?: string | undefined;
  v?: string | undefined;
  before?: MapLayerPosition | undefined;
  condition?: ConditionProto | undefined;
}

export interface StackProto {
  substacks?: SubstackProto[] | undefined;
  id?: FeatureIDProto | undefined;
}

export interface SubstackProto {
  lines?: LineProto[] | undefined;
  collapsable?: boolean | undefined;
}

export interface LineProto {
  value?: ValueLineProto | undefined;
  leftRightValue?: LeftRightValueLineProto | undefined;
  expression?: ExpressionLineProto | undefined;
  tags?: TagsLineProto | undefined;
  histogramBar?: HistogramBarLineProto | undefined;
  swatch?: SwatchLineProto | undefined;
  shell?: ShellLineProto | undefined;
  choice?: ChoiceLineProto | undefined;
  header?: HeaderLineProto | undefined;
  error?: ErrorLineProto | undefined;
  action?: ActionLineProto | undefined;
  comparison?: ComparisonLineProto | undefined;
}

export interface ValueLineProto {
  atom?: AtomProto | undefined;
  clickExpression?: NodeProto | undefined;
}

export interface LeftRightValueLineProto {
  left?: ClickableAtomProto[] | undefined;
  right?: ClickableAtomProto | undefined;
}

export interface ClickableAtomProto {
  atom?: AtomProto | undefined;
  clickExpression?: NodeProto | undefined;
}

export interface ExpressionLineProto {
  expression?: string | undefined;
}

export interface TagsLineProto {
  tags?: TagAtomProto[] | undefined;
}

export interface TagAtomProto {
  prefix?: string | undefined;
  key?: string | undefined;
  value?: string | undefined;
  clickExpression?: NodeProto | undefined;
}

export interface HistogramBarLineProto {
  range?: AtomProto | undefined;
  value?: number | undefined;
  total?: number | undefined;
  index?: number | undefined;
}

export interface SwatchLineProto {
  label?: AtomProto | undefined;
  index?: number | undefined;
}

export interface ShellLineProto {
  functions?: string[] | undefined;
}

export interface ChoiceLineProto {
  label?: AtomProto | undefined;
  chips?: AtomProto[] | undefined;
}

export interface ChoiceProto {
  chipValues?: number[] | undefined;
  label?: AtomProto | undefined;
}

export interface HeaderLineProto {
  title?: AtomProto | undefined;
  close?: boolean | undefined;
  share?: boolean | undefined;
}

export interface ErrorLineProto {
  error?: string | undefined;
}

export interface ActionLineProto {
  atom?: AtomProto | undefined;
  clickExpression?: NodeProto | undefined;
  inContext?: boolean | undefined;
}

export interface ComparisonHistogramProto {
  bars?: HistogramBarLineProto[] | undefined;
}

export interface ComparisonLineProto {
  baseline?: ComparisonHistogramProto | undefined;
  scenarios?: ComparisonHistogramProto[] | undefined;
}

export interface AtomProto {
  value?: string | undefined;
  labelledIcon?: LabelledIconProto | undefined;
  download?: string | undefined;
  chip?: ChipProto | undefined;
  conditional?: ConditionalProto | undefined;
}

export interface LabelledIconProto {
  icon?: string | undefined;
  label?: string | undefined;
}

export interface ChipProto {
  index?: number | undefined;
  labels?: string[] | undefined;
}

export interface ConditionProto {
  /**
   * The value of the chips specified in indices need to match
   * the respective value for the element to be rendered.
   */
  indices?: number[] | undefined;
  values?: number[] | undefined;
}

export interface ConditionalProto {
  conditions?: ConditionProto[] | undefined;
  atoms?: AtomProto[] | undefined;
}

export interface GeoJSONProto {
  condition?: ConditionProto | undefined;
  index?: number | undefined;
}

export interface FeatureIDsProto {
  namespaces?: string[] | undefined;
  ids?: IDsProto[] | undefined;
}

export interface IDsProto {
  ids?: number[] | undefined;
}

export interface ComparisonRequestProto {
  /** The ID of the analysis to run in different worlds */
  analysis?:
    | FeatureIDProto
    | undefined;
  /** The ID of the baseline world in which to run the analysis */
  baseline?:
    | FeatureIDProto
    | undefined;
  /** The IDs of the scenario worlds in which to run the analysis */
  scenarios?: FeatureIDProto[] | undefined;
}

function createBaseUIRequestProto(): UIRequestProto {
  return {
    node: undefined,
    expression: "",
    root: undefined,
    locked: false,
    logEvent: "",
    logMapCenter: undefined,
    logMapZoom: 0,
    session: 0,
  };
}

export const UIRequestProto = {
  encode(message: UIRequestProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.node !== undefined) {
      NodeProto.encode(message.node, writer.uint32(10).fork()).ldelim();
    }
    if (message.expression !== undefined && message.expression !== "") {
      writer.uint32(18).string(message.expression);
    }
    if (message.root !== undefined) {
      FeatureIDProto.encode(message.root, writer.uint32(26).fork()).ldelim();
    }
    if (message.locked !== undefined && message.locked !== false) {
      writer.uint32(32).bool(message.locked);
    }
    if (message.logEvent !== undefined && message.logEvent !== "") {
      writer.uint32(42).string(message.logEvent);
    }
    if (message.logMapCenter !== undefined) {
      PointProto.encode(message.logMapCenter, writer.uint32(50).fork()).ldelim();
    }
    if (message.logMapZoom !== undefined && message.logMapZoom !== 0) {
      writer.uint32(61).float(message.logMapZoom);
    }
    if (message.session !== undefined && message.session !== 0) {
      writer.uint32(64).uint64(message.session);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): UIRequestProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseUIRequestProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.node = NodeProto.decode(reader, reader.uint32());
          continue;
        case 2:
          if (tag !== 18) {
            break;
          }

          message.expression = reader.string();
          continue;
        case 3:
          if (tag !== 26) {
            break;
          }

          message.root = FeatureIDProto.decode(reader, reader.uint32());
          continue;
        case 4:
          if (tag !== 32) {
            break;
          }

          message.locked = reader.bool();
          continue;
        case 5:
          if (tag !== 42) {
            break;
          }

          message.logEvent = reader.string();
          continue;
        case 6:
          if (tag !== 50) {
            break;
          }

          message.logMapCenter = PointProto.decode(reader, reader.uint32());
          continue;
        case 7:
          if (tag !== 61) {
            break;
          }

          message.logMapZoom = reader.float();
          continue;
        case 8:
          if (tag !== 64) {
            break;
          }

          message.session = longToNumber(reader.uint64() as Long);
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): UIRequestProto {
    return {
      node: isSet(object.node) ? NodeProto.fromJSON(object.node) : undefined,
      expression: isSet(object.expression) ? globalThis.String(object.expression) : "",
      root: isSet(object.root) ? FeatureIDProto.fromJSON(object.root) : undefined,
      locked: isSet(object.locked) ? globalThis.Boolean(object.locked) : false,
      logEvent: isSet(object.logEvent) ? globalThis.String(object.logEvent) : "",
      logMapCenter: isSet(object.logMapCenter) ? PointProto.fromJSON(object.logMapCenter) : undefined,
      logMapZoom: isSet(object.logMapZoom) ? globalThis.Number(object.logMapZoom) : 0,
      session: isSet(object.session) ? globalThis.Number(object.session) : 0,
    };
  },

  toJSON(message: UIRequestProto): unknown {
    const obj: any = {};
    if (message.node !== undefined) {
      obj.node = NodeProto.toJSON(message.node);
    }
    if (message.expression !== undefined && message.expression !== "") {
      obj.expression = message.expression;
    }
    if (message.root !== undefined) {
      obj.root = FeatureIDProto.toJSON(message.root);
    }
    if (message.locked !== undefined && message.locked !== false) {
      obj.locked = message.locked;
    }
    if (message.logEvent !== undefined && message.logEvent !== "") {
      obj.logEvent = message.logEvent;
    }
    if (message.logMapCenter !== undefined) {
      obj.logMapCenter = PointProto.toJSON(message.logMapCenter);
    }
    if (message.logMapZoom !== undefined && message.logMapZoom !== 0) {
      obj.logMapZoom = message.logMapZoom;
    }
    if (message.session !== undefined && message.session !== 0) {
      obj.session = Math.round(message.session);
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<UIRequestProto>, I>>(base?: I): UIRequestProto {
    return UIRequestProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<UIRequestProto>, I>>(object: I): UIRequestProto {
    const message = createBaseUIRequestProto();
    message.node = (object.node !== undefined && object.node !== null) ? NodeProto.fromPartial(object.node) : undefined;
    message.expression = object.expression ?? "";
    message.root = (object.root !== undefined && object.root !== null)
      ? FeatureIDProto.fromPartial(object.root)
      : undefined;
    message.locked = object.locked ?? false;
    message.logEvent = object.logEvent ?? "";
    message.logMapCenter = (object.logMapCenter !== undefined && object.logMapCenter !== null)
      ? PointProto.fromPartial(object.logMapCenter)
      : undefined;
    message.logMapZoom = object.logMapZoom ?? 0;
    message.session = object.session ?? 0;
    return message;
  },
};

function createBaseUIResponseProto(): UIResponseProto {
  return {
    stack: undefined,
    node: undefined,
    expression: "",
    highlighted: undefined,
    geoJSON: [],
    layers: [],
    mapCenter: undefined,
    locked: false,
    chipValues: [],
    logDetail: "",
    tilesChanged: false,
  };
}

export const UIResponseProto = {
  encode(message: UIResponseProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.stack !== undefined) {
      StackProto.encode(message.stack, writer.uint32(10).fork()).ldelim();
    }
    if (message.node !== undefined) {
      NodeProto.encode(message.node, writer.uint32(18).fork()).ldelim();
    }
    if (message.expression !== undefined && message.expression !== "") {
      writer.uint32(26).string(message.expression);
    }
    if (message.highlighted !== undefined) {
      FeatureIDsProto.encode(message.highlighted, writer.uint32(34).fork()).ldelim();
    }
    if (message.geoJSON !== undefined && message.geoJSON.length !== 0) {
      for (const v of message.geoJSON) {
        GeoJSONProto.encode(v!, writer.uint32(42).fork()).ldelim();
      }
    }
    if (message.layers !== undefined && message.layers.length !== 0) {
      for (const v of message.layers) {
        MapLayerProto.encode(v!, writer.uint32(58).fork()).ldelim();
      }
    }
    if (message.mapCenter !== undefined) {
      PointProto.encode(message.mapCenter, writer.uint32(66).fork()).ldelim();
    }
    if (message.locked !== undefined && message.locked !== false) {
      writer.uint32(72).bool(message.locked);
    }
    if (message.chipValues !== undefined && message.chipValues.length !== 0) {
      writer.uint32(82).fork();
      for (const v of message.chipValues) {
        writer.int32(v);
      }
      writer.ldelim();
    }
    if (message.logDetail !== undefined && message.logDetail !== "") {
      writer.uint32(90).string(message.logDetail);
    }
    if (message.tilesChanged !== undefined && message.tilesChanged !== false) {
      writer.uint32(96).bool(message.tilesChanged);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): UIResponseProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseUIResponseProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.stack = StackProto.decode(reader, reader.uint32());
          continue;
        case 2:
          if (tag !== 18) {
            break;
          }

          message.node = NodeProto.decode(reader, reader.uint32());
          continue;
        case 3:
          if (tag !== 26) {
            break;
          }

          message.expression = reader.string();
          continue;
        case 4:
          if (tag !== 34) {
            break;
          }

          message.highlighted = FeatureIDsProto.decode(reader, reader.uint32());
          continue;
        case 5:
          if (tag !== 42) {
            break;
          }

          message.geoJSON!.push(GeoJSONProto.decode(reader, reader.uint32()));
          continue;
        case 7:
          if (tag !== 58) {
            break;
          }

          message.layers!.push(MapLayerProto.decode(reader, reader.uint32()));
          continue;
        case 8:
          if (tag !== 66) {
            break;
          }

          message.mapCenter = PointProto.decode(reader, reader.uint32());
          continue;
        case 9:
          if (tag !== 72) {
            break;
          }

          message.locked = reader.bool();
          continue;
        case 10:
          if (tag === 80) {
            message.chipValues!.push(reader.int32());

            continue;
          }

          if (tag === 82) {
            const end2 = reader.uint32() + reader.pos;
            while (reader.pos < end2) {
              message.chipValues!.push(reader.int32());
            }

            continue;
          }

          break;
        case 11:
          if (tag !== 90) {
            break;
          }

          message.logDetail = reader.string();
          continue;
        case 12:
          if (tag !== 96) {
            break;
          }

          message.tilesChanged = reader.bool();
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): UIResponseProto {
    return {
      stack: isSet(object.stack) ? StackProto.fromJSON(object.stack) : undefined,
      node: isSet(object.node) ? NodeProto.fromJSON(object.node) : undefined,
      expression: isSet(object.expression) ? globalThis.String(object.expression) : "",
      highlighted: isSet(object.highlighted) ? FeatureIDsProto.fromJSON(object.highlighted) : undefined,
      geoJSON: globalThis.Array.isArray(object?.geoJSON)
        ? object.geoJSON.map((e: any) => GeoJSONProto.fromJSON(e))
        : [],
      layers: globalThis.Array.isArray(object?.layers) ? object.layers.map((e: any) => MapLayerProto.fromJSON(e)) : [],
      mapCenter: isSet(object.mapCenter) ? PointProto.fromJSON(object.mapCenter) : undefined,
      locked: isSet(object.locked) ? globalThis.Boolean(object.locked) : false,
      chipValues: globalThis.Array.isArray(object?.chipValues)
        ? object.chipValues.map((e: any) => globalThis.Number(e))
        : [],
      logDetail: isSet(object.logDetail) ? globalThis.String(object.logDetail) : "",
      tilesChanged: isSet(object.tilesChanged) ? globalThis.Boolean(object.tilesChanged) : false,
    };
  },

  toJSON(message: UIResponseProto): unknown {
    const obj: any = {};
    if (message.stack !== undefined) {
      obj.stack = StackProto.toJSON(message.stack);
    }
    if (message.node !== undefined) {
      obj.node = NodeProto.toJSON(message.node);
    }
    if (message.expression !== undefined && message.expression !== "") {
      obj.expression = message.expression;
    }
    if (message.highlighted !== undefined) {
      obj.highlighted = FeatureIDsProto.toJSON(message.highlighted);
    }
    if (message.geoJSON?.length) {
      obj.geoJSON = message.geoJSON.map((e) => GeoJSONProto.toJSON(e));
    }
    if (message.layers?.length) {
      obj.layers = message.layers.map((e) => MapLayerProto.toJSON(e));
    }
    if (message.mapCenter !== undefined) {
      obj.mapCenter = PointProto.toJSON(message.mapCenter);
    }
    if (message.locked !== undefined && message.locked !== false) {
      obj.locked = message.locked;
    }
    if (message.chipValues?.length) {
      obj.chipValues = message.chipValues.map((e) => Math.round(e));
    }
    if (message.logDetail !== undefined && message.logDetail !== "") {
      obj.logDetail = message.logDetail;
    }
    if (message.tilesChanged !== undefined && message.tilesChanged !== false) {
      obj.tilesChanged = message.tilesChanged;
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<UIResponseProto>, I>>(base?: I): UIResponseProto {
    return UIResponseProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<UIResponseProto>, I>>(object: I): UIResponseProto {
    const message = createBaseUIResponseProto();
    message.stack = (object.stack !== undefined && object.stack !== null)
      ? StackProto.fromPartial(object.stack)
      : undefined;
    message.node = (object.node !== undefined && object.node !== null) ? NodeProto.fromPartial(object.node) : undefined;
    message.expression = object.expression ?? "";
    message.highlighted = (object.highlighted !== undefined && object.highlighted !== null)
      ? FeatureIDsProto.fromPartial(object.highlighted)
      : undefined;
    message.geoJSON = object.geoJSON?.map((e) => GeoJSONProto.fromPartial(e)) || [];
    message.layers = object.layers?.map((e) => MapLayerProto.fromPartial(e)) || [];
    message.mapCenter = (object.mapCenter !== undefined && object.mapCenter !== null)
      ? PointProto.fromPartial(object.mapCenter)
      : undefined;
    message.locked = object.locked ?? false;
    message.chipValues = object.chipValues?.map((e) => e) || [];
    message.logDetail = object.logDetail ?? "";
    message.tilesChanged = object.tilesChanged ?? false;
    return message;
  },
};

function createBaseMapLayerProto(): MapLayerProto {
  return { path: "", q: "", v: "", before: 0, condition: undefined };
}

export const MapLayerProto = {
  encode(message: MapLayerProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.path !== undefined && message.path !== "") {
      writer.uint32(10).string(message.path);
    }
    if (message.q !== undefined && message.q !== "") {
      writer.uint32(18).string(message.q);
    }
    if (message.v !== undefined && message.v !== "") {
      writer.uint32(26).string(message.v);
    }
    if (message.before !== undefined && message.before !== 0) {
      writer.uint32(32).int32(message.before);
    }
    if (message.condition !== undefined) {
      ConditionProto.encode(message.condition, writer.uint32(42).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): MapLayerProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseMapLayerProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.path = reader.string();
          continue;
        case 2:
          if (tag !== 18) {
            break;
          }

          message.q = reader.string();
          continue;
        case 3:
          if (tag !== 26) {
            break;
          }

          message.v = reader.string();
          continue;
        case 4:
          if (tag !== 32) {
            break;
          }

          message.before = reader.int32() as any;
          continue;
        case 5:
          if (tag !== 42) {
            break;
          }

          message.condition = ConditionProto.decode(reader, reader.uint32());
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): MapLayerProto {
    return {
      path: isSet(object.path) ? globalThis.String(object.path) : "",
      q: isSet(object.q) ? globalThis.String(object.q) : "",
      v: isSet(object.v) ? globalThis.String(object.v) : "",
      before: isSet(object.before) ? mapLayerPositionFromJSON(object.before) : 0,
      condition: isSet(object.condition) ? ConditionProto.fromJSON(object.condition) : undefined,
    };
  },

  toJSON(message: MapLayerProto): unknown {
    const obj: any = {};
    if (message.path !== undefined && message.path !== "") {
      obj.path = message.path;
    }
    if (message.q !== undefined && message.q !== "") {
      obj.q = message.q;
    }
    if (message.v !== undefined && message.v !== "") {
      obj.v = message.v;
    }
    if (message.before !== undefined && message.before !== 0) {
      obj.before = mapLayerPositionToJSON(message.before);
    }
    if (message.condition !== undefined) {
      obj.condition = ConditionProto.toJSON(message.condition);
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<MapLayerProto>, I>>(base?: I): MapLayerProto {
    return MapLayerProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<MapLayerProto>, I>>(object: I): MapLayerProto {
    const message = createBaseMapLayerProto();
    message.path = object.path ?? "";
    message.q = object.q ?? "";
    message.v = object.v ?? "";
    message.before = object.before ?? 0;
    message.condition = (object.condition !== undefined && object.condition !== null)
      ? ConditionProto.fromPartial(object.condition)
      : undefined;
    return message;
  },
};

function createBaseStackProto(): StackProto {
  return { substacks: [], id: undefined };
}

export const StackProto = {
  encode(message: StackProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.substacks !== undefined && message.substacks.length !== 0) {
      for (const v of message.substacks) {
        SubstackProto.encode(v!, writer.uint32(10).fork()).ldelim();
      }
    }
    if (message.id !== undefined) {
      FeatureIDProto.encode(message.id, writer.uint32(18).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): StackProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseStackProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.substacks!.push(SubstackProto.decode(reader, reader.uint32()));
          continue;
        case 2:
          if (tag !== 18) {
            break;
          }

          message.id = FeatureIDProto.decode(reader, reader.uint32());
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): StackProto {
    return {
      substacks: globalThis.Array.isArray(object?.substacks)
        ? object.substacks.map((e: any) => SubstackProto.fromJSON(e))
        : [],
      id: isSet(object.id) ? FeatureIDProto.fromJSON(object.id) : undefined,
    };
  },

  toJSON(message: StackProto): unknown {
    const obj: any = {};
    if (message.substacks?.length) {
      obj.substacks = message.substacks.map((e) => SubstackProto.toJSON(e));
    }
    if (message.id !== undefined) {
      obj.id = FeatureIDProto.toJSON(message.id);
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<StackProto>, I>>(base?: I): StackProto {
    return StackProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<StackProto>, I>>(object: I): StackProto {
    const message = createBaseStackProto();
    message.substacks = object.substacks?.map((e) => SubstackProto.fromPartial(e)) || [];
    message.id = (object.id !== undefined && object.id !== null) ? FeatureIDProto.fromPartial(object.id) : undefined;
    return message;
  },
};

function createBaseSubstackProto(): SubstackProto {
  return { lines: [], collapsable: false };
}

export const SubstackProto = {
  encode(message: SubstackProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.lines !== undefined && message.lines.length !== 0) {
      for (const v of message.lines) {
        LineProto.encode(v!, writer.uint32(10).fork()).ldelim();
      }
    }
    if (message.collapsable !== undefined && message.collapsable !== false) {
      writer.uint32(16).bool(message.collapsable);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): SubstackProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseSubstackProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.lines!.push(LineProto.decode(reader, reader.uint32()));
          continue;
        case 2:
          if (tag !== 16) {
            break;
          }

          message.collapsable = reader.bool();
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): SubstackProto {
    return {
      lines: globalThis.Array.isArray(object?.lines) ? object.lines.map((e: any) => LineProto.fromJSON(e)) : [],
      collapsable: isSet(object.collapsable) ? globalThis.Boolean(object.collapsable) : false,
    };
  },

  toJSON(message: SubstackProto): unknown {
    const obj: any = {};
    if (message.lines?.length) {
      obj.lines = message.lines.map((e) => LineProto.toJSON(e));
    }
    if (message.collapsable !== undefined && message.collapsable !== false) {
      obj.collapsable = message.collapsable;
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<SubstackProto>, I>>(base?: I): SubstackProto {
    return SubstackProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<SubstackProto>, I>>(object: I): SubstackProto {
    const message = createBaseSubstackProto();
    message.lines = object.lines?.map((e) => LineProto.fromPartial(e)) || [];
    message.collapsable = object.collapsable ?? false;
    return message;
  },
};

function createBaseLineProto(): LineProto {
  return {
    value: undefined,
    leftRightValue: undefined,
    expression: undefined,
    tags: undefined,
    histogramBar: undefined,
    swatch: undefined,
    shell: undefined,
    choice: undefined,
    header: undefined,
    error: undefined,
    action: undefined,
    comparison: undefined,
  };
}

export const LineProto = {
  encode(message: LineProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.value !== undefined) {
      ValueLineProto.encode(message.value, writer.uint32(10).fork()).ldelim();
    }
    if (message.leftRightValue !== undefined) {
      LeftRightValueLineProto.encode(message.leftRightValue, writer.uint32(18).fork()).ldelim();
    }
    if (message.expression !== undefined) {
      ExpressionLineProto.encode(message.expression, writer.uint32(26).fork()).ldelim();
    }
    if (message.tags !== undefined) {
      TagsLineProto.encode(message.tags, writer.uint32(34).fork()).ldelim();
    }
    if (message.histogramBar !== undefined) {
      HistogramBarLineProto.encode(message.histogramBar, writer.uint32(42).fork()).ldelim();
    }
    if (message.swatch !== undefined) {
      SwatchLineProto.encode(message.swatch, writer.uint32(50).fork()).ldelim();
    }
    if (message.shell !== undefined) {
      ShellLineProto.encode(message.shell, writer.uint32(58).fork()).ldelim();
    }
    if (message.choice !== undefined) {
      ChoiceLineProto.encode(message.choice, writer.uint32(66).fork()).ldelim();
    }
    if (message.header !== undefined) {
      HeaderLineProto.encode(message.header, writer.uint32(74).fork()).ldelim();
    }
    if (message.error !== undefined) {
      ErrorLineProto.encode(message.error, writer.uint32(82).fork()).ldelim();
    }
    if (message.action !== undefined) {
      ActionLineProto.encode(message.action, writer.uint32(90).fork()).ldelim();
    }
    if (message.comparison !== undefined) {
      ComparisonLineProto.encode(message.comparison, writer.uint32(98).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): LineProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseLineProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.value = ValueLineProto.decode(reader, reader.uint32());
          continue;
        case 2:
          if (tag !== 18) {
            break;
          }

          message.leftRightValue = LeftRightValueLineProto.decode(reader, reader.uint32());
          continue;
        case 3:
          if (tag !== 26) {
            break;
          }

          message.expression = ExpressionLineProto.decode(reader, reader.uint32());
          continue;
        case 4:
          if (tag !== 34) {
            break;
          }

          message.tags = TagsLineProto.decode(reader, reader.uint32());
          continue;
        case 5:
          if (tag !== 42) {
            break;
          }

          message.histogramBar = HistogramBarLineProto.decode(reader, reader.uint32());
          continue;
        case 6:
          if (tag !== 50) {
            break;
          }

          message.swatch = SwatchLineProto.decode(reader, reader.uint32());
          continue;
        case 7:
          if (tag !== 58) {
            break;
          }

          message.shell = ShellLineProto.decode(reader, reader.uint32());
          continue;
        case 8:
          if (tag !== 66) {
            break;
          }

          message.choice = ChoiceLineProto.decode(reader, reader.uint32());
          continue;
        case 9:
          if (tag !== 74) {
            break;
          }

          message.header = HeaderLineProto.decode(reader, reader.uint32());
          continue;
        case 10:
          if (tag !== 82) {
            break;
          }

          message.error = ErrorLineProto.decode(reader, reader.uint32());
          continue;
        case 11:
          if (tag !== 90) {
            break;
          }

          message.action = ActionLineProto.decode(reader, reader.uint32());
          continue;
        case 12:
          if (tag !== 98) {
            break;
          }

          message.comparison = ComparisonLineProto.decode(reader, reader.uint32());
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): LineProto {
    return {
      value: isSet(object.value) ? ValueLineProto.fromJSON(object.value) : undefined,
      leftRightValue: isSet(object.leftRightValue)
        ? LeftRightValueLineProto.fromJSON(object.leftRightValue)
        : undefined,
      expression: isSet(object.expression) ? ExpressionLineProto.fromJSON(object.expression) : undefined,
      tags: isSet(object.tags) ? TagsLineProto.fromJSON(object.tags) : undefined,
      histogramBar: isSet(object.histogramBar) ? HistogramBarLineProto.fromJSON(object.histogramBar) : undefined,
      swatch: isSet(object.swatch) ? SwatchLineProto.fromJSON(object.swatch) : undefined,
      shell: isSet(object.shell) ? ShellLineProto.fromJSON(object.shell) : undefined,
      choice: isSet(object.choice) ? ChoiceLineProto.fromJSON(object.choice) : undefined,
      header: isSet(object.header) ? HeaderLineProto.fromJSON(object.header) : undefined,
      error: isSet(object.error) ? ErrorLineProto.fromJSON(object.error) : undefined,
      action: isSet(object.action) ? ActionLineProto.fromJSON(object.action) : undefined,
      comparison: isSet(object.comparison) ? ComparisonLineProto.fromJSON(object.comparison) : undefined,
    };
  },

  toJSON(message: LineProto): unknown {
    const obj: any = {};
    if (message.value !== undefined) {
      obj.value = ValueLineProto.toJSON(message.value);
    }
    if (message.leftRightValue !== undefined) {
      obj.leftRightValue = LeftRightValueLineProto.toJSON(message.leftRightValue);
    }
    if (message.expression !== undefined) {
      obj.expression = ExpressionLineProto.toJSON(message.expression);
    }
    if (message.tags !== undefined) {
      obj.tags = TagsLineProto.toJSON(message.tags);
    }
    if (message.histogramBar !== undefined) {
      obj.histogramBar = HistogramBarLineProto.toJSON(message.histogramBar);
    }
    if (message.swatch !== undefined) {
      obj.swatch = SwatchLineProto.toJSON(message.swatch);
    }
    if (message.shell !== undefined) {
      obj.shell = ShellLineProto.toJSON(message.shell);
    }
    if (message.choice !== undefined) {
      obj.choice = ChoiceLineProto.toJSON(message.choice);
    }
    if (message.header !== undefined) {
      obj.header = HeaderLineProto.toJSON(message.header);
    }
    if (message.error !== undefined) {
      obj.error = ErrorLineProto.toJSON(message.error);
    }
    if (message.action !== undefined) {
      obj.action = ActionLineProto.toJSON(message.action);
    }
    if (message.comparison !== undefined) {
      obj.comparison = ComparisonLineProto.toJSON(message.comparison);
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<LineProto>, I>>(base?: I): LineProto {
    return LineProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<LineProto>, I>>(object: I): LineProto {
    const message = createBaseLineProto();
    message.value = (object.value !== undefined && object.value !== null)
      ? ValueLineProto.fromPartial(object.value)
      : undefined;
    message.leftRightValue = (object.leftRightValue !== undefined && object.leftRightValue !== null)
      ? LeftRightValueLineProto.fromPartial(object.leftRightValue)
      : undefined;
    message.expression = (object.expression !== undefined && object.expression !== null)
      ? ExpressionLineProto.fromPartial(object.expression)
      : undefined;
    message.tags = (object.tags !== undefined && object.tags !== null)
      ? TagsLineProto.fromPartial(object.tags)
      : undefined;
    message.histogramBar = (object.histogramBar !== undefined && object.histogramBar !== null)
      ? HistogramBarLineProto.fromPartial(object.histogramBar)
      : undefined;
    message.swatch = (object.swatch !== undefined && object.swatch !== null)
      ? SwatchLineProto.fromPartial(object.swatch)
      : undefined;
    message.shell = (object.shell !== undefined && object.shell !== null)
      ? ShellLineProto.fromPartial(object.shell)
      : undefined;
    message.choice = (object.choice !== undefined && object.choice !== null)
      ? ChoiceLineProto.fromPartial(object.choice)
      : undefined;
    message.header = (object.header !== undefined && object.header !== null)
      ? HeaderLineProto.fromPartial(object.header)
      : undefined;
    message.error = (object.error !== undefined && object.error !== null)
      ? ErrorLineProto.fromPartial(object.error)
      : undefined;
    message.action = (object.action !== undefined && object.action !== null)
      ? ActionLineProto.fromPartial(object.action)
      : undefined;
    message.comparison = (object.comparison !== undefined && object.comparison !== null)
      ? ComparisonLineProto.fromPartial(object.comparison)
      : undefined;
    return message;
  },
};

function createBaseValueLineProto(): ValueLineProto {
  return { atom: undefined, clickExpression: undefined };
}

export const ValueLineProto = {
  encode(message: ValueLineProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.atom !== undefined) {
      AtomProto.encode(message.atom, writer.uint32(10).fork()).ldelim();
    }
    if (message.clickExpression !== undefined) {
      NodeProto.encode(message.clickExpression, writer.uint32(18).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): ValueLineProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseValueLineProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.atom = AtomProto.decode(reader, reader.uint32());
          continue;
        case 2:
          if (tag !== 18) {
            break;
          }

          message.clickExpression = NodeProto.decode(reader, reader.uint32());
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): ValueLineProto {
    return {
      atom: isSet(object.atom) ? AtomProto.fromJSON(object.atom) : undefined,
      clickExpression: isSet(object.clickExpression) ? NodeProto.fromJSON(object.clickExpression) : undefined,
    };
  },

  toJSON(message: ValueLineProto): unknown {
    const obj: any = {};
    if (message.atom !== undefined) {
      obj.atom = AtomProto.toJSON(message.atom);
    }
    if (message.clickExpression !== undefined) {
      obj.clickExpression = NodeProto.toJSON(message.clickExpression);
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<ValueLineProto>, I>>(base?: I): ValueLineProto {
    return ValueLineProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<ValueLineProto>, I>>(object: I): ValueLineProto {
    const message = createBaseValueLineProto();
    message.atom = (object.atom !== undefined && object.atom !== null) ? AtomProto.fromPartial(object.atom) : undefined;
    message.clickExpression = (object.clickExpression !== undefined && object.clickExpression !== null)
      ? NodeProto.fromPartial(object.clickExpression)
      : undefined;
    return message;
  },
};

function createBaseLeftRightValueLineProto(): LeftRightValueLineProto {
  return { left: [], right: undefined };
}

export const LeftRightValueLineProto = {
  encode(message: LeftRightValueLineProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.left !== undefined && message.left.length !== 0) {
      for (const v of message.left) {
        ClickableAtomProto.encode(v!, writer.uint32(10).fork()).ldelim();
      }
    }
    if (message.right !== undefined) {
      ClickableAtomProto.encode(message.right, writer.uint32(18).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): LeftRightValueLineProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseLeftRightValueLineProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.left!.push(ClickableAtomProto.decode(reader, reader.uint32()));
          continue;
        case 2:
          if (tag !== 18) {
            break;
          }

          message.right = ClickableAtomProto.decode(reader, reader.uint32());
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): LeftRightValueLineProto {
    return {
      left: globalThis.Array.isArray(object?.left) ? object.left.map((e: any) => ClickableAtomProto.fromJSON(e)) : [],
      right: isSet(object.right) ? ClickableAtomProto.fromJSON(object.right) : undefined,
    };
  },

  toJSON(message: LeftRightValueLineProto): unknown {
    const obj: any = {};
    if (message.left?.length) {
      obj.left = message.left.map((e) => ClickableAtomProto.toJSON(e));
    }
    if (message.right !== undefined) {
      obj.right = ClickableAtomProto.toJSON(message.right);
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<LeftRightValueLineProto>, I>>(base?: I): LeftRightValueLineProto {
    return LeftRightValueLineProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<LeftRightValueLineProto>, I>>(object: I): LeftRightValueLineProto {
    const message = createBaseLeftRightValueLineProto();
    message.left = object.left?.map((e) => ClickableAtomProto.fromPartial(e)) || [];
    message.right = (object.right !== undefined && object.right !== null)
      ? ClickableAtomProto.fromPartial(object.right)
      : undefined;
    return message;
  },
};

function createBaseClickableAtomProto(): ClickableAtomProto {
  return { atom: undefined, clickExpression: undefined };
}

export const ClickableAtomProto = {
  encode(message: ClickableAtomProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.atom !== undefined) {
      AtomProto.encode(message.atom, writer.uint32(10).fork()).ldelim();
    }
    if (message.clickExpression !== undefined) {
      NodeProto.encode(message.clickExpression, writer.uint32(18).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): ClickableAtomProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseClickableAtomProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.atom = AtomProto.decode(reader, reader.uint32());
          continue;
        case 2:
          if (tag !== 18) {
            break;
          }

          message.clickExpression = NodeProto.decode(reader, reader.uint32());
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): ClickableAtomProto {
    return {
      atom: isSet(object.atom) ? AtomProto.fromJSON(object.atom) : undefined,
      clickExpression: isSet(object.clickExpression) ? NodeProto.fromJSON(object.clickExpression) : undefined,
    };
  },

  toJSON(message: ClickableAtomProto): unknown {
    const obj: any = {};
    if (message.atom !== undefined) {
      obj.atom = AtomProto.toJSON(message.atom);
    }
    if (message.clickExpression !== undefined) {
      obj.clickExpression = NodeProto.toJSON(message.clickExpression);
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<ClickableAtomProto>, I>>(base?: I): ClickableAtomProto {
    return ClickableAtomProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<ClickableAtomProto>, I>>(object: I): ClickableAtomProto {
    const message = createBaseClickableAtomProto();
    message.atom = (object.atom !== undefined && object.atom !== null) ? AtomProto.fromPartial(object.atom) : undefined;
    message.clickExpression = (object.clickExpression !== undefined && object.clickExpression !== null)
      ? NodeProto.fromPartial(object.clickExpression)
      : undefined;
    return message;
  },
};

function createBaseExpressionLineProto(): ExpressionLineProto {
  return { expression: "" };
}

export const ExpressionLineProto = {
  encode(message: ExpressionLineProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.expression !== undefined && message.expression !== "") {
      writer.uint32(10).string(message.expression);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): ExpressionLineProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseExpressionLineProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.expression = reader.string();
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): ExpressionLineProto {
    return { expression: isSet(object.expression) ? globalThis.String(object.expression) : "" };
  },

  toJSON(message: ExpressionLineProto): unknown {
    const obj: any = {};
    if (message.expression !== undefined && message.expression !== "") {
      obj.expression = message.expression;
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<ExpressionLineProto>, I>>(base?: I): ExpressionLineProto {
    return ExpressionLineProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<ExpressionLineProto>, I>>(object: I): ExpressionLineProto {
    const message = createBaseExpressionLineProto();
    message.expression = object.expression ?? "";
    return message;
  },
};

function createBaseTagsLineProto(): TagsLineProto {
  return { tags: [] };
}

export const TagsLineProto = {
  encode(message: TagsLineProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.tags !== undefined && message.tags.length !== 0) {
      for (const v of message.tags) {
        TagAtomProto.encode(v!, writer.uint32(10).fork()).ldelim();
      }
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): TagsLineProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseTagsLineProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.tags!.push(TagAtomProto.decode(reader, reader.uint32()));
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): TagsLineProto {
    return {
      tags: globalThis.Array.isArray(object?.tags) ? object.tags.map((e: any) => TagAtomProto.fromJSON(e)) : [],
    };
  },

  toJSON(message: TagsLineProto): unknown {
    const obj: any = {};
    if (message.tags?.length) {
      obj.tags = message.tags.map((e) => TagAtomProto.toJSON(e));
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<TagsLineProto>, I>>(base?: I): TagsLineProto {
    return TagsLineProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<TagsLineProto>, I>>(object: I): TagsLineProto {
    const message = createBaseTagsLineProto();
    message.tags = object.tags?.map((e) => TagAtomProto.fromPartial(e)) || [];
    return message;
  },
};

function createBaseTagAtomProto(): TagAtomProto {
  return { prefix: "", key: "", value: "", clickExpression: undefined };
}

export const TagAtomProto = {
  encode(message: TagAtomProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.prefix !== undefined && message.prefix !== "") {
      writer.uint32(10).string(message.prefix);
    }
    if (message.key !== undefined && message.key !== "") {
      writer.uint32(18).string(message.key);
    }
    if (message.value !== undefined && message.value !== "") {
      writer.uint32(26).string(message.value);
    }
    if (message.clickExpression !== undefined) {
      NodeProto.encode(message.clickExpression, writer.uint32(34).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): TagAtomProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseTagAtomProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.prefix = reader.string();
          continue;
        case 2:
          if (tag !== 18) {
            break;
          }

          message.key = reader.string();
          continue;
        case 3:
          if (tag !== 26) {
            break;
          }

          message.value = reader.string();
          continue;
        case 4:
          if (tag !== 34) {
            break;
          }

          message.clickExpression = NodeProto.decode(reader, reader.uint32());
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): TagAtomProto {
    return {
      prefix: isSet(object.prefix) ? globalThis.String(object.prefix) : "",
      key: isSet(object.key) ? globalThis.String(object.key) : "",
      value: isSet(object.value) ? globalThis.String(object.value) : "",
      clickExpression: isSet(object.clickExpression) ? NodeProto.fromJSON(object.clickExpression) : undefined,
    };
  },

  toJSON(message: TagAtomProto): unknown {
    const obj: any = {};
    if (message.prefix !== undefined && message.prefix !== "") {
      obj.prefix = message.prefix;
    }
    if (message.key !== undefined && message.key !== "") {
      obj.key = message.key;
    }
    if (message.value !== undefined && message.value !== "") {
      obj.value = message.value;
    }
    if (message.clickExpression !== undefined) {
      obj.clickExpression = NodeProto.toJSON(message.clickExpression);
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<TagAtomProto>, I>>(base?: I): TagAtomProto {
    return TagAtomProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<TagAtomProto>, I>>(object: I): TagAtomProto {
    const message = createBaseTagAtomProto();
    message.prefix = object.prefix ?? "";
    message.key = object.key ?? "";
    message.value = object.value ?? "";
    message.clickExpression = (object.clickExpression !== undefined && object.clickExpression !== null)
      ? NodeProto.fromPartial(object.clickExpression)
      : undefined;
    return message;
  },
};

function createBaseHistogramBarLineProto(): HistogramBarLineProto {
  return { range: undefined, value: 0, total: 0, index: 0 };
}

export const HistogramBarLineProto = {
  encode(message: HistogramBarLineProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.range !== undefined) {
      AtomProto.encode(message.range, writer.uint32(10).fork()).ldelim();
    }
    if (message.value !== undefined && message.value !== 0) {
      writer.uint32(16).int32(message.value);
    }
    if (message.total !== undefined && message.total !== 0) {
      writer.uint32(24).int32(message.total);
    }
    if (message.index !== undefined && message.index !== 0) {
      writer.uint32(32).int32(message.index);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): HistogramBarLineProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseHistogramBarLineProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.range = AtomProto.decode(reader, reader.uint32());
          continue;
        case 2:
          if (tag !== 16) {
            break;
          }

          message.value = reader.int32();
          continue;
        case 3:
          if (tag !== 24) {
            break;
          }

          message.total = reader.int32();
          continue;
        case 4:
          if (tag !== 32) {
            break;
          }

          message.index = reader.int32();
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): HistogramBarLineProto {
    return {
      range: isSet(object.range) ? AtomProto.fromJSON(object.range) : undefined,
      value: isSet(object.value) ? globalThis.Number(object.value) : 0,
      total: isSet(object.total) ? globalThis.Number(object.total) : 0,
      index: isSet(object.index) ? globalThis.Number(object.index) : 0,
    };
  },

  toJSON(message: HistogramBarLineProto): unknown {
    const obj: any = {};
    if (message.range !== undefined) {
      obj.range = AtomProto.toJSON(message.range);
    }
    if (message.value !== undefined && message.value !== 0) {
      obj.value = Math.round(message.value);
    }
    if (message.total !== undefined && message.total !== 0) {
      obj.total = Math.round(message.total);
    }
    if (message.index !== undefined && message.index !== 0) {
      obj.index = Math.round(message.index);
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<HistogramBarLineProto>, I>>(base?: I): HistogramBarLineProto {
    return HistogramBarLineProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<HistogramBarLineProto>, I>>(object: I): HistogramBarLineProto {
    const message = createBaseHistogramBarLineProto();
    message.range = (object.range !== undefined && object.range !== null)
      ? AtomProto.fromPartial(object.range)
      : undefined;
    message.value = object.value ?? 0;
    message.total = object.total ?? 0;
    message.index = object.index ?? 0;
    return message;
  },
};

function createBaseSwatchLineProto(): SwatchLineProto {
  return { label: undefined, index: 0 };
}

export const SwatchLineProto = {
  encode(message: SwatchLineProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.label !== undefined) {
      AtomProto.encode(message.label, writer.uint32(10).fork()).ldelim();
    }
    if (message.index !== undefined && message.index !== 0) {
      writer.uint32(16).int32(message.index);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): SwatchLineProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseSwatchLineProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.label = AtomProto.decode(reader, reader.uint32());
          continue;
        case 2:
          if (tag !== 16) {
            break;
          }

          message.index = reader.int32();
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): SwatchLineProto {
    return {
      label: isSet(object.label) ? AtomProto.fromJSON(object.label) : undefined,
      index: isSet(object.index) ? globalThis.Number(object.index) : 0,
    };
  },

  toJSON(message: SwatchLineProto): unknown {
    const obj: any = {};
    if (message.label !== undefined) {
      obj.label = AtomProto.toJSON(message.label);
    }
    if (message.index !== undefined && message.index !== 0) {
      obj.index = Math.round(message.index);
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<SwatchLineProto>, I>>(base?: I): SwatchLineProto {
    return SwatchLineProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<SwatchLineProto>, I>>(object: I): SwatchLineProto {
    const message = createBaseSwatchLineProto();
    message.label = (object.label !== undefined && object.label !== null)
      ? AtomProto.fromPartial(object.label)
      : undefined;
    message.index = object.index ?? 0;
    return message;
  },
};

function createBaseShellLineProto(): ShellLineProto {
  return { functions: [] };
}

export const ShellLineProto = {
  encode(message: ShellLineProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.functions !== undefined && message.functions.length !== 0) {
      for (const v of message.functions) {
        writer.uint32(10).string(v!);
      }
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): ShellLineProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseShellLineProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.functions!.push(reader.string());
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): ShellLineProto {
    return {
      functions: globalThis.Array.isArray(object?.functions)
        ? object.functions.map((e: any) => globalThis.String(e))
        : [],
    };
  },

  toJSON(message: ShellLineProto): unknown {
    const obj: any = {};
    if (message.functions?.length) {
      obj.functions = message.functions;
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<ShellLineProto>, I>>(base?: I): ShellLineProto {
    return ShellLineProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<ShellLineProto>, I>>(object: I): ShellLineProto {
    const message = createBaseShellLineProto();
    message.functions = object.functions?.map((e) => e) || [];
    return message;
  },
};

function createBaseChoiceLineProto(): ChoiceLineProto {
  return { label: undefined, chips: [] };
}

export const ChoiceLineProto = {
  encode(message: ChoiceLineProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.label !== undefined) {
      AtomProto.encode(message.label, writer.uint32(10).fork()).ldelim();
    }
    if (message.chips !== undefined && message.chips.length !== 0) {
      for (const v of message.chips) {
        AtomProto.encode(v!, writer.uint32(18).fork()).ldelim();
      }
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): ChoiceLineProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseChoiceLineProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.label = AtomProto.decode(reader, reader.uint32());
          continue;
        case 2:
          if (tag !== 18) {
            break;
          }

          message.chips!.push(AtomProto.decode(reader, reader.uint32()));
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): ChoiceLineProto {
    return {
      label: isSet(object.label) ? AtomProto.fromJSON(object.label) : undefined,
      chips: globalThis.Array.isArray(object?.chips) ? object.chips.map((e: any) => AtomProto.fromJSON(e)) : [],
    };
  },

  toJSON(message: ChoiceLineProto): unknown {
    const obj: any = {};
    if (message.label !== undefined) {
      obj.label = AtomProto.toJSON(message.label);
    }
    if (message.chips?.length) {
      obj.chips = message.chips.map((e) => AtomProto.toJSON(e));
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<ChoiceLineProto>, I>>(base?: I): ChoiceLineProto {
    return ChoiceLineProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<ChoiceLineProto>, I>>(object: I): ChoiceLineProto {
    const message = createBaseChoiceLineProto();
    message.label = (object.label !== undefined && object.label !== null)
      ? AtomProto.fromPartial(object.label)
      : undefined;
    message.chips = object.chips?.map((e) => AtomProto.fromPartial(e)) || [];
    return message;
  },
};

function createBaseChoiceProto(): ChoiceProto {
  return { chipValues: [], label: undefined };
}

export const ChoiceProto = {
  encode(message: ChoiceProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.chipValues !== undefined && message.chipValues.length !== 0) {
      writer.uint32(10).fork();
      for (const v of message.chipValues) {
        writer.int32(v);
      }
      writer.ldelim();
    }
    if (message.label !== undefined) {
      AtomProto.encode(message.label, writer.uint32(18).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): ChoiceProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseChoiceProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag === 8) {
            message.chipValues!.push(reader.int32());

            continue;
          }

          if (tag === 10) {
            const end2 = reader.uint32() + reader.pos;
            while (reader.pos < end2) {
              message.chipValues!.push(reader.int32());
            }

            continue;
          }

          break;
        case 2:
          if (tag !== 18) {
            break;
          }

          message.label = AtomProto.decode(reader, reader.uint32());
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): ChoiceProto {
    return {
      chipValues: globalThis.Array.isArray(object?.chipValues)
        ? object.chipValues.map((e: any) => globalThis.Number(e))
        : [],
      label: isSet(object.label) ? AtomProto.fromJSON(object.label) : undefined,
    };
  },

  toJSON(message: ChoiceProto): unknown {
    const obj: any = {};
    if (message.chipValues?.length) {
      obj.chipValues = message.chipValues.map((e) => Math.round(e));
    }
    if (message.label !== undefined) {
      obj.label = AtomProto.toJSON(message.label);
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<ChoiceProto>, I>>(base?: I): ChoiceProto {
    return ChoiceProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<ChoiceProto>, I>>(object: I): ChoiceProto {
    const message = createBaseChoiceProto();
    message.chipValues = object.chipValues?.map((e) => e) || [];
    message.label = (object.label !== undefined && object.label !== null)
      ? AtomProto.fromPartial(object.label)
      : undefined;
    return message;
  },
};

function createBaseHeaderLineProto(): HeaderLineProto {
  return { title: undefined, close: false, share: false };
}

export const HeaderLineProto = {
  encode(message: HeaderLineProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.title !== undefined) {
      AtomProto.encode(message.title, writer.uint32(10).fork()).ldelim();
    }
    if (message.close !== undefined && message.close !== false) {
      writer.uint32(16).bool(message.close);
    }
    if (message.share !== undefined && message.share !== false) {
      writer.uint32(24).bool(message.share);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): HeaderLineProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseHeaderLineProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.title = AtomProto.decode(reader, reader.uint32());
          continue;
        case 2:
          if (tag !== 16) {
            break;
          }

          message.close = reader.bool();
          continue;
        case 3:
          if (tag !== 24) {
            break;
          }

          message.share = reader.bool();
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): HeaderLineProto {
    return {
      title: isSet(object.title) ? AtomProto.fromJSON(object.title) : undefined,
      close: isSet(object.close) ? globalThis.Boolean(object.close) : false,
      share: isSet(object.share) ? globalThis.Boolean(object.share) : false,
    };
  },

  toJSON(message: HeaderLineProto): unknown {
    const obj: any = {};
    if (message.title !== undefined) {
      obj.title = AtomProto.toJSON(message.title);
    }
    if (message.close !== undefined && message.close !== false) {
      obj.close = message.close;
    }
    if (message.share !== undefined && message.share !== false) {
      obj.share = message.share;
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<HeaderLineProto>, I>>(base?: I): HeaderLineProto {
    return HeaderLineProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<HeaderLineProto>, I>>(object: I): HeaderLineProto {
    const message = createBaseHeaderLineProto();
    message.title = (object.title !== undefined && object.title !== null)
      ? AtomProto.fromPartial(object.title)
      : undefined;
    message.close = object.close ?? false;
    message.share = object.share ?? false;
    return message;
  },
};

function createBaseErrorLineProto(): ErrorLineProto {
  return { error: "" };
}

export const ErrorLineProto = {
  encode(message: ErrorLineProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.error !== undefined && message.error !== "") {
      writer.uint32(10).string(message.error);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): ErrorLineProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseErrorLineProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.error = reader.string();
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): ErrorLineProto {
    return { error: isSet(object.error) ? globalThis.String(object.error) : "" };
  },

  toJSON(message: ErrorLineProto): unknown {
    const obj: any = {};
    if (message.error !== undefined && message.error !== "") {
      obj.error = message.error;
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<ErrorLineProto>, I>>(base?: I): ErrorLineProto {
    return ErrorLineProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<ErrorLineProto>, I>>(object: I): ErrorLineProto {
    const message = createBaseErrorLineProto();
    message.error = object.error ?? "";
    return message;
  },
};

function createBaseActionLineProto(): ActionLineProto {
  return { atom: undefined, clickExpression: undefined, inContext: false };
}

export const ActionLineProto = {
  encode(message: ActionLineProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.atom !== undefined) {
      AtomProto.encode(message.atom, writer.uint32(10).fork()).ldelim();
    }
    if (message.clickExpression !== undefined) {
      NodeProto.encode(message.clickExpression, writer.uint32(18).fork()).ldelim();
    }
    if (message.inContext !== undefined && message.inContext !== false) {
      writer.uint32(24).bool(message.inContext);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): ActionLineProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseActionLineProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.atom = AtomProto.decode(reader, reader.uint32());
          continue;
        case 2:
          if (tag !== 18) {
            break;
          }

          message.clickExpression = NodeProto.decode(reader, reader.uint32());
          continue;
        case 3:
          if (tag !== 24) {
            break;
          }

          message.inContext = reader.bool();
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): ActionLineProto {
    return {
      atom: isSet(object.atom) ? AtomProto.fromJSON(object.atom) : undefined,
      clickExpression: isSet(object.clickExpression) ? NodeProto.fromJSON(object.clickExpression) : undefined,
      inContext: isSet(object.inContext) ? globalThis.Boolean(object.inContext) : false,
    };
  },

  toJSON(message: ActionLineProto): unknown {
    const obj: any = {};
    if (message.atom !== undefined) {
      obj.atom = AtomProto.toJSON(message.atom);
    }
    if (message.clickExpression !== undefined) {
      obj.clickExpression = NodeProto.toJSON(message.clickExpression);
    }
    if (message.inContext !== undefined && message.inContext !== false) {
      obj.inContext = message.inContext;
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<ActionLineProto>, I>>(base?: I): ActionLineProto {
    return ActionLineProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<ActionLineProto>, I>>(object: I): ActionLineProto {
    const message = createBaseActionLineProto();
    message.atom = (object.atom !== undefined && object.atom !== null) ? AtomProto.fromPartial(object.atom) : undefined;
    message.clickExpression = (object.clickExpression !== undefined && object.clickExpression !== null)
      ? NodeProto.fromPartial(object.clickExpression)
      : undefined;
    message.inContext = object.inContext ?? false;
    return message;
  },
};

function createBaseComparisonHistogramProto(): ComparisonHistogramProto {
  return { bars: [] };
}

export const ComparisonHistogramProto = {
  encode(message: ComparisonHistogramProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.bars !== undefined && message.bars.length !== 0) {
      for (const v of message.bars) {
        HistogramBarLineProto.encode(v!, writer.uint32(10).fork()).ldelim();
      }
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): ComparisonHistogramProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseComparisonHistogramProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.bars!.push(HistogramBarLineProto.decode(reader, reader.uint32()));
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): ComparisonHistogramProto {
    return {
      bars: globalThis.Array.isArray(object?.bars)
        ? object.bars.map((e: any) => HistogramBarLineProto.fromJSON(e))
        : [],
    };
  },

  toJSON(message: ComparisonHistogramProto): unknown {
    const obj: any = {};
    if (message.bars?.length) {
      obj.bars = message.bars.map((e) => HistogramBarLineProto.toJSON(e));
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<ComparisonHistogramProto>, I>>(base?: I): ComparisonHistogramProto {
    return ComparisonHistogramProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<ComparisonHistogramProto>, I>>(object: I): ComparisonHistogramProto {
    const message = createBaseComparisonHistogramProto();
    message.bars = object.bars?.map((e) => HistogramBarLineProto.fromPartial(e)) || [];
    return message;
  },
};

function createBaseComparisonLineProto(): ComparisonLineProto {
  return { baseline: undefined, scenarios: [] };
}

export const ComparisonLineProto = {
  encode(message: ComparisonLineProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.baseline !== undefined) {
      ComparisonHistogramProto.encode(message.baseline, writer.uint32(10).fork()).ldelim();
    }
    if (message.scenarios !== undefined && message.scenarios.length !== 0) {
      for (const v of message.scenarios) {
        ComparisonHistogramProto.encode(v!, writer.uint32(18).fork()).ldelim();
      }
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): ComparisonLineProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseComparisonLineProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.baseline = ComparisonHistogramProto.decode(reader, reader.uint32());
          continue;
        case 2:
          if (tag !== 18) {
            break;
          }

          message.scenarios!.push(ComparisonHistogramProto.decode(reader, reader.uint32()));
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): ComparisonLineProto {
    return {
      baseline: isSet(object.baseline) ? ComparisonHistogramProto.fromJSON(object.baseline) : undefined,
      scenarios: globalThis.Array.isArray(object?.scenarios)
        ? object.scenarios.map((e: any) => ComparisonHistogramProto.fromJSON(e))
        : [],
    };
  },

  toJSON(message: ComparisonLineProto): unknown {
    const obj: any = {};
    if (message.baseline !== undefined) {
      obj.baseline = ComparisonHistogramProto.toJSON(message.baseline);
    }
    if (message.scenarios?.length) {
      obj.scenarios = message.scenarios.map((e) => ComparisonHistogramProto.toJSON(e));
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<ComparisonLineProto>, I>>(base?: I): ComparisonLineProto {
    return ComparisonLineProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<ComparisonLineProto>, I>>(object: I): ComparisonLineProto {
    const message = createBaseComparisonLineProto();
    message.baseline = (object.baseline !== undefined && object.baseline !== null)
      ? ComparisonHistogramProto.fromPartial(object.baseline)
      : undefined;
    message.scenarios = object.scenarios?.map((e) => ComparisonHistogramProto.fromPartial(e)) || [];
    return message;
  },
};

function createBaseAtomProto(): AtomProto {
  return { value: undefined, labelledIcon: undefined, download: undefined, chip: undefined, conditional: undefined };
}

export const AtomProto = {
  encode(message: AtomProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.value !== undefined) {
      writer.uint32(10).string(message.value);
    }
    if (message.labelledIcon !== undefined) {
      LabelledIconProto.encode(message.labelledIcon, writer.uint32(18).fork()).ldelim();
    }
    if (message.download !== undefined) {
      writer.uint32(26).string(message.download);
    }
    if (message.chip !== undefined) {
      ChipProto.encode(message.chip, writer.uint32(34).fork()).ldelim();
    }
    if (message.conditional !== undefined) {
      ConditionalProto.encode(message.conditional, writer.uint32(42).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): AtomProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseAtomProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.value = reader.string();
          continue;
        case 2:
          if (tag !== 18) {
            break;
          }

          message.labelledIcon = LabelledIconProto.decode(reader, reader.uint32());
          continue;
        case 3:
          if (tag !== 26) {
            break;
          }

          message.download = reader.string();
          continue;
        case 4:
          if (tag !== 34) {
            break;
          }

          message.chip = ChipProto.decode(reader, reader.uint32());
          continue;
        case 5:
          if (tag !== 42) {
            break;
          }

          message.conditional = ConditionalProto.decode(reader, reader.uint32());
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): AtomProto {
    return {
      value: isSet(object.value) ? globalThis.String(object.value) : undefined,
      labelledIcon: isSet(object.labelledIcon) ? LabelledIconProto.fromJSON(object.labelledIcon) : undefined,
      download: isSet(object.download) ? globalThis.String(object.download) : undefined,
      chip: isSet(object.chip) ? ChipProto.fromJSON(object.chip) : undefined,
      conditional: isSet(object.conditional) ? ConditionalProto.fromJSON(object.conditional) : undefined,
    };
  },

  toJSON(message: AtomProto): unknown {
    const obj: any = {};
    if (message.value !== undefined) {
      obj.value = message.value;
    }
    if (message.labelledIcon !== undefined) {
      obj.labelledIcon = LabelledIconProto.toJSON(message.labelledIcon);
    }
    if (message.download !== undefined) {
      obj.download = message.download;
    }
    if (message.chip !== undefined) {
      obj.chip = ChipProto.toJSON(message.chip);
    }
    if (message.conditional !== undefined) {
      obj.conditional = ConditionalProto.toJSON(message.conditional);
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<AtomProto>, I>>(base?: I): AtomProto {
    return AtomProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<AtomProto>, I>>(object: I): AtomProto {
    const message = createBaseAtomProto();
    message.value = object.value ?? undefined;
    message.labelledIcon = (object.labelledIcon !== undefined && object.labelledIcon !== null)
      ? LabelledIconProto.fromPartial(object.labelledIcon)
      : undefined;
    message.download = object.download ?? undefined;
    message.chip = (object.chip !== undefined && object.chip !== null) ? ChipProto.fromPartial(object.chip) : undefined;
    message.conditional = (object.conditional !== undefined && object.conditional !== null)
      ? ConditionalProto.fromPartial(object.conditional)
      : undefined;
    return message;
  },
};

function createBaseLabelledIconProto(): LabelledIconProto {
  return { icon: "", label: "" };
}

export const LabelledIconProto = {
  encode(message: LabelledIconProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.icon !== undefined && message.icon !== "") {
      writer.uint32(10).string(message.icon);
    }
    if (message.label !== undefined && message.label !== "") {
      writer.uint32(18).string(message.label);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): LabelledIconProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseLabelledIconProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.icon = reader.string();
          continue;
        case 2:
          if (tag !== 18) {
            break;
          }

          message.label = reader.string();
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): LabelledIconProto {
    return {
      icon: isSet(object.icon) ? globalThis.String(object.icon) : "",
      label: isSet(object.label) ? globalThis.String(object.label) : "",
    };
  },

  toJSON(message: LabelledIconProto): unknown {
    const obj: any = {};
    if (message.icon !== undefined && message.icon !== "") {
      obj.icon = message.icon;
    }
    if (message.label !== undefined && message.label !== "") {
      obj.label = message.label;
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<LabelledIconProto>, I>>(base?: I): LabelledIconProto {
    return LabelledIconProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<LabelledIconProto>, I>>(object: I): LabelledIconProto {
    const message = createBaseLabelledIconProto();
    message.icon = object.icon ?? "";
    message.label = object.label ?? "";
    return message;
  },
};

function createBaseChipProto(): ChipProto {
  return { index: 0, labels: [] };
}

export const ChipProto = {
  encode(message: ChipProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.index !== undefined && message.index !== 0) {
      writer.uint32(8).int32(message.index);
    }
    if (message.labels !== undefined && message.labels.length !== 0) {
      for (const v of message.labels) {
        writer.uint32(18).string(v!);
      }
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): ChipProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseChipProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 8) {
            break;
          }

          message.index = reader.int32();
          continue;
        case 2:
          if (tag !== 18) {
            break;
          }

          message.labels!.push(reader.string());
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): ChipProto {
    return {
      index: isSet(object.index) ? globalThis.Number(object.index) : 0,
      labels: globalThis.Array.isArray(object?.labels) ? object.labels.map((e: any) => globalThis.String(e)) : [],
    };
  },

  toJSON(message: ChipProto): unknown {
    const obj: any = {};
    if (message.index !== undefined && message.index !== 0) {
      obj.index = Math.round(message.index);
    }
    if (message.labels?.length) {
      obj.labels = message.labels;
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<ChipProto>, I>>(base?: I): ChipProto {
    return ChipProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<ChipProto>, I>>(object: I): ChipProto {
    const message = createBaseChipProto();
    message.index = object.index ?? 0;
    message.labels = object.labels?.map((e) => e) || [];
    return message;
  },
};

function createBaseConditionProto(): ConditionProto {
  return { indices: [], values: [] };
}

export const ConditionProto = {
  encode(message: ConditionProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.indices !== undefined && message.indices.length !== 0) {
      writer.uint32(10).fork();
      for (const v of message.indices) {
        writer.int32(v);
      }
      writer.ldelim();
    }
    if (message.values !== undefined && message.values.length !== 0) {
      writer.uint32(18).fork();
      for (const v of message.values) {
        writer.int32(v);
      }
      writer.ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): ConditionProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseConditionProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag === 8) {
            message.indices!.push(reader.int32());

            continue;
          }

          if (tag === 10) {
            const end2 = reader.uint32() + reader.pos;
            while (reader.pos < end2) {
              message.indices!.push(reader.int32());
            }

            continue;
          }

          break;
        case 2:
          if (tag === 16) {
            message.values!.push(reader.int32());

            continue;
          }

          if (tag === 18) {
            const end2 = reader.uint32() + reader.pos;
            while (reader.pos < end2) {
              message.values!.push(reader.int32());
            }

            continue;
          }

          break;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): ConditionProto {
    return {
      indices: globalThis.Array.isArray(object?.indices) ? object.indices.map((e: any) => globalThis.Number(e)) : [],
      values: globalThis.Array.isArray(object?.values) ? object.values.map((e: any) => globalThis.Number(e)) : [],
    };
  },

  toJSON(message: ConditionProto): unknown {
    const obj: any = {};
    if (message.indices?.length) {
      obj.indices = message.indices.map((e) => Math.round(e));
    }
    if (message.values?.length) {
      obj.values = message.values.map((e) => Math.round(e));
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<ConditionProto>, I>>(base?: I): ConditionProto {
    return ConditionProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<ConditionProto>, I>>(object: I): ConditionProto {
    const message = createBaseConditionProto();
    message.indices = object.indices?.map((e) => e) || [];
    message.values = object.values?.map((e) => e) || [];
    return message;
  },
};

function createBaseConditionalProto(): ConditionalProto {
  return { conditions: [], atoms: [] };
}

export const ConditionalProto = {
  encode(message: ConditionalProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.conditions !== undefined && message.conditions.length !== 0) {
      for (const v of message.conditions) {
        ConditionProto.encode(v!, writer.uint32(10).fork()).ldelim();
      }
    }
    if (message.atoms !== undefined && message.atoms.length !== 0) {
      for (const v of message.atoms) {
        AtomProto.encode(v!, writer.uint32(18).fork()).ldelim();
      }
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): ConditionalProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseConditionalProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.conditions!.push(ConditionProto.decode(reader, reader.uint32()));
          continue;
        case 2:
          if (tag !== 18) {
            break;
          }

          message.atoms!.push(AtomProto.decode(reader, reader.uint32()));
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): ConditionalProto {
    return {
      conditions: globalThis.Array.isArray(object?.conditions)
        ? object.conditions.map((e: any) => ConditionProto.fromJSON(e))
        : [],
      atoms: globalThis.Array.isArray(object?.atoms) ? object.atoms.map((e: any) => AtomProto.fromJSON(e)) : [],
    };
  },

  toJSON(message: ConditionalProto): unknown {
    const obj: any = {};
    if (message.conditions?.length) {
      obj.conditions = message.conditions.map((e) => ConditionProto.toJSON(e));
    }
    if (message.atoms?.length) {
      obj.atoms = message.atoms.map((e) => AtomProto.toJSON(e));
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<ConditionalProto>, I>>(base?: I): ConditionalProto {
    return ConditionalProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<ConditionalProto>, I>>(object: I): ConditionalProto {
    const message = createBaseConditionalProto();
    message.conditions = object.conditions?.map((e) => ConditionProto.fromPartial(e)) || [];
    message.atoms = object.atoms?.map((e) => AtomProto.fromPartial(e)) || [];
    return message;
  },
};

function createBaseGeoJSONProto(): GeoJSONProto {
  return { condition: undefined, index: 0 };
}

export const GeoJSONProto = {
  encode(message: GeoJSONProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.condition !== undefined) {
      ConditionProto.encode(message.condition, writer.uint32(10).fork()).ldelim();
    }
    if (message.index !== undefined && message.index !== 0) {
      writer.uint32(16).int32(message.index);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): GeoJSONProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseGeoJSONProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.condition = ConditionProto.decode(reader, reader.uint32());
          continue;
        case 2:
          if (tag !== 16) {
            break;
          }

          message.index = reader.int32();
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): GeoJSONProto {
    return {
      condition: isSet(object.condition) ? ConditionProto.fromJSON(object.condition) : undefined,
      index: isSet(object.index) ? globalThis.Number(object.index) : 0,
    };
  },

  toJSON(message: GeoJSONProto): unknown {
    const obj: any = {};
    if (message.condition !== undefined) {
      obj.condition = ConditionProto.toJSON(message.condition);
    }
    if (message.index !== undefined && message.index !== 0) {
      obj.index = Math.round(message.index);
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<GeoJSONProto>, I>>(base?: I): GeoJSONProto {
    return GeoJSONProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<GeoJSONProto>, I>>(object: I): GeoJSONProto {
    const message = createBaseGeoJSONProto();
    message.condition = (object.condition !== undefined && object.condition !== null)
      ? ConditionProto.fromPartial(object.condition)
      : undefined;
    message.index = object.index ?? 0;
    return message;
  },
};

function createBaseFeatureIDsProto(): FeatureIDsProto {
  return { namespaces: [], ids: [] };
}

export const FeatureIDsProto = {
  encode(message: FeatureIDsProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.namespaces !== undefined && message.namespaces.length !== 0) {
      for (const v of message.namespaces) {
        writer.uint32(26).string(v!);
      }
    }
    if (message.ids !== undefined && message.ids.length !== 0) {
      for (const v of message.ids) {
        IDsProto.encode(v!, writer.uint32(34).fork()).ldelim();
      }
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): FeatureIDsProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseFeatureIDsProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 3:
          if (tag !== 26) {
            break;
          }

          message.namespaces!.push(reader.string());
          continue;
        case 4:
          if (tag !== 34) {
            break;
          }

          message.ids!.push(IDsProto.decode(reader, reader.uint32()));
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): FeatureIDsProto {
    return {
      namespaces: globalThis.Array.isArray(object?.namespaces)
        ? object.namespaces.map((e: any) => globalThis.String(e))
        : [],
      ids: globalThis.Array.isArray(object?.ids) ? object.ids.map((e: any) => IDsProto.fromJSON(e)) : [],
    };
  },

  toJSON(message: FeatureIDsProto): unknown {
    const obj: any = {};
    if (message.namespaces?.length) {
      obj.namespaces = message.namespaces;
    }
    if (message.ids?.length) {
      obj.ids = message.ids.map((e) => IDsProto.toJSON(e));
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<FeatureIDsProto>, I>>(base?: I): FeatureIDsProto {
    return FeatureIDsProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<FeatureIDsProto>, I>>(object: I): FeatureIDsProto {
    const message = createBaseFeatureIDsProto();
    message.namespaces = object.namespaces?.map((e) => e) || [];
    message.ids = object.ids?.map((e) => IDsProto.fromPartial(e)) || [];
    return message;
  },
};

function createBaseIDsProto(): IDsProto {
  return { ids: [] };
}

export const IDsProto = {
  encode(message: IDsProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.ids !== undefined && message.ids.length !== 0) {
      writer.uint32(10).fork();
      for (const v of message.ids) {
        writer.uint64(v);
      }
      writer.ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): IDsProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseIDsProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag === 8) {
            message.ids!.push(longToNumber(reader.uint64() as Long));

            continue;
          }

          if (tag === 10) {
            const end2 = reader.uint32() + reader.pos;
            while (reader.pos < end2) {
              message.ids!.push(longToNumber(reader.uint64() as Long));
            }

            continue;
          }

          break;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): IDsProto {
    return { ids: globalThis.Array.isArray(object?.ids) ? object.ids.map((e: any) => globalThis.Number(e)) : [] };
  },

  toJSON(message: IDsProto): unknown {
    const obj: any = {};
    if (message.ids?.length) {
      obj.ids = message.ids.map((e) => Math.round(e));
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<IDsProto>, I>>(base?: I): IDsProto {
    return IDsProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<IDsProto>, I>>(object: I): IDsProto {
    const message = createBaseIDsProto();
    message.ids = object.ids?.map((e) => e) || [];
    return message;
  },
};

function createBaseComparisonRequestProto(): ComparisonRequestProto {
  return { analysis: undefined, baseline: undefined, scenarios: [] };
}

export const ComparisonRequestProto = {
  encode(message: ComparisonRequestProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.analysis !== undefined) {
      FeatureIDProto.encode(message.analysis, writer.uint32(10).fork()).ldelim();
    }
    if (message.baseline !== undefined) {
      FeatureIDProto.encode(message.baseline, writer.uint32(18).fork()).ldelim();
    }
    if (message.scenarios !== undefined && message.scenarios.length !== 0) {
      for (const v of message.scenarios) {
        FeatureIDProto.encode(v!, writer.uint32(26).fork()).ldelim();
      }
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): ComparisonRequestProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseComparisonRequestProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.analysis = FeatureIDProto.decode(reader, reader.uint32());
          continue;
        case 2:
          if (tag !== 18) {
            break;
          }

          message.baseline = FeatureIDProto.decode(reader, reader.uint32());
          continue;
        case 3:
          if (tag !== 26) {
            break;
          }

          message.scenarios!.push(FeatureIDProto.decode(reader, reader.uint32()));
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): ComparisonRequestProto {
    return {
      analysis: isSet(object.analysis) ? FeatureIDProto.fromJSON(object.analysis) : undefined,
      baseline: isSet(object.baseline) ? FeatureIDProto.fromJSON(object.baseline) : undefined,
      scenarios: globalThis.Array.isArray(object?.scenarios)
        ? object.scenarios.map((e: any) => FeatureIDProto.fromJSON(e))
        : [],
    };
  },

  toJSON(message: ComparisonRequestProto): unknown {
    const obj: any = {};
    if (message.analysis !== undefined) {
      obj.analysis = FeatureIDProto.toJSON(message.analysis);
    }
    if (message.baseline !== undefined) {
      obj.baseline = FeatureIDProto.toJSON(message.baseline);
    }
    if (message.scenarios?.length) {
      obj.scenarios = message.scenarios.map((e) => FeatureIDProto.toJSON(e));
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<ComparisonRequestProto>, I>>(base?: I): ComparisonRequestProto {
    return ComparisonRequestProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<ComparisonRequestProto>, I>>(object: I): ComparisonRequestProto {
    const message = createBaseComparisonRequestProto();
    message.analysis = (object.analysis !== undefined && object.analysis !== null)
      ? FeatureIDProto.fromPartial(object.analysis)
      : undefined;
    message.baseline = (object.baseline !== undefined && object.baseline !== null)
      ? FeatureIDProto.fromPartial(object.baseline)
      : undefined;
    message.scenarios = object.scenarios?.map((e) => FeatureIDProto.fromPartial(e)) || [];
    return message;
  },
};

type Builtin = Date | Function | Uint8Array | string | number | boolean | undefined;

export type DeepPartial<T> = T extends Builtin ? T
  : T extends globalThis.Array<infer U> ? globalThis.Array<DeepPartial<U>>
  : T extends ReadonlyArray<infer U> ? ReadonlyArray<DeepPartial<U>>
  : T extends {} ? { [K in keyof T]?: DeepPartial<T[K]> }
  : Partial<T>;

type KeysOfUnion<T> = T extends T ? keyof T : never;
export type Exact<P, I extends P> = P extends Builtin ? P
  : P & { [K in keyof P]: Exact<P[K], I[K]> } & { [K in Exclude<keyof I, KeysOfUnion<P>>]: never };

function longToNumber(long: Long): number {
  if (long.gt(globalThis.Number.MAX_SAFE_INTEGER)) {
    throw new globalThis.Error("Value is larger than Number.MAX_SAFE_INTEGER");
  }
  return long.toNumber();
}

if (_m0.util.Long !== Long) {
  _m0.util.Long = Long as any;
  _m0.configure();
}

function isSet(value: any): boolean {
  return value !== null && value !== undefined;
}
