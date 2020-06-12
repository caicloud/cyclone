package leaderelection

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/caicloud/nirvana/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/server/healthz"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
)

// Option defines the parameters required to start the leader election component.
type Option struct {
	// LeaseLockName is the lease lock resource name, recommended to use the component name.
	LeaseLockName string
	// LeaseLockNamespace is the lease lock resource namespace, recommended to use the component namespace.
	LeaseLockNamespace string
	// ID is the the holder identity name, recommended to use the component pod name.
	// If not set, the value of the POD_NAME environment variable will be used
	// +optional
	ID string
	// KubeClient is the kube client of a cluster.
	KubeClient kubernetes.Interface
	// Run is the main controller code loop starter.
	Run func(ctx context.Context)
	// LivenessChecker defines the liveness healthz checker.
	// +optional
	LivenessChecker func(req *http.Request) error
	// Port is the healthz server port.
	Port int
	// StopCh is the stop channel used to shut down the component
	StopCh <-chan struct{}
}

// RunOrDie starts the leader election code loop with the provided config or panics if the config fails to validate.
// A wrapper of Kubernetes leaderelection package, more info here: https://github.com/caicloud/leader-election-example
func RunOrDie(opt Option) {
	id := opt.ID
	if id == "" {
		id = os.Getenv("POD_NAME")
	}
	if id == "" {
		panic("The ID option or POD_NAME environment variable must be set")
	}

	lock := &resourcelock.EndpointsLock{
		EndpointsMeta: metav1.ObjectMeta{
			Name:      opt.LeaseLockName,
			Namespace: opt.LeaseLockNamespace,
		},
		Client: opt.KubeClient.CoreV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: id,
		},
	}

	electionChecker := leaderelection.NewLeaderHealthzAdaptor(time.Second * 20)
	leaderElector, err := leaderelection.NewLeaderElector(leaderelection.LeaderElectionConfig{
		Lock:          lock,
		LeaseDuration: 60 * time.Second,
		RenewDeadline: 15 * time.Second,
		RetryPeriod:   3 * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: opt.Run,
			OnStoppedLeading: func() {
				log.Infof("lost: %s", id)
				os.Exit(0)
			},
			OnNewLeader: func(identity string) {
				if identity == id {
					return
				}
				log.Infof("new leader elected: %s", identity)
			},
		},
		WatchDog:        electionChecker,
		ReleaseOnCancel: true,
	})
	if err != nil {
		panic(err)
	}

	// setup healthz checks
	checks := []healthz.HealthChecker{electionChecker}
	if opt.LivenessChecker != nil {
		checks = append(checks, newNamedChecker("liveness", leaderElector, opt.LivenessChecker))
	}
	mux := http.NewServeMux()
	healthz.InstallHandler(mux, checks...)

	go func() {
		log.Infof("[healthz] Start listening to %d", opt.Port)

		if err := http.ListenAndServe(fmt.Sprintf(":%d", opt.Port), mux); err != nil {
			log.Fatalf("[healthz] Error starting server: %v", err)
		}
	}()

	// use a Go context so we can tell the leaderelection code when we want to step down
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		<-opt.StopCh
		log.Info("Received termination, signaling shutdown")
		cancel()
	}()

	leaderElector.Run(ctx)
}
