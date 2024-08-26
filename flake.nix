{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/24.05";
    unstable.url = "github:NixOS/nixpkgs/nixos-unstable";
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

      pyproject-file = (pkgs.runCommand "make-pyproject" { } ''
        substitute ${./python/pyproject.toml.template} $out \
          --subst-var-by VERSION ''$(${b6-go}/bin/b6-api --pip-version)
      '');

      pythonProject = pyproject-nix.lib.project.loadPyproject {
        projectRoot = ./python;
        pyproject = pkgs.lib.importTOML pyproject-file;
      };

      get-function-docs = pkgs.writeShellScriptBin "get-function-docs" ''
        ${b6-go}/bin/b6-api --docs --functions | ${pkgs.lib.getExe pkgs.jq} ".Functions[] | select(.Name == \"''$1\")"
      '';

      b6-py = python.pkgs.buildPythonPackage
        (pythonProject.renderers.buildPythonPackage
          {
            inherit python;
          } // {
          patchPhase = ''
            cat ${pyproject-file} > pyproject.toml
          '';
        });

      pythonEnv = python.withPackages (ps:
        [
          b6-py

          # For `make python`
          ps.grpcio-tools

          # For hacking
          ps.jupyter
        ]
      );

      # Go setup
      b6-go = with pkgs; gomod2nix.legacyPackages.${system}.buildGoApplication {
        name = "b6";
        src = ./src/diagonal.works/b6;
        buildInputs = [
          gdal
        ];
        nativeBuildInputs = [
          pkg-config
        ];

        # TODO: Nix can't run the tests because they refer to files that are
        # outside the nix closure, ultimately. Some potential fixes are:
        #
        #   - Fix gomod2nix to allow us to bring these files in somehow
        #   - Move the files that are for go-tests into a sub-folder of the go source
        #   - ???
        #
        doCheck = false;

        # Must be added due to bug https://github.com/nix-community/gomod2nix/issues/120
        pwd = ./src/diagonal.works/b6;
      };

      pkgs = import nixpkgs { inherit system; overlays = [ overlay ]; };
      unstablePkgs = import unstable { inherit system; };
    in
    {
      # Development shells for hacking/building with the Makefile
      devShells.default = pkgs.mkShell {
        packages = with pkgs; [
          # Misc tools
          get-function-docs

          # Running the Makefile tasks
          gdal
          pkg-config
          protobuf
          protoc-gen-go
          protoc-gen-go-grpc

          # Python hacking
          pythonEnv

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


      packages = {
        # Run like `nix run . -- --help` or access all the binaries with
        # `nix build` and look in `./result/bin`.
        default = b6-go;

        # Not an application; but can be built `nix build .#python`.
        python = b6-py;
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
