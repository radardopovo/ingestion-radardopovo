# Radar do Povo ETL - https://radardopovo.com

.PHONY: build vet test run clean

build:
	go build -o bin/importer ./cmd/importer

vet:
	go vet ./...

test:
	go test ./...

run:
	go run ./cmd/importer

clean:
	rm -f bin/importer
