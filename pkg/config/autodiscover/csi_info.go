package autodiscover

import (
	"fmt"
	"strings"
)

// PodSetList holds the data from an `oc get deployment/statefulset -o json` command

// PodSetResource defines deployment/statefulset resources
var (
	csiCommand          = "oc get csidriver -o go-template='{{ range .items}}{{.metadata.name}} {{end}}'"
	depNameCommand      = "oc get pods -A -o go-template='{{ range .items}}{{ $alllabels := .metadata.labels}}{{ $namespace := .metadata.namespace}}{{ range .spec.containers }}{{ range .args }}{{if eq . \"--driver-name=%s\"}}{{ range $label,$value := $alllabels}}{{if eq $label \"app.kubernetes.io/managed-by\"}}{{$value}} {{$namespace}}{{end}}{{end}}{{end}}{{end}}{{end}}{{end}}'"
	operatorNameCommand = "oc get deployment %s -n %s -o go-template='{{ range $label,$value := .metadata.labels}}{{$label}}{{print \"\n\"}}{{end}}' |grep \"operators.coreos.com\"| sed \"s#operators.coreos.com/##g\""
	subscriptionCommand = "oc get operator %s -o go-template='{{ range .status.components.refs }}{{if eq .kind \"Subscription\"}}{{.name}}{{end}}{{end}}'"
	orgPackCommand      = "oc get subscription -n %s %s -o go-template='{{.spec.source}} {{.spec.name}}'"
)

func GetPackageandOrg(csi string) (packag, organization string) {
	command := fmt.Sprintf(depNameCommand, csi)
	out := execCommandOutput(command)
	operatorName := ""
	nameSpace := ""
	subscription := ""
	if out != "" {
		out := strings.Split(out, " ")
		operatorName = out[0]
		nameSpace = out[1]
	}
	command = fmt.Sprintf(operatorNameCommand, operatorName, nameSpace)
	out = execCommandOutput(command)
	if out != "" {
		operatorName = out
	}
	command = fmt.Sprintf(subscriptionCommand, operatorName)
	out = execCommandOutput(command)
	if out != "" {
		subscription = out
	}
	command = fmt.Sprintf(orgPackCommand, nameSpace, subscription)
	out = execCommandOutput(command)
	if out != "" {
		out := strings.Split(out, " ")
		organization = out[0]
		packag = out[1]
	}
	return packag, organization
}

// GetTargetCsi will return the csidriver list.
func GetTargetCsi() ([]string, error) {
	ocCmd := csiCommand

	out := execCommandOutput(ocCmd)

	csiList := strings.Split(out, " ")
	return csiList, nil
}
