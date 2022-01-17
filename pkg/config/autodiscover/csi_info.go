package autodiscover

import (
	_ "embed"
	"strings"
)

// PodSetList holds the data from an `oc get deployment/statefulset -o json` command

// PodSetResource defines deployment/statefulset resources
var (
	csiCommand          = "oc get csidriver -o go-template='{{ range .items}}{{.metadata.name}} {{end}}'"
	//go:embed csi-mapping.json
   csiMappingString []byte
)

func GetPackageandOrg(csi string) (string, error){
	csiNameToOperatorName:=make(map[string]string)
	err:=jsonUnmarshal(csiMappingString,&csiNameToOperatorName)
	if err!=nil{
		return "",err
	}
	return csiNameToOperatorName[csi], nil
}

// GetTargetCsi will return the csidriver list.
func GetTargetCsi() ([]string, error) {
	ocCmd := csiCommand

	out := execCommandOutput(ocCmd)

	csiList := strings.Split(out, " ")
	return csiList, nil
}
/*
func TestOperatorVersion() {
	csilist, err := GetTargetCsi()
	if err != nil {
		log.Error("Unable to get csi list  Error: ", err)
		return
	}	
	mapOperatorVersions:=csimapping.GetOperatorVersions()
	ocpVersion:=GetOcpVersion()
	operatorVersionMap:=GetOperatorVersionMap()

	for _, csi := range csilist {
		if csi != "" {
			pack, _,_:= GetPackageandOrg(csi)
			if pack!="" {
				aKey := csimapping.OperatorKey{OperatorName: pack, OcpVersion: ocpVersion}
				for _,version:=range mapOperatorVersions[aKey]{
					if operatorVersionMap[pack]==version{
						log.Infof("Operator: %s currently running version: %s this version is certified to run with Current OCP version %s",pack,version, ocpVersion)
					}else{
						log.Infof("Operator: %s currently running version: %s this version is NOT certified to run with OCP version %s",pack,version, ocpVersion)
					}
				}
			} else {
				log.Infof("Driver: %s is not provided by a certified operator or csimapping.json needs to be updated",csi)
			}
		}
	}
}*/