package configsections

type Hpa struct {
	MinReplicas int
	MaxReplicas int
	HpaName     string
}

func (hpa Hpa) GetHpaName() string {
	return hpa.HpaName
}
