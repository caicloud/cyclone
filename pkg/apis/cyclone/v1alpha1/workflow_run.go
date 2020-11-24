package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WorkflowRun describes one workflow run, giving concrete runtime parameters and
// recording workflow run status.
type WorkflowRun struct {
	// Metadata for the resource, like kind and apiversion
	metav1.TypeMeta `json:",inline"`
	// Metadata for the particular object, including name, namespace, labels, etc
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Workflow run specification
	Spec WorkflowRunSpec `json:"spec"`
	// Status of workflow execution
	Status WorkflowRunStatus `json:"status,omitempty"`
}

// WorkflowRunSpec defines workflow run specification.
type WorkflowRunSpec struct {
	// Reference to a Workflow
	WorkflowRef *corev1.ObjectReference `json:"workflowRef"`
	// Stages in the workflow to start execution
	StartStages []string `json:"startStages"`
	// Stages in the workflow to end execution
	EndStages []string `json:"endStages"`
	// Maximum time this workflow can run
	Timeout string `json:"timeout"`
	// ServiceAccount used in the workflow execution
	ServiceAccount string `json:"serviceAccount"`
	// ResourceParams represents parameters for resources
	ResourceParams []ParameterConfig `json:"resources"`
	// StageParams represents parameters for stages
	StageParams []ParameterConfig `json:"stages"`
	// Execution context which specifies namespace and PVC used
	ExecutionContext *ExecutionContext `json:"executionContext"`
	// PresetVolumes volumes are preset volumes that will be mounted to all stage pods. For the moment, two kinds
	// of volumes supported, namely HostPath, PV. Users can make use of preset volumes to inject timezone, certificates
	// from host to containers, or mount data from PV to be used in containers.
	PresetVolumes []PresetVolume `json:"volumes,omitempty"`
	// GlobalVariables are global variables that can be used in all stages within the workflow in a run. Different to
	// global values defined in workflow, values defined here MUST be concrete value (string only for the moment). We
	// can regard global variables in workflow as value templates (for example, $(random:5)), and final values must be
	// populated in workflow run.
	// Values defined here will override those defined in the workflow. If a variable is defined in workflow but not here,
	// it would be populated (final value generated) here when workflowrun created by workflowrun controller.
	GlobalVariables []GlobalVariable `json:"globalVariables,omitempty"`
}

// PresetVolume defines a preset volume
type PresetVolume struct {
	// Type of the volume
	Type PresetVolumeType `json:"type"`
	// Path is path in host, PVC, or key path in Secret, ConfigMap.
	Path string `json:"path"`
	// MountPath is path in container that this preset volume will be mounted.
	MountPath string `json:"mountPath"`
	// SubPath is subpath to mount in container, for example, if MountPath is '/etc', SubPath is 'app.config', then
	// final file would be '/etc/app.config' in container
	SubPath string `json:"subPath"`
	// ObjectName is name of the source object if preset volume type is Secret, ConfigMap, PVC
	ObjectName *string `json:"objectName,omitempty"`
	// Scope defines the which containers to apply the preset volumes
	ContainerGroup ContainerGroup `json:"containerGroup"`
}

// PresetVolumeType is type of preset volumes, HostPath, PV supported.
type PresetVolumeType string

const (
	// PresetVolumeTypeHostPath ...
	PresetVolumeTypeHostPath PresetVolumeType = "HostPath"
	// PresetVolumeTypePVC ...
	PresetVolumeTypePVC PresetVolumeType = "PVC"
	// PresetVolumeTypeSecret ...
	PresetVolumeTypeSecret PresetVolumeType = "Secret"
	// PresetVolumeTypeConfigMap ...
	PresetVolumeTypeConfigMap PresetVolumeType = "ConfigMap"
)

// ContainerGroup defines group of containers in a stage pod, for example, 'sidecar', 'workload', 'initContainer'
type ContainerGroup string

const (
	// ContainerGroupAll represents all containers in a pod
	ContainerGroupAll ContainerGroup = "All"
	// ContainerGroupSidecar represents sidecar containers in a pod
	ContainerGroupSidecar ContainerGroup = "Sidecar"
	// ContainerGroupWorkload repressents user containers in a pod
	ContainerGroupWorkload ContainerGroup = "Workload"
)

// ParameterConfig configures parameters of a resource or a stage.
type ParameterConfig struct {
	// Whose parameters to configure
	Name string `json:"name"`
	// Parameters ...
	Parameters []ParameterItem `json:"parameters"`
}

