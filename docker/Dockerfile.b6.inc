# Build a docker image for b6, and a Python package, assuming the
# source repo is passed as the docker context. For example, at the root
# of a clone of the git repo:
# docker build --build-arg=TARGETOS=linux --build-arg=TARGETARCH=amd64 -f docker/Dockerfile -t b6-amd64 .
# Written to extend Dockerfile.b6-build.inc
COPY .git /build/.git
COPY data /build/data
COPY etc /build/etc
COPY frontend /build/frontend
COPY Makefile /build/Makefile
COPY proto /build/proto
COPY python /build/python
COPY src /build/src
RUN TARGETOS=$TARGETOS TARGETARCH=$TARGETARCH make -C /build
RUN TARGETOS=$TARGETOS TARGETARCH=$TARGETARCH make -C /build test
RUN TARGETOS=$TARGETOS TARGETARCH=$TARGETARCH make -C /build python-test
RUN cd /build/python; python3 -m build
RUN /build/bin/$TARGETOS/$TARGETARCH/b6-ingest-osm --input=/build/data/tests/camden.osm.pbf --output=/build/data/camden.VERSION.index
RUN /build/bin/$TARGETOS/$TARGETARCH/b6-connect --input=/build/data/camden.VERSION.index --output=/build/data/camden.connected.VERSION.index
FROM ubuntu:mantic
ARG TARGETOS
ARG TARGETARCH
RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get -yq install ca-certificates libgdal33
COPY --from=b6-build /build/bin/$TARGETOS/$TARGETARCH/* /diagonal/bin/
COPY --from=b6-build /build/data/camden.connected.index /world/camden.connected.index
COPY --from=b6-build /build/python/dist/* /diagonal/python/dist/
COPY --from=b6-build /build/src/diagonal.works/b6/cmd/b6/js/bundle.js /diagonal/js/bundle.js
COPY --from=b6-build /build/src/diagonal.works/b6/cmd/b6/js/static /diagonal/static
COPY --from=b6-build /build/frontend/dist /diagonal/frontend/dist
COPY --from=b6-build /build/frontend/dist-storybook /diagonal/frontend/dist-storybook
COPY --from=b6-build /build/VERSION /diagonal/VERSION
RUN tar -czf /diagonal-$TARGETOS-$TARGETARCH-`cat /diagonal/VERSION`.tar.gz /diagonal
EXPOSE 8001 8002
CMD ["/diagonal/bin/b6", "--world=/world", "--js=/diagonal/js", "--static=/diagonal/static", "--static-v2=/diagonal/frontend/dist", "--http=0.0.0.0:8001", "--grpc=0.0.0.0:8002"]
