package types

type Endpoint struct {
	IA         string   `json:"ia"`
	IP         string   `json:"ip"`
	ScrapePort string   `json:"scrape_port"`
	ManagePort string   `json:"manage_port"`
	Paths      []string `json:"paths"`
}

func (end *Endpoint) Equal(end_b *Endpoint) bool {
	return end.IA == end_b.IA && end.IP == end_b.IP && end.ManagePort == end_b.ManagePort
}
