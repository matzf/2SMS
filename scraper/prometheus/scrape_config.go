package prometheus

type ScrapeConfig struct {
	JobName		string		`yaml:"job_name"`
	ScrapeInterval	string `yaml:"scrape_interval,omitempty"`
	ScrapeTimeout	string `yaml:"scrape_timeout,omitempty"`
	MetricsPath		string `yaml:"metrics_path,omitempty"`
	StaticConfigs	[]*StaticConfig `yaml:"static_configs"`
	ProxyUrl		string `yaml:"proxy_url,omitempty"`
}