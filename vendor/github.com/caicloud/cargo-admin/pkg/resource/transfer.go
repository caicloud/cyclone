package resource

// 此文件中的绝大多数代码都是从个 v1.4.0 的 harbor 抄过来的，参考了 harbor jobservice 中 transfer image 的实现

import (
	"fmt"
	"strings"

	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/common/utils/registry"
	"github.com/vmware/harbor/src/common/utils/registry/auth"
	"github.com/vmware/harbor/src/jobservice/utils"
)

const (
	// StateInitialize ...
	StateInitialize = "initialize"
	// StateCheck ...
	StateCheck = "check"
	// StatePullManifest ...
	StatePullManifest = "pull_manifest"
	// StateTransferBlob ...
	StateTransferBlob = "transfer_blob"
	// StatePushManifest ...
	StatePushManifest = "push_manifest"
)

// BaseHandler holds informations shared by other state handlers
type BaseHandler struct {
	srcURL        string // url of source registry
	srcUsr        string // username ...
	srcPwd        string // password ...
	srcProject    string // project_name
	srcRepository string // prject_name/repo_name
	srcTag        string

	dstURL        string // url of target registry
	dstUsr        string // username ...
	dstPwd        string // password ...
	dstProject    string // project_name
	dstRepository string // prject_name/repo_name
	dstTag        string

	insecure bool // whether skip secure check when using https

	srcClient *registry.Repository
	dstClient *registry.Repository

	manifest distribution.Manifest // manifest of tags[0]
	digest   string                //digest of tags[0]'s manifest
	blobs    []string              // blobs need to be transferred for tags[0]

	blobsExistence map[string]bool //key: digest of blob, value: existence

	logger *log.Logger
}

// InitBaseHandler initializes a BaseHandler.
func InitBaseHandler(src, dst *ImageInfo) *BaseHandler {
	return &BaseHandler{
		srcURL:         src.Registry.Host,
		srcUsr:         src.Registry.Username,
		srcPwd:         src.Registry.Password,
		srcProject:     src.Project.Name,
		srcRepository:  src.Repositry,
		srcTag:         src.Tag,
		dstURL:         dst.Registry.Host,
		dstUsr:         dst.Registry.Username,
		dstPwd:         dst.Registry.Password,
		dstProject:     dst.Project.Name,
		dstRepository:  dst.Repositry,
		dstTag:         dst.Tag,
		blobsExistence: make(map[string]bool, 10),
		logger:         log.DefaultLogger(),
	}
}

// Exit ...
func (b *BaseHandler) Exit() error {
	return nil
}

// Initializer creates clients for source and destination registry,
// lists tags of the repository if parameter tags is nil.
type Initializer struct {
	*BaseHandler
}

// Enter ...
func (i *Initializer) Enter() (string, error) {
	i.logger.Infof("initializing: ")
	i.logger.Infof("source URL: %s, source project: %s, source repository: %s, source tag: %s", i.srcURL, i.srcProject, i.srcRepository, i.srcTag)
	i.logger.Infof("destination URL: %s, destination project: %s, destination repository: %s, destination tag: %s", i.dstURL, i.dstProject, i.dstRepository, i.dstTag)

	state, err := i.enter()
	if err != nil && retry(err) {
		i.logger.Info("waiting for retrying...")
		return models.JobRetrying, nil
	}

	return state, err
}

func (i *Initializer) enter() (string, error) {
	srcCred := auth.NewBasicAuthCredential(i.srcUsr, i.srcPwd)
	srcClient, err := utils.NewRepositoryClient(i.srcURL, true, srcCred,
		"", fmt.Sprintf("%s/%s", i.srcProject, i.srcRepository))
	if err != nil {
		i.logger.Errorf("an error occurred while creating source repository client: %v", err)
		return "", err
	}
	i.srcClient = srcClient

	dstCred := auth.NewBasicAuthCredential(i.dstUsr, i.dstPwd)
	dstClient, err := utils.NewRepositoryClient(i.dstURL, true, dstCred,
		"", fmt.Sprintf("%s/%s", i.dstProject, i.dstRepository))
	if err != nil {
		i.logger.Errorf("an error occurred while creating destination repository client: %v", err)
		return "", err
	}
	i.dstClient = dstClient

	i.logger.Infof("initialization completed:")
	i.logger.Infof("source URL: %s, source project: %s, source repository: %s, source tag: %s", i.srcURL, i.srcProject, i.srcRepository, i.srcTag)
	i.logger.Infof("destination URL: %s, destination project: %s, destination repository: %s, destination tag: %s", i.dstURL, i.dstProject, i.dstRepository, i.dstTag)

	return StateCheck, nil
}

// ManifestPuller pulls the manifest of a tag. And if no tag needs to be pulled,
// the next state that state machine should enter is "finished".
type ManifestPuller struct {
	*BaseHandler
}

// Enter pulls manifest of a tag and checks if all blobs exist in the destination registry
func (m *ManifestPuller) Enter() (string, error) {
	state, err := m.enter()
	if err != nil && retry(err) {
		m.logger.Info("waiting for retrying...")
		return models.JobRetrying, nil
	}

	return state, err

}

