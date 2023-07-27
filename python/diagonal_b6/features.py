from diagonal_b6 import expression

from diagonal_b6 import api_pb2
from diagonal_b6 import features_pb2

FEATURE_TYPE_POINT = features_pb2.FeatureType.FeatureTypePoint
FEATURE_TYPE_PATH = features_pb2.FeatureType.FeatureTypePath
FEATURE_TYPE_AREA = features_pb2.FeatureType.FeatureTypeArea
FEATURE_TYPE_RELATION = features_pb2.FeatureType.FeatureTypeRelation

NAMESPACE_OSM_NODE = "openstreetmap.org/node"
NAMESPACE_OSM_WAY = "openstreetmap.org/way"
NAMESPACE_OSM_RELATION = "openstreetmap.org/relation"
NAMESPACE_UK_ONS_BOUNDARIES = "statistics.gov.uk/datasets/regions"
NAMESPACE_DIAGONAL_ACCESS_POINTS = "diagonal.works/ns/access-point"

class FeatureID(expression.Literal):

    def __init__(self, type, namespace, value):
        self.type = type
        self.namespace = namespace
        self.value = value

    def is_point(self):
        return self.type == FEATURE_TYPE_POINT

    def is_path(self):
        return self.type == FEATURE_TYPE_PATH
    
    def is_area(self):
        return self.type == FEATURE_TYPE_AREA

    def is_relation(self):
        return self.type == FEATURE_TYPE_RELATION

    def to_literal_proto(self):
        l = api_pb2.LiteralNodeProto()
        l.featureIDValue.type = self.type
        l.featureIDValue.namespace = self.namespace
        l.featureIDValue.value = self.value
        return l

    def __str__(self):
        type = features_pb2.FeatureType.Name(self.type).replace("FeatureType", "").lower()
        return "/%s/%s/%d" % (type, self.namespace, self.value)

    def __repr__(self):
        return str(self)

    def __eq__(self, other):
        return self.type == other.type and self.namespace == other.namespace and self.value == other.value

    def __hash__(self):
        return hash(self.type) ^ hash(self.namespace) ^ hash(self.value)

    def _fill_query(self, query):
        query.spatial.area.id.type = self.type
        query.spatial.area.id.namespace = self.namespace
        query.spatial.area.id.value = self.value

class Feature(expression.Node):

    def is_point(self):
        return self.id.is_point()

    def is_path(self):
        return self.id.is_path()
    
    def is_area(self):
        return self.id.is_area()

    def is_relation(self):
        return self.id.is_relation()

    def get(self, key):
        for tag in self._pb.tags:
            if tag.key == key:
                return (key, tag.value)
        return (None, None)

    def get_string(self, key):
        _, value = self.get(key)
        if value is not None:
            return value
        return ""

    def get_int(self, key):
        _, value = self.get(key)
        if value is not None:
            try:
                return int(value)
            except:
                pass
        return 0

    def get_float(self, key):
        _, value = self.get(key)
        if value is not None:
            try:
                return float(value)
            except:
                pass
        return 0

    def all_tags(self):
        return [(tag.key, tag.value) for tag in self._pb.tags]

    def to_node_proto(self):
        node = api_pb2.NodeProto()
        node.call.function.symbol = "find-feature"
        node.call.args.add().CopyFrom(self.id.to_node_proto())
        return node

    def __str__(self):
        type = features_pb2.FeatureType.Name(self.id.type).replace("FeatureType", "").title()
        return "<%s %s>" % (type, self.id)

    def _fill_query(self, query):
        return self.id._fill_query(query)

class PointFeature(Feature):

    def __init__(self, p):
        self.id = from_id_proto(p.point.id)
        self._pb = p.point

class PathFeature(Feature):

    def __init__(self, pb):
        self.id = from_id_proto(pb.path.id)
        self._pb = pb.path

    def length_meters(self):
        return self._pb.lengthMeters

class AreaFeature(Feature):

    def __init__(self, pb):
        self.id = from_id_proto(pb.area.id)
        self._pb = pb.area

class RelationFeature(Feature):

    def __init__(self, pb):
        self.id = from_id_proto(pb.relation.id)
        self._pb = pb.relation

    def members(self):
        return [_from_relation_member_proto(m) for m in self._pb.members]

class RelationMember:

    def __init__(self, id, role=None):
        self.id = id
        self.role = role

    def is_point(self):
        return self.id.is_point()

    def is_path(self):
        return self.id.is_path()
    
    def is_area(self):
        return self.id.is_area()

    def is_relation(self):
        return self.id.is_relation()

    def __str__(self):
        return "<RelationMember %s" % (str(self.id),)

def from_id_proto(p):
    return FeatureID(p.type, p.namespace, p.value)

expression.register_literal_from_proto("featureIDValue", from_id_proto)

def from_applied_change_proto(change):
    applied = {}
    for i in range(0, len(change.original)):
        applied[from_id_proto(change.original[i])] = from_id_proto(change.modified[i])
    return applied

expression.register_literal_from_proto("appliedChangeValue", from_applied_change_proto)

def id_to_proto(id):
    pb = features_pb2.FeatureIDProto()
    pb.type = id.type
    pb.namespace = id.namespace
    pb.value = id.value
    return pb

def _from_point_proto(p):
    return PointFeature(p)

def _from_path_proto(p):
    return PathFeature(p)

def _from_area_proto(p):
    return AreaFeature(p)

def _from_relation_proto(p):
    return RelationFeature(p)

def _from_relation_member_proto(p):
    return RelationMember(from_id_proto(p.id), p.role)

def from_proto(p):
    oneof = p.WhichOneof("feature")
    if oneof == "point":
        return _from_point_proto(p)
    elif oneof == "path":
        return _from_path_proto(p)
    elif oneof == "area":
        return _from_area_proto(p)
    elif oneof == "relation":
        return _from_relation_proto(p)
    elif oneof == None:
        return None
    raise Exception("Unexpected feature %s" % (p,))

expression.register_literal_from_proto("featureValue", from_proto)

def osm_node_id(id):
    return FeatureID(FEATURE_TYPE_POINT, NAMESPACE_OSM_NODE, id)

def osm_way_id(id):
    return FeatureID(FEATURE_TYPE_PATH, NAMESPACE_OSM_WAY, id)

def osm_way_area_id(id):
    return FeatureID(FEATURE_TYPE_AREA, NAMESPACE_OSM_WAY, id)

def osm_relation_area_id(id):
    return FeatureID(FEATURE_TYPE_AREA, NAMESPACE_OSM_RELATION, id)

def osm_relation_id(id):
    return FeatureID(FEATURE_TYPE_RELATION, NAMESPACE_OSM_RELATION, id)

def uk_ons_boundary_id(id, year=2011):
    # See GBONS2011IDStrategy in src/diagonal.works/b6/ingest/gdal/source.go
    if len(id) != 9:
        raise "Expected a string of 9 characters"
    codeBits = ord(id[0]) << 40
    yearBits = (year-1900) << 32
    return FeatureID(FEATURE_TYPE_AREA, NAMESPACE_UK_ONS_BOUNDARIES, codeBits|yearBits|int(id[1:]))