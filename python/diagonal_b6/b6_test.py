#!/usr/bin/env python3

import argparse
import json
import s2sphere
import subprocess
import time
import urllib.request
import unittest
import sys

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
HIGHWAY_AREAS_IN_GRANARY_SQUARE = 5
BIKE_PARKING_IN_GRANARY_SQUARE = 11
FOUNTAINS_IN_GRANARY_SQUARE = 4 # Within the pedestrian square itself defined by WKT below not the entire area
STABLE_STREET_BRIDGE_NORTH_END_DEGREE = 7 # After it has been connected

GRANARY_SQUARE_POLYGON_WKT="POLYGON ((-0.1260475 51.5357019,-0.1261001 51.5355674,-0.1261596 51.5354153,-0.1262097 51.535287,-0.1259034 51.5352365,-0.1259462 51.5351347,-0.1255806 51.5350765,-0.1255202 51.5350667,-0.1255004 51.5350372,-0.1254536 51.5349963,-0.1254346 51.5350013,-0.1252611 51.535049,-0.125219 51.5350629,-0.124904 51.5350121,-0.1247915 51.5350326,-0.124709 51.5350541,-0.1247491 51.5351308,-0.1247727 51.5351758,-0.1246766 51.5353808,-0.1246363 51.5354737,-0.125082 51.5355458,-0.1259754 51.5356902,-0.1260475 51.5357019))"
GRANARY_SQUARE_MULTIPOLYGON_WKT="MULTIPOLYGON (((-0.1260475 51.5357019,-0.1261001 51.5355674,-0.1261596 51.5354153,-0.1262097 51.535287,-0.1259034 51.5352365,-0.1259462 51.5351347,-0.1255806 51.5350765,-0.1255202 51.5350667,-0.1255004 51.5350372,-0.1254536 51.5349963,-0.1254346 51.5350013,-0.1252611 51.535049,-0.125219 51.5350629,-0.124904 51.5350121,-0.1247915 51.5350326,-0.124709 51.5350541,-0.1247491 51.5351308,-0.1247727 51.5351758,-0.1246766 51.5353808,-0.1246363 51.5354737,-0.125082 51.5355458,-0.1259754 51.5356902,-0.1260475 51.5357019)))"

EARTH_RADIUS_METERS = 6371.01 * 1000.0

def angle_to_meters(angle):
	return angle.radians * EARTH_RADIUS_METERS

