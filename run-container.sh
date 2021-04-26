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

REQUIRED_VARS=('LOCAL_KUBECONFIG' 'LOCAL_TNF_CONFIG' 'OUTPUT_LOC')
REQUIRED_VARS_ERROR_MESSAGES=(
	'KUBECONFIG is invalid or not given. Use the -k option to provide path to one or more kubeconfig files.'
	'TNFCONFIG is required. Use the -t option to specify the directory containing the TNF configuration files.'
	'OUTPUT_LOC is required. Use the -o option to specify the output location for the test results.'
)

TNF_IMAGE_NAME=test-network-function
TNF_IMAGE_TAG=latest
TNF_OFFICIAL_ORG=quay.io/testnetworkfunction/
TNF_OFFICIAL_IMAGE="${TNF_OFFICIAL_ORG}${TNF_IMAGE_NAME}:${TNF_IMAGE_TAG}"

CONTAINER_TNF_DIR=/usr/tnf
CONTAINER_TNF_KUBECONFIG_FILE_BASE_PATH="$CONTAINER_TNF_DIR/kubeconfig/config"

usage() {
	read -d '' usage_prompt <<- EOF
	Usage: $0 -t TNFCONFIG -o OUTPUT_LOC [ -k KUBECONFIG ] [-i IMAGE ] SUITE [ ... SUITE ]

	Configure and run the containerised TNF test offering.

	Options
	  -t: set the directory containing TNF config files set up for the test.
	  -o: set the output location for the test results.
	  -k: set path to one or more local kubeconfigs, separated by a colon.
	      The -k option takes precedence, overwriting the results of local kubeconfig autodiscovery.
	      See the 'Kubeconfig lookup order' section below for more details.
	  -i: set the TNF container image. Supports local images, as well as images from external registries.

	Kubeconfig lookup order
	  1. If -k is specified, use the paths provided with the -k option.
	  2. If -k is not specified, use paths defined in \$KUBECONFIG on the underlying host.
	  3. If no paths are defined, use the default kubeconfig file located in '\$HOME/.kube/config'
	     (currently: $HOME/.kube/config).

	Examples
	  $0 -t ~/tnf/config -o ~/tnf/output diagnostic generic

	  Because -k is omitted, $(basename $0) will first try to autodiscover local kubeconfig files.
	  If it succeeds, the diagnostic and generic tests will be run using the autodiscovered configuration.
	  The test results will be saved to the '~/tnf/output' directory on the host.

	  $0 -k ~/.kube/ABC:~/.kube/DEF -t ~/tnf/config -o ~/tnf/output diagnostic generic

	  The command will bind two kubeconfig files (~/.kube/ABC and ~/.kube/DEF) to the TNF container,
	  run the diagnostic and generic tests, and save the test results into the '~/tnf/output' directory
	  on the host.

	  $0 -i custom-tnf-image:v1.2-dev -t ~/tnf/config -o ~/tnf/output diagnostic generic

	  The command will run the diagnostic and generic tests as implemented in the custom-tnf-image:v1.2-dev
	  local image set by the -i parameter. The test results will be saved to the '~/tnf/output' directory.

	Test suites
	  Allowed tests are listed in the README.
	  Note: Tests must be specified after all other arguments!
	EOF

	echo -e "$usage_prompt"
}

usage_error() {
	usage
	exit 1
}

