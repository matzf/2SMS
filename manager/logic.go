package main

import (
	"io/ioutil"
	"log"
	"encoding/json"
	"github.com/baehless/2SMS/common/types"
	"github.com/pkg/errors"
)

func getScrapers() []types.Scraper {
	jsonScrs, err := ioutil.ReadFile("scrapers.json")
	if err != nil {
		log.Println("Error while reading file:", err)
		return nil
	}
	var scrs []types.Scraper
	err = json.Unmarshal(jsonScrs, &scrs)
	if err != nil {
		log.Println("Error while unmarshalling json:", err)
		return nil
	}
	return scrs
}

func getScraperByIP(ip string) *types.Scraper {
	for _, scr := range getScrapers() {
		if scr.IP == ip {
			return &scr
		}
	}
	return nil
}

func addScraper(scraper *types.Scraper) error {
	scrapers := getScrapers()
	// If scraper already registered just return
	for _, scr := range scrapers {
		if scraper.Equal(&scr) {
			return nil
		}
	}
	// Add new scraper to the list and write to file
	scrapers = append(scrapers, *scraper)
	jsonScrs, err := json.Marshal(scrapers)
	if err != nil {
		return errors.New("Error marshalling json: " + err.Error())
	}
	return ioutil.WriteFile("scrapers.json", jsonScrs, 0644)
}

func RemoveScraper(scraper *types.Scraper) error {
	scrapers := getScrapers()
	new_scrapers := []types.Scraper{}

	// Copy other storages
	for _, scr := range scrapers {
		if !scraper.Equal(&scr) {
			new_scrapers = append(new_scrapers, scr)
		}
	}
	jsonScrs, err := json.Marshal(new_scrapers)
	if err != nil {
		return errors.New("Error marshalling json: " + err.Error())
	}
	return ioutil.WriteFile("scrapers.json", jsonScrs, 0644)
}

func getEndpoints() []types.Endpoint {
	jsonEnds, err := ioutil.ReadFile("endpoints.json")
	if err != nil {
		log.Println("Error while reading file:", err)
		return nil
	}
	var ends []types.Endpoint
	err = json.Unmarshal(jsonEnds, &ends)
	if err != nil {
		log.Println("Error while unmarshalling json:", err)
		return nil
	}
	return ends
}

func addEndpoint(endpoint *types.Endpoint) error {
	endpoints := getEndpoints()
	// If endpoint already registered just return
	for _, end := range endpoints {
		if endpoint.Equal(&end) {
			return nil
		}
	}
	// Add new scraper to the list and write to file
	endpoints = append(endpoints, *endpoint)
	jsonEnds, err := json.Marshal(endpoints)
	if err != nil {
		return errors.New("Error marshalling json: " + err.Error())
	}
	return ioutil.WriteFile("endpoints.json", jsonEnds, 0644)
}

func RemoveEndpoint(endpoint *types.Endpoint) error {
	endpoints := getEndpoints()
	new_endpoints := []types.Endpoint{}

	// Copy other storages
	for _, end := range endpoints {
		if !endpoint.Equal(&end) {
			new_endpoints = append(new_endpoints, end)
		}
	}
	jsonEnds, err := json.Marshal(new_endpoints)
	if err != nil {
		return errors.New("Error marshalling json: " + err.Error())
	}
	return ioutil.WriteFile("endpoints.json", jsonEnds, 0644)
}

func getEndpointByIP(ip string) *types.Endpoint {
	for _, end := range getEndpoints() {
		if end.IP == ip {
			return &end
		}
	}
	return nil
}

func addStorage(storage *types.Storage) error {
	storages := getStorages()
	// If scraper already registered just return
	for _, str := range storages {
		if storage.Equal(&str) {
			return nil
		}
	}
	// Add new scraper to the list and write to file
	storages = append(storages, *storage)
	jsonScrs, err := json.Marshal(storages)
	if err != nil {
		return errors.New("Error marshalling json: " + err.Error())
	}
	return ioutil.WriteFile("storages.json", jsonScrs, 0644)
}

func RemoveStorage(storage *types.Storage) error {
	storages := getStorages()
	new_storages := []types.Storage{}

	// Copy other storages
	for _, str := range storages {
		if !storage.Equal(&str) {
			new_storages = append(new_storages, str)
		}
	}
	jsonScrs, err := json.Marshal(storages)
	if err != nil {
		return errors.New("Error marshalling json: " + err.Error())
	}
	return ioutil.WriteFile("storages.json", jsonScrs, 0644)
}

func getStorages() []types.Storage {
	jsonStrs, err := ioutil.ReadFile("storages.json")
	if err != nil {
		log.Println("Error while reading file:", err)
		return nil
	}
	var strs []types.Storage
	err = json.Unmarshal(jsonStrs, &strs)
	if err != nil {
		log.Println("Error while unmarshalling json:", err)
		return nil
	}
	return strs
}