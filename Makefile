.PHONY: build \
	build-generic-cnf-tests \
	clean \
	cnf-tests \
	deps-update \
	generic-cnf-tests \
	mocks \
	run-cnf-tests \
	run-generic-cnf-tests \
	unit-tests

# Export GO111MODULE=on to enable project to be built from within GOPATH/src
export GO111MODULE=on

export COMMON_GINKGO_ARGS=-ginkgo.v -junit . -report .
export COMMON_GO_ARGS=-race

build:
	go fmt ./...
	go build ${COMMON_GO_ARGS} ./...

generic-cnf-tests: build build-cnf-tests run-generic-cnf-tests

cnf-tests: build build-cnf-tests run-cnf-tests

build-cnf-tests:
	ginkgo build ./test-network-function

run-generic-cnf-tests:
	cd ./test-network-function && ./test-network-function.test -ginkgo.focus="generic" ${COMMON_GINKGO_ARGS}

run-cnf-tests:
	cd ./test-network-function && ./test-network-function.test $COMMON_GINKGO_ARGS

deps-update:
	go mod tidy && \
	go mod vendor

mocks:
	mockgen -source=pkg/tnf/interactive/spawner.go -destination=pkg/tnf/interactive/mocks/mock_spawner.go
	mockgen -source=pkg/tnf/test.go -destination=pkg/tnf/mocks/mock_tester.go
	mockgen -source=./internal/reel/reel.go -destination=./internal/reel/mocks/mock_reel.go

unit-tests:
	go test -coverprofile=cover.out `go list ./... | grep -v "github.com/redhat-nfvpe/test-network-function/test-network-function" | grep -v mock` && go tool cover -html=cover.out

.PHONY: clean
clean:
	go clean
	rm -f ./test-network-function/test-network-function.test
	rm -f ./test-network-function/cnf-certification-tests_junit.xml
