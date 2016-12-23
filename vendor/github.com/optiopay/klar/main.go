package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/optiopay/klar/clair"
	"github.com/optiopay/klar/docker"
)

var priorities = []string{"Unknown", "Negligible", "Low", "Medium", "High", "Critical", "Defcon1"}
var store = make(map[string][]clair.Vulnerability)

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Image name must be provided")
		os.Exit(1)
	}

	clairAddr := os.Getenv("CLAIR_ADDR")
	if clairAddr == "" {
		fmt.Printf("Clair address must be provided")
		os.Exit(1)
	}

	threshold := 0
	thresholdStr := os.Getenv("CLAIR_THRESHOLD")
	if thresholdStr != "" {
		threshold, _ = strconv.Atoi(thresholdStr)
	}

	dockerUser := os.Getenv("DOCKER_USER")
	dockerPassword := os.Getenv("DOCKER_PASSWORD")

	image, err := docker.NewImage(os.Args[1], dockerUser, dockerPassword)
	if err != nil {
		fmt.Printf("Can't parse qname: %s", err)
		os.Exit(1)
	}

	err = image.Pull()
	if err != nil {
		fmt.Printf("Can't pull image: %s", err)
		os.Exit(1)
	}
	if len(image.FsLayers) == 0 {
		fmt.Printf("Can't pull fsLayers")
		os.Exit(1)
	} else {
		fmt.Printf("Analysing %d layers\n", len(image.FsLayers))
	}

	c := clair.NewClair(clairAddr)
	vs := c.Analyse(image)
	groupBySeverity(vs)
	fmt.Printf("Found %d vulnerabilities \n", len(vs))
	highSevNumber := len(store["High"]) + len(store["Critical"]) + len(store["Defcon1"])

	iteratePriorities(func(sev string) {
		for _, v := range store[sev] {
			fmt.Printf("%s: [%s] \n%s\n%s\n", v.Name, v.Severity, v.Description, v.Link)
			fmt.Println("-----------------------------------------")
		}
	})
	iteratePriorities(func(sev string) { fmt.Printf("%s: %d\n", sev, len(store[sev])) })

	if highSevNumber > threshold {
		os.Exit(1)
	}
}

func iteratePriorities(f func(sev string)) {
	for _, sev := range priorities {
		if len(store[sev]) != 0 {
			f(sev)
		}
	}

}

func groupBySeverity(vs []clair.Vulnerability) {
	for _, v := range vs {
		sevRow := vulnsBy(v.Severity, store)
		store[v.Severity] = append(sevRow, v)
	}
}

func vulnsBy(sev string, store map[string][]clair.Vulnerability) []clair.Vulnerability {
	items, found := store[sev]
	if !found {
		items = make([]clair.Vulnerability, 0)
		store[sev] = items
	}
	return items
}
