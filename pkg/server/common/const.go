package common

const (
	// DefaultTenant is the name of the cyclone default tenant, it's a default tenant created
	// when Cyclone start.
	DefaultTenant = "system"

	// TenantNamespacePrefix is the prefix of namespace which related to a specific tenant
	TenantNamespacePrefix = "cyclone-"

	// TenantPVCPrefix is the prefix of pvc which related to a specific tenant
	TenantPVCPrefix = "cyclone-pvc-"

	// DefaultPVCSize is the default size of pvc
	DefaultPVCSize = "5Gi"

	// SceneCICD is the label value used to indicate cyclone CI/CD scenario
	SceneCICD = "cicd"

	// SceneAI is the label value used to indicate cyclone AI scenario
	SceneAI = "ai"

	// CronTimerTrigger represents the trigger of workflowruns triggered by cron timer.
	CronTimerTrigger = "cron-timer"

	// QuotaCPULimit represents default value of 'limits.cpu'
	QuotaCPULimit = "2"
	// QuotaCPURequest represents default value of 'requests.cpu'
	QuotaCPURequest = "1"
	// QuotaMemoryLimit represents default value of 'limits.memory'
	QuotaMemoryLimit = "4Gi"
	// QuotaMemoryRequest represents default value of 'requests.memory'
	QuotaMemoryRequest = "2Gi"

	// SecretKeyIntegration is the key of the secret dada to indicate its value is about integration information.
	SecretKeyIntegration = "integration"
	// SecretKeyRepos is the key of the secret dada to records webhooks created for workflowtriggers.
	// Only for SCM integration secrets.
	SecretKeyRepos = "repos"

	// CachePrefixPath is the prefix path of acceleration caches
	CachePrefixPath = "caches"
)
