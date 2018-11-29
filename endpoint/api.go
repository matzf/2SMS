package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/netsec-ethz/2SMS/common/types"
)

func listMappings(w http.ResponseWriter, r *http.Request) {
	// Try loading into a temp map
	mappings, err := LoadMappings()
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
	err = UpdateMappings(*mapping)
	if err != nil {
		w.WriteHeader(500)
		log.Println("Failed adding internal mapping:", err)
		return
	}

	// Add metric permissions for the new target to "owner_role"
	thisMapping := types.EndpointMappings{mapping.Path: mapping.Port}
	SyncPermissions(thisMapping, types.EndpointMappings{})
	err = SyncManager(thisMapping, types.EndpointMappings{})
	if err != nil {
		w.WriteHeader(500)
		log.Println("Failed to sync against the manager:", err)
		return
	}
	w.WriteHeader(201)
}

// putMappings receives a Message
func putMappings(w http.ResponseWriter, r *http.Request) {
	type Message struct {
		RemoveRegex []string        `json:"removeRegex,omitempty"`
		Add         []types.Mapping `json:"add,omitempty"`
	}
	var mappings Message
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Could not read the body: %v", err)
		w.WriteHeader(500)
		return
	}
	err = json.Unmarshal(data, &mappings)
	if err != nil {
		w.WriteHeader(400)
		log.Printf("Failed parsing request body %v: %v\n", string(data), err)
		return
	}
	removals := mappings.RemoveRegex
	removed, err := RemoveMappingRegexpBatch(removals)
	if err != nil {
		w.WriteHeader(400)
		log.Printf("Failed removing: %v.\nRequest body: %v\n", err, string(data))
		return
	}
	additions := mappings.Add
	added, err := UpdateMappingBatch(additions)
	if err != nil {
		w.WriteHeader(400)
		log.Printf("Failed adding: %v.\nRequest body: %v\n", err, string(data))
		return
	}
	SyncPermissions(added, removed)
	err = SyncManager(added, removed)
	if err != nil {
		w.WriteHeader(500)
		log.Println("Failed to sync against the manager:", err)
		return
	}
	w.WriteHeader(204)
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
	err = RemoveMapping(mapping.Path)
	if err != nil {
		w.WriteHeader(500)
		log.Println("Failed removing internal mapping:", err)
		return
	}
	// Remove all scrape and temporal permissions associated with the removed mapping
	SyncPermissions(types.EndpointMappings{}, types.EndpointMappings{mapping.Path: mapping.Port})
	err = SyncManager(types.EndpointMappings{}, types.EndpointMappings{mapping.Path: mapping.Port})
	if err != nil {
		w.WriteHeader(500)
		log.Printf("Could not sync with manager while removing target. Error is: %v", err)
		return
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
	Name string `json:"name,omitempty"`
	Type string `json:"type,omitempty"`
	Help string `json:"help,omitempty"`
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
	accessController.BlockSource(source, "/"+mapping)
	w.WriteHeader(204)
}

// Allows the source to scrape with the current permissions
func enableSource(w http.ResponseWriter, r *http.Request) {
	source := mux.Vars(r)["source"]
	mapping := mux.Vars(r)["mapping"]
	accessController.AllowSource(source, "/"+mapping)
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
			} else if strings.HasPrefix(perm, "frequency:") {
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
	CanScrape bool
	Frequency string
	Until     string
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
	mapping := "/" + mux.Vars(r)["mapping"]
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
	mapping := "/" + mux.Vars(r)["mapping"]
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
	mapping := "/" + mux.Vars(r)["mapping"]
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
	mapping := "/" + mux.Vars(r)["mapping"]
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

func enableAccessControl(w http.ResponseWriter, r *http.Request) {
	accessController.Enable()
	w.WriteHeader(204)
}

func disableAccessControl(w http.ResponseWriter, r *http.Request) {
	accessController.Disable()
	w.WriteHeader(204)
}

func listSources(w http.ResponseWriter, r *http.Request) {
	jsonSources, err := json.Marshal(accessController.GetAllSources())
	if err != nil {
		log.Println("Error while marshalling json:", err)
		w.WriteHeader(500)
		return
	}
	w.Write(jsonSources)
}
