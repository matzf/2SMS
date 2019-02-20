package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"

	"github.com/netsec-ethz/2SMS/common"
	"github.com/netsec-ethz/2SMS/common/types"
)

func LoadMappings() (types.EndpointMappings, error) {
	dat, err := ioutil.ReadFile("mappings.json")
	if err != nil {
		return nil, err
	}
	var list []types.Mapping
	if err := json.Unmarshal(dat, &list); err != nil {
		return nil, err
	}
	tmpMapping := make(types.EndpointMappings)
	for _, mapping := range list {
		tmpMapping[mapping.Path] = mapping.Port
	}
	return tmpMapping, nil
}

func SaveMappings(mappings types.EndpointMappings) error {
	var list []types.Mapping
	for k, v := range mappings {
		list = append(list, types.Mapping{Path: k, Port: v})
	}
	bytes, err := json.Marshal(list)
	if err != nil {
		return err
	}
	return ioutil.WriteFile("mappings.json", bytes, 0644)
}

func UpdateMappings(newMapping types.Mapping) error {
	_, err := UpdateMappingBatch([]types.Mapping{newMapping})
	return err
}
func UpdateMappingBatch(mappings []types.Mapping) (types.EndpointMappings, error) {
	log.Println("Adding the following mappings: ", mappings)
	reloadMappingsMutex.Lock()
	defer reloadMappingsMutex.Unlock()
	added := types.EndpointMappings{}
	for _, m := range mappings {
		internalMapping[m.Path] = m.Port
		added[m.Path] = m.Port
	}
	return added, SaveMappings(internalMapping)
}

func RemoveMapping(toRemove string) error {
	return RemoveMappingBatch([]string{toRemove})
}

func RemoveMappingBatch(pathsToRemove []string) error {
	reloadMappingsMutex.Lock()
	defer reloadMappingsMutex.Unlock()
	for _, p := range pathsToRemove {
		delete(internalMapping, p)
	}
	return SaveMappings(internalMapping)
}

func RemoveMappingRegexpBatch(pathsToRemove []string) (types.EndpointMappings, error) {
	reloadMappingsMutex.Lock()
	defer reloadMappingsMutex.Unlock()
	log.Println("Removing these regular expressions: ", pathsToRemove)
	var removeRegExprs []*regexp.Regexp
	for _, p := range pathsToRemove {
		r, err := regexp.Compile(p)
		if err != nil {
			log.Printf("Error compiling regular expression: %s: %v\n", p, err)
			continue
		}
		removeRegExprs = append(removeRegExprs, r)
	}
	removed := types.EndpointMappings{}
	for s := range internalMapping {
		removeIt := false
		for _, e := range removeRegExprs {
			if e.MatchString(s) {
				removeIt = true
				break
			}
		}
		if removeIt {
			log.Printf("Removed %s\n", s)
			removed[s] = internalMapping[s]
			delete(internalMapping, s)
		}
	}
	return removed, SaveMappings(internalMapping)
}

func SyncPermissions(addMappings, delMappings types.EndpointMappings) {
	for path := range delMappings {
		accessController.DeleteAllMappingPermissions(path)
	}
	for mapping := range addMappings {
		metricInfos := GetMetricsInfoForMapping(mapping, localHTTPClient)
		perms := make([]string, len(metricInfos))
		for i, info := range metricInfos {
			perms[i] = info.Name
		}
		accessController.AddRolePermissions("owner", mapping, perms)
	}
}

func SyncManager(addMappings, delMappings types.EndpointMappings) error {
	// Register at manager
	if managerIP == "" {
		return nil
	}
	var paths []string

	// Remove all delMappings:
	// Remove mapping from any scraper that has it as target
	target := types.Target{}
	target.AS = local.IA.A.String()
	target.ISD = fmt.Sprint(local.IA.I)
	target.IP = endpointIP
	target.Port = fmt.Sprint(local.Host.L4.Port())
	target.Labels = make(map[string]string)

	for path := range delMappings {
		target.Path = path
		target.Name = path[1:]
		jsonBytes, err := json.Marshal(target)
		if err != nil {
			return fmt.Errorf("Error while marshaling json: %v", err)
		}
		req, err := http.NewRequest("DELETE", "https://"+managerIP+":"+managerVerifPort+"/endpoint/mappings/notify", bytes.NewReader(jsonBytes))
		if err != nil {
			return fmt.Errorf("Error creating request: %v", err)
		}
		_, err = httpsClient.Do(req)
		if err != nil {
			return fmt.Errorf("Error contacting manager to delete a target: %v", err)
		}
	}

	// Add all addMappings:
	for mapping := range addMappings {
		paths = append(paths, mapping)
	}
	data, err := json.Marshal(types.Endpoint{
		IA:         local.IA.String(),
		IP:         local.Host.L3.IP().String(),
		ScrapePort: fmt.Sprint(local.Host.L4.Port()),
		ManagePort: managementAPIPort,
		Paths:      paths,
	})
	if err != nil {
		return fmt.Errorf("Failed marshalling Endpoint struct: %v", err)
	}
	resp, err := httpsClient.Post("https://"+managerIP+":"+managerVerifPort+"/endpoints/register", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("Failed sending registration request: %v", err)
	}
	if resp.StatusCode != 200 {
		message, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("Registration failed. Status code: %v , Message: %v", resp.StatusCode, message)
	}
	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Failed reading body: %v", err)
	}
	var addedToScrapers []types.Scraper
	err = json.Unmarshal(data, &addedToScrapers)
	if err != nil {
		return fmt.Errorf("Could not unmarshal manager response. Error is: %v\nBody is: %s", err, string(data))
	}
	// Add owner role and scrape permission to each scraper
	for _, path := range paths {
		for _, scr := range addedToScrapers {
			accessController.AddRole(scr.IA+":"+scr.IP, path[1:]+"_"+common.OwnerRole)
			accessController.AllowSource(scr.IA+":"+scr.IP, path)
		}
	}
	return nil
}

func GetMetricsInfoForMapping(mapping string, client *http.Client) []*MetricInfo {
	resp, err := LocalhostGet(mapping, client)
	if err != nil {
		return []*MetricInfo{}
	}
	// Parse response body to metric families
	metrics := DecodeResponseBody(resp)
	// Keep only name, type and help fields
	metricsInfo := make([]*MetricInfo, len(metrics))
	for i, metric := range metrics {
		metricsInfo[i] = &MetricInfo{*metric.Name, metric.Type.String(), *metric.Help}
	}
	return metricsInfo
}