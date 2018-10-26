package main

import (
	"io/ioutil"
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/baehless/2SMS/common/types"
)

func InitInternalMappings(nodeListenPort, genFolder string) []types.Mapping {
	mappings := []types.Mapping{}
	if nodeListenPort != "" {
		nodeMapping := types.Mapping{Path: "/node", Port: nodeListenPort}
		mappings = append(mappings, nodeMapping)
	}
	if genFolder != "" {
		// TODO: Load mapping for SCION service from gen folder
	}
	return mappings
}

func LoadInternalMappings() (map[string]string, error) {
	list, err := readMappings()
	if err != nil {
		return nil, err
	}
	tmpMapping := make(map[string]string)
	for _, mapping := range list {
		tmpMapping[mapping.Path] = mapping.Port
	}
	return tmpMapping, nil
}


func AddInternalMapping(mapping types.Mapping) error {
	list, err := readMappings()
	if err != nil {
		return err
	}
	for _, mp := range list {
		if mp.Equal(&mapping) {
			return errors.New("Mapping already present.")
		}
	}
	// Write back to file extended list
	return writeMappings(append(list, mapping))
}

func RemoveInternalMapping(mapping types.Mapping) error {
	list, err := readMappings()
	if err != nil {
		return err
	}
	// Scan list and copy to new one any element but the one to remove
	var newList []types.Mapping
	for _, elem := range list {
		if elem != mapping {
			newList = append(newList, elem)
		}
	}
	if len(newList) == len(list) {
		return errors.New("Mapping doesn't exists and cannot be removed.")
	}
	writeMappings(newList)
	return nil
}

func readMappings() ([]types.Mapping, error) {
	dat, err := ioutil.ReadFile("mappings.json")
	if err != nil {
		return nil, err
	}
	var list []types.Mapping
	if err := json.Unmarshal(dat, &list); err != nil {
		return nil, err
	}
	return list, nil
}

func writeMappings(mappings []types.Mapping) error {
	bytes, err := json.Marshal(mappings)
	if err != nil {
		return err
	}
	return ioutil.WriteFile("mappings.json", bytes, 0644)
}