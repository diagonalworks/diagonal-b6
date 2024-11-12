{ pkgs
}:
let
  b6-js = pkgs.buildNpmPackage {
    pname = "b6-js";
    version = "v.0.0.0";
    src = ./../src/diagonal.works/b6/cmd/b6/js;
    npmDepsHash = "sha256-w332KqVpqdZSoNjwRxAaB4PZKaMfnM0MHl3lcpkmmQU=";

    buildPhase = ''
      npm run build
    '';

    installPhase = ''
      mkdir $out
      mv bundle.js $out
    '';
  };

  # Note: We obtain a list of all the "features" (i.e. the folders on this
  # specific directory) and use that to construct a set of derivations
  # that turn on/off every feature combination.
  frontend-feature-matrix =
    let
      allPaths = builtins.readDir ./../frontend/src/features;
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
      configurations = pkgs.lib.cartesianProduct allOptions;

      # { "abc=true,xyz=true"  = make-frontend { ABC = "true"; XYZ = "true" };
      # , "abc=true,xyz=false" = make-frontend { ABC = "true"; XYZ = false  };
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
              value = make-frontend c;
            })
            configurations;
        in
        builtins.listToAttrs elements;
    in
    matrix;

  # From <https://github.com/NixOS/nixpkgs/blob/master/doc/languages-frameworks/javascript.section.md#pnpm-javascript-pnpm>
  make-frontend = envVars: pkgs.stdenv.mkDerivation (finalAttrs: {
    pname = "b6-frontend";
    version = "0.0.0";

    src = ./../frontend;

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
      ${pkgs.lib.strings.stringAsChars (x: if x == "\n" then " " else x) (pkgs.lib.strings.toShellVars envVars)} pnpm build
    '';

    installPhase = ''
      rm dist/index-vite.html
      mv dist/ $out
    '';
  });

  # The default frontend configuration.
  frontend = make-frontend {
    # Note: If you want something to be true it must equal "true".
    VITE_FEATURES_SCENARIOS = false;
    VITE_FEATURES_SHELL = false;
  };

  # A frontend for development; shell is on, scenarios are off.
  frontend-dev = make-frontend {
    VITE_FEATURES_SCENARIOS = false;
    VITE_FEATURES_SHELL = "true";
  };
in
{
  inherit
    b6-js
    frontend
    frontend-dev
    ;
} // frontend-feature-matrix
