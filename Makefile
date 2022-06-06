NAME=github.com/odpf/entropy
VERSION=$(shell git describe --tags --always --first-parent 2>/dev/null)
COMMIT=$(shell git rev-parse --short HEAD)
BUILD_TIME=$(shell date)
COVERAGE_DIR=coverage
BUILD_DIR=dist
EXE=entropy

.PHONY: all build clean tidy format test test-coverage

all: clean test build format lint

tidy:
	@echo "Tidy up go.mod..."
	@go mod tidy -v

format:
	@echo "Running gofumpt..."
	@gofumpt -l -w .

lint:
	@echo "Running lint checks using golangci-lint..."
	@golangci-lint run

clean: tidy
	@echo "Cleaning up build directories..."
	@rm -rf ${COVERAGE_DIR} ${BUILD_DIR}
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

download:
	go mod download