class B6Test(unittest.TestCase):

    def __init__(self, name, connection, grpc_address):
        unittest.TestCase.__init__(self, name)
        self.connection = connection
        self.grpc_address = grpc_address

    def test_get_tag(self):
        name = b6.find_feature(b6.osm_way_area_id(LIGHTERMAN_WAY_ID)).get("name")
        self.assertEqual(("name", "The Lighterman"), self.connection(name))

    def test_find_areas(self):
        names = [building.get_string("name") for (id, building) in self.connection(b6.find_areas(b6.keyed("#building")))]
        self.assertEqual(len(names), BUILDINGS_IN_GRANARY_SQUARE)

    def test_find_point_by_id(self):
        area = self.connection(b6.find_feature(b6.osm_node_id(STABLE_STREET_BRIDGE_SOUTH_END_ID)))
        self.assertEqual(area.id.value, STABLE_STREET_BRIDGE_SOUTH_END_ID)

    def test_find_area_by_id(self):
        area = self.connection(b6.find_area(b6.osm_way_area_id(COAL_DROPS_YARD_WEST_BUILDING_ID)))
        self.assertEqual(area.id.value, COAL_DROPS_YARD_WEST_BUILDING_ID)

    def test_find_non_existant_id(self):
        with self.assertRaises(Exception):
            self.assertEqual(self.connection(b6.find_feature(b6.osm_node_id(42))), None)

    def test_find_area_by_wrong_id_type(self):
        with self.assertRaises(Exception): # TODO: Make more specific
            self.connection(b6.find_area(b6.osm_node_id(STABLE_STREET_BRIDGE_SOUTH_END_ID)))

    def test_uk_ons_boundary_id(self):
        self.assertEqual(b6.uk_ons_boundary_id("E01000953").value, 76343044687353)

    def test_area_str(self):
        area = self.connection(b6.find_area(b6.osm_way_area_id(COAL_DROPS_YARD_WEST_BUILDING_ID)))
        self.assertEqual("<Area /area/openstreetmap.org/way/222021572>", str(area))

    def test_relation_members(self):
        greenway = self.connection(b6.find_relation(b6.osm_relation_id(JUBILEE_GREENWAY_ID)))
        paths = ([str(m) for m in greenway.members() if m.is_path()])
        self.assertGreater(len(paths), 10)
        self.assertLess(len(paths), 800)

    def test_or_query(self):
        names = [amenity.get_string("name") for (id, amenity) in self.connection(b6.find(b6.tagged("#amenity", "restaurant").or_(b6.tagged("#amenity", "cafe"))))]
        self.assertIn("Le Cafe Alain Ducasse", names)

    def test_point_degree(self):
        self.assertEqual(self.connection(b6.find_feature(b6.osm_node_id(STABLE_STREET_BRIDGE_NORTH_END_ID)).degree()), STABLE_STREET_BRIDGE_NORTH_END_DEGREE)
        degrees = [degree for (id, degree) in self.connection(b6.find(b6.within_cap(b6.ll(51.535241, -0.124364), 100)).degree())]
        for d in degrees:
            self.assertGreaterEqual(d, 0)
            self.assertLess(d, 10)

    def test_send_evaluated_feature_back_to_server(self):
        point = self.connection(b6.find_feature(b6.osm_node_id(STABLE_STREET_BRIDGE_NORTH_END_ID)))
        degree = self.connection(b6.degree(point))
        self.assertEqual(degree, self.connection(b6.find_feature(b6.osm_node_id(STABLE_STREET_BRIDGE_NORTH_END_ID)).degree()))

    def test_path_lengths(self):
        lengths = [length for (id, length) in self.connection(b6.find(b6.typed("path", b6.keyed("#highway"))).length())]
        self.assertGreater(len(lengths), 0)
        for l in lengths: # noqa: E741
            self.assertGreater(l, 0)
            self.assertLess(l, 1000)

    def test_relation_names(self):
        names = [r.get_string("name") for (id, r) in self.connection(b6.find_relations(b6.keyed("#route")))]
        self.assertIn("Jubilee Greenway", names)

    def test_reachable_areas_from_point(self):
        expression = b6.find_feature(b6.osm_node_id(STABLE_STREET_BRIDGE_SOUTH_END_ID)).reachable({"mode": "walk"}, 200.0, b6.keyed("#amenity")).get_string("name")
        names = set([name for (_, name) in self.connection(expression)])
        self.assertIn("The Lighterman", names)

    def test_reachable_with_distance(self):
        small = b6.find_feature(b6.osm_node_id(STABLE_STREET_BRIDGE_SOUTH_END_ID)).reachable({"mode": "walk"}, 100.0, b6.keyed("#amenity")).count()
        large = b6.find_feature(b6.osm_node_id(STABLE_STREET_BRIDGE_SOUTH_END_ID)).reachable({"mode": "walk"}, 200.0, b6.keyed("#amenity")).count()
        self.assertGreater(self.connection(large), self.connection(small))

    def test_paths_to_reach(self):
        expression = b6.find_feature(b6.osm_node_id(STABLE_STREET_BRIDGE_SOUTH_END_ID)).paths_to_reach({"mode": "walk"}, 200.0, b6.keyed("#amenity"))
        paths = list(self.connection(expression))
        self.assertGreaterEqual(len(paths), 4)
        for path, count in paths:
            self.assertGreaterEqual(count, 1)
            self.assertLess(count, 100)

    def test_accessible_all(self):
        origins = [b6.find_feature(b6.osm_node_id(STABLE_STREET_BRIDGE_NORTH_END_ID))]
        destinatations = list(self.connection(b6.accessible_all(origins, b6.keyed("entrance"), 500, {"mode": "walk"})))
        self.assertGreater(len(destinatations), 2)

    def test_accessible_routes(self):
        origin = b6.find_feature(b6.osm_node_id(STABLE_STREET_BRIDGE_NORTH_END_ID))
        routes = list(self.connection(b6.accessible_routes(origin, b6.keyed("entrance"), 500, {"mode": "walk"})))
        self.assertGreater(len(routes), 2)
        for (_, route) in routes:
            self.assertGreater(len(route), 4) # At least 4 steps
            self.assertGreater(route.cost(), 100.0)
            self.assertLess(route.cost(), 500.0)

    def test_closest_from_point(self):
        expression = b6.find_feature(b6.osm_node_id(STABLE_STREET_BRIDGE_SOUTH_END_ID)).closest({"mode": "walk"}, 1000.0, b6.tagged("#amenity", "pub"))
        pub = self.connection(expression)
        self.assertEqual("The Lighterman", pub.get_string("name"))

    def test_closest_from_point_distance(self):
        expression = b6.find_feature(b6.osm_node_id(STABLE_STREET_BRIDGE_SOUTH_END_ID)).closest_distance({"mode": "walk"}, 1000.0, b6.tagged("#amenity", "pub"))
        distance = self.connection(expression)
        self.assertGreater(distance, 128.0)
        self.assertLess(distance, 129.0)

    def test_closest_from_area(self):
        expression = b6.find_area(b6.osm_way_area_id(COAL_DROPS_YARD_WEST_BUILDING_ID)).closest({"mode": "walk"}, 1000.0, b6.tagged("#amenity", "pub"))
        pub = self.connection(expression)
        self.assertEqual("The Lighterman", pub.get_string("name"))

    def test_closest_from_point_non_existant(self):
        expression = b6.find_feature(b6.osm_node_id(STABLE_STREET_BRIDGE_SOUTH_END_ID)).closest({"mode": "walk"}, 1000.0, b6.tagged("#amenity", "nonexistant"))
        self.assertEqual(None, self.connection(expression))

    def test_containing_areas_from_point(self):
        expression = b6.find_feature(b6.osm_node_id(STABLE_STREET_BRIDGE_SOUTH_END_ID)).reachable({"mode": "walk"}, 1000.0, b6.all()).containing_areas(b6.keyed("#shop")).get_string("name")
        names = set([name for (_, name) in self.connection(expression)])
        self.assertIn("Coal Drops Yard", names)

    def test_containing_areas_from_area(self):
        areas = self.connection(b6.find_area(b6.osm_way_area_id(COAL_DROPS_YARD_WEST_BUILDING_ID)).reachable({"mode": "walk"}, 1000.0, b6.all()).containing_areas(b6.all()))
        self.assertGreater(len(areas), 10)

    def test_count_features(self):
        self.assertEqual(self.connection(b6.find(b6.tagged("#amenity", "bicycle_parking")).count()), BIKE_PARKING_IN_GRANARY_SQUARE)
        self.assertEqual(self.connection(b6.find(b6.typed("path", b6.keyed("#highway"))).count()), HIGHWAYS_IN_GRANARY_SQUARE)
        self.assertEqual(self.connection(b6.find(b6.typed("area", b6.keyed("#highway"))).count()), HIGHWAY_AREAS_IN_GRANARY_SQUARE)
        self.assertEqual(self.connection(b6.find_areas(b6.keyed("#building")).count()), BUILDINGS_IN_GRANARY_SQUARE)

    def test_sum(self):
        self.assertEqual(self.connection(b6.sum(b6.collection(b6.pair("one", 1), b6.pair("two", 2)))), 3)

    def test_divide_count_features(self):
        self.assertAlmostEqual(self.connection(b6.find(b6.tagged("#amenity", "bicycle_parking")).count().divide(10.0)), BIKE_PARKING_IN_GRANARY_SQUARE / 10.0)

    def test_to_str(self):
        self.connection(b6.add_tags(b6.find_areas(b6.keyed("#building")).map(lambda building: b6.tag("#reachable-within-km", building.reachable({"mode": "walk"}, 1000, b6.keyed("#highway")).count().to_str()))))
        self.assertEqual(self.connection(b6.find_area(b6.osm_way_area_id(COAL_DROPS_YARD_WEST_BUILDING_ID)).get_string("#reachable-within-km")), "9")

    def test_filter(self):
        filtered = self.connection(b6.find_areas(b6.keyed("#amenity")).filter(lambda a: b6.matches(a, b6.keyed("addr:postcode"))))
        self.assertGreater(len(filtered), 0)
        for (_, feature) in filtered:
            self.assertNotEqual(feature.get_string("addr:postcode"), "")

    def test_filter_with_implicit_function(self):
        filtered = self.connection(b6.find(b6.tagged("#amenity", "restaurant")).filter(b6.tagged("cuisine", "indian")).map(lambda f: b6.get_string(f, "name")))
        self.assertEqual(["Dishoom"], [name for (id, name) in filtered])

    def test_to_geojson_collection(self):
        geojson = self.connection(b6.to_geojson_collection(b6.find_areas(b6.keyed("#building"))))
        self.assertGreater(len(geojson["features"]), 4)
        for feature in geojson["features"]:
            self.assertIn("#building", feature["properties"])

    def test_to_geojson_with_feature(self):
        closest = b6.find_feature(b6.osm_node_id(STABLE_STREET_BRIDGE_SOUTH_END_ID)).closest({"mode": "walk"}, 1000.0, b6.tagged("#amenity", "pub"))
        geojson = self.connection(b6.to_geojson(closest))
        self.assertEqual(geojson["type"], "Feature")

    def test_to_geojson_with_missing_feature(self):
        with self.assertRaises(Exception):
            geojson = self.connection(b6.to_geojson(b6.find_feature(b6.osm_node_id(1))))
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
        applied = self.connection(b6.add_tags(b6.find(b6.tagged("#highway", "footway")).filter(b6.keyed("bicycle")).map(lambda h: b6.tag("#bicycle", h.get_string("bicycle")))))
        self.assertGreater(len(applied), 0)
        self.assertEqual(self.connection(b6.find(b6.keyed("#bicycle")).count()), len(applied))

    def test_search_for_newly_added_tag(self):
        reachable = b6.find_feature(b6.osm_node_id(STABLE_STREET_BRIDGE_SOUTH_END_ID)).reachable({"mode": "walk"}, 1000.0, b6.keyed("#amenity"))
        modified = self.connection(b6.add_tags(reachable.map(lambda building: b6.tag("#reachable", "yes"))))
        self.assertGreater(len(modified), 1)
        self.assertEqual(self.connection(b6.find(b6.keyed("#reachable")).count()), len(modified))

    def test_sample_points_along_path(self):
        count = len(self.connection(b6.sample_points(b6.find_feature(b6.osm_way_id(STABLE_STREET_BRIDGE_ID)), 1.0)))
        self.assertGreater(count, 20)
        self.assertLess(count, 40)

    def test_sample_points_along_paths(self):
        count = 0
        center = s2sphere.LatLng.from_degrees(51.53539, -0.12537)
        for _, ll in self.connection(b6.find(b6.keyed("#highway")).sample_points_along_paths(20.0)):
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
        geojson = self.connection(b6.to_geojson_collection(b6.sample_points(b6.find_feature(bridge), 5.0).map(lambda point: b6.sightline(point, 250.0))))
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
        for _, token in grid:
            cell = s2sphere.Cell(s2sphere.CellId.from_token(token))
            self.assertEqual(cell.level(), 21)
            self.assertTrue(rect.intersects(cell.get_rect_bound()))

    def test_s2_covering(self):
        topLeft = (51.5146, -0.1140)
        bottomRight = (51.5124, -0.0951)
        covering = self.connection(b6.rectangle_polygon(b6.ll(*topLeft), b6.ll(*bottomRight)).s2_covering(0, 30))
        rect = s2sphere.LatLngRect.from_point_pair(s2sphere.LatLng.from_degrees(*topLeft), s2sphere.LatLng.from_degrees(*bottomRight))
        count = 0
        for _, token in covering:
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
        a = self.connection(b6.find_feature(b6.osm_way_id(377974549)))
        b = self.connection(b6.find_feature(b6.osm_way_id(834245629)))
        joined = self.connection(b6.join(b6.find_feature(b6.osm_way_id(377974549)), b6.find_feature(b6.osm_way_id(834245629))))
        self.assertLess(abs((joined.length_meters() / (a.length_meters() + b.length_meters())) - 1.0), 0.0001)

    def test_ordered_join_paths(self):
        a = self.connection(b6.find_feature(b6.osm_way_id(377974549)))
        b = self.connection(b6.find_feature(b6.osm_way_id(834245629)))
        joined = self.connection(b6.ordered_join(b6.find_feature(b6.osm_way_id(377974549)), b6.find_feature(b6.osm_way_id(834245629))))
        self.assertLess(abs((joined.length_meters() / (a.length_meters() + b.length_meters())) - 1.0), 0.0001)

    def test_points(self):
        points = self.connection(b6.points(b6.find_feature(b6.osm_way_id(STABLE_STREET_BRIDGE_ID))))
        expected = s2sphere.LatLng.from_degrees(51.535035, -0.1247934)
        ll = s2sphere.LatLng.from_degrees(*points[0][1])
        self.assertLess(ll.get_distance(expected).radians, 0.000001)

    def test_point_features(self):
        points = self.connection(b6.find_feature(b6.osm_way_id(STABLE_STREET_BRIDGE_ID)).point_features())
        self.assertEqual(len(points), 2)
        self.assertEqual(b6.osm_node_id(STABLE_STREET_BRIDGE_NORTH_END_ID), points[0][0])

    def test_paths_by_point(self):
        paths = self.connection(b6.find_feature(b6.osm_node_id(STABLE_STREET_BRIDGE_NORTH_END_ID)).point_paths())
        self.assertIn(b6.osm_way_id(STABLE_STREET_BRIDGE_ID), [id for (id, _) in paths])

    def test_interpolate(self):
        (lat, lng) = self.connection(b6.interpolate(b6.find_feature(b6.osm_way_id(377974549)), 0.5))
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

    def test_collect_areas(self):
        collected_areas = self.connection(b6.collect_areas(b6.find_areas(b6.keyed("#building"))).area())
        summed_areas = 0
        for _, area in self.connection(b6.find_areas(b6.keyed("#building")).map(lambda b: b.area())):
            summed_areas += area
        self.assertLess((collected_areas - summed_areas)/summed_areas, 0.0001)

    def test_distance_to_point_meters(self):
        distance = self.connection(b6.distance_to_point_meters(b6.find_feature(b6.osm_way_id(377974549)), b6.ll(51.53586, -0.12564)))
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
        bridge = self.connection(b6.find_feature(b6.osm_way_id(STABLE_STREET_BRIDGE_ID)))
        self.assertGreater(bridge.length_meters(), 20.0)
        self.assertLess(bridge.length_meters(), 30.0)

    def test_connect_points(self):
        modified = self.connection(b6.find_feature(b6.osm_node_id(VERMUTERIA_NODE_ID)).connect(b6.find_feature(b6.osm_node_id(STABLE_STREET_BRIDGE_NORTH_END_ID))))
        self.assertEqual(len(modified), 1)
        for (_, id) in modified:
            self.assertTrue(id.is_path())
            self.assertEqual(id.namespace, b6.NAMESPACE_DIAGONAL_ACCESS_POINTS)

    def test_connect_point_to_network(self):
        modified = self.connection(b6.find_feature(b6.osm_node_id(VERMUTERIA_NODE_ID)).connect_to_network())
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
        self.assertEqual(ids[0][1].namespace, "diagonal.works/test")
        self.assertEqual(self.connection(b6.find_feature(ids[0][1]).get_string("name")), "Ruby Violet Truck")

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
        self.assertEqual(self.connection(b6.find_feature(ids[0][1])).get_string("bridge"), "yes")

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
        self.assertEqual(self.connection(b6.find_area(ids[0][1])).get_string("building"), "yes")

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
        self.assertEqual(self.connection(b6.find_area(ids[0][1])).get_string("building"), "yes")

    def test_import_geojson_file(self):
        ids = self.connection(b6.import_geojson_file("data/tests/granary-square.geojson", "diagonal.works/test"))
        self.assertGreater(len(ids), 0)
        self.assertGreater(self.connection(b6.find_area(ids[0][1]).area()), 100.0)

    def test_parse_geojson_file(self):
        areas = b6.parse_geojson_file("data/tests/granary-square.geojson").geojson_areas()
        area = self.connection(b6.convex_hull(areas).area())
        self.assertGreater(area, 2400.0)
        self.assertLess(area, 2500.0)

    def test_reachable_with_changed_world(self):
        close_road = b6.remove_tag(b6.osm_way_id(STABLE_STREET_BRIDGE_ID), "#highway")
        reachable = b6.find_feature(b6.osm_node_id(STABLE_STREET_BRIDGE_SOUTH_END_ID)).reachable({"mode": "walk"}, 200.0, b6.keyed("#amenity")).get_string("name")
        before = len(self.connection(reachable))
        after = len(self.connection(b6.with_change(close_road, lambda: reachable)))
        self.assertGreater(before, after)

    def test_remove_tags(self):
        roads = b6.find(b6.keyed("#highway"))
        before = self.connection(roads.count())
        close_roads = b6.remove_tags(roads.map(lambda road: "#highway"))
        after = self.connection(b6.with_change(close_roads, lambda: roads.count()))
        self.assertGreater(before, 0)
        self.assertEqual(after, 0)

    def test_merge_changes(self):
        roads = b6.find(b6.keyed("#highway"))
        before = self.connection(roads.count())
        close_roads = b6.merge_changes(roads.map(lambda h: b6.remove_tag(h, "#highway")))
        after = self.connection(b6.with_change(close_roads, lambda: roads.count()))
        self.assertGreater(before, 0)
        self.assertEqual(after, 0)

    def test_get_tags_from_list_of_ids(self):
        names = b6.map([b6.osm_way_area_id(id) for id in (LIGHTERMAN_WAY_ID, GRANARY_SQUARE_WAY_ID)], lambda f: b6.get_string(f, "name"))
        expected = [(0, "The Lighterman"), (1, "Granary Square")]
        self.assertEqual(expected, self.connection(names))

    def test_make_tags_from_list_of_strings(self):
        tags = b6.map(["primary", "secondary"], lambda v: b6.tag("#highway", v))
        expected = [(0, ("#highway", "primary")), (1, ("#highway", "secondary"))]
        self.assertEqual(expected, self.connection(tags))

    def test_convex_hull_from_list_of_lat_lngs(self):
        caps = b6.map([b6.ll(51.535387, -0.125277), b6.ll(51.537088, -0.125781)], lambda c: b6.cap_polygon(c, 20.0))
        areas = self.connection(caps.map(b6.area))
        hull_area = self.connection(b6.convex_hull(caps).area())
        self.assertGreater(hull_area, sum([a for _, a in areas]))

    def test_collection(self):
        ids = b6.collection(b6.pair(0, b6.osm_way_area_id(GRANARY_SQUARE_WAY_ID)), b6.pair(1, b6.osm_way_area_id(LIGHTERMAN_WAY_ID)))
        areas = self.connection(ids.map(lambda id: b6.area(b6.find_area(id))))
        for (i, (j, area)) in enumerate(areas):
            self.assertEqual(i, j)
            self.assertGreater(area, 0.0)
            self.assertLess(area, 6000.0)

    def test_map_literal_collection_from_dict(self):
        collection = {
            b6.tag("highway", "motorway"): 3,
            b6.tag("highway", "primary"): 7,
        }
        result = self.connection(b6.map(collection, lambda count: b6.add(count, 1)))
        self.assertEqual([4, 8], sorted([count for ((key, value), count) in result]))

    def test_map_literal_collection_from_list(self):
        collection = [36, 42]
        result = self.connection(b6.map(collection, lambda count: b6.add(count, 1)))
        self.assertEqual([37, 43], sorted([count for (key, count) in result]))

    def test_flatten(self):
        parks = b6.tagged("#leisure", "park")
        grass = b6.tagged("#landuse", "grass")
        parks_count = self.connection(b6.find(parks).count())
        grass_count = self.connection(b6.find(grass).count())
        self.assertGreater(parks_count, 0)
        self.assertGreater(grass_count, 0)
        parks_and_grass = b6.map([parks, grass], lambda q: b6.find(q)).flatten()
        self.assertEqual(parks_count + grass_count, self.connection(parks_and_grass.count()))

    def test_add_point(self):
        id = b6.FeatureID(b6.FEATURE_TYPE_POINT, "diagonal.works/restaurants", 0)
        add = b6.add_point(b6.ll(51.537165, -0.125737), id, [b6.tag("#amenity", "restaurant"), b6.tag("name", "noma")])
        names = self.connection(b6.with_change(add, lambda: b6.map(b6.find(b6.tagged("#amenity", "restaurant")), lambda r: b6.get_string(r, "name"))))
        self.assertIn("noma", [name for (id, name) in names])

    def test_add_relation(self):
        id = b6.id_to_relation_id("diagonal.works/test", b6.osm_way_id(STABLE_STREET_BRIDGE_ID))
        add = b6.add_relation(id, [b6.tag("#route", "bicycle")], {b6.osm_way_id(STABLE_STREET_BRIDGE_ID): "forwards"})
        route = self.connection(b6.with_change(add, lambda: b6.find_feature(id).get_string("#route")))
        self.assertEqual("bicycle", route)

    def test_materialise(self):
        collection_id = b6.FeatureID(b6.FEATURE_TYPE_COLLECTION, "diagonal.works/test", 1)
        roads = b6.find(b6.keyed("#highway"))
        n = self.connection(b6.with_change(b6.materialise(collection_id, lambda: roads), lambda: b6.count(b6.find_feature(collection_id))))
        self.assertGreater(n, 100)
        self.assertLess(n, 200)
        collection = self.connection(b6.with_change(b6.materialise(collection_id, lambda: roads), lambda: b6.find_feature(collection_id)))
        self.assertEqual(n, len(collection))
        self.assertEqual(n, len(list(collection)))
        self.assertIn(b6.osm_way_id(STABLE_STREET_BRIDGE_ID), [id for (id, _) in collection])

    def test_materialise_includes_expression(self):
        collection_id = b6.FeatureID(b6.FEATURE_TYPE_COLLECTION, "diagonal.works/test", 1)
        expression_id = b6.FeatureID(b6.FEATURE_TYPE_EXPRESSION, "diagonal.works/test", 1)
        roads = b6.find(b6.keyed("#highway"))
        expression = self.connection(b6.with_change(b6.materialise(collection_id, lambda: roads), lambda: b6.find_feature(expression_id)))
        # TODO: Test expression behaviour in Python, once it exists.
        self.assertIsNotNone(expression)

    def test_name_expression(self):
        origin = b6.name(b6.find_feature(b6.osm_node_id(STABLE_STREET_BRIDGE_SOUTH_END_ID)), "bridge")
        query = b6.name(b6.keyed("#amenity"), "amenities")
        distance = b6.name(200.0, "200m")
        count = b6.count(b6.reachable(origin, {"mode": "walk"}, distance, query))
        # The functionality of name itself is tested at the UI level. Here
        # we're just verifying that the API works.
        self.assertGreater(self.connection(count), 0)

    def test_modify_different_world(self):
        bridge = b6.osm_way_id(STABLE_STREET_BRIDGE_ID)
        self.connection(b6.add_tag(bridge, b6.tag("maxspeed", "10")))
        root = b6.FeatureID(b6.FEATURE_TYPE_COLLECTION, "diagonal.works/test/world", 0)
        new_connection = b6.connect_insecure(self.grpc_address, root=root)
        new_connection(b6.add_tag(bridge, b6.tag("maxspeed", "5")))
        self.assertEqual(self.connection(b6.get_string(bridge, "maxspeed")), "10")
        self.assertEqual(new_connection(b6.get_string(bridge, "maxspeed")), "5")

    def test_list_worlds(self):
        bridge = b6.osm_way_id(STABLE_STREET_BRIDGE_ID)
        root = b6.FeatureID(b6.FEATURE_TYPE_COLLECTION, "diagonal.works/test_list_worlds", 0)
        new_connection = b6.connect_insecure(self.grpc_address, root=root)
        new_connection(b6.add_tag(bridge, b6.tag("maxspeed", "5")))
        self.assertIn(root, new_connection.list_worlds())

    def test_delete_world(self):
        bridge = b6.osm_way_id(STABLE_STREET_BRIDGE_ID)
        self.connection(b6.add_tag(bridge, b6.tag("maxspeed", "10")))
        root = b6.FeatureID(b6.FEATURE_TYPE_COLLECTION, "diagonal.works/test_delete_world", 0)
        new_connection = b6.connect_insecure(self.grpc_address, root=root)
        new_connection(b6.add_tag(bridge, b6.tag("maxspeed", "5")))
        self.assertEqual(new_connection(b6.get_string(bridge, "maxspeed")), "5")
        new_connection.delete_world(root)
        self.assertEqual(new_connection(b6.get_string(bridge, "maxspeed")), "")

    def test_add_world_with_change(self):
        bridge = b6.osm_way_id(STABLE_STREET_BRIDGE_ID)
        change = b6.add_tag(bridge, b6.tag("maxspeed", "10"))
        root = b6.FeatureID(b6.FEATURE_TYPE_COLLECTION, "diagonal.works/test_add_world_with_change", 0)
        self.connection(b6.add_world_with_change(root, change))
        new_connection = b6.connect_insecure(self.grpc_address, root=root)
        self.assertEqual(new_connection(b6.get_string(bridge, "maxspeed")), "10")

    def test_add_and_call_expression(self):
        id = b6.FeatureID(b6.FEATURE_TYPE_EXPRESSION, "diagonal.works/test_add_and_call_expression", 0)
        add = b6.add_expression(id, [b6.tag("help", "Add 10")], lambda i: b6.add(i, 10))
        self.connection(add)
        self.assertEqual(self.connection(b6.call(b6.evaluate_feature(id), 20)), 30)

    def test_filter_invalid(self):
      c = b6.keyed("#building")
      o = [b6.find_feature(b6.osm_node_id(STABLE_STREET_BRIDGE_NORTH_END_ID))]
      q = b6.accessible_all(o, c, 10.0, {"mode": "walk"})
      m = len(self.connection(q))
      n = len(self.connection(q.filter(lambda f: b6.matches(f, b6.is_valid()))))
      # Should've filtered away the invalid item; i.e.
      #   n < m
      self.assertLess(n, m)

    def test_get_centroid(self):
      f = b6.find_feature(b6.osm_node_id(STABLE_STREET_BRIDGE_NORTH_END_ID))
      ll = self.connection(b6.distance_meters(f.get_centroid(), f.get_centroid()))
      # Distance to itself is 0.0.
      self.assertEqual(ll, 0.0)

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--http-port", default="10080", help="Host and port on which to serve HTTP")
    parser.add_argument("--grpc-port", default="10081", help="Host and port on which to serve GRPC")
    parser.add_argument("tests", type=str, nargs="*")
    args = parser.parse_args()
    http_address = "localhost:%d" % int(args.http_port)
    grpc_address = "localhost:%d" % int(args.grpc_port)
    p = subprocess.Popen([
        "bin/b6",
        "--http=%s" % http_address,
        "--grpc=%s" % grpc_address,
        "--world=data/tests/granary-square.index"
    ])

    ready, response = wait_until_ready(p, http_address)
    if not ready:
        print("Server is not ready: %s" % response)
        return
    connection = b6.connect_insecure(grpc_address)
    suite = unittest.TestSuite()
    if len(args.tests) > 0:
        for test in args.tests:
            suite.addTest(B6Test(test, connection, grpc_address))
    else:
        for method in dir(B6Test):
            if method.startswith("test_"):
                suite.addTest(B6Test(method, connection, grpc_address))
    runner = unittest.TextTestRunner()
    result = runner.run(suite)
    p.terminate()
    p.wait()
    if result.wasSuccessful():
        sys.exit(0)
    else:
        sys.exit(1)

def wait_until_ready(p, http_address):
    max_attempts = 20
    for i in range(max_attempts):
        if p.poll() is not None:
            return False, "Server exited with status %d" % p.poll()
        try:
            r = urllib.request.urlopen("http://%s/healthy" % http_address)
            if r.status == 200:
                response = r.read().strip()
                ok = b"ok"
                return response == ok, "Server got response: %s, expected %s" % (response, ok)
        except urllib.error.URLError:
            pass
        print("Waiting for server (attempt %d/%d)..." % (i+1, max_attempts))
        time.sleep(0.5)
    return False, "Timeout while waiting for server to start"

if __name__ == "__main__":
    main()
