/*
Copyright 2017 The Kubernetes Authors.

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

package gce

import (
	"context"
	"errors"
	"fmt"

	"github.com/golang/glog"
	container "google.golang.org/api/container/v1"
)

func newClustersMetricContext(request, zone string) *metricContext {
	return newGenericMetricContext("clusters", request, unusedMetricLabel, zone, computeV1Version)
}

func (gce *GCECloud) ListClusters(ctx context.Context) ([]string, error) {
	allClusters := []string{}

	for _, zone := range gce.managedZones {
		clusters, err := gce.listClustersInZone(zone)
		if err != nil {
			return nil, err
		}
		// TODO: Scoping?  Do we need to qualify the cluster name?
		allClusters = append(allClusters, clusters...)
	}

	return allClusters, nil
}

func (gce *GCECloud) GetManagedClusters(ctx context.Context) ([]*container.Cluster, error) {
	managedClusters := []*container.Cluster{}

	if gce.regional {
		var err error
		managedClusters, err = gce.getClustersInLocation(gce.region)
		if err != nil {
			return nil, err
		}
	} else if len(gce.managedZones) >= 1 {
		for _, zone := range gce.managedZones {
			clusters, err := gce.getClustersInLocation(zone)
			if err != nil {
				return nil, err
			}
			managedClusters = append(managedClusters, clusters...)
		}
	} else {
		return nil, errors.New(fmt.Sprintf("no zones associated with this cluster(%s)", gce.ProjectID()))
	}

	return managedClusters, nil
}

func (gce *GCECloud) Master(ctx context.Context, clusterName string) (string, error) {
	return "k8s-" + clusterName + "-master.internal", nil
}

func (gce *GCECloud) listClustersInZone(zone string) ([]string, error) {
	clusters, err := gce.getClustersInLocation(zone)
	if err != nil {
		return nil, err
	}

	result := []string{}
	for _, cluster := range clusters {
		result = append(result, cluster.Name)
	}
	return result, nil
}

func (gce *GCECloud) getClustersInLocation(zoneOrRegion string) ([]*container.Cluster, error) {
	// TODO: Issue/68913 migrate metric to list_location instead of list_zone.
	mc := newClustersMetricContext("list_zone", zoneOrRegion)
	// TODO: use PageToken to list all not just the first 500
	location := getLocationName(gce.projectID, zoneOrRegion)
	list, err := gce.containerService.Projects.Locations.Clusters.List(location).Do()
	if err != nil {
		return nil, mc.Observe(err)
	}
	if list.Header.Get("nextPageToken") != "" {
		glog.Errorf("Failed to get all clusters for request, received next page token %s", list.Header.Get("nextPageToken"))
	}

	return list.Clusters, mc.Observe(nil)
}
