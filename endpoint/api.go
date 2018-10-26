package main

import (
	"net/http"
	"log"
	"encoding/json"
	"io/ioutil"
	"github.com/baehless/2SMS/common"
	"github.com/baehless/2SMS/common/types"
	"bytes"
	"fmt"
	"github.com/gorilla/mux"
	"strings"
)

func listMappings(w http.ResponseWriter, r *http.Request) {
	// Try loading into a temp map
	mappings, err := LoadInternalMappings()
	if err != nil {
		log.Println("Failed reloading internal mappings:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	jsonMappings, err := json.Marshal(mappings)
	if err != nil {
		log.Println("Error while marshalling json:", err)
		w.WriteHeader(500)
		return
	}
	w.Write(jsonMappings)
}

func addMapping(w http.ResponseWriter, r *http.Request) {
	// Parse request
	mapping, err := parseChangeRequest(r)
	if err != nil {
		w.WriteHeader(400)
		log.Println("Failed parsing request body:", err)
		return
	}
	// Add mapping to the forwarding list
	err = AddInternalMapping(*mapping)
	if err != nil {
		w.WriteHeader(500)
		log.Println("Failed adding internal mapping:", err)
		return
	}
	updateMappings()

	// Add metric permissions for the new target to "owner_role"
	metricsInfo := GetMetricsInfoForMapping(mapping.Path, http.Client{})
	metricsNames := make([]string, len(metricsInfo))
	for i, metric := range metricsInfo {
		metricsNames[i] = metric.Name
	}
	accessController.AddRolePermissions(common.OwnerRole, mapping.Path, metricsNames)

	if managerIP != "" {
		// Add mapping as target to all scrapers
		client := common.CreateHttpsClient(caCertsDir, endpointCert, endpointPrivKey)
		// Build target, marshal it and pass it as body
		target := types.Target{}
		target.AS = local.IA.A.String()
		target.ISD = fmt.Sprint(local.IA.I)
		target.IP = endpointIP
		target.Path = mapping.Path
		target.Port = fmt.Sprint(local.L4Port)
		target.Labels = make(map[string]string)
		target.Labels["AS"] = target.AS
		target.Labels["ISD"] = target.ISD
		target.Labels["service"] = target.Path[1:] // Assumes path is of the form `/<service-name>`
		target.Name = target.Path[1:]
		jsonBytes, err := json.Marshal(target)
		if err != nil {
			log.Println("Error while marshaling json:", err)
			w.WriteHeader(500)
			return
		}
		resp, err := client.Post("https://"+managerIP+":"+managerVerifPort+"/endpoint/mappings/notify", "application/json", bytes.NewReader(jsonBytes))
		// Add role owner and scrape permission to each scraper
		addScraperPermissions(resp, mapping.Path)
	}
	w.WriteHeader(201)
}

// Add owner role and scrape permission to each scraper
func addScraperPermissions(resp *http.Response, mapping string) {
	var addedToScrapers []types.Scraper
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Failed reading body:", err)
		return
	}
	err = json.Unmarshal(data, &addedToScrapers)
	// TODO: handle error
	for _, scr := range addedToScrapers {
		accessController.AddRole(scr.IA + ":" + scr.IP, mapping[1:] + "_" + common.OwnerRole)
		accessController.AllowSource(scr.IA + ":" + scr.IP, mapping)
	}
}

func removeMapping(w http.ResponseWriter, r *http.Request) {
	// Parse request
	mapping, err := parseChangeRequest(r)
	if err != nil {
		w.WriteHeader(400)
		log.Println("Failed parsing request body:", err)
		return
	}

	// Remove mapping from local forwarding list
	err = RemoveInternalMapping(*mapping)
	if err != nil {
		w.WriteHeader(500)
		log.Println("Failed removing internal mapping:", err)
		return
	}
	updateMappings()

	// Remove all scrape and temporal permissions associated with the removed mapping
	accessController.DeleteAllMappingPermissions(mapping.Path)

	if managerIP != "" {
		// Remove mapping from any scraper that has it as target
		client := common.CreateHttpsClient(caCertsDir, endpointCert, endpointPrivKey)
		target := types.Target{}
		target.AS = local.IA.A.String()
		target.ISD = fmt.Sprint(local.IA.I)
		target.IP = endpointIP
		target.Path = mapping.Path
		target.Port = fmt.Sprint(local.L4Port)
		target.Labels = make(map[string]string)
		target.Name = target.Path[1:]
		jsonBytes, err := json.Marshal(target)
		if err != nil {
			log.Println("Error while marshaling json:", err)
			w.WriteHeader(500)
			return
		}
		req, err := http.NewRequest("DELETE", "https://"+managerIP+":"+managerVerifPort+"/endpoint/mappings/notify", bytes.NewReader(jsonBytes))
		_, err = client.Do(req)
	}
	w.WriteHeader(204)
}

func parseChangeRequest(r *http.Request) (*types.Mapping, error) {
	var mapping types.Mapping
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &mapping)
	if err != nil {
		return nil, err
	}
	return &mapping, nil
}

