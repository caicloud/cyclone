package resource

import (
	"context"
	"fmt"
	"strings"

	. "github.com/caicloud/cargo-admin/pkg/errors"
	"github.com/caicloud/cargo-admin/pkg/models"

	"github.com/caicloud/nirvana/log"

	"github.com/vmware/harbor/src/common/utils"
)

type ImageCopyReq struct {
	Source *ImageInfo
	Target *ImageInfo
}

// image copy 在 PRD 中被设计为仅提供同镜像仓库内的镜像拷贝，但在实现中，没有这个限制，方便之后扩展、
// image copy 整体流程
// 1. Initializer： 确认可以连接到源 registry 和目的 registry；
// 2. ManifestPuller：将镜像 pull 下俩；
// 3. BlobTransfer：将镜像 Blob 推送到目的 registry（经测试，此步骤不可或缺）；
// 4. ManifestPusher：将镜像的 Manifest 推送到目的 registry.
func TriggerImageCopy(ctx context.Context, tenant string, username string, icReq *ImageCopyReq) error {
	base := InitBaseHandler(icReq.Source, icReq.Target)

	var state string
	var err error
	state, err = (&Initializer{BaseHandler: base}).Enter()
	if err != nil {
		log.Errorf("Initializer error: %v", err)
		return err
	}
	log.Infof("image copy in state: %s", state)
	state, err = (&ManifestPuller{BaseHandler: base}).Enter()
	if err != nil {
		log.Errorf("ManifestPuller error: %v", err)
		return err
	}
	log.Infof("image copy in state: %s", state)
	state, err = (&BlobTransfer{BaseHandler: base}).Enter()
	if err != nil {
		log.Errorf("BlobTransfer error: %v", err)
		return err
	}
	log.Infof("image copy in state: %s", state)
	state, err = (&ManifestPusher{BaseHandler: base}).Enter()
	if err != nil {
		log.Errorf("ManifestPusher error: %v", err)
		return err
	}
	log.Infof("image copy in state: %s", state)

	return nil
}

// =================================================================================================

type ImageInfo struct {
	Registry  *models.RegistryInfo
	Project   *models.ProjectInfo
	Repositry string
	Tag       string
}

// 通过 cargo.caicloudprivatetest.com/library/nginx:v1.0 镜像，获得这个镜像所在的 registry、project 信息
// 并将此 image 的 repository 和 tag 解析出来。
func ParseImageInfo(image string) (*ImageInfo, error) {
	image = strings.TrimLeft(image, "/")
	image = strings.TrimRight(image, "/")
	index1 := strings.Index(image, "/")
	index2 := strings.LastIndex(image, ":")
	if index2 <= index1 || index1 == -1 || index2 == -1 {
		return nil, ErrorUnknownRequest.Error(fmt.Sprintf("%s is invalid", image))
	}
	domain := image[0:index1]
	tag := image[index2+1:]
	project, repository := utils.ParseRepository(image[index1+1 : index2])

	reginfo, err := models.Registry.FindByDomain(domain)
	if err != nil {
		return nil, err
	}
	pinfo, err := models.Project.FindByNameWithoutTenant(reginfo.Name, project)
	if err != nil {
		return nil, err
	}

	return &ImageInfo{
		Registry:  reginfo,
		Project:   pinfo,
		Repositry: repository,
		Tag:       tag,
	}, nil
}
