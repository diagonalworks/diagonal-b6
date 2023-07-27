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