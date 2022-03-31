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

GO_PACKAGES=$(shell go list ./... | grep -v vendor)

.PHONY:	build \
	mocks \
	clean \
	lint \
	test \
	coverage-html \
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
	install-tools \
	vet

# Get default value of $GOBIN if not explicitly set
GO_PATH=$(shell go env GOPATH)
ifeq (,$(shell go env GOBIN))
  GOBIN=${GO_PATH}/bin
else
  GOBIN=$(shell go env GOBIN)
endif

COMMON_GO_ARGS=-race
GIT_COMMIT=$(shell git rev-list -1 HEAD)
GIT_RELEASE=$(shell git tag --points-at HEAD | head -n 1)
GIT_PREVIOUS_RELEASE=$(shell git tag --no-contains HEAD --sort=v:refname | tail -n 1)
GOLANGCI_VERSION=v1.45.2

# Run the unit tests and build all binaries
build:
	make test
	make build-cnf-tests


build-tnf-tool:
	go build -o tnf -v cmd/tnf/main.go

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
	rm -f ./tnf

# Run configured linters
lint:
	golangci-lint run --timeout 5m0s

# Build and run unit tests
test: mocks
	go build ${COMMON_GO_ARGS} ./...
	UNIT_TEST="true" go test -coverprofile=cover.out ./...

coverage-html: test
	go tool cover -html cover.out

# generate the test catalog in JSON
build-catalog-json: build-tnf-tool
	./tnf generate catalog json > catalog.json

# generate the test catalog in Markdown
build-catalog-md: build-tnf-tool
	./tnf generate catalog markdown > CATALOG.md

# build the CNF test binary
build-cnf-tests:
	PATH=${PATH}:${GOBIN} ginkgo build -ldflags "-X github.com/test-network-function/test-network-function/test-network-function.GitCommit=${GIT_COMMIT} -X github.com/test-network-function/test-network-function/test-network-function.GitRelease=${GIT_RELEASE} -X github.com/test-network-function/test-network-function/test-network-function.GitPreviousRelease=${GIT_PREVIOUS_RELEASE}" ./test-network-function 
	make build-catalog-md

build-cnf-tests-debug:
	PATH=${PATH}:${GOBIN} ginkgo build -gcflags "all=-N -l" -ldflags "-X github.com/test-network-function/test-network-function/test-network-function.GitCommit=${GIT_COMMIT} -X github.com/test-network-function/test-network-function/test-network-function.GitRelease=${GIT_RELEASE} -X github.com/test-network-function/test-network-function/test-network-function.GitPreviousRelease=${GIT_PREVIOUS_RELEASE} -extldflags '-z relro -z now'" ./test-network-function
	make build-catalog-md

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
	go install github.com/onsi/ginkgo/v2/ginkgo@v2.1.3
	go install github.com/onsi/gomega
	go install github.com/golang/mock/mockgen@v1.6.0
	wget https://get.helm.sh/helm-v3.8.1-linux-amd64.tar.gz && \
    tar -xvf helm-v3.8.1-linux-amd64.tar.gz && \
    cp linux-amd64/helm /usr/local/bin/helm
	
# Install golangci-lint	
install-lint:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ${GO_PATH}/bin ${GOLANGCI_VERSION}

vet:
	go vet ${GO_PACKAGES}
