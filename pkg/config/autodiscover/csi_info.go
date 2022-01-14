package autodiscover

import (
	"strings"
	_ "embed"
)

// PodSetList holds the data from an `oc get deployment/statefulset -o json` command

// PodSetResource defines deployment/statefulset resources
var (
	csiCommand          = "oc get csidriver -o go-template='{{ range .items}}{{.metadata.name}} {{end}}'"
	//go:embed csi-mapping.json
   csiMappingString []byte
)

func GetPackageandOrg(csi string) (string,  string, error){
	csiNameToOperatorName:=make(map[string]string)
	err:=jsonUnmarshal(csiMappingString,&csiNameToOperatorName)
	if err!=nil{
		return "","",err
	}
	return csiNameToOperatorName[csi], "certified-operators", nil
}

// GetTargetCsi will return the csidriver list.
func GetTargetCsi() ([]string, error) {
	ocCmd := csiCommand

	out := execCommandOutput(ocCmd)

	csiList := strings.Split(out, " ")
	return csiList, nil
}
