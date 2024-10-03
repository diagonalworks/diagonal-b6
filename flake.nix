{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

    # Latest 3.7.1 release from nixpkgs
    # https://github.com/NixOS/nixpkgs/commits/nixpkgs-unstable/pkgs/development/libraries/gdal/default.nix
    gdalNixpkgs.url = "github:NixOS/nixpkgs/a93ab55b415e8c50f01cb6c9ebd705c458409d57";

    gomod2nix = {
      url = "github:nix-community/gomod2nix";
      inputs.nixpkgs.follows = "nixpkgs";
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
    , ...
    }:
    flake-utils.lib.eachDefaultSystem
      (system:
      let
        pkgs = import nixpkgs {
          inherit system;
          overlays = [ overlay ];
          # crossSystem.config = "aarch64-linux";
        };

        aarch64Pkgs = import nixpkgs {
          inherit system;
          overlays = [ overlay ];
          crossSystem.config = "aarch64-linux";
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
                version = "1.65.1";
                src = pkgs.fetchPypi {
                  pname = "grpcio";
                  inherit version;
                  hash = "sha256-PEkjAZiM1yDNFF2E4XMY1FrzQuKe+TFBIo+c1zIiNos=";
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
                  owner = "silky";
                  repo = "s2sphere";
                  rev = "d1d067e8c06e5fbaf0cc0158bade947b4a03a438";
                  sha256 = "sha256-6hNIuyLTcGcXpLflw2ajCOjel0IaZSFRlPFi81Z5LUo=";
                };
              };
            };
          };
        };

        b6-py = import ./nix/python.nix {
          inherit
            pkgs
            python
            b6-go-packages
            pyproject-nix
            ;
        };

        pythonEnv = python.withPackages (ps:
          [
            b6-py

            # For `make python`
            ps.grpcio-tools

            # For hacking
            ps.jupyter
          ]);


        # Use a pinned version of gdal.
        ourGdal = (import gdalNixpkgs { inherit system; }).gdal;

        b6-js-packages = import ./nix/js.nix {
          inherit
            pkgs
            ;
        };

        b6-go-packages = import ./nix/go.nix {
          inherit
            pkgs
            system
            ourGdal
            gomod2nix
            ;
        };

        make-b6-image = import ./nix/docker.nix
          {
            inherit
              pkgs
              ;
          } // b6-js-packages;
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
            go_1_21
            gotools
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
          default = b6-go-packages.everything;

          # Add an explicit 'go' entrypoint for the full go build+test.
          go = b6-go-packages.everything;

          # Not an application; but can be built `nix build .#python312`.
          python312 = b6-py;
          # TODO:
          # python311 = ...;

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
          #     FRONTEND_CONFIGURATION="frontend-with-scenarios=true" \
          #     -p 8001:8001 \
          #     -p 8002:8002 \
          #     -v ./data:/data \
          #     b6 \
          #     -world /data/camden.index
          #
          # to enable a specific frontend configuration.
          b6-image = make-b6-image "b6" b6-go-packages.everything;
          b6-minimal-image = make-b6-image "b6-minimal" b6-go-packages.go-executables.b6;
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
        // b6-go-packages.go-executables

        # All the frontend packages
        #
        # Examples:
        #
        #   nix build .#b6-js
        #   nix build .#frontend-with-scenarios=false
        #   nix build .#frontend-with-scenarios=true
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
