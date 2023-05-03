# Sets TARGETARCH to something like x86_64 or arm64
TARGETARCH ?= $(shell uname -m | tr A-Z a-z)
# Sets TARGETOS to something like linux or darwin
TARGETOS ?= $(shell uname -s | tr A-Z a-z)
# Sets TARGETPLATFORM to something like linux/x86_64 or darwin/arm64
export TARGETPLATFORM ?= ${TARGETOS}/${TARGETARCH}


all: b6 b6-ingest-osm b6-ingest-gdal b6-ingest-terrain b6-ingest-gb-uprn b6-ingest-gb-codepoint b6-connect b6-api python

b6: b6-backend
	make -C src/diagonal.works/b6/cmd/b6/js

VERSION: b6-api
	bin/${TARGETPLATFORM}/b6-api --version > $@

b6-backend: proto src/diagonal.works/b6/api/y.go VERSION
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

proto:
	protoc -I=proto --go_out=src proto/tiles.proto
	protoc -I=proto --go_out=src proto/geometry.proto
	protoc -I=proto --go_out=src proto/features.proto
	protoc -I=proto --go_out=src --go-grpc_out=src proto/api.proto
	protoc -I=src/diagonal.works/b6/osm/proto --go_out=src src/diagonal.works/b6/osm/proto/pbf.proto

src/diagonal.works/b6/api/y.go: src/diagonal.works/b6/api/shell.y
	cd src/diagonal.works/b6/api; goyacc shell.y

python: python/diagonal_b6/api_generated.py python/pyproject.toml

python/diagonal_b6/api_generated.py: b6-api
	python3 -m grpc_tools.protoc -Iproto --python_out=python/diagonal_b6 --grpc_python_out=python/diagonal_b6  proto/geometry.proto proto/features.proto proto/api.proto
	sed -e 's/import geometry_pb2/import diagonal_b6.geometry_pb2/' python/diagonal_b6/features_pb2.py > python/diagonal_b6/features_pb2.py.modified
	mv python/diagonal_b6/features_pb2.py.modified python/diagonal_b6/features_pb2.py
	sed -e 's/import geometry_pb2/import diagonal_b6.geometry_pb2/' python/diagonal_b6/api_pb2.py > python/diagonal_b6/api_pb2.py.modified
	mv python/diagonal_b6/api_pb2.py.modified python/diagonal_b6/api_pb2.py
	sed -e 's/import features_pb2/import diagonal_b6.features_pb2/' python/diagonal_b6/api_pb2.py > python/diagonal_b6/api_pb2.py.modified
	mv python/diagonal_b6/api_pb2.py.modified python/diagonal_b6/api_pb2.py
	sed -e 's/import api_pb2/import diagonal_b6.api_pb2/' python/diagonal_b6/api_pb2_grpc.py > python/diagonal_b6/api_pb2_grpc.py.modified
	mv python/diagonal_b6/api_pb2_grpc.py.modified python/diagonal_b6/api_pb2_grpc.py
	bin/${TARGETPLATFORM}/b6-api | python/diagonal_b6/generate_api.py > $@

python/pyproject.toml: python/pyproject.toml.template python/VERSION
	sed -e s/VERSION/`cat python/VERSION`/ $< > $@

python/VERSION:
	bin/${TARGETPLATFORM}/b6-api --pip-version > $@

python-test: python b6-backend
	PYTHONPATH=python TARGETPLATFORM=${TARGETPLATFORM} python3 python/diagonal_b6/b6_test.py

test: proto src/diagonal.works/b6/api/y.go
	cd src/diagonal.works/b6; go test diagonal.works/b6/...

clean:
	cd src/diagonal.works/b6; go clean
	rm -f src/diagonal.works/b6/proto/*.pb.go
	rm -f src/diagonal.works/b6/osm/proto/*.pb.go
	rm -f python/diagonal_b6/*_pb2.py
	rm -f python/diagonal_b6/*_pb2_grpc.py

.PHONY: python proto docker
