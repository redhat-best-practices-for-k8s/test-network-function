.PHONY: build-cnf-tests \
	cnf-tests \
	dependencies \
	deps-update \
	mocks-clean \
	run-cnf-tests \
	run-generic-cnf-tests \
	run-container-tests \
	run-operator-tests

ifeq (,$(shell go env GOBIN))
  GOBIN=$(shell go env GOPATH)/bin
else
  GOBIN=$(shell go env GOBIN)
endif

COMMON_GO_ARGS=-race

build:
	make mocks
	go fmt ./...
	make lint
	go build ${COMMON_GO_ARGS} ./...
	make unit-tests

json-catalog:
	go run cmd/catalog/main.go generate json > catalog.json

markdown-catalog:
	go run cmd/catalog/main.go generate markdown > CATALOG.md

cnf-tests: build build-cnf-tests run-cnf-tests

generic-cnf-tests: build build-cnf-tests run-generic-cnf-tests

operator-cnf-tests: build build-cnf-tests run-operator-tests

container-cnf-tests: build build-cnf-tests run-container-tests

.PHONY: build-cnf-tests
build-cnf-tests:
	PATH=${PATH}:${GOBIN} ginkgo build ./test-network-function

.PHONY: run-generic-cnf-tests
run-generic-cnf-tests:
	./run-cnf-suites.sh diagnostic generic

.PHONY: run-cnf-tests
run-cnf-tests:
	./run-cnf-suites.sh diagnostic generic multus operator container

.PHONY: run-operator-tests
run-operator-tests:
	./run-cnf-suites.sh diagnostic operator

.PHONY: run-container-tests
run-container-tests:
	./run-cnf-suites.sh diagnostic container

deps-update:
	go mod tidy && \
	go mod vendor

mocks:
	make mocks-clean
	mkdir -p pkg/tnf/interactive/mocks
	mockgen -source=pkg/tnf/interactive/spawner.go -destination=pkg/tnf/interactive/mocks/mock_spawner.go
	mkdir -p pkg/tnf/mocks
	mockgen -source=pkg/tnf/test.go -destination=pkg/tnf/mocks/mock_tester.go
	mkdir -p pkg/tnf/reel/mocks
	mockgen -source=pkg/tnf/reel/reel.go -destination=pkg/tnf/reel/mocks/mock_reel.go

.PHONY: mocks-clean
mocks-clean:
	rm -f pkg/tnf/interactive/mocks/mock_spawner.go
	rm -f pkg/tnf/mocks/mock_tester.go
	rm -f pkg/tnf/reel/mocks/mock_reel.go


unit-tests:
	go test -coverprofile=cover.out `go list ./... | grep -v "github.com/redhat-nfvpe/test-network-function/test-network-function" | grep -v mock`

lint:
	golint `go list ./... | grep -v vendor`
	golangci-lint run

jsontest-cli:
	go build -o jsontest-cli -v cmd/generic/main.go

clean:
	go clean
	make mocks-clean
	rm -f ./test-network-function/test-network-function.test
	rm -f ./test-network-function/cnf-certification-tests_junit.xml

dependencies:
	go get github.com/onsi/ginkgo/ginkgo
	go get github.com/onsi/gomega/...
	go get golang.org/x/lint/golint
	go get github.com/golang/mock/mockgen
