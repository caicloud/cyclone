package docker

import (
	"regexp"
	"strings"
)

const BadTarFile = "Error processing tar file(exit status 1)"

var pattern = regexp.MustCompile(`Loaded image:\s*([\w:./-]+)`)

func LoadedImages(output string) []string {
	matches := pattern.FindAllStringSubmatch(output, -1)
	imgs := make([]string, 0)
	for _, m := range matches {
		imgs = append(imgs, m[1])
	}
	return imgs
}

func IsBadTarFile(result string) bool {
	return strings.Index(result, BadTarFile) >= 0
}

func IsBuildSucceed(output string) bool {
	return strings.Index(output, "errorDetail") == -1
}

func IsPushSucceed(output string) bool {
	return strings.Index(output, "errorDetail") == -1
}
