---
sidebar_position: 1
---

# Nix

All the b6 go executables are able to be run via various `nix run ...`
invocations; for example `nix run .#b6 -- -world some.index`. Note that you
need to use the `--` to start the input of the arguments to the given
executable.

> :::tip
> You can use the argument `--print-out-paths` to get nix to
> print the output of a particular build. This can be particularly
> handy for the frontend:
> ```shell
> nix run .#b6 -- \
>   -static-v2=$(nix build .#frontend --print-out-paths) \
>   -enable-v2-ui --world data/camden.index
> ```
>
> Will build the frontend (from the present source) and also the `b6`
> executable, and run them together.


## Direnv

TODO

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

### Build the go binaries

```shell
nix build .
```

### Run <tt>b6-connect</tt>

```shell
nix build .#b6-connect -- \
      -input some.index \
      -output some.connected.index
```

### Running the frontend with vite

This also for hot-reloading/fast frontend development, in combination with a
backend hosting a particular dataset.

```shell
# In one folder
cp frontend && npm run start

# In other folder
nix run .#b6 -- \
      -static-v2=./frontend/dist \
      -enable-v2-ui \
      -enable-vite \
      -world data/camden.index
```
