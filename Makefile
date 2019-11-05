all:
	protoc -I=proto --go_out=src proto/geography.proto
	cd src/diagonal.works/diagonal; go build diagonal.works/diagonal/...
	cd src/diagonal.works/diagonal/experimental/publiclife; go build
	cd src/diagonal.works/diagonal/experimental/osm; go build

test:
	cd src/diagonal.works/diagonal; go test diagonal.works/diagonal/...

clean:
	find . -type f -perm +a+x | xargs rm

