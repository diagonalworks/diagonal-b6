# Build a docker image to run the b6 continuous integration pipeline,
# including the standard b6 build environment, but populating the go
# cache.
# Written to extend Dockerfile.b6-build.inc
COPY src/diagonal.works/b6/go.* /bootstrap/src/diagonal.works/b6/
RUN cd /bootstrap/src/diagonal.works/b6/; go mod download
CMD ["/bin/bash"]