/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/pflag"
	graceful "gopkg.in/tylerb/graceful.v1"

	"github.com/caicloud/devops-admin/cmd/options"
	"github.com/caicloud/devops-admin/pkg/api/v1/handler"
	"github.com/caicloud/devops-admin/pkg/store"
)

const (
	DefaultPort               = 7088
	DefaultTerminationTimeout = 5 * time.Minute
)

func main() {
	// Log to standard error instead of files.
	flag.Set("logtostderr", "true")

	// Flushes all pending log I/O.
	defer glog.Flush()

	// Initialize flags.
	f := &options.Flags{}
	f.AddFlags(pflag.CommandLine)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	http.Handle("/metrics", prometheus.Handler())

	closing := make(chan struct{})
	mongoConfig := &store.MgoConfig{
		Addrs: f.MongoServers,
		DB:    f.MongoDatabase,
		Mode:  "strong",
	}
	mclosed, err := store.InitMongo(mongoConfig, closing)
	if err != nil {
		glog.Fatalf("init MongoDB error: %v", err)
	}
	go background(closing, mclosed)

	handler, err := handler.NewHandler(f.CycloneServer)
	if err != nil {
		glog.Fatalf("new handler error: %v", err)
	}

	port := DefaultPort
	if f.Port != 0 {
		port = int(f.Port)
	}

	glog.Infof("devops admin starts listening on %d", port)
	graceful.Run(
		fmt.Sprintf("%s:%d", f.Address, port),
		DefaultTerminationTimeout,
		handler,
	)
	glog.Error("Server stopped")
}

// background must be a daemon goroutine for devops-admin.
// It can catch system signal and send signal to other goroutine before program exits.
func background(closing, mclosed chan struct{}) {
	closed := []chan struct{}{mclosed}
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-sigs
	glog.Info("capture system signal, will close \"closing\" channel")
	close(closing)
	for _, c := range closed {
		<-c
	}
	glog.Info("exit the process with 0")
	os.Exit(0)
}
