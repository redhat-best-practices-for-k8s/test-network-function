package autodiscover

import (
	_ "embed"
	"strings"
)

// PodSetList holds the data from an `oc get deployment/statefulset -o json` command

// PodSetResource defines deployment/statefulset resources
var (
	csiCommand = "oc get subscriptions.operators.coreos.com -A -ogo-template='{{range .items}}{{.spec.source}}_{{.status.currentCSV}},{{end}}'"
	//go:embed csi-mapping.json
	csiMappingString []byte
)

func GetPackageandOrg(csi string) (string, error) {
	csiNameToOperatorName := make(map[string]string)
	err := jsonUnmarshal(csiMappingString, &csiNameToOperatorName)
	if err != nil {
		return "", err
	}
	return csiNameToOperatorName[csi], nil
}

// GetTargetCsi will return the csidriver list.
func GetTargetCsi() ([]string, error) {
	ocCmd := csiCommand

	out := execCommandOutput(ocCmd)

	csiList := strings.Split(out, ",")
	return csiList, nil
}
