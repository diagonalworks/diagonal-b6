## v0.2.3: Jan 2025

* Publishing packages on PyPI via CI.
* Reinstate test related to GeoJSON parsing in Go.

Released from Edinburgh, UK.

## v0.2.2: Jan 2025

* Render `amenity` layers similarly to `building` layers.
* Return GeoJSON if a `b6.feature` contains it geometry; this allows
  collections of features to added to the UI to be highlighted when clicked..
* Draw _points_ in the colour specified by `b6:colour`.

Released from Edinburgh, UK.

## v0.2.1: Dec 2024

* Change the backend to return "header" lines and normal lines.
* Add a "Target" function to move the UI to focus on a given feature.
* Reinstate the Copy/Show/Hide buttons on the header elements
* Introduce a simple Nix entrypoint, `.#run-b6-dev`, for running the backend
  in dev mode.

Released from Edinburgh, UK.

## v0.2.0: Nov 2024

* Add `is-valid` to drop invalid features.
* Add `get-centroid` to compute the centroid of any feature.
* Fixes to the Nix flake setup.

Released from Edinburgh, UK.

## v0.1.0: Nov 2024

* Move to first non-minor version so we can use minor versions!

Released from Edinburgh, UK.

## v0.0.4: July 2023

* Added merge-changes, parse-geojson-file.
* Added support for collection literals in Python and the b6 shell.
* Batch transformations in b6-ingest-gdal to support large files.
* Support underscores in tag keys.

Released from Pula, Croatia.

## v0.0.3: June 2023

* Redesigned the web interface, adding the ability to show
  multiple expression results at the same time, add features to
  the basemap by tag, and highlight features returned from expressions.
* Added reachability by public transit.
* Added elevation penaltites to walking reachability.

Released from Prizren, Kosovo.

## v0.0.2: May 2023

* Add a version number to the b6 binary, and to index files.
* Use a memory mapped hashtable for search index tokens. Invalidates the
  previous index format, detected by using a new magic header value.
* Read large world indices using memory allocated by mmap.

Released from London, UK.

## v0.0.1: May 2023

* Skipped to synchronise version numbers with pypi.org

## v0.0.0: April 2023

* Our first open source release, for the [Geospatial Systems CDT](https://geospatialcdt.ac.uk/) Challenge Week 2023.

Released from Nottingham, UK.
