package constants

import (
	"fmt"
	"time"
)

const (
	APIVersion = "v1alpha1"
)

var (
	RootPath = fmt.Sprintf("/api/%s", APIVersion)
)

const (
	ParameterStart = "start"
	ParameterLimit = "limit"

	ParameterStorageType    = "type"
	ParameterStorageService = "service"
	ParameterStorageClass   = "storageclass"
	ParameterCluster        = "cluster"
	ParameterPartition      = "partition"
	ParameterVolume         = "volume"
	ParameterName           = "name"

	MimeJson          = "application/json"
	MimeText          = "text/plain"
	HeaderContentType = "Content-Type"

	TimeFormatObjectMeta = time.RFC3339

	GLogLogLevelFlagName     = "stderrthreshold"
	GLogLogVerbosityFlagName = "v"

	ControllerMaxRetryTimesUnlimited = 0 // continue retry

	DefaultControllerMaxRetryTimes      = ControllerMaxRetryTimesUnlimited
	DefaultControllerResyncPeriodSecond = 600
	DefaultWatchIntervalSecond          = 60
	DefaultListenPort                   = 2333
	DefaultLogLevel                     = "info"
	DefaultLogVerbosity                 = "0"

	ClassNameMaxLength = 21

	// storage quota label for all
	// LabelKeyStorageNumAll  = "persistentvolumeclaims"
	// LabelKeyStorageSizeAll = "requests.storage"

	// storage quota label by class
	FormatLabelKeyStorageQuotaNum  = "%s.storageclass.storage.k8s.io/persistentvolumeclaims"
	FormatLabelKeyStorageQuotaSize = "%s.storageclass.storage.k8s.io/requests.storage"

	// storage class annotations
	LabelKeyStorageType         = "storage.resource.caicloud.io/type"
	LabelKeyStorageService      = "storage.resource.caicloud.io/service"
	LabelKeyStorageAdminMarkKey = "storage.resource.caicloud.io/process"
	LabelKeyStorageAdminMarkVal = "true"
	LabelKeyIsSystemKey         = "storage.resource.caicloud.io/system"
	LabelKeyIsSystemVal         = "true"

	// annotations for all
	LabelKeyStorageAdminAlias       = "storage.resource.caicloud.io/alias"
	LabelKeyStorageAdminDescription = "storage.resource.caicloud.io/description"

	// finalizers
	StorageServiceControllerFinalizerName = "storage.resource.caicloud.io/storageservice-deletion"
	StorageClassControllerFinalizerName   = "storage.resource.caicloud.io/storageclass-deletion"

	// just for log
	ControlClusterName = "control-cluster"

	SystemStorageClass        = "heketi-storageclass"
	SystemNamespaceEmpty      = ""
	SystemNamespaceDefault    = "default"
	SystemNamespaceKubePublic = "kube-public"
	SystemNamespaceKubeSystem = "kube-system"
)

var (
	SystemNamespaces = map[string]struct{}{
		SystemNamespaceEmpty:      {},
		SystemNamespaceDefault:    {},
		SystemNamespaceKubePublic: {},
		SystemNamespaceKubeSystem: {},
	}
)