func updateMappings() error {
	// Try loading into a temp map
	tmpMappings, err := LoadInternalMappings()
	if err != nil {
		return err
	}
	// Lock global variable
	reloadMappingsMutex.Lock()
	// Reassign mappings
	internalMapping = tmpMappings
	// Unlock global variable
	reloadMappingsMutex.Unlock()
	log.Println("Successfully reloaded internal mappings:", internalMapping)
	return nil
}

func listMetrics(w http.ResponseWriter, r *http.Request) {
	mapping := mux.Vars(r)["mapping"]
	if mapping == "" {
		w.WriteHeader(400)
		return
	}
	metricsInfo := GetMetricsInfoForMapping(mapping, http.Client{})
	// Marshal to json and write in response
	jsonMetricsInfo, err := json.Marshal(metricsInfo)
	if err != nil {
		log.Println("Error while marshalling json:", err)
		w.WriteHeader(500)
		return
	}
	w.Write(jsonMetricsInfo)
}

type MetricInfo struct {
	Name	string	`json:"name,omitempty"`
	Type	string	`json:"type,omitempty"`
	Help	string	`json:"help,omitempty"`
}

func listSourceRoles(w http.ResponseWriter, r *http.Request) {
	source := mux.Vars(r)["source"]
	jsonRoles, err := json.Marshal(accessController.GetRoles(source))
	if err != nil {
		log.Println("Error while marshalling json:", err)
		w.WriteHeader(500)
		return
	}
	w.Write(jsonRoles)
}

func addSourceRole(w http.ResponseWriter, r *http.Request) {
	source := mux.Vars(r)["source"]
	var role string
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &role)
	if err != nil {
		return
	}
	err = accessController.AddRole(source, role)
	if err != nil {
		log.Println("Error in adding role:", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(204)
}

func removeSourceRole(w http.ResponseWriter, r *http.Request) {
	source := mux.Vars(r)["source"]
	var role string
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &role)
	if err != nil {
		log.Println("Error while unmarshalling json:", err)
		w.WriteHeader(500)
		return
	}
	accessController.RemoveRole(source, role)
	w.WriteHeader(204)
}

// Blocks the source's scrape permissions without removing them
func blockSource(w http.ResponseWriter, r *http.Request) {
	source := mux.Vars(r)["source"]
	mapping := mux.Vars(r)["mapping"]
	accessController.BlockSource(source, "/" + mapping)
	w.WriteHeader(204)
}

// Allows the source to scrape with the current permissions
func enableSource(w http.ResponseWriter, r *http.Request) {
	source := mux.Vars(r)["source"]
	mapping := mux.Vars(r)["mapping"]
	accessController.AllowSource(source, "/" + mapping)
	w.WriteHeader(204)
}

// Shows if the source can scrape, how often and until when which mapping
func sourceStatus(w http.ResponseWriter, r *http.Request) {
	source := mux.Vars(r)["source"]
	allPerms := accessController.GetSubjectPermissions(source)
	globalStatus := make(map[string]Status)
	for obj, perms := range allPerms {
		status := Status{false, "unlimited", "unlimited"}
		for _, perm := range perms {
			if perm == "scrape" {
				status.CanScrape = true
			} else if strings.HasPrefix(perm,"frequency:") {
				status.Frequency = strings.Split(perm, ":")[1]

			} else if strings.HasPrefix(perm, "window:") {
				status.Until = strings.SplitAfterN(perm, ":", 2)[1]
			}
		}
		globalStatus[obj] = status
	}
	jsonStatus, err := json.Marshal(globalStatus)
	if err != nil {
		log.Println("Error while marshalling json:", err)
		w.WriteHeader(500)
		return
	}
	w.Write(jsonStatus)
}

type Status struct {
	CanScrape	bool
	Frequency	string
	Until		string
}

func removeAllSourcePermissions(w http.ResponseWriter, r *http.Request) {
	source := mux.Vars(r)["source"]
	accessController.DeleteAllPermissions(source)
	w.WriteHeader(204)
}

