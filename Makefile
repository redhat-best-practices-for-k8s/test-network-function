# There are four main groups of operations provided by this Makefile: build,
# clean, run and tasks.

# Build operations will create artefacts from code. This includes things such as
# binaries, mock files, or catalogs of CNF tests.

# Clean operations remove the results of the build tasks, or other files not
# considered permanent.

# Run operations provide shortcuts to execute built binaries in common
# configurations or with default options. They are part convenience and part
# documentation.

# Tasks provide shortcuts to common operations that occur frequently during
# development. This includes running configured linters and executing unit tests

.PHONY:	build \
	mocks \
	clean \
	lint \
	test \
	build-jsontest-cli \
	build-catalog-json \
	build-catalog-md \
	build-cnf-tests \
	run-cnf-tests \
	run-generic-cnf-tests \
	run-container-tests \
	run-operator-tests \
	run-generic-cnf-tests \
	run-operator-tests \
	run-container-tests \
	clean-mocks \
	update-deps \
	install-tools

# Get default value of $GOBIN if not explicitly set
ifeq (,$(shell go env GOBIN))
  GOBIN=$(shell go env GOPATH)/bin
else
  GOBIN=$(shell go env GOBIN)
endif

COMMON_GO_ARGS=-race

# Run the unit tests and build all binaries
build:
	make test
	make build-cnf-tests
	make build-jsontest-cli

# (Re)generate mock files as needed
mocks: pkg/tnf/interactive/mocks/mock_spawner.go \
    pkg/tnf/mocks/mock_tester.go \
    pkg/tnf/reel/mocks/mock_reel.go

# Cleans up auto-generated and report files
clean:
	go clean
	make clean-mocks
	rm -f ./test-network-function/test-network-function.test
	rm -f ./test-network-function/cnf-certification-tests_junit.xml

# Run configured linters
lint:
	golint `go list ./... | grep -v vendor`
	golangci-lint run

# Build and run unit tests
test: mocks
	go build ${COMMON_GO_ARGS} ./...
	go test -coverprofile=cover.out `go list ./... | grep -v "github.com/test-network-function/test-network-function/test-network-function" | grep -v mock`


# build the binary that can be used to run JSON-defined tests.
build-jsontest-cli:
	go build -o jsontest-cli -v cmd/generic/main.go

# generate the test catalog in JSON
build-catalog-json:
	go run cmd/catalog/main.go generate json > catalog.json

# generate the test catalog in Markdown
build-catalog-md:
	go run cmd/catalog/main.go generate markdown > CATALOG.md

# build the CNF test binary
build-cnf-tests:
	PATH=${PATH}:${GOBIN} ginkgo build ./test-network-function


# run all CNF tests
run-cnf-tests: build-cnf-tests
	./run-cnf-suites.sh diagnostic generic multus operator container

# run only the generic CNF tests
run-generic-cnf-tests: build-cnf-tests
	./run-cnf-suites.sh diagnostic generic

# Run operator CNF tests
run-operator-tests: build-cnf-tests
	./run-cnf-suites.sh diagnostic operator

# Run container CNF tests
run-container-tests: build-cnf-tests
	./run-cnf-suites.sh diagnostic container

# Each mock depends on one source file
pkg/tnf/interactive/mocks/mock_spawner.go: pkg/tnf/interactive/spawner.go
	mockgen -source=pkg/tnf/interactive/spawner.go -destination=pkg/tnf/interactive/mocks/mock_spawner.go

pkg/tnf/mocks/mock_tester.go: pkg/tnf/test.go
	mockgen -source=pkg/tnf/test.go -destination=pkg/tnf/mocks/mock_tester.go

pkg/tnf/reel/mocks/mock_reel.go: pkg/tnf/reel/reel.go
	mockgen -source=pkg/tnf/reel/reel.go -destination=pkg/tnf/reel/mocks/mock_reel.go

# Remove generated mock files
clean-mocks:
	rm -f pkg/tnf/interactive/mocks/mock_spawner.go
	rm -f pkg/tnf/mocks/mock_tester.go
	rm -f pkg/tnf/reel/mocks/mock_reel.go


# Update source dependencies and fix versions
update-deps:
	make mocks
	go mod tidy && \
	go mod vendor

# Install build tools and other required software.
install-tools:
	go get github.com/onsi/ginkgo/ginkgo
	go get github.com/onsi/gomega/...
	go get golang.org/x/lint/golint
	go get github.com/golang/mock/mockgen
