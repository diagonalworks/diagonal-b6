{
  inputs = {
    # Note: We're pinned to 24.05 here as unstable doesn't contain the version
    # of go we are using; we'd need to upgrade the go dependency to be able to
    # move to unstable fully (or just use an old version of nixpkgs for go.)
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

    # We're pinned to 1.22.6 for now.
    goNixpkgs.url = "github:NixOS/nixpkgs/nixos-24.05";

    # Latest 3.7.1 release from nixpkgs
    # https://github.com/NixOS/nixpkgs/commits/nixpkgs-unstable/pkgs/development/libraries/gdal/default.nix
    gdalNixpkgs.url = "github:NixOS/nixpkgs/a93ab55b415e8c50f01cb6c9ebd705c458409d57";

    gomod2nix = {
      url = "github:nix-community/gomod2nix";
      inputs.nixpkgs.follows = "goNixpkgs";
    };

    pyproject-nix.url = "github:nix-community/pyproject.nix";
    flake-utils.url = "github:numtide/flake-utils";
    treefmt-nix.url = "github:numtide/treefmt-nix";
  };


  outputs =
    { self
    , nixpkgs
    , gomod2nix
    , pyproject-nix
    , flake-utils
    , treefmt-nix
    , gdalNixpkgs
    , goNixpkgs
    , ...
    }:
    flake-utils.lib.eachDefaultSystem
      (system:
      let
        pkgs = import nixpkgs {
          inherit system;
          overlays = [ overlay ];
        };

        gopkgs = import goNixpkgs {
          inherit system;
        };

        # Python setup
        python-version = "python312";
        python = pkgs.${python-version};

        overlay = _: prev: {
          ${python-version} = prev.${python-version}.override {
            packageOverrides = _: p: {
              # Note: We have to refer to a spceific version here to make sure
              # grpcio _and_ grpcio-tools match.
              #
              # At present on unstable they are at _different_ versions, and
              # this causes a warning. This can be removed if we pin to a
              # newer version of a stable nixpkgs, or when unstable has these
              # two packages at the same version.
              grpcio = p.grpcio.overridePythonAttrs (old: rec {
                version = "1.66.1";
                src = pkgs.fetchPypi {
                  pname = "grpcio";
                  inherit version;
                  hash = "sha256-NTNPnJdFrdPjV+M3J1b9MtklvVLEHal/Tf2vveC/DuI=";
                };
              });

              s2sphere = p.buildPythonPackage rec {
                version = "0.2.5";
                pname = "s2sphere";
                format = "pyproject";
                nativeBuildInputs = with p.pythonPackages; [
                  setuptools
                ];
                propagatedBuildInputs = with p.pythonPackages; [
                  future
                ];
                # The original repo <https://github.com/sidewalklabs/s2sphere>
                # is archived, so this refers to a fork.
                src = pkgs.fetchFromGitHub {
                  owner = "diagonalworks";
                  repo = "s2sphere";
                  rev = "d1d067e8c06e5fbaf0cc0158bade947b4a03a438";
                  sha256 = "sha256-6hNIuyLTcGcXpLflw2ajCOjel0IaZSFRlPFi81Z5LUo=";
                };
              };
            };
          };
        };

        b6-py = ourGdal: import ./nix/python.nix {
          inherit
            pkgs
            pyproject-nix
            ;
          b6-go-packages = b6-go-packages ourGdal;
        };

        pythonEnv = python.withPackages (ps:
          [
            (b6-py ourGdal python)

            # For `make python`
            ps.grpcio-tools

            # For hacking
            ps.jupyter
          ]);


        # Use a pinned version of gdal.
        ourGdal = (import gdalNixpkgs { inherit system; }).gdal;
        ourGdal-aarch64-linux = (import gdalNixpkgs {
          inherit system;
          crossSystem.config = "aarch64-linux";
        }).gdal;

        b6-js-packages = import ./nix/js.nix {
          inherit
            pkgs
            ;
        };

        b6-go-packages = ourGdal: import ./nix/go.nix {
          inherit
            system
            ourGdal
            gomod2nix
            ;
          pkgs = gopkgs;
        };

        make-b6-image = import ./nix/docker.nix {
          inherit
            pkgs
            b6-js-packages
            ;
        };

        b6-image = make-b6-image "b6" (b6-go-packages ourGdal).everything;
        b6-minimal-image = make-b6-image "b6-minimal" (b6-go-packages ourGdal).go-executables.b6;
      in
      rec {
        # Development shells for hacking/building with the Makefile. Note
        # that, importantly, this shell does _not_ contain the Python
        # derivations; so can't be used to run any 'Makefile' tasks relevant
        # to Python.
        #
        # If you need both, use the 'combined' shell:
        #
        # > nix develop .#combined
        #
        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            # Running the Makefile tasks
            ourGdal
            pkg-config
            protobuf
            protoc-gen-go
            protoc-gen-go-grpc

            # Go
            # Note: The vesion here should match the one in `go.mod`.
            gopkgs.go_1_22
            gopkgs.gotools
            gomod2nix.packages.${system}.default # gomod2nix CLI

            # JavaScript (docs/front-end)
            nodejs
            pnpm # Need version 9

            # Other
            osmium-tool # Extract OSM files
          ];

          shellHook = ''
            export PYTHONPATH=''$(pwd)/python
          '';
        };

        # Note: We have a separate Python development shell because this
        # _also_ requires the Go package completely built, which makes it
        # quite inconvenient for actual go hacking (i.e. if you change go.mod
        # and haven't yet run gomod2nix, for example, the shell will fail.)
        devShells.python = pkgs.mkShell {
          packages = [
            # Python hacking
            pythonEnv
          ];
        };

        # Finally, we have a combined devShell for building everything at once
        # (i.e. if you want to run `make all-tests`).
        devShells.combined = pkgs.mkShell {
          inputsFrom = with devShells; [
            default
          ];

          packages = [
            pythonEnv
          ];

          shellHook = ''
            export PYTHONPATH=''$(pwd)/python
          '';
        };

        packages = {
          # Run like `nix run . -- --help` or access all the binaries with
          # `nix build` and look in `./result/bin`.
          default = (b6-go-packages ourGdal).everything;

          # Add an explicit 'go' entrypoint for the full go build+test.
          go = (b6-go-packages ourGdal).everything;

          # Not an application; but can be built `nix build .#python312`.
          python312 = b6-py ourGdal python;

          # Docker images
          #
          # To build/load the resulting image:
          #
          # > nix build .#b6-image
          # > ./result | docker load
          #
          # To run:
          # > docker run -p 8001:8001 -p 8002:8002 -v ./data:/data b6 -world /data/camden.index
          #
          # or:
          #
          # > docker run -e \
          #     FRONTEND_CONFIGURATION="frontend-with-scenarios=true,shell=false" \
          #     -p 8001:8001 \
          #     -p 8002:8002 \
          #     -v ./data:/data \
          #     b6 \
          #     -world /data/camden.index
          #
          # to enable a specific frontend configuration.
          inherit
            b6-image
            b6-minimal-image
            ;
        }
        # All the explicit go executables
        #
        # Examples:
        #
        # > nix run .#b6
        # > nix run .#b6-connect
        # > nix run .#b6-ingest-osm
        #
        #
        // (b6-go-packages ourGdal).go-executables

        # All the frontend packages
        #
        # Examples:
        #
        #   nix build .#b6-js
        #   nix build .#frontend-with-scenarios=false,shell=true
        #   nix build .#frontend-with-scenarios=true,shell=false
        #
        // b6-js-packages
        ;

        # Run via `nix fmt`
        formatter =
          let
            fmt = treefmt-nix.lib.evalModule pkgs (_: {
              projectRootFile = "flake.nix";
              programs.nixpkgs-fmt.enable = true;
            });
          in
          fmt.config.build.wrapper;
      }
      ) // {

      # Nix templates
      templates = {
        default = {
          path = ./nix/templates/python-client;
          description = "A local setup for b6: python client + local b6 executable";
        };
      };
    };


  nixConfig = {
    extra-substituters = [
      "https://diagonalworks.cachix.org"
    ];
    extra-trusted-public-keys = [
      "diagonalworks.cachix.org-1:7U834B3foDCfa1EeV6xpyOz9JhdfUXj2yxRv0rAdYMk="
    ];
  };
}
