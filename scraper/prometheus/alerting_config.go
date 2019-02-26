package prometheus

type AlertingConfig struct {
	Alertmanagers	[]*AlertmanagerConfig `yaml:",omitempty"`
}