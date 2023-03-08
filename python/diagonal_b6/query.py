from diagonal_b6 import expression

from diagonal_b6 import api_pb2

class Query(expression.Literal):

    def __init__(self, query):
        self.query = query

    def tagged(self, key, value=None):        
        return self._and(tagged(key, value))

    def type(self, t):        
        return type(t, self)

    def within(self, area):        
        return self._and(within(area))

    def union(self, query):
        return union(self, query)

    def _and(self, other):
        if self.query.WhichOneof("query") == "all":
            return Query(other)
        elif self.query.WhichOneof("query") == "intersection":
            query = api_pb2.QueryProto()
            query.CopyFrom(self.query)
            query.intersection.queries.add().CopyFrom(other)
            return Query(query)
        else:
            query = api_pb2.QueryProto()
            query.intersection.queries.add().CopyFrom(self.query)
            query.intersection.queries.add().CopyFrom(other.query)
            return Query(query)

    def to_proto(self):
        return self.query

    def to_literal_proto(self):
        l = api_pb2.LiteralNodeProto()
        l.queryValue.CopyFrom(self.query)
        return l

def from_proto(pb):
    return Query(pb)

expression.register_literal("queryValue", from_proto)

def union(*queries):
    q = api_pb2.QueryProto()
    for query in queries:
        _flattern_union(q, query.query)
    return Query(q)

def _flattern_union(dest, source):
    if source.WhichOneof("query") == "union":
        for q in source.union.queries:
            _flattern_union(dest, q)
    else:
        dest.union.queries.add().CopyFrom(source)

def tagged(key, value=None):
    q = api_pb2.QueryProto()
    if value is None:
        q.key.key = key
    else:
        q.keyValue.key = key
        q.keyValue.value = value
    return Query(q)

def within(area):
    q = api_pb2.QueryProto()
    area._fill_query(q)
    return Query(q)

def type(t, query):
    q = api_pb2.QueryProto()
    q.type.type = t
    q.query.CopyFrom(query)
    return Query(q)

def all():
    q = api_pb2.QueryProto()
    q.all.CopyFrom(api_pb2.AllQueryProto())
    return Query(q)