// ExecutionContext is execution context of a workflow. Namespace, pvc
// cluster info would be defined here.
type ExecutionContext struct {
	// Name of the execution cluster
	Cluster string `json:"cluster"`
	// Namespace is namespace where to run workflow
	Namespace string `json:"namespace"`
	// PVC is the PVC used to run workflow
	PVC string `json:"pvc"`
}

// WorkflowRunStatus records workflow running status.
type WorkflowRunStatus struct {
	// Status of all stages
	Stages map[string]*StageStatus `json:"stages"`
	// Overall status
	Overall Status `json:"overall"`
	// Whether gc is performed on this WorkflowRun, such as deleting pods.
	Cleaned bool `json:"cleaned"`
	// Notifications represents the status of sending notifications.
	Notifications map[string]NotificationStatus `json:"notifications,omitempty"`
}

// StageStatus describes status of a stage execution.
type StageStatus struct {
	// Information of the pod
	Pod *PodInfo `json:"pod"`
	// Conditions of a stage
	Status Status `json:"status"`
	// Key-value outputs of this stage
	Outputs []KeyValue `json:"outputs"`
	// Stages that this stage depends on.
	Depends []string `json:"depends"`
	// Trivial indicates whether this stage is critical in the workflow. If set to true, it means the workflow
	// can tolerate failure of this stage. In this case, all other stages can continue to execute and the overall
	// status of the workflow execution can still be succeed.
	Trivial bool `json:"trivial"`
	// Events of the stage
	Events []StageEvent `json:"events"`
}

// StageEvent describes pod warning events for a stage
type StageEvent struct {
	// Event name
	Name string `json:"name"`
	// This should be a short, machine understandable string that gives the reason
	// for the transition into the object's current status.
	Reason string `json:"reason,omitempty"`
	// A human-readable description of the status of this operation.
	Message string `json:"message,omitempty"`
	// The time at which the most recent occurrence of this event was recorded.
	LastTimestamp metav1.Time `json:"lastTimestamp,omitempty"`
	// The number of times this event has occurred.
	Count int32 `json:"count,omitempty"`
}

// StatusPhase represents the phase of stage status or workflowrun status.
type StatusPhase string

const (
	// StatusPending means stage is not executed yet when used for stage. When
	// used for WorkflowRun overall status, it means no stages in WorkflowRun
	// are started to execute.
	StatusPending StatusPhase = "Pending"
	// StatusRunning means Stage or WorkflowRun is running.
	StatusRunning StatusPhase = "Running"
	// StatusWaiting means Stage or WorkflowRun have finished, but need to wait
	// for external events to continue. For example, a stage's executing result
	// needs approval of users, so that following stages can proceed.
	StatusWaiting StatusPhase = "Waiting"
	// StatusSucceeded means Stage or WorkflowRun gotten completed without errors.
	StatusSucceeded StatusPhase = "Succeeded"
	// StatusFailed indicates something wrong in the execution of Stage or WorkflowRun.
	StatusFailed StatusPhase = "Failed"
	// StatusCancelled indicates WorkflowRun have been cancelled.
	StatusCancelled StatusPhase = "Cancelled"
)

// PodInfo describes the pod a stage created.
type PodInfo struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// Status of a Stage in a WorkflowRun or the whole WorkflowRun.
// +k8s:deepcopy-gen=true
type Status struct {
	// Phase with value: Running, Waiting, Completed, Error
	Phase StatusPhase `json:"phase"`

	// LastTransitionTime is the last time the status transitioned from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`

	// The reason for the status's last transition.
	// +optional
	Reason string `json:"reason,omitempty"`

	// A human readable message indicating details about the transition.
	// +optional
	Message string `json:"message,omitempty"`

	// StartTime is the start time of processing stage/workflowrun
	StartTime metav1.Time `json:"startTime,omitempty"`
}

// NotificationResult represents the result of sending notifications.
type NotificationResult string

const (
	// NotificationResultSucceeded means success result of sending notifications.
	NotificationResultSucceeded NotificationResult = "Succeeded"

	// NotificationResultFailed means failure result of sending notifications.
	NotificationResultFailed NotificationResult = "Failed"
)

// NotificationStatus represents the status of sending notifications.
type NotificationStatus struct {
	// Result represents the result of sending notifications.
	Result NotificationResult `json:"result"`
	// Message represents the detailed message for result.
	Message string `json:"message"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WorkflowRunList describes an array of WorkflowRun instances.
type WorkflowRunList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []WorkflowRun `json:"items"`
}
