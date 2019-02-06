package prometheus

import (
	"fmt"
	"github.com/netsec-ethz/2SMS/common/types"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	config2 "github.com/prometheus/common/config"
	"github.com/prometheus/prometheus/config"
)

type ConfigManager struct {
	configFile     string // Path to the prometheus configuration file
	scrapeProxyURL string // Proxy URL from the Prometheus server to the Scraper component
	promListenURL  string // Base URL where the Prometheus API is exposed
	addChannel     chan *types.Target
	removeChannel  chan *types.Target
	updateTicker   *time.Ticker
}

func CreateConfigManager(configFilePath, scrapeProxyURL, promListenURL string, updateFrequency, updatesBufferSize int) *ConfigManager {
	return &ConfigManager{
		configFile:     configFilePath,
		scrapeProxyURL: scrapeProxyURL,
		promListenURL:  promListenURL,
		addChannel:     make(chan *types.Target, updatesBufferSize),
		removeChannel:  make(chan *types.Target, updatesBufferSize),
		updateTicker:   time.NewTicker(time.Duration(updateFrequency) * time.Second),
	}
}

func (cm *ConfigManager) Start() {
	go func() {
		for range cm.updateTicker.C {
			toAdd := readChannelTargets(cm.addChannel)
			additions := len(toAdd)
			toRemove := readChannelTargets(cm.removeChannel)
			removals := len(toRemove)

			if additions+removals > 0 {
				log.Println("Updating Prometheus configuration file")

				parsedConfig, err := cm.LoadFile()
				if err != nil {
					log.Printf("Error while loading Prometheus configuration from file. Error is: %v", err)
					continue
				}

				added, removed := 0, 0
				if additions > 0 {
					added = cm.addTargets(toAdd, parsedConfig)
					log.Printf("Added %d/%d targets to the configuration.", added, additions)
				}
				if removals > 0 {
					removed = cm.removeTargets(toRemove, parsedConfig)
					log.Printf("Removed %d/%d targets from the configuration.", removed, removals)
				}

				if added+removed > 0 {
					// Write updated configuration to file
					cm.WriteConfig(parsedConfig)
					log.Printf("Updated configuration written to disk.")

					err = cm.ReloadPrometheus()
					if err != nil {
						log.Printf("Failed reloading the Prometheus server. Error is: %v", err)
						continue
					}
					log.Printf("Successfully reloaded the Prometheus server.")
				}
			}
		}
	}()
}

func readChannelTargets(channel <-chan *types.Target) []*types.Target {
	var targets []*types.Target
	// Non-blocking reading of all values in channel
	for {
		select {
		case target := <-channel:
			targets = append(targets, target)
		default:
			return targets
		}
	}
}

func (cm *ConfigManager) AddTarget(target types.Target) {
	cm.addChannel <- &target
}

func (cm *ConfigManager) AddTargets(targets []types.Target) {
	for _, target := range targets {
		cm.addChannel <- &target
	}
}

func (cm *ConfigManager) RemoveTarget(target types.Target) {
	cm.removeChannel <- &target
}

func (cm *ConfigManager) RemoveTargets(targets []types.Target) {
	for _, target := range targets {
		cm.removeChannel <- &target
	}
}

func (cm *ConfigManager) ReloadPrometheus() error {
	resp, err := http.Post(cm.promListenURL+"/-/reload", "application/json", nil)
	if err != nil {
		return errors.New(fmt.Sprintf("Error while executing reloading POST request. Error is: %v", err))
	}
	if resp.StatusCode != 200 {
		message, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		return errors.New(fmt.Sprintf("Failed reloading Prometheus configuration. Status code is: %d. Message is: %s", resp.StatusCode, string(message)))
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
	err = os.Chdir(filepath.Dir(cm.configFile))
	if err != nil {
		log.Fatalf("Cannot chdir to the directory where prometheus lives (%s). Fatal error is: %v", filepath.Dir(cm.configFile), err)
	}
	// we need to specify a path without subdirectories, for config.LoadFile will prepend those
	// to the filepaths contained in the config file
	return config.LoadFile(filepath.Base(cm.configFile))
}

// Adds all the given targets but duplicates to the configuration and returns the number of added targets.
func (cm *ConfigManager) addTargets(targets []*types.Target, configuration *config.Config) int {
	added := 0
	for _, target := range targets {
		// Check if name not already used
		if target.ExistsInConfig(configuration) {
			log.Printf("AddTargets: target with name %s is already present in the configuration.", target.BuildJobName())
			continue
		}
		newScrapeConfig := target.ToScrapeConfig()
		proxyURL, _ := url.Parse(cm.scrapeProxyURL) // Error is not checked because scrapeProxyURL assumed to be correct
		newScrapeConfig.HTTPClientConfig = config2.HTTPClientConfig{ProxyURL: config2.URL{proxyURL}}

		// Add new ScrapeConfig to Config.ScrapeConfigs
		configuration.ScrapeConfigs = append(configuration.ScrapeConfigs, &newScrapeConfig)
		added++
		log.Printf("AddTargets: Added target with name %s to the configuration.", target.BuildJobName())
	}
	return added
}

func (cm *ConfigManager) removeTargets(targets []*types.Target, configuration *config.Config) int {
	removed := 0
	for _, target := range targets {
		// Check if name exists
		if !target.ExistsInConfig(configuration) {
			log.Printf("RemoveTargets: Target with name %s not found in the configuration.", target.BuildJobName())
			continue
		}
		var newScrapeConfigs []*config.ScrapeConfig
		jobName := target.BuildJobName()
		for _, job := range configuration.ScrapeConfigs {
			if job.JobName != jobName {
				newScrapeConfigs = append(newScrapeConfigs, job)
			}
		}
		configuration.ScrapeConfigs = newScrapeConfigs
		removed++
		log.Printf("RemoveTargets: Removed target with name %s from the configuration.", target.BuildJobName())
	}
	return removed
}
