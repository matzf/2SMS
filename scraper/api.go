package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/netsec-ethz/2SMS/common/types"
	config2 "github.com/prometheus/common/config"
	"github.com/prometheus/prometheus/config"
)

// TODO: call config manager instead of doing everything here (just parse request and build answer)

func AddTarget(w http.ResponseWriter, r *http.Request) {
	parsedConfig, err := config.LoadFile(configManager.ConfigFile)
	if err != nil {
		log.Println("Error while loading parsedConfig from file:", err)
		w.WriteHeader(500)
		return
	}
	// Parse body
	var target types.Target
	_ = json.NewDecoder(r.Body).Decode(&target)

	// Check if name not already used
	if target.ExistsInConfig(parsedConfig) {
		w.WriteHeader(400)
		w.Write([]byte("Job name already in use."))
		return
	}

	newScrapeConfig := target.ToScrapeConfig()
	proxyURL, _ := url.Parse(configManager.ProxyURL) // Error is not checked because ProxyURL assumed to be correct
	newScrapeConfig.HTTPClientConfig = config2.HTTPClientConfig{ProxyURL: config2.URL{proxyURL}}

	// Add new ScrapeConfig to Config.ScrapeConfigs
	parsedConfig.ScrapeConfigs = append(parsedConfig.ScrapeConfigs, &newScrapeConfig)

	// Write new parsedConfig to file
	configManager.WriteConfig(parsedConfig, configManager.ConfigFile)
	log.Println("Added job to config:", fmt.Sprint(target.ISD)+"-"+fmt.Sprint(target.AS)+" "+target.Name)

	err = configManager.ReloadPrometheus()
	if err != nil {
		log.Println("Failed reloading Prometheus:", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(201)

}

func ListTargets(w http.ResponseWriter, r *http.Request) {
	parsedConfig, err := config.LoadFile(configManager.ConfigFile)
	if err != nil {
		fmt.Println("Error while loading parsedConfig from file:", err)
		w.WriteHeader(500)
	} else {
		targets := []types.Target{}
		for _, job := range parsedConfig.ScrapeConfigs {
			var target types.Target
			target.FromScrapeConfig(job)
			// Extend
			targets = append(targets, target)
		}
		json.NewEncoder(w).Encode(targets)
	}
}

func RemoveTarget(w http.ResponseWriter, r *http.Request) {
	parsedConfig, err := config.LoadFile(configManager.ConfigFile)
	if err != nil {
		fmt.Println("Error while loading parsedConfig from file:", err)
		w.WriteHeader(500)
	} else {
		// Parse body
		var target types.Target
		_ = json.NewDecoder(r.Body).Decode(&target)

		// Check if name exists
		if !target.ExistsInConfig(parsedConfig) {
			w.WriteHeader(400)
			w.Write([]byte("Job name not found."))
			return
		}
		var newScrapeConfigs []*config.ScrapeConfig
		jobName := target.BuildJobName()
		for _, job := range parsedConfig.ScrapeConfigs {
			if job.JobName != jobName {
				newScrapeConfigs = append(newScrapeConfigs, job)
			}
		}
		parsedConfig.ScrapeConfigs = newScrapeConfigs

		configManager.WriteConfig(parsedConfig, configManager.ConfigFile)
		log.Println("Removed job from config:", fmt.Sprint(target.ISD)+"-"+fmt.Sprint(target.AS)+" "+target.Name)

		err := configManager.ReloadPrometheus()
		if err != nil {
			log.Println("Failed reloading Prometheus:", err)
			w.WriteHeader(500)
			return
		}
		w.WriteHeader(204)
	}
}

func RemoveStorage(w http.ResponseWriter, r *http.Request) {
	parsedConfig, err := config.LoadFile(configManager.ConfigFile)
	if err != nil {
		fmt.Println("Error while loading parsedConfig from file:", err)
		w.WriteHeader(400)
	} else {
		// Parse body
		var storage types.Storage
		_ = json.NewDecoder(r.Body).Decode(&storage)

		// Check if name exists
		if !storage.ExistsInConfig(parsedConfig) {
			w.WriteHeader(400)
			w.Write([]byte("Remote read/write not found."))
			return
		}

		// Remove remote read/write
		var newWriteConfigs []*config.RemoteWriteConfig
		for _, writeConfig := range parsedConfig.RemoteWriteConfigs {
			if writeConfig.URL.String() != storage.BuildWriteURL() {
				newWriteConfigs = append(newWriteConfigs, writeConfig)
			}
		}
		parsedConfig.RemoteWriteConfigs = newWriteConfigs

		var newReadConfigs []*config.RemoteReadConfig
		for _, readConfig := range parsedConfig.RemoteReadConfigs {
			if readConfig.URL.String() != storage.BuildReadURL() {
				newReadConfigs = append(newReadConfigs, readConfig)
			}
		}
		parsedConfig.RemoteReadConfigs = newReadConfigs

		configManager.WriteConfig(parsedConfig, configManager.ConfigFile)
		configManager.ReloadPrometheus()
		log.Println("Removed remote read/write from config:", fmt.Sprint(storage.IA)+" "+storage.IP)

		err := configManager.ReloadPrometheus()
		if err != nil {
			log.Println("Failed reloading Prometheus:", err)
			w.WriteHeader(500)
			return
		}
		w.WriteHeader(204)
	}
}

func AddStorage(w http.ResponseWriter, r *http.Request) {
	// Parse Request
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("Error while reading request body:", err)
		w.WriteHeader(500)
		return
	}
	var storage types.Storage
	err = json.Unmarshal(data, &storage)
	if err != nil {
		log.Println("Error while reading request body:", err)
		w.WriteHeader(500)
		return
	}

	parsedConfig, err := config.LoadFile(configManager.ConfigFile)
	if err != nil {
		log.Println("Error while parsing config file:", err)
		w.WriteHeader(500)
		return
	}
	if !storage.ExistsInConfig(parsedConfig) {
		newRemoteWriteConfig, newRemoteReadConfig := storage.ToRemoteConfigs(configManager.ProxyURL)
		parsedConfig.RemoteWriteConfigs = append(parsedConfig.RemoteWriteConfigs, newRemoteWriteConfig)
		parsedConfig.RemoteReadConfigs = append(parsedConfig.RemoteReadConfigs, newRemoteReadConfig)

		configManager.WriteConfig(parsedConfig, configManager.ConfigFile)
		configManager.ReloadPrometheus()
		log.Println("Added remote read/write to config:", fmt.Sprint(storage.IA)+" "+storage.IP)
		w.WriteHeader(201)
		return
	}
	w.Write([]byte("Storage already present in the configuration."))
	w.WriteHeader(400)
}

func ListStorages(w http.ResponseWriter, r *http.Request) {
	parsedConfig, err := config.LoadFile(configManager.ConfigFile)
	if err != nil {
		fmt.Println("Error while loading config from file:", err)
		w.WriteHeader(500)
		return
	}

	storages := []*types.Storage{}
	for _, rWrite := range parsedConfig.RemoteWriteConfigs {
		var s types.Storage
		s.FromRemoteConfig(rWrite)
		// Extend
		storages = append(storages, &s)
	}
	json.NewEncoder(w).Encode(storages)
}
