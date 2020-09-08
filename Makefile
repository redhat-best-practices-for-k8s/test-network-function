.PHONY: build \
	build-generic-cnf-tests \
	clean \
	cnf-tests \
	deps-update \
	generic-cnf-tests \
	run-cnf-tests \
	run-generic-cnf-tests

# Export GO111MODULE=on to enable project to be built from within GOPATH/src
export GO111MODULE=on

export COMMON_GINKGO_ARGS="-ginkgo.v -junit . -report ."
export COMMON_GO_ARGS="-race"

build:
	go fmt ./...
	go build ${COMMON_GO_ARGS} ./...

generic-cnf-tests: build build-cnf-tests run-generic-cnf-tests

cnf-tests: build build-cnf-tests run-cnf-tests

build-cnf-tests:
	ginkgo build ./test-network-function

run-generic-cnf-tests:
	cd ./test-network-function && ./test-network-function.test -ginkgo.focus="generic" $COMMON_GINKGO_ARGS

run-cnf-tests:
	cd ./test-network-function && ./test-network-function.test $COMMON_GINKGO_ARGS

deps-update:
	go mod tidy && \
	go mod vendor

.PHONY: clean
clean:
	go clean
	rm -f ./test-network-function/test-network-function.test
	rm -f ./test-network-function/cnf-certification-tests_junit.xml
