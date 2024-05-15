/* eslint-disable */
import _m0 from "protobufjs/minimal";

export const protobufPackage = "geometry";

export interface PolylineProto {
  points?: PointProto[] | undefined;
  lengthMeters?: number | undefined;
}

export interface MultiPolygonProto {
  polygons?: PolygonProto[] | undefined;
}

export interface PolygonProto {
  /**
   * All loops are ordered counter-clockwise, and a point is defined to be
   * inside the polygon if it's enclosed by an odd number of loops.
   */
  loops?: LoopProto[] | undefined;
}

export interface LoopProto {
  points?: PointProto[] | undefined;
}

export interface PointProto {
  latE7?: number | undefined;
  lngE7?: number | undefined;
}

function createBasePolylineProto(): PolylineProto {
  return { points: [], lengthMeters: 0 };
}

export const PolylineProto = {
  encode(message: PolylineProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.points !== undefined && message.points.length !== 0) {
      for (const v of message.points) {
        PointProto.encode(v!, writer.uint32(10).fork()).ldelim();
      }
    }
    if (message.lengthMeters !== undefined && message.lengthMeters !== 0) {
      writer.uint32(17).double(message.lengthMeters);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): PolylineProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBasePolylineProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.points!.push(PointProto.decode(reader, reader.uint32()));
          continue;
        case 2:
          if (tag !== 17) {
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

  fromJSON(object: any): PolylineProto {
    return {
      points: globalThis.Array.isArray(object?.points) ? object.points.map((e: any) => PointProto.fromJSON(e)) : [],
      lengthMeters: isSet(object.lengthMeters) ? globalThis.Number(object.lengthMeters) : 0,
    };
  },

  toJSON(message: PolylineProto): unknown {
    const obj: any = {};
    if (message.points?.length) {
      obj.points = message.points.map((e) => PointProto.toJSON(e));
    }
    if (message.lengthMeters !== undefined && message.lengthMeters !== 0) {
      obj.lengthMeters = message.lengthMeters;
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<PolylineProto>, I>>(base?: I): PolylineProto {
    return PolylineProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<PolylineProto>, I>>(object: I): PolylineProto {
    const message = createBasePolylineProto();
    message.points = object.points?.map((e) => PointProto.fromPartial(e)) || [];
    message.lengthMeters = object.lengthMeters ?? 0;
    return message;
  },
};

function createBaseMultiPolygonProto(): MultiPolygonProto {
  return { polygons: [] };
}

export const MultiPolygonProto = {
  encode(message: MultiPolygonProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.polygons !== undefined && message.polygons.length !== 0) {
      for (const v of message.polygons) {
        PolygonProto.encode(v!, writer.uint32(10).fork()).ldelim();
      }
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): MultiPolygonProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseMultiPolygonProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.polygons!.push(PolygonProto.decode(reader, reader.uint32()));
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): MultiPolygonProto {
    return {
      polygons: globalThis.Array.isArray(object?.polygons)
        ? object.polygons.map((e: any) => PolygonProto.fromJSON(e))
        : [],
    };
  },

  toJSON(message: MultiPolygonProto): unknown {
    const obj: any = {};
    if (message.polygons?.length) {
      obj.polygons = message.polygons.map((e) => PolygonProto.toJSON(e));
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<MultiPolygonProto>, I>>(base?: I): MultiPolygonProto {
    return MultiPolygonProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<MultiPolygonProto>, I>>(object: I): MultiPolygonProto {
    const message = createBaseMultiPolygonProto();
    message.polygons = object.polygons?.map((e) => PolygonProto.fromPartial(e)) || [];
    return message;
  },
};

function createBasePolygonProto(): PolygonProto {
  return { loops: [] };
}

export const PolygonProto = {
  encode(message: PolygonProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.loops !== undefined && message.loops.length !== 0) {
      for (const v of message.loops) {
        LoopProto.encode(v!, writer.uint32(10).fork()).ldelim();
      }
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): PolygonProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBasePolygonProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.loops!.push(LoopProto.decode(reader, reader.uint32()));
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): PolygonProto {
    return {
      loops: globalThis.Array.isArray(object?.loops) ? object.loops.map((e: any) => LoopProto.fromJSON(e)) : [],
    };
  },

  toJSON(message: PolygonProto): unknown {
    const obj: any = {};
    if (message.loops?.length) {
      obj.loops = message.loops.map((e) => LoopProto.toJSON(e));
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<PolygonProto>, I>>(base?: I): PolygonProto {
    return PolygonProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<PolygonProto>, I>>(object: I): PolygonProto {
    const message = createBasePolygonProto();
    message.loops = object.loops?.map((e) => LoopProto.fromPartial(e)) || [];
    return message;
  },
};

function createBaseLoopProto(): LoopProto {
  return { points: [] };
}

export const LoopProto = {
  encode(message: LoopProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.points !== undefined && message.points.length !== 0) {
      for (const v of message.points) {
        PointProto.encode(v!, writer.uint32(10).fork()).ldelim();
      }
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): LoopProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseLoopProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.points!.push(PointProto.decode(reader, reader.uint32()));
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): LoopProto {
    return {
      points: globalThis.Array.isArray(object?.points) ? object.points.map((e: any) => PointProto.fromJSON(e)) : [],
    };
  },

  toJSON(message: LoopProto): unknown {
    const obj: any = {};
    if (message.points?.length) {
      obj.points = message.points.map((e) => PointProto.toJSON(e));
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<LoopProto>, I>>(base?: I): LoopProto {
    return LoopProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<LoopProto>, I>>(object: I): LoopProto {
    const message = createBaseLoopProto();
    message.points = object.points?.map((e) => PointProto.fromPartial(e)) || [];
    return message;
  },
};

function createBasePointProto(): PointProto {
  return { latE7: 0, lngE7: 0 };
}

export const PointProto = {
  encode(message: PointProto, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.latE7 !== undefined && message.latE7 !== 0) {
      writer.uint32(8).int32(message.latE7);
    }
    if (message.lngE7 !== undefined && message.lngE7 !== 0) {
      writer.uint32(16).int32(message.lngE7);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): PointProto {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBasePointProto();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 8) {
            break;
          }

          message.latE7 = reader.int32();
          continue;
        case 2:
          if (tag !== 16) {
            break;
          }

          message.lngE7 = reader.int32();
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): PointProto {
    return {
      latE7: isSet(object.latE7) ? globalThis.Number(object.latE7) : 0,
      lngE7: isSet(object.lngE7) ? globalThis.Number(object.lngE7) : 0,
    };
  },

  toJSON(message: PointProto): unknown {
    const obj: any = {};
    if (message.latE7 !== undefined && message.latE7 !== 0) {
      obj.latE7 = Math.round(message.latE7);
    }
    if (message.lngE7 !== undefined && message.lngE7 !== 0) {
      obj.lngE7 = Math.round(message.lngE7);
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<PointProto>, I>>(base?: I): PointProto {
    return PointProto.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<PointProto>, I>>(object: I): PointProto {
    const message = createBasePointProto();
    message.latE7 = object.latE7 ?? 0;
    message.lngE7 = object.lngE7 ?? 0;
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

function isSet(value: any): boolean {
  return value !== null && value !== undefined;
}
