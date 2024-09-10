# Sets TARGETARCH to something like x86_64 or arm64
TARGETARCH ?= $(shell uname -m | tr A-Z a-z)
# Sets TARGETOS to something like linux or darwin
TARGETOS ?= $(shell uname -s | tr A-Z a-z)
# Sets TARGETPLATFORM to something like linux/x86_64 or darwin/arm64
export TARGETPLATFORM ?= ${TARGETOS}/${TARGETARCH}

all: .git/hooks/pre-commit b6 b6-ingest-osm b6-ingest-gdal b6-ingest-terrain b6-ingest-gb-uprn b6-ingest-gb-codepoint b6-connect b6-api python docs

.git/hooks/pre-commit: etc/pre-commit
	cp $< $@

b6: b6-backend b6-frontend
	make -C src/diagonal.works/b6/cmd/b6/js

b6-frontend:
	make -C frontend

VERSION: b6-api
	bin/${TARGETPLATFORM}/b6-api --version > $@

b6-backend: proto-go src/diagonal.works/b6/api/y.go VERSION
	cd src/diagonal.works/b6/cmd/b6; go build -o ../../../../../bin/${TARGETPLATFORM}/b6 -ldflags "-X=diagonal.works/b6.BackendVersion=`cat ../../../../../VERSION`"

b6-ingest-osm:
	cd src/diagonal.works/b6/cmd/$@; go build -o ../../../../../bin/${TARGETPLATFORM}/$@

b6-ingest-gdal:
	cd src/diagonal.works/b6/cmd/$@; go build -o ../../../../../bin/${TARGETPLATFORM}/$@

b6-ingest-gtfs:
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
	bin/${TARGETPLATFORM}/b6-api --docs > src/diagonal.works/b6/api/functions/docs.generated
	mv src/diagonal.works/b6/api/functions/docs.generated src/diagonal.works/b6/api/functions/docs.go
	cd src/diagonal.works/b6/cmd/$@; go build -o ../../../../../bin/${TARGETPLATFORM}/$@

proto: proto-go proto-python

proto-go: src/diagonal.works/b6/proto/tiles.pb.go src/diagonal.works/b6/proto/geometry.pb.go src/diagonal.works/b6/proto/compact.pb.go src/diagonal.works/b6/proto/ui.pb.go src/diagonal.works/b6/proto/api.pb.go src/diagonal.works/b6/proto/api_grpc.pb.go src/diagonal.works/b6/osm/proto/pbf.pb.go

src/diagonal.works/b6/proto/ui.pb.go: proto/ui.proto proto/api.proto proto/geometry.proto

src/diagonal.works/b6/proto/api.pb.go: proto/api.proto proto/geometry.proto

src/diagonal.works/b6/proto/api_grpc.pb.go: proto/api.proto proto/geometry.proto

src/diagonal.works/b6/osm/proto/pbf.pb.go: proto/pbf.proto

%_grpc.pb.go:
	protoc -I=proto --go_out=src --go-grpc_out=src $<

%.pb.go:
	protoc -I=proto --go_out=src $<

src/diagonal.works/b6/api/y.go: src/diagonal.works/b6/api/shell.y
	cd src/diagonal.works/b6/api; goyacc shell.y

python: proto-python python/diagonal_b6/api_generated.py python/pyproject.toml

proto-python: python/diagonal_b6/geometry_pb2.py python/diagonal_b6/api_pb2.py python/diagonal_b6/api_pb2_grpc.py

python/diagonal_b6/api_pb2.py: proto/api.proto proto/geometry.proto

python/diagonal_b6/api_pb2_grpc.py: proto/api.proto proto/geometry.proto

python/diagonal_b6/%_pb2.py: proto/%.proto
	python3 -m grpc_tools.protoc -Iproto --python_out=python/diagonal_b6 $<
	sed -e 's/import geometry_pb2/import diagonal_b6.geometry_pb2/' $@ > $@.modified
	mv $@.modified $@

python/diagonal_b6/%_pb2_grpc.py: proto/%.proto
	python3 -m grpc_tools.protoc -Iproto --grpc_python_out=python/diagonal_b6 $<
	sed -e 's/import geometry_pb2/import diagonal_b6.geometry_pb2/' $@ > $@.modified
	mv $@.modified $@
	sed -e 's/import api_pb2/import diagonal_b6.api_pb2/' $@ > $@.modified
	mv $@.modified $@

python/diagonal_b6/api_generated.py: proto-python b6-api
	bin/${TARGETPLATFORM}/b6-api --functions | python/diagonal_b6/generate_api.py > $@

python/pyproject.toml: python/pyproject.toml.template python/VERSION
	sed -e s/VERSION/`cat python/VERSION`/ $< > $@

python/VERSION:
	bin/${TARGETPLATFORM}/b6-api --pip-version > $@

python-test: python b6-backend b6-ingest-osm
	bin/${TARGETPLATFORM}/b6-ingest-osm --input=data/tests/granary-square.osm.pbf --output=data/tests/granary-square.index
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

clean-api-docs:
	rm docs/docs/b6-api-documentation.md

docs: docs/docs/b6-api-documentation.md

docs/docs/b6-api-documentation.md: clean-api-docs
	bin/${TARGETPLATFORM}/b6-api --docs --functions | ./scripts/api-docs-to-docusaurus.py >docs/docs/b6-api-documentation.md

clean:
	cd src/diagonal.works/b6; go clean
	rm -f src/diagonal.works/b6/proto/*.pb.go
	rm -f src/diagonal.works/b6/osm/proto/*.pb.go
	rm -f python/diagonal_b6/*_pb2.py
	rm -f python/diagonal_b6/*_pb2_grpc.py

.PHONY: python proto proto-go proto-python docker b6-frontend clean-api-docs docs
