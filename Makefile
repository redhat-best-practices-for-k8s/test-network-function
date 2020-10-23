.PHONY: build \
	build-generic-cnf-tests \
	clean \
	cnf-tests \
	dependencies \
	deps-update \
	generic-cnf-tests \
	lint \
	mocks \
	run-cnf-tests \
	run-generic-cnf-tests \
	unit-tests

# Export GO111MODULE=on to enable project to be built from within GOPATH/src
export GO111MODULE=on

ifeq (,$(shell go env GOBIN))
  GOBIN=$(shell go env GOPATH)/bin
else
  GOBIN=$(shell go env GOBIN)
endif

export COMMON_GINKGO_ARGS=-ginkgo.v -junit . -report .
export COMMON_GO_ARGS=-race

build: lint
	go fmt ./...
	go build ${COMMON_GO_ARGS} ./...
	make unit-tests

generic-cnf-tests: build build-cnf-tests run-generic-cnf-tests

cnf-tests: build build-cnf-tests run-cnf-tests

operator-cnf-tests: build build-cnf-operator-tests run-operator-tests

build-cnf-tests:
	PATH=${PATH}:${GOBIN} ginkgo build ./test-network-function

build-cnf-operator-tests:
	PATH=${PATH}:${GOBIN} ginkgo build ./test-network-function/operator-test --tags operator_suite

run-generic-cnf-tests:
	cd ./test-network-function && ./test-network-function.test -ginkgo.focus="generic" ${COMMON_GINKGO_ARGS}

run-cnf-tests:
	cd ./test-network-function && ./test-network-function.test $COMMON_GINKGO_ARGS

run-operator-tests:
	cd ./test-network-function/operator-test && ./operator-test.test  $COMMON_GINKGO_ARGS

deps-update:
	go mod tidy && \
	go mod vendor

mocks:
	mockgen -source=pkg/tnf/interactive/spawner.go -destination=pkg/tnf/interactive/mocks/mock_spawner.go
	mockgen -source=pkg/tnf/test.go -destination=pkg/tnf/mocks/mock_tester.go
	mockgen -source=./internal/reel/reel.go -destination=./internal/reel/mocks/mock_reel.go

unit-tests:
	go test -coverprofile=cover.out `go list ./... | grep -v "github.com/redhat-nfvpe/test-network-function/test-network-function" | grep -v mock` && go tool cover -html=cover.out

lint:
	golint `go list ./... | grep -v vendor`

.PHONY: clean
clean:
	go clean
	rm -f ./test-network-function/test-network-function.test
	rm -f ./test-network-function/cnf-certification-tests_junit.xml
	rm -f ./test-network-function/operator-test/operator-test.test
	rm -f ./test-network-function/operator-test/cnf-operator-certification-tests_junit.xml

dependencies:
	go get github.com/onsi/ginkgo/ginkgo
	go get github.com/onsi/gomega/...
