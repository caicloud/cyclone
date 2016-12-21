// +build linux

package plugin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/chrootarchive"
	"github.com/docker/docker/pkg/stringid"
	"github.com/docker/docker/plugin/distribution"
	"github.com/docker/docker/plugin/v2"
	"github.com/docker/docker/reference"
	"golang.org/x/net/context"
)

var (
	validFullID    = regexp.MustCompile(`^([a-f0-9]{64})$`)
	validPartialID = regexp.MustCompile(`^([a-f0-9]{1,64})$`)
)

// Disable deactivates a plugin, which implies that they cannot be used by containers.
func (pm *Manager) Disable(name string) error {
	p, err := pm.pluginStore.GetByName(name)
	if err != nil {
		return err
	}
	if err := pm.disable(p); err != nil {
		return err
	}
	pm.pluginEventLogger(p.GetID(), name, "disable")
	return nil
}

// Enable activates a plugin, which implies that they are ready to be used by containers.
func (pm *Manager) Enable(name string, config *types.PluginEnableConfig) error {

	p, err := pm.pluginStore.GetByName(name)
	if err != nil {
		return err
	}

	p.TimeoutInSecs = config.Timeout
	if err := pm.enable(p, false); err != nil {
		return err
	}
	pm.pluginEventLogger(p.GetID(), name, "enable")
	return nil
}

// Inspect examines a plugin config
func (pm *Manager) Inspect(refOrID string) (tp types.Plugin, err error) {
	// Match on full ID
	if validFullID.MatchString(refOrID) {
		p, err := pm.pluginStore.GetByID(refOrID)
		if err == nil {
			return p.PluginObj, nil
		}
	}

	// Match on full name
	if pluginName, err := getPluginName(refOrID); err == nil {
		if p, err := pm.pluginStore.GetByName(pluginName); err == nil {
			return p.PluginObj, nil
		}
	}

	// Match on partial ID
	if validPartialID.MatchString(refOrID) {
		p, err := pm.pluginStore.Search(refOrID)
		if err == nil {
			return p.PluginObj, nil
		}
		return tp, err
	}

	return tp, fmt.Errorf("no such plugin name or ID associated with %q", refOrID)
}

func (pm *Manager) pull(ref reference.Named, metaHeader http.Header, authConfig *types.AuthConfig, pluginID string) (types.PluginPrivileges, error) {
	pd, err := distribution.Pull(ref, pm.registryService, metaHeader, authConfig)
	if err != nil {
		logrus.Debugf("error in distribution.Pull(): %v", err)
		return nil, err
	}

	if err := distribution.WritePullData(pd, filepath.Join(pm.libRoot, pluginID), true); err != nil {
		logrus.Debugf("error in distribution.WritePullData(): %v", err)
		return nil, err
	}

	tag := distribution.GetTag(ref)
	p := v2.NewPlugin(ref.Name(), pluginID, pm.runRoot, pm.libRoot, tag)
	if err := p.InitPlugin(); err != nil {
		return nil, err
	}
	pm.pluginStore.Add(p)

	pm.pluginEventLogger(pluginID, ref.String(), "pull")
	return p.ComputePrivileges(), nil
}

// Pull pulls a plugin and computes the privileges required to install it.
func (pm *Manager) Pull(name string, metaHeader http.Header, authConfig *types.AuthConfig) (types.PluginPrivileges, error) {
	ref, err := distribution.GetRef(name)
	if err != nil {
		logrus.Debugf("error in distribution.GetRef: %v", err)
		return nil, err
	}
	name = ref.String()

	if p, _ := pm.pluginStore.GetByName(name); p != nil {
		logrus.Debug("plugin already exists")
		return nil, fmt.Errorf("%s exists", name)
	}

	pluginID := stringid.GenerateNonCryptoID()
	pluginDir := filepath.Join(pm.libRoot, pluginID)
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		logrus.Debugf("error in MkdirAll: %v", err)
		return nil, err
	}

	priv, err := pm.pull(ref, metaHeader, authConfig, pluginID)
	if err != nil {
		if err := os.RemoveAll(pluginDir); err != nil {
			logrus.Warnf("unable to remove %q from failed plugin pull: %v", pluginDir, err)
		}
		return nil, err
	}

	return priv, nil
}

