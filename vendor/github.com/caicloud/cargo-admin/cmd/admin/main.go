package main

import (
	"os"
	"os/signal"
	"syscall"

	api "github.com/caicloud/cargo-admin/pkg/api/admin/descriptor"
	"github.com/caicloud/cargo-admin/pkg/env"
	"github.com/caicloud/cargo-admin/pkg/harbor"
	"github.com/caicloud/cargo-admin/pkg/models"
	"github.com/caicloud/cargo-admin/pkg/resource"
	"github.com/caicloud/cargo-admin/pkg/utils/parser"

	"github.com/caicloud/nirvana"
	"github.com/caicloud/nirvana/log"
	"github.com/caicloud/nirvana/plugins/metrics"
	"github.com/caicloud/nirvana/plugins/profiling"

	config "qiniupkg.com/x/config.v7"
)

func main() {
	config.Init("f", "", "cargo-admin-config.json")
	var conf env.AdminConfig
	err := config.Load(&conf)
	if err != nil {
		log.Fatalf("load cargo-admin config file failed: %v", err)
		return
	}
	closing := make(chan struct{})

	closed := initialize(conf, closing)

	go run(conf.Address)
	go gracefulShutdown(closing)
	waitCleanup(closed)
}

func initialize(conf env.AdminConfig, closing chan struct{}) []chan struct{} {
	mClosed, err := models.InitAdminMongo(conf.Mongo, closing)
	if err != nil {
		log.Fatalf("init mongodb error: %v", err)
		return nil
	}

	// Initialize default registry and tenant
	resource.InitRegistry(conf.DefaultRegistry)
	env.SystemTenant = conf.SystemTenant

	// Initialize harbor, harbor sessions will be refreshed periodically.
	hClosed, err := harbor.InitHarbor(closing)
	if err != nil {
		log.Fatalf("init harbor failed: %v", err)
		return nil
	}
	log.Info("init harbor success")

	// Initialize default projects
	err = resource.InitProject(conf.Projects)
	if err != nil {
		log.Fatalf("init projects error: %v", err)
		return nil
	}
	log.Info("init projects success")

	return []chan struct{}{mClosed, hClosed}
}

func run(addr string) {
	ip, port, err := parser.HostPort(addr)
	if err != nil {
		log.Fatalf("parse address %s error: %v", addr, err)
		return
	}
	config := nirvana.NewDefaultConfig()
	nirvana.IP(ip)(config)
	nirvana.Port(port)(config)
	config.Configure(
		metrics.Path("/metrics"),
		profiling.Path("/debug/pprof/"),
		profiling.Contention(true),
	)
	api.ConfigService(config)

	log.Infof("API service listening on %s:%d", ip, port)
	if err = nirvana.NewServer(config).Serve(); err != nil {
		log.Fatal(err)
	}
}

func gracefulShutdown(closing chan struct{}) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	log.Infof("capture system signal %s, to close \"closing\" channel", <-signals)
	close(closing)
}

func waitCleanup(closed []chan struct{}) {
	for _, c := range closed {
		<-c
	}
}
