package types

import (
	"sync"
)

type ImageItem struct {
	Origin string `json:"origin"`
	Image  string `json:"image"`
}

type UploadImages struct {
	Images []ImageItem `json:"images"`
}

func (images *UploadImages) Add(item ImageItem) {
	images.Images = append(images.Images, item)
}

type ImageUploadStats struct {
	Succeed []string `json:"succeed"`
	Failed  []string `json:"failed"`
}

type ImageUploadResult struct {
	Stats *ImageUploadStats
	mux   sync.Mutex
}

func (result *ImageUploadResult) AddSucceed(tag string) {
	result.mux.Lock()
	defer result.mux.Unlock()
	result.Stats.Succeed = append(result.Stats.Succeed, tag)
}

func (result *ImageUploadResult) AddFailed(tag string) {
	result.mux.Lock()
	defer result.mux.Unlock()
	result.Stats.Failed = append(result.Stats.Failed, tag)
}
