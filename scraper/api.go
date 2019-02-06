package main

import (
	"encoding/json"
	"fmt"
	"github.com/netsec-ethz/2SMS/common/types"
	"github.com/prometheus/prometheus/config"
	"log"
	"net/http"
)

func AddTarget(w http.ResponseWriter, r *http.Request) {
	// Parse body
	var target types.Target
	err := json.NewDecoder(r.Body).Decode(&target)

	if err != nil {
		log.Printf("Failed parsing request's body. Error is: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	configManager.AddTarget(target)
	w.WriteHeader(http.StatusCreated)
}

func ListTargets(w http.ResponseWriter, r *http.Request) {
	parsedConfig, err := configManager.LoadFile()
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
	// Parse body
	var target types.Target
	err := json.NewDecoder(r.Body).Decode(&target)
	if err != nil {
		log.Printf("Failed parsing request's body. Error is: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	configManager.RemoveTarget(target)
	w.WriteHeader(http.StatusNoContent)
}

func RemoveStorage(w http.ResponseWriter, r *http.Request) {
	parsedConfig, err := configManager.LoadFile()
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

		configManager.WriteConfig(parsedConfig)
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

// TODO: reimplement in config manager and call from there
func AddStorage(w http.ResponseWriter, r *http.Request) {
	// Parse Request
	//data, err := ioutil.ReadAll(r.Body)
	//if err != nil {
	//	log.Println("Error while reading request body:", err)
	//	w.WriteHeader(500)
	//	return
	//}
	//var storage types.Storage
	//err = json.Unmarshal(data, &storage)
	//if err != nil {
	//	log.Println("Error while reading request body:", err)
	//	w.WriteHeader(500)
	//	return
	//}
	//
	//parsedConfig, err := configManager.LoadFile()
	//if err != nil {
	//	log.Println("Error while parsing config file:", err)
	//	w.WriteHeader(500)
	//	return
	//}
	//if !storage.ExistsInConfig(parsedConfig) {
	//	newRemoteWriteConfig, newRemoteReadConfig := storage.ToRemoteConfigs(configManager.scrapeProxyURL)
	//	parsedConfig.RemoteWriteConfigs = append(parsedConfig.RemoteWriteConfigs, newRemoteWriteConfig)
	//	parsedConfig.RemoteReadConfigs = append(parsedConfig.RemoteReadConfigs, newRemoteReadConfig)
	//
	//	configManager.WriteConfig(parsedConfig)
	//	configManager.ReloadPrometheus()
	//	log.Println("Added remote read/write to config:", fmt.Sprint(storage.IA)+" "+storage.IP)
	//	w.WriteHeader(201)
	//	return
	//}
	//w.Write([]byte("Storage already present in the configuration."))
	//w.WriteHeader(400)
}

func ListStorages(w http.ResponseWriter, r *http.Request) {
	parsedConfig, err := configManager.LoadFile()
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
