# Frontend

## Map styling

Almost all the map rendering rules are controlled by the map style json
function, defined here:
[diagonal-map-style.json on GitHub](https://github.com/diagonalworks/diagonal-b6/blob/main/frontend/src/assets/map/diagonal-map-style.json).

This, in combination with the rendering rules defined in the go backend -
[renderer.go on GitHub](https://github.com/diagonalworks/diagonal-b6/blob/main/src/diagonal.works/b6/renderer/renderer.go#L171)
almost entirely controls what layers are displayed in what colour and when.

The caveats are GeoJSON layers, which can be returned from the [gRPC API
occasionally](https://github.com/diagonalworks/diagonal-b6/blob/main/src/diagonal.works/b6/ui/ui.go#L644),
and other custom functionality that may be defined in the relevant [Map frontend code](https://github.com/diagonalworks/diagonal-b6/blob/main/frontend/src/components/Map.tsx#L100).

## Feature flags with Vite

The frontend has a couple of "features" that are enabled/disabled with
environment variables.

The features correspond to folder names in the [`features`
folder](https://github.com/diagonalworks/diagonal-b6/tree/main/frontend/src/features); at present:

- `scenarios`
- `shell`

When running vite manually (i.e. via `pnpm dev`) you can set them like so:

```shell
VITE_FEATURES_SHELL=true pnpm dev
```

### Nix

When building with nix, we define a specific [feature
matrix](https://github.com/diagonalworks/diagonal-b6/blob/main/nix/js.nix#L24)
which contains all the relevant deriations; so you can build a version like:

```shell
nix build .#frontend-with-scenarios=true,shell=false
# Or
nix build .#frontend-with-scenarios=false,shell=false
```

### Docker

The [docker images use an _environment variable_
again](https://github.com/diagonalworks/diagonal-b6/blob/main/nix/docker.nix#L20),
to select the appropriate source package; so in your docker definition you
would set something like:

```shell
FRONTEND_CONFIGURATION="frontend-with-scenarios=true,shell=true"
```

in your docker environment.

## The Shell

## Tiles
