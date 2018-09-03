package router

import (
	"testing"
)

func TestGetTemplates(t *testing.T) {
	_, err := getTemplates("../../../config/templates/templates.yaml")

	if err != nil {
		t.Fatalf("get templates error:%v", err)
	}
}
