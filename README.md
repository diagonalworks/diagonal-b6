# b6

Bedrock, or [b6](https://diagonal.works/b6), is Diagonal's geospatial analysis
engine. It reads a compact representation of the world into memory, and makes
it available for analysis, for example from a Python script or iPython
notebook. It also provides a simple web interface for exploring data.
Communication between Python and b6 happens over [gRPC](https://grpc.io).

We use b6 for the analysis behind our
[work for clients](http://diagonal.works/journal). We use the web interface
to explore data at the outset of a project, before generating analysis
results for different scenarios as JSON. We load these results into
interactive visualisations and tools built with [d3](https://d3js.org/).
When working with larger datasets, or building more complex tools, we embed
the [Go library](src/diagonal.works/b6/world.go) into custom
binaries to build tools that support dynamic analysis.

Working in the urban environment, the impact of the projects we support on
the communities around them outlasts our own involvement. We have a duty to
allow these communities to understand and build on our work beyond our time
with the project. We open sourced b6 to enable this - and to comply with our
[charter](http://diagonal.works/charter), which requires our work to be
 transparent.

We built b6 for ourselves, and don't expect it to be useful for a wide range
of people. We don't have as much public documentation as we'd like yet.
However, if you do find it useful, or interesting, we'd be excited to [hear
about what you're up to](mailto:hello@diagonal.works).

## Quickstart

The simplest way to try b6 is with the docker image we provide:

```sh
# Clone the repo
git clone https://github.com/diagonalworks/diagonal-b6.git

# Run the docker image and point it to some test data
docker run \
  -p 8001:8001 \
  -p 8002:8002 \
  -v ./data:/data \
  -e FRONTEND_CONFIGURATION="frontend-dev" \
  ghcr.io/diagonalworks/diagonal-b6:latest \
  --world /data/tests/camden.osm.pbf
```

> [!note]
> We provide a specific environment variable, `FRONTEND_CONFIGURATION`, to
> select the features we want to enable; in this case we want the _shell_
> feature to be on, but the _scenarios_ feature off. You can read more above
> these in the [/nix/js.nix](./nix/js.nix) file and
> [/nix/docker.nix](./nix/docker.nix).

This starts an instance of b6, with a web interface on port 8001, and
a gRPC interface for analysis from Python on port 8002, hosting a small amount
of data from OpenStreetMap for the area of London around
[Diagonal's spiritual home](https://www.dishoom.com/kings-cross/). Viewing
[localhost:8001](http://localhost:8001) should show you a map.

You can also run b6 directly via Nix:

```sh
nix run github:diagonalworks/diagonal-b6#run-b6 \
  -- \
  --world data/tests/camden.osm.pbf
```

<!--

TODO: Get this package deployed again; and/or update instructions on how to
use with the Nix template.

To try out analysis, you'll need the Python client library. For the omment,
the most convenient way is through the Nix shell, or via a

```
python -m pip install diagonal_b6
```
-->

There's a [Python notebook](python/docs/01_Search.ipynb) that introduces the
client library, and an overview b6 concepts in our
[FOSS4G 2023 talk](https://diagonal.works/foss4g-2023). The
[unit tests](python/diagonal_b6/b6_test.py) are the best place to discover the
functions b6 provides, and see how they're used.

## The web interface, and the b6 shell

![The b6 web interface](python/docs/b6-screenshot.jpg)

In the web interface, a left click will show a result window for the lat, lng
of that location. Holding shift while left clicking will show a result window
with the feature rendered on the map at that location. The result window can
be dragged if you'd like to keep it around, otherwise it will be replaced by
the next click.

There's an input box labelled _b6_ at the bottom of each result window. We call
this the b6 shell. The shell lets you enter a b6 function to run on the value
shown in the result window. For example, clicking a point on the map to show
the lat, lng and entering `sightline 300` will show an estimated viewshed
polygon from that point, with a cutoff of 300m. The Python client library and
the shell provide the same set of functions.

Pressing backtick will slide open a b6 shell that's not associated with a
result. This is a good starting point for jumping to new locations and finding
data. Entering `51.537028, -0.128169` will jump the map to a pocket park.

You can search for features by tags that start with a `#`. For example,
`find (tagged "#amenity" "bench")` will show places to sit. `find` returns
feaures matching a query, and `tagged` returns a query that will match
features by tag value. As searching is so common, the shell provides a shorthand
for queries: `find [#amenity=bench]`. `find` will highlight the matching
features on the map, if they're shown. A result window for a tag will add
features with that tag to the map, so if you'd like to see benches, enter
`#amenity=bench` (which is shorthand for `tag "#amenity" "bench"`), and drag
the window to keep it around. (To close windows, you have to reload the UI -
it's a work in progress!).

We often restrict searches to a radius around a location.
`find (and (intersecting-cap 51.537028, -0.128169 500) [#building])` will
return buildings within 500m of the park.

Nesting brackets in the shell quickly becomes tedious, so we provide a shorthand
for piplining functions with `|`, which calls the next function the result of
the current call as the first argument. `take (find [#amenity=bench]) 10`, which
returns the first 10 benches ordered by ID, can be written as
`find [#amenity=bench] | take 10`. When you use the shell at the bottom of
a result window, you're adding to a pipeline that starts with the result in the
window.

## Ingesting data

If you have a small amount of data you'd like to work with, in OSM PBF format,
you can read it directly by putting it in a directory by itself and making
sure that directory is readable. Via docker:

```sh
# Run the docker image and point it to some test data
docker run \
  -p 8001:8001 \
  -p 8002:8002 \
  -v ./data:/data \
  -e FRONTEND_CONFIGURATION="frontend-with-scenarios=false,shell=true" \
  ghcr.io/diagonalworks/diagonal-b6:latest \
  --world /data
```

For larger datasets, or datasets in formats other than OSM PBF, you'll need to
use one of the [ingestion tools](src/diagonal.works/b6/cmd). These tools
convert source data into a compact representation for efficient reading by the
backend, that we call an index. `b6-ingest-osm` produces an index for
OpenStreetMap data in PBF format, while `b6-ingest-gdal` will produce an index
for shp or geojson files read via the GDAL library. We typically use the
`.index` extension for ingested data.

To ingest an OSM PBF file `granary-square.osm.pbf`, use:

```sh
# Docker
docker run \
  -v ./data:/data \
  --entrypoint b6-ingest-osm \
  ghcr.io/diagonalworks/diagonal-b6:latest \
  --input data/tests/granary-square.osm.pbf --output data/granary-square.index

# Nix
nix run .#b6-ingest-osm -- \
      --input data/tests/granary-square.osm.pbf \
      --output granary-square.index
```

To ingest a shapefile via GDAL, use something like:

```
b6-ingest-gdal \
    --input SG_DataZone_Bdry_2011.shp \
    --output data/region/scottish-borders/data-zones-2011.index \
    --namespace maps.scot.gov/data-zone-2011 \
    --id DataZone \
    --id-strategy strip \
    --add-tags "#boundary=datazone" \
    --copy-tags "name=Name,code=DataZone,population:2011=TotPop2011"
```

In this example:

  * `--input` is the name of the GDAL readable file to ingest.
  * `--namespace` is the namespace to use when generating IDs for features.
  * `--id DataZone --id-strategy strip` uses the value of the `DataZone` field
     as the integer part of the feature's identifier, stripping non-numeric
     characters from the field's value. Another common option is
     `--id=Code --id-strategy=hash`, which hashes the value of the `Code`
    field. If `--id-strategy` isn't supplied, an ID is generated by
    incrementing an integer for each ingested feature.
  * `--copy-tags "name=Name"` copies the `Name` field into the feature as a
    tag named `name`.
  * `--add-tags "#boundary=datazone"` adds the tag `#boundary=datazone` to all
    features.

Indexing large inputs takes time and memory, but results in a reasonably sized
index file. For example, indexing a 10Gb planet extract for the UK takes ~20
minutes on a machine with 8 cores, and uses ~40Gb RAM. The resulting index is
around 10Gb. As the index is mapped directly into memory, the size of the file
is the upper bound on the amount of memory b6 will use to read the index. We
normally ingest large datasets on cloud VMs, but use the index on our own
machines.

## Building and running from source

The best way to build the project locally is via the Nix development shells.

A [Nix](https://nixos.org/) [flake](https://nixos.wiki/wiki/flakes) is
provided to build the binaries and/or do development.

If you use [direnv](https://direnv.net/) you will get a development shell
automatically, and otherwise you can get one with:

```shell
nix develop
```

You can build all the go binaries with `nix build` and run specific binaries,
such as `b6` or `b6-ingest-gdal` like so:

```sh
nix run .#b6 -- --help
nix run .#b6-ingest-gdal -- --help
```

> [!note]
> We provide a helpful nix package, `nix run .#run-b6` that pre-defines a few
> of a common command-line options so you don't need to set them explicitly.
> See the [flake.nix](./flake.nix) for more information.

The go application is built with
[gomod2nix](https://github.com/nix-community/gomod2nix/).

For day-to-day development, it is convenient to use the Makefile; so you can
run `make b6`, or any other make target, from the Nix shell. Note that the
resulting binaries are placed in the `./bin` folder.

Because the Python library depends on _running_ the Go binaries, we have
provided a special `combined` shell that can be used to run both:

```sh
nix develop .#combined
```

This is useful for running the Python _and_ Go tests via the Makefile:

```sh
> nix develop .#combined
> make all-tests
```

There is a Python project defined which can be built with `nix build
.#python312`; but this is only useful as a flake input to another project, and
it used for the flake template (see below). You can jump into a Python
environment with `nix develop .#python`.

> [!Important]
> The version of Python you use must match when you bring in the library; i.e.
> if you python312 you need to use the `python312` package.

#### Running with Nix

It is also possible to run the go binaries directly with nix:

```shell
> nix run github:diagonalworks/diagonal-b6/#b6

# or

> nix run github:diagonalworks/diagonal-b6/#b6-ingest-osm
```

#### Updating go dependencies

If the go dependencies change, you need to run `gomod2nix`. You can do this
from the normal devShell:

```shell
cd src/diagonal.works/b6/
gomod2nix
```

#### Building the docker image with Nix

You can build the docker image with Nix:

```shell
nix build .#b6-image
./result | docker load
```

This provides the docker image `b6`, which can be run in the typical way:

```shell
docker run -p 8001:8001 -p 8002:8002 -v ./data:/data b6 -world /data/camden.index
```

See the [flake.nix](flake.nix) file for more information.


#### Flake template

There is a flake template to demonstrate how to make a b6 "client" project;
i.e. something that comes with the b6 Python library installed. You can use it
like so:


```shell
mkdir some-project
cd some-project
nix flake init --template github:diagonalworks/diagonal-b6/
nix develop
```

Within here you will have a `b6` binary to run a local b6, and also the
`diagonal_b6` python package available.

You can read more about the template here:
[nix/python-client/flake.nix](./nix/python-client/flake.nix)

#### Running the docs with nix

To just open the docs:

```shell
nix run .#b6-docs
```