func (m *ManifestPuller) enter() (string, error) {
	name := fmt.Sprintf("%s/%s", m.srcProject, m.srcRepository)
	tag := m.srcTag

	acceptMediaTypes := []string{schema1.MediaTypeManifest, schema2.MediaTypeManifest}
	digest, mediaType, payload, err := m.srcClient.PullManifest(tag, acceptMediaTypes)
	if err != nil {
		m.logger.Errorf("an error occurred while pulling manifest of %s:%s from %s: %v", name, tag, m.srcURL, err)
		return "", err
	}
	m.digest = digest
	m.logger.Infof("manifest of %s:%s pulled successfully from %s: %s", name, tag, m.srcURL, digest)

	if strings.Contains(mediaType, "application/json") {
		mediaType = schema1.MediaTypeManifest
	}

	manifest, _, err := registry.UnMarshal(mediaType, payload)
	if err != nil {
		m.logger.Errorf("an error occurred while parsing manifest of %s:%s from %s: %v", name, tag, m.srcURL, err)
		return "", err
	}

	m.manifest = manifest

	// all blobs(layers and config)
	var blobs []string

	for _, discriptor := range manifest.References() {
		blobs = append(blobs, discriptor.Digest.String())
	}

	m.logger.Infof("all blobs of %s:%s from %s: %v", name, tag, m.srcURL, blobs)

	for _, blob := range blobs {
		exist, ok := m.blobsExistence[blob]
		if !ok {
			exist, err = m.dstClient.BlobExist(blob)
			if err != nil {
				m.logger.Errorf("an error occurred while checking existence of blob %s of %s:%s on %s: %v", blob, name, tag, m.dstURL, err)
				return "", err
			}
			m.blobsExistence[blob] = exist
		}

		if !exist {
			m.blobs = append(m.blobs, blob)
		} else {
			m.logger.Infof("blob %s of %s:%s already exists in %s", blob, name, tag, m.dstURL)
		}
	}
	m.logger.Infof("blobs of %s:%s need to be transferred to %s: %v", name, tag, m.dstURL, m.blobs)

	return StateTransferBlob, nil
}

// BlobTransfer transfers blobs of a tag
type BlobTransfer struct {
	*BaseHandler
}

// Enter pulls blobs and then pushs them to destination registry.
func (b *BlobTransfer) Enter() (string, error) {
	state, err := b.enter()
	if err != nil && retry(err) {
		b.logger.Info("waiting for retrying...")
		return models.JobRetrying, nil
	}

	return state, err

}

func (b *BlobTransfer) enter() (string, error) {
	name := fmt.Sprintf("%s/%s", b.srcProject, b.srcRepository)
	tag := b.srcTag
	for _, blob := range b.blobs {
		b.logger.Infof("transferring blob %s of %s:%s to %s ...", blob, name, tag, b.dstURL)
		size, data, err := b.srcClient.PullBlob(blob)
		if err != nil {
			b.logger.Errorf("an error occurred while pulling blob %s of %s:%s from %s: %v", blob, name, tag, b.srcURL, err)
			return "", err
		}
		if data != nil {
			defer data.Close()
		}
		if err = b.dstClient.PushBlob(blob, size, data); err != nil {
			b.logger.Errorf("an error occurred while pushing blob %s of %s:%s to %s : %v", blob, name, tag, b.dstURL, err)
			return "", err
		}
		b.logger.Infof("blob %s of %s:%s transferred to %s completed", blob, name, tag, b.dstURL)
	}

	return StatePushManifest, nil
}

// ManifestPusher pushs the manifest to destination registry
type ManifestPusher struct {
	*BaseHandler
}

// Enter checks the existence of manifest in the source registry first, and if it
// exists, pushs it to destination registry. The checking operation is to avoid
// the situation that the tag is deleted during the blobs transfering
func (m *ManifestPusher) Enter() (string, error) {
	state, err := m.enter()
	if err != nil && retry(err) {
		m.logger.Info("waiting for retrying...")
		return models.JobRetrying, nil
	}

	return state, err

}

func (m *ManifestPusher) enter() (string, error) {
	name := fmt.Sprintf("%s/%s", m.dstProject, m.dstRepository)
	tag := m.srcTag
	_, exist, err := m.srcClient.ManifestExist(tag)
	if err != nil {
		m.logger.Infof("an error occurred while checking the existence of manifest of %s:%s on %s: %v", name, tag, m.srcURL, err)
		return "", err
	}
	if !exist {
		m.logger.Infof("manifest of %s:%s does not exist on source registry %s, cancel manifest pushing", name, tag, m.srcURL)
	} else {
		m.logger.Infof("manifest of %s:%s exists on source registry %s, continue manifest pushing", name, tag, m.srcURL)

		tag = m.dstTag
		digest, manifestExist, err := m.dstClient.ManifestExist(tag)
		if manifestExist && digest == m.digest {
			m.logger.Infof("manifest of %s:%s exists on destination registry %s, skip manifest pushing", name, tag, m.dstURL)
			m.manifest = nil
			m.digest = ""
			m.blobs = nil
			return StatePullManifest, nil
		}

		mediaType, data, err := m.manifest.Payload()
		if err != nil {
			m.logger.Errorf("an error occurred while getting payload of manifest for %s:%s : %v", name, tag, err)
			return "", err
		}

		if _, err = m.dstClient.PushManifest(tag, mediaType, data); err != nil {
			m.logger.Errorf("an error occurred while pushing manifest of %s:%s to %s : %v", name, tag, m.dstURL, err)
			return "", err
		}
		m.logger.Infof("manifest of %s:%s has been pushed to %s", name, tag, m.dstURL)
	}

	// m.tags = m.tags[1:]
	m.manifest = nil
	m.digest = ""
	m.blobs = nil

	return StatePullManifest, nil
}
