#!/usr/bin/env bash

# Requires
# kube config file mounted to /usr/tnf/kubeconfig/config
# TNF config files mounted into /usr/tnf/config
# A directory to output claim into mounted at /usr/tnf/claim


usage() {
	echo "$0 -k KUBECONFIG -t TNFCONFIG -o OUTPUT_LOC SUITE [... SUITE]"
	echo "Configure and run the containerised TNF test offering."
	echo "  e.g."
	echo "    $0 -k ~/.kube/config -t ~/tnf/config -o ~/tnf/output diagnostic generic"
	echo "  will run the diagnostic and generic tests, and output the result into"
	echo "  ~/tnf/output on the host."
	echo ""
	echo "Allowed tests are listed in the README."
	echo "Note: Tests muse be specified after all other arguments"

}

usage_error() {
	usage
	exit 1
}

CONTAINER_TNF_DIR=/usr/tnf

if (($# < 7)); then
	usage_error
fi;

# Parge args beginning with -
while [[ $1 == -* ]]; do
	echo $1 $2
    case "$1" in
      -h|--help|-\?) usage; exit 0;;
      -k) if (($# > 1)); then
            LOCAL_KUBECONFIG=$2; shift 2
          else
            echo "-k requires an argument" 1>&2
            exit 1
          fi ;;
      -t) if (($# > 1)); then
            LOCAL_TNF_CONFIG=$2; shift 2
          else
            echo "-t requires an argument" 1>&2
            exit 1
          fi ;;
      -o) if (($# > 1)); then
            OUTPUT_LOC=$2; shift 2
          else
            echo "-o requires an argument" 1>&2
            exit 1
          fi ;;
      --) shift; break;;
      -*) echo "invalid option: $1" 1>&2; usage; exit 1;;
    esac
done

set -x
docker run --rm -v $LOCAL_TNF_CONFIG:$CONTAINER_TNF_DIR/config -v $LOCAL_KUBECONFIG:$CONTAINER_TNF_DIR/kubeconfig/config -v $OUTPUT_LOC:$CONTAINER_TNF_DIR/claim quay.io/testnetworkfunction/test-network-function:latest ./run-cnf-suites.sh -o $CONTAINER_TNF_DIR/claim $@
