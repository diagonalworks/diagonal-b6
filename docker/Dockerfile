# Build a docker image for b6, and a Python package, assuming the
# source tree is passed as the docker context. For example, at the root
# of the source tree:
# docker build --build-arg=TARGETOS=linux,TARGETARCH=amd64 -f docker/Dockerfile -t b6-amd64 .
FROM ubuntu:jammy AS build
ARG TARGETOS
ARG TARGETARCH
RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get -yq install git make ca-certificates curl npm protobuf-compiler gdal-bin libgdal-dev python3-pip python3.10-venv
RUN curl -L -O https://go.dev/dl/go1.20.3.$TARGETOS-$TARGETARCH.tar.gz
RUN tar -C /usr/local -xzf go1.20.3.$TARGETOS-$TARGETARCH.tar.gz && rm go1.20.3.$TARGETOS-$TARGETARCH.tar.gz
ENV GOBIN=/usr/local/go/bin
ENV PATH=/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest && GOBIN=/usr/local/go/bin go install golang.org/x/tools/cmd/goyacc@latest
RUN python3 -m pip install build grpcio grpcio-tools s2sphere
COPY Makefile /build/Makefile
COPY src /build/src
COPY python /build/python
COPY proto /build/proto
COPY data /build/data
RUN TARGETOS=$TARGETOS TARGETARCH=$TARGETARCH make -C /build
RUN TARGETOS=$TARGETOS TARGETARCH=$TARGETARCH make -C /build test
RUN TARGETOS=$TARGETOS TARGETARCH=$TARGETARCH make -C /build python-test
RUN cd /build/python; python3 -m build
RUN /build/bin/$TARGETOS/$TARGETARCH/b6-ingest-osm --input=/build/data/tests/camden.osm.pbf --output=/build/data/camden.index
RUN /build/bin/$TARGETOS/$TARGETARCH/b6-connect --input=/build/data/camden.index --output=/build/data/camden.connected.index
FROM ubuntu:jammy
ARG TARGETOS
ARG TARGETARCH
COPY --from=build /build/bin/$TARGETOS/$TARGETARCH/* /diagonal/bin/
COPY --from=build /build/data/camden.connected.index /world/camden.connected.index
COPY --from=build /build/python/dist/* /diagonal/python/dist
COPY --from=build /build/src/diagonal.works/b6/cmd/b6/js/bundle.js /diagonal/js/bundle.js
COPY --from=build /build/src/diagonal.works/b6/cmd/b6/js/static /diagonal/static
RUN tar -czf /diagonal-$TARGETOS-$TARGETARCH.tar.gz /diagonal
EXPOSE 8001 8002
CMD ["/diagonal/bin/b6", "--world=/world", "--js=/diagonal/js", "--static=/diagonal/static", "--http=0.0.0.0:8001", "--grpc=0.0.0.0:8002"]
