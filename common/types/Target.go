package types

import (
	"fmt"
	"log"
	"regexp"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/config"
	config2 "github.com/prometheus/prometheus/discovery/config"
	"github.com/prometheus/prometheus/discovery/targetgroup"
)

type Target struct {
	Name   string            `json:"name,omitempty"`
	ISD    string            `json:"isd,omitempty"`
	AS     string            `json:"as,omitempty"`
	IP     string            `json:"ip,omitempty"`
	Port   string            `json:"port,omitempty"`
	Path   string            `json:"path,omitempty"`
	Labels map[string]string `json:"labels,omitempty"`
}

func (t *Target) BuildJobName() string {
	return fmt.Sprint(t.ISD) + "-" + fmt.Sprint(t.AS) + " " + t.IP + " " + t.Name
}

func (t *Target) ExistsInConfig(currentConfig *config.Config) bool {
	jobName := t.BuildJobName()
	for _, job := range currentConfig.ScrapeConfigs {
		if job.JobName == jobName {
			return true
		}
	}
	return false
}

// Create and return a new ScrapeConfig object from the Target
func (t *Target) ToScrapeConfig() config.ScrapeConfig {
	fmt.Println(t)
	target1 := make(map[model.LabelName]model.LabelValue)
	target1["__address__"] = model.LabelValue(fmt.Sprint(t.IP) + ":" + fmt.Sprint(t.Port))
	targets := []model.LabelSet{target1}
	labels := make(map[model.LabelName]model.LabelValue)
	for k, v := range t.Labels {
		labels[model.LabelName(k)] = model.LabelValue(fmt.Sprint(v))
	}
	targetGroup := targetgroup.Group{targets, labels, ""}
	scConfigs := []*targetgroup.Group{&targetGroup}
	sdConfig := config2.ServiceDiscoveryConfig{StaticConfigs: scConfigs}
	return config.ScrapeConfig{JobName: t.BuildJobName(), MetricsPath: "/" + t.ISD + "-" + t.AS + t.Path, ServiceDiscoveryConfig: sdConfig}
}

// FromScrapeConfig parses given ScrapeConfig and fill Target fields accordingly
// Assumption: every job has only a single static config and a single target
// If the name is not of the form "17-ffaa:1:43 127.0.0.5 node" it may result in the AS or ISD fields
// of Target not being populated. The labels may be then used to populate them.
func (t *Target) FromScrapeConfig(sc *config.ScrapeConfig) {
	// Parse job name into name, ISD and AS (e.g. 17-ffaa:1:43 127.0.0.5 node)
	re := regexp.MustCompile(`(.+)-(.+) (.+) (.+)`)
	groups := re.FindStringSubmatch(sc.JobName)
	if len(groups) == 5 {
		t.ISD = groups[1]
		t.AS = groups[2]
		t.Name = groups[4]
	} else {
		// cannot guess the ISD or AS
		t.ISD = ""
		t.AS = ""
		t.Name = sc.JobName
	}
	// Parse url into IP and Port
	re = regexp.MustCompile(`(.+):(\d+)`)
	groups = re.FindStringSubmatch(string(sc.ServiceDiscoveryConfig.StaticConfigs[0].Targets[0]["__address__"]))
	if len(groups) != 3 {
		log.Printf("Reading Target from prometheus configuration: could not parse address of '%s'", sc.JobName)
		t.IP = ""
		t.Port = ""
	} else {
		t.IP = groups[1]
		t.Port = groups[2]
	}
	// Get metrics path
	t.Path = sc.MetricsPath
	// Get labels
	labels := make(map[string]string)
	for k, v := range sc.ServiceDiscoveryConfig.StaticConfigs[0].Labels {
		kk := string(k)
		vv := string(v)
		labels[kk] = vv
		// if we couldn't get ISD or AS, but we have those labels, populate them
		if t.ISD == "" && kk == "ISD" {
			t.ISD = vv
		}
		if t.AS == "" && kk == "AS" {
			t.AS = vv
		}
	}
	t.Labels = labels
}
