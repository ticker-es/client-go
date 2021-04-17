
# Makefile for client-go

VERSION ?= 0.0.0
BINARY_NAME ?= ticker

build:

test:

archive:

release:

build-prerequisites:
	mkdir -p bin dist

release-prerequisites:

test-prerequisites:

install-tools:

### BUILD ###################################################################

generate-rpc:
	protoc -I../rpc \
		--go_out=rpc --go_opt=paths=import --go_opt=module=github.com/ticker-es/client-go/rpc \
		--go-grpc_out=rpc --go-grpc_opt=paths=import --go-grpc_opt=module=github.com/ticker-es/client-go/rpc \
		../rpc/maintenance.proto \
		../rpc/event_stream.proto

build-client-go: build-prerequisites
	go build -ldflags "-X main.version=${VERSION} -X main.commit=$$(git rev-parse --short HEAD 2>/dev/null || echo \"none\")" -o bin/$(OUTPUT_DIR)$(BINARY_NAME) cli/*.go
build-client-go-linux_amd64: build-prerequisites
	$(MAKE) GOOS=linux GOARCH=amd64 OUTPUT_DIR=linux_amd64/ build
build-client-go-darwin_amd64: build-prerequisites
	$(MAKE) GOOS=darwin GOARCH=amd64 OUTPUT_DIR=darwin_amd64/ build
build-client-go-windows_amd64: build-prerequisites
	$(MAKE) GOOS=windows GOARCH=amd64 OUTPUT_DIR=windows_amd64/ build

build-linux_amd64: build-client-go-linux_amd64
build-darwin_amd64: build-client-go-darwin_amd64
build-windows_amd64: build-client-go-windows_amd64

build: build-client-go
build-all: build-linux_amd64 build-darwin_amd64 build-windows_amd64

### ARCHIVE #################################################################

archive-client-go-linux_amd64: build-client-go-linux_amd64
	tar czf dist/$(BINARY_NAME)-${VERSION}-linux_amd64.tar.gz -C bin/linux_amd64/ .
archive-client-go-darwin_amd64: build-client-go-darwin_amd64
	tar czf dist/$(BINARY_NAME)-${VERSION}-darwin_amd64.tar.gz -C bin/darwin_amd64/ .
archive-client-go-windows_amd64: build-client-go-windows_amd64
	tar czf dist/$(BINARY_NAME)-${VERSION}-windows_amd64.tar.gz -C bin/windows_amd64/ .

archive-linux_amd64: archive-client-go-linux_amd64
archive-darwin_amd64: archive-client-go-darwin_amd64
archive-windows_amd64: archive-client-go-windows_amd64

archive: archive-linux_amd64 archive-darwin_amd64 archive-windows_amd64

release: archive
	sha1sum dist/*.tar.gz > dist/$(BINARY_NAME)-${VERSION}.shasums

### TEST ####################################################################

test-client-go:
	ginkgo -r
test-client-go-watch:
	ginkgo watch
test: test-client-go
.PHONY: test-client-go
.PHONY: test

clean:
	rm -r bin/* dist/*

### DATABASE ################################################################

db-up:
	psql < db/up.sql

db-down:
	psql < db/down.sql

db-seed:
	psql < db/seed.sql

