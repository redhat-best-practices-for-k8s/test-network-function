.PHONY: build \
	cnftests \
	deps-update \
	clean

# Export GO111MODULE=on to enable project to be built from within GOPATH/src
export GO111MODULE=on

build:
	go build ./...

cnftests: build build-cnftests

build-cnftests:
	ginkgo build ./test-network-functions

deps-update:
	go mod tidy && \
	go mod vendor

.PHONY: clean
clean:
	go clean
	rm -rf ./cnf-tests
