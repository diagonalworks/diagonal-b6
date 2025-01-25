# Ingest: Reading data

"Ingesting" in the b6 ecosystem means taking in data in some format, and
computing it into b6's indexed format. The output is a `.index` file.

The best place to see the list of formats that can be read is here:
[b6/cmd on GitHub](https://github.com/diagonalworks/diagonal-b6/tree/main/src/diagonal.works/b6/cmd).

All the commands like `b6-ingest-*` are for ingesting.

The most important is `b6-ingest-osm`, to ingest data available directly from
OSM, and `b6-ingest-gdal`, which allows you to ingest shapefiles and GeoJSON
files.

See [Contributing#Downloading data](/docs/contributing#downloading-data) for
some other examples, and see the `--help` option in the various programs for
further details.

Example:

```shell
nix run .#b6-ingest-osm -- --help
nix run .#b6-ingest-gdal -- --help
```
