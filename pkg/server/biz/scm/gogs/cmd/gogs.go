package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/caicloud/cyclone/pkg/server/apis/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/biz/scm"
	"github.com/caicloud/cyclone/pkg/server/biz/scm/gogs"
)

func main() {
	var server, username, password string
	flag.StringVar(&server, "server", "", "Gogs's server address")
	flag.StringVar(&username, "username", "", "Gogs's username")
	flag.StringVar(&password, "password", "", "Gogs's password")
	flag.Parse()

	var config = &v1alpha1.SCMSource{
		Type:     "Gogs",
		Server:   server,
		User:     username,
		Password: password,
		AuthType: v1alpha1.AuthTypePassword,
	}

	var err error

	var provider scm.Provider

	if provider, err = gogs.NewGogs(config); err != nil {
		log.Fatal(err)
	}

	var repos []scm.Repository
	if repos, err = provider.ListRepos(); err != nil {
		log.Fatal(err)
	}
	for _, r := range repos {
		fmt.Printf("Repositority %s tags:\n", r.Name)
		var tags []string
		if tags, err = provider.ListTags(r.Name); err != nil {
			log.Fatal(err)
		}
		for _, t := range tags {
			fmt.Printf("\t%s\n", t)
		}
	}
}
