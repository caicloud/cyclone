package gitlab

type MergeCommentEvent struct {
	ObjectKind string `json:"object_kind"`
	//User       *User  `json:"user"`
	ProjectID int `json:"project_id"`
	Project   struct {
		Name              string `json:"name"`
		Description       string `json:"description"`
		AvatarURL         string `json:"avatar_url"`
		GitSSHURL         string `json:"git_ssh_url"`
		GitHTTPURL        string `json:"git_http_url"`
		Namespace         string `json:"namespace"`
		PathWithNamespace string `json:"path_with_namespace"`
		DefaultBranch     string `json:"default_branch"`
		Homepage          string `json:"homepage"`
		URL               string `json:"url"`
		SSHURL            string `json:"ssh_url"`
		HTTPURL           string `json:"http_url"`
		WebURL            string `json:"web_url"`
		//VisibilityLevel   VisibilityLevelValue `json:"visibility_level"`
	} `json:"project"`
	//Repository       *Repository `json:"repository"`
	ObjectAttributes struct {
		ID           int    `json:"id"`
		Note         string `json:"note"`
		NoteableType string `json:"noteable_type"`
		AuthorID     int    `json:"author_id"`
		CreatedAt    string `json:"created_at"`
		UpdatedAt    string `json:"updated_at"`
		ProjectID    int    `json:"project_id"`
		Attachment   string `json:"attachment"`
		LineCode     string `json:"line_code"`
		CommitID     string `json:"commit_id"`
		NoteableID   int    `json:"noteable_id"`
		System       bool   `json:"system"`
		//StDiff       *Diff  `json:"st_diff"`
		URL string `json:"url"`
	} `json:"object_attributes"`
	MergeRequest *MergeRequest `json:"merge_request"`
}

type MergeRequest struct {
	ID             int    `json:"id"`
	IID            int    `json:"iid"`
	ProjectID      int    `json:"project_id"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	WorkInProgress bool   `json:"work_in_progress"`
	State          string `json:"state"`
	//CreatedAt      *time.Time `json:"created_at"`
	//UpdatedAt      *time.Time `json:"updated_at"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
	TargetBranch string `json:"target_branch"`
	SourceBranch string `json:"source_branch"`
	Upvotes      int    `json:"upvotes"`
	Downvotes    int    `json:"downvotes"`
	Author       struct {
		Name      string `json:"name"`
		Username  string `json:"username"`
		ID        int    `json:"id"`
		State     string `json:"state"`
		AvatarURL string `json:"avatar_url"`
	} `json:"author"`
	Assignee struct {
		Name      string `json:"name"`
		Username  string `json:"username"`
		ID        int    `json:"id"`
		State     string `json:"state"`
		AvatarURL string `json:"avatar_url"`
	} `json:"assignee"`
	SourceProjectID int      `json:"source_project_id"`
	TargetProjectID int      `json:"target_project_id"`
	Labels          []string `json:"labels"`
	Milestone       struct {
		ID          int    `json:"id"`
		Iid         int    `json:"iid"`
		ProjectID   int    `json:"project_id"`
		Title       string `json:"title"`
		Description string `json:"description"`
		State       string `json:"state"`
		//CreatedAt   *time.Time `json:"created_at"`
		//UpdatedAt   *time.Time `json:"updated_at"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
		DueDate   string `json:"due_date"`
	} `json:"milestone"`
	MergeWhenBuildSucceeds  bool   `json:"merge_when_build_succeeds"`
	MergeStatus             string `json:"merge_status"`
	Subscribed              bool   `json:"subscribed"`
	UserNotesCount          int    `json:"user_notes_count"`
	SouldRemoveSourceBranch bool   `json:"should_remove_source_branch"`
	ForceRemoveSourceBranch bool   `json:"force_remove_source_branch"`
	Changes                 []struct {
		OldPath     string `json:"old_path"`
		NewPath     string `json:"new_path"`
		AMode       string `json:"a_mode"`
		BMode       string `json:"b_mode"`
		Diff        string `json:"diff"`
		NewFile     bool   `json:"new_file"`
		RenamedFile bool   `json:"renamed_file"`
		DeletedFile bool   `json:"deleted_file"`
	} `json:"changes"`
	WebURL     string `json:"web_url"`
	LastCommit struct {
		ID      string `json:"id"`
		Message string `json:"message"`
	} `json:"last_commit"`
}
