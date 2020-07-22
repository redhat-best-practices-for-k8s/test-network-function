#!/bin/bash

# Directly lifted from:  https://github.com/openshift-kni/cnf-features-deploy/blob/master/hack/common.sh

set -e

pushd .
cd "$(dirname "$0")/.."

function finish {
    popd
}
trap finish EXIT
