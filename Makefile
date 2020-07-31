.PHONY: build \
	cnftests \
	build-cnftests \
	run-cnftests \
	deps-update \
	clean

# Export GO111MODULE=on to enable project to be built from within GOPATH/src
export GO111MODULE=on

build:
	go build ./...

cnftests: build build-cnftests run-cnftests

build-cnftests:
	ginkgo build ./test-network-function

run-cnftests:
	cd ./test-network-function && ./test-network-function.test -ginkgo.v -junit . -report .

deps-update:
	go mod tidy && \
	go mod vendor

.PHONY: clean
clean:
	go clean
