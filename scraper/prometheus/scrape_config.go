package prometheus

import "net/url"

type ScrapeConfig struct {
	JobName		string
	ScrapeInterval	string
	ScrapeTimeout	string
	MetricsPath		string
	StaticConfigs	[]*StaticConfig
	ProxyUrl		*url.URL
}