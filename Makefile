export GO111MODULE=on

# The name of the executable (default is current directory name)
TARGET := $(shell echo $${PWD\#\#*/})
.DEFAULT_GOAL: $(TARGET)

# go source files, ignore vendor directory
SRC = $(shell find . -type f -name '*.go' -not -path "./vendor/*")

# protobuf sources
PROTO_SRCS := $(shell find . -name *.proto)

.PHONY: all build clean install uninstall fmt simplify check deps generate run proto

all: check build

$(TARGET): $(SRC)
	@go build -o build/$(TARGET) -tags=jsoniter

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
ifeq ($(shell which protoc-gen-go), )
	@echo "Installing dependency: protoc-gen-go"
	@go get -u github.com/golang/protobuf/protoc-gen-go
endif
ifeq ($(shell which protoc), )
	$(error protoc is not installed. You must install it manually on https://developers.google.com/protocol-buffers/)
endif

proto: deps
	@for PROTO in $(PROTO_SRCS); do \
		protoc -I/usr/local/include -I. \
			--go_out=plugins=grpc:. \
			$$PROTO; \
	done

generate: deps
	$(eval GENERATE_SRCS := $(shell git --no-pager grep -wl "go:generate" -- "*.go" ":(exclude)vendor"))
	@for GENERATE_SRC in $(GENERATE_SRCS); do \
		go generate $$GENERATE_SRC; \
	done

run: build
	@./build/$(TARGET)

test: test-all

test-all:
	@go test -v `go list ./... | grep -v test/e2e`

test-e2e:
	@go test -v `go list ./test/e2e` $(FLAGS)
