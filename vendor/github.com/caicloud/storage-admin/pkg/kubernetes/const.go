package kubernetes

import (
	corev1 "k8s.io/api/core/v1"
)

const (
	// pvc annotations
	LabelKeyKubeStorageClass       = corev1.BetaStorageClassAnnotation
	LabelKeyKubeStorageProvisioner = "volume.beta.kubernetes.io/storage-provisioner"
)

type StorageClassProvisioner = string

const (
	StorageClassProvisionerGlusterfs StorageClassProvisioner = "kubernetes.io/glusterfs"
	StorageClassProvisionerCephRBD   StorageClassProvisioner = "kubernetes.io/rbd"
	StorageClassProvisionerAzureDisk StorageClassProvisioner = "kubernetes.io/azure-disk"
	StorageClassProvisionerNetappNAS StorageClassProvisioner = "netapp.io/trident"
	StorageClassProvisionerNFS       StorageClassProvisioner = "caicloud.io/nfs"

	StorageClassParamNameRestUser        = "restuser"
	StorageClassParamNameRestUserKey     = "restuserkey"
	StorageClassParamNameSecretName      = "secretName"
	StorageClassParamNameSecretNamespace = "secretNamespace"
	StorageClassParamNameGidMin          = "gidMin"
	StorageClassParamNameGidMax          = "gidMax"
	StorageClassParamGidRangeMin         = 2000
	StorageClassParamGidRangeMax         = 2147483647
)

const (
	HeaderK8SHost  = "K8S-Host"
	HeaderToken    = "K8S-Token"
	HeaderUsername = "K8S-Username"
	HeaderPassword = "K8S-Password"
)
