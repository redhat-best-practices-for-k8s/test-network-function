package autodiscover

type HelmSetList struct {
	Items []HelmList `json:""`
}
type HelmList struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Chart     string `json:"chart"`
}

func Gethelm() *HelmSetList {
	out := execCommandOutput("helm list -A -o json")
	var helmList HelmSetList
	if out != "" {
		err := jsonUnmarshal([]byte(out), &helmList.Items)
		if err != nil {
			return nil
		}
	}
	return &helmList
}
