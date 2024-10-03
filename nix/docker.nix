{ pkgs
, b6-js-packages
}:
let
  make-b6-image = name: b6-drv:
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
            STATIC_ARG=${b6-js-packages."frontend-with-scenarios=true".outPath} ;;
          "frontend-with-scenarios=false")
            STATIC_ARG=${b6-js-packages."frontend-with-scenarios=false".outPath} ;;
          *)
            STATIC_ARG=${b6-js-packages.frontend.outPath} ;;
        esac

        # So we can kill it with Ctrl-C
        _term() {
          kill "$child" 2>/dev/null
        }
        trap _term INT

        ${b6-drv}/bin/b6 \
          -http=0.0.0.0:8001 \
          -grpc=0.0.0.0:8002 \
          -js=${b6-js-packages.b6-js.outPath} \
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
make-b6-image
