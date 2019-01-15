package v1alpha1

import (
	core_v1 "k8s.io/api/core/v1"
)

// Tenant contains information about tenant
type Tenant struct {
	// Metadata contains metadata information about tenant
	Metadata TenantMetadata `json:"metadata"`
	// Spec contains tenant spec
	Spec TenantSpec `json:"spec"`
}

// TenantMetadata contains metadata information about tenant
type TenantMetadata struct {
	// Name is the name of tenant, unique.
	Name string `json:"name"`
	// Description describes the tenant
	Description string `json:"description,omitempty"`
	// CreationTime records the time of tenant creation
	CreationTime string `json:"creationTime"`
}

// TenantSpec contains the tenant spec information
type TenantSpec struct {
	// PersistentVolumeClaim describes information about persistent volume claim
	PersistentVolumeClaim PersistentVolumeClaim `json:"persistentVolumeClaim"`

	// ResourceQuota describes the resource quota of the namespace,
	// eg map[string]string{"cpu": "2Gi", "memory": "512Mi"}
	ResourceQuota map[core_v1.ResourceName]string `json:"resourceQuota"`
}

// PersistentVolumeClaim describes information about pvc belongs to a tenant
type PersistentVolumeClaim struct {
	// Name is the pvc name specified by user, and if Name is not nil, cyclone will
	// use this pvc and not to create another one.
	Name string `json:"name"`

	// StorageClass represents the strorageclass used to create pvc
	StorageClass string `json:"storageclass"`

	// Size represents the capacity of the pvc, unit supports 'Gi' or 'Mi'
	// More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#capacity
	Size string `json:"size"`
}
