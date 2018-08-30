package resource

import (
	"github.com/caicloud/cargo-admin/pkg/api/admin/types"
	"github.com/davecgh/go-spew/spew"
	"reflect"
	"testing"
)

const data = `
- group: Tomcat
  dockerfiles:
    - image: tomcat:7.0-alpine
      content: |
        FROM tomcat:7.0-alpine
    - image: tomcat:8.0-alpine
      content: |
        FROM tomcat:8.0-alpine
    - image: tomcat:9.0-alpine
      content: |
        FROM tomcat:9.0-alpine

        # openjdk:9-jre-alpine
        # tomcat: 9.0.10

        # Please note that you war file name would become your application context. For example, if your
        # war file name is 'app.war', then you should access your web pages as 127.0.0.1:8080/app
        ADD {{ war_file_name }} ${CATALINA_HOME}/webapps

        EXPOSE 8080

        WORKDIR $CATALINA_HOME
- group: Node
  dockerfiles:
    - image: node:9-alpine
      content: |
        FROM node:9-alpine
`

func TestGetDockerfiles(t *testing.T) {
	groups, err := getDockerfiles([]byte(data))
	if err != nil {
		t.Errorf("get dockerfiles error")
	}

	expected := []*types.DockerfileGroup{
		{
			Group: "Tomcat",
			Dockerfiles: []*types.Dockerfile{
				{
					Image:   "tomcat:7.0-alpine",
					Content: "FROM tomcat:7.0-alpine\n",
				},
				{
					Image:   "tomcat:8.0-alpine",
					Content: "FROM tomcat:8.0-alpine\n",
				},
				{
					Image:   "tomcat:9.0-alpine",
					Content: "FROM tomcat:9.0-alpine\n\n# openjdk:9-jre-alpine\n# tomcat: 9.0.10\n\n# Please note that you war file name would become your application context. For example, if your\n# war file name is 'app.war', then you should access your web pages as 127.0.0.1:8080/app\nADD {{ war_file_name }} ${CATALINA_HOME}/webapps\n\nEXPOSE 8080\n\nWORKDIR $CATALINA_HOME\n",
				},
			},
		},
		{
			Group: "Node",
			Dockerfiles: []*types.Dockerfile{
				{
					Image:   "node:9-alpine",
					Content: "FROM node:9-alpine\n",
				},
			},
		},
	}
	if !reflect.DeepEqual(expected, groups) {
		t.Errorf("expected %s, but got %s", spew.Sdump(expected), spew.Sdump(groups))
	}
}
