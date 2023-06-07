from diagonal_b6 import api_pb2
from diagonal_b6 import expression
from diagonal_b6 import features
from diagonal_b6 import geometry
from diagonal_b6 import api_generated

class LiteralCollection(expression.Node):

    def __init__(self, node):
        self.node = node

    def to_node_proto(self):
        return self.node

def feature_ids(values):
    node = api_pb2.NodeProto()
    node.literal.collectionValue.CopyFrom(api_pb2.CollectionProto())
    for (i, v) in enumerate(values):
        node.literal.collectionValue.keys.add().intValue = i
        node.literal.collectionValue.values.add().CopyFrom(v.to_literal_proto())
    return api_generated.IntFeatureIDCollectionResult(LiteralCollection(node))

def lls(values):
    node = api_pb2.NodeProto()
    node.literal.collectionValue.CopyFrom(api_pb2.CollectionProto())
    for (i, ll) in enumerate(values):
        node.literal.collectionValue.keys.add().intValue = i
        node.literal.collectionValue.values.add().pointValue.CopyFrom(geometry.to_point_proto(ll))
    return api_generated.IntFeatureIDCollectionResult(LiteralCollection(node))

def strings(values):
    node = api_pb2.NodeProto()
    node.literal.collectionValue.CopyFrom(api_pb2.CollectionProto())
    for (i, v) in enumerate(values):
        node.literal.collectionValue.keys.add().intValue = i
        node.literal.collectionValue.values.add().stringValue = v
    return api_generated.IntStringCollectionResult(LiteralCollection(node))
