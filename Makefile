all: protos experimental fe ingest transit fe-js
	cd src/diagonal.works/diagonal; go build diagonal.works/diagonal/...
	cd src/diagonal.works/diagonal/cmd/inspect; go build
	cd src/diagonal.works/diagonal/cmd/splitosm; go build
	cd src/diagonal.works/diagonal/cmd/tile; go build
	make -C data

fe: protos
	cd src/diagonal.works/diagonal/cmd/fe; go build

fe-js:
	make -C js

ingest: protos
	cd src/diagonal.works/diagonal/cmd/ingest; go build

transit: protos
	cd src/diagonal.works/diagonal/cmd/transit; go build

docker: protos
	mkdir -p docker/bin/linux-amd64
	cd src/diagonal.works/diagonal/cmd/ingest; GOOS=linux GOARCH=amd64 go build -o ../../../../../docker/bin/linux-amd64/ingest
	cd src/diagonal.works/diagonal/cmd/splitosm; GOOS=linux GOARCH=amd64 go build -o ../../../../../docker/bin/linux-amd64/splitosm
	docker build -f docker/Dockerfile.diagonal -t diagonal docker
	docker tag diagonal eu.gcr.io/diagonal-platform/diagonal
	docker push eu.gcr.io/diagonal-platform/diagonal
	docker build -f docker/Dockerfile.monitoring -t monitoring docker
	docker tag monitoring eu.gcr.io/diagonal-platform/monitoring
	docker push eu.gcr.io/diagonal-platform/monitoring
	docker build -f docker/Dockerfile.planet -t planet docker
	docker tag planet eu.gcr.io/diagonal-platform/planet
	docker push eu.gcr.io/diagonal-platform/planet

docker-atlas-internal: fe-js
	mkdir -p docker/bin/linux-amd64
	cd src/diagonal.works/diagonal/cmd/fe; GOOS=linux GOARCH=amd64 go build -o ../../../../../docker/bin/linux-amd64/fe
	mkdir -p docker/js
	rm -rf docker/js/dist
	cp -r js/dist docker/js/dist
	cp data/earth/ne_10m_land.shp docker/data/atlas-internal
	cp data/earth/ne_10m_land.prj docker/data/atlas-internal
	docker build -f docker/Dockerfile.atlas-internal -t atlas-internal docker

docker-atlas-dev: fe-js
	mkdir -p docker/bin/linux-amd64
	cd src/diagonal.works/diagonal/cmd/fe; GOOS=linux GOARCH=amd64 go build -o ../../../../../docker/bin/linux-amd64/fe
	mkdir -p docker/js
	rm -rf docker/js/dist
	cp -r js/dist docker/js/dist
	cp data/earth/ne_10m_land.shp docker/data/atlas-dev
	cp data/earth/ne_10m_land.prj docker/data/atlas-dev
	docker build -f docker/Dockerfile.atlas-dev -t atlas-dev docker
	docker tag atlas-dev eu.gcr.io/diagonal-platform/atlas-dev
	docker push eu.gcr.io/diagonal-platform/atlas-dev

protos:
	protoc --plugin=${HOME}/go/bin/protoc-gen-go -I=proto --go_out=src proto/tiles.proto
	protoc --plugin=${HOME}/go/bin/protoc-gen-go -I=proto --go_out=src proto/osm.proto
	protoc --plugin=${HOME}/go/bin/protoc-gen-go -I=proto --go_out=src --python_out=python/diagonal/proto proto/geometry.proto
	protoc --plugin=${HOME}/go/bin/protoc-gen-go -I=proto --go_out=src --python_out=python/diagonal/proto proto/features.proto
	protoc --plugin=${HOME}/go/bin/protoc-gen-go -I=proto --go_out=src --go-grpc_out=src proto/api.proto
	protoc --plugin=${HOME}/go/bin/protoc-gen-go -I=src/diagonal.works/diagonal/osm --go_out=src src/diagonal.works/diagonal/osm/import.proto
	protoc --plugin=${HOME}/go/bin/protoc-gen-go -I=src/diagonal.works/diagonal/osm/pbf --go_out=src src/diagonal.works/diagonal/osm/pbf/pbf.proto
	flatc -o src/diagonal.works/diagonal/ingest --go src/diagonal.works/diagonal/ingest/fbs/ingest.fbs

experimental: experimental_geojson experimental_grpc
	cd src/diagonal.works/diagonal/experimental/mr; go build
	cd src/diagonal.works/diagonal/experimental/osmpbf; go build

experimental_geojson:
	cd src/diagonal.works/diagonal/experimental/geojson; go build

experimental_grpc:
	cd src/diagonal.works/diagonal/experimental/grpc; go build

python:
	python3 -m grpc.tools.protoc -Iproto --python_out=python/diagonal/proto --grpc_python_out=python/diagonal/proto proto/api.proto
	sed -i"" -e 's|import geometry_pb2|import diagonal.proto.geometry_pb2|' python/diagonal/proto/*.py
	sed -i"" -e 's|import features_pb2|import diagonal.proto.features_pb2|' python/diagonal/proto/*.py
	sed -i"" -e 's|import api_pb2|import diagonal.proto.api_pb2|' python/diagonal/proto/*.py

python_test: python
	PYTHONPATH=python python3 python/tests/all.py

test:
	cd src/diagonal.works/diagonal; go test -v diagonal.works/diagonal/...

clean:
	find . -type f -perm +a+x | xargs rm

.PHONY: python
