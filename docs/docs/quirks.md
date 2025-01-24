---
sidebar_position: 5
---

# Quirks

- The `b6.collection(*[ ... ])` limitations. This seems to break, in
  combination with `add_tags`; it seems to drop if the list has more than 60k
  elements.

- Searching for things added with `add_world_with_change( ... )`; the default
  connection uses the default `root` world. You'd need to open another
  connection with the specific world you are after.


