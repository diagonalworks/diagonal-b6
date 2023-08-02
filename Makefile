# Sets TARGETARCH to something like x86_64 or arm64
TARGETARCH ?= $(shell uname -m | tr A-Z a-z)
# Sets TARGETOS to something like linux or darwin
TARGETOS ?= $(shell uname -s | tr A-Z a-z)
# Sets TARGETPLATFORM to something like linux/x86_64 or darwin/arm64
export TARGETPLATFORM ?= ${TARGETOS}/${TARGETARCH}

all: .git/hooks/pre-commit b6 b6-ingest-osm b6-ingest-gdal b6-ingest-terrain b6-ingest-gb-uprn b6-ingest-gb-codepoint b6-connect b6-api python

.git/hooks/pre-commit: etc/pre-commit
	cp $< $@

b6: b6-backend
	make -C src/diagonal.works/b6/cmd/b6/js

VERSION: b6-api
	bin/${TARGETPLATFORM}/b6-api --version > $@

b6-backend: proto-go src/diagonal.works/b6/api/y.go VERSION
	cd src/diagonal.works/b6/cmd/b6; go build -o ../../../../../bin/${TARGETPLATFORM}/b6 -ldflags "-X=diagonal.works/b6.BackendVersion=`cat ../../../../../VERSION`"

b6-ingest-osm:
	cd src/diagonal.works/b6/cmd/$@; go build -o ../../../../../bin/${TARGETPLATFORM}/$@

b6-ingest-gdal:
	cd src/diagonal.works/b6/cmd/$@; go build -o ../../../../../bin/${TARGETPLATFORM}/$@

b6-ingest-terrain:
	cd src/diagonal.works/b6/cmd/$@; go build -o ../../../../../bin/${TARGETPLATFORM}/$@

b6-ingest-gb-uprn:
	cd src/diagonal.works/b6/cmd/$@; go build -o ../../../../../bin/${TARGETPLATFORM}/$@

b6-ingest-gb-codepoint:
	cd src/diagonal.works/b6/cmd/$@; go build -o ../../../../../bin/${TARGETPLATFORM}/$@

b6-connect:
	cd src/diagonal.works/b6/cmd/$@; go build -o ../../../../../bin/${TARGETPLATFORM}/$@

b6-api:
	cd src/diagonal.works/b6/cmd/$@; go build -o ../../../../../bin/${TARGETPLATFORM}/$@

proto: proto-go proto-python

proto-go: src/diagonal.works/b6/proto/tiles.pb.go src/diagonal.works/b6/proto/geometry.pb.go src/diagonal.works/b6/proto/features.pb.go src/diagonal.works/b6/proto/compact.pb.go src/diagonal.works/b6/proto/ui.pb.go src/diagonal.works/b6/proto/api.pb.go src/diagonal.works/b6/proto/api_grpc.pb.go src/diagonal.works/b6/osm/proto/pbf.pb.go

src/diagonal.works/b6/proto/features.pb.go: proto/features.proto proto/geometry.proto

src/diagonal.works/b6/proto/ui.pb.go: proto/ui.proto proto/api.proto proto/features.proto

src/diagonal.works/b6/proto/api.pb.go: proto/api.proto proto/features.proto proto/geometry.proto

src/diagonal.works/b6/proto/api_grpc.pb.go: proto/api.proto proto/features.proto proto/geometry.proto

src/diagonal.works/b6/osm/proto/pbf.pb.go: proto/pbf.proto

%_grpc.pb.go:
	protoc -I=proto --go_out=src --go-grpc_out=src $<

%.pb.go:
	protoc -I=proto --go_out=src $<

src/diagonal.works/b6/api/y.go: src/diagonal.works/b6/api/shell.y
	cd src/diagonal.works/b6/api; goyacc shell.y

python: proto-python python/diagonal_b6/api_generated.py python/pyproject.toml

proto-python: python/diagonal_b6/geometry_pb2.py python/diagonal_b6/features_pb2.py python/diagonal_b6/api_pb2.py python/diagonal_b6/api_pb2_grpc.py

python/diagonal_b6/features_pb2.py: proto/features.proto proto/geometry.proto

python/diagonal_b6/api_pb2.py: proto/api.proto proto/features.proto proto/geometry.proto

python/diagonal_b6/api_pb2_grpc.py: proto/api.proto proto/features.proto proto/geometry.proto

python/diagonal_b6/%_pb2.py: proto/%.proto
	python3 -m grpc_tools.protoc -Iproto --python_out=python/diagonal_b6 $<
	sed -e 's/import geometry_pb2/import diagonal_b6.geometry_pb2/' $@ > $@.modified
	mv $@.modified $@
	sed -e 's/import features_pb2/import diagonal_b6.features_pb2/' $@ > $@.modified
	mv $@.modified $@

python/diagonal_b6/%_pb2_grpc.py: proto/%.proto
	python3 -m grpc_tools.protoc -Iproto --grpc_python_out=python/diagonal_b6 $<
	sed -e 's/import geometry_pb2/import diagonal_b6.geometry_pb2/' $@ > $@.modified
	mv $@.modified $@
	sed -e 's/import features_pb2/import diagonal_b6.features_pb2/' $@ > $@.modified	
	mv $@.modified $@
	sed -e 's/import api_pb2/import diagonal_b6.api_pb2/' $@ > $@.modified
	mv $@.modified $@

python/diagonal_b6/api_generated.py: proto-python b6-api
	bin/${TARGETPLATFORM}/b6-api | python/diagonal_b6/generate_api.py > $@

python/pyproject.toml: python/pyproject.toml.template python/VERSION
	sed -e s/VERSION/`cat python/VERSION`/ $< > $@

python/VERSION:
	bin/${TARGETPLATFORM}/b6-api --pip-version > $@

python-test: python b6-backend
	PYTHONPATH=python TARGETPLATFORM=${TARGETPLATFORM} python3 python/diagonal_b6/b6_test.py

test: proto-go src/diagonal.works/b6/api/y.go
	cd src/diagonal.works/b6; go test diagonal.works/b6/...

docker: docker/Dockerfile.b6 docker/Dockerfile.b6-ci

docker/Dockerfile.b6: docker/Dockerfile.b6-build.inc docker/Dockerfile.b6.inc

docker/Dockerfile.b6-ci: docker/Dockerfile.b6-build.inc docker/Dockerfile.b6-ci.inc

docker/Dockerfile.%:
	cat $^ > $@

docker-b6-ci: docker/Dockerfile.b6-ci
	docker build -t europe-docker.pkg.dev/diagonal-public/b6/b6-ci -f docker/Dockerfile.b6-ci .
	docker push europe-docker.pkg.dev/diagonal-public/b6/b6-ci

clean:
	cd src/diagonal.works/b6; go clean
	rm -f src/diagonal.works/b6/proto/*.pb.go
	rm -f src/diagonal.works/b6/osm/proto/*.pb.go
	rm -f python/diagonal_b6/*_pb2.py
	rm -f python/diagonal_b6/*_pb2_grpc.py

.PHONY: python proto proto-go proto-python docker
