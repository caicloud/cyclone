package common

const (
	// AdminTenant is name of the system admin tenant, it's a default tenant created when Cyclone
	// start, and resources shared among all tenants would be placed in this tenant, such as stage
	// templates.
	AdminTenant = "admin"

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

	// ControlClusterName is the name of control cluster
	ControlClusterName = "control-cluster"

	// CachePrefixPath is the prefix path of acceleration caches
	CachePrefixPath = "caches"
)
