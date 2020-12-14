#!/usr/bin/env bash
COMMON_GINKGO_ARGS="-ginkgo.v -junit . -report ."
FOCUS=""

usage() {
	echo "Call the script and list the test suites to run"
	echo "  e.g."
	echo "    $0 generic container"
	echo "  will run the generic and container suites"
	echo ""
	echo "Allowed suites are listed in the README."
	exit 1
}

for var in "$@"
do
	case "$var" in
	  diagnostic) FOCUS="diagnostic|$FOCUS"
	    ;;
		generic) FOCUS="generic|$FOCUS"
			;;
		multus) FOCUS="multus|$FOCUS"
			;;
		operator) FOCUS="operator|$FOCUS"
			;;
		container) FOCUS="container|$FOCUS"
			;;
		*) usage
			;;
	esac
done

[ -z "$FOCUS" ] && usage

FOCUS=${FOCUS%?}  # strip the trailing "|" from the concatenation

echo "running with focus '$FOCUS'"
cd ./test-network-function && ./test-network-function.test -ginkgo.focus="$FOCUS" ${COMMON_GINKGO_ARGS}