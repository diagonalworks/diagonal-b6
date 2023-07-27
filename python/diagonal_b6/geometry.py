from diagonal_b6 import expression

from diagonal_b6 import api_pb2
from diagonal_b6 import geometry_pb2

def from_point_proto(p):
    return (float(p.lat_e7) / 1e7, float(p.lng_e7) / 1e7)

expression.register_literal_from_proto("pointValue", from_point_proto)

def to_point_proto(ll):
    p = geometry_pb2.PointProto()
    p.lat_e7 = int(ll[0] * 1e7)
    p.lng_e7 = int(ll[1] * 1e7)
    return p

def from_polyline_proto(p):
    return Polyline(p)

expression.register_literal_from_proto("pathValue", from_polyline_proto)

def from_multipolygon_proto(p):
    return MultiPolygon(p)

expression.register_literal_from_proto("areaValue", from_multipolygon_proto)

class Polyline(expression.Node):

    def __init__(self, pb):
        self._pb = pb

    def length_meters(self):
        return self._pb.length_meters

    def to_node_proto(self):
        n = api_pb2.NodeProto()
        n.literal.pathValue.CopyFrom(self._pb)
        return n

    def __str__(self):
        return str(self._pb)

class Polygon(expression.Node):

    def __init__(self, pb):
        self._pb = pb

    def to_node_proto(self):
        n = api_pb2.NodeProto()
        n.literal.areaValue.polygons.add().CopyFrom(self._pb)
        return n

    def __str__(self):
        return str(self._pb)

class MultiPolygon(expression.Node):

    def __init__(self, pb):
        self._pb = pb

    def to_node_proto(self):
        n = api_pb2.NodeProto()
        n.literal.areaValue.CopyFrom(self._pb)
        return n

    def __str__(self):
        return str(self._pb)

def wkt(s):
    if s.startswith("POLYGON "):
        return _wkt_polygon(s)
    elif s.startswith("MULTIPOLYGON "):
        return _wkt_multipolygon(s)
    raise ValueError("Can't parse WKT %s" % (repr(s),))

def _wkt_polygon(s):
    header = "POLYGON "
    polygon_wkt = s[len(header):]
    polygon = geometry_pb2.PolygonProto()
    _fill_polygon_from_wkt(polygon, polygon_wkt)
    return Polygon(polygon)

def _wkt_multipolygon(s):
    header = "MULTIPOLYGON "
    multipolygon_wkt = s[len(header):]
    multipolygon = geometry_pb2.MultiPolygonProto()
    for group in _parse_wkt_groups(multipolygon_wkt):
        _fill_polygon_from_wkt(multipolygon.polygons.add(), group)
    return MultiPolygon(multipolygon)

def _fill_polygon_from_wkt(polygon, wkt):
    for group in _parse_wkt_groups(wkt):
        _fill_loop_from_wkt(polygon.loops.add(), group)

def _fill_loop_from_wkt(loop, wkt):
    for (lat, lng) in _parse_wkt_points(wkt):
        point = loop.points.add()
        point.lat_e7 = int(lat * 1e7)
        point.lng_e7 = int(lng * 1e7)

def _parse_wkt_groups(s):
    if s[0] != "(" and s[len(s)-1] != ")":
        raise ValueError("Expected a bracketed WKT group: %s" % (repr(s),))
    groups = []
    i = 1
    while i < len(s) - 1:        
        if s[i] == "(":
            depth = 1
            start = i
            i += 1
            while True:
                if i >= len(s) - 1:
                    raise ValueError("Expected closing bracket in group %s" % (repr(s),))
                if s[i] == ")":
                    depth -= 1
                    if depth == 0:
                        i += 1
                        break
                if s[i] == "(":
                    depth += 1
                i += 1
            groups.append(s[start:i])
        elif s[i] == " " or s[i] == ",":
            i += 1
        else:
            raise ValueError("Invalid character %s in group %s" % (repr(s[i]), repr(s)))
    return groups

def _parse_wkt_points(s):
    if s[0] != "(" and s[len(s)-1] != ")":
        raise ValueError("Expected a bracketed WKT group of points %s" % (repr(s),))
    parsed = []
    for coordinates in s[1:-1].split(","):
        cs = [float(c) for c in coordinates.split(" ") if len(c) > 0]
        if len(cs) != 2:
            raise ValueError("Expected a coordinate with 2 values, found %s" % (repr(coordinates),))
        parsed.append((cs[1], cs[0]))
    return parsed

