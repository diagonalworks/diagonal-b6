{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-24.11";

    # We're pinned to go version 1.22.6 for now.
    goNixpkgs.url = "github:NixOS/nixpkgs/nixos-24.05";

    # Pin to pnpm to 9.7.0, as there is an issue using the latest version of
    # pnpm in this project.
    # TODO: Consider upgrading this at some point.
    pnpmNixpkgs.url = "github:NixOS/nixpkgs/nixos-24.05";

    # Latest 3.7.1 release from nixpkgs
    # https://github.com/NixOS/nixpkgs/commits/nixpkgs-unstable/pkgs/development/libraries/gdal/default.nix
    gdalNixpkgs.url = "github:NixOS/nixpkgs/a93ab55b415e8c50f01cb6c9ebd705c458409d57";

    gomod2nix = {
      url = "github:nix-community/gomod2nix";
      inputs.nixpkgs.follows = "goNixpkgs";
    };

    pyproject-nix = {
      url = "github:nix-community/pyproject.nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };

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
    , pnpmNixpkgs
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

        pnpmpkgs = import pnpmNixpkgs {
          inherit system;
        };

        # Python setup
        python-version = "python312";
        python = pkgs.${python-version};

        overlay = _: prev: {
          ${python-version} = prev.${python-version}.override {
            packageOverrides = _: p: {
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
            ps.pandas
            ps.matplotlib
          ]);


        # Use a pinned version of gdal.
        ourGdal = (import gdalNixpkgs { inherit system; }).gdal;

        b6-js-packages = import ./nix/js.nix {
          inherit
            pkgs
            ;

          # Pin the version of pnpm
          pnpm = pnpmpkgs.pnpm;
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

          # We define a wrapped version of the b6 executable, using a particular
          # build of the frontend, for easy invocation, such as:
          #
          # > nix run .#run-b6 -- -world data
          run-b6 = pkgs.writeShellScriptBin "run-b6" ''
              ${packages.b6}/bin/b6 \
                -http=0.0.0.0:8001 \
                -grpc=0.0.0.0:8002 \
                -enable-v2-ui \
                -static-v2=${packages.frontend-dev.outPath} \
                "$@"
              '';

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
          # > docker run -p 8001:8001 -p 8002:8002 -v ./data:/data b6 --world /data/camden.index
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
        #   nix build .#frontend-with-scenarios=false,shell=true
        #   nix build .#frontend-with-scenarios=true,shell=false
        #   nix build .#b6-js # Old v1 UI
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
