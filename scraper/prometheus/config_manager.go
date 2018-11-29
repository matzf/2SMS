package prometheus

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/netsec-ethz/2SMS/common/types"
	"github.com/pkg/errors"
	"github.com/prometheus/prometheus/config"
)

// TODO: create interface and specific prometheus config manager
type ConfigManager struct {
	configFile                string
	prometheusBin             string
	_configRelativeFromBinary string // relative path to configFile from prometheusBin directory
	ProxyURL                  string // TODO: change to scrape and read/write
	ListenAddress             string
}

// NewConfigManager creates a new ConfigManager and returns a pointer to it
func NewConfigManager(configFile, prometheusBin, proxyURL, listenAddress string) *ConfigManager {
	cm := ConfigManager{configFile: configFile, prometheusBin: prometheusBin, ProxyURL: proxyURL, ListenAddress: listenAddress}
	cm._recomputeConfPathFromPromBinPath()
	return &cm
}

// this will recompute the relative path to the configuration file from the prometheus binary path
func (cm *ConfigManager) _recomputeConfPathFromPromBinPath() {
	configDirs := filepath.SplitList(filepath.Dir(filepath.Clean(cm.configFile)))
	binDirs := filepath.SplitList(filepath.Dir(filepath.Clean(cm.prometheusBin)))
	commonPrefixCount := 0
	for p := range configDirs {
		if configDirs[p] != binDirs[p] {
			break
		}
		commonPrefixCount++
	}
	configDirs = configDirs[commonPrefixCount:]
	cm._configRelativeFromBinary = filepath.Join(configDirs...)
	cm._configRelativeFromBinary = filepath.Join(cm._configRelativeFromBinary, filepath.Base(cm.configFile))
}

// GetConfigPath returns the configuration path
func (cm *ConfigManager) GetConfigPath() string {
	return cm.configFile
}

// SetConfigPath sets the configuration path
func (cm *ConfigManager) SetConfigPath(configPath string) {
	cm.configFile = configPath
	cm._recomputeConfPathFromPromBinPath()
}

// GetPrometheusBinPath returns the prometheus binary path
func (cm *ConfigManager) GetPrometheusBinPath() string {
	return cm.prometheusBin
}

// SetPrometheusBinPath sets the prometheus binary path
func (cm *ConfigManager) SetPrometheusBinPath(prometheusBinPath string) {
	cm.prometheusBin = prometheusBinPath
	cm._recomputeConfPathFromPromBinPath()
}

// TODO: write
func (cm *ConfigManager) AddTarget(target *types.Target) error {
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
func (cm *ConfigManager) ListTargets() ([]*types.Target, error) {
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
func (cm *ConfigManager) RemoveTarget(target *types.Target) error {
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

func (cm *ConfigManager) ReloadPrometheus() error {
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

// WriteConfig writes the prometheus native Config structure to the YML file set in this ConfigManager
func (cm *ConfigManager) WriteConfig(config *config.Config) error {
	f, err := os.Create(cm.configFile)
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

// LoadFile reads the configuration file and returns a prometheus configuration Config struct.
// The configuration file can be anywhere, and LoadFile can be called from any working directory.
func (cm *ConfigManager) LoadFile() (*config.Config, error) {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Cannot obtain the CWD. Fatal error is: %v", err)
	}
	defer func(d string) {
		err := os.Chdir(d)
		if err != nil {
			log.Fatalf("Cannot chdir back from the directory where prometheus lives (%s). Fatal error is: %v", d, err)
		}
	}(cwd)
	err = os.Chdir(filepath.Dir(cm.prometheusBin))
	if err != nil {
		log.Fatalf("Cannot chdir to the directory where prometheus lives (%s). Fatal error is: %v", filepath.Dir(cm.prometheusBin), err)
	}
	return config.LoadFile(cm._configRelativeFromBinary)
}
