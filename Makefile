NAME=github.com/odpf/entropy
VERSION=$(shell git describe --tags --always --first-parent 2>/dev/null)
COMMIT=$(shell git rev-parse --short HEAD)
BUILD_TIME=$(shell date)
COVERAGE_DIR=coverage
BUILD_DIR=dist
EXE=entropy

.PHONY: all build clean

all: clean test build

build:
	mkdir -p ${BUILD_DIR}
	CGO_ENABLED=0 go build -ldflags '-X "${NAME}/pkg/version.Version=${VERSION}" -X "${NAME}/pkg/version.Commit=${COMMIT}" -X "${NAME}/pkg/version.BuildTime=${BUILD_TIME}"' -o ${BUILD_DIR}/${EXE}

clean:
	rm -rf ${COVERAGE_DIR} ${BUILD_DIR}

download:
	go mod download

test:
	mkdir -p ${COVERAGE_DIR}
	go test ./... -coverprofile=${COVERAGE_DIR}/coverage.out

test-coverage: test
	go tool cover -html=${COVERAGE_DIR}/coverage.out
