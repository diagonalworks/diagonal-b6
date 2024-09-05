{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/24.05";
    unstable.url = "github:NixOS/nixpkgs/nixos-unstable";

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
    , unstable
    , pyproject-nix
    , flake-utils
    , treefmt-nix
    , gdalNixpkgs
    , ...
    }:
    flake-utils.lib.eachDefaultSystem (system:
    let
      # Python setup
      overlay = _: prev: {
        python3 = prev.python3.override {
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

      python = pkgs.python3;

      # From <https://github.com/NixOS/nixpkgs/blob/master/doc/languages-frameworks/javascript.section.md#pnpm-javascript-pnpm>
      frontend = pkgs.stdenv.mkDerivation (finalAttrs: {
        pname = "b6-frontend";
        version = "0.0.0";

        src = ./frontend;

        nativeBuildInputs = [
          pkgs.nodejs
          unstablePkgs.pnpm.configHook
        ];

        pnpmDeps = unstablePkgs.pnpm.fetchDeps {
          inherit (finalAttrs) pname version src;
          hash = "sha256-8Kc8dxchO/Gu/j3QSR52hOf4EnJuxzsaWMy9kMNgOCc=";
        };

        # Override the phases as there is already a Makefile present, which is
        # used by Nix by default.
        buildPhase = ''
          pnpm build
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
        npmDepsHash = "sha256-qzMHjOVRINRZzeTdabz2u+75QrhULS1YzvdeDzWNwLs=";

        buildPhase = ''
          npm run build
        '';

        installPhase = ''
          mkdir $out
          mv bundle.js $out
        '';
      };

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

      # A derivation to _only_ build the 'b6' program; helps to keep the
      # docker image small.
      b6-go-only-b6 = with pkgs; gomod2nix.legacyPackages.${system}.buildGoApplication {
        name = "b6";
        src = ./src/diagonal.works/b6;
        buildInputs = [
          ourGdal
        ];
        nativeBuildInputs = [
          pkg-config
        ];
        subPackages = [ "cmd/b6" ];
        doCheck = false;
        pwd = ./src/diagonal.works/b6;
      };

      # Run like:
      # docker run -p 8001:8001 -p 8002:8002 -v data:/data b6 -world /data/camden.index
      b6-image = pkgs.dockerTools.streamLayeredImage {
        name = "b6";
        tag = "latest";
        created = "now";
        contents = [
          pkgs.busybox
        ];
        config = {
          ExposedPorts = {
            "8001" = { };
            "8002" = { };
          };
          Entrypoint = [
            "${b6-go-only-b6}/bin/b6"
            "-http=0.0.0.0:8001"
            "-grpc=0.0.0.0:8002"
            "-js=${b6-js.outPath}"
            "-enable-v2-ui"
            "-static-v2=${frontend.outPath}"
            # Note: We don't specify any '-world' parameter and instead
            # force people to provide it.
            # "-world=/world"
          ];
        };
      };

      pkgs = import nixpkgs { inherit system; overlays = [ overlay ]; };
      unstablePkgs = import unstable { inherit system; };
    in
    {
      # Development shells for hacking/building with the Makefile
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
          unstablePkgs.pnpm # Need version 9

          # Other
          osmium-tool # Extract OSM files
        ];

        shellHook = ''
          export PYTHONPATH=''$(pwd)/python
        '';
      };

      # Note: We have a separate Python development shell because this _also_
      # requires the Go package completely built, which makes it quite
      # inconvenient for actual go hacking (i.e. if you change go.mod and
      # haven't yet run gomod2nix, for example.)
      devShells.python = pkgs.mkShell {
        packages = [
          # Python hacking
          pythonEnv
        ];
      };

      packages = {
        # Run like `nix run . -- --help` or access all the binaries with
        # `nix build` and look in `./result/bin`.
        default = b6-go;

        go = b6-go;

        # Not an application; but can be built `nix build .#python`.
        python = b6-py;

        b6-image = b6-image;

        frontend = frontend;

        b6-js = b6-js;
      };


      # Run via `nix fmt`
      formatter =
        let
          fmt = treefmt-nix.lib.evalModule pkgs (_: {
            projectRootFile = "flake.nix";
            programs.nixpkgs-fmt.enable = true;
          });
        in
        fmt.config.build.wrapper;
    });
}
