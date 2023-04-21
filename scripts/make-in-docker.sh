#!/bin/sh
# Run a build inside a docker container, using the build image generated from
# docker/Dockerfile.build
if [ -f Makefile ]; then
    docker run --mount type=bind,source=${PWD},target=/diagonal -it --name=build --rm build make -C /diagonal $@
else
    echo "Run script from the root of the b6 source tree"
fi
