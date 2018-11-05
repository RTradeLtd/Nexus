GO=env GO111MODULE=on go
GONOMOD=env GO111MODULE=off go
IPFSCONTAINERS=`docker ps -a -q --filter="name=ipfs-*"`
TESTCOMPOSE=https://raw.githubusercontent.com/RTradeLtd/Temporal/V2/test/docker-compose.yml
COMPOSECOMMAND=env ADDR_NODE1=1 ADDR_NODE2=2 docker-compose -f test/docker-compose.yml

all: deps check build

.PHONY: build
build:
	go build

.PHONY: install
install: deps
	go install

# Install dependencies
.PHONY: deps
deps:
	$(GO) mod vendor
	$(GO) get github.com/maxbrunsfeld/counterfeiter
	$(GO) mod tidy

# Run simple checks
.PHONY: check
check:
	go vet ./...
	go test -run xxxx ./...

# Execute tests
.PHONY: test
test:
	go test -race -cover ./...

.PHONY: testenv
testenv:
	mkdir -p test
	curl $(TESTCOMPOSE) --output test/docker-compose.yml
	$(COMPOSECOMMAND) up -d postgres

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

.PHONY: mock
mock:
	counterfeiter -o ./ipfs/mock/ipfs.mock.go \
		./ipfs/ipfs.go NodeClient