check_cli_required_num_of_args() {
	if (($# < 5)); then
		usage_error
	fi;
}

check_required_vars() {
	local var_missing=false

	for index in "${!REQUIRED_VARS[@]}"; do
		var=${REQUIRED_VARS[$index]}
		if [[ -z ${!var} ]]; then
			error_message=${REQUIRED_VARS_ERROR_MESSAGES[$index]}
			echo "$0: error: $error_message" 1>&2
			var_missing=true
		fi
	done

	if $var_missing; then
		echo ""
		usage_error
	fi
}

perform_kubeconfig_autodiscovery() {
	if [[ -n "$KUBECONFIG" ]]; then
		LOCAL_KUBECONFIG=$KUBECONFIG
		kubeconfig_autodiscovery_source='$KUBECONFIG'
	elif [[ -f "$HOME/.kube/config" ]]; then
		LOCAL_KUBECONFIG=$HOME/.kube/config
		kubeconfig_autodiscovery_source="\$HOME/.kube/config ($HOME/.kube/config)"
	fi
}

display_kubeconfig_autodiscovery_summary() {
	if [[ -n "$kubeconfig_autodiscovery_source" ]]; then
		echo "Kubeconfig Autodiscovery: configuration loaded from $kubeconfig_autodiscovery_source"
	fi
}

get_container_tnf_kubeconfig_path_from_index() {
	local local_path_index="$1"
	kubeconfig_path=$CONTAINER_TNF_KUBECONFIG_FILE_BASE_PATH

	# To maintain backward compatiblity with the TNF container image,
	# indexing of kubeconfigs starts from the second file.
	# For example:
	# - /usr/tnf/kubeconfig/config
	# - /usr/tnf/kubeconfig/config.2
	# - /usr/tnf/kubeconfig/config.3
	if (($local_path_index > 0)); then
		kubeconfig_index=$(($local_path_index + 1))
		kubeconfig_path="$kubeconfig_path.$kubeconfig_index"
	fi
	echo $kubeconfig_path
}

display_config_summary() {
	printf "Mounting %d kubeconfig volume(s):\n" "${#container_tnf_kubeconfig_volume_bindings[@]}"
	printf -- "-v %s\n" "${container_tnf_kubeconfig_volume_bindings[@]}"

	# Checks whether a prefix of the selected image path matches the address of the official TNF repository
	if [[ "$TNF_IMAGE" != $TNF_OFFICIAL_ORG* ]]; then
		printf "Warning: Could not verify whether '%s' is an official TNF image.\n" "$TNF_IMAGE"
		printf "\t Official TNF images can be pulled directly from '%s'.\n" "$TNF_OFFICIAL_ORG"
	fi
}

join_paths() {
	local IFS=':'; echo "$*"
}

check_cli_required_num_of_args $@
perform_kubeconfig_autodiscovery

# Parge args beginning with -
while [[ $1 == -* ]]; do
	echo $1 $2
    case "$1" in
      -h|--help|-\?) usage; exit 0;;
      -k) if (($# > 1)); then
            LOCAL_KUBECONFIG=$2
            unset kubeconfig_autodiscovery_source
            shift 2
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
      -i) if (($# > 1)); then
            TNF_IMAGE=$2; shift 2
          else
            echo "-i requires an argument" 1>&2
            exit 1
          fi ;;
      --) shift; break;;
      -*) echo "invalid option: $1" 1>&2; usage_error;;
    esac
done

display_kubeconfig_autodiscovery_summary
check_required_vars

# Explode loaded KUBECONFIG (e.g. /kubeconfig/path1:/kubeconfig/path2:...)
# into an array of individual paths to local kubeconfigs.
IFS=":" read -a local_kubeconfig_paths <<< $LOCAL_KUBECONFIG

declare -a container_tnf_kubeconfig_paths
declare -a container_tnf_kubeconfig_volume_bindings

# Assign a file in the TNF container for each provided local kubeconfig
for local_path_index in "${!local_kubeconfig_paths[@]}"; do
	local_path=${local_kubeconfig_paths[$local_path_index]}
	container_path=$(get_container_tnf_kubeconfig_path_from_index $local_path_index)

	container_tnf_kubeconfig_paths+=($container_path)
	container_tnf_kubeconfig_volume_bindings+=("$local_path:$container_path:ro")
done

TNF_IMAGE="${TNF_IMAGE:-$TNF_OFFICIAL_IMAGE}"

display_config_summary

# Construct new $KUBECONFIG env variable containing all paths to kubeconfigs mounted to the container.
# This environment variable is passed to the TNF container and is made available for use by oc/kubectl.
CONTAINER_TNF_KUBECONFIG=$(join_paths ${container_tnf_kubeconfig_paths[@]})

container_tnf_kubeconfig_volumes_cmd_args=$(printf -- "-v %s " "${container_tnf_kubeconfig_volume_bindings[@]}")

set -x
docker run --rm \
	${container_tnf_kubeconfig_volumes_cmd_args[@]} \
	-v $LOCAL_TNF_CONFIG:$CONTAINER_TNF_DIR/config \
	-v $OUTPUT_LOC:$CONTAINER_TNF_DIR/claim \
	-e KUBECONFIG=$CONTAINER_TNF_KUBECONFIG \
	$TNF_IMAGE \
	./run-cnf-suites.sh \
	-o $CONTAINER_TNF_DIR/claim $@
