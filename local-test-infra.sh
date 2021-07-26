mkdir -p ./temp

export res=$(oc get pod -A -l test-network-function.com/generic=fs_diff_master 2>&1)
if [ "$res" == "No resources found" ]; then
    cat $TNF_PARTNER_SRC_DIR/local-test-infra/fsdiff-pod.yaml | ./script/mo > ./temp/rendered-fsdiff-template.yaml
    oc apply -f ./temp/rendered-fsdiff-template.yaml
    rm ./temp/rendered-fsdiff-template.yaml
else
    echo "fsDiffMasterPod already exists, no reason to recreate"
fi

export res=$(oc get pod -A -l test-network-function.com/generic=orchestrator 2>&1)
if [ "$res" == "No resources found" ]; then
    cat $TNF_PARTNER_SRC_DIR/local-test-infra/local-partner-pod.yaml | ./script/mo > ./temp/rendered-partner-template.yaml
    oc apply -f ./temp/rendered-partner-template.yaml
    rm ./temp/rendered-partner-template.yaml
else
    echo "partner pod already exists, no reason to recreate"
fi