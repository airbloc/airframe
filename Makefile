export GO111MODULE=on

# The name of the executable (default is current directory name)
TARGET := $(shell echo $${PWD\#\#*/})
.DEFAULT_GOAL: $(TARGET)

# go source files, ignore vendor directory
SRC = $(shell find . -type f -name '*.go' -not -path "./vendor/*")

# sources that needs to be run with go generate
GENERATE_SRCS := $(shell grep -rwl --exclude-dir={./build,./vendor} --include=*.go . -e "go:generate")

.PHONY: all build clean install uninstall fmt simplify check deps generate run

all: check build

$(TARGET): $(SRC)
	@go build -o build/$(TARGET)

build: $(TARGET)
	@true

clean:
	@go clean
	@rm -rf build/

install: build
	@cp build/$(TARGET) $(GOPATH)/bin/

uninstall: clean
	@rm -f $$(which ${TARGET})

fmt:
	@gofmt -l -w $(SRC)

simplify:
	@gofmt -s -l -w $(SRC)

check:
	@go tool vet ${SRC}

deps:
ifeq ($(shell which easyjson), )
	@echo "Installing dependency: easyjson"
	@go get -u github.com/mailru/easyjson/...
endif

generate: deps
	@for GENERATE_SRC in $(GENERATE_SRCS); do \
		go generate $$GENERATE_SRC; \
	done

run: build
	@./build/$(TARGET)

test: test-all

test-all:
	@go test -v `go list ./... | grep -v test/e2e`

test-e2e:
	@go test -v `go list ./test/e2e`
