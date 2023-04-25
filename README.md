# b6

Bedrock, or [b6](https://diagonal.works/b6), is Diagonal's geospatial analysis
engine. It reads a structured representation of the world into memory, and
makes it available for analysis, for example from a Python script or iPython
notebook. It also provides a rudimentary web interface for exploring the
available data. Communication between Python and b6 happens over
[GRPC](https://grpc.io).

We use b6 for the analysis behind our
[work for clients](http://diagonal.works/journal). Typically, we generate
results for different scenarios as JSON, and load them into bespoke
visualisations built with [d3](https://d3js.org/). We also directly embed
the [Go library](src/diagonal.works/b6/world.go) into custom
binaries to build tools that support dynamic analysis.

Working in the urban environment, the impact of the projects we support on
the communities around them outlasts our own involvement. We have a duty to
allow these communities to understand and build on our work beyond our time
with the project. We open sourced b6 to enable this - and to comply with our
[charter](http://diagonal.works/charter), which requires transparency.

We built b6 for ourselves, and don't expect it to be useful for a wide range
of people. We've also yet to write as much public documentation as we'd like.
However, if you do find it useful, or interesting, we'd be excited to [hear
about what you're up to](mailto:hello@diagonal.works).

## Quickstart

We provide a docker package for the b6 backend. Running:

```
docker run -p 8001:8001 -p 8002:8002 europe-docker.pkg.dev/diagonal-public/b6/b6
```

Will start an instance of the backend, with a web interface on port 8001, and
a GRPC interface for analysis from Python on port 8002, hosting a small amount
of data from OpenStreetMap for the area of London around
[Diagonal's spiritual home](https://www.dishoom.com/kings-cross/). Viewing
[localhost:8001](http://localhost:8001) should show you a map.

To try out analysis, you'll need to install the Python client library, via:

```
python -m pip install diagonal_b6
```

Right now, the best (and only) documentation for analysis functionality are the [unit tests](python/diagonal_b6/b6_test.py).

In the web interface, a left click will recenter the
map. Pressing backtick will drop down a terminal window. Shift click will
tell you about the geographic feature at that location. Entering a lat, lng
in the terminal will jump you to that location. Entering
`find [#amenity=cafe] | take 10 | show` will highlight 10 cafe, as the terminal
supports all the functionality available through Python. We've yet to properly
document it publicly, however.

## Building and running from source

We depend on the [protocol buffer](https://protobuf.dev/) compiler, and
[npm](https://www.npmjs.com/) at build time. To ingest and reproject data from
shapefiles, we depend on [gdal](https://gdal.org/), though it's not required
when working with OpenStreetMap, and we don't use it at run time. To install these on an Ubuntu based system, for example:
```
apt-get install npm protobuf-compiler gdal-bin libgdal-dev
```
or on OSX, using [brew](http://brew.sh):
```
brew install protobuf gdal
```
We also require Python >= 3.10 and Go >= 1.20.

To build all binaries, including data ingestion with gdal:
```
make
```

To build just the b6 backend, and OpenStreetMap ingestion, without gdal:
```
make b6 b6-ingest-osm
```
To generate the sample world data included in the docker image (replace the platform as appropriate):
```
bin/linux/x86_64/b6-ingest-osm --input=data/tests/camden.osm.pbf --output=data/camden.index
```
To start the backend:
```
bin/linux/x86_64/b6 --world=data/camden.index
```
You can run the entire build inside a docker container with:
```
docker build --build-arg=TARGETOS=linux --build-arg=TARGETARCH=amd64 -f docker/Dockerfile
```
## Ingesting data

If you have a small amount of data you'd like to work with, in OSM PBF format,
you can read it directly by putting in a directory by itself and replacing the
`/world` directory in the image:

```
docker run -v /path/with/data:/world -p 8001:8001 -p 8002:8002 europe-docker.pkg.dev/diagonal-public/b6/b6
```

For larger datasets, or datasets in formats other than OSM PBF, you'll need to
use one of the [ingestion tools](src/diagonal.works/b6/cmd) from either the
docker image. These tools convert source data into a compact representation for
efficient reading by the backend, that we call an index. `b6-ingest-osm` produces
an index for OpenStreetMap data, while `b6-ingest-gdal` will produce an index for
anything the GDAL library can ready. Our ingest tools aren't documented
publicly yet.


