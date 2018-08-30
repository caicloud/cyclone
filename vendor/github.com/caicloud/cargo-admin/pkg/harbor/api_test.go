package harbor

import "testing"

func TestProjectsPath(t *testing.T) {
	cases := []struct {
		page     int
		pageSize int
		name     string
		public   string
		truth    string
	}{
		{1, 10, "", "", "/api/projects?page=1&page_size=10&name=&public="},
		{1, 10, "foobar", "false", "/api/projects?page=1&page_size=10&name=foobar&public=false"},
		{1, 10, "foo bar", "", "/api/projects?page=1&page_size=10&name=foo+bar&public="},
	}

	for _, c := range cases {
		path := ProjectsPath(c.page, c.pageSize, c.name, c.public)
		if c.truth != path {
			t.Errorf("Get projects path error, expected %s, but got %s", c.truth, path)
		}
	}
}

func TestProjectPath(t *testing.T) {
	path := ProjectPath(1)
	truth := "/api/projects/1"
	if path != truth {
		t.Errorf("Get project path error: expected %s, but got %s", truth, path)
	}
}

func TestRepoPath(t *testing.T) {
	path := RepoPath("foo", "bar")
	truth := "/api/repositories/foo/bar"
	if path != truth {
		t.Errorf("Get repo path error, expected %s, but got %s", truth, path)
	}
}

func TestTagsPath(t *testing.T) {
	path := TagsPath("foo", "bar")
	truth := "/api/repositories/foo/bar/tags"
	if path != truth {
		t.Errorf("Get tags path error, expected %s, but got %s", truth, path)
	}
}

func TestTagPath(t *testing.T) {
	path := TagPath("foo", "bar", "v1")
	truth := "/api/repositories/foo/bar/tags/v1"
	if path != truth {
		t.Errorf("Get tag path error, expected %s, but got %s", truth, path)
	}
}

func TestTagVulnerabilityPath(t *testing.T) {
	path := TagVulnerabilityPath("foo", "bar", "v1")
	truth := "/api/repositories/foo/bar/tags/v1/vulnerability/details"
	if path != truth {
		t.Errorf("Get tag vulnerability path error, expected %s, but got %s", truth, path)
	}
}

func TestLoginUrl(t *testing.T) {
	url := LoginUrl("foobar.com", "baz", "pwd")
	truth := "foobar.com/login?principal=baz&password=pwd"
	if url != truth {
		t.Errorf("Get login url error, expected %s, but got %s", truth, url)
	}
}

func TestLogsPath(t *testing.T) {
	path := LogsPath(100, 200, "pull")
	truth := "/api/logs?begin_timestamp=100&end_timestamp=200&operation=pull"
	if path != truth {
		t.Errorf("Get logs path error, expected %s, but got %s", truth, path)
	}
}

func TestProjectLogsPath(t *testing.T) {
	path := ProjectLogsPath(1, 100, 200, "pull")
	truth := "/api/projects/1/logs?begin_timestamp=100&end_timestamp=200&operation=pull"
	if path != truth {
		t.Errorf("Get projects logs path error, expected %s, but got %s", path, truth)
	}
}

func TestReposPath(t *testing.T) {
	cases := []struct {
		pid      int64
		query    string
		page     int
		pageSize int
		truth    string
	}{
		{1, "", 1, 10, "/api/repositories?project_id=1&q=&page=1&page_size=10"},
		{1, "foobar", 1, 10, "/api/repositories?project_id=1&q=foobar&page=1&page_size=10"},
		{1, "foo bar", 1, 10, "/api/repositories?project_id=1&q=foo+bar&page=1&page_size=10"},
	}

	for _, c := range cases {
		path := ReposPath(c.pid, c.query, c.page, c.pageSize)
		if c.truth != path {
			t.Errorf("Get repos path error, expected %s, but got %s", c.truth, path)
		}
	}
}
