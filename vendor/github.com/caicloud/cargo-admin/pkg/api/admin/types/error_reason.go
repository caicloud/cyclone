package types

const (
	Unknown         = "unknown"
	BadUrl          = "cargo:BadUrl"
	NoIPAllowed     = "cargo:IpNotAllowed"
	BadScheme       = "cargo:BadScheme"
	RegistryExisted = "cargo:RegistryExisted"
	AccessFailed    = "cargo:AccessRegistryFailed"
	UsedAlready     = "cargo:UsedAlready"

	RegistryNotExist = "cargo:RegistryNotFound"
	ProjectNotExist  = "cargo:ProjectNotExit"
	ProjectProtected = "cargo:ProjectProtected"
	BadImageFile     = "cargo:BadImageFile"
)
