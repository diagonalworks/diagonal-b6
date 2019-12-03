all:
	protoc -I=proto --go_out=src proto/geography.proto
	protoc -I=proto --go_out=src proto/tiles.proto
	cd src/diagonal.works/diagonal; go build diagonal.works/diagonal/...
	cd src/diagonal.works/diagonal/cmd/fe; go build
	cd src/diagonal.works/diagonal/cmd/osm; go build
	cd src/diagonal.works/diagonal/experimental/publiclife; go build

test:
	cd src/diagonal.works/diagonal; go test -v diagonal.works/diagonal/...

clean:
	find . -type f -perm +a+x | xargs rm

