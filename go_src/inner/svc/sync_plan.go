package svc

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/cockroachdb/errors"
)

// BuildSyncPlan 根据主库配置和用户选择生成覆盖同步计划，不修改文件。
func (s *VaultService) BuildSyncPlan(req SyncRequest) (SyncPlan, error) {
	mainVaultPath, err := precheckVaultPath(req.MainVaultPath)
	if err != nil {
		return SyncPlan{}, errors.WithStack(err)
	}

	configItems, err := s.ListConfigItems(mainVaultPath)
	if err != nil {
		return SyncPlan{}, errors.WithStack(err)
	}

	selectedPaths := make([]string, 0, len(req.SelectedPaths))
	selectedPathSet := make(map[string]struct{}, len(req.SelectedPaths))
	for _, path := range req.SelectedPaths {
		path = normalizeSyncPath(path)
		if path == "" {
			return SyncPlan{}, errors.New("同步配置路径不能为空")
		}
		if err := precheckSyncPath(path); err != nil {
			return SyncPlan{}, err
		}
		if _, exists := selectedPathSet[path]; exists {
			continue
		}
		selectedPathSet[path] = struct{}{}
		selectedPaths = append(selectedPaths, path)
	}

	knownPaths := make(map[string]struct{}, len(configItems))
	for _, item := range configItems {
		knownPaths[normalizeSyncPath(item.Path)] = struct{}{}
	}
	for _, path := range selectedPaths {
		if _, ok := knownPaths[path]; !ok {
			return SyncPlan{}, errors.Errorf("主库不存在配置项: %s", path)
		}
	}

	targets := make([]TargetSyncPlan, 0, len(req.TargetVaultPaths))
	for _, targetPath := range req.TargetVaultPaths {
		targetVaultPath, err := precheckVaultPath(targetPath)
		if err != nil {
			return SyncPlan{}, errors.WithStack(err)
		}
		if samePath(mainVaultPath, targetVaultPath) {
			return SyncPlan{}, errors.Errorf("主库不能作为从库: %s", targetVaultPath)
		}

		items := make([]SyncPlanItem, 0, len(selectedPaths))
		for _, path := range selectedPaths {
			_, err := os.Stat(filepath.Join(targetVaultPath, ".obsidian", filepath.FromSlash(path)))
			action := SyncPlanActionCreate
			if err == nil {
				action = SyncPlanActionOverwrite
			} else if !os.IsNotExist(err) {
				return SyncPlan{}, errors.Wrapf(err, "检查目标配置项失败: %s", path)
			}
			items = append(items, SyncPlanItem{Path: path, Action: action})
		}
		targets = append(targets, TargetSyncPlan{
			VaultPath: targetVaultPath,
			Items:     items,
		})
	}

	return SyncPlan{
		MainVaultPath: mainVaultPath,
		Targets:       targets,
	}, nil
}

// normalizeSyncPath 统一同步路径格式，保留目录路径末尾的斜杠语义。
func normalizeSyncPath(path string) string {
	path = filepath.ToSlash(strings.TrimSpace(path))
	return path
}

// samePath 判断两个已解析路径是否指向同一个目录。
func samePath(left string, right string) bool {
	return strings.EqualFold(filepath.Clean(left), filepath.Clean(right))
}
