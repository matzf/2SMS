package prometheus

type AlertmanagerConfig struct {
	PathPrefix		string `yaml:"path_prefix,omitempty"`
	StaticConfigs	[]*StaticConfig `yaml:"static_configs,omitempty"`
}