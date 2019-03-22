package v1alpha1

// ResourceItem defines a resource
type ResourceItem struct {
	// Resource name
	Name string `json:"name"`
	// For input resource, this is the path that resource will be mounted in workload container.
	// Resolver would resolve resources and mount it in this path.
	// For output resource, this is the path in the workload container specify output data, for
	// the moment, only one path can be given. In the resolver container, data specified here would
	// be mounted in /workspace/data by default. Resolver will then push resource to remote server.
	// TODO(ChenDe): For output resource, need support multiple paths.
	Path string `json:"path"`
}

// ArtifactItem defines an artifact
type ArtifactItem struct {
	// Artifact name
	Name string `json:"name"`
	// Path of the artifact
	Path string `json:"path"`
	// Source of the artifact. When artifact is used as input, this is needed.
	// It's in the format of: <stage name>/<artifact name>
	// +Optional
	Source string `json:"source"`
}

// ParameterItem defines a parameter
type ParameterItem struct {
	// Name of the parameter
	Name string `json:"name"`
	// Value of the parameter
	Value string `json:"value"`
}

// ArgumentValue defines a argument value
type ArgumentValue ParameterItem

// KeyValue defines a key-value pair
type KeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// PresetVolumeConfig defines a preset volume
type PresetVolumeConfig struct {
	// Type of the volume
	Type PresetVolumeType `json:"type"`
	// VolumePath is the path of volume. with HostPath type volume, it is the path of the host,
	// with PV type volume, it is the path of the pv.
	VolumePath string `json:"volumePath"`
	// Path in the pod where the VolumePath will mount to
	Path string `json:"path"`
}

// PresetVolumeType is one type of stages preset mount volumes, there are some kinds of preset volume
// supported to mount to a workload pod: HostPath, PV.
type PresetVolumeType string

const (
	// PresetVolumeTypeHostPath ...
	PresetVolumeTypeHostPath PresetVolumeType = "HostPath"
	// PresetVolumeTypePV ...
	PresetVolumeTypePV PresetVolumeType = "PV"
)
