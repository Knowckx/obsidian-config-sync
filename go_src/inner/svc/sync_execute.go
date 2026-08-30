package svc

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/cockroachdb/errors"
)

// ExecuteSyncPlan 执行覆盖同步；已有同步任务运行时拒绝新的调用。
func (t *VaultService) ExecuteSyncPlan(plan SyncPlan) (SyncResult, error) {
	if !t.isSyncing.CompareAndSwap(false, true) {
		return SyncResult{}, errors.New("同步正在执行，请等待当前任务完成")
	}
	defer t.isSyncing.Store(false)

	currentPlan, err := t.rebuildSyncPlan(plan)
	if err != nil {
		return SyncResult{}, errors.WithStack(err)
	}
	if !reflect.DeepEqual(plan, currentPlan) {
		return SyncResult{}, errors.New("同步计划已变化，请重新确认")
	}

	backupPath := ""
	// 2026-08-30 备份功能已冻结，暂时停止实际调用。
	// backupPath, err = createSyncBackup(currentPlan)
	// if err != nil {
	// 	return SyncResult{}, errors.Wrap(err, "创建同步备份失败")
	// }

	result := SyncResult{
		BackupPath: backupPath,
		Targets:    make([]TargetSyncResult, 0, len(currentPlan.Targets)),
	}
	for _, target := range currentPlan.Targets {
		targetResult := TargetSyncResult{
			VaultPath: target.VaultPath,
			Items:     make([]SyncResultItem, 0, len(target.Items)),
		}

		for _, item := range target.Items {
			resultItem := SyncResultItem{
				Path:   item.Path,
				Status: resultStatusForAction(item.Action),
			}
			src := filepath.Join(currentPlan.MainVaultPath, ".obsidian", filepath.FromSlash(item.Path))
			dst := filepath.Join(target.VaultPath, ".obsidian", filepath.FromSlash(item.Path))
			if err := copySyncItem(src, dst); err != nil {
				resultItem.Status = SyncResultStatusFailed
				resultItem.Error = errors.Wrapf(err, "同步配置失败: %s", item.Path).Error()
			}
			targetResult.Items = append(targetResult.Items, resultItem)
		}
		result.Targets = append(result.Targets, targetResult)
	}
	// 2026-08-30 备份功能已冻结，暂时停止实际调用。
	// if backupPath != "" {
	// 	if err := pruneSyncBackups(filepath.Dir(backupPath), syncBackupKeepCount); err != nil {
	// 		log.Printf("清理旧同步备份失败: %+v", err)
	// 	}
	// }
	return result, nil
}

// rebuildSyncPlan 根据待执行计划重新生成当前计划。
func (t *VaultService) rebuildSyncPlan(plan SyncPlan) (SyncPlan, error) {
	if len(plan.Targets) == 0 {
		return SyncPlan{}, errors.New("同步目标不能为空")
	}
	if len(plan.Targets[0].Items) == 0 {
		return SyncPlan{}, errors.New("同步配置项不能为空")
	}

	targetVaultPaths := make([]string, 0, len(plan.Targets))
	targetPathSet := make(map[string]struct{}, len(plan.Targets))
	for _, target := range plan.Targets {
		targetPath := strings.ToLower(filepath.Clean(target.VaultPath))
		if _, exists := targetPathSet[targetPath]; exists {
			return SyncPlan{}, errors.Errorf("同步目标重复: %s", target.VaultPath)
		}
		targetPathSet[targetPath] = struct{}{}
		targetVaultPaths = append(targetVaultPaths, target.VaultPath)
	}

	selectedPaths := make([]string, 0, len(plan.Targets[0].Items))
	for _, item := range plan.Targets[0].Items {
		selectedPaths = append(selectedPaths, item.Path)
	}

	return t.BuildSyncPlan(SyncRequest{
		MainVaultPath:    plan.MainVaultPath,
		TargetVaultPaths: targetVaultPaths,
		SelectedPaths:    selectedPaths,
	})
}

// resultStatusForAction 将计划动作转换为成功时的结果状态。
func resultStatusForAction(action SyncPlanAction) SyncResultStatus {
	if action == SyncPlanActionCreate {
		return SyncResultStatusCreated
	}
	return SyncResultStatusOverwrote
}

// precheckSyncPath 检查同步路径是否为 .obsidian 下的相对路径。
func precheckSyncPath(path string) error {
	path = filepath.Clean(filepath.FromSlash(normalizeSyncPath(path)))
	if path == "." || filepath.IsAbs(path) || path == ".." || strings.HasPrefix(path, ".."+string(filepath.Separator)) {
		return errors.Errorf("无效的同步配置路径: %s", path)
	}
	return nil
}

// copySyncItem 将单个文件或目录覆盖复制到目标路径。
func copySyncItem(src string, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return copySyncFile(src, dst, info.Mode())
	}

	return filepath.WalkDir(src, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(dst, relPath)
		if entry.IsDir() {
			info, err := entry.Info()
			if err != nil {
				return err
			}
			return os.MkdirAll(targetPath, info.Mode())
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		return copySyncFile(path, targetPath, info.Mode())
	})
}

// copySyncFile 覆盖复制文件并保留权限位。
func copySyncFile(src string, dst string, mode fs.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		dstFile.Close()
		return err
	}
	return dstFile.Close()
}
