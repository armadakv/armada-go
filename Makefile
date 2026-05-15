LDFLAGS = -X github.com/armadakv/armada-go/client.Version=$(VERSION)
VERSION ?= $(shell git describe --tags --always --dirty)
CGO_ENABLED ?= 0
ARMADA_PROTO_SRC_DIR = internal/proto
ARMADA_PROTO_VERSION = v0.10.0

.PHONY: all
all: getproto test build

.PHONY: test
test:
	go test ./... -cover -race -v

.PHONY: build
build:
	test $(VERSION) || (echo "version not set"; exit 1)
	CGO_ENABLED=$(CGO_ENABLED) go build -ldflags="$(LDFLAGS)" -v ./...


.PHONY: getproto-cleanup
# Cleanup temporary directory
getproto-cleanup:
	rm -Rf ${ARMADA_PROTO_SRC_DIR}

.PHONY: getproto
getproto: getproto-cleanup
	mkdir -p ${ARMADA_PROTO_SRC_DIR}
	curl -sSL https://api.github.com/repos/armadakv/armada/tarball/${ARMADA_PROTO_VERSION} | tar -x --strip-components=2 -C ${ARMADA_PROTO_SRC_DIR} */armadapb/*.pb.go

