all: experimental
	protoc -I=proto --go_out=src proto/geography.proto
	protoc -I=proto --go_out=src proto/tiles.proto
	protoc -I=proto --go_out=src proto/osm.proto
	protoc -I=src/diagonal.works/diagonal/osm --go_out=src src/diagonal.works/diagonal/osm/import.proto
	cd src/diagonal.works/diagonal; go build diagonal.works/diagonal/...
	cd src/diagonal.works/diagonal/cmd/fe; go build
	cd src/diagonal.works/diagonal/cmd/osm; go build
	cd src/diagonal.works/diagonal/cmd/osmbeam; go build
	cd src/diagonal.works/diagonal/cmd/inspect; go build
	make -C data

experimental:
	cd src/diagonal.works/diagonal/experimental/mr; go build

test:
	cd src/diagonal.works/diagonal; go test -v diagonal.works/diagonal/...

clean:
	find . -type f -perm +a+x | xargs rm

