package prometheus

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/netsec-ethz/2SMS/common/types"
	"github.com/pkg/errors"
	"github.com/prometheus/prometheus/config"
)

// TODO: create interface and specific prometheus config manager
type ConfigManager struct {
	ConfigFile    string
	ProxyURL      string // TODO: change to scrape and read/write
	ListenAddress string
}

// TODO: write
func (cm ConfigManager) AddTarget(target *types.Target) error {
	//parsedConfig, err := config.LoadFile(cm.ConfigFile)
	//if err != nil {
	//	fmt.Println("Error while loading parsedConfig from file:", err)
	//	w.WriteHeader(400)
	//} else {
	//	/*fmt.Println(parsedConfig.ScrapeConfigs[0].JobName)
	//	fmt.Println(parsedConfig.ScrapeConfigs[0].HTTPClientConfig)
	//	fmt.Println(parsedConfig.ScrapeConfigs[0].MetricsPath)
	//	fmt.Println(parsedConfig.ScrapeConfigs[0].Params)
	//	fmt.Println(parsedConfig.ScrapeConfigs[0].Scheme)*/
	//	//fmt.Println(parsedConfig.ScrapeConfigs[0].ServiceDiscoveryConfig.StaticConfigs[0].Targets)
	//
	//	// Parse body
	//	var target common.Target
	//	_ = json.NewDecoder(r.Body).Decode(&target)
	//
	//	// Check if name not already used
	//	if target.ExistsInConfig(parsedConfig) {
	//		w.WriteHeader(400)
	//		return
	//	}
	//
	//	newScrapeConfig := target.ToScrapeConfig()
	//	proxyURL, _ := url.Parse(cm.ProxyURL) // Error is not checked because ProxyURL assumed to be correct
	//	newScrapeConfig.HTTPClientConfig = config2.HTTPClientConfig{ProxyURL: config2.URL{proxyURL}}
	//
	//	// Add new ScrapeConfig to Config.ScrapeConfigs
	//	parsedConfig.ScrapeConfigs = append(parsedConfig.ScrapeConfigs, &newScrapeConfig)
	//
	//	// Write new parsedConfig to file
	//	writeConfig(parsedConfig, cm.ConfigFile)
	//	log.Println("Added job to config:", fmt.Sprint(target.ISD) + "-" + fmt.Sprint(target.AS) + " " + target.Name)
	//
	//	reloadPrometheus()
	//	w.WriteHeader(204)
	//}
	return nil
}

// TODO: write
func (cm ConfigManager) ListTargets() ([]*types.Target, error) {
	//parsedConfig, err := config.LoadFile(cm.ConfigFile)
	//if err != nil {
	//	fmt.Println("Error while loading parsedConfig from file:", err)
	//	w.WriteHeader(400)
	//} else {
	//	var targets []common.Target
	//	for _, job := range parsedConfig.ScrapeConfigs {
	//		var target common.Target
	//		target.FromScrapeConfig(job)
	//		// Extend
	//		targets = append(targets, target)
	//	}
	//	json.NewEncoder(w).Encode(targets)
	//}
	return nil, nil
}

// TODO: write
func (cm ConfigManager) RemoveTarget(target *types.Target) error {
	//parsedConfig, err := config.LoadFile(cm.ConfigFile)
	//if err != nil {
	//	fmt.Println("Error while loading parsedConfig from file:", err)
	//	w.WriteHeader(400)
	//} else {
	//	// Parse body
	//	var target common.Target
	//	_ = json.NewDecoder(r.Body).Decode(&target)
	//
	//	// Check if name exists
	//	if !target.ExistsInConfig(parsedConfig) {
	//		w.WriteHeader(400)
	//		return
	//	}
	//	var newScrapeConfigs []*config.ScrapeConfig
	//	jobName := target.BuildJobName()
	//	for _, job := range parsedConfig.ScrapeConfigs {
	//		if job.JobName != jobName {
	//			newScrapeConfigs = append(newScrapeConfigs, job)
	//		}
	//	}
	//	parsedConfig.ScrapeConfigs = newScrapeConfigs
	//
	//	writeConfig(parsedConfig, cm.ConfigFile)
	//	log.Println("Removed job from config:", fmt.Sprint(target.ISD) + "-" + fmt.Sprint(target.AS) + " " + target.Name)
	//
	//	reloadPrometheus()
	//	w.WriteHeader(204)
	//}
	return nil
}

func (cm ConfigManager) ReloadPrometheus() error {
	resp, err := http.Post("http://"+cm.ListenAddress+"/-/reload", "application/json", nil)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		message, _ := ioutil.ReadAll(resp.Body)
		return errors.New("Failed reloading Prometheus configuration. Status code: " + fmt.Sprint(resp.StatusCode) + ". Message: " + string(message))
	}
	return nil
}

func (cm ConfigManager) WriteConfig(config *config.Config) error {
	f, err := os.Create(cm.ConfigFile)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(config.String())
	if err != nil {
		return err
	}
	err = f.Sync()
	if err != nil {
		return err
	}
	return nil
}
