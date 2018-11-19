package common

// ContainerState represents container state.
type ContainerState string

const (
	// ContainerStateTerminated represents container is stopped.
	ContainerStateTerminated ContainerState = "Terminated"

	// ContainerStateInitialized represents container is Running or Stopped, not Init or Creating.
	ContainerStateInitialized ContainerState = "Initialized"
)
