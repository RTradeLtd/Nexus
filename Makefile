GO=env GO111MODULE=on go
GONOMOD=env GO111MODULE=off go
IPFSCONTAINERS=`docker ps -a -q --filter="name=ipfs-*"`
COMPOSECOMMAND=env ADDR_NODE1=1 ADDR_NODE2=2 docker-compose -f testenv/docker-compose.yml
VERSION=`git describe --always --tags`

all: deps check build

.PHONY: build
build:
	go build -ldflags "-X main.Version=$(VERSION)" ./cmd/nexus

.PHONY: install
install: deps
	go install -ldflags "-X main.Version=$(VERSION)" ./cmd/nexus

# Install dependencies
.PHONY: deps
deps:
	$(GO) mod vendor
	$(GO) get github.com/UnnoTed/fileb0x
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
	$(COMPOSECOMMAND) up -d postgres

# Clean up containers and things
.PHONY: clean
clean:
	$(COMPOSECOMMAND) down
	docker stop $(IPFSCONTAINERS) || true
	docker rm $(IPFSCONTAINERS) || true
	rm -f ./nexus
	find . -name tmp -type d -exec rm -f -r {} +

# Gen runs all code generators
.PHONY: gen
gen:
	fileb0x b0x.yml
	counterfeiter -o ./ipfs/mock/ipfs.mock.go \
		./ipfs/ipfs.go NodeClient
	counterfeiter -o ./temporal/mock/access.mock.go \
		./temporal/database.go AccessChecker
	counterfeiter -o ./temporal/mock/networks.mock.go \
		./temporal/database.go PrivateNetworks

.PHONY: release
release:
	bash .scripts/release.sh

#####################
# DEVELOPMENT UTILS #
#####################

NETWORK=test_network
TESTFLAGS=-dev -config ./config.dev.json

.PHONY: example-config
example-config: build
	./nexus -config ./config.example.json init

.PHONY: dev-config
dev-config: build
	./nexus $(TESTFLAGS) init

.PHONY: config
config: example-config dev-config

.PHONY: daemon
daemon: build
	./nexus $(TESTFLAGS) daemon

.PHONY: new-network
new-network: build
	./nexus $(TESTFLAGS) dev network $(NETWORK)

.PHONY: start-network
start-network: build
	./nexus $(TESTFLAGS) ctl --pretty StartNetwork Network=$(NETWORK)

.PHONY: stat-network
stat-network:
	./nexus $(TESTFLAGS) ctl --pretty NetworkStats Network=$(NETWORK)

.PHONY: diag-network
diag-network:
	./nexus $(TESTFLAGS) ctl NetworkDiagnostics Network=$(NETWORK)
