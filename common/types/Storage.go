package types

import (
	"github.com/prometheus/prometheus/config"
	config2 "github.com/prometheus/common/config"
	"net/url"
	"regexp"
)

type Storage struct {
	IA         string `json:"ia"`
	IP         string `json:"ip"`
	Port       string `json:"port"`
	ManagePort string `json:"manage_port"`
}

func (str *Storage) Equal(str_b *Storage) bool {
	return str.IA == str_b.IA && str.IP == str_b.IP && str.ManagePort == str_b.ManagePort && str.Port == str_b.Port
}

func (str *Storage) BuildWriteURL() string {
	return "http://" + str.IP +":" + str.Port + "/" + str.IA + "/write"
}

func (str *Storage) BuildReadURL() string {
	return "http://" + str.IP +":" + str.Port + "/" + str.IA + "/read"
}

func (str *Storage) ExistsInConfig(currentConfig *config.Config) bool {
	read, write := false, false
	for _, rWrite := range currentConfig.RemoteWriteConfigs {
		if rWrite.URL.String() == str.BuildWriteURL() {
			write = true
		}
	}
	for _, rRead := range currentConfig.RemoteReadConfigs {
		if rRead.URL.String() == str.BuildReadURL() {
			read = true
		}
	}
	return read && write
}

func (str *Storage) ToRemoteConfigs(proxyUrl string) (*config.RemoteWriteConfig, *config.RemoteReadConfig) {
	proxy, err := url.Parse(proxyUrl)
	if err != nil {
		// TODO
	}

	rWrite := config.RemoteWriteConfig{}
	url, err := url.Parse(str.BuildWriteURL())
	if err != nil {
		// TODO
	}
	rWrite.URL = &config2.URL{url}
	rWrite.HTTPClientConfig.ProxyURL = config2.URL{proxy}

	rRead := config.RemoteReadConfig{}
	url, err = url.Parse(str.BuildReadURL())
	if err != nil {
		// TODO
	}
	rRead.URL = &config2.URL{url}
	rRead.HTTPClientConfig.ProxyURL = config2.URL{proxy}

	return &rWrite, &rRead
}

// Assumption: for every remote write config there is a corresponding read config
func (str *Storage) FromRemoteConfig(rw *config.RemoteWriteConfig) {
	re := regexp.MustCompile( `http://(.+):(.+)/(.+)/(.+)`) // IP:Port/IA/write
	// Parse write url into its subcomponents
	groups := re.FindStringSubmatch(rw.URL.String())
	str.IP = groups[1]
	str.Port = groups[2]
	str.IA = groups[3]
}