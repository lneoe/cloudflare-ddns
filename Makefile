# workdir info
PACKAGE=ddns
PREFIX=$(shell pwd)
CMD_PACKAGE=${PACKAGE}
OUTPUT_DIR=${PREFIX}/bin
OUTPUT_FILE=${OUTPUT_DIR}/ddns
COMMIT_ID=$(shell git rev-parse --short HEAD)
VERSION=$(shell git describe --tags || echo "v0.0.0")
VERSION_IMPORT_PATH=main
BUILD_TIME=$(shell date '+%Y-%m-%dT%H:%M:%S%Z')
VCS_BRANCH=$(shell git symbolic-ref --short -q HEAD)

# build args
BUILD_ARGS := \
    -ldflags "-w -s -buildid= \
    -X $(VERSION_IMPORT_PATH).appName=$(PACKAGE) \
    -X $(VERSION_IMPORT_PATH).version=$(VERSION) \
    -X $(VERSION_IMPORT_PATH).revision=$(COMMIT_ID) \
    -X $(VERSION_IMPORT_PATH).branch=$(VCS_BRANCH) \
    -X $(VERSION_IMPORT_PATH).buildDate=$(BUILD_TIME)"
EXTRA_BUILD_ARGS=

# which golint
GOLINT=$(shell which golangci-lint || echo '')

export GOCACHE=
export GOPROXY=
export GOSUMDB=

default: lint test build

lint:
	@echo "+ $@"
	@$(if $(GOLINT), , \
		$(error Please install golint: "https://golangci-lint.run/usage/install"))
	golangci-lint run --deadline=10m -E gofmt  -E errcheck ./...

test:
	@echo "+ test"
	go test -cover $(EXTRA_BUILD_ARGS) ./...

.PHONY:build
linux-armv7:
	@echo "+ $@"
	GOARCH=arm GOOS=linux GOARM=7 \
	go build -trimpath $(BUILD_ARGS) $(EXTRA_BUILD_ARGS) \
		-o ${OUTPUT_FILE}-$@ $(CMD_PACKAGE)

build:
	@echo "+ build"
	go build $(BUILD_ARGS) $(EXTRA_BUILD_ARGS) -o ${OUTPUT_FILE} $(CMD_PACKAGE)

clean:
	@echo "+ $@"
	@rm -r "${OUTPUT_DIR}"
