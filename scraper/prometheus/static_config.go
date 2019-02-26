package prometheus

type StaticConfig struct {
	Targets	[]string
	Labels	map[string]string `yaml:",omitempty"`
}
