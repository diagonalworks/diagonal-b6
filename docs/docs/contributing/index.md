# Contributing

## Data

To do almost anything interesting, you will need to have some data.

The repository contains test data in the `./data/tests` folder.

### Downloading data

You will almost certainly want to work with your own data for your own area of
interest. The most convenient way to obtain data, from
[OpenStreetMap](https://www.openstreetmap.org/), is through
[geofabrik](https://www.geofabrik.de/).

#### Downloading `.osm.pbf` format

Docker is probably the easiest way:

```shell
docker run --rm -it \
         -v $PWD:/download openmaptiles/openmaptiles-tools \
         download-osm \
         geofabrik \
         scotland
         -- -d /download
```

This downloads all of Scotland in `.osm.pbf` format.

From here you should either run `b6-ingest` to compute the index, or, use
osmium to focus down to a small area of interest. You can do this by finding a
polygon file defining your region, [cities
polygons](https://github.com/JamesChevalier/cities/tree/master) - let's say
[Edinburgh](https://github.com/JamesChevalier/cities/blob/master/united_kingdom/scotland/city-of-edinburgh_scotland.poly)
- and then running [`osmium`](https://github.com/osmcode/osmium-tool) (which is provided in the default nix shell):

```shell
osmium extract \
         --polygon city-of-edinburgh_scotland.poly \
         -F pbf \
         scotland-latest.osm.pbf \
         -o edinburgh-latest.osm.pbf
```

Then you can just run `b6` on the resulting file

```shell
nix run .#run-b6-dev -- --world ./edinburgh-latest.osm.pbf
```

### Indexing and connecting

In order to do more interesting interesting analysis you will almost certainly
want to index the data, and connect the buildings to the road network (to
allow you to compute reachability.)

In order to two this, there are two `go` executables of interest; `b6-ingest`
and `b6-connect`.

#### Indexing: `b6-ingest`

There are actually a few `b6-ingest-*` executables, but given that we have
downloaded OSM data above, we will use `b6-ingest-osm`.

To run it:

```shell
nix run .#b6-ingest-osm -- \
      --input edinburgh-latest.osm.pbf \
      --output edinburgh-latest.index
```

This will produce `edinburgh-latest.index`, which again can be loaded
directly:

```shell
nix run .#run-b6-dev -- --world ./edinburgh-latest.index
```

#### Connecting: `b6-connect`

This tool takes in an index, and connects all the buildings to the network.

```shell
nix run .#b6-connect -- \
      --input edinburgh-latest.index \
      --output edinburgh-latest.connected.index
```

As usual, we can just load this directly:

```shell
nix run .#run-b6-dev -- --world ./edinburgh-latest.connected.index
```

## Hacking on the code

[Nix](https://nixos.org/) is used heavily to provide a consistent build
environment.

### Nix

All the b6 go executables are able to be run via various `nix run ...`
invocations; for example `nix run .#b6 -- -world some.index`. Note that you
need to use the `--` to start the input of the arguments to the given
executable.

> :::tip
> You can use the argument `--print-out-paths` to get nix to
> print the output of a particular build. This can be particularly
> handy for the frontend. For example,
>
> ```shell
> nix run .#b6 -- \
>   -static-v2=$(nix build .#frontend --print-out-paths) \
>   -enable-v2-ui --world data/tests/camden.osm.pbf
> ```
>
> will build the frontend (from the present source) and also the `b6`
> executable, and run them together.


### Direnv

For convenient hacking we recommend using [direnv](https://direnv.net/) (or
[nix-direnv](https://github.com/nix-community/nix-direnv)) to load the shell
when we enter the folder. The default shell allows you to build the `go`
binaries with the Makefile, and also provides nodejs to build the JavaScript
projects.

## Hacking on the frontend

[vite](https://vite.dev/) is used to do a hot-reloading style of development
for the frontend.

To get this running, you need to, of course, start the vite development server
in the frontend:

```shell
# In one folder
cp frontend && npm run start
```

and then, run the backend by pointing it to the local copy. For this purpose
there is a short nix entrypoint `#run-b6-dev`; used like so:

```shell
nix run .#run-b6-dev -- --world ./data/camden.index
```

Which just calls b6 with arguments to load the frontend from the location that
vite builds into: `./frontend/dist`.

> :::note
> Note that this builds the go binaries _through nix_, which will be slower
> than building them via the Makefile/go directly. I prefer this because it's
> less thinking; but you can choose the method you prefer.

## Tips and tricks

### The `combined` shell

The `combined` devShell is necessary for certain tasks, in particular it is
required for running the Python tests

```sh
make all-tests
```

To enter the shell, use:

```sh
nix develop .#combined
```

> :::note
> We introduce the `combined` shell because, otherwise, the normal development
> shell would need to build the b6 go code in order to load (because building
> the Python library depends on _running_ the `b6-api` executable).
>
> In a devShell, it is not appropriate to build the _project_ itself,
> otherwise the shell would be unloadable whenever the project doesn't build
> (i.e. you're making a change to the go code that prevents it from
> compiling). For this reason, we provide the combined shell.

### Build the go binaries

To build _all_ the go binaries, simply run a `nix build`:

```shell
nix build .
```

They will all be accessible in `./result/bin/`.

#### Build and run specific go executables

To run (and build) only a specific executable; say `b6-connect`, you
can do:

```shell
nix run .#b6-connect -- \
      --help
```

### Code Formatting

Run `nix fmt` to format all the code.

### Running the docs through Nix

There's a simple Nix derivation to _run_ the docs:

```shell
nix run .#b6-docs
```
