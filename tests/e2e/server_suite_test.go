package cyclone_test

import (
	"github.com/caicloud/cyclone/api/server"
	"github.com/caicloud/cyclone/cloud"
	//"github.com/caicloud/cyclone/pkg/executil"
	"github.com/caicloud/cyclone/pkg/api"
	"github.com/caicloud/cyclone/pkg/server/manager"
	"github.com/caicloud/cyclone/pkg/server/router"
	"github.com/caicloud/cyclone/store"
	//dockerCli "github.com/docker/docker/client"
	restful "github.com/emicklei/go-restful"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	//"github.com/zoumo/logdog"

	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestServer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Server Suite")
}

const (
	//dockerHost       = "tcp://localhost:2375"
	mongoPara        = "mongo"
	mongoDefaultHost = "localhost:27017"
	//etcdDefaultHost  = "localhost:2379"
	//mongoImage       = "mongo:3.0.5"
	//etcdImage = "quay.io/coreos/etcd:v3.1.3"

	projectCollectionName        string = "projects"
	pipelineCollectionName       string = "pipelines"
	pipelineRecordCollectionName string = "pipelineRecords"
)

func TestManager(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Project Suite")
}

type ServerOptions struct {
	WorkerOptions    *cloud.WorkerOptions
	APIServerOptions *server.APIServerOptions
}

var (
	mongoHost             string
	cycloneServer         *http.Server
	port                  int
	dataStore             store.DataStore
	closing               chan struct{}
	mclosed               chan struct{}
	pipelineRecordManager manager.PipelineRecordManager
	pipelineManager       manager.PipelineManager
	projectManager        manager.ProjectManager
	//mongoID               []byte
	//etcdID                []byte
)

var _ = Describe("Project Suite", func() {
	BeforeSuite(func() {
		initEnv()
	})

	AfterSuite(func() {
		clearEnv()
	})

	Context("ProjectManager works correctly", func() {
		var mongo Mongo
		var cases TestCases
		BeforeEach(func() {
			mongo = NewMongoClient(mongoHost, []string{projectCollectionName})

			h := make([]*header, 0)
			h = append(h, &header{k: "Content-Type", v: "application/json"})
			cases = NewTestCases(h)
		})

		AfterEach(func() {
			mongo.Close()
			cases.Clear()
		})

		// Create project test.
		It("Given a projectManager, when creating new project, it should creat project correctly.", func() {

			// testCase1 Create project
			uri1 := "/projects"
			method1 := "POST"
			testReq1 := &api.Project{
				Name: "test1",
			}
			testRsp1 := &api.Project{
				Name: "test1",
			}
			cases.Add(uri1, method1, testReq1, testRsp1, &api.Project{}, http.StatusCreated)

			uri2 := "/projects"
			method2 := "POST"
			testReq2 := &api.Project{
				Name: "test2",
			}
			testRsp2 := &api.Project{
				Name: "test2",
			}
			cases.Add(uri2, method2, testReq2, testRsp2, &api.Project{}, http.StatusCreated)

			cases.Test()
		})

		// Update project test.
		It("Given a projectManager, when update new project, it should creat project correctly.", func() {
			baesProjectName := "base"
			baseProject := &api.Project{
				Name: "base",
			}
			mongo.Insert(projectCollectionName, baseProject)

			uri1 := fmt.Sprintf("/projects/%s", baesProjectName)
			method1 := "POST"
			testReq1 := &api.Project{
				Name: "test1",
			}
			testRsp1 := &api.Project{
				Name: "test1",
			}
			cases.Add(uri1, method1, testReq1, testRsp1, &api.Project{}, http.StatusCreated)

			uri2 := fmt.Sprintf("/projects/%s", baesProjectName)
			method2 := "POST"
			testReq2 := &api.Project{
				Name: "test2",
			}
			testRsp2 := &api.Project{
				Name: "test2",
			}
			cases.Add(uri2, method2, testReq2, testRsp2, &api.Project{}, http.StatusCreated)

			cases.Test()
		})
	})

	Context("PipelineManager works correctly", func() {
		var mongo Mongo
		var cases TestCases
		var baseProjectID string
		BeforeEach(func() {
			mongo = NewMongoClient(mongoHost, []string{projectCollectionName, pipelineCollectionName})

			baseProjectID = "testID"
			insertBaseProject(mongo, baseProjectID)

			h := make([]*header, 0)
			h = append(h, &header{k: "Content-Type", v: "application/json"})
			cases = NewTestCases(h)
		})

		AfterEach(func() {
			mongo.Close()
			cases.Clear()
		})

		It("Given a pipelineManager, when creating new pipeline, it should creat pipeline correctly.", func() {
			testReq1 := &api.Pipeline{
				Name:      "test1",
				ProjectID: "nonexist",
			}
			cases.Add("/projects/testProject/pipelines", "POST", testReq1, nil, nil, http.StatusNotFound)

			testReq2 := &api.Pipeline{
				Name:      "test2",
				ProjectID: baseProjectID,
			}
			testRsp2 := &api.Pipeline{
				Name:      "test2",
				ProjectID: baseProjectID,
			}
			cases.Add("/projects/testProject/pipelines", "POST", testReq2, testRsp2, &api.Pipeline{}, http.StatusBadRequest)

			testReq3 := &api.Pipeline{
				Name:      "test3",
				ProjectID: baseProjectID,
			}
			testRsp3 := &api.Pipeline{
				Name:      "test3",
				ProjectID: baseProjectID,
			}
			cases.Add("/projects/testProject/pipelines", "POST", testReq3, testRsp3, &api.Pipeline{}, http.StatusBadRequest)

			cases.Test()
		})
	})
})

func insertBaseProject(m Mongo, baseProjectID string) {
	project := &api.Project{
		Name: "baseProject",
		ID:   baseProjectID,
	}
	err := m.Insert(projectCollectionName, project)
	Expect(err).NotTo(HaveOccurred())
}

func initEnv() {
	//initDockerClient()
	//initMongo()
	//initEtcd()
	initParas()

	_, _, err := store.Init(mongoHost, time.Second, closing)
	Expect(err).NotTo(HaveOccurred())

	dataStore := store.NewStore()
	defer dataStore.Close()

	// Initialize the V1 API.
	err = router.InitRouters(dataStore)
	Expect(err).NotTo(HaveOccurred())

	// start server
	port = GetFreeTcpPort()
	cycloneServer := &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: restful.DefaultContainer}

	go cycloneServer.ListenAndServe()

	println("========           Starting cyclone-test-server ...           ========")
	time.Sleep(3 * time.Second)
}

func initParas() {
	if len(os.Args) > 1 {
		for _, arg := range os.Args[1:] {
			para := strings.Split(arg, "=")
			if len(para) != 2 {
				continue
			}

			switch para[0] {
			case mongoPara:
				mongoHost = para[1]
			default:
				continue
			}
		}
	} else {
		mongoHost = mongoDefaultHost
	}
	println(fmt.Sprintf("========          Set %s to mongoHost            ========", mongoHost))
}

func clearEnv() {
	cycloneServer.Close()
}

func GetFreeTcpPort() (port int) {
	listener, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		fmt.Println("Can not get free port:", err)
	} else {
		//fmt.Println("Get free port:", listener.Addr())
		addr := fmt.Sprint(listener.Addr())
		port, _ = strconv.Atoi(strings.Split(addr, ":")[3])
		listener.Close()
	}
	return
}
