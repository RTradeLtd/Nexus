GO=env GO111MODULE=on go

all: deps check

# Install dependencies
.PHONY: deps
deps:
	$(GO) mod vendor

# Run simple checks
.PHONY: check
check:
	go vet ./...
	go test -run xxxx ./...

# Execute tests
.PHONY: test
test:
	go test -race -cover ./...

# Generate protobuf code from definitions
.PHONY: proto
proto:
	protoc -I protobuf service.proto --go_out=plugins=grpc:protobuf
