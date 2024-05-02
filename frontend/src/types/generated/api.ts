/* eslint-disable */
import Long from "long";
import _m0 from "protobufjs/minimal";
import { MultiPolygonProto, PointProto, PolylineProto } from "./geometry";

export const protobufPackage = "api";

export enum FeatureType {
  FeatureTypeInvalid = 0,
  FeatureTypePoint = 1,
  FeatureTypePath = 2,
  FeatureTypeArea = 3,
  FeatureTypeRelation = 4,
  FeatureTypeCollection = 5,
  FeatureTypeExpression = 6,
  UNRECOGNIZED = -1,
}

export function featureTypeFromJSON(object: any): FeatureType {
  switch (object) {
    case 0:
    case "FeatureTypeInvalid":
      return FeatureType.FeatureTypeInvalid;
    case 1:
    case "FeatureTypePoint":
      return FeatureType.FeatureTypePoint;
    case 2:
    case "FeatureTypePath":
      return FeatureType.FeatureTypePath;
    case 3:
    case "FeatureTypeArea":
      return FeatureType.FeatureTypeArea;
    case 4:
    case "FeatureTypeRelation":
      return FeatureType.FeatureTypeRelation;
    case 5:
    case "FeatureTypeCollection":
      return FeatureType.FeatureTypeCollection;
    case 6:
    case "FeatureTypeExpression":
      return FeatureType.FeatureTypeExpression;
    case -1:
    case "UNRECOGNIZED":
    default:
      return FeatureType.UNRECOGNIZED;
  }
}

export function featureTypeToJSON(object: FeatureType): string {
  switch (object) {
    case FeatureType.FeatureTypeInvalid:
      return "FeatureTypeInvalid";
    case FeatureType.FeatureTypePoint:
      return "FeatureTypePoint";
    case FeatureType.FeatureTypePath:
      return "FeatureTypePath";
    case FeatureType.FeatureTypeArea:
      return "FeatureTypeArea";
    case FeatureType.FeatureTypeRelation:
      return "FeatureTypeRelation";
    case FeatureType.FeatureTypeCollection:
      return "FeatureTypeCollection";
    case FeatureType.FeatureTypeExpression:
      return "FeatureTypeExpression";
    case FeatureType.UNRECOGNIZED:
    default:
      return "UNRECOGNIZED";
  }
}

export interface TagProto {
  key: string;
  value: string;
}

export interface FeatureIDProto {
  type: FeatureType;
  namespace: string;
  value: number;
}

export interface PointFeatureProto {
  id: FeatureIDProto | undefined;
  tags: TagProto[];
  point: PointProto | undefined;
}

export interface PathFeatureProto {
  id: FeatureIDProto | undefined;
  tags: TagProto[];
  features: PointFeatureProto[];
  lengthMeters: number;
}

export interface PathFeaturesProto {
  paths: PathFeatureProto[];
}

export interface AreaFeatureProto {
  id: FeatureIDProto | undefined;
  tags: TagProto[];
  features: PathFeaturesProto[];
}

export interface RelationMemberProto {
  id: FeatureIDProto | undefined;
  role: string;
}

export interface RelationFeatureProto {
  id: FeatureIDProto | undefined;
  tags: TagProto[];
  members: RelationMemberProto[];
}

export interface CollectionFeatureProto {
  id: FeatureIDProto | undefined;
  tags: TagProto[];
  collection: CollectionProto | undefined;
}

export interface ExpressionFeatureProto {
  id: FeatureIDProto | undefined;
  tags: TagProto[];
  expression: NodeProto | undefined;
}

export interface FeatureProto {
  point?: PointFeatureProto | undefined;
  path?: PathFeatureProto | undefined;
  area?: AreaFeatureProto | undefined;
  relation?: RelationFeatureProto | undefined;
  collection?: CollectionFeatureProto | undefined;
  expression?: ExpressionFeatureProto | undefined;
}

export interface CollectionProto {
  keys: LiteralNodeProto[];
  values: LiteralNodeProto[];
}

export interface PairProto {
  first: LiteralNodeProto | undefined;
  second: LiteralNodeProto | undefined;
}

export interface ModifiedFeaturesProto {
  ids: FeatureIDProto[];
}

export interface AppliedChangeProto {
  original: FeatureIDProto[];
  modified: FeatureIDProto[];
}

export interface NodeProto {
  symbol?: string | undefined;
  literal?: LiteralNodeProto | undefined;
  call?: CallNodeProto | undefined;
  lambda?: LambdaNodeProto | undefined;
  name: string;
  begin: number;
  end: number;
}

export interface LiteralNodeProto {
  nilValue?: boolean | undefined;
  boolValue?: boolean | undefined;
  stringValue?: string | undefined;
  intValue?: number | undefined;
  floatValue?: number | undefined;
  collectionValue?: CollectionProto | undefined;
  pairValue?: PairProto | undefined;
  featureValue?: FeatureProto | undefined;
  queryValue?: QueryProto | undefined;
  featureIDValue?: FeatureIDProto | undefined;
  pointValue?: PointProto | undefined;
  pathValue?: PolylineProto | undefined;
  areaValue?: MultiPolygonProto | undefined;
  appliedChangeValue?:
    | AppliedChangeProto
    | undefined;
  /** gzipped */
  geoJSONValue?: Uint8Array | undefined;
  tagValue?: TagProto | undefined;
  routeValue?: RouteProto | undefined;
}

export interface CallNodeProto {
  function: NodeProto | undefined;
  args: NodeProto[];
  pipelined: boolean;
}

export interface LambdaNodeProto {
  args: string[];
  node: NodeProto | undefined;
}

export interface KeyQueryProto {
  key: string;
}

export interface KeyValueQueryProto {
  key: string;
  value: string;
}

export interface TypedQueryProto {
  type: FeatureType;
  query: QueryProto | undefined;
}

export interface QueriesProto {
  queries: QueryProto[];
}

export interface AllQueryProto {
}

export interface EmptyQueryProto {
}

export interface CapProto {
  center: PointProto | undefined;
  radiusMeters: number;
}

export interface S2CellIDsProto {
  s2CellIDs: number[];
}

export interface QueryProto {
  all?: AllQueryProto | undefined;
  empty?: EmptyQueryProto | undefined;
  keyed?: string | undefined;
  tagged?: TagProto | undefined;
  typed?: TypedQueryProto | undefined;
  intersection?: QueriesProto | undefined;
  union?: QueriesProto | undefined;
  intersectsCap?: CapProto | undefined;
  intersectsFeature?: FeatureIDProto | undefined;
  intersectsPoint?: PointProto | undefined;
  intersectsPolyline?: PolylineProto | undefined;
  intersectsMultiPolygon?: MultiPolygonProto | undefined;
  intersectsCells?: S2CellIDsProto | undefined;
  mightIntersect?: S2CellIDsProto | undefined;
}

export interface StepProto {
  destination: FeatureIDProto | undefined;
  via: FeatureIDProto | undefined;
  cost: number;
}

export interface RouteProto {
  origin: FeatureIDProto | undefined;
  steps: StepProto[];
}

export interface FindFeatureByIDRequestProto {
  id: FeatureIDProto | undefined;
}

export interface FindFeatureByIDResponseProto {
  feature: FeatureProto | undefined;
}

export interface FindFeaturesRequestProto {
  query: QueryProto | undefined;
}

export interface FindFeaturesResponseProto {
  features: FeatureProto[];
}

export interface ModifyTagsRequestProto {
  id: FeatureIDProto | undefined;
  tags: TagProto[];
}

export interface ModifyTagsBatchRequestProto {
  requests: ModifyTagsRequestProto[];
}

export interface ModifyTagsBatchResponseProto {
}

export interface EvaluateRequestProto {
  request: NodeProto | undefined;
  version: string;
  root: FeatureIDProto | undefined;
}

export interface EvaluateResponseProto {
  result: NodeProto | undefined;
}

function createBaseTagProto(): TagProto {
  return { key: "", value: "" };
}

