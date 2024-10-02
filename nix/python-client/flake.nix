{
  description = "A local setup for b6: python client + local b6 executable";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/24.05";
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
      py-env = pkgs.python3.withPackages (ps: with ps; [
        jupyter
        # Note: Here is where you would add extra Python libraries
        # ex: numpy
        diagonal-b6.packages."${system}".python
      ]);


      # b6 executable
      #
      # Note: We define a wrapped version of the b6 executable, using a
      # particular build of the frontend, for easy invocation, such as:
      #
      # > b6 -world data
      #
      b6-wrapped =
        let b6 = diagonal-b6.packages.${system};
        in
        pkgs.writeShellScriptBin "b6" ''
          ${b6.b6}/bin/b6 \
            -http=0.0.0.0:8001 \
            -grpc=0.0.0.0:8002 \
            -js=${b6.b6-js.outPath} \
            -enable-v2-ui \
            -static-v2=${b6."frontend-with-scenarios=true".outPath} \
            "$@"
        '';
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
          b6-wrapped
          # We'll also take a whole bunch of b6-ingest-* executables, in case
          # we would like to run any ad-hoc data ingestions.
          diagonal-b6.packages."${system}".go
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
