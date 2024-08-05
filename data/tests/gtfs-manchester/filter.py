#!/usr/bin/env python3
# Generate a sample of a GTFS feed, suitable for using in tests. Reads a full
# set of GTFS files from the current directory, and output a sample of the
# routes to a new set of files, called stops.filtered.csv etc. Deliberately
# doesn't rely on csvreader etc to copy the input data verbatim.

SAMPLE_EACH_N_ROUTES = 100

sampled_routes = set()
with open("routes.txt") as input:
    next(input) # Skip header
    routes = []
    for line in input:
        routes.append(line.split(",")[0])
    routes.sort()
    for (i, route) in enumerate(routes):
        if i % SAMPLE_EACH_N_ROUTES == 0:
            sampled_routes.add(route)

with open("routes.txt") as input:
    with open("routes.filtered.txt", "w") as output:
        output.write(next(input))
        for line in input:
            if line.split(",")[0] in sampled_routes:
                output.write(line)

TRIPS_ROUTE_ID = 0
TRIPS_TRIP_ID = 2

sampled_trips = set()
with open("trips.txt") as input:
    with open("trips.filtered.txt", "w") as output:
        output.write(next(input))
        for line in input:
            columns = line.split(",")
            if columns[TRIPS_ROUTE_ID] in sampled_routes:
                sampled_trips.add(columns[TRIPS_TRIP_ID])
                output.write(line)

STOP_TIMES_TRIP_ID = 0
STOP_TIMES_STOP_ID = 3

sampled_stops = set()
with open("stop_times.txt") as input:
    with open("stop_times.filtered.txt", "w") as output:
        output.write(next(input))
        for line in input:
            columns = line.split(",")
            if columns[STOP_TIMES_TRIP_ID] in sampled_trips:
                sampled_stops.add(columns[STOP_TIMES_STOP_ID])
                output.write(line)

with open("stops.txt") as input:
    with open("stops.filtered.txt", "w") as output:
        output.write(next(input))
        for line in input:
            if line.split(",")[0]  in sampled_stops:
                output.write(line)


