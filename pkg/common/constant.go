package common

const (
	// CloneDir represents the dir which the repo clone to.
	CloneDir = "/tmp/code"

	// GoTestReport represents the file name of golang test report.
	// If user configured the golang test report in unit-test commands (eg: go test -coverprofile=coverage.out),
	// We will cp the file(coverage.out) to go_test_report.cyclone, and use it at code-scan stage.
	GoTestReport = "go_test_report.cyclone"
)
