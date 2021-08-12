#!/usr/bin/env bash

# defaults
export OUTPUT_LOC="$PWD/test-network-function"

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
GINKGO_ARGS="-ginkgo.v -junit $OUTPUT_LOC -claimloc $OUTPUT_LOC -ginkgo.reportFile $OUTPUT_LOC/cnf-certification-tests_junit.xml"
FOCUS=""

for var in "$@"
do
	FOCUS="$var|$FOCUS"
	# case "$var" in
	# 	diagnostic) FOCUS="diagnostic|$FOCUS";;
	# 	access-control) FOCUS="ac|$FOCUS";;
	# 	affiliated-certification) FOCUS="affiliated-certification|$FOCUS";;
	# 	lifecycle) FOCUS="lifecycle|$FOCUS";;
	# 	platform-alteration) FOCUS="platform-alteration|$FOCUS";;
	# 	generic) FOCUS="generic|$FOCUS";;
	# 	observability) FOCUS="observability|$FOCUS";;
	# 	operator) FOCUS="operator|$FOCUS";;
	# 	networking) FOCUS="networking|$FOCUS";;
	# 	*) usage_error;;
	# esac
done

# If no focus is set then display usage and quit with a non-zero exit code.
[ -z "$FOCUS" ] && usage_error

FOCUS=${FOCUS%?}  # strip the trailing "|" from the concatenation

# Run cnf-feature-deploy test container if not running inside a container
# cgroup file doesn't exist on MacOS. Consider that as not running in container as well
if [[ ! -f "/proc/1/cgroup" ]] || grep -q init\.scope /proc/1/cgroup; then
	cd script
	./run-cfd-container.sh
	cd ..
fi

if [[ -z "${TNF_PARTNER_SRC_DIR}" ]]; then
	echo "env var \"TNF_PARTNER_SRC_DIR\" not set, please set it before running this command!"
	exit 1
fi

make -C $TNF_PARTNER_SRC_DIR install-partner-pods

echo "Running with focus '$FOCUS'. Report will be output to '$OUTPUT_LOC'"
cd ./test-network-function && ./test-network-function.test -ginkgo.focus="$FOCUS" ${GINKGO_ARGS}
