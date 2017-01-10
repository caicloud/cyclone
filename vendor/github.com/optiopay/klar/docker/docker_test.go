package docker

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewImage(t *testing.T) {
	tcs := map[string]struct {
		image    string
		registry string
		name     string
		tag      string
	}{
		"full": {
			image:    "docker-registry.domain.com:8080/nginx:1b29e1531c",
			registry: "https://docker-registry.domain.com:8080/v2",
			name:     "nginx",
			tag:      "1b29e1531c",
		},
		"regular": {
			image:    "docker-registry.domain.com/nginx:1b29e1531c",
			registry: "https://docker-registry.domain.com/v2",
			name:     "nginx",
			tag:      "1b29e1531c",
		},
		"no_tag": {
			image:    "docker-registry.domain.com/nginx",
			registry: "https://docker-registry.domain.com/v2",
			name:     "nginx",
			tag:      "latest",
		},
		"no_tag_with_port": {
			image:    "docker-registry.domain.com:8080/nginx",
			registry: "https://docker-registry.domain.com:8080/v2",
			name:     "nginx",
			tag:      "latest",
		},

		"no_registry": {
			image:    "skynetservices/skydns:2.3",
			registry: "https://registry-1.docker.io/v2",
			name:     "skynetservices/skydns",
			tag:      "2.3",
		},
		"no_registry_root": {
			image:    "postgres:9.5.1",
			registry: "https://registry-1.docker.io/v2",
			name:     "library/postgres",
			tag:      "9.5.1",
		},
	}
	for name, tc := range tcs {
		image, err := NewImage(tc.image, "", "")
		if err != nil {
			t.Fatalf("%s: Can't parse image name: %s", name, err)
		}
		if image.Registry != tc.registry {
			t.Fatalf("%s: Expected registry name %s, got %s", name, tc.registry, image.Registry)
		}
		if image.Name != tc.name {
			t.Fatalf("%s: Expected image name %s, got %s", name, tc.name, image.Name)
		}
		if image.Tag != tc.tag {
			t.Fatalf("%s: Expected image tag %s, got %s", name, tc.tag, image.Tag)
		}
	}

}

func TestPull(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp, err := ioutil.ReadFile("testdata/registry-response.json")
		if err != nil {
			t.Fatalf("Can't load registry test response %s", err.Error())
		}
		fmt.Fprintln(w, string(resp))
	}))
	defer ts.Close()

	image, err := NewImage("docker-registry.domain.com/nginx:1b29e1531c", "", "")
	image.Registry = ts.URL
	err = image.Pull()
	if err != nil {
		t.Fatalf("Can't pull image: %s", err)
	}
	if len(image.FsLayers) == 0 {
		t.Fatal("Can't pull fsLayers")
	}
}
