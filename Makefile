all: build

.PHONY: build
build: 
	go build -o tparallel ./cmd/tparallel/

.PHONY: test
test: 
	go test -v ./...
