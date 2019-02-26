package prometheus

import "github.com/netsec-ethz/2SMS/common/types"

type Config struct {
	Global	map[string]string		`yaml:"global"`
	RuleFiles []string				`yaml:"rule_files"`
	Alerting AlertingConfig			`yaml:"alerting"`
	ScrapeConfigs []*ScrapeConfig	`yaml:"scrape_configs"`
	RemoteWrites []*RemoteWriteConfig	`yaml:"remote_write"`
	RemoteReads []*RemoteReadConfig	`yaml:"remote_read"`
}

func (config *Config) ContainsTarget(target *types.Target) bool {
	jobName := target.BuildJobName()
	for _, job := range config.ScrapeConfigs {
		if job.JobName == jobName {
			return true
		}
	}
	return false
}