package common

import (
	"github.com/casbin/casbin"
	"github.com/prometheus/client_model/go"
	"time"
	"github.com/scionproto/scion/go/lib/addr"
	"strings"
	"github.com/pkg/errors"
	"log"
	"github.com/netsec-ethz/2SMS/common/types"
	"sync"
	"io/ioutil"
	"encoding/json"
)

type AccessController struct {

	enforcer *casbin.Enforcer
	active bool
	lastScrape map[string]time.Time
	rwMutex sync.RWMutex
	CoreASes []*addr.IA
	NeighboringASes []*addr.IA
}

const ScrapePermission = "scrape"
const CoreRole = "core"
const NeighborRole = "neighbor"
const OwnerRole = "owner"

func NewAccessController(modelFile, policyFile string, active bool, ia *addr.IA) *AccessController {
	enforcer := casbin.NewEnforcer(modelFile, policyFile)
	return &AccessController{
		enforcer,
		active,
		make(map[string]time.Time),
		sync.RWMutex{},
		GetCoreASes(*ia),
		GetNeighboringASes(*ia),
	}
}

func (ac *AccessController) LoadPermsFromFile(file string) error {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	var roleDefs []types.Role
	err = json.Unmarshal(data, &roleDefs)
	if err != nil {
		return err
	}

	// Add roles to the policy
	for _, role := range roleDefs {
		for mapping, permissions := range role.Permissions {
			ac.AddRolePermissions(role.Name, mapping, permissions)
		}
	}
	return nil
}

func (ac *AccessController) Disable() {
	ac.active = false
}

func (ac *AccessController) Enable() {
	ac.active = true
}

func (ac *AccessController) Authorized(source, path string) error {
	if ac.active {
		if ac.enforcer.Enforce(source, path, ScrapePermission) {
			// Find window and frequency permissions
			var window, frequency string
			for _, perm := range ac.enforcer.GetPermissionsForUser(source) {
				if strings.HasPrefix(perm[2], "window:") {
					window = perm[2]
				} else if strings.HasPrefix(perm[2], "frequency:") {
					frequency = perm[2]
				}
			}
			if window != "" {
				// Ensure window not expired
				expiration, err := time.Parse(time.RFC3339, strings.SplitAfterN(window, ":", 2)[1])
				if err != nil {
					log.Println(err)
				}
				log.Println("expiration time:", expiration)
				if time.Now().After(expiration) {
					// Remove "scrape" and window permissions
					ac.enforcer.DeletePermissionForUser(source, ScrapePermission)
					ac.enforcer.DeletePermissionForUser(source, window)
					return errors.New("Time window for " + source + " has exipred")
				}
			}
			if frequency != "" {
				// Check last access with current time
				ac.rwMutex.RLock()
				last := ac.lastScrape[source]
				ac.rwMutex.RUnlock()
				now := time.Now()
				freqDuration, _ := time.ParseDuration(strings.Split(frequency, ":")[1])
				if (last == time.Time{}) || now.After(last.Add(freqDuration)) { // TODO: introduce few seconds tolerance?
					// Write new time
					ac.rwMutex.Lock()
					ac.lastScrape[source] = now
					ac.rwMutex.Unlock()
				} else {
					remainingTime := now.Add(freqDuration).Sub(last)
					return errors.New("Next scrape for " + source + " authorized in " + remainingTime.String())
				}
			}
			return nil
		} else {
			return errors.New(source + " not authorized to scrape")
		}
	}
	return nil
}

func (ac *AccessController) FilterMetrics(source, path string, metrics []*io_prometheus_client.MetricFamily) []*io_prometheus_client.MetricFamily {
	filteredMetrics := []*io_prometheus_client.MetricFamily{}
	for _, fam := range metrics {
		if ac.enforcer.Enforce(source, path, *fam.Name) {
			filteredMetrics = append(filteredMetrics, fam)
		}
	}
	return filteredMetrics
}

func (ac *AccessController) GetAllRoles() []string {
	roles := []string{}
	for _, subj := range ac.enforcer.GetAllSubjects() {
		if strings.HasSuffix(subj, "_role") {
			roles = append(roles, subj[:(len(subj) - 5)])
		}
	}
	return roles
}

func (ac *AccessController) CreateRole(role types.Role) {
	internalRoleName := role.Name + "_role"
	for obj, objPerms := range role.Permissions {
		for _, perm := range objPerms {
			ac.enforcer.AddPermissionForUser(internalRoleName, obj, perm)
		}
	}
	ac.enforcer.SavePolicy()
}

func (ac *AccessController) DeleteRole(role string) {
	ac.enforcer.DeleteRole(role + "_role")
	ac.enforcer.SavePolicy()
}

func (ac *AccessController) GetRoles(source string) []string {
	roles := ac.enforcer.GetRolesForUser(source)
	roleNames := make([]string, len(roles))
	for i, role := range roles {
		roleNames[i] = role[:len(role) - 5]
	}
	return roleNames
}

func (ac *AccessController) GetRoleInfo(role string) *types.Role {
	perms := ac.GetSubjectPermissions(role + "_role")
	if len(perms) > 0 {
		return &types.Role{
			Name: role,
			Permissions: perms,
		}
	}
	return nil
}

