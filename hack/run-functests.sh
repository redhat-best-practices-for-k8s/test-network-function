#!/bin/bash

# Heavily inspired by: https://github.com/openshift-kni/cnf-features-deploy/blob/master/hack/run-functests.sh

set -e
. $(dirname "$0")/common.sh

if ! which go; then
  echo "No go command available"
  exit 1
fi

GOPATH="${GOPATH:-~/go}"
export GOFLAGS="${GOFLAGS:-"-mod=vendor"}"

export PATH=$PATH:$GOPATH/bin

if ! which gingko; then
	echo "Downloading ginkgo tool"
	go get github.com/onsi/ginkgo/ginkgo
fi

ginkgo build ./configsuite

mkdir -p cnf-tests/bin
mv ./configsuite/configsuite.test ./cnf-tests/bin/cnftests

./cnf-tests/bin/cnftests -junit . -ginkgo.v