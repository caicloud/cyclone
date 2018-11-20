package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
)

// basic
type ObjectMetaData struct {
	ID                 string `json:"id,omitempty"`
	Name               string `json:"name"`
	CreationTime       string `json:"creationTime"`
	LastUpdateTime     string `json:"lastUpdateTime"`
	LastTransitionTime string `json:"lastTransitionTime,omitempty"` // TODO 是否需要

	Alias       string `json:"alias,omitempty"`
	Description string `json:"description,omitempty"`
}
type ListMetaData struct {
	Total int `json:"total"`
}

// type
type StorageTypeObject struct {
	Metadata ObjectMetaData `json:"metadata"`
	// self
	Provisioner string `json:"provisioner"`
	// var
	CommonParameters   map[string]string `json:"commonParameters"`
	OptionalParameters map[string]string `json:"optionalParameters"`
}
type ListStorageTypeResponse struct {
	MetaData ListMetaData        `json:"metadata"`
	Items    []StorageTypeObject `json:"items"`
}
type StorageTypeInfoObject struct {
	Name string `json:"name"`
	// self
	Provisioner string `json:"provisioner"`
	// var
	OptionalParameters map[string]string `json:"optionalParameters"`
}

// service
type StorageServiceObject struct {
	Metadata ObjectMetaData `json:"metadata"`
	// upper
	Type StorageTypeInfoObject `json:"type"`
	// var
	Parameters map[string]string `json:"parameters"`
}
type ListStorageServiceResponse struct {
	MetaData ListMetaData           `json:"metadata"`
	Items    []StorageServiceObject `json:"items"`
}
type CreateStorageServiceRequest struct {
	// self
	Name        string `json:"name"`
	Alias       string `json:"alias,omitempty"`
	Description string `json:"description,omitempty"`
	// upper
	Type string `json:"type"`
	// var
	Parameters map[string]string `json:"parameters"`
}
type CreateStorageServiceResponse = StorageServiceObject
type ReadStorageServiceResponse = StorageServiceObject
type UpdateStorageServiceRequest struct {
	Alias       string `json:"alias,omitempty"`
	Description string `json:"description,omitempty"`
}
type UpdateStorageServiceResponse = StorageServiceObject

// class
type StorageClassStatus string

const (
	StorageClassPending     StorageClassStatus = "Pending"
	StorageClassActive      StorageClassStatus = "Active"
	StorageClassTerminating StorageClassStatus = "Terminating"
)

type StorageClassObject struct {
	storagev1.StorageClass
	Status StorageClassStatus `json:"status"` // TODO remove status in api
}
type ListStorageClassResponse struct {
	MetaData ListMetaData         `json:"metadata"`
	Items    []StorageClassObject `json:"items"`
}

type CreateStorageClassRequest struct { // cluster 信息在 path 中传入
	// self
	Name        string `json:"name"` // 该名称在 volume 创建有使用，用于关联，用户有指定
	Alias       string `json:"alias,omitempty"`
	Description string `json:"description,omitempty"`
	// upper
	Service string `json:"service"`
	// var
	Parameters map[string]string `json:"parameters"`
}
type CreateStorageClassResponse = StorageClassObject
type ReadStorageClassResponse = StorageClassObject
type UpdateStorageClassRequest struct { // 上级/位置信息在 path 中传入
	Alias       string `json:"alias,omitempty"`
	Description string `json:"description,omitempty"`
}
type UpdateStorageClassResponse = StorageClassObject

// volume // pvc
type VolumeAccessMode = corev1.PersistentVolumeAccessMode

const (
	ReadWriteOnce VolumeAccessMode = corev1.ReadWriteOnce
	ReadOnlyMany  VolumeAccessMode = corev1.ReadOnlyMany
	ReadWriteMany VolumeAccessMode = corev1.ReadWriteMany
)

type DataVolumeObject = corev1.PersistentVolumeClaim
type ListDataVolumeResponse struct {
	MetaData ListMetaData       `json:"metadata"`
	Items    []DataVolumeObject `json:"items"`
}
type CreateDataVolumeRequest struct {
	// self
	Name string `json:"name"` // 该名称在 volume 创建有使用，用于关联，用户有指定
	// upper
	StorageClass string `json:"storageClass"`
	// var
	AccessModes []VolumeAccessMode `json:"accessModes"`
	Size        int                `json:"size"`
}
type CreateDataVolumeResponse = DataVolumeObject
type ReadDataVolumeResponse = DataVolumeObject
