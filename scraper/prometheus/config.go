package prometheus

import "github.com/netsec-ethz/2SMS/common/types"

type Config struct {
	Global	map[string]string		`yaml:"global"`
	RuleFiles []string				`yaml:"rule_files"`
	Alerting AlertingConfig			`yaml:"alerting"`
	ScrapeConfigs []*ScrapeConfig	`yaml:"scraper_configs"`
	RemoteWrite map[string]string	`yaml:"remote_write"`
	RemoteRead map[string]string	`yaml:"remote_read"`
}

func (config *Config) ContainsTarget(target *types.Target) bool {
	return true
	// TODO
}