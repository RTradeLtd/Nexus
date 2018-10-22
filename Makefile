GO=env GO111MODULE=on go
IPFSCONTAINERS=`docker ps -a -q --filter="name=ipfs-*"`

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

# Clean up containers and things
.PHONY: clean
clean:
	docker stop $(IPFSCONTAINERS) || true
	docker rm $(IPFSCONTAINERS) || true
	rm -f ./ipfs-orchestrator
	find . -name tmp -type d -exec rm -f -r {} +
