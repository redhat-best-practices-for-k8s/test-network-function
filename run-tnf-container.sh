#!/usr/bin/env bash

# Requires
# - kubeconfig file mounted to /usr/tnf/kubeconfig/config
#   If more than one kubeconfig needs to be used, bind
#   additional volumes for each kubeconfig, e.g.
#     - /usr/tnf/kubeconfig/config
#     - /usr/tnf/kubeconfig/config.2
#     - /usr/tnf/kubeconfig/config.3
# - TNF config files mounted into /usr/tnf/config
# - A directory to output claim into mounted at /usr/tnf/claim
# - A $KUBECONFIG environment variable passed to the TNF container
#   containing all paths to kubeconfigs located in the container, e.g.
#   KUBECONFIG=/usr/tnf/kubeconfig/config:/usr/tnf/kubeconfig/config.2

export REQUIRED_NUM_OF_ARGS=5
export REQUIRED_VARS=('LOCAL_KUBECONFIG' 'LOCAL_TNF_CONFIG' 'OUTPUT_LOC')
export REQUIRED_VARS_ERROR_MESSAGES=(
	'KUBECONFIG is invalid or not given. Use the -k option to provide path to one or more kubeconfig files.'
	'TNFCONFIG is required. Use the -t option to specify the directory containing the TNF configuration files.'
	'OUTPUT_LOC is required. Use the -o option to specify the output location for the test results.'
)

export TNF_IMAGE_NAME=test-network-function
export TNF_IMAGE_TAG=latest
export TNF_OFFICIAL_ORG=quay.io/testnetworkfunction/
export TNF_OFFICIAL_IMAGE="${TNF_OFFICIAL_ORG}${TNF_IMAGE_NAME}:${TNF_IMAGE_TAG}"
export TNF_CMD="./run-cnf-suites.sh"
export OUTPUT_ARG="-o"

./run-container.sh "$@"