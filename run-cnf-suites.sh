#!/usr/bin/env bash

# defaults
OUTPUT_LOC="."

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

GINKGO_ARGS="-ginkgo.v -junit . -report $OUTPUT_LOC -claimloc $OUTPUT_LOC"
FOCUS=""

for var in "$@"
do
	case "$var" in
		diagnostic) FOCUS="diagnostic|$FOCUS";;
		generic) FOCUS="generic|$FOCUS";;
		multus) FOCUS="multus|$FOCUS";;
		operator) FOCUS="operator|$FOCUS";;
		container) FOCUS="container|$FOCUS";;
		turnium) FOCUS="turnium|$FOCUS";;
		*) usage_error;;
	esac
done

# If no focus is set then display usage and quit with a non-zero exit code.
[ -z "$FOCUS" ] && usage_error

FOCUS=${FOCUS%?}  # strip the trailing "|" from the concatenation

echo "Running with focus '$FOCUS'. Report will be output to '$OUTPUT_LOC'"
cd ./test-network-function && ./test-network-function.test -ginkgo.focus="$FOCUS" ${GINKGO_ARGS}
