#!/usr/bin/env python3

import argparse
import json
from operator import ge
import os
import s2sphere
import subprocess
import time
import urllib.request
import unittest

import diagonal_b6 as b6

# TODO: Harmonise these IDs with those in src/diagonal.works/diagonal/test/world.go
COAL_DROPS_YARD_WEST_BUILDING_ID = 222021572
COAL_DROPS_YARD_ENCLOSURE_ID = 500008118
JUBILEE_GREENWAY_ID = 380856
STABLE_STREET_BRIDGE_ID = 140633010
STABLE_STREET_BRIDGE_NORTH_END_ID = 1447052073
STABLE_STREET_BRIDGE_SOUTH_END_ID = 1540349979
VERMUTERIA_NODE_ID = 6082053666
GRANARY_SQUARE_WAY_ID = 222021571
LIGHTERMAN_WAY_ID = 427900370

BUILDINGS_IN_GRANARY_SQUARE = 13
HIGHWAYS_IN_GRANARY_SQUARE = 117
BIKE_PARKING_IN_GRANARY_SQUARE = 11
FOUNTAINS_IN_GRANARY_SQUARE = 4 # Within the pedestrian square itself defined by WKT below not the entire area
STABLE_STREET_BRIDGE_NORTH_END_DEGREE = 7 # After it has been connected

GRANARY_SQUARE_POLYGON_WKT="POLYGON ((-0.1260475 51.5357019,-0.1261001 51.5355674,-0.1261596 51.5354153,-0.1262097 51.535287,-0.1259034 51.5352365,-0.1259462 51.5351347,-0.1255806 51.5350765,-0.1255202 51.5350667,-0.1255004 51.5350372,-0.1254536 51.5349963,-0.1254346 51.5350013,-0.1252611 51.535049,-0.125219 51.5350629,-0.124904 51.5350121,-0.1247915 51.5350326,-0.124709 51.5350541,-0.1247491 51.5351308,-0.1247727 51.5351758,-0.1246766 51.5353808,-0.1246363 51.5354737,-0.125082 51.5355458,-0.1259754 51.5356902,-0.1260475 51.5357019))"
GRANARY_SQUARE_MULTIPOLYGON_WKT="MULTIPOLYGON (((-0.1260475 51.5357019,-0.1261001 51.5355674,-0.1261596 51.5354153,-0.1262097 51.535287,-0.1259034 51.5352365,-0.1259462 51.5351347,-0.1255806 51.5350765,-0.1255202 51.5350667,-0.1255004 51.5350372,-0.1254536 51.5349963,-0.1254346 51.5350013,-0.1252611 51.535049,-0.125219 51.5350629,-0.124904 51.5350121,-0.1247915 51.5350326,-0.124709 51.5350541,-0.1247491 51.5351308,-0.1247727 51.5351758,-0.1246766 51.5353808,-0.1246363 51.5354737,-0.125082 51.5355458,-0.1259754 51.5356902,-0.1260475 51.5357019)))"

EARTH_RADIUS_METERS = 6371.01 * 1000.0

def angle_to_meters(angle):
	return angle.radians * EARTH_RADIUS_METERS

class B6Test(unittest.TestCase):

    def __init__(self, name, connection):
        unittest.TestCase.__init__(self, name)
        self.connection = connection

    def test_find_areas(self):
        names = [building.get_string("name") for (id, building) in self.connection(b6.find_areas(b6.keyed("#building")))]
        self.assertEqual(len(names), BUILDINGS_IN_GRANARY_SQUARE)

    def test_find_point_by_id(self):        
        area = self.connection(b6.find_point(b6.osm_node_id(STABLE_STREET_BRIDGE_SOUTH_END_ID)))
        self.assertEqual(area.id.value, STABLE_STREET_BRIDGE_SOUTH_END_ID)

    def test_find_area_by_id(self):
        area = self.connection(b6.find_area(b6.osm_way_area_id(COAL_DROPS_YARD_WEST_BUILDING_ID)))
        self.assertEqual(area.id.value, COAL_DROPS_YARD_WEST_BUILDING_ID)

    def test_find_non_existant_id(self):
        self.assertEqual(self.connection(b6.find_point(b6.osm_node_id(42))), None)

    def test_find_area_by_wrong_id_type(self):
        with self.assertRaises(Exception): # TODO: Make more specific
            self.connection(b6.find_area(b6.osm_node_id(STABLE_STREET_BRIDGE_SOUTH_END_ID)))

    def test_uk_ons_boundary_id(self):
        self.assertEqual(b6.uk_ons_boundary_id("E01000953").value, 76343044687353)

    def test_relation_members(self):
        greenway = self.connection(b6.find_relation(b6.osm_relation_id(JUBILEE_GREENWAY_ID)))
        paths = ([str(m) for m in greenway.members() if m.is_path()])
        self.assertGreater(len(paths), 10)
        self.assertLess(len(paths), 800)

    def test_or_query(self):
        names = [amenity.get_string("name") for (id, amenity) in self.connection(b6.find(b6.tagged("#amenity", "restaurant").or_(b6.tagged("#amenity", "cafe"))))]
        self.assertIn("Le Cafe Alain Ducasse", names)

    def test_point_degree(self):
        self.assertEqual(self.connection(b6.find_point(b6.osm_node_id(STABLE_STREET_BRIDGE_NORTH_END_ID)).degree()), STABLE_STREET_BRIDGE_NORTH_END_DEGREE)
        degrees = [degree for (id, degree) in self.connection(b6.find_points(b6.within_cap(b6.ll(51.535241, -0.124364), 100)).degree())]
        for d in degrees:
            self.assertGreaterEqual(d, 0)
            self.assertLess(d, 10)

    def test_path_lengths(self):
        lengths = [length for (id, length) in self.connection(b6.find_paths(b6.keyed("#highway")).length())]
        self.assertGreater(len(lengths), 0)
        for l in lengths:
            self.assertGreater(l, 0)
            self.assertLess(l, 1000)

    def test_relation_names(self):
        names = [r.get_string("name") for (id, r) in self.connection(b6.find_relations(b6.keyed("#route")))]
        self.assertIn("Jubilee Greenway", names)

    def test_reachable_areas_from_point(self):
        expression = b6.find_point(b6.osm_node_id(STABLE_STREET_BRIDGE_SOUTH_END_ID)).reachable("walk", 200.0, b6.keyed("#amenity")).get_string("name")
        names = set([name for (_, name) in self.connection(expression)])
        self.assertIn("The Lighterman", names)

    def test_reachable_with_distance(self):
        small = b6.find_point(b6.osm_node_id(STABLE_STREET_BRIDGE_SOUTH_END_ID)).reachable("walk", 100.0, b6.keyed("#amenity")).count()
        large = b6.find_point(b6.osm_node_id(STABLE_STREET_BRIDGE_SOUTH_END_ID)).reachable("walk", 200.0, b6.keyed("#amenity")).count()
        self.assertGreater(self.connection(large), self.connection(small))

    def test_paths_to_reach(self):
        expression = b6.find_point(b6.osm_node_id(STABLE_STREET_BRIDGE_SOUTH_END_ID)).paths_to_reach("walk", 200.0, b6.keyed("#amenity"))
        paths = list(self.connection(expression))
        self.assertGreaterEqual(len(paths), 4)
        for path, count in paths:
            self.assertGreaterEqual(count, 1)
            self.assertLess(count, 100)

    def test_closest_from_point(self):
        expression = b6.find_point(b6.osm_node_id(STABLE_STREET_BRIDGE_SOUTH_END_ID)).closest("walk", 1000.0, b6.tagged("#amenity", "pub"))
        pub = self.connection(expression)
        self.assertEqual("The Lighterman", pub.get_string("name"))

    def test_closest_from_point_distance(self):
        expression = b6.find_point(b6.osm_node_id(STABLE_STREET_BRIDGE_SOUTH_END_ID)).closest_distance("walk", 1000.0, b6.tagged("#amenity", "pub"))
        distance = self.connection(expression)
        self.assertGreater(distance, 100.0)
        self.assertLess(distance, 105.0)

    def test_closest_from_area(self):
        expression = b6.find_area(b6.osm_way_area_id(COAL_DROPS_YARD_WEST_BUILDING_ID)).closest("walk", 1000.0, b6.tagged("#amenity", "pub"))
        pub = self.connection(expression)
        self.assertEqual("The Lighterman", pub.get_string("name"))

    def test_closest_from_point_non_existant(self):
        expression = b6.find_point(b6.osm_node_id(STABLE_STREET_BRIDGE_SOUTH_END_ID)).closest("walk", 1000.0, b6.tagged("#amenity", "nonexistant"))
        self.assertEqual(None, self.connection(expression))

    def test_containing_areas_from_point(self):
        expression = b6.find_point(b6.osm_node_id(STABLE_STREET_BRIDGE_SOUTH_END_ID)).reachable_points("walk", 1000.0, b6.all()).containing_areas(b6.keyed("#shop")).get_string("name")
        names = set([name for (_, name) in self.connection(expression)])
        self.assertIn("Coal Drops Yard", names)

    def test_containing_areas_from_area(self):
        areas = self.connection(b6.find_area(b6.osm_way_area_id(COAL_DROPS_YARD_WEST_BUILDING_ID)).reachable_points("walk", 1000.0, b6.all()).containing_areas(b6.all()))
        self.assertGreater(len(areas), 10)

    def test_count_features(self):
        self.assertEqual(self.connection(b6.find_points(b6.tagged("#amenity", "bicycle_parking")).count()), BIKE_PARKING_IN_GRANARY_SQUARE)
        self.assertEqual(self.connection(b6.find_paths(b6.keyed("#highway")).count()), HIGHWAYS_IN_GRANARY_SQUARE)
        self.assertEqual(self.connection(b6.find_areas(b6.keyed("#building")).count()), BUILDINGS_IN_GRANARY_SQUARE)

    def test_divide_count_features(self):
        self.assertAlmostEqual(self.connection(b6.find_points(b6.tagged("#amenity", "bicycle_parking")).count().divide(10.0)), BIKE_PARKING_IN_GRANARY_SQUARE / 10.0)

    def test_filter(self):
        filtered = self.connection(b6.find_areas(b6.keyed("#amenity")).filter(lambda a: b6.has_key(a, "addr:postcode")))
        self.assertGreater(len(filtered), 0)
        for (_, feature) in filtered:
            self.assertNotEqual(feature.get_string("addr:postcode"), "")

    def test_to_geojson_collection(self):
        geojson = self.connection(b6.to_geojson_collection(b6.find_areas(b6.keyed("#building"))))
        self.assertGreater(len(geojson["features"]), 4)
        for feature in geojson["features"]:
            self.assertIn("#building", feature["properties"])

    def test_to_geojson_with_feature(self):
        closest = b6.find_point(b6.osm_node_id(STABLE_STREET_BRIDGE_SOUTH_END_ID)).closest("walk", 1000.0, b6.tagged("#amenity", "pub"))
        geojson = self.connection(b6.to_geojson(closest))
        self.assertEqual(geojson["type"], "Feature")

    def test_to_geojson_with_missing_feature(self):
        geojson = self.connection(b6.to_geojson(b6.find_point(b6.osm_node_id(1))))
        self.assertEqual(len(geojson["features"]), 0)

    def test_search_within_wkt_polygon(self):
        granarySquare = b6.wkt(GRANARY_SQUARE_POLYGON_WKT)
        self.assertEqual(self.connection(b6.find_areas(b6.tagged("#amenity", "fountain").and_(b6.intersecting(granarySquare))).count()), FOUNTAINS_IN_GRANARY_SQUARE)

    def test_search_within_wkt_multipolygon(self):
        granarySquare = b6.wkt(GRANARY_SQUARE_MULTIPOLYGON_WKT)
        self.assertEqual(self.connection(b6.find_areas(b6.tagged("#amenity", "fountain").and_(b6.intersecting(granarySquare))).count()), FOUNTAINS_IN_GRANARY_SQUARE)

    def test_add_tags(self):
        applied = self.connection(b6.add_tags(b6.find_areas(b6.keyed("#building")).map(lambda building: b6.tag("diagonal:colour", building.get_string("building:levels")))))
        self.assertEqual(len(applied), BUILDINGS_IN_GRANARY_SQUARE)

    def test_add_tags_with_filter(self):
        applied = self.connection(b6.add_tags(b6.find_paths(b6.tagged("#highway", "footway")).filter(lambda h: b6.has_key(h, "bicycle")).map(lambda h: b6.tag("#bicycle", h.get_string("bicycle")))))
        self.assertGreater(len(applied), 0)
        self.assertEqual(self.connection(b6.find_paths(b6.keyed("#bicycle")).count()), len(applied))

    def test_search_for_newly_added_tag(self):
        reachable = b6.find_point(b6.osm_node_id(STABLE_STREET_BRIDGE_SOUTH_END_ID)).reachable("walk", 1000.0, b6.keyed("#amenity"))
        modified = self.connection(b6.add_tags(reachable.map(lambda building: b6.tag("#reachable", "yes"))))
        self.assertGreater(len(modified), 1)
        self.assertEqual(self.connection(b6.find(b6.keyed("#reachable")).count()), len(modified))

    def test_sample_points_along_path(self):
        count = len(self.connection(b6.find_path(b6.osm_way_id(STABLE_STREET_BRIDGE_ID)).sample_points(1.0)))
        self.assertGreater(count, 20)
        self.assertLess(count, 40)

    def test_sample_points_along_paths(self):
        count = 0
        center = s2sphere.LatLng.from_degrees(51.53539, -0.12537)
        for _, ll in self.connection(b6.find_paths(b6.keyed("#highway")).sample_points_along_paths(20.0)):
            d = s2sphere.LatLng.from_degrees(ll[0], ll[1]).get_distance(center)
            self.assertLess(angle_to_meters(d), 500.0)
            count += 1
        self.assertGreater(count, 300)
        self.assertLess(count, 350)

    def test_sightline(self):
        a1 = self.connection(b6.ll(51.53557, -0.12585).sightline(250.0).area())
        a2 = self.connection(b6.ll(51.53557, -0.12585).cap_polygon(250.0).area())
        self.assertGreater(a1/a2, 0.20)
        self.assertLess(a1/a2, 0.30)

    def test_sightline_geojson(self):
        geojson = self.connection(b6.to_geojson(b6.ll(51.53557, -0.12585).sightline(250.0)))
        self.assertEqual(geojson["type"], "Feature")
        self.assertEqual(geojson["geometry"]["type"], "Polygon")
        bridge = b6.osm_way_id(STABLE_STREET_BRIDGE_ID)
        geojson = self.connection(b6.to_geojson_collection(b6.find_path(bridge).sample_points(5.0).map(lambda point: b6.sightline(point, 250.0))))
        self.assertGreater(len(geojson["features"]), 5)
        self.assertLess(len(geojson["features"]), 10)

    def test_s2_points(self):
        count = len(self.connection(b6.find_area(b6.osm_way_area_id(GRANARY_SQUARE_WAY_ID)).s2_points(21, 21)))
        self.assertGreater(count, 400)
        self.assertLess(count, 500)

    def test_rectangle_polygon(self):
        area = self.connection(b6.rectangle_polygon(b6.ll(51.5146, -0.1140), b6.ll(51.5124, -0.0951)).area())
        self.assertGreater(area, 300000)
        self.assertLess(area, 400000)

    def test_s2_grid(self):
        topLeft = (51.5146, -0.1140)
        bottomRight = (51.5124, -0.0951)
        grid = self.connection(b6.rectangle_polygon(b6.ll(*topLeft), b6.ll(*bottomRight)).s2_grid(21))
        rect = s2sphere.LatLngRect.from_point_pair(s2sphere.LatLng.from_degrees(*topLeft), s2sphere.LatLng.from_degrees(*bottomRight))
        for token, _ in grid:
            cell = s2sphere.Cell(s2sphere.CellId.from_token(token))
            self.assertEqual(cell.level(), 21)
            self.assertTrue(rect.intersects(cell.get_rect_bound()))

    def test_s2_covering(self):
        topLeft = (51.5146, -0.1140)
        bottomRight = (51.5124, -0.0951)
        covering = self.connection(b6.rectangle_polygon(b6.ll(*topLeft), b6.ll(*bottomRight)).s2_covering(0, 30))
        rect = s2sphere.LatLngRect.from_point_pair(s2sphere.LatLng.from_degrees(*topLeft), s2sphere.LatLng.from_degrees(*bottomRight))
        count = 0
        for token, _ in covering:
            count += 1
            cell = s2sphere.Cell(s2sphere.CellId.from_token(token))
            self.assertTrue(rect.intersects(cell.get_rect_bound()))
        self.assertLess(count, 10)

    def test_s2_center(self):
        # TODO: Rewrite without using string() once we've sorted out node types
        (lat, lng) = self.connection(b6.s2_center("487604b4fbdc"))
        ll = s2sphere.LatLng.from_degrees(lat, lng)
        expected = s2sphere.LatLng.from_degrees(51.5126733, -0.1140124)
        self.assertLess(ll.get_distance(expected).radians, 0.000001)

    def test_join_paths(self):
        a = self.connection(b6.find_path(b6.osm_way_id(377974549)))
        b = self.connection(b6.find_path(b6.osm_way_id(834245629)))
        joined = self.connection(b6.find_path(b6.osm_way_id(377974549)).join(b6.find_path(b6.osm_way_id(834245629))))
        self.assertLess(abs((joined.length_meters() / (a.length_meters() + b.length_meters())) - 1.0), 0.0001)

    def test_ordered_join_paths(self):
        a = self.connection(b6.find_path(b6.osm_way_id(377974549)))
        b = self.connection(b6.find_path(b6.osm_way_id(834245629)))
        joined = self.connection(b6.find_path(b6.osm_way_id(377974549)).ordered_join(b6.find_path(b6.osm_way_id(834245629))))
        self.assertLess(abs((joined.length_meters() / (a.length_meters() + b.length_meters())) - 1.0), 0.0001)

    def test_points(self):
        points = self.connection(b6.find_path(b6.osm_way_id(STABLE_STREET_BRIDGE_ID)).points())
        expected = s2sphere.LatLng.from_degrees(51.535035, -0.1247934)
        ll = s2sphere.LatLng.from_degrees(*points[0][1])
        self.assertLess(ll.get_distance(expected).radians, 0.000001)

    def test_point_features(self):
        points = self.connection(b6.find_path(b6.osm_way_id(STABLE_STREET_BRIDGE_ID)).point_features())
        self.assertEqual(len(points), 2)
        self.assertEqual(b6.osm_node_id(STABLE_STREET_BRIDGE_NORTH_END_ID), points[0][0])

    def test_paths_by_point(self):
        paths = self.connection(b6.find_point(b6.osm_node_id(STABLE_STREET_BRIDGE_NORTH_END_ID)).point_paths())
        self.assertIn(b6.osm_way_id(STABLE_STREET_BRIDGE_ID), [id for (id, _) in paths])

    def test_interpolate(self):
        (lat, lng) = self.connection(b6.find_path(b6.osm_way_id(377974549)).interpolate(0.5))
        ll = s2sphere.LatLng.from_degrees(lat, lng)
        expected = s2sphere.LatLng.from_degrees(51.5361869, -0.1258445)
        self.assertLess(ll.get_distance(expected).radians, 0.000001)

    def test_snap_area_edges(self):
        building = b6.find_area(b6.osm_way_area_id(COAL_DROPS_YARD_WEST_BUILDING_ID))
        original_area = self.connection(building.area())
        snapped_area = self.connection(b6.snap_area_edges(building, b6.keyed("#highway"), 40.0).area())
        self.assertGreater(snapped_area, original_area)

    def test_geojson_map_areas(self):
        building = b6.find_area(b6.osm_way_area_id(COAL_DROPS_YARD_WEST_BUILDING_ID))
        original_area = self.connection(building.area())
        snapped = b6.to_geojson(building).map_geometries(b6.apply_to_area(lambda a: b6.snap_area_edges(a, b6.keyed("#highway"), 40.0)))
        areas = snapped.geojson_areas().map(lambda a: a.area())
        _, snapped_area = self.connection(areas)[0]
        self.assertGreater(snapped_area, original_area)

    def test_distance_to_point_meters(self):
        distance = self.connection(b6.find_path(b6.osm_way_id(377974549)).distance_to_point_meters(b6.ll(51.53586, -0.12564)))
        self.assertGreater(distance, 24.0)
        self.assertLess(distance, 25.0)

    def test_centroid(self):
        expected = b6.ll(51.5352611, -0.1243803)
        distance = self.connection(b6.distance_meters(b6.find_area(b6.osm_way_area_id(LIGHTERMAN_WAY_ID)).centroid(), expected))
        self.assertLess(distance, 0.1)

    def test_centroids(self):
        origin = b6.ll(51.5352611, -0.1243803)
        distances = self.connection(b6.find_areas(b6.keyed("#building")).centroid().map(lambda p: p.distance_meters(origin)))
        for f, distance in distances:
            self.assertLess(distance, 1000.0)

    def test_find_building_intersecting_point(self):
        point = b6.ll(51.5352611, -0.1243803)
        features = self.connection(b6.find(b6.and_(b6.tagged("#building", "yes"), b6.intersecting(point))))
        self.assertIn("The Lighterman", [f.get_string("name") for (_, f) in features])

    def test_map_area(self):
        areas = list(self.connection(b6.find_areas(b6.keyed("#building")).map(lambda building: building.area())))
        self.assertEqual(len(areas), BUILDINGS_IN_GRANARY_SQUARE)
        for (_, area) in areas:
            self.assertGreater(area, 50)
            self.assertLess(area, 10000)

    def test_path_length(self):
        bridge = self.connection(b6.find_path(b6.osm_way_id(STABLE_STREET_BRIDGE_ID)))
        self.assertGreater(bridge.length_meters(), 20.0)
        self.assertLess(bridge.length_meters(), 30.0)

    def test_connect_points(self):
        modified = self.connection(b6.find_point(b6.osm_node_id(VERMUTERIA_NODE_ID)).connect(b6.find_point(b6.osm_node_id(STABLE_STREET_BRIDGE_NORTH_END_ID))))
        self.assertEqual(len(modified), 1)
        for id in modified.values():
            self.assertTrue(id.is_path())
            self.assertEqual(id.namespace, b6.NAMESPACE_DIAGONAL_ACCESS_POINTS)

    def test_connect_point_to_network(self):
        modified = self.connection(b6.find_point(b6.osm_node_id(VERMUTERIA_NODE_ID)).connect_to_network())
        # The attempt to connect the point to the network fails, as the sample area
        # used by the tests is too small for us to consider any street as being part
        # of the street network, but the test at least verifies the wrapping of
        # network connection code
        self.assertEqual(len(modified), 0)

    def test_connect_area_to_network(self):
        modified = self.connection(b6.find_area(b6.osm_way_area_id(LIGHTERMAN_WAY_ID)).connect_to_network())
        self.assertEqual(len(modified), 0)

    def test_import_geojson_point(self):
        g = {
            "type": "FeatureCollection",
            "features": [
                {
                    "type": "Feature",
                    "geometry": {
                        "type": "Point",
                        "coordinates": [-0.1249292, 51.5352547],
                    },
                    "properties": {
                        "name": "Ruby Violet Truck",
                    }
                }
            ]
        }
        ids = self.connection(b6.import_geojson(b6.parse_geojson(json.dumps(g)), "diagonal.works/test"))
        self.assertEqual(len(ids), 1)
        id = list(ids.values())[0]
        self.assertEqual(self.connection(b6.find_point(id).get_string("name")), "Ruby Violet Truck")

    def test_import_geojson_path(self):
        g = {
            "type": "FeatureCollection",
            "features": [
                {
                    "type": "Feature",
                    "geometry": {
                        "type": "LineString",
                        "coordinates": [[-0.1251651, 51.5349089], [-0.1251580, 51.5347263]],
                    },
                    "properties": {
                        "bridge": "yes",
                    }
                }
            ]
        }
        ids = self.connection(b6.import_geojson(b6.parse_geojson(json.dumps(g)), "diagonal.works/test"))
        self.assertEqual(len(ids), 1)
        id = list(ids.values())[0]
        self.assertEqual(self.connection(b6.find_path(id)).get_string("bridge"), "yes")

    def test_import_geojson_polygon(self):
        g = {
            "type": "FeatureCollection",
            "features": [
                {
                    "type": "Feature",
                    "geometry": {
                        "type": "Polygon",
                        "coordinates": [[[-0.1243817, 51.5354124], [-0.1243411, 51.5351416], [-0.1242415, 51.5353736]]],
                    },
                    "properties": {
                        "building": "yes",
                    }
                }
            ]
        }
        ids = self.connection(b6.import_geojson(b6.parse_geojson(json.dumps(g)), "diagonal.works/test"))
        self.assertEqual(len(ids), 1)
        id = list(ids.values())[0]
        self.assertEqual(self.connection(b6.find_area(id)).get_string("building"), "yes")

    def test_import_geojson_multipolygon(self):
        g = {
            "type": "FeatureCollection",
            "features": [
                {
                    "type": "Feature",
                    "geometry": {
                        "type": "MultiPolygon",
                        "coordinates": [
                            [[[-0.1243817, 51.5354124], [-0.1243411, 51.5351416], [-0.1242415, 51.5353736]]],
                            [[[-0.1239823, 51.5358407], [-0.1240998, 51.5355521], [-0.1238063, 51.5358096]]],
                        ],
                    },
                    "properties": {
                        "building": "yes",
                    }
                }
            ]
        }
        ids = self.connection(b6.import_geojson(b6.parse_geojson(json.dumps(g)), "diagonal.works/test"))
        self.assertEqual(len(ids), 1)
        id = list(ids.values())[0]
        self.assertEqual(self.connection(b6.find_area(id)).get_string("building"), "yes")

    def test_import_geojson_file(self):
        ids = self.connection(b6.import_geojson_file("data/tests/granary-square.geojson", "diagonal.works/test"))
        self.assertGreater(len(ids), 0)
        id = list(ids.values())[0]
        self.assertGreater(self.connection(b6.find_area(id).area()), 100.0)

    def test_evaluate_with_changed_world(self):
        close_road = b6.remove_tag(b6.osm_way_id(STABLE_STREET_BRIDGE_ID), "#highway")
        reachable = b6.find_point(b6.osm_node_id(STABLE_STREET_BRIDGE_SOUTH_END_ID)).reachable("walk", 200.0, b6.keyed("#amenity")).get_string("name")
        before = len(self.connection(reachable))
        after = len(self.connection(b6.with_change(close_road, lambda: reachable)))
        self.assertGreater(before, after)

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--http-port", default="10080", help="Host and port on which to serve HTTP")
    parser.add_argument("--grpc-port", default="10081", help="Host and port on which to serve GRPC")
    parser.add_argument("--platform", default=None, help="Current platform, eg darwin/x86_64")
    parser.add_argument("tests", type=str, nargs="*")
    args = parser.parse_args()
    http_address = "localhost:%d" % int(args.http_port)
    grpc_address = "localhost:%d" % int(args.grpc_port)
    platform = args.platform or os.environ.get("TARGETPLATFORM", None) or ("%s/%s" % (os.uname().sysname.lower(), os.uname().machine))
    p = subprocess.Popen([
        "bin/%s/b6" % platform,
        "--http=%s" % http_address,
        "--grpc=%s" % grpc_address,
        "--world=osm:data/tests/granary-square.osm.pbf"
    ])

    ready = False
    for _ in range(20):
        if p.poll() is not None:
            print("Server exited with status %d" % p.poll())
            return
        try:
            r = urllib.request.urlopen("http://%s/healthy" % http_address)
            if r.status == 200:
                if r.read().strip() == b"ok":
                    ready = True
                break
        except urllib.error.URLError:
            pass
        time.sleep(0.5)
    if not ready:
        print("Server was never healthy")
        return
    connection = b6.connect_insecure(grpc_address)
    suite = unittest.TestSuite()
    if len(args.tests) > 0:
        for test in args.tests:
            suite.addTest(B6Test(test, connection))
    else:
        for method in dir(B6Test):
            if method.startswith("test_"):
                suite.addTest(B6Test(method, connection))
    runner = unittest.TextTestRunner()
    runner.run(suite)
    p.terminate()
    p.wait()

if __name__ == "__main__":
    main()