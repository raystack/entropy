NAME=github.com/goto/entropy
VERSION=$(shell git describe --tags --always --first-parent 2>/dev/null)
COMMIT=$(shell git rev-parse --short HEAD)
PROTON_COMMIT="5b5dc727b525925bcec025b355983ca61d7ccf68"
BUILD_TIME=$(shell date)
COVERAGE_DIR=coverage
BUILD_DIR=dist
EXE=entropy

.PHONY: all build clean tidy format test test-coverage proto

all: format clean test build 

tidy:
	@echo "Tidy up go.mod..."
	@go mod tidy -v

install: ## install required dependencies
	@echo "> installing dependencies"
	go install github.com/vektra/mockery/v2@v2.14.0
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.1
	go get -d google.golang.org/protobuf/proto@v1.28.1
	go get -d google.golang.org/grpc@v1.49.0
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2.0
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.11.3
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.11.3
	go install github.com/bufbuild/buf/cmd/buf@v1.7.0
	go install github.com/envoyproxy/protoc-gen-validate@v0.6.7
	
format:
	@echo "Running gofumpt..."
	@gofumpt -l -w .

lint:
	@echo "Running lint checks using golangci-lint..."
	@golangci-lint run

clean: tidy
	@echo "Cleaning up build directories..."
	@rm -rf ${COVERAGE_DIR} ${BUILD_DIR}

generate:
	@echo "Running go-generate..."
	@go generate ./...

test: tidy
	@mkdir -p ${COVERAGE_DIR}
	@echo "Running unit tests..."
	@go test ./... -coverprofile=${COVERAGE_DIR}/coverage.out

test-coverage: test
	@echo "Generating coverage report..."
	@go tool cover -html=${COVERAGE_DIR}/coverage.out

build: clean
	@mkdir -p ${BUILD_DIR}
	@echo "Running build for '${VERSION}' in '${BUILD_DIR}/'..."
	@CGO_ENABLED=0 go build -ldflags '-X "${NAME}/pkg/version.Version=${VERSION}" -X "${NAME}/pkg/version.Commit=${COMMIT}" -X "${NAME}/pkg/version.BuildTime=${BUILD_TIME}"' -o ${BUILD_DIR}/${EXE}

proto: ## Generate the protobuf files
	@echo " > generating protobuf from goto/proton"
	@echo " > [info] make sure correct version of dependencies are installed using 'make install'"
	@rm -rf ./proto
	@buf generate https://github.com/goto/proton/archive/${PROTON_COMMIT}.zip#strip_components=1 --template buf.gen.yaml --path gotocompany/entropy --path gotocompany/common
	@echo " > protobuf compilation finished"

download:
	@go mod download

setup:
	@go install github.com/vektra/mockery/v2@v2.10.4
