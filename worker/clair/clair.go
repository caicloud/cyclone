/*
Copyright 2016 caicloud authors. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package clair

import (
	"fmt"
	"sort"

	"github.com/caicloud/cyclone/docker"
	"github.com/caicloud/cyclone/pkg/log"
	"github.com/caicloud/cyclone/pkg/osutil"
	klar_clair "github.com/optiopay/klar/clair"
	klar_docker "github.com/optiopay/klar/docker"

	"github.com/caicloud/cyclone/api"
)

const (
	// CLAIR_SERVER_IP is the IP of clair server.
	CLAIR_SERVER_IP = "CLAIR_SERVER_IP"
)

// Serverity is the type for level of checks.
type Serverity string

const (
	// Unknown is either a security problem that has not been
	// assigned to a priority yet or a priority that our system
	// did not recognize
	Unknown Serverity = "Unknown"
	// Negligible is technically a security problem, but is
	// only theoretical in nature, requires a very special
	// situation, has almost no install base, or does no real
	// damage. These tend not to get backport from upstreams,
	// and will likely not be included in security updates unless
	// there is an easy fix and some other issue causes an update.
	Negligible Serverity = "Negligible"
	// Low is a security problem, but is hard to
	// exploit due to environment, requires a user-assisted
	// attack, a small install base, or does very little damage.
	// These tend to be included in security updates only when
	// higher priority issues require an update, or if many
	// low priority issues have built up.
	Low Serverity = "Low"
	// Medium is a real security problem, and is exploitable
	// for many people. Includes network daemon denial of service
	// attacks, cross-site scripting, and gaining user privileges.
	// Updates should be made soon for this priority of issue.
	Medium Serverity = "Medium"
	// High is a real problem, exploitable for many people in a default
	// installation. Includes serious remote denial of services,
	// local root privilege escalations, or data loss.
	High Serverity = "High"
	// Critical is a world-burning problem, exploitable for nearly all people
	// in a default installation of Linux. Includes remote root
	// privilege escalations, or massive data loss.
	Critical Serverity = "Critical"
	// Defcon1 is a Critical problem which has been manually highlighted by
	// the team. It requires an immediate attention.
	Defcon1 Serverity = "Defcon1"
)

var SeverityWeight map[Serverity]int

type VulnerabilityList []klar_clair.Vulnerability

func init() {
	// init serverity weight map
	SeverityWeight = make(map[Serverity]int)
	weight := 0
	SeverityWeight[Unknown] = weight
	weight++
	SeverityWeight[Negligible] = weight
	weight++
	SeverityWeight[Low] = weight
	weight++
	SeverityWeight[Medium] = weight
	weight++
	SeverityWeight[High] = weight
	weight++
	SeverityWeight[Critical] = weight
	weight++
	SeverityWeight[Defcon1] = weight
}

// AnalysisImage analysis image by Clair server
func AnalysisImage(dockerManager *docker.Manager,
	imageName string) ([]klar_clair.Vulnerability, error) {
	var vulnerabilities []klar_clair.Vulnerability

	image, err := klar_docker.NewImage(imageName, dockerManager.AuthConfig.Username,
		dockerManager.AuthConfig.Password)
	if err != nil {
		log.Errorf("new images err: %v", err)
		return vulnerabilities, err
	}

	err = image.Pull()
	if err != nil {
		log.Errorf("get image layer info err: %v", err)
		return vulnerabilities, err
	}

	clairServerAddr := osutil.GetStringEnv(CLAIR_SERVER_IP,
		"http://localhost:6060")
	clairClient := klar_clair.NewClair(clairServerAddr)
	vulnerabilities = clairClient.Analyse(image)
	sort.Sort(VulnerabilityList(vulnerabilities))

	return vulnerabilities, nil
}

func compareSeverity(severity1 Serverity, severity2 Serverity) bool {
	return SeverityWeight[severity1] > SeverityWeight[severity2]
}

// Len realize function of interface sort
func (list VulnerabilityList) Len() int {
	return len(list)
}

// Less realize function of interface sort
func (list VulnerabilityList) Less(i, j int) bool {
	return compareSeverity(Serverity(list[i].Severity), Serverity(list[j].Severity))
}

// Swap realize function of interface sort
func (list VulnerabilityList) Swap(i, j int) {
	temp := list[i]
	list[i] = list[j]
	list[j] = temp
}

// Analysis Analyses the image.
func Analysis(event *api.Event, dmanager *docker.Manager) error {
	imagename, ok := event.Data["image-name"]
	tagname, ok2 := event.Data["tag-name"]

	if !ok || !ok2 {
		return fmt.Errorf("Unable to retrieve image name")
	}
	imageName := imagename.(string) + ":" + tagname.(string)

	vulnerabilities, err := AnalysisImage(dmanager, imageName)

	if err != nil {
		log.Errorf("clair analysis %s err: %v", imageName, err)
		return err
	}
	event.Version.SecurityCheck = true
	for _, vulnerability := range vulnerabilities {
		if vulnerability.Severity == string(Medium) ||
			vulnerability.Severity == string(High) ||
			vulnerability.Severity == string(Critical) ||
			vulnerability.Severity == string(Defcon1) {
			security := api.Security{}
			security.Name = vulnerability.Name
			security.Description = vulnerability.Description
			security.Severity = vulnerability.Severity
			event.Version.SecurityInfo = append(event.Version.SecurityInfo, security)
		}
	}

	return nil
}
