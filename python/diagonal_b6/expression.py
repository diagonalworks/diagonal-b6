import gzip
import inspect
import json

from diagonal_b6 import api_pb2

class Node:

    def __init__(self):
        self.name = None

    def set_name(self, name):
        self.name = name

    def to_node_proto(self):
        raise NotImplementedError()

class Literal(Node):

    def __init__(self):
        Node.__init__(self)

    def to_node_proto(self):
        n = api_pb2.NodeProto()
        if self.name is not None:
            n.name = self.name
        n.literal.CopyFrom(self.to_literal_proto())
        return n

    def to_literal_proto(self):
        raise NotImplementedError()

class LiteralInt(Literal):

    def __init__(self, v):
        Literal.__init__(self)
        self.v = v

    def to_literal_proto(self):
        n = api_pb2.LiteralNodeProto()
        n.intValue = self.v
        return n

class LiteralFloat(Literal):

    def __init__(self, v):
        Literal.__init__(self)
        self.v = v

    def to_literal_proto(self):
        n = api_pb2.LiteralNodeProto()
        n.floatValue = self.v
        return n

class LiteralString(Literal):

    def __init__(self, v):
        Literal.__init__(self)
        self.v = v

    def to_literal_proto(self):
        n = api_pb2.LiteralNodeProto()
        n.stringValue = self.v
        return n

class Symbol(Node):

    def __init__(self, symbol):
        Node.__init__(self)
        self.symbol = symbol

    def to_node_proto(self):
        n = api_pb2.NodeProto()
        if self.name is not None:
            n.name = self.name
        n.symbol = self.symbol
        return n

class Call(Node):

    def __init__(self, function, args):
        Node.__init__(self)
        self.function = function
        self.args = args

    def to_node_proto(self):
        n = api_pb2.NodeProto()
        if self.name is not None:
            n.name = self.name
        n.call.function.CopyFrom(self.function.to_node_proto())
        for arg in self.args:
            node = to_node(arg)
            n.call.args.add().CopyFrom(node.to_node_proto())
        return n

class Lambda(Node):

    def __init__(self, function, arg_types):
        Node.__init__(self)
        self.function = function
        self.arg_types = arg_types

    def to_callable(self):
        return self

    def with_arg_types(self, arg_types):
        return Lambda(self.function, arg_types)

    def to_node_proto(self):
        n = api_pb2.NodeProto()
        if self.name is not None:
            n.name = self.name
        args = ["_%s_%d" % (name, id(n)) for name in inspect.signature(self.function).parameters]
        for arg in args:
            n.lambda_.args.append(arg)
        node = to_node(self.function(*[type(Symbol(name)) for (name, type) in zip(args, self.arg_types)]))
        n.lambda_.node.CopyFrom(node.to_node_proto())
        return n

    def __call__(self, *args):
        return self.function(*args)

class QueryConversionTraits:

    def to_callable(self):
        return self

    def with_arg_types(self, arg_types):
        return self

def to_lambda(f):
    if getattr(f, "to_callable", None) is not None:
        return f.to_callable()
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

    def set_name(self, name):
        self.node.set_name(name)

    def to_node_proto(self):
        return self.node.to_node_proto()

    def _values(cls):
        return Result

class ListCollectionResult(Result):

    @classmethod
    def _values(cls):
        return Result

class DictCollectionResult(Result):

    @classmethod
    def _values(cls):
        return Result

class Placeholder(Node):

    def to_node_proto(self):
        raise Exception("Placeholder can't be converted to a proto")

def to_node(v):
    if isinstance(v, Node):
        return v
    elif isinstance(v, int):
       return LiteralInt(v)
    elif isinstance(v, float):
       return LiteralFloat(v)
    elif isinstance(v, str):
       return LiteralString(v)
    elif isinstance(v, list):
        pairs = [Call(Symbol("pair"), [LiteralInt(i), to_node(vv)]) for (i, vv) in enumerate(v)]
        return ListCollectionResult(Call(Symbol("collection"), pairs))
    elif isinstance(v, dict):
        pairs = [Call(Symbol("pair"), [to_node(k), to_node(vv)]) for (k, vv) in v.items()]
        return DictCollectionResult(Call(Symbol("collection"), pairs))
    elif callable(v):
        return to_lambda(v)
    else:
        raise ValueError("Can't convert %s to proto" % (v,))

def from_node_proto(n):
    if n.WhichOneof("node") != "literal":
        raise Exception("Can't convert node type %s to a value" % (n.WhichOneof("node"),))
    return from_literal_node_proto(n.literal)

_literal_from_proto = {}

def register_literal_from_proto(oneof, from_proto):
    _literal_from_proto[oneof] = from_proto

register_literal_from_proto("nilValue", lambda v: None)
register_literal_from_proto("intValue", lambda v: v)
register_literal_from_proto("floatValue", lambda v: v)
register_literal_from_proto("stringValue", lambda v: v)
register_literal_from_proto("boolValue", lambda v: v)
register_literal_from_proto("tagValue", lambda v: (v.key, v.value))
register_literal_from_proto("geoJSONValue", lambda v: json.loads(gzip.decompress(v)))

def from_literal_node_proto(literal):
    oneof = literal.WhichOneof("value")
    from_proto = _literal_from_proto.get(oneof, None)
    if from_proto is not None:
        return from_proto(getattr(literal, oneof))
    else:
        raise Exception("can't convert %s to value" % (oneof,))

def from_collection_proto(collection):
    return [(from_literal_node_proto(collection.keys[i]), from_literal_node_proto(collection.values[i]))
            for i in range(0, len(collection.keys))]

register_literal_from_proto("collectionValue", from_collection_proto)

_builtin_results = {}

def register_builtin_result(type, result):
    _builtin_results[type] = result

def collection_result(result):
    if type(result) in _builtin_results:
        return _builtin_results[type(result)]._collection()
    return result._collection()

def _map(collection, f):
    collection = to_node(collection)
    result = f(collection._values()(Placeholder()))
    return collection_result(result)(Call(Symbol("map"), [collection, to_lambda(f).with_arg_types((collection._values(),))]))

def _filter(collection, f):
    collection = to_node(collection)
    return type(collection)(Call(Symbol("filter"), [collection, to_lambda(f).with_arg_types((collection._values(),))]))

def _name(expression, name):
    expression = to_node(expression)
    expression.set_name(name)
    return expression


