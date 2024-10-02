# Build a docker image containing the tools and dependencies necessary to
# build b6 from source
FROM ubuntu:mantic AS b6-build
ARG TARGETOS
ARG TARGETARCH
RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get -yq install git make ca-certificates curl npm protobuf-compiler gdal-bin libgdal-dev python3-full
RUN npm install -g pnpm
RUN curl -L -O https://go.dev/dl/go1.21.6.$TARGETOS-$TARGETARCH.tar.gz
RUN tar -C /usr/local -xzf go1.21.6.$TARGETOS-$TARGETARCH.tar.gz && rm go1.21.6.$TARGETOS-$TARGETARCH.tar.gz
RUN mkdir /go-cache
ENV GOBIN=/usr/local/go/bin
ENV GOCACHE=/go-cache
ENV PATH=/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest && GOBIN=/usr/local/go/bin go install golang.org/x/tools/cmd/goyacc@latest
RUN python3 -m venv /build
RUN /build/bin/pip install build grpcio grpcio-tools s2sphere
ENV PATH="/build/bin:$PATH"
