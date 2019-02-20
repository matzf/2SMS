package prometheus

import (
	"fmt"
	"github.com/netsec-ethz/2SMS/common/types"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"time"

	"github.com/pkg/errors"
)

type ConfigManager struct {
	configFile    string	// Path to the prometheus configuration file
	scraperProxyURL	*url.URL	// Proxy URL from the Prometheus server to the Scraper component
	promListenURL string	// Base URL where the Prometheus API is exposed
	addChannel chan *types.Target
	removeChannel chan *types.Target
	updateTicker *time.Ticker
	config 		*Config
}

func CreateConfigManager(configFilePath, promListenURL string, scraperProxyURL *url.URL, updateFrequency, updatesBufferSize int) (*ConfigManager, error) {
	configManager := ConfigManager{
		configFile: configFilePath,
		scraperProxyURL: scraperProxyURL,
		promListenURL: promListenURL,
		addChannel: make(chan *types.Target, updatesBufferSize),
		removeChannel: make(chan *types.Target, updatesBufferSize),
		updateTicker: time.NewTicker(time.Duration(updateFrequency) * time.Second),
	}

	err := configManager.loadConfig()
	if err != nil {
		return nil, err
	}

	return &configManager, nil
}

func (cm *ConfigManager) Start() {
	go func() {
		log.Printf("ConfigManager: Started.")
		for range cm.updateTicker.C {
			toAdd := readChannelTargets(cm.addChannel)
			additions := len(toAdd)
			toRemove := readChannelTargets(cm.removeChannel)
			removals := len(toRemove)
			if additions + removals > 0 {
				log.Println("ConfigManager: Updating Prometheus configuration.")
				added, removed := 0, 0
				if additions > 0 {
					added = cm.addTargets(toAdd, cm.config)
					log.Printf("ConfigManager: Added %d/%d targets to the configuration.", added, additions)
				}
				if removals > 0 {
					removed = cm.removeTargets(toRemove, cm.config)
					log.Printf("ConfigManager: Removed %d/%d targets from the configuration.", removed, removals)
				}

				if added + removed > 0 {
					// Write updated configuration to file
					err := cm.writeConfig()
					if err != nil {
						log.Printf("ConfigManager: Failed writing configuration to file. Error is: %v", err)
						continue
					}
					log.Printf("ConfigManager: Successfully written configuration to disk.")

					err = cm.ReloadPrometheus()
					if err != nil {
						log.Printf("ConfigManager: Failed reloading the Prometheus server. Error is: %v", err)
						continue
					}
					log.Printf("ConfigManager: Successfully reloaded the Prometheus server.")
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
	resp, err := http.Post("http://"+cm.promListenURL+"/-/reload", "application/json", nil)
	if err != nil {
		return errors.New(fmt.Sprintf("ConfigManger: Failed executing reloading POST request. Error is: %v", err))
	}
	if resp.StatusCode != 200 {
		message, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		return errors.New(fmt.Sprintf("ConfigManager: Failed reloading Prometheus configuration. Status code is: %d. Message is: %s", resp.StatusCode, string(message)))
	}
	return nil
}

// WriteConfig writes the internal Config structure to the YML file set in this ConfigManager
func (cm *ConfigManager) writeConfig() error {
	f, err := os.Create(cm.configFile)
	if err != nil {
		return err
	}
	defer f.Close()

	yamlConfig, err := yaml.Marshal(cm.config)
	if err != nil {
		return err
	}

	_, err = f.Write(yamlConfig)
	if err != nil {
		return err
	}

	err = f.Sync()
	if err != nil {
		return err
	}
	return nil
}

// LoadFile reads the configuration file, parses it and stores it in the Configuration Manager's state.
func (cm *ConfigManager) loadConfig() (error) {
	// Read configuration file
	yamlFile, err := ioutil.ReadFile(cm.configFile)
	if err != nil {
		return errors.Errorf("Couldn't read configuration file. Error is: %v", err)
	}

	// Unmarshal from yaml
	err = yaml.Unmarshal(yamlFile, &cm.config)
	if err != nil {
		return errors.Errorf("Couldn't unmarshal yaml. Error is: %v", err)
	}

	return nil
}

// Adds all the given targets but duplicates to the configuration and returns the number of added targets.
func (cm *ConfigManager) addTargets(targets []*types.Target, configuration *Config) int {
	added := 0
	for _, target := range targets {
		// Check if name not already used
		targetName := target.BuildJobName()
		if configuration.ContainsTarget(target) {
			log.Printf("ConfigManager: Job with name %s is already present in the configuration.", targetName)
			continue
		}
		newScrapeConfig := targetToScrapeConfig(target)
		newScrapeConfig.ProxyUrl = cm.scraperProxyURL

		// Add new ScrapeConfig to Config.ScrapeConfigs
		configuration.ScrapeConfigs = append(configuration.ScrapeConfigs, newScrapeConfig)
		added++
		log.Printf("ConfigManager: Added job with name %s to the configuration.", newScrapeConfig.JobName)
	}
	return added
}

func (cm *ConfigManager) removeTargets(targets []*types.Target, configuration *Config) int {
	removed := 0
	for _, target := range targets {
		// Check if name exists
		targetName := target.BuildJobName()
		if !configuration.ContainsTarget(target) {
			log.Printf("ConfigManager: Job with name %s not found in the configuration.", targetName)
			continue
		}
		var newScrapeConfigs []*ScrapeConfig
		for _, job := range configuration.ScrapeConfigs {
			if job.JobName != targetName {
				newScrapeConfigs = append(newScrapeConfigs, job)
			}
		}
		configuration.ScrapeConfigs = newScrapeConfigs
		removed++
		log.Printf("ConfigManager: Removed target with name %s from the configuration.", targetName)
	}
	return removed
}

func (cm *ConfigManager) GetTargets() []*types.Target {
	var targets []*types.Target
	for _, sc := range cm.config.ScrapeConfigs {
		targets = append(targets, targetFromScrapeConfig(sc))
	}
	return targets
}

func targetToScrapeConfig(target *types.Target) *ScrapeConfig {
	staticConfig := StaticConfig{
		Targets:[]string{fmt.Sprintf("%s:%s", target.IP, target.Port)},
		Labels: target.Labels,
	}

	return &ScrapeConfig{
		JobName: target.BuildJobName(),
		MetricsPath: fmt.Sprintf("/%s-%s%s", target.ISD, target.AS, target.Path),
		StaticConfigs: []*StaticConfig{&staticConfig},
	}
}

func targetFromScrapeConfig(scrapeConfig *ScrapeConfig) *types.Target {
	target := types.Target{}
	re := regexp.MustCompile(`(.+)-(.+) (.+) (.+)`)
	groups := re.FindStringSubmatch(scrapeConfig.JobName)
	if len(groups) == 5 {
		target.ISD = groups[1]
		target.AS = groups[2]
		target.Name = groups[4]
	} else {
		// cannot guess the ISD or AS
		target.ISD = ""
		target.AS = ""
		target.Name = scrapeConfig.JobName
	}
	// Parse url into IP and Port
	re = regexp.MustCompile(`(.+):(\d+)`)
	groups = re.FindStringSubmatch(string(scrapeConfig.StaticConfigs[0].Targets[0]))
	if len(groups) != 3 {
		log.Printf("Reading Target from prometheus configuration: could not parse address of '%s'", scrapeConfig.JobName)
		target.IP = ""
		target.Port = ""
	} else {
		target.IP = groups[1]
		target.Port = groups[2]
	}
	// Get metrics path
	target.Path = scrapeConfig.MetricsPath
	// Get labels
	target.Labels = scrapeConfig.StaticConfigs[0].Labels
	// If we couldn't get ISD or AS, try populating them from the labels
	if target.ISD == "" {
		target.ISD = target.Labels["ISD"]
	}
	if target.AS == "" {
		target.AS = target.Labels["AS"]
	}
	return &target
}