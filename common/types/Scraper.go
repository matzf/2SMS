package types

type Scraper struct {
	IA string   `json:"ia"`
	IP string 	`json:"ip"`
	ManagePort string `json:"manage_port"`
	ISDs []string `json:"isds"`
}

func (scr *Scraper) Equal(scr_b *Scraper) bool {
	return scr.IA == scr_b.IA && scr.IP == scr_b.IP && scr.ManagePort == scr_b.ManagePort
}

func (scr *Scraper) Covers(isd string) bool {
	for _, i := range(scr.ISDs) {
		if i == isd {
			return true
		}
	}
	return false
}