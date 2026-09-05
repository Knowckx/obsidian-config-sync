package svc_test

import (
	"path/filepath"
	"testing"

	"obsidian-config-sync/go_src/inner/svc"
)

func Test_ScanVaults(t *testing.T) {
	testPath := filepath.Join("..", "..", "..", "test_cases")

	service := &svc.VaultService{}
	result, err := service.ScanVaults(testPath)
	if err != nil {
		t.Fatalf("ScanVaults 失败: %v", err)
	}
	t.Logf("result:\n%+v", result)
}
