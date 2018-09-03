package types

// Status represents cargo-admin current status
type HealthCheckResult struct {
	Mongo string
	Cargo string
}
