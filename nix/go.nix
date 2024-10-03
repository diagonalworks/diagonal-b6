{ pkgs
, ourGdal
, system
, gomod2nix
}:
let
  # Go setup
  everything = with pkgs; gomod2nix.legacyPackages.${system}.buildGoApplication {
    name = "b6";
    src = ./../src/diagonal.works/b6;
    # Must be added due to bug https://github.com/nix-community/gomod2nix/issues/120
    pwd = ./../src/diagonal.works/b6;
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
      cp -r ${./../data/tests}/* ./data/tests
    '';

    doCheck = true;

  };

  # A collection of derivations for each go cmd; can be useful to keep
  # closures small, if for example you only want to depend on a specific
  # binary.
  go-executables =
    let
      allPaths = builtins.readDir ./../src/diagonal.works/b6/cmd;
      onlyDirs = pkgs.lib.attrsets.filterAttrs (_: v: v == "directory") allPaths;
      cmds = builtins.attrNames onlyDirs;
      mkGoApp = cmd: with pkgs; gomod2nix.legacyPackages.${system}.buildGoApplication {
        name = "${cmd}";
        src = ./../src/diagonal.works/b6;
        pwd = ./../src/diagonal.works/b6;
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
in
{
  inherit
    everything
    go-executables
    ;
}