export const TagProto = {
  encode(message: TagProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.key !== "") {
      writer.uint32(10).string(message.key);
    }
    if (message.value !== "") {
      writer.uint32(18).string(message.value);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): TagProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseTagProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.key = reader.string();
          continue;
        case 2:
          if (tag !== 18) {
            break;
          }

          message.value = reader.string();
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): TagProto {
    return {
      key: isSet(object.key) ? globalThis.String(object.key) : "",
      value: isSet(object.value) ? globalThis.String(object.value) : "",
    };
  },

  toJSON(message: TagProto): unknown {
    const obj: any = {};
    if (message.key !== "") {
      obj.key = message.key;
    }
    if (message.value !== "") {
      obj.value = message.value;
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<TagProto>, I>>(base?: I): TagProto {
    return TagProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<TagProto>, I>>(object: I): TagProto {
    const message = createBaseTagProto();
    message.key = object.key ?? "";
    message.value = object.value ?? "";
    return message;
  },
};

function createBaseFeatureIDProto(): FeatureIDProto {
  return { type: 0, namespace: "", value: 0 };
}

export const FeatureIDProto = {
  encode(message: FeatureIDProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.type !== 0) {
      writer.uint32(8).int32(message.type);
    }
    if (message.namespace !== "") {
      writer.uint32(18).string(message.namespace);
    }
    if (message.value !== 0) {
      writer.uint32(24).uint64(message.value);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): FeatureIDProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseFeatureIDProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 8) {
            break;
          }

          message.type = reader.int32() as any;
          continue;
        case 2:
          if (tag !== 18) {
            break;
          }

          message.namespace = reader.string();
          continue;
        case 3:
          if (tag !== 24) {
            break;
          }

          message.value = longToNumber(reader.uint64() as Long);
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): FeatureIDProto {
    return {
      type: isSet(object.type) ? featureTypeFromJSON(object.type) : 0,
      namespace: isSet(object.namespace) ? globalThis.String(object.namespace) : "",
      value: isSet(object.value) ? globalThis.Number(object.value) : 0,
    };
  },

  toJSON(message: FeatureIDProto): unknown {
    const obj: any = {};
    if (message.type !== 0) {
      obj.type = featureTypeToJSON(message.type);
    }
    if (message.namespace !== "") {
      obj.namespace = message.namespace;
    }
    if (message.value !== 0) {
      obj.value = Math.round(message.value);
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<FeatureIDProto>, I>>(base?: I): FeatureIDProto {
    return FeatureIDProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<FeatureIDProto>, I>>(object: I): FeatureIDProto {
    const message = createBaseFeatureIDProto();
    message.type = object.type ?? 0;
    message.namespace = object.namespace ?? "";
    message.value = object.value ?? 0;
    return message;
  },
};

function createBasePointFeatureProto(): PointFeatureProto {
  return { id: undefined, tags: [], point: undefined };
}

export const PointFeatureProto = {
  encode(message: PointFeatureProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.id !== undefined) {
      FeatureIDProto.encode(message.id, writer.uint32(10).fork()).ldelim();
    }
    for (const v of message.tags) {
      TagProto.encode(v!, writer.uint32(18).fork()).ldelim();
    }
    if (message.point !== undefined) {
      PointProto.encode(message.point, writer.uint32(26).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): PointFeatureProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBasePointFeatureProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.id = FeatureIDProto.decode(reader, reader.uint32());
          continue;
        case 2:
          if (tag !== 18) {
            break;
          }

          message.tags.push(TagProto.decode(reader, reader.uint32()));
          continue;
        case 3:
          if (tag !== 26) {
            break;
          }

          message.point = PointProto.decode(reader, reader.uint32());
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): PointFeatureProto {
    return {
      id: isSet(object.id) ? FeatureIDProto.fromJSON(object.id) : undefined,
      tags: globalThis.Array.isArray(object?.tags) ? object.tags.map((e: any) => TagProto.fromJSON(e)) : [],
      point: isSet(object.point) ? PointProto.fromJSON(object.point) : undefined,
    };
  },

  toJSON(message: PointFeatureProto): unknown {
    const obj: any = {};
    if (message.id !== undefined) {
      obj.id = FeatureIDProto.toJSON(message.id);
    }
    if (message.tags?.length) {
      obj.tags = message.tags.map((e) => TagProto.toJSON(e));
    }
    if (message.point !== undefined) {
      obj.point = PointProto.toJSON(message.point);
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<PointFeatureProto>, I>>(base?: I): PointFeatureProto {
    return PointFeatureProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<PointFeatureProto>, I>>(object: I): PointFeatureProto {
    const message = createBasePointFeatureProto();
    message.id = (object.id !== undefined && object.id !== null) ? FeatureIDProto.fromPartial(object.id) : undefined;
    message.tags = object.tags?.map((e) => TagProto.fromPartial(e)) || [];
    message.point = (object.point !== undefined && object.point !== null)
      ? PointProto.fromPartial(object.point)
      : undefined;
    return message;
  },
};

function createBasePathFeatureProto(): PathFeatureProto {
  return { id: undefined, tags: [], features: [], lengthMeters: 0 };
}

export const PathFeatureProto = {
  encode(message: PathFeatureProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.id !== undefined) {
      FeatureIDProto.encode(message.id, writer.uint32(10).fork()).ldelim();
    }
    for (const v of message.tags) {
      TagProto.encode(v!, writer.uint32(18).fork()).ldelim();
    }
    for (const v of message.features) {
      PointFeatureProto.encode(v!, writer.uint32(26).fork()).ldelim();
    }
    if (message.lengthMeters !== 0) {
      writer.uint32(33).double(message.lengthMeters);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): PathFeatureProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBasePathFeatureProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.id = FeatureIDProto.decode(reader, reader.uint32());
          continue;
        case 2:
          if (tag !== 18) {
            break;
          }

          message.tags.push(TagProto.decode(reader, reader.uint32()));
          continue;
        case 3:
          if (tag !== 26) {
            break;
          }

          message.features.push(PointFeatureProto.decode(reader, reader.uint32()));
          continue;
        case 4:
          if (tag !== 33) {
            break;
          }

          message.lengthMeters = reader.double();
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): PathFeatureProto {
    return {
      id: isSet(object.id) ? FeatureIDProto.fromJSON(object.id) : undefined,
      tags: globalThis.Array.isArray(object?.tags) ? object.tags.map((e: any) => TagProto.fromJSON(e)) : [],
      features: globalThis.Array.isArray(object?.features)
        ? object.features.map((e: any) => PointFeatureProto.fromJSON(e))
        : [],
      lengthMeters: isSet(object.lengthMeters) ? globalThis.Number(object.lengthMeters) : 0,
    };
  },

  toJSON(message: PathFeatureProto): unknown {
    const obj: any = {};
    if (message.id !== undefined) {
      obj.id = FeatureIDProto.toJSON(message.id);
    }
    if (message.tags?.length) {
      obj.tags = message.tags.map((e) => TagProto.toJSON(e));
    }
    if (message.features?.length) {
      obj.features = message.features.map((e) => PointFeatureProto.toJSON(e));
    }
    if (message.lengthMeters !== 0) {
      obj.lengthMeters = message.lengthMeters;
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<PathFeatureProto>, I>>(base?: I): PathFeatureProto {
    return PathFeatureProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<PathFeatureProto>, I>>(object: I): PathFeatureProto {
    const message = createBasePathFeatureProto();
    message.id = (object.id !== undefined && object.id !== null) ? FeatureIDProto.fromPartial(object.id) : undefined;
    message.tags = object.tags?.map((e) => TagProto.fromPartial(e)) || [];
    message.features = object.features?.map((e) => PointFeatureProto.fromPartial(e)) || [];
    message.lengthMeters = object.lengthMeters ?? 0;
    return message;
  },
};

function createBasePathFeaturesProto(): PathFeaturesProto {
  return { paths: [] };
}

export const PathFeaturesProto = {
  encode(message: PathFeaturesProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    for (const v of message.paths) {
      PathFeatureProto.encode(v!, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): PathFeaturesProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBasePathFeaturesProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.paths.push(PathFeatureProto.decode(reader, reader.uint32()));
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): PathFeaturesProto {
    return {
      paths: globalThis.Array.isArray(object?.paths) ? object.paths.map((e: any) => PathFeatureProto.fromJSON(e)) : [],
    };
  },

  toJSON(message: PathFeaturesProto): unknown {
    const obj: any = {};
    if (message.paths?.length) {
      obj.paths = message.paths.map((e) => PathFeatureProto.toJSON(e));
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<PathFeaturesProto>, I>>(base?: I): PathFeaturesProto {
    return PathFeaturesProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<PathFeaturesProto>, I>>(object: I): PathFeaturesProto {
    const message = createBasePathFeaturesProto();
    message.paths = object.paths?.map((e) => PathFeatureProto.fromPartial(e)) || [];
    return message;
  },
};

function createBaseAreaFeatureProto(): AreaFeatureProto {
  return { id: undefined, tags: [], features: [] };
}

export const AreaFeatureProto = {
  encode(message: AreaFeatureProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.id !== undefined) {
      FeatureIDProto.encode(message.id, writer.uint32(10).fork()).ldelim();
    }
    for (const v of message.tags) {
      TagProto.encode(v!, writer.uint32(18).fork()).ldelim();
    }
    for (const v of message.features) {
      PathFeaturesProto.encode(v!, writer.uint32(26).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): AreaFeatureProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseAreaFeatureProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.id = FeatureIDProto.decode(reader, reader.uint32());
          continue;
        case 2:
          if (tag !== 18) {
            break;
          }

          message.tags.push(TagProto.decode(reader, reader.uint32()));
          continue;
        case 3:
          if (tag !== 26) {
            break;
          }

          message.features.push(PathFeaturesProto.decode(reader, reader.uint32()));
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): AreaFeatureProto {
    return {
      id: isSet(object.id) ? FeatureIDProto.fromJSON(object.id) : undefined,
      tags: globalThis.Array.isArray(object?.tags) ? object.tags.map((e: any) => TagProto.fromJSON(e)) : [],
      features: globalThis.Array.isArray(object?.features)
        ? object.features.map((e: any) => PathFeaturesProto.fromJSON(e))
        : [],
    };
  },

  toJSON(message: AreaFeatureProto): unknown {
    const obj: any = {};
    if (message.id !== undefined) {
      obj.id = FeatureIDProto.toJSON(message.id);
    }
    if (message.tags?.length) {
      obj.tags = message.tags.map((e) => TagProto.toJSON(e));
    }
    if (message.features?.length) {
      obj.features = message.features.map((e) => PathFeaturesProto.toJSON(e));
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<AreaFeatureProto>, I>>(base?: I): AreaFeatureProto {
    return AreaFeatureProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<AreaFeatureProto>, I>>(object: I): AreaFeatureProto {
    const message = createBaseAreaFeatureProto();
    message.id = (object.id !== undefined && object.id !== null) ? FeatureIDProto.fromPartial(object.id) : undefined;
    message.tags = object.tags?.map((e) => TagProto.fromPartial(e)) || [];
    message.features = object.features?.map((e) => PathFeaturesProto.fromPartial(e)) || [];
    return message;
  },
};

function createBaseRelationMemberProto(): RelationMemberProto {
  return { id: undefined, role: "" };
}

export const RelationMemberProto = {
  encode(message: RelationMemberProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.id !== undefined) {
      FeatureIDProto.encode(message.id, writer.uint32(10).fork()).ldelim();
    }
    if (message.role !== "") {
      writer.uint32(26).string(message.role);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): RelationMemberProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseRelationMemberProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.id = FeatureIDProto.decode(reader, reader.uint32());
          continue;
        case 3:
          if (tag !== 26) {
            break;
          }

          message.role = reader.string();
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): RelationMemberProto {
    return {
      id: isSet(object.id) ? FeatureIDProto.fromJSON(object.id) : undefined,
      role: isSet(object.role) ? globalThis.String(object.role) : "",
    };
  },

  toJSON(message: RelationMemberProto): unknown {
    const obj: any = {};
    if (message.id !== undefined) {
      obj.id = FeatureIDProto.toJSON(message.id);
    }
    if (message.role !== "") {
      obj.role = message.role;
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<RelationMemberProto>, I>>(base?: I): RelationMemberProto {
    return RelationMemberProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<RelationMemberProto>, I>>(object: I): RelationMemberProto {
    const message = createBaseRelationMemberProto();
    message.id = (object.id !== undefined && object.id !== null) ? FeatureIDProto.fromPartial(object.id) : undefined;
    message.role = object.role ?? "";
    return message;
  },
};

function createBaseRelationFeatureProto(): RelationFeatureProto {
  return { id: undefined, tags: [], members: [] };
}

export const RelationFeatureProto = {
  encode(message: RelationFeatureProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.id !== undefined) {
      FeatureIDProto.encode(message.id, writer.uint32(10).fork()).ldelim();
    }
    for (const v of message.tags) {
      TagProto.encode(v!, writer.uint32(18).fork()).ldelim();
    }
    for (const v of message.members) {
      RelationMemberProto.encode(v!, writer.uint32(26).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): RelationFeatureProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseRelationFeatureProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.id = FeatureIDProto.decode(reader, reader.uint32());
          continue;
        case 2:
          if (tag !== 18) {
            break;
          }

          message.tags.push(TagProto.decode(reader, reader.uint32()));
          continue;
        case 3:
          if (tag !== 26) {
            break;
          }

          message.members.push(RelationMemberProto.decode(reader, reader.uint32()));
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): RelationFeatureProto {
    return {
      id: isSet(object.id) ? FeatureIDProto.fromJSON(object.id) : undefined,
      tags: globalThis.Array.isArray(object?.tags) ? object.tags.map((e: any) => TagProto.fromJSON(e)) : [],
      members: globalThis.Array.isArray(object?.members)
        ? object.members.map((e: any) => RelationMemberProto.fromJSON(e))
        : [],
    };
  },

  toJSON(message: RelationFeatureProto): unknown {
    const obj: any = {};
    if (message.id !== undefined) {
      obj.id = FeatureIDProto.toJSON(message.id);
    }
    if (message.tags?.length) {
      obj.tags = message.tags.map((e) => TagProto.toJSON(e));
    }
    if (message.members?.length) {
      obj.members = message.members.map((e) => RelationMemberProto.toJSON(e));
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<RelationFeatureProto>, I>>(base?: I): RelationFeatureProto {
    return RelationFeatureProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<RelationFeatureProto>, I>>(object: I): RelationFeatureProto {
    const message = createBaseRelationFeatureProto();
    message.id = (object.id !== undefined && object.id !== null) ? FeatureIDProto.fromPartial(object.id) : undefined;
    message.tags = object.tags?.map((e) => TagProto.fromPartial(e)) || [];
    message.members = object.members?.map((e) => RelationMemberProto.fromPartial(e)) || [];
    return message;
  },
};

function createBaseCollectionFeatureProto(): CollectionFeatureProto {
  return { id: undefined, tags: [], collection: undefined };
}

export const CollectionFeatureProto = {
  encode(message: CollectionFeatureProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.id !== undefined) {
      FeatureIDProto.encode(message.id, writer.uint32(10).fork()).ldelim();
    }
    for (const v of message.tags) {
      TagProto.encode(v!, writer.uint32(18).fork()).ldelim();
    }
    if (message.collection !== undefined) {
      CollectionProto.encode(message.collection, writer.uint32(26).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): CollectionFeatureProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseCollectionFeatureProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.id = FeatureIDProto.decode(reader, reader.uint32());
          continue;
        case 2:
          if (tag !== 18) {
            break;
          }

          message.tags.push(TagProto.decode(reader, reader.uint32()));
          continue;
        case 3:
          if (tag !== 26) {
            break;
          }

          message.collection = CollectionProto.decode(reader, reader.uint32());
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): CollectionFeatureProto {
    return {
      id: isSet(object.id) ? FeatureIDProto.fromJSON(object.id) : undefined,
      tags: globalThis.Array.isArray(object?.tags) ? object.tags.map((e: any) => TagProto.fromJSON(e)) : [],
      collection: isSet(object.collection) ? CollectionProto.fromJSON(object.collection) : undefined,
    };
  },

  toJSON(message: CollectionFeatureProto): unknown {
    const obj: any = {};
    if (message.id !== undefined) {
      obj.id = FeatureIDProto.toJSON(message.id);
    }
    if (message.tags?.length) {
      obj.tags = message.tags.map((e) => TagProto.toJSON(e));
    }
    if (message.collection !== undefined) {
      obj.collection = CollectionProto.toJSON(message.collection);
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<CollectionFeatureProto>, I>>(base?: I): CollectionFeatureProto {
    return CollectionFeatureProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<CollectionFeatureProto>, I>>(object: I): CollectionFeatureProto {
    const message = createBaseCollectionFeatureProto();
    message.id = (object.id !== undefined && object.id !== null) ? FeatureIDProto.fromPartial(object.id) : undefined;
    message.tags = object.tags?.map((e) => TagProto.fromPartial(e)) || [];
    message.collection = (object.collection !== undefined && object.collection !== null)
      ? CollectionProto.fromPartial(object.collection)
      : undefined;
    return message;
  },
};

function createBaseExpressionFeatureProto(): ExpressionFeatureProto {
  return { id: undefined, tags: [], expression: undefined };
}

export const ExpressionFeatureProto = {
  encode(message: ExpressionFeatureProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.id !== undefined) {
      FeatureIDProto.encode(message.id, writer.uint32(10).fork()).ldelim();
    }
    for (const v of message.tags) {
      TagProto.encode(v!, writer.uint32(18).fork()).ldelim();
    }
    if (message.expression !== undefined) {
      NodeProto.encode(message.expression, writer.uint32(26).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): ExpressionFeatureProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseExpressionFeatureProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.id = FeatureIDProto.decode(reader, reader.uint32());
          continue;
        case 2:
          if (tag !== 18) {
            break;
          }

          message.tags.push(TagProto.decode(reader, reader.uint32()));
          continue;
        case 3:
          if (tag !== 26) {
            break;
          }

          message.expression = NodeProto.decode(reader, reader.uint32());
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): ExpressionFeatureProto {
    return {
      id: isSet(object.id) ? FeatureIDProto.fromJSON(object.id) : undefined,
      tags: globalThis.Array.isArray(object?.tags) ? object.tags.map((e: any) => TagProto.fromJSON(e)) : [],
      expression: isSet(object.expression) ? NodeProto.fromJSON(object.expression) : undefined,
    };
  },

  toJSON(message: ExpressionFeatureProto): unknown {
    const obj: any = {};
    if (message.id !== undefined) {
      obj.id = FeatureIDProto.toJSON(message.id);
    }
    if (message.tags?.length) {
      obj.tags = message.tags.map((e) => TagProto.toJSON(e));
    }
    if (message.expression !== undefined) {
      obj.expression = NodeProto.toJSON(message.expression);
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<ExpressionFeatureProto>, I>>(base?: I): ExpressionFeatureProto {
    return ExpressionFeatureProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<ExpressionFeatureProto>, I>>(object: I): ExpressionFeatureProto {
    const message = createBaseExpressionFeatureProto();
    message.id = (object.id !== undefined && object.id !== null) ? FeatureIDProto.fromPartial(object.id) : undefined;
    message.tags = object.tags?.map((e) => TagProto.fromPartial(e)) || [];
    message.expression = (object.expression !== undefined && object.expression !== null)
      ? NodeProto.fromPartial(object.expression)
      : undefined;
    return message;
  },
};

function createBaseFeatureProto(): FeatureProto {
  return {
    point: undefined,
    path: undefined,
    area: undefined,
    relation: undefined,
    collection: undefined,
    expression: undefined,
  };
}

export const FeatureProto = {
  encode(message: FeatureProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.point !== undefined) {
      PointFeatureProto.encode(message.point, writer.uint32(10).fork()).ldelim();
    }
    if (message.path !== undefined) {
      PathFeatureProto.encode(message.path, writer.uint32(18).fork()).ldelim();
    }
    if (message.area !== undefined) {
      AreaFeatureProto.encode(message.area, writer.uint32(26).fork()).ldelim();
    }
    if (message.relation !== undefined) {
      RelationFeatureProto.encode(message.relation, writer.uint32(34).fork()).ldelim();
    }
    if (message.collection !== undefined) {
      CollectionFeatureProto.encode(message.collection, writer.uint32(42).fork()).ldelim();
    }
    if (message.expression !== undefined) {
      ExpressionFeatureProto.encode(message.expression, writer.uint32(50).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): FeatureProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseFeatureProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.point = PointFeatureProto.decode(reader, reader.uint32());
          continue;
        case 2:
          if (tag !== 18) {
            break;
          }

          message.path = PathFeatureProto.decode(reader, reader.uint32());
          continue;
        case 3:
          if (tag !== 26) {
            break;
          }

          message.area = AreaFeatureProto.decode(reader, reader.uint32());
          continue;
        case 4:
          if (tag !== 34) {
            break;
          }

          message.relation = RelationFeatureProto.decode(reader, reader.uint32());
          continue;
        case 5:
          if (tag !== 42) {
            break;
          }

          message.collection = CollectionFeatureProto.decode(reader, reader.uint32());
          continue;
        case 6:
          if (tag !== 50) {
            break;
          }

          message.expression = ExpressionFeatureProto.decode(reader, reader.uint32());
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): FeatureProto {
    return {
      point: isSet(object.point) ? PointFeatureProto.fromJSON(object.point) : undefined,
      path: isSet(object.path) ? PathFeatureProto.fromJSON(object.path) : undefined,
      area: isSet(object.area) ? AreaFeatureProto.fromJSON(object.area) : undefined,
      relation: isSet(object.relation) ? RelationFeatureProto.fromJSON(object.relation) : undefined,
      collection: isSet(object.collection) ? CollectionFeatureProto.fromJSON(object.collection) : undefined,
      expression: isSet(object.expression) ? ExpressionFeatureProto.fromJSON(object.expression) : undefined,
    };
  },

  toJSON(message: FeatureProto): unknown {
    const obj: any = {};
    if (message.point !== undefined) {
      obj.point = PointFeatureProto.toJSON(message.point);
    }
    if (message.path !== undefined) {
      obj.path = PathFeatureProto.toJSON(message.path);
    }
    if (message.area !== undefined) {
      obj.area = AreaFeatureProto.toJSON(message.area);
    }
    if (message.relation !== undefined) {
      obj.relation = RelationFeatureProto.toJSON(message.relation);
    }
    if (message.collection !== undefined) {
      obj.collection = CollectionFeatureProto.toJSON(message.collection);
    }
    if (message.expression !== undefined) {
      obj.expression = ExpressionFeatureProto.toJSON(message.expression);
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<FeatureProto>, I>>(base?: I): FeatureProto {
    return FeatureProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<FeatureProto>, I>>(object: I): FeatureProto {
    const message = createBaseFeatureProto();
    message.point = (object.point !== undefined && object.point !== null)
      ? PointFeatureProto.fromPartial(object.point)
      : undefined;
    message.path = (object.path !== undefined && object.path !== null)
      ? PathFeatureProto.fromPartial(object.path)
      : undefined;
    message.area = (object.area !== undefined && object.area !== null)
      ? AreaFeatureProto.fromPartial(object.area)
      : undefined;
    message.relation = (object.relation !== undefined && object.relation !== null)
      ? RelationFeatureProto.fromPartial(object.relation)
      : undefined;
    message.collection = (object.collection !== undefined && object.collection !== null)
      ? CollectionFeatureProto.fromPartial(object.collection)
      : undefined;
    message.expression = (object.expression !== undefined && object.expression !== null)
      ? ExpressionFeatureProto.fromPartial(object.expression)
      : undefined;
    return message;
  },
};

function createBaseCollectionProto(): CollectionProto {
  return { keys: [], values: [] };
}

export const CollectionProto = {
  encode(message: CollectionProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    for (const v of message.keys) {
      LiteralNodeProto.encode(v!, writer.uint32(18).fork()).ldelim();
    }
    for (const v of message.values) {
      LiteralNodeProto.encode(v!, writer.uint32(34).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): CollectionProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseCollectionProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 2:
          if (tag !== 18) {
            break;
          }

          message.keys.push(LiteralNodeProto.decode(reader, reader.uint32()));
          continue;
        case 4:
          if (tag !== 34) {
            break;
          }

          message.values.push(LiteralNodeProto.decode(reader, reader.uint32()));
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): CollectionProto {
    return {
      keys: globalThis.Array.isArray(object?.keys) ? object.keys.map((e: any) => LiteralNodeProto.fromJSON(e)) : [],
      values: globalThis.Array.isArray(object?.values)
        ? object.values.map((e: any) => LiteralNodeProto.fromJSON(e))
        : [],
    };
  },

  toJSON(message: CollectionProto): unknown {
    const obj: any = {};
    if (message.keys?.length) {
      obj.keys = message.keys.map((e) => LiteralNodeProto.toJSON(e));
    }
    if (message.values?.length) {
      obj.values = message.values.map((e) => LiteralNodeProto.toJSON(e));
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<CollectionProto>, I>>(base?: I): CollectionProto {
    return CollectionProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<CollectionProto>, I>>(object: I): CollectionProto {
    const message = createBaseCollectionProto();
    message.keys = object.keys?.map((e) => LiteralNodeProto.fromPartial(e)) || [];
    message.values = object.values?.map((e) => LiteralNodeProto.fromPartial(e)) || [];
    return message;
  },
};

function createBasePairProto(): PairProto {
  return { first: undefined, second: undefined };
}

export const PairProto = {
  encode(message: PairProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.first !== undefined) {
      LiteralNodeProto.encode(message.first, writer.uint32(10).fork()).ldelim();
    }
    if (message.second !== undefined) {
      LiteralNodeProto.encode(message.second, writer.uint32(18).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): PairProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBasePairProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.first = LiteralNodeProto.decode(reader, reader.uint32());
          continue;
        case 2:
          if (tag !== 18) {
            break;
          }

          message.second = LiteralNodeProto.decode(reader, reader.uint32());
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): PairProto {
    return {
      first: isSet(object.first) ? LiteralNodeProto.fromJSON(object.first) : undefined,
      second: isSet(object.second) ? LiteralNodeProto.fromJSON(object.second) : undefined,
    };
  },

  toJSON(message: PairProto): unknown {
    const obj: any = {};
    if (message.first !== undefined) {
      obj.first = LiteralNodeProto.toJSON(message.first);
    }
    if (message.second !== undefined) {
      obj.second = LiteralNodeProto.toJSON(message.second);
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<PairProto>, I>>(base?: I): PairProto {
    return PairProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<PairProto>, I>>(object: I): PairProto {
    const message = createBasePairProto();
    message.first = (object.first !== undefined && object.first !== null)
      ? LiteralNodeProto.fromPartial(object.first)
      : undefined;
    message.second = (object.second !== undefined && object.second !== null)
      ? LiteralNodeProto.fromPartial(object.second)
      : undefined;
    return message;
  },
};

function createBaseModifiedFeaturesProto(): ModifiedFeaturesProto {
  return { ids: [] };
}

export const ModifiedFeaturesProto = {
  encode(message: ModifiedFeaturesProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    for (const v of message.ids) {
      FeatureIDProto.encode(v!, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): ModifiedFeaturesProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseModifiedFeaturesProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.ids.push(FeatureIDProto.decode(reader, reader.uint32()));
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): ModifiedFeaturesProto {
    return { ids: globalThis.Array.isArray(object?.ids) ? object.ids.map((e: any) => FeatureIDProto.fromJSON(e)) : [] };
  },

  toJSON(message: ModifiedFeaturesProto): unknown {
    const obj: any = {};
    if (message.ids?.length) {
      obj.ids = message.ids.map((e) => FeatureIDProto.toJSON(e));
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<ModifiedFeaturesProto>, I>>(base?: I): ModifiedFeaturesProto {
    return ModifiedFeaturesProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<ModifiedFeaturesProto>, I>>(object: I): ModifiedFeaturesProto {
    const message = createBaseModifiedFeaturesProto();
    message.ids = object.ids?.map((e) => FeatureIDProto.fromPartial(e)) || [];
    return message;
  },
};

function createBaseAppliedChangeProto(): AppliedChangeProto {
  return { original: [], modified: [] };
}

export const AppliedChangeProto = {
  encode(message: AppliedChangeProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    for (const v of message.original) {
      FeatureIDProto.encode(v!, writer.uint32(10).fork()).ldelim();
    }
    for (const v of message.modified) {
      FeatureIDProto.encode(v!, writer.uint32(18).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): AppliedChangeProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseAppliedChangeProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.original.push(FeatureIDProto.decode(reader, reader.uint32()));
          continue;
        case 2:
          if (tag !== 18) {
            break;
          }

          message.modified.push(FeatureIDProto.decode(reader, reader.uint32()));
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): AppliedChangeProto {
    return {
      original: globalThis.Array.isArray(object?.original)
        ? object.original.map((e: any) => FeatureIDProto.fromJSON(e))
        : [],
      modified: globalThis.Array.isArray(object?.modified)
        ? object.modified.map((e: any) => FeatureIDProto.fromJSON(e))
        : [],
    };
  },

  toJSON(message: AppliedChangeProto): unknown {
    const obj: any = {};
    if (message.original?.length) {
      obj.original = message.original.map((e) => FeatureIDProto.toJSON(e));
    }
    if (message.modified?.length) {
      obj.modified = message.modified.map((e) => FeatureIDProto.toJSON(e));
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<AppliedChangeProto>, I>>(base?: I): AppliedChangeProto {
    return AppliedChangeProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<AppliedChangeProto>, I>>(object: I): AppliedChangeProto {
    const message = createBaseAppliedChangeProto();
    message.original = object.original?.map((e) => FeatureIDProto.fromPartial(e)) || [];
    message.modified = object.modified?.map((e) => FeatureIDProto.fromPartial(e)) || [];
    return message;
  },
};

function createBaseNodeProto(): NodeProto {
  return { symbol: undefined, literal: undefined, call: undefined, lambda: undefined, name: "", begin: 0, end: 0 };
}

export const NodeProto = {
  encode(message: NodeProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.symbol !== undefined) {
      writer.uint32(10).string(message.symbol);
    }
    if (message.literal !== undefined) {
      LiteralNodeProto.encode(message.literal, writer.uint32(18).fork()).ldelim();
    }
    if (message.call !== undefined) {
      CallNodeProto.encode(message.call, writer.uint32(26).fork()).ldelim();
    }
    if (message.lambda !== undefined) {
      LambdaNodeProto.encode(message.lambda, writer.uint32(34).fork()).ldelim();
    }
    if (message.name !== "") {
      writer.uint32(42).string(message.name);
    }
    if (message.begin !== 0) {
      writer.uint32(48).int32(message.begin);
    }
    if (message.end !== 0) {
      writer.uint32(56).int32(message.end);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): NodeProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseNodeProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.symbol = reader.string();
          continue;
        case 2:
          if (tag !== 18) {
            break;
          }

          message.literal = LiteralNodeProto.decode(reader, reader.uint32());
          continue;
        case 3:
          if (tag !== 26) {
            break;
          }

          message.call = CallNodeProto.decode(reader, reader.uint32());
          continue;
        case 4:
          if (tag !== 34) {
            break;
          }

          message.lambda = LambdaNodeProto.decode(reader, reader.uint32());
          continue;
        case 5:
          if (tag !== 42) {
            break;
          }

          message.name = reader.string();
          continue;
        case 6:
          if (tag !== 48) {
            break;
          }

          message.begin = reader.int32();
          continue;
        case 7:
          if (tag !== 56) {
            break;
          }

          message.end = reader.int32();
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): NodeProto {
    return {
      symbol: isSet(object.symbol) ? globalThis.String(object.symbol) : undefined,
      literal: isSet(object.literal) ? LiteralNodeProto.fromJSON(object.literal) : undefined,
      call: isSet(object.call) ? CallNodeProto.fromJSON(object.call) : undefined,
      lambda: isSet(object.lambda) ? LambdaNodeProto.fromJSON(object.lambda) : undefined,
      name: isSet(object.name) ? globalThis.String(object.name) : "",
      begin: isSet(object.begin) ? globalThis.Number(object.begin) : 0,
      end: isSet(object.end) ? globalThis.Number(object.end) : 0,
    };
  },

  toJSON(message: NodeProto): unknown {
    const obj: any = {};
    if (message.symbol !== undefined) {
      obj.symbol = message.symbol;
    }
    if (message.literal !== undefined) {
      obj.literal = LiteralNodeProto.toJSON(message.literal);
    }
    if (message.call !== undefined) {
      obj.call = CallNodeProto.toJSON(message.call);
    }
    if (message.lambda !== undefined) {
      obj.lambda = LambdaNodeProto.toJSON(message.lambda);
    }
    if (message.name !== "") {
      obj.name = message.name;
    }
    if (message.begin !== 0) {
      obj.begin = Math.round(message.begin);
    }
    if (message.end !== 0) {
      obj.end = Math.round(message.end);
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<NodeProto>, I>>(base?: I): NodeProto {
    return NodeProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<NodeProto>, I>>(object: I): NodeProto {
    const message = createBaseNodeProto();
    message.symbol = object.symbol ?? undefined;
    message.literal = (object.literal !== undefined && object.literal !== null)
      ? LiteralNodeProto.fromPartial(object.literal)
      : undefined;
    message.call = (object.call !== undefined && object.call !== null)
      ? CallNodeProto.fromPartial(object.call)
      : undefined;
    message.lambda = (object.lambda !== undefined && object.lambda !== null)
      ? LambdaNodeProto.fromPartial(object.lambda)
      : undefined;
    message.name = object.name ?? "";
    message.begin = object.begin ?? 0;
    message.end = object.end ?? 0;
    return message;
  },
};

function createBaseLiteralNodeProto(): LiteralNodeProto {
  return {
    nilValue: undefined,
    boolValue: undefined,
    stringValue: undefined,
    intValue: undefined,
    floatValue: undefined,
    collectionValue: undefined,
    pairValue: undefined,
    featureValue: undefined,
    queryValue: undefined,
    featureIDValue: undefined,
    pointValue: undefined,
    pathValue: undefined,
    areaValue: undefined,
    appliedChangeValue: undefined,
    geoJSONValue: undefined,
    tagValue: undefined,
    routeValue: undefined,
  };
}

export const LiteralNodeProto = {
  encode(message: LiteralNodeProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.nilValue !== undefined) {
      writer.uint32(8).bool(message.nilValue);
    }
    if (message.boolValue !== undefined) {
      writer.uint32(16).bool(message.boolValue);
    }
    if (message.stringValue !== undefined) {
      writer.uint32(26).string(message.stringValue);
    }
    if (message.intValue !== undefined) {
      writer.uint32(32).int64(message.intValue);
    }
    if (message.floatValue !== undefined) {
      writer.uint32(41).double(message.floatValue);
    }
    if (message.collectionValue !== undefined) {
      CollectionProto.encode(message.collectionValue, writer.uint32(50).fork()).ldelim();
    }
    if (message.pairValue !== undefined) {
      PairProto.encode(message.pairValue, writer.uint32(58).fork()).ldelim();
    }
    if (message.featureValue !== undefined) {
      FeatureProto.encode(message.featureValue, writer.uint32(66).fork()).ldelim();
    }
    if (message.queryValue !== undefined) {
      QueryProto.encode(message.queryValue, writer.uint32(74).fork()).ldelim();
    }
    if (message.featureIDValue !== undefined) {
      FeatureIDProto.encode(message.featureIDValue, writer.uint32(82).fork()).ldelim();
    }
    if (message.pointValue !== undefined) {
      PointProto.encode(message.pointValue, writer.uint32(90).fork()).ldelim();
    }
    if (message.pathValue !== undefined) {
      PolylineProto.encode(message.pathValue, writer.uint32(98).fork()).ldelim();
    }
    if (message.areaValue !== undefined) {
      MultiPolygonProto.encode(message.areaValue, writer.uint32(106).fork()).ldelim();
    }
    if (message.appliedChangeValue !== undefined) {
      AppliedChangeProto.encode(message.appliedChangeValue, writer.uint32(114).fork()).ldelim();
    }
    if (message.geoJSONValue !== undefined) {
      writer.uint32(122).bytes(message.geoJSONValue);
    }
    if (message.tagValue !== undefined) {
      TagProto.encode(message.tagValue, writer.uint32(130).fork()).ldelim();
    }
    if (message.routeValue !== undefined) {
      RouteProto.encode(message.routeValue, writer.uint32(138).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): LiteralNodeProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseLiteralNodeProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 8) {
            break;
          }

          message.nilValue = reader.bool();
          continue;
        case 2:
          if (tag !== 16) {
            break;
          }

          message.boolValue = reader.bool();
          continue;
        case 3:
          if (tag !== 26) {
            break;
          }

          message.stringValue = reader.string();
          continue;
        case 4:
          if (tag !== 32) {
            break;
          }

          message.intValue = longToNumber(reader.int64() as Long);
          continue;
        case 5:
          if (tag !== 41) {
            break;
          }

          message.floatValue = reader.double();
          continue;
        case 6:
          if (tag !== 50) {
            break;
          }

          message.collectionValue = CollectionProto.decode(reader, reader.uint32());
          continue;
        case 7:
          if (tag !== 58) {
            break;
          }

          message.pairValue = PairProto.decode(reader, reader.uint32());
          continue;
        case 8:
          if (tag !== 66) {
            break;
          }

          message.featureValue = FeatureProto.decode(reader, reader.uint32());
          continue;
        case 9:
          if (tag !== 74) {
            break;
          }

          message.queryValue = QueryProto.decode(reader, reader.uint32());
          continue;
        case 10:
          if (tag !== 82) {
            break;
          }

          message.featureIDValue = FeatureIDProto.decode(reader, reader.uint32());
          continue;
        case 11:
          if (tag !== 90) {
            break;
          }

          message.pointValue = PointProto.decode(reader, reader.uint32());
          continue;
        case 12:
          if (tag !== 98) {
            break;
          }

          message.pathValue = PolylineProto.decode(reader, reader.uint32());
          continue;
        case 13:
          if (tag !== 106) {
            break;
          }

          message.areaValue = MultiPolygonProto.decode(reader, reader.uint32());
          continue;
        case 14:
          if (tag !== 114) {
            break;
          }

          message.appliedChangeValue = AppliedChangeProto.decode(reader, reader.uint32());
          continue;
        case 15:
          if (tag !== 122) {
            break;
          }

          message.geoJSONValue = reader.bytes();
          continue;
        case 16:
          if (tag !== 130) {
            break;
          }

          message.tagValue = TagProto.decode(reader, reader.uint32());
          continue;
        case 17:
          if (tag !== 138) {
            break;
          }

          message.routeValue = RouteProto.decode(reader, reader.uint32());
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): LiteralNodeProto {
    return {
      nilValue: isSet(object.nilValue) ? globalThis.Boolean(object.nilValue) : undefined,
      boolValue: isSet(object.boolValue) ? globalThis.Boolean(object.boolValue) : undefined,
      stringValue: isSet(object.stringValue) ? globalThis.String(object.stringValue) : undefined,
      intValue: isSet(object.intValue) ? globalThis.Number(object.intValue) : undefined,
      floatValue: isSet(object.floatValue) ? globalThis.Number(object.floatValue) : undefined,
      collectionValue: isSet(object.collectionValue) ? CollectionProto.fromJSON(object.collectionValue) : undefined,
      pairValue: isSet(object.pairValue) ? PairProto.fromJSON(object.pairValue) : undefined,
      featureValue: isSet(object.featureValue) ? FeatureProto.fromJSON(object.featureValue) : undefined,
      queryValue: isSet(object.queryValue) ? QueryProto.fromJSON(object.queryValue) : undefined,
      featureIDValue: isSet(object.featureIDValue) ? FeatureIDProto.fromJSON(object.featureIDValue) : undefined,
      pointValue: isSet(object.pointValue) ? PointProto.fromJSON(object.pointValue) : undefined,
      pathValue: isSet(object.pathValue) ? PolylineProto.fromJSON(object.pathValue) : undefined,
      areaValue: isSet(object.areaValue) ? MultiPolygonProto.fromJSON(object.areaValue) : undefined,
      appliedChangeValue: isSet(object.appliedChangeValue)
        ? AppliedChangeProto.fromJSON(object.appliedChangeValue)
        : undefined,
      geoJSONValue: isSet(object.geoJSONValue) ? bytesFromBase64(object.geoJSONValue) : undefined,
      tagValue: isSet(object.tagValue) ? TagProto.fromJSON(object.tagValue) : undefined,
      routeValue: isSet(object.routeValue) ? RouteProto.fromJSON(object.routeValue) : undefined,
    };
  },

  toJSON(message: LiteralNodeProto): unknown {
    const obj: any = {};
    if (message.nilValue !== undefined) {
      obj.nilValue = message.nilValue;
    }
    if (message.boolValue !== undefined) {
      obj.boolValue = message.boolValue;
    }
    if (message.stringValue !== undefined) {
      obj.stringValue = message.stringValue;
    }
    if (message.intValue !== undefined) {
      obj.intValue = Math.round(message.intValue);
    }
    if (message.floatValue !== undefined) {
      obj.floatValue = message.floatValue;
    }
    if (message.collectionValue !== undefined) {
      obj.collectionValue = CollectionProto.toJSON(message.collectionValue);
    }
    if (message.pairValue !== undefined) {
      obj.pairValue = PairProto.toJSON(message.pairValue);
    }
    if (message.featureValue !== undefined) {
      obj.featureValue = FeatureProto.toJSON(message.featureValue);
    }
    if (message.queryValue !== undefined) {
      obj.queryValue = QueryProto.toJSON(message.queryValue);
    }
    if (message.featureIDValue !== undefined) {
      obj.featureIDValue = FeatureIDProto.toJSON(message.featureIDValue);
    }
    if (message.pointValue !== undefined) {
      obj.pointValue = PointProto.toJSON(message.pointValue);
    }
    if (message.pathValue !== undefined) {
      obj.pathValue = PolylineProto.toJSON(message.pathValue);
    }
    if (message.areaValue !== undefined) {
      obj.areaValue = MultiPolygonProto.toJSON(message.areaValue);
    }
    if (message.appliedChangeValue !== undefined) {
      obj.appliedChangeValue = AppliedChangeProto.toJSON(message.appliedChangeValue);
    }
    if (message.geoJSONValue !== undefined) {
      obj.geoJSONValue = base64FromBytes(message.geoJSONValue);
    }
    if (message.tagValue !== undefined) {
      obj.tagValue = TagProto.toJSON(message.tagValue);
    }
    if (message.routeValue !== undefined) {
      obj.routeValue = RouteProto.toJSON(message.routeValue);
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<LiteralNodeProto>, I>>(base?: I): LiteralNodeProto {
    return LiteralNodeProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<LiteralNodeProto>, I>>(object: I): LiteralNodeProto {
    const message = createBaseLiteralNodeProto();
    message.nilValue = object.nilValue ?? undefined;
    message.boolValue = object.boolValue ?? undefined;
    message.stringValue = object.stringValue ?? undefined;
    message.intValue = object.intValue ?? undefined;
    message.floatValue = object.floatValue ?? undefined;
    message.collectionValue = (object.collectionValue !== undefined && object.collectionValue !== null)
      ? CollectionProto.fromPartial(object.collectionValue)
      : undefined;
    message.pairValue = (object.pairValue !== undefined && object.pairValue !== null)
      ? PairProto.fromPartial(object.pairValue)
      : undefined;
    message.featureValue = (object.featureValue !== undefined && object.featureValue !== null)
      ? FeatureProto.fromPartial(object.featureValue)
      : undefined;
    message.queryValue = (object.queryValue !== undefined && object.queryValue !== null)
      ? QueryProto.fromPartial(object.queryValue)
      : undefined;
    message.featureIDValue = (object.featureIDValue !== undefined && object.featureIDValue !== null)
      ? FeatureIDProto.fromPartial(object.featureIDValue)
      : undefined;
    message.pointValue = (object.pointValue !== undefined && object.pointValue !== null)
      ? PointProto.fromPartial(object.pointValue)
      : undefined;
    message.pathValue = (object.pathValue !== undefined && object.pathValue !== null)
      ? PolylineProto.fromPartial(object.pathValue)
      : undefined;
    message.areaValue = (object.areaValue !== undefined && object.areaValue !== null)
      ? MultiPolygonProto.fromPartial(object.areaValue)
      : undefined;
    message.appliedChangeValue = (object.appliedChangeValue !== undefined && object.appliedChangeValue !== null)
      ? AppliedChangeProto.fromPartial(object.appliedChangeValue)
      : undefined;
    message.geoJSONValue = object.geoJSONValue ?? undefined;
    message.tagValue = (object.tagValue !== undefined && object.tagValue !== null)
      ? TagProto.fromPartial(object.tagValue)
      : undefined;
    message.routeValue = (object.routeValue !== undefined && object.routeValue !== null)
      ? RouteProto.fromPartial(object.routeValue)
      : undefined;
    return message;
  },
};

function createBaseCallNodeProto(): CallNodeProto {
  return { function: undefined, args: [], pipelined: false };
}

export const CallNodeProto = {
  encode(message: CallNodeProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.function !== undefined) {
      NodeProto.encode(message.function, writer.uint32(10).fork()).ldelim();
    }
    for (const v of message.args) {
      NodeProto.encode(v!, writer.uint32(18).fork()).ldelim();
    }
    if (message.pipelined !== false) {
      writer.uint32(24).bool(message.pipelined);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): CallNodeProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseCallNodeProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.function = NodeProto.decode(reader, reader.uint32());
          continue;
        case 2:
          if (tag !== 18) {
            break;
          }

          message.args.push(NodeProto.decode(reader, reader.uint32()));
          continue;
        case 3:
          if (tag !== 24) {
            break;
          }

          message.pipelined = reader.bool();
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): CallNodeProto {
    return {
      function: isSet(object.function) ? NodeProto.fromJSON(object.function) : undefined,
      args: globalThis.Array.isArray(object?.args) ? object.args.map((e: any) => NodeProto.fromJSON(e)) : [],
      pipelined: isSet(object.pipelined) ? globalThis.Boolean(object.pipelined) : false,
    };
  },

  toJSON(message: CallNodeProto): unknown {
    const obj: any = {};
    if (message.function !== undefined) {
      obj.function = NodeProto.toJSON(message.function);
    }
    if (message.args?.length) {
      obj.args = message.args.map((e) => NodeProto.toJSON(e));
    }
    if (message.pipelined !== false) {
      obj.pipelined = message.pipelined;
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<CallNodeProto>, I>>(base?: I): CallNodeProto {
    return CallNodeProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<CallNodeProto>, I>>(object: I): CallNodeProto {
    const message = createBaseCallNodeProto();
    message.function = (object.function !== undefined && object.function !== null)
      ? NodeProto.fromPartial(object.function)
      : undefined;
    message.args = object.args?.map((e) => NodeProto.fromPartial(e)) || [];
    message.pipelined = object.pipelined ?? false;
    return message;
  },
};

function createBaseLambdaNodeProto(): LambdaNodeProto {
  return { args: [], node: undefined };
}

export const LambdaNodeProto = {
  encode(message: LambdaNodeProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    for (const v of message.args) {
      writer.uint32(10).string(v!);
    }
    if (message.node !== undefined) {
      NodeProto.encode(message.node, writer.uint32(18).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): LambdaNodeProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseLambdaNodeProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.args.push(reader.string());
          continue;
        case 2:
          if (tag !== 18) {
            break;
          }

          message.node = NodeProto.decode(reader, reader.uint32());
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): LambdaNodeProto {
    return {
      args: globalThis.Array.isArray(object?.args) ? object.args.map((e: any) => globalThis.String(e)) : [],
      node: isSet(object.node) ? NodeProto.fromJSON(object.node) : undefined,
    };
  },

  toJSON(message: LambdaNodeProto): unknown {
    const obj: any = {};
    if (message.args?.length) {
      obj.args = message.args;
    }
    if (message.node !== undefined) {
      obj.node = NodeProto.toJSON(message.node);
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<LambdaNodeProto>, I>>(base?: I): LambdaNodeProto {
    return LambdaNodeProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<LambdaNodeProto>, I>>(object: I): LambdaNodeProto {
    const message = createBaseLambdaNodeProto();
    message.args = object.args?.map((e) => e) || [];
    message.node = (object.node !== undefined && object.node !== null) ? NodeProto.fromPartial(object.node) : undefined;
    return message;
  },
};

function createBaseKeyQueryProto(): KeyQueryProto {
  return { key: "" };
}

export const KeyQueryProto = {
  encode(message: KeyQueryProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.key !== "") {
      writer.uint32(10).string(message.key);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): KeyQueryProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseKeyQueryProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.key = reader.string();
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): KeyQueryProto {
    return { key: isSet(object.key) ? globalThis.String(object.key) : "" };
  },

  toJSON(message: KeyQueryProto): unknown {
    const obj: any = {};
    if (message.key !== "") {
      obj.key = message.key;
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<KeyQueryProto>, I>>(base?: I): KeyQueryProto {
    return KeyQueryProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<KeyQueryProto>, I>>(object: I): KeyQueryProto {
    const message = createBaseKeyQueryProto();
    message.key = object.key ?? "";
    return message;
  },
};

function createBaseKeyValueQueryProto(): KeyValueQueryProto {
  return { key: "", value: "" };
}

export const KeyValueQueryProto = {
  encode(message: KeyValueQueryProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.key !== "") {
      writer.uint32(10).string(message.key);
    }
    if (message.value !== "") {
      writer.uint32(18).string(message.value);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): KeyValueQueryProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseKeyValueQueryProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.key = reader.string();
          continue;
        case 2:
          if (tag !== 18) {
            break;
          }

          message.value = reader.string();
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): KeyValueQueryProto {
    return {
      key: isSet(object.key) ? globalThis.String(object.key) : "",
      value: isSet(object.value) ? globalThis.String(object.value) : "",
    };
  },

  toJSON(message: KeyValueQueryProto): unknown {
    const obj: any = {};
    if (message.key !== "") {
      obj.key = message.key;
    }
    if (message.value !== "") {
      obj.value = message.value;
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<KeyValueQueryProto>, I>>(base?: I): KeyValueQueryProto {
    return KeyValueQueryProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<KeyValueQueryProto>, I>>(object: I): KeyValueQueryProto {
    const message = createBaseKeyValueQueryProto();
    message.key = object.key ?? "";
    message.value = object.value ?? "";
    return message;
  },
};

function createBaseTypedQueryProto(): TypedQueryProto {
  return { type: 0, query: undefined };
}

export const TypedQueryProto = {
  encode(message: TypedQueryProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.type !== 0) {
      writer.uint32(8).int32(message.type);
    }
    if (message.query !== undefined) {
      QueryProto.encode(message.query, writer.uint32(18).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): TypedQueryProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseTypedQueryProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 8) {
            break;
          }

          message.type = reader.int32() as any;
          continue;
        case 2:
          if (tag !== 18) {
            break;
          }

          message.query = QueryProto.decode(reader, reader.uint32());
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): TypedQueryProto {
    return {
      type: isSet(object.type) ? featureTypeFromJSON(object.type) : 0,
      query: isSet(object.query) ? QueryProto.fromJSON(object.query) : undefined,
    };
  },

  toJSON(message: TypedQueryProto): unknown {
    const obj: any = {};
    if (message.type !== 0) {
      obj.type = featureTypeToJSON(message.type);
    }
    if (message.query !== undefined) {
      obj.query = QueryProto.toJSON(message.query);
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<TypedQueryProto>, I>>(base?: I): TypedQueryProto {
    return TypedQueryProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<TypedQueryProto>, I>>(object: I): TypedQueryProto {
    const message = createBaseTypedQueryProto();
    message.type = object.type ?? 0;
    message.query = (object.query !== undefined && object.query !== null)
      ? QueryProto.fromPartial(object.query)
      : undefined;
    return message;
  },
};

function createBaseQueriesProto(): QueriesProto {
  return { queries: [] };
}

export const QueriesProto = {
  encode(message: QueriesProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    for (const v of message.queries) {
      QueryProto.encode(v!, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): QueriesProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseQueriesProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.queries.push(QueryProto.decode(reader, reader.uint32()));
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): QueriesProto {
    return {
      queries: globalThis.Array.isArray(object?.queries) ? object.queries.map((e: any) => QueryProto.fromJSON(e)) : [],
    };
  },

  toJSON(message: QueriesProto): unknown {
    const obj: any = {};
    if (message.queries?.length) {
      obj.queries = message.queries.map((e) => QueryProto.toJSON(e));
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<QueriesProto>, I>>(base?: I): QueriesProto {
    return QueriesProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<QueriesProto>, I>>(object: I): QueriesProto {
    const message = createBaseQueriesProto();
    message.queries = object.queries?.map((e) => QueryProto.fromPartial(e)) || [];
    return message;
  },
};

function createBaseAllQueryProto(): AllQueryProto {
  return {};
}

export const AllQueryProto = {
  encode(_: AllQueryProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): AllQueryProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseAllQueryProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(_: any): AllQueryProto {
    return {};
  },

  toJSON(_: AllQueryProto): unknown {
    const obj: any = {};
    return obj;
  },

  create<I extends Exact<DeepPartial<AllQueryProto>, I>>(base?: I): AllQueryProto {
    return AllQueryProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<AllQueryProto>, I>>(_: I): AllQueryProto {
    const message = createBaseAllQueryProto();
    return message;
  },
};

function createBaseEmptyQueryProto(): EmptyQueryProto {
  return {};
}

export const EmptyQueryProto = {
  encode(_: EmptyQueryProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): EmptyQueryProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseEmptyQueryProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(_: any): EmptyQueryProto {
    return {};
  },

  toJSON(_: EmptyQueryProto): unknown {
    const obj: any = {};
    return obj;
  },

  create<I extends Exact<DeepPartial<EmptyQueryProto>, I>>(base?: I): EmptyQueryProto {
    return EmptyQueryProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<EmptyQueryProto>, I>>(_: I): EmptyQueryProto {
    const message = createBaseEmptyQueryProto();
    return message;
  },
};

function createBaseCapProto(): CapProto {
  return { center: undefined, radiusMeters: 0 };
}

export const CapProto = {
  encode(message: CapProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.center !== undefined) {
      PointProto.encode(message.center, writer.uint32(10).fork()).ldelim();
    }
    if (message.radiusMeters !== 0) {
      writer.uint32(17).double(message.radiusMeters);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): CapProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseCapProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.center = PointProto.decode(reader, reader.uint32());
          continue;
        case 2:
          if (tag !== 17) {
            break;
          }

          message.radiusMeters = reader.double();
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): CapProto {
    return {
      center: isSet(object.center) ? PointProto.fromJSON(object.center) : undefined,
      radiusMeters: isSet(object.radiusMeters) ? globalThis.Number(object.radiusMeters) : 0,
    };
  },

  toJSON(message: CapProto): unknown {
    const obj: any = {};
    if (message.center !== undefined) {
      obj.center = PointProto.toJSON(message.center);
    }
    if (message.radiusMeters !== 0) {
      obj.radiusMeters = message.radiusMeters;
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<CapProto>, I>>(base?: I): CapProto {
    return CapProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<CapProto>, I>>(object: I): CapProto {
    const message = createBaseCapProto();
    message.center = (object.center !== undefined && object.center !== null)
      ? PointProto.fromPartial(object.center)
      : undefined;
    message.radiusMeters = object.radiusMeters ?? 0;
    return message;
  },
};

function createBaseS2CellIDsProto(): S2CellIDsProto {
  return { s2CellIDs: [] };
}

export const S2CellIDsProto = {
  encode(message: S2CellIDsProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    writer.uint32(10).fork();
    for (const v of message.s2CellIDs) {
      writer.uint64(v);
    }
    writer.ldelim();
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): S2CellIDsProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseS2CellIDsProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag === 8) {
            message.s2CellIDs.push(longToNumber(reader.uint64() as Long));

            continue;
          }

          if (tag === 10) {
            const end2 = reader.uint32() + reader.pos;
            while (reader.pos < end2) {
              message.s2CellIDs.push(longToNumber(reader.uint64() as Long));
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

  fromJSON(object: any): S2CellIDsProto {
    return {
      s2CellIDs: globalThis.Array.isArray(object?.s2CellIDs)
        ? object.s2CellIDs.map((e: any) => globalThis.Number(e))
        : [],
    };
  },

  toJSON(message: S2CellIDsProto): unknown {
    const obj: any = {};
    if (message.s2CellIDs?.length) {
      obj.s2CellIDs = message.s2CellIDs.map((e) => Math.round(e));
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<S2CellIDsProto>, I>>(base?: I): S2CellIDsProto {
    return S2CellIDsProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<S2CellIDsProto>, I>>(object: I): S2CellIDsProto {
    const message = createBaseS2CellIDsProto();
    message.s2CellIDs = object.s2CellIDs?.map((e) => e) || [];
    return message;
  },
};

function createBaseQueryProto(): QueryProto {
  return {
    all: undefined,
    empty: undefined,
    keyed: undefined,
    tagged: undefined,
    typed: undefined,
    intersection: undefined,
    union: undefined,
    intersectsCap: undefined,
    intersectsFeature: undefined,
    intersectsPoint: undefined,
    intersectsPolyline: undefined,
    intersectsMultiPolygon: undefined,
    intersectsCells: undefined,
    mightIntersect: undefined,
  };
}

export const QueryProto = {
  encode(message: QueryProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.all !== undefined) {
      AllQueryProto.encode(message.all, writer.uint32(10).fork()).ldelim();
    }
    if (message.empty !== undefined) {
      EmptyQueryProto.encode(message.empty, writer.uint32(18).fork()).ldelim();
    }
    if (message.keyed !== undefined) {
      writer.uint32(26).string(message.keyed);
    }
    if (message.tagged !== undefined) {
      TagProto.encode(message.tagged, writer.uint32(34).fork()).ldelim();
    }
    if (message.typed !== undefined) {
      TypedQueryProto.encode(message.typed, writer.uint32(42).fork()).ldelim();
    }
    if (message.intersection !== undefined) {
      QueriesProto.encode(message.intersection, writer.uint32(50).fork()).ldelim();
    }
    if (message.union !== undefined) {
      QueriesProto.encode(message.union, writer.uint32(58).fork()).ldelim();
    }
    if (message.intersectsCap !== undefined) {
      CapProto.encode(message.intersectsCap, writer.uint32(66).fork()).ldelim();
    }
    if (message.intersectsFeature !== undefined) {
      FeatureIDProto.encode(message.intersectsFeature, writer.uint32(74).fork()).ldelim();
    }
    if (message.intersectsPoint !== undefined) {
      PointProto.encode(message.intersectsPoint, writer.uint32(82).fork()).ldelim();
    }
    if (message.intersectsPolyline !== undefined) {
      PolylineProto.encode(message.intersectsPolyline, writer.uint32(90).fork()).ldelim();
    }
    if (message.intersectsMultiPolygon !== undefined) {
      MultiPolygonProto.encode(message.intersectsMultiPolygon, writer.uint32(98).fork()).ldelim();
    }
    if (message.intersectsCells !== undefined) {
      S2CellIDsProto.encode(message.intersectsCells, writer.uint32(106).fork()).ldelim();
    }
    if (message.mightIntersect !== undefined) {
      S2CellIDsProto.encode(message.mightIntersect, writer.uint32(114).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): QueryProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseQueryProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.all = AllQueryProto.decode(reader, reader.uint32());
          continue;
        case 2:
          if (tag !== 18) {
            break;
          }

          message.empty = EmptyQueryProto.decode(reader, reader.uint32());
          continue;
        case 3:
          if (tag !== 26) {
            break;
          }

          message.keyed = reader.string();
          continue;
        case 4:
          if (tag !== 34) {
            break;
          }

          message.tagged = TagProto.decode(reader, reader.uint32());
          continue;
        case 5:
          if (tag !== 42) {
            break;
          }

          message.typed = TypedQueryProto.decode(reader, reader.uint32());
          continue;
        case 6:
          if (tag !== 50) {
            break;
          }

          message.intersection = QueriesProto.decode(reader, reader.uint32());
          continue;
        case 7:
          if (tag !== 58) {
            break;
          }

          message.union = QueriesProto.decode(reader, reader.uint32());
          continue;
        case 8:
          if (tag !== 66) {
            break;
          }

          message.intersectsCap = CapProto.decode(reader, reader.uint32());
          continue;
        case 9:
          if (tag !== 74) {
            break;
          }

          message.intersectsFeature = FeatureIDProto.decode(reader, reader.uint32());
          continue;
        case 10:
          if (tag !== 82) {
            break;
          }

          message.intersectsPoint = PointProto.decode(reader, reader.uint32());
          continue;
        case 11:
          if (tag !== 90) {
            break;
          }

          message.intersectsPolyline = PolylineProto.decode(reader, reader.uint32());
          continue;
        case 12:
          if (tag !== 98) {
            break;
          }

          message.intersectsMultiPolygon = MultiPolygonProto.decode(reader, reader.uint32());
          continue;
        case 13:
          if (tag !== 106) {
            break;
          }

          message.intersectsCells = S2CellIDsProto.decode(reader, reader.uint32());
          continue;
        case 14:
          if (tag !== 114) {
            break;
          }

          message.mightIntersect = S2CellIDsProto.decode(reader, reader.uint32());
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): QueryProto {
    return {
      all: isSet(object.all) ? AllQueryProto.fromJSON(object.all) : undefined,
      empty: isSet(object.empty) ? EmptyQueryProto.fromJSON(object.empty) : undefined,
      keyed: isSet(object.keyed) ? globalThis.String(object.keyed) : undefined,
      tagged: isSet(object.tagged) ? TagProto.fromJSON(object.tagged) : undefined,
      typed: isSet(object.typed) ? TypedQueryProto.fromJSON(object.typed) : undefined,
      intersection: isSet(object.intersection) ? QueriesProto.fromJSON(object.intersection) : undefined,
      union: isSet(object.union) ? QueriesProto.fromJSON(object.union) : undefined,
      intersectsCap: isSet(object.intersectsCap) ? CapProto.fromJSON(object.intersectsCap) : undefined,
      intersectsFeature: isSet(object.intersectsFeature)
        ? FeatureIDProto.fromJSON(object.intersectsFeature)
        : undefined,
      intersectsPoint: isSet(object.intersectsPoint) ? PointProto.fromJSON(object.intersectsPoint) : undefined,
      intersectsPolyline: isSet(object.intersectsPolyline)
        ? PolylineProto.fromJSON(object.intersectsPolyline)
        : undefined,
      intersectsMultiPolygon: isSet(object.intersectsMultiPolygon)
        ? MultiPolygonProto.fromJSON(object.intersectsMultiPolygon)
        : undefined,
      intersectsCells: isSet(object.intersectsCells) ? S2CellIDsProto.fromJSON(object.intersectsCells) : undefined,
      mightIntersect: isSet(object.mightIntersect) ? S2CellIDsProto.fromJSON(object.mightIntersect) : undefined,
    };
  },

  toJSON(message: QueryProto): unknown {
    const obj: any = {};
    if (message.all !== undefined) {
      obj.all = AllQueryProto.toJSON(message.all);
    }
    if (message.empty !== undefined) {
      obj.empty = EmptyQueryProto.toJSON(message.empty);
    }
    if (message.keyed !== undefined) {
      obj.keyed = message.keyed;
    }
    if (message.tagged !== undefined) {
      obj.tagged = TagProto.toJSON(message.tagged);
    }
    if (message.typed !== undefined) {
      obj.typed = TypedQueryProto.toJSON(message.typed);
    }
    if (message.intersection !== undefined) {
      obj.intersection = QueriesProto.toJSON(message.intersection);
    }
    if (message.union !== undefined) {
      obj.union = QueriesProto.toJSON(message.union);
    }
    if (message.intersectsCap !== undefined) {
      obj.intersectsCap = CapProto.toJSON(message.intersectsCap);
    }
    if (message.intersectsFeature !== undefined) {
      obj.intersectsFeature = FeatureIDProto.toJSON(message.intersectsFeature);
    }
    if (message.intersectsPoint !== undefined) {
      obj.intersectsPoint = PointProto.toJSON(message.intersectsPoint);
    }
    if (message.intersectsPolyline !== undefined) {
      obj.intersectsPolyline = PolylineProto.toJSON(message.intersectsPolyline);
    }
    if (message.intersectsMultiPolygon !== undefined) {
      obj.intersectsMultiPolygon = MultiPolygonProto.toJSON(message.intersectsMultiPolygon);
    }
    if (message.intersectsCells !== undefined) {
      obj.intersectsCells = S2CellIDsProto.toJSON(message.intersectsCells);
    }
    if (message.mightIntersect !== undefined) {
      obj.mightIntersect = S2CellIDsProto.toJSON(message.mightIntersect);
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<QueryProto>, I>>(base?: I): QueryProto {
    return QueryProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<QueryProto>, I>>(object: I): QueryProto {
    const message = createBaseQueryProto();
    message.all = (object.all !== undefined && object.all !== null) ? AllQueryProto.fromPartial(object.all) : undefined;
    message.empty = (object.empty !== undefined && object.empty !== null)
      ? EmptyQueryProto.fromPartial(object.empty)
      : undefined;
    message.keyed = object.keyed ?? undefined;
    message.tagged = (object.tagged !== undefined && object.tagged !== null)
      ? TagProto.fromPartial(object.tagged)
      : undefined;
    message.typed = (object.typed !== undefined && object.typed !== null)
      ? TypedQueryProto.fromPartial(object.typed)
      : undefined;
    message.intersection = (object.intersection !== undefined && object.intersection !== null)
      ? QueriesProto.fromPartial(object.intersection)
      : undefined;
    message.union = (object.union !== undefined && object.union !== null)
      ? QueriesProto.fromPartial(object.union)
      : undefined;
    message.intersectsCap = (object.intersectsCap !== undefined && object.intersectsCap !== null)
      ? CapProto.fromPartial(object.intersectsCap)
      : undefined;
    message.intersectsFeature = (object.intersectsFeature !== undefined && object.intersectsFeature !== null)
      ? FeatureIDProto.fromPartial(object.intersectsFeature)
      : undefined;
    message.intersectsPoint = (object.intersectsPoint !== undefined && object.intersectsPoint !== null)
      ? PointProto.fromPartial(object.intersectsPoint)
      : undefined;
    message.intersectsPolyline = (object.intersectsPolyline !== undefined && object.intersectsPolyline !== null)
      ? PolylineProto.fromPartial(object.intersectsPolyline)
      : undefined;
    message.intersectsMultiPolygon =
      (object.intersectsMultiPolygon !== undefined && object.intersectsMultiPolygon !== null)
        ? MultiPolygonProto.fromPartial(object.intersectsMultiPolygon)
        : undefined;
    message.intersectsCells = (object.intersectsCells !== undefined && object.intersectsCells !== null)
      ? S2CellIDsProto.fromPartial(object.intersectsCells)
      : undefined;
    message.mightIntersect = (object.mightIntersect !== undefined && object.mightIntersect !== null)
      ? S2CellIDsProto.fromPartial(object.mightIntersect)
      : undefined;
    return message;
  },
};

function createBaseStepProto(): StepProto {
  return { destination: undefined, via: undefined, cost: 0 };
}

export const StepProto = {
  encode(message: StepProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.destination !== undefined) {
      FeatureIDProto.encode(message.destination, writer.uint32(10).fork()).ldelim();
    }
    if (message.via !== undefined) {
      FeatureIDProto.encode(message.via, writer.uint32(18).fork()).ldelim();
    }
    if (message.cost !== 0) {
      writer.uint32(25).double(message.cost);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): StepProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseStepProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.destination = FeatureIDProto.decode(reader, reader.uint32());
          continue;
        case 2:
          if (tag !== 18) {
            break;
          }

          message.via = FeatureIDProto.decode(reader, reader.uint32());
          continue;
        case 3:
          if (tag !== 25) {
            break;
          }

          message.cost = reader.double();
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): StepProto {
    return {
      destination: isSet(object.destination) ? FeatureIDProto.fromJSON(object.destination) : undefined,
      via: isSet(object.via) ? FeatureIDProto.fromJSON(object.via) : undefined,
      cost: isSet(object.cost) ? globalThis.Number(object.cost) : 0,
    };
  },

  toJSON(message: StepProto): unknown {
    const obj: any = {};
    if (message.destination !== undefined) {
      obj.destination = FeatureIDProto.toJSON(message.destination);
    }
    if (message.via !== undefined) {
      obj.via = FeatureIDProto.toJSON(message.via);
    }
    if (message.cost !== 0) {
      obj.cost = message.cost;
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<StepProto>, I>>(base?: I): StepProto {
    return StepProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<StepProto>, I>>(object: I): StepProto {
    const message = createBaseStepProto();
    message.destination = (object.destination !== undefined && object.destination !== null)
      ? FeatureIDProto.fromPartial(object.destination)
      : undefined;
    message.via = (object.via !== undefined && object.via !== null)
      ? FeatureIDProto.fromPartial(object.via)
      : undefined;
    message.cost = object.cost ?? 0;
    return message;
  },
};

function createBaseRouteProto(): RouteProto {
  return { origin: undefined, steps: [] };
}

export const RouteProto = {
  encode(message: RouteProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.origin !== undefined) {
      FeatureIDProto.encode(message.origin, writer.uint32(10).fork()).ldelim();
    }
    for (const v of message.steps) {
      StepProto.encode(v!, writer.uint32(18).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): RouteProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseRouteProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.origin = FeatureIDProto.decode(reader, reader.uint32());
          continue;
        case 2:
          if (tag !== 18) {
            break;
          }

          message.steps.push(StepProto.decode(reader, reader.uint32()));
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): RouteProto {
    return {
      origin: isSet(object.origin) ? FeatureIDProto.fromJSON(object.origin) : undefined,
      steps: globalThis.Array.isArray(object?.steps) ? object.steps.map((e: any) => StepProto.fromJSON(e)) : [],
    };
  },

  toJSON(message: RouteProto): unknown {
    const obj: any = {};
    if (message.origin !== undefined) {
      obj.origin = FeatureIDProto.toJSON(message.origin);
    }
    if (message.steps?.length) {
      obj.steps = message.steps.map((e) => StepProto.toJSON(e));
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<RouteProto>, I>>(base?: I): RouteProto {
    return RouteProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<RouteProto>, I>>(object: I): RouteProto {
    const message = createBaseRouteProto();
    message.origin = (object.origin !== undefined && object.origin !== null)
      ? FeatureIDProto.fromPartial(object.origin)
      : undefined;
    message.steps = object.steps?.map((e) => StepProto.fromPartial(e)) || [];
    return message;
  },
};

function createBaseFindFeatureByIDRequestProto(): FindFeatureByIDRequestProto {
  return { id: undefined };
}

export const FindFeatureByIDRequestProto = {
  encode(message: FindFeatureByIDRequestProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.id !== undefined) {
      FeatureIDProto.encode(message.id, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): FindFeatureByIDRequestProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseFindFeatureByIDRequestProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
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

  fromJSON(object: any): FindFeatureByIDRequestProto {
    return { id: isSet(object.id) ? FeatureIDProto.fromJSON(object.id) : undefined };
  },

  toJSON(message: FindFeatureByIDRequestProto): unknown {
    const obj: any = {};
    if (message.id !== undefined) {
      obj.id = FeatureIDProto.toJSON(message.id);
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<FindFeatureByIDRequestProto>, I>>(base?: I): FindFeatureByIDRequestProto {
    return FindFeatureByIDRequestProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<FindFeatureByIDRequestProto>, I>>(object: I): FindFeatureByIDRequestProto {
    const message = createBaseFindFeatureByIDRequestProto();
    message.id = (object.id !== undefined && object.id !== null) ? FeatureIDProto.fromPartial(object.id) : undefined;
    return message;
  },
};

function createBaseFindFeatureByIDResponseProto(): FindFeatureByIDResponseProto {
  return { feature: undefined };
}

export const FindFeatureByIDResponseProto = {
  encode(message: FindFeatureByIDResponseProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.feature !== undefined) {
      FeatureProto.encode(message.feature, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): FindFeatureByIDResponseProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseFindFeatureByIDResponseProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.feature = FeatureProto.decode(reader, reader.uint32());
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): FindFeatureByIDResponseProto {
    return { feature: isSet(object.feature) ? FeatureProto.fromJSON(object.feature) : undefined };
  },

  toJSON(message: FindFeatureByIDResponseProto): unknown {
    const obj: any = {};
    if (message.feature !== undefined) {
      obj.feature = FeatureProto.toJSON(message.feature);
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<FindFeatureByIDResponseProto>, I>>(base?: I): FindFeatureByIDResponseProto {
    return FindFeatureByIDResponseProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<FindFeatureByIDResponseProto>, I>>(object: I): FindFeatureByIDResponseProto {
    const message = createBaseFindFeatureByIDResponseProto();
    message.feature = (object.feature !== undefined && object.feature !== null)
      ? FeatureProto.fromPartial(object.feature)
      : undefined;
    return message;
  },
};

function createBaseFindFeaturesRequestProto(): FindFeaturesRequestProto {
  return { query: undefined };
}

export const FindFeaturesRequestProto = {
  encode(message: FindFeaturesRequestProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.query !== undefined) {
      QueryProto.encode(message.query, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): FindFeaturesRequestProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseFindFeaturesRequestProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.query = QueryProto.decode(reader, reader.uint32());
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): FindFeaturesRequestProto {
    return { query: isSet(object.query) ? QueryProto.fromJSON(object.query) : undefined };
  },

  toJSON(message: FindFeaturesRequestProto): unknown {
    const obj: any = {};
    if (message.query !== undefined) {
      obj.query = QueryProto.toJSON(message.query);
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<FindFeaturesRequestProto>, I>>(base?: I): FindFeaturesRequestProto {
    return FindFeaturesRequestProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<FindFeaturesRequestProto>, I>>(object: I): FindFeaturesRequestProto {
    const message = createBaseFindFeaturesRequestProto();
    message.query = (object.query !== undefined && object.query !== null)
      ? QueryProto.fromPartial(object.query)
      : undefined;
    return message;
  },
};

function createBaseFindFeaturesResponseProto(): FindFeaturesResponseProto {
  return { features: [] };
}

export const FindFeaturesResponseProto = {
  encode(message: FindFeaturesResponseProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    for (const v of message.features) {
      FeatureProto.encode(v!, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): FindFeaturesResponseProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseFindFeaturesResponseProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.features.push(FeatureProto.decode(reader, reader.uint32()));
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): FindFeaturesResponseProto {
    return {
      features: globalThis.Array.isArray(object?.features)
        ? object.features.map((e: any) => FeatureProto.fromJSON(e))
        : [],
    };
  },

  toJSON(message: FindFeaturesResponseProto): unknown {
    const obj: any = {};
    if (message.features?.length) {
      obj.features = message.features.map((e) => FeatureProto.toJSON(e));
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<FindFeaturesResponseProto>, I>>(base?: I): FindFeaturesResponseProto {
    return FindFeaturesResponseProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<FindFeaturesResponseProto>, I>>(object: I): FindFeaturesResponseProto {
    const message = createBaseFindFeaturesResponseProto();
    message.features = object.features?.map((e) => FeatureProto.fromPartial(e)) || [];
    return message;
  },
};

function createBaseModifyTagsRequestProto(): ModifyTagsRequestProto {
  return { id: undefined, tags: [] };
}

export const ModifyTagsRequestProto = {
  encode(message: ModifyTagsRequestProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.id !== undefined) {
      FeatureIDProto.encode(message.id, writer.uint32(10).fork()).ldelim();
    }
    for (const v of message.tags) {
      TagProto.encode(v!, writer.uint32(18).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): ModifyTagsRequestProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseModifyTagsRequestProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.id = FeatureIDProto.decode(reader, reader.uint32());
          continue;
        case 2:
          if (tag !== 18) {
            break;
          }

          message.tags.push(TagProto.decode(reader, reader.uint32()));
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): ModifyTagsRequestProto {
    return {
      id: isSet(object.id) ? FeatureIDProto.fromJSON(object.id) : undefined,
      tags: globalThis.Array.isArray(object?.tags) ? object.tags.map((e: any) => TagProto.fromJSON(e)) : [],
    };
  },

  toJSON(message: ModifyTagsRequestProto): unknown {
    const obj: any = {};
    if (message.id !== undefined) {
      obj.id = FeatureIDProto.toJSON(message.id);
    }
    if (message.tags?.length) {
      obj.tags = message.tags.map((e) => TagProto.toJSON(e));
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<ModifyTagsRequestProto>, I>>(base?: I): ModifyTagsRequestProto {
    return ModifyTagsRequestProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<ModifyTagsRequestProto>, I>>(object: I): ModifyTagsRequestProto {
    const message = createBaseModifyTagsRequestProto();
    message.id = (object.id !== undefined && object.id !== null) ? FeatureIDProto.fromPartial(object.id) : undefined;
    message.tags = object.tags?.map((e) => TagProto.fromPartial(e)) || [];
    return message;
  },
};

function createBaseModifyTagsBatchRequestProto(): ModifyTagsBatchRequestProto {
  return { requests: [] };
}

export const ModifyTagsBatchRequestProto = {
  encode(message: ModifyTagsBatchRequestProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    for (const v of message.requests) {
      ModifyTagsRequestProto.encode(v!, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): ModifyTagsBatchRequestProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseModifyTagsBatchRequestProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.requests.push(ModifyTagsRequestProto.decode(reader, reader.uint32()));
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): ModifyTagsBatchRequestProto {
    return {
      requests: globalThis.Array.isArray(object?.requests)
        ? object.requests.map((e: any) => ModifyTagsRequestProto.fromJSON(e))
        : [],
    };
  },

  toJSON(message: ModifyTagsBatchRequestProto): unknown {
    const obj: any = {};
    if (message.requests?.length) {
      obj.requests = message.requests.map((e) => ModifyTagsRequestProto.toJSON(e));
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<ModifyTagsBatchRequestProto>, I>>(base?: I): ModifyTagsBatchRequestProto {
    return ModifyTagsBatchRequestProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<ModifyTagsBatchRequestProto>, I>>(object: I): ModifyTagsBatchRequestProto {
    const message = createBaseModifyTagsBatchRequestProto();
    message.requests = object.requests?.map((e) => ModifyTagsRequestProto.fromPartial(e)) || [];
    return message;
  },
};

function createBaseModifyTagsBatchResponseProto(): ModifyTagsBatchResponseProto {
  return {};
}

export const ModifyTagsBatchResponseProto = {
  encode(_: ModifyTagsBatchResponseProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): ModifyTagsBatchResponseProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseModifyTagsBatchResponseProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(_: any): ModifyTagsBatchResponseProto {
    return {};
  },

  toJSON(_: ModifyTagsBatchResponseProto): unknown {
    const obj: any = {};
    return obj;
  },

  create<I extends Exact<DeepPartial<ModifyTagsBatchResponseProto>, I>>(base?: I): ModifyTagsBatchResponseProto {
    return ModifyTagsBatchResponseProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<ModifyTagsBatchResponseProto>, I>>(_: I): ModifyTagsBatchResponseProto {
    const message = createBaseModifyTagsBatchResponseProto();
    return message;
  },
};

function createBaseEvaluateRequestProto(): EvaluateRequestProto {
  return { request: undefined, version: "", root: undefined };
}

export const EvaluateRequestProto = {
  encode(message: EvaluateRequestProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.request !== undefined) {
      NodeProto.encode(message.request, writer.uint32(10).fork()).ldelim();
    }
    if (message.version !== "") {
      writer.uint32(18).string(message.version);
    }
    if (message.root !== undefined) {
      FeatureIDProto.encode(message.root, writer.uint32(26).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): EvaluateRequestProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseEvaluateRequestProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.request = NodeProto.decode(reader, reader.uint32());
          continue;
        case 2:
          if (tag !== 18) {
            break;
          }

          message.version = reader.string();
          continue;
        case 3:
          if (tag !== 26) {
            break;
          }

          message.root = FeatureIDProto.decode(reader, reader.uint32());
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): EvaluateRequestProto {
    return {
      request: isSet(object.request) ? NodeProto.fromJSON(object.request) : undefined,
      version: isSet(object.version) ? globalThis.String(object.version) : "",
      root: isSet(object.root) ? FeatureIDProto.fromJSON(object.root) : undefined,
    };
  },

  toJSON(message: EvaluateRequestProto): unknown {
    const obj: any = {};
    if (message.request !== undefined) {
      obj.request = NodeProto.toJSON(message.request);
    }
    if (message.version !== "") {
      obj.version = message.version;
    }
    if (message.root !== undefined) {
      obj.root = FeatureIDProto.toJSON(message.root);
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<EvaluateRequestProto>, I>>(base?: I): EvaluateRequestProto {
    return EvaluateRequestProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<EvaluateRequestProto>, I>>(object: I): EvaluateRequestProto {
    const message = createBaseEvaluateRequestProto();
    message.request = (object.request !== undefined && object.request !== null)
      ? NodeProto.fromPartial(object.request)
      : undefined;
    message.version = object.version ?? "";
    message.root = (object.root !== undefined && object.root !== null)
      ? FeatureIDProto.fromPartial(object.root)
      : undefined;
    return message;
  },
};

function createBaseEvaluateResponseProto(): EvaluateResponseProto {
  return { result: undefined };
}

export const EvaluateResponseProto = {
  encode(message: EvaluateResponseProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.result !== undefined) {
      NodeProto.encode(message.result, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): EvaluateResponseProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseEvaluateResponseProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.result = NodeProto.decode(reader, reader.uint32());
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): EvaluateResponseProto {
    return { result: isSet(object.result) ? NodeProto.fromJSON(object.result) : undefined };
  },

  toJSON(message: EvaluateResponseProto): unknown {
    const obj: any = {};
    if (message.result !== undefined) {
      obj.result = NodeProto.toJSON(message.result);
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<EvaluateResponseProto>, I>>(base?: I): EvaluateResponseProto {
    return EvaluateResponseProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<EvaluateResponseProto>, I>>(object: I): EvaluateResponseProto {
    const message = createBaseEvaluateResponseProto();
    message.result = (object.result !== undefined && object.result !== null)
      ? NodeProto.fromPartial(object.result)
      : undefined;
    return message;
  },
};

export interface B6 {
  Evaluate(request: EvaluateRequestProto): Promise<EvaluateResponseProto>;
}

export const B6ServiceName = "api.B6";
export class B6ClientImpl implements B6 {
  private readonly rpc: Rpc;
  private readonly service: string;
  constructor(rpc: Rpc, opts?: { service?: string }) {
    this.service = opts?.service || B6ServiceName;
    this.rpc = rpc;
    this.Evaluate = this.Evaluate.bind(this);
  }
  Evaluate(request: EvaluateRequestProto): Promise<EvaluateResponseProto> {
    const data = EvaluateRequestProto.encode(request).finish();
    const promise = this.rpc.request(this.service, "Evaluate", data);
    return promise.then((data) => EvaluateResponseProto.decode(_m0.Reader.create(data)));
  }
}

interface Rpc {
  request(service: string, method: string, data: Uint8Array): Promise<Uint8Array>;
}

function bytesFromBase64(b64: string): Uint8Array {
  if ((globalThis as any).Buffer) {
    return Uint8Array.from(globalThis.Buffer.from(b64, "base64"));
  } else {
    const bin = globalThis.atob(b64);
    const arr = new Uint8Array(bin.length);
    for (let i = 0; i < bin.length; ++i) {
      arr[i] = bin.charCodeAt(i);
    }
    return arr;
  }
}

function base64FromBytes(arr: Uint8Array): string {
  if ((globalThis as any).Buffer) {
    return globalThis.Buffer.from(arr).toString("base64");
  } else {
    const bin: string[] = [];
    arr.forEach((byte) => {
      bin.push(globalThis.String.fromCharCode(byte));
    });
    return globalThis.btoa(bin.join(""));
  }
}

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
