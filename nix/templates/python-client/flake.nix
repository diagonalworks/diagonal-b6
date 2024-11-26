{
  inputs = {
    # Note: This _must_ match the one coming from the version in diagonal-b6.
    # TODO: Work out how to do this a bit more cleanly.
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-24.11";

    flake-utils.url = "github:numtide/flake-utils";

    # Note: This has to stay as a single expression as we use sed to replace
    # it in the 'ci-nix' CI task.
    diagonal-b6.url = "github:diagonalworks/diagonal-b6";
    diagonal-b6.inputs.nixpkgs.follows = "nixpkgs";
  };


  outputs =
    { self
    , nixpkgs
    , flake-utils
    , diagonal-b6
    , ...
    }:
    flake-utils.lib.eachDefaultSystem (system:
    let
      pkgs = import nixpkgs { inherit system; };

      # Python envrionment
      #
      # Here we depend on the b6 python library, as well as any additional
      # libraries we might like to use.
      py-env = pkgs.python312.withPackages (ps: with ps; [
        # The b6 python library
        diagonal-b6.packages."${system}".python312
        # Additional python dependencies
        geopandas
        jupyter
        more-itertools
        pandas
        # Note: Here is where you would add extra Python libraries
        # ex: numpy
        #
        # If you want to specify a library from, say, GitHub, it will look
        # something like this:
        #
        # (
        #   buildPythonPackage rec {
        #     version = "1.3";
        #     pname = "altair-nx";
        #
        #     # Hack: So that the runtime dependency check doesn't fail.
        #     propagatedBuildInputs = [
        #       altair
        #       networkx
        #       pandas
        #     ];
        #
        #     # This has to be determined by either trial-and-error, or
        #     investigating the pyproject.toml or setup.py of the relevant
        #     project. It's a bit annoying.
        #
        #     format = "pyproject";
        #     nativeBuildInputs = [
        #       hatchling
        #       hatch-vcs
        #     ];
        #
        #     # This is the easy part; the GitHub details:
        #     src = pkgs.fetchFromGitHub {
        #       owner = "T-Flet";
        #       repo = "altair-nx";
        #       rev = "master";
        #       hash = "sha256-AlZuFqq1GaZeW6xfvxvAPIXABAm3ipJuTASqce7AD+s=";
        #     };
        #   }
        # )
        #
      ]);
    in
    {
      # Development shell
      #
      # One default shell is provided, containing the Python envrionment, the
      # (wrapped) b6 executable, as well as all the other b6-* executables
      # that are defined in the project.
      #
      devShells.default = pkgs.mkShell {
        packages = with pkgs; [
          py-env
          diagonal-b6.packages.${system}.run-b6
          # We'll also take a whole bunch of b6-ingest-* executables, in case
          # we would like to run any ad-hoc data ingestions.
          diagonal-b6.packages.${system}.go
        ];
      };
    });


  # Cachix setup: Use the diagonalworks cachix by default, to save extra
  # building!
  nixConfig = {
    extra-substituters = [
      "https://diagonalworks.cachix.org"
    ];
    extra-trusted-public-keys = [
      "diagonalworks.cachix.org-1:7U834B3foDCfa1EeV6xpyOz9JhdfUXj2yxRv0rAdYMk="
    ];
  };
}
