{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    diagonal-b6.url = "github:diagonalworks/diagonal-b6";
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
