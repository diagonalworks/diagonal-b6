all: protos experimental fe ingest ingestons transit fe-js dfe scaffold
	cd src/diagonal.works/diagonal/monitoring; go generate
	cd src/diagonal.works/diagonal; go build diagonal.works/diagonal/...
	cd src/diagonal.works/diagonal/cmd/inspect; go build
	cd src/diagonal.works/diagonal/cmd/splitosm; go build
	cd src/diagonal.works/diagonal/cmd/tile; go build
	make -C data

fe: protos
	cd src/diagonal.works/diagonal/monitoring; go generate
	cd src/diagonal.works/diagonal/cmd/fe; go build

fe-js:
	make -C js

ingest: protos
	cd src/diagonal.works/diagonal/cmd/ingest; go build

ingestons: protos
	cd src/diagonal.works/diagonal/cmd/ingestons; go build

transit: protos
	cd src/diagonal.works/diagonal/cmd/transit; go build

mbtiles:
	cd src/diagonal.works/diagonal/cmd/mbtiles; go build

dfe:
	cd src/diagonal.works/diagonal/cmd/dfe; go build

scaffold:
	cd src/diagonal.works/diagonal/cmd/scaffold; go build

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

docker-dfe: dfe
	mkdir -p docker/bin/linux-amd64
	cd src/diagonal.works/diagonal/cmd/dfe; GOOS=linux GOARCH=amd64 go build -o ../../../../../docker/bin/linux-amd64/dfe
	rm -rf docker/www/
	mkdir -p docker/www/
	cp -r src/diagonal.works/diagonal/experimental/website docker/www/staging.diagonal.works
	docker build -f docker/Dockerfile.dfe -t dfe docker
	docker tag dfe eu.gcr.io/diagonal-platform/dfe
	docker push eu.gcr.io/diagonal-platform/dfe

protos:
	protoc --plugin=${HOME}/go/bin/protoc-gen-go -I=proto --go_out=src proto/cookie.proto
	protoc --plugin=${HOME}/go/bin/protoc-gen-go -I=proto --go_out=src proto/tiles.proto
	protoc --plugin=${HOME}/go/bin/protoc-gen-go -I=proto --go_out=src proto/osm.proto
	protoc --plugin=${HOME}/go/bin/protoc-gen-go -I=proto --go_out=src proto/geometry.proto
	protoc --plugin=${HOME}/go/bin/protoc-gen-go -I=proto --go_out=src proto/features.proto
	protoc --plugin=${HOME}/go/bin/protoc-gen-go --plugin=${HOME}/go/bin/protoc-gen-go-grpc -I=proto --go_out=src --go-grpc_out=src proto/api.proto
	protoc --plugin=${HOME}/go/bin/protoc-gen-go -I=src/diagonal.works/diagonal/osm --go_out=src src/diagonal.works/diagonal/osm/import.proto
	protoc --plugin=${HOME}/go/bin/protoc-gen-go -I=src/diagonal.works/diagonal/osm/pbf --go_out=src src/diagonal.works/diagonal/osm/pbf/pbf.proto
	flatc -o src/diagonal.works/diagonal/ingest --go src/diagonal.works/diagonal/ingest/fbs/ingest.fbs

experimental: experimental_geojson experimental_staging
	cd src/diagonal.works/diagonal/experimental/mr; go build
	cd src/diagonal.works/diagonal/experimental/osmpbf; go build

experimental_geojson:
	cd src/diagonal.works/diagonal/experimental/geojson; go build

experimental_staging:
	cd src/diagonal.works/diagonal/experimental/staging; go build

python:
	python3 -m grpc.tools.protoc -Iproto --python_out=python/diagonal/proto proto/geometry.proto
	python3 -m grpc.tools.protoc -Iproto --python_out=python/diagonal/proto proto/features.proto
	python3 -m grpc.tools.protoc -Iproto --python_out=python/diagonal/proto --grpc_python_out=python/diagonal/proto proto/api.proto
	sed -e 's/import geometry_pb2/import diagonal.proto.geometry_pb2/' python/diagonal/proto/features_pb2.py > python/diagonal/proto/features_pb2.py.new
	mv python/diagonal/proto/features_pb2.py.new python/diagonal/proto/features_pb2.py
	sed -e 's/import geometry_pb2/import diagonal.proto.geometry_pb2/' python/diagonal/proto/api_pb2.py > python/diagonal/proto/api_pb2.py.new
	mv python/diagonal/proto/api_pb2.py.new python/diagonal/proto/api_pb2.py
	sed -e 's/import features_pb2/import diagonal.proto.features_pb2/' python/diagonal/proto/api_pb2.py > python/diagonal/proto/api_pb2.py.new
	mv python/diagonal/proto/api_pb2.py.new python/diagonal/proto/api_pb2.py
	sed -e 's/import api_pb2/import diagonal.proto.api_pb2/' python/diagonal/proto/api_pb2_grpc.py > python/diagonal/proto/api_pb2_grpc.py.new
	mv python/diagonal/proto/api_pb2_grpc.py.new python/diagonal/proto/api_pb2_grpc.py

ipython: python
	cd python; pip3 install . --upgrade --target ${HOME}/.ipython/

python-test: python fe
	PYTHONPATH=python python3 python/tests/all.py

test:
	make -C data test
	cd src/diagonal.works/diagonal; go test diagonal.works/diagonal/...

clean:
	find . -type f -perm +a+x | xargs rm

.PHONY: python
