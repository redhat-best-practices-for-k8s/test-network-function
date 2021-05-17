#!/usr/bin/env bash

# defaults
OUTPUT_LOC="$PWD/test-network-function"

usage() {
	echo "$0 [-o OUTPUT_LOC] SUITE [... SUITE]"
	echo "Call the script and list the test suites to run"
	echo "  e.g."
	echo "    $0 [ARGS] generic container"
	echo "  will run the generic and container suites"
	echo ""
	echo "Allowed suites are listed in the README."
}

usage_error() {
	usage
	exit 1
}

# Parge args beginning with "-"
while [[ $1 == -* ]]; do
	case "$1" in
		-h|--help|-\?) usage; exit 0;;
		-o) if (($# > 1)); then
				OUTPUT_LOC=$2; shift 2
			else
				echo "-o requires an argument" 1>&2
				exit 1
			fi ;;
		--) shift; break;;
		-*) echo "invalid option: $1" 1>&2; usage_error;;
	esac
done
# specify Junit report file name.
GINKGO_ARGS="-ginkgo.v -junit $OUTPUT_LOC -report $OUTPUT_LOC -claimloc $OUTPUT_LOC -ginkgo.reportFile $OUTPUT_LOC/cnf-certification-tests_junit.xml"
FOCUS=""

for var in "$@"
do
	case "$var" in
		diagnostic) FOCUS="diagnostic|$FOCUS";;
		generic) FOCUS="generic|$FOCUS";;
		multus) FOCUS="multus|$FOCUS";;
		operator) FOCUS="operator|$FOCUS";;
		container) FOCUS="container|$FOCUS";;
		*) usage_error;;
	esac
done

# If no focus is set then display usage and quit with a non-zero exit code.
[ -z "$FOCUS" ] && usage_error

FOCUS=${FOCUS%?}  # strip the trailing "|" from the concatenation

if [ "$VERIFY_CNF_FEATURES" == "true" ] && [ "$TNF_MINIKUBE_ONLY" != "true" ]; then

	export TNF_IMAGE_NAME=cnf-tests
	export TNF_IMAGE_TAG=latest
	export TNF_OFFICIAL_ORG=quay.io/openshift-kni/
	export TNF_OFFICIAL_IMAGE="${TNF_OFFICIAL_ORG}${TNF_IMAGE_NAME}:${TNF_IMAGE_TAG}"
	export TNF_CMD="/usr/bin/test-run.sh"
	export OUTPUT_ARG="--junit"

	./run-container.sh -d 127.0.0.53 -n host -o $OUTPUT_LOC -ginkgo.v -ginkgo.skip="performance|sriov|ptp|sctp|xt_u32|dpdk|ovn"

fi

echo "Running with focus '$FOCUS'. Report will be output to '$OUTPUT_LOC'"
cd ./test-network-function && ./test-network-function.test -ginkgo.focus="$FOCUS" ${GINKGO_ARGS}