// List displays the list of plugins and associated metadata.
func (pm *Manager) List() ([]types.Plugin, error) {
	plugins := pm.pluginStore.GetAll()
	out := make([]types.Plugin, 0, len(plugins))
	for _, p := range plugins {
		out = append(out, p.PluginObj)
	}
	return out, nil
}

// Push pushes a plugin to the store.
func (pm *Manager) Push(name string, metaHeader http.Header, authConfig *types.AuthConfig) error {
	p, err := pm.pluginStore.GetByName(name)
	if err != nil {
		return err
	}
	dest := filepath.Join(pm.libRoot, p.GetID())
	config, err := ioutil.ReadFile(filepath.Join(dest, "config.json"))
	if err != nil {
		return err
	}

	var dummy types.Plugin
	err = json.Unmarshal(config, &dummy)
	if err != nil {
		return err
	}

	rootfs, err := archive.Tar(filepath.Join(dest, "rootfs"), archive.Gzip)
	if err != nil {
		return err
	}
	defer rootfs.Close()

	_, err = distribution.Push(name, pm.registryService, metaHeader, authConfig, ioutil.NopCloser(bytes.NewReader(config)), rootfs)
	// XXX: Ignore returning digest for now.
	// Since digest needs to be written to the ProgressWriter.
	return err
}

// Remove deletes plugin's root directory.
func (pm *Manager) Remove(name string, config *types.PluginRmConfig) error {
	p, err := pm.pluginStore.GetByName(name)
	if err != nil {
		return err
	}

	if !config.ForceRemove {
		p.RLock()
		if p.RefCount > 0 {
			p.RUnlock()
			return fmt.Errorf("plugin %s is in use", p.Name())
		}
		p.RUnlock()

		if p.IsEnabled() {
			return fmt.Errorf("plugin %s is enabled", p.Name())
		}
	}

	if p.IsEnabled() {
		if err := pm.disable(p); err != nil {
			logrus.Errorf("failed to disable plugin '%s': %s", p.Name(), err)
		}
	}

	pm.pluginStore.Remove(p)
	os.RemoveAll(filepath.Join(pm.libRoot, p.GetID()))
	pm.pluginEventLogger(p.GetID(), name, "remove")
	return nil
}

// Set sets plugin args
func (pm *Manager) Set(name string, args []string) error {
	p, err := pm.pluginStore.GetByName(name)
	if err != nil {
		return err
	}
	return p.Set(args)
}

// CreateFromContext creates a plugin from the given pluginDir which contains
// both the rootfs and the config.json and a repoName with optional tag.
func (pm *Manager) CreateFromContext(ctx context.Context, tarCtx io.Reader, options *types.PluginCreateOptions) error {
	pluginID := stringid.GenerateNonCryptoID()

	pluginDir := filepath.Join(pm.libRoot, pluginID)
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return err
	}

	// In case an error happens, remove the created directory.
	if err := pm.createFromContext(ctx, pluginID, pluginDir, tarCtx, options); err != nil {
		if err := os.RemoveAll(pluginDir); err != nil {
			logrus.Warnf("unable to remove %q from failed plugin creation: %v", pluginDir, err)
		}
		return err
	}

	return nil
}

func (pm *Manager) createFromContext(ctx context.Context, pluginID, pluginDir string, tarCtx io.Reader, options *types.PluginCreateOptions) error {
	if err := chrootarchive.Untar(tarCtx, pluginDir, nil); err != nil {
		return err
	}

	repoName := options.RepoName
	ref, err := distribution.GetRef(repoName)
	if err != nil {
		return err
	}
	name := ref.Name()
	tag := distribution.GetTag(ref)

	p := v2.NewPlugin(name, pluginID, pm.runRoot, pm.libRoot, tag)
	if err := p.InitPlugin(); err != nil {
		return err
	}

	if err := pm.pluginStore.Add(p); err != nil {
		return err
	}

	pm.pluginEventLogger(p.GetID(), repoName, "create")

	return nil
}

func getPluginName(name string) (string, error) {
	named, err := reference.ParseNamed(name) // FIXME: validate
	if err != nil {
		return "", err
	}
	if reference.IsNameOnly(named) {
		named = reference.WithDefaultTag(named)
	}
	ref, ok := named.(reference.NamedTagged)
	if !ok {
		return "", fmt.Errorf("invalid name: %s", named.String())
	}
	return ref.String(), nil
}
