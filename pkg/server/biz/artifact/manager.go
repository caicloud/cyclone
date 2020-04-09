package artifact

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"k8s.io/kubernetes/pkg/volume/util/fs"

	"github.com/caicloud/cyclone/pkg/server/common"
	"github.com/caicloud/nirvana/log"
)

// Manager manages artifacts
type Manager struct {
	cleanPeriod     time.Duration
	artifactHomeDir string
}

// NewManager Initiates a artifacts manager.
// If artifactHomeDir not passed, the default '/var/lib/cyclone' will be used.
func NewManager(artifactHomeDir ...string) *Manager {
	var home = common.CycloneHome
	if artifactHomeDir != nil && artifactHomeDir[0] != "" {
		home = artifactHomeDir[0]
	}
	m := &Manager{
		cleanPeriod:     time.Duration(time.Hour),
		artifactHomeDir: home,
	}

	return m
}

// CleanPeriodically will clean up artifacts which exceeded retention time periodically.
// This func will run forever unless panics, you'd better invoke it by a go-routine.
func (m *Manager) CleanPeriodically(retention time.Duration) {
	t := time.NewTicker(time.Duration(time.Hour))
	defer t.Stop()

	for ; true; <-t.C {
		log.Info("Start to scan and clean artifacts")
		if err := m.scanAndClean(selectArtifact, retention); err != nil {
			log.Warningf("Clean artifacts error: ", err)
		}
	}
}

type artifactSelector func(path string) bool

func selectArtifact(path string) bool {
	path = strings.TrimPrefix(path, "/")
	// artifact path must be in format of {tenant}/{project}/{workflow}/{workflowrun}/artifacts/{stage}/xxx
	splitPaths := strings.SplitN(path, "/", 7)

	if len(splitPaths) < 7 {
		return false
	}

	if splitPaths[4] != "artifacts" {
		return false
	}
	return true
}

// scanAndClean scans the artifacts folders and finds out exceeded retention time artifacts, and
// then delete them.
func (m *Manager) scanAndClean(selectArtifact artifactSelector, retention time.Duration) error {
	return filepath.Walk(m.artifactHomeDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("walk in %s error: %v", path, err)
		}

		if !selectArtifact(strings.TrimPrefix(path, m.artifactHomeDir)) {
			return nil
		}

		if time.Now().Before(info.ModTime().Add(retention)) {
			return nil
		}

		log.Infof("Start to remove artifact: %s", path)
		return os.RemoveAll(path)
	})
}

// GetDiskAvailablePercentage returns available space in percentage format of the artifact home folder.
func (m *Manager) GetDiskAvailablePercentage() (float64, error) {
	s := time.Now()

	available, capacity, _, _, _, _, err := fs.FsInfo(m.artifactHomeDir)
	if err != nil {
		return 0, err
	}

	e := time.Now()
	log.Infof("fsInfo cost time: %s, available: %d, capacity: %d, percentage: %v",
		e.Sub(s), available, capacity, float64(available)/float64(capacity))

	return float64(available) / float64(capacity), nil
}
