package autodiscover

import (
	"github.com/test-network-function/test-network-function/pkg/tnf/interactive"
	"github.com/test-network-function/test-network-function/pkg/utils"
)

type HelmSetList struct {
	Items []HelmChart `json:""`
}
type HelmChart struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Chart     string `json:"chart"`
}

func GetClusterHelmCharts() (*HelmSetList, error) {
	var helmList HelmSetList

	out, err := utils.ExecuteCommand("helm list -A -o json", ocCommandTimeOut, interactive.GetContext(expectersVerboseModeEnabled))
	if err != nil {
		return &helmList, err
	}

	if out != "" {
		err := jsonUnmarshal([]byte(out), &helmList.Items)
		if err != nil {
			return nil, err
		}
	}
	return &helmList, nil
}
