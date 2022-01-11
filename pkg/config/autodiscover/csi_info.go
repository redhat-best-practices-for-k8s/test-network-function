package autodiscover

import (
	"fmt"
	"strings"
)

// PodSetList holds the data from an `oc get deployment/statefulset -o json` command

// PodSetResource defines deployment/statefulset resources

func getpackageandorg(csi string) (packag, organization string) {
	orgpack := fmt.Sprintf("./get-csi-info.sh %s", csi)
	out := execCommandOutput(orgpack)
	packag = ""
	organization = ""
	if out != "" {
		out := strings.Split(out, " ")
		packag = out[0]
		organization = out[1]
	}
	return packag, organization

}

// GetTargetCsi will return all podsets(deployments/statefulset )that have pods with a given label.
func GetTargetCsi() ([]string, error) {
	ocCmd := "oc get csidriver -o go-template='{{ range .items}}{{.metadata.name}} {{end}}'"

	out := execCommandOutput(ocCmd)

	csiList := strings.Split(out, " ")
	ocCmd = "pwd"

	out = execCommandOutput(ocCmd)
	return csiList, nil
}
