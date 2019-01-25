package prometheus

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/prometheus/prometheus/config"
)

type ConfigManager struct {
	ConfigFile    string
	PathPrefix	  string
	ProxyURL      string // TODO: change to scrape and read/write
	ListenAddress string
}

func (cm *ConfigManager) ReloadPrometheus() error {
	resp, err := http.Post(fmt.Sprintf("http://%s%s/-/reload", cm.ListenAddress, cm.PathPrefix), "application/json", nil)
	if err != nil {
		return errors.New(fmt.Sprintf("Error while executing reloading POST request. Error is: %v", err))
	}
	if resp.StatusCode != 200 {
		message, _ := ioutil.ReadAll(resp.Body)
		return errors.New("Failed reloading Prometheus configuration. Status code: " + fmt.Sprint(resp.StatusCode) + ". Message: " + string(message))
	}
	return nil
}

// WriteConfig writes the prometheus native Config structure to the YML file set in this ConfigManager
func (cm *ConfigManager) WriteConfig(config *config.Config) error {
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
	err = os.Chdir(filepath.Dir(cm.ConfigFile))
	if err != nil {
		log.Fatalf("Cannot chdir to the directory where prometheus lives (%s). Fatal error is: %v", filepath.Dir(cm.ConfigFile), err)
	}
	// we need to specify a path without subdirectories, for config.LoadFile will prepend those
	// to the filepaths contained in the config file
	return config.LoadFile(filepath.Base(cm.ConfigFile))
}
