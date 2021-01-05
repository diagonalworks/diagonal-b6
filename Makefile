all: protos experimental fe ingest transit fe-js
	cd src/diagonal.works/diagonal; go build diagonal.works/diagonal/...
	cd src/diagonal.works/diagonal/cmd/osm; go build
	cd src/diagonal.works/diagonal/cmd/osmbeam; go build
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

protos:
	protoc --plugin=${HOME}/go/bin/protoc-gen-go -I=proto --go_out=src proto/geography.proto
	protoc --plugin=${HOME}/go/bin/protoc-gen-go -I=proto --go_out=src proto/tiles.proto
	protoc --plugin=${HOME}/go/bin/protoc-gen-go -I=proto --go_out=src proto/osm.proto
	protoc --plugin=${HOME}/go/bin/protoc-gen-go -I=proto --go_out=src proto/geometry.proto
	protoc --plugin=${HOME}/go/bin/protoc-gen-go -I=proto --go_out=src proto/features.proto
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

test:
	cd src/diagonal.works/diagonal; go test -v diagonal.works/diagonal/...

clean:
	find . -type f -perm +a+x | xargs rm

