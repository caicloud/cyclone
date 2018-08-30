package resource

import (
	"context"

	. "github.com/caicloud/cargo-admin/pkg/errors"
	"github.com/caicloud/cargo-admin/pkg/harbor"
	"github.com/caicloud/cargo-admin/pkg/models"

	"github.com/caicloud/nirvana/log"
)

func createOneTargetToAllRegistries(registry, host, username, password string) error {
	registries, err := models.Registry.FindAll()
	if err != nil {
		return err
	}
	for _, r := range registries {
		c, err := harbor.ClientMgr.GetClient(r.Name)
		if err != nil {
			return err
		}
		_, err = c.CreateTarget(registry, host, username, password)
		if err != nil {
			log.Errorf("create target into registry: %s, error %v", r.Name, err)
			return ErrorUnknownInternal.Error(err)
		}
	}
	return nil
}

func createAllTargetsToOneRegistry(host, username, password string) error {
	cli, err := harbor.NewClient(host, username, password)
	if err != nil {
		return err
	}
	registries, err := models.Registry.FindAll()
	if err != nil {
		return err
	}

	for _, r := range registries {
		_, err := cli.CreateTarget(r.Name, r.Host, r.Username, r.Password)
		if err != nil {
			log.Errorf("create target into registry: %s, error %v", r.Name, err)
			return err
		}
	}

	return nil
}

func deleteAllHarborTargets(ctx context.Context, registry string) error {
	cli, err := harbor.ClientMgr.GetClient(registry)
	if err != nil {
		return err
	}

	replications, err := cli.ListReplicationPolicies()
	if err != nil {
		return ErrorUnknownInternal.Error(err)
	}
	if len(replications) != 0 {
		log.Errorf("there are %d replication policies in registry: %s", len(replications), registry)
		return ErrorUnknownRequest.Error("please delete all replications first")
	}

	targets, err := cli.ListTargets()
	if err != nil {
		log.Errorf("list targets from registry: %s error: %v", registry, err)
		return ErrorUnknownInternal.Error(err)
	}

	for _, t := range targets {
		err = cli.DeleteTarget(t.ID)
		if err != nil {
			log.Errorf("delete target: %d from registry: %s error: %v", t.ID, registry)
		}
	}
	return nil
}

func deleteOneTargetFromAllRegistries(registry string) error {
	registries, err := models.Registry.FindAll()
	if err != nil {
		return err
	}
	for _, r := range registries {
		if r.Name == registry {
			continue
		}
		c, err := harbor.ClientMgr.GetClient(r.Name)
		if err != nil {
			return err
		}
		targets, err := c.ListTargets()
		if err != nil {
			log.Errorf("list all targets from registry: %s error: %v", r.Name, err)
			return ErrorUnknownInternal.Error(err)
		}
		for _, t := range targets {
			if t.Name == registry {
				log.Infof("delete target: %s from registry: %s, error: %v", t.Name, r.Name, c.DeleteTarget(t.ID))
			}
		}
	}
	return nil
}
