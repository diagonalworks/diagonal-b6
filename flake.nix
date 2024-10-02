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
        pkgs = import nixpkgs { inherit system; overlays = [ overlay ]; };

        # Python setup
        python = pkgs.python312;

        overlay = _: prev: {
          python312 = prev.python312.override {
            packageOverrides = _: p: {
              # Note: We have to refer to a spceific version here to make sure
              # grpcio _and_ grpcio-tools match.
              #
              # At present on unstable they are at _different_ versions, and
              # this causes a warning. This can be removed if we pin to a
              # newer version of a stable nixpkgs, or when unstable has these
              # two packages at the same version.
              grpcio = p.grpcio.overridePythonAttrs(old: rec {
                version = "1.65.1";
                src =  pkgs.fetchPypi {
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

        # We have to write the version into here from the output of the go
        # binary.
        pyproject-file = (pkgs.runCommand "make-pyproject" { } ''
          substitute ${./python/pyproject.toml.template} $out \
            --subst-var-by VERSION ''$(${b6-go}/bin/b6-api --pip-version)
        '');

        pythonProject = pyproject-nix.lib.project.loadPyproject {
          projectRoot = ./python;
          pyproject = pkgs.lib.importTOML pyproject-file;
        };

        renderedPyProject = pythonProject.renderers.buildPythonPackage {
          inherit python;
        };

        b6-py = python.pkgs.buildPythonPackage (renderedPyProject // {
          # Set the pyproject to be the one we computed via our b6 binary.
          patchPhase = ''
            cat ${pyproject-file} > pyproject.toml
          '';

          nativeBuildInputs = renderedPyProject.nativeBuildInputs ++ [
            python.pkgs.grpcio-tools
          ];

          # A couple of hacks necessary to build the proto files and the API.
          preBuild = ''
            # Bring in the necessary proto files and the Makefile
            cp -r ${./proto} ./proto
            cat ${./Makefile} > some-Makefile

            # Hack: Run the b6-api command outside of the Makefile, using the
            # Nix version of the binary.
            ${b6-go}/bin/b6-api --functions | python diagonal_b6/generate_api.py > diagonal_b6/api_generated.py

            # Hack: Make the directory structure that the Makefile expects,
            # then move things to where we want them
            mkdir python
            mkdir python/diagonal_b6

            make proto-python -f some-Makefile

            # Cleanup
            mv python/diagonal_b6/* ./diagonal_b6
            rm -rf python/diagonal_b6
          '';

          pythonImportsCheck = [ "diagonal_b6" ];
        });


        pythonEnv = python.withPackages (ps:
          [
            b6-py

            # For `make python`
            ps.grpcio-tools

            # For hacking
            ps.jupyter
          ]);


        # The default frontend configuration.
        frontend = mkFrontend {
          VITE_FEATURES_SCENARIOS = false;
        };

        # Note: We obtain a list of all the "features" (i.e. the folders on this
        # specific directory) and use that to construct a set of derivations
        # that turn on/off every feature combination.
        #
        # Example:
        #
        #   nix build .#frontend-with-scenarios=false
        #   nix build .#frontend-with-scenarios=true
        #
        frontend-feature-matrix =
          let
            allPaths = builtins.readDir ./frontend/src/features;
            onlyDirs = pkgs.lib.attrsets.filterAttrs (_: v: v == "directory") allPaths;

            # [ scenarios ... ]
            featureNames = builtins.attrNames onlyDirs;

            # [ ABC, XYZ, ... ]
            viteEnvVars = map (x: "VITE_FEATURES_${pkgs.lib.toUpper x}") featureNames;

            # { ABC = [ "true" false ];
            #   XYZ = [ "true" false ];
            #   ...
            # }
            # Note: Because of the way the JS code checks for the value of the
            # flag, this has to be the _string_ "true".
            allOptions = builtins.listToAttrs (map (x: { name = x; value = [ "true" false ]; }) viteEnvVars);

            # [ { ABC = "true"; XYZ = "true" }
            # , { ABC = "true"; XYZ = false  }
            # , ...
            # ]
            configurations = pkgs.lib.cartesianProduct
              allOptions
            ;

            # { "abc=true,xyz=true"  = mkFrontend { ABC = "true"; XYZ = "true" };
            # , "abc=true,xyz=false" = mkFrontend { ABC = "true"; XYZ = false  };
            # , ...
            # }
            matrix =
              let
                nameFor = k: v:
                  let
                    featName = builtins.substring 14 (-1) (pkgs.lib.toLower k);
                    val = if v == false then "false" else toString v;
                  in
                  "${featName}=${val}";
                elements = map
                  (c: {
                    # "abc=true,xyz=true" ...
                    name =
                      let conf = pkgs.lib.strings.concatStringsSep "," (pkgs.lib.mapAttrsToList nameFor c);
                      in "frontend-with-${conf}";
                    value = mkFrontend c;
                  })
                  configurations;
              in
              builtins.listToAttrs elements;
          in
          matrix;

        # From <https://github.com/NixOS/nixpkgs/blob/master/doc/languages-frameworks/javascript.section.md#pnpm-javascript-pnpm>
        mkFrontend = envVars: pkgs.stdenv.mkDerivation (finalAttrs: {
          pname = "b6-frontend";
          version = "0.0.0";

          src = ./frontend;

          nativeBuildInputs = [
            pkgs.nodejs
            pkgs.pnpm.configHook
          ];

          pnpmDeps = pkgs.pnpm.fetchDeps {
            inherit (finalAttrs) pname version src;
            hash = "sha256-8Kc8dxchO/Gu/j3QSR52hOf4EnJuxzsaWMy9kMNgOCc=";
          };

          # Override the phases as there is already a Makefile present, which is
          # used by Nix by default.
          buildPhase = ''
            ${pkgs.lib.strings.toShellVars envVars} pnpm build
          '';

          installPhase = ''
            rm dist/index-vite.html
            mv dist/ $out
          '';
        });

        b6-js = pkgs.buildNpmPackage {
          pname = "b6-js";
          version = "v.0.0.0";
          src = ./src/diagonal.works/b6/cmd/b6/js;
          npmDepsHash = "sha256-w332KqVpqdZSoNjwRxAaB4PZKaMfnM0MHl3lcpkmmQU=";

          buildPhase = ''
            npm run build
          '';

          installPhase = ''
            mkdir $out
            mv bundle.js $out
          '';
        };

        # Use a pinned version of gdal.
        ourGdal = (import gdalNixpkgs { inherit system; }).gdal;

        # Go setup
        b6-go = with pkgs; gomod2nix.legacyPackages.${system}.buildGoApplication {
          name = "b6";
          src = ./src/diagonal.works/b6;
          buildInputs = [
            ourGdal
          ];
          nativeBuildInputs = [
            pkg-config
          ];

          # Bring in test data to the root directory; this is where it will be
          # found by the tests (see b6/test/data.go: and the 'testDataDirectory' function)
          preCheck = ''
            mkdir data
            mkdir data/tests
            cp -r ${./data/tests}/* ./data/tests
          '';

          doCheck = true;

          # Must be added due to bug https://github.com/nix-community/gomod2nix/issues/120
          pwd = ./src/diagonal.works/b6;
        };

        # A collection of derivations for each go cmd; can be useful to keep
        # closures small, if for example you only want to depend on a specific
        # binary.
        #
        # Examples:
        #
        # > nix run .#b6
        # > nix run .#b6-connect
        # > nix run .#b6-ingest-osm
        #
        go-executables =
          let
            allPaths = builtins.readDir ./src/diagonal.works/b6/cmd;
            onlyDirs = pkgs.lib.attrsets.filterAttrs (_: v: v == "directory") allPaths;
            cmds = builtins.attrNames onlyDirs;
            mkGoApp = cmd: with pkgs; gomod2nix.legacyPackages.${system}.buildGoApplication {
              name = "${cmd}";
              src = ./src/diagonal.works/b6;
              pwd = ./src/diagonal.works/b6;
              buildInputs = [
                ourGdal
              ];
              nativeBuildInputs = [
                pkg-config
              ];
              subPackages = [ "cmd/${cmd}" ];
              # Don't run the tests now; they've been tested by the main
              # derivation.
              doCheck = false;
            };
          in
          builtins.listToAttrs (map (n: { name = n; value = mkGoApp n; }) cmds);


        # Run like:
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
        #
        b6-image = name: b6-drv:
          let
            # The main entrypoint.
            #
            # Note: We don't specify any '-world' parameter and instead
            # force people to provide it.
            # "-world=/world"
            #
            # Note:
            #
            # We look to an environment variable, `FRONTEND_CONFIGURATION`, to
            # decide which frontend JavaScript build to depend on.
            #
            # For now these are manually defined; but we could imagine
            # computing them from the `frontend-feature-matrix`, if we wished.
            launch-script = pkgs.writeShellScriptBin "launch-b6" ''
              case "''${FRONTEND_CONFIGURATION}" in
                "frontend-with-scenarios=true")
                  STATIC_ARG=${frontend-feature-matrix."frontend-with-scenarios=true".outPath} ;;
                "frontend-with-scenarios=false")
                  STATIC_ARG=${frontend-feature-matrix."frontend-with-scenarios=false".outPath} ;;
                *)
                  STATIC_ARG=${frontend.outPath} ;;
              esac

              # So we can kill it with Ctrl-C
              _term() {
                kill "$child" 2>/dev/null
              }
              trap _term INT

              ${b6-drv}/bin/b6 \
                -http=0.0.0.0:8001 \
                -grpc=0.0.0.0:8002 \
                -js=${b6-js.outPath} \
                -enable-v2-ui \
                -static-v2=''$STATIC_ARG \
                "$@" &

              child=$!
              wait "$child"
            '';
          in
          pkgs.dockerTools.streamLayeredImage {
            name = "${name}";
            tag = "latest";
            created = "now";
            contents = [
              # For navigating around/debugging, if necessary
              pkgs.busybox
            ];
            config = {
              Labels = {
                "org.opencontainers.image.source" = "https://github.com/diagonalworks/diagonal-b6";
                "org.opencontainers.image.description" = "b6";
              };
              Env = [
                # Make sure all the b6 binaries are in the path.
                "PATH=/bin:/sbin:/usr/bin:/usr/sbin:/usr/local/bin:/usr/local/sbin:${b6-drv}/bin"
              ];
              ExposedPorts = {
                "8001" = { };
                "8002" = { };
              };
              Entrypoint = [ "${launch-script}/bin/launch-b6" ];
            };
          };
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
          default = b6-go;

          # Add an explicit 'go' entrypoint for the full go build+test.
          go = b6-go;

          # Not an application; but can be built `nix build .#python312`.
          python312 = b6-py;
          # TODO:
          # python311 = ...;

          # Docker images
          b6-image = b6-image "b6" b6-go;
          b6-minimal-image = b6-image "b6-minimal" go-executables.b6;

          inherit
            b6-js
            frontend
            ;
        }
        # All the frontend feature configurations
        // frontend-feature-matrix
        # All the explicit go executables
        // go-executables
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
          path = ./nix-templates/python-client;
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
