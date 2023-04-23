# Build a docker image for b6, and a Python package, assuming the
# source tree is passed as the docker context. For example, at the root
# of the source tree:
# docker build --build-arg=TARGETPLATFORM=linux/x86_64 -f docker/Dockerfile.package -t b6-x86_64 .
FROM ubuntu:jammy AS build
ARG TARGETPLATFORM
RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get -yq install git make ca-certificates curl npm protobuf-compiler gdal-bin libgdal-dev python3-pip python3.10-venv
RUN curl -L -O https://go.dev/dl/go1.20.3.linux-amd64.tar.gz && tar -C /usr/local -xzf go1.20.3.linux-amd64.tar.gz && rm go1.20.3.linux-amd64.tar.gz
ENV GOBIN=/usr/local/go/bin
ENV PATH=/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest && GOBIN=/usr/local/go/bin go install golang.org/x/tools/cmd/goyacc@latest
RUN python3 -m pip install grpcio-tools build
COPY Makefile /build/Makefile
COPY src /build/src
COPY python /build/python
COPY proto /build/proto
COPY data /build/data
RUN make -C /build
RUN cd /build/python; python3 -m build
RUN /build/bin/$TARGETPLATFORM/b6-ingest-osm --input=/build/data/tests/camden.osm.pbf --output=/build/data/camden.index
FROM ubuntu:jammy
ARG TARGETPLATFORM
COPY --from=build /build/bin/$TARGETPLATFORM/* /diagonal/bin/
COPY --from=build /build/data/camden.index /world/camden.index
COPY --from=build /build/python/dist/* /diagonal/python/dist
COPY --from=build /build/src/diagonal.works/b6/cmd/b6/js/bundle.js /diagonal/js/bundle.js
COPY --from=build /build/src/diagonal.works/b6/cmd/b6/js/static /diagonal/static
EXPOSE 8001 8002
CMD ["/diagonal/bin/b6", "--world=/world", "--js=/diagonal/js", "--static=/diagonal/static", "--http=0.0.0.0:8001", "--grpc=0.0.0.0:8002"]