func listAllSourcePermissions(w http.ResponseWriter, r *http.Request) {
	source := mux.Vars(r)["source"]
	jsonPerms, err := json.Marshal(accessController.GetAllPermissions(source))
	if err != nil {
		log.Println("Error while marshalling json:", err)
		w.WriteHeader(500)
		return
	}
	w.Write(jsonPerms)
}

func listRoles(w http.ResponseWriter, r *http.Request) {
	jsonRoles, err := json.Marshal(accessController.GetAllRoles())
	if err != nil {
		log.Println("Error while marshalling json:", err)
		w.WriteHeader(500)
		return
	}
	w.Write(jsonRoles)
}

func createRole(w http.ResponseWriter, r *http.Request) {
	var role types.Role
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("Error while reading request body:", err)
		w.WriteHeader(500)
		return
	}
	err = json.Unmarshal(data, &role)
	if err != nil {
		log.Println("Error while unmarshalling json:", err)
		w.WriteHeader(400)
		return
	}
	accessController.CreateRole(role)
	w.WriteHeader(201)
}

func deleteRole(w http.ResponseWriter, r *http.Request) {
	role := mux.Vars(r)["role"]
	accessController.DeleteRole(role)
	w.WriteHeader(204)
}

func getRoleInfo(w http.ResponseWriter, r *http.Request) {
	role := mux.Vars(r)["role"]
	info := accessController.GetRoleInfo(role)
	if info == nil {
		w.Write([]byte("Inexisting role: " + role))
		w.WriteHeader(500)
		return
	}
	jsonInfo, err := json.Marshal(info)
	if err != nil {
		log.Println("Error while marshalling json:", err)
		w.WriteHeader(400)
		return
	}
	w.Write(jsonInfo)
}

// Adds all given permissions for a mapping to a role
func addRolePermissions(w http.ResponseWriter, r *http.Request) {
	role := mux.Vars(r)["role"]
	mapping :=  "/" + mux.Vars(r)["mapping"]
	var perms []string
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.Write([]byte("Inexisting role: " + role))
		w.WriteHeader(500)
		return
	}
	err = json.Unmarshal(data, &perms)
	if err != nil {
		log.Println("Error while marshalling json:", err)
		w.WriteHeader(400)
		return
	}
	accessController.AddRolePermissions(role, mapping, perms)
	w.WriteHeader(204)
}

func removeRolePermissions(w http.ResponseWriter, r *http.Request) {
	role := mux.Vars(r)["role"]
	mapping :=  "/" + mux.Vars(r)["mapping"]
	var perms []string
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.Write([]byte("Inexisting role: " + role))
		w.WriteHeader(500)
		return
	}
	err = json.Unmarshal(data, &perms)
	if err != nil {
		log.Println("Error while marshalling json:", err)
		w.WriteHeader(400)
		return
	}
	accessController.RemoveRolePermissions(role, mapping, perms)
	w.WriteHeader(204)
}

func removeSourceFrequency(w http.ResponseWriter, r *http.Request) {
	source := mux.Vars(r)["source"]
	mapping :=  "/" + mux.Vars(r)["mapping"]
	accessController.DeleteTimingPermission(source, mapping, "frequency")
	w.WriteHeader(204)
}

func setSourceFrequency(w http.ResponseWriter, r *http.Request) {
	source := mux.Vars(r)["source"]
	mapping := "/" + mux.Vars(r)["mapping"]
	var duration string
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &duration)
	if err != nil {
		return
	}
	accessController.AddTimingPermission(source, mapping, "frequency", duration)
	w.WriteHeader(204)
}

func removeSourceWindow(w http.ResponseWriter, r *http.Request) {
	source := mux.Vars(r)["source"]
	mapping :=  "/" + mux.Vars(r)["mapping"]
	accessController.DeleteTimingPermission(source, mapping, "window")
	w.WriteHeader(204)
}

func setSourceWindow(w http.ResponseWriter, r *http.Request) {
	source := mux.Vars(r)["source"]
	mapping := "/" + mux.Vars(r)["mapping"]
	var duration string
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &duration)
	if err != nil {
		return
	}
	accessController.AddTimingPermission(source, mapping, "window", duration)
	w.WriteHeader(204)
}

func enableAccessControl (w http.ResponseWriter, r *http.Request) {
	accessController.Enable()
	w.WriteHeader(204)
}

func disableAccessControl (w http.ResponseWriter, r *http.Request) {
	accessController.Disable()
	w.WriteHeader(204)
}

func listSources (w http.ResponseWriter, r *http.Request) {
	jsonSources, err := json.Marshal(accessController.GetAllSources())
	if err != nil {
		log.Println("Error while marshalling json:", err)
		w.WriteHeader(500)
		return
	}
	w.Write(jsonSources)
}