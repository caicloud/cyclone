package handler

import (
	"testing"
)

func TestGetTemplates(t *testing.T) {
	_, err := getConfigTemplates("../../../../config/templates/templates.yaml")

	if err != nil {
		t.Fatalf("get config templates error:%v", err)
	}
}
