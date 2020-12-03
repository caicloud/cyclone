package controller

import (
	"context"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/common"
	"github.com/caicloud/cyclone/pkg/k8s/clientset"
)

// InitLogger inits logging
func InitLogger(logging *LoggingConfig) {
	log.WithField("level", logging.Level).Info("Setting log level")

	switch strings.ToLower(logging.Level) {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	default:
		log.Fatalf("Unknown level: %s", logging.Level)
	}

	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
		ForceColors:   true,
		FullTimestamp: true,
	})
	log.SetOutput(os.Stdout)
}

// InitControlCluster initializes control cluster for workflow to run.
func InitControlCluster(client clientset.Interface) error {
	// Create ExecutionCluster instance for control cluster. This makes it possible to
	// use only workflow engine to run workflow in control cluster.
	// Create ExecutionCluster resource for Workflow Engine to use
	_, err := client.CycloneV1alpha1().ExecutionClusters().Create(context.TODO(), &v1alpha1.ExecutionCluster{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: common.ControlClusterName,
		},
	}, meta_v1.CreateOptions{})

	// If the CR already exists, just ignore it.
	if err != nil && errors.IsAlreadyExists(err) {
		log.Infof("ExecutionCluster resource '%s' already exist", common.ControlClusterName)
		return nil
	}

	return err
}