func (ac *AccessController) sourceInCoreAS(source string) bool {
	for _, coreAS := range ac.CoreASes {
		if strings.HasPrefix(source, coreAS.String()) {
			return true
		}
	}
	return false
}

func (ac *AccessController) sourceInNeighborAS(source string) bool {
	for _, neighAS := range ac.NeighboringASes {
		if strings.HasPrefix(source, neighAS.String()) {
			return true
		}
	}
	return false
}

func (ac *AccessController) AddRole(source string, role string) error {
	if (strings.Contains(role, CoreRole) && ac.sourceInCoreAS(source)) ||
		(strings.Contains(role, NeighborRole) && !ac.sourceInNeighborAS(source)) {
		return errors.New(source + " is not allowed to have reserved role " + role)
	}
	ac.enforcer.AddRoleForUser(source, role + "_role")
	ac.enforcer.SavePolicy()
	return nil
}

func (ac *AccessController) RemoveRole(source string, role string) {
	ac.enforcer.DeleteRoleForUser(source, role + "_role")
	ac.enforcer.SavePolicy()
}

// Expects role to be just the role name and mapping to have e heading /
func (ac *AccessController) AddRolePermissions(role string, mapping string, permissions []string) {
	for _, perm := range permissions {
		ac.enforcer.AddPermissionForUser(mapping[1:] + "_" + role + "_role", mapping, perm)
	}
	ac.enforcer.SavePolicy()
}

func (ac *AccessController) RemoveRolePermissions(role string, mapping string, permissions []string) {
	for _, perm := range permissions {
		ac.enforcer.DeletePermissionForUser(role, mapping, perm)
	}
	ac.enforcer.SavePolicy()
}

// Blocks the given source from scraping the given mapping
func (ac *AccessController) BlockSource(source, mapping string) {
	ac.enforcer.DeletePermissionForUser(source, mapping, ScrapePermission)
	ac.enforcer.SavePolicy()
}

// Allows the given source to scrape the given mapping
func (ac *AccessController) AllowSource(source, mapping string) {
	ac.enforcer.AddPermissionForUser(source, mapping, ScrapePermission)
	ac.enforcer.SavePolicy()
}

// Returns all permissions for the given user (this includes permissions from roles)
func (ac *AccessController) GetAllPermissions(source string) map[string][]string {
	permsMap := ac.GetSubjectPermissions(source)
	// Get permissions from roles
	for _, role := range ac.enforcer.GetRolesForUser(source) {
		rolePerms := ac.enforcer.GetPermissionsForUser(role)
		for _, rolePerm := range rolePerms {
			permsMap[rolePerm[1]] = append(permsMap[rolePerm[1]], rolePerm[2])
		}
	}
	return permsMap
}

// Deletes all permissions associated with a user (i.e. timing, "scrape" and all role assignments)
func (ac *AccessController) DeleteAllPermissions(source string) {
	ac.enforcer.DeletePermissionsForUser(source)
	ac.enforcer.DeleteRolesForUser(source)
	ac.enforcer.SavePolicy()
}

// Deletes all permissions associated with an object (i.e. owner role, scrape and temporal permissions)
func (ac *AccessController) DeleteAllMappingPermissions(mapping string) {
	role := mapping[1:] + "_" + OwnerRole + "_role"
	ac.enforcer.DeleteRole(role)
	ac.enforcer.RemoveFilteredPolicy(1, mapping)
	ac.enforcer.SavePolicy()
}

func (ac *AccessController) GetPermissionsForObject(subject, object string) []string {
	return ac.GetSubjectPermissions(subject)[object]
}

// Returns a map(mapping->permissions) for the given subject
func (ac *AccessController) GetSubjectPermissions(subject string) map[string][]string {
	allPerms := ac.enforcer.GetPermissionsForUser(subject)
	permsMap := make(map[string][]string)
	for _, perm := range allPerms {
		permsMap[perm[1]] = append(permsMap[perm[1]], perm[2])
	}
	return permsMap
}

func (ac *AccessController) AddTimingPermission(source, mapping, typ, duration string) {
	ac.DeleteTimingPermission(source, mapping, typ)
	var permission string
	switch typ {
	case "frequency":
		permission = typ + ":" + duration
	case "window":
		windowDuration, _ := time.ParseDuration(duration) // TODO: handle error
		exp := time.Now().Add(windowDuration).Format(time.RFC3339)
		permission = typ + ":" + exp
	}
	ac.enforcer.AddPermissionForUser(source, mapping, permission)
	ac.enforcer.SavePolicy()
}

func (ac *AccessController) DeleteTimingPermission(source, mapping, typ string) {
	for _, perm := range ac.GetPermissionsForObject(source, mapping) {
		if strings.HasPrefix(perm, typ + ":") {
			ac.enforcer.DeletePermissionForUser(source, mapping, perm)
			ac.enforcer.SavePolicy()
			break
		}
	}
}

// Returns a list of all sources that have some permission
func (ac *AccessController) GetAllSources() []string {
	sources := []string{}
	for _, subj := range ac.enforcer.GetAllSubjects() {
		if !strings.HasSuffix(subj, "_role") {
			sources = append(sources, subj)
		}
	}
	return sources
}