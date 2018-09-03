package harbor

type harborCreateProjectReq struct {
	Name     string            `json:"project_name"`
	Metadata map[string]string `json:"metadata"`
}

type HarborCreateTargetReq struct {
	URL      string `json:"endpoint"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Password string `json:"password"`
	Type     int    `json:"type"`
	Insecure bool   `json:"insecure"`
}

type HarborUpdateTargetReq struct {
	Name     string `json:"name"`
	Endpoint string `json:"endpoint"`
	Username string `json:"username"`
	Password string `json:"password"`
	Insecure bool   `json:"insecure"`
}

type HarborCreateRepPolicyReq struct {
	Name                      string             `json:"name"`
	Description               string             `json:"description"`
	Filters                   []HarborFilter     `json:"filters"`
	ReplicateDeletion         bool               `json:"replicate_deletion"`
	Trigger                   *HarborTrigger     `json:"trigger"`
	Projects                  []*HarborProject   `json:"projects"`
	Targets                   []*HarborRepTarget `json:"targets"`
	ReplicateExistingImageNow bool               `json:"replicate_existing_image_now"`
}

type HarborUpdateRepPolicyReq HarborCreateRepPolicyReq

type HarborStopJobsReq struct {
	PolicyID int64  `json:"policy_id"`
	Status   string `json:"status"`
}
