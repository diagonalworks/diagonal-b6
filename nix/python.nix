{ pkgs
, python
, b6-go-packages
, pyproject-nix
}:
let
  # We have to write the version into here from the output of the go
  # binary.
  pyproject-file = (pkgs.runCommand "make-pyproject" { } ''
    substitute ${./../python/pyproject.toml.template} $out \
      --subst-var-by VERSION ''$(${b6-go-packages.go-executables.b6-api}/bin/b6-api --pip-version)
  '');

  pythonProject = pyproject-nix.lib.project.loadPyproject {
    projectRoot = ./../python;
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
      cp -r ${./../proto} ./proto
      cat ${./../Makefile} > some-Makefile

      # Hack: Run the b6-api command outside of the Makefile, using the
      # Nix version of the binary.
      ${b6-go-packages.go-executables.b6-api}/bin/b6-api --functions | python diagonal_b6/generate_api.py > diagonal_b6/api_generated.py

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
in
b6-py
