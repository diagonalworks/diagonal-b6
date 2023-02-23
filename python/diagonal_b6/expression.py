import gzip
import inspect
import json

from diagonal_b6 import api_pb2

class Node:

    def to_node_proto(self):
        raise NotImplementedError()

class Literal(Node):

    def to_node_proto(self):
        node = api_pb2.NodeProto()
        node.literal.CopyFrom(self.to_literal_proto())
        return node

    def to_literal_proto(self):
        raise NotImplementedError()

class Symbol(Node):

    def __init__(self, name):
        self.name = name

    def to_node_proto(self):
        node = api_pb2.NodeProto()
        node.symbol = self.name
        return node

class Call(Node):

    def __init__(self, function, args):
        self.function = function
        self.args = args

    def to_node_proto(self):
        node = api_pb2.NodeProto()
        node.call.function.CopyFrom(self.function.to_node_proto())
        for arg in self.args:
            node.call.args.add().CopyFrom(to_node_proto(arg))
        return node

class Lambda(Node):

    def __init__(self, function, arg_types):
        self.function = function
        self.arg_types = arg_types

    def with_arg_types(self, arg_types):
        return Lambda(self.function, arg_types)

    def to_node_proto(self):
        n = api_pb2.NodeProto()
        args = ["_%s_%d" % (name, id(n)) for name in inspect.signature(self.function).parameters]
        for arg in args:
            n.lambda_.args.append(arg)
        n.lambda_.node.CopyFrom(to_node_proto(self.function(*[type(Symbol(name)) for (name, type) in zip(args, self.arg_types)])))
        return n

    def __call__(self, *args):
        return self.function(*args)

def to_lambda(f):
    if isinstance(f, Lambda):
        return f
    arg_types = []
    for p in inspect.signature(f).parameters.values():
        if p.annotation != inspect.Signature.empty:
            arg_types.append(p.annotation)
        else:
            arg_types.append(Result)
    return Lambda(f, arg_types)

class Result(Node):

    def __init__(self, node):
        self.node = node

    def to_node_proto(self):
        return self.node.to_node_proto()

class Placeholder(Node):

    def to_node_proto(self):
        raise Exception("Placeholder can't be converted to a proto")

def to_node_proto(v):
    if isinstance(v, Node):
        return v.to_node_proto()
    elif isinstance(v, int):
       n = api_pb2.NodeProto()
       n.literal.intValue = v
       return n
    elif isinstance(v, float):
        n = api_pb2.NodeProto()
        n.literal.floatValue = v
        return n
    elif isinstance(v, str):
        n = api_pb2.NodeProto()
        n.literal.stringValue = v
        return n
    elif callable(v):
        return to_lambda(v).to_node_proto()
    else:
        raise ValueError("Can't convert %s to proto" % (v,))

def from_node_proto(n):
    if n.WhichOneof("node") != "literal":
        raise Exception("Can't convert node type %s to a value" % (n.WhichOneof("node"),))
    return from_literal_node_proto(n.literal)

_literals = {}

def register_literal(oneof, from_proto):
    _literals[oneof] = from_proto

register_literal("nilValue", lambda v: None)
register_literal("intValue", lambda v: v)
register_literal("floatValue", lambda v: v)
register_literal("stringValue", lambda v: v)
register_literal("geoJSONValue", lambda v: json.loads(gzip.decompress(v)))

def from_literal_node_proto(literal):
    oneof = literal.WhichOneof("value")
    from_proto = _literals.get(oneof, None)
    if from_proto is not None:
        return from_proto(getattr(literal, oneof))
    else:
        raise Exception("can't convert %s to value" % (oneof,))

def from_collection_proto(collection):
    return [(from_literal_node_proto(collection.keys[i]), from_literal_node_proto(collection.values[i]))
            for i in range(0, len(collection.keys))]

register_literal("collectionValue", from_collection_proto)

def _map(collection, f):
    result = f(collection._values()(Placeholder()))
    return result._collection()(Call(Symbol("map"), [collection, to_lambda(f).with_arg_types((collection._values(),))]))

def _filter(collection, f):
    return type(collection)(Call(Symbol("filter"), [collection, to_lambda(f).with_arg_types((collection._values(),))]))


