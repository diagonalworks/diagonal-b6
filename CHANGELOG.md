## v0.0.3: June 2023

* Redesigned the web interface, adding the ability to show
  multiple expression results at the same time, add features to
  the basemap by tag, and highlight features returned from expressions.
* Added reachability by public transit.
* Added elevation penaltites to walking reachability.

## v0.0.2: May 2023

* Add a version number to the b6 binary, and to index files.
* Use a memory mapped hashtable for search index tokens. Invalidates the
  previous index format, detected by using a new magic header value.
* Read large world indices using memory allocated by mmap.

## v0.0.1: May 2023

* Skipped to synchronise version numbers with pypi.org

## v0.0.0: April 2023

* Our first open source release, for the [Geospatial Systems CDT](https://geospatialcdt.ac.uk/) Challenge Week 2023.