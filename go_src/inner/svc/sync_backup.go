// 2026-08-30 已冻结，暂时不要向本文件增加代码。

package svc

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
)

const syncBackupKeepCount = 10

var (
	backupDirNamePattern   = regexp.MustCompile(`^\d{8}-\d{6}-[0-9a-f]{4}$`)
	creatingDirNamePattern = regexp.MustCompile(`^\.creating-\d{8}-\d{6}-[0-9a-f]{4}$`)
	backupTargetIDPattern  = regexp.MustCompile(`^\d{3,}$`)
)

// OpenSyncBackupDir 使用系统文件管理器打开完整同步备份目录。
func (t *VaultService) OpenSyncBackupDir(backupPath string) error {
	root, err := syncBackupRoot()
	if err != nil {
		return errors.WithStack(err)
	}
	if err := precheckBackupRoot(root); err != nil {
		return errors.WithStack(err)
	}
	target, err := validateSyncBackupDir(root, backupPath)
	if err != nil {
		return errors.WithStack(err)
	}
	if err := openDir(target); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// backupManifest 记录一次完整同步备份。
type backupManifest struct {
	CreatedAt     string         `json:"createdAt"`
	MainVaultPath string         `json:"mainVaultPath"`
	Targets       []backupTarget `json:"targets"`
}

// backupTarget 记录一个从库的备份目录和配置项。
type backupTarget struct {
	ID        string       `json:"id"`
	VaultPath string       `json:"vaultPath"`
	Items     []backupItem `json:"items"`
}

// backupItem 记录一个已备份的覆盖项。
type backupItem struct {
	Path   string         `json:"path"`
	Action SyncPlanAction `json:"action"`
}

// syncBackupCandidate 记录可参与保留清理的完整备份。
type syncBackupCandidate struct {
	Path      string
	Name      string
	CreatedAt time.Time
}

// createSyncBackup 为当前同步计划创建覆盖项备份。
func createSyncBackup(plan SyncPlan) (backupPath string, err error) {
	if !hasOverwriteItem(plan) {
		return "", nil
	}

	root, err := syncBackupRoot()
	if err != nil {
		return "", errors.WithStack(err)
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return "", errors.WithStack(err)
	}
	if err := precheckBackupRoot(root); err != nil {
		return "", errors.WithStack(err)
	}
	if err := cleanupCreatingBackupDirs(root); err != nil {
		return "", errors.WithStack(err)
	}

	now := time.Now()
	suffix, err := shortUUID()
	if err != nil {
		return "", errors.WithStack(err)
	}
	backupID := fmt.Sprintf("%s-%s", now.Format("20060102-150405"), suffix)
	finalPath := filepath.Join(root, backupID)
	if _, err := os.Lstat(finalPath); err == nil {
		return "", errors.Errorf("同步备份目录已存在: %s", finalPath)
	} else if !os.IsNotExist(err) {
		return "", errors.WithStack(err)
	}

	tempPath := filepath.Join(root, ".creating-"+backupID)
	if err := os.Mkdir(tempPath, 0o755); err != nil {
		return "", errors.WithStack(err)
	}
	defer func() {
		if tempPath == "" {
			return
		}
		if removeErr := os.RemoveAll(tempPath); removeErr != nil {
			if err == nil {
				err = errors.WithStack(removeErr)
				return
			}
			err = errors.Wrapf(err, "清理临时备份目录失败: %v", removeErr)
		}
	}()

	manifest := backupManifest{
		CreatedAt:     now.Format(time.RFC3339Nano),
		MainVaultPath: plan.MainVaultPath,
		Targets:       make([]backupTarget, 0, len(plan.Targets)),
	}
	for i, target := range plan.Targets {
		targetBackup := backupTarget{
			ID:        fmt.Sprintf("%03d", i+1),
			VaultPath: target.VaultPath,
			Items:     make([]backupItem, 0, len(target.Items)),
		}
		obsidianRoot := filepath.Join(target.VaultPath, ".obsidian")
		for _, item := range target.Items {
			if item.Action != SyncPlanActionOverwrite {
				continue
			}
			if err := precheckBackupSourcePath(obsidianRoot, item.Path); err != nil {
				return "", errors.Wrapf(err, "检查备份配置项失败: %s", item.Path)
			}

			src := filepath.Join(obsidianRoot, filepath.FromSlash(item.Path))
			dst := filepath.Join(tempPath, "targets", targetBackup.ID, filepath.FromSlash(item.Path))
			if err := backupSyncItem(src, dst); err != nil {
				return "", errors.Wrapf(err, "备份配置项失败: %s", item.Path)
			}
			targetBackup.Items = append(targetBackup.Items, backupItem{
				Path:   item.Path,
				Action: item.Action,
			})
		}
		if len(targetBackup.Items) > 0 {
			manifest.Targets = append(manifest.Targets, targetBackup)
		}
	}

	if err := writeBackupManifest(tempPath, manifest); err != nil {
		return "", errors.WithStack(err)
	}
	if err := os.Rename(tempPath, finalPath); err != nil {
		return "", errors.Wrap(err, "完成同步备份失败")
	}
	tempPath = ""
	return finalPath, nil
}

// pruneSyncBackups 尝试将完整同步备份保留到指定数量。
func pruneSyncBackups(root string, keep int) error {
	if err := precheckBackupRoot(root); err != nil {
		return errors.WithStack(err)
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		return errors.WithStack(err)
	}

	backups := make([]syncBackupCandidate, 0, len(entries))
	for _, entry := range entries {
		if !backupDirNamePattern.MatchString(entry.Name()) {
			continue
		}
		path := filepath.Join(root, entry.Name())
		info, err := os.Lstat(path)
		if err != nil || info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			continue
		}
		manifest, err := readValidBackupManifest(path)
		if err != nil {
			continue
		}
		createdAt, err := time.Parse(time.RFC3339Nano, manifest.CreatedAt)
		if err != nil {
			continue
		}
		backups = append(backups, syncBackupCandidate{
			Path:      path,
			Name:      entry.Name(),
			CreatedAt: createdAt,
		})
	}

	sort.Slice(backups, func(i int, j int) bool {
		if backups[i].CreatedAt.Equal(backups[j].CreatedAt) {
			return backups[i].Name > backups[j].Name
		}
		return backups[i].CreatedAt.After(backups[j].CreatedAt)
	})
	if len(backups) <= keep {
		return nil
	}
	for _, backup := range backups[keep:] {
		ok, err := isDirectChild(root, backup.Path)
		if err != nil {
			return errors.WithStack(err)
		}
		if !ok {
			return errors.Errorf("无效的旧备份目录: %s", backup.Path)
		}
		if err := os.RemoveAll(backup.Path); err != nil {
			return errors.Wrapf(err, "删除旧备份失败: %s", backup.Path)
		}
	}
	return nil
}

// hasOverwriteItem 判断计划中是否包含需要备份的覆盖项。
func hasOverwriteItem(plan SyncPlan) bool {
	for _, target := range plan.Targets {
		for _, item := range target.Items {
			if item.Action == SyncPlanActionOverwrite {
				return true
			}
		}
	}
	return false
}

// syncBackupRoot 返回应用同步备份根目录。
func syncBackupRoot() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Abs(filepath.Join(cacheDir, "obsi-conf-sync", "backups"))
}

// precheckBackupRoot 检查备份根路径是普通目录而非符号链接。
func precheckBackupRoot(root string) error {
	info, err := os.Lstat(root)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return errors.Errorf("无效的同步备份根目录: %s", root)
	}
	return nil
}

// cleanupCreatingBackupDirs 清理备份根目录下符合规则的遗留临时目录。
func cleanupCreatingBackupDirs(root string) error {
	entries, err := os.ReadDir(root)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if !creatingDirNamePattern.MatchString(entry.Name()) {
			continue
		}
		path := filepath.Join(root, entry.Name())
		info, err := os.Lstat(path)
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			continue
		}
		ok, err := isDirectChild(root, path)
		if err != nil {
			return err
		}
		if !ok {
			return errors.Errorf("无效的临时备份目录: %s", path)
		}
		if err := os.RemoveAll(path); err != nil {
			return errors.Wrapf(err, "清理临时备份目录失败: %s", path)
		}
	}
	return nil
}

// precheckBackupSourcePath 检查备份项及其现有父路径不包含符号链接。
func precheckBackupSourcePath(obsidianRoot string, syncPath string) error {
	if err := precheckSyncPath(syncPath); err != nil {
		return err
	}

	current, err := filepath.Abs(obsidianRoot)
	if err != nil {
		return err
	}
	info, err := os.Lstat(current)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return errors.Errorf("禁止备份符号链接: %s", current)
	}
	if !info.IsDir() {
		return errors.Errorf("配置目录不是普通目录: %s", current)
	}

	parts := strings.Split(filepath.Clean(filepath.FromSlash(normalizeSyncPath(syncPath))), string(os.PathSeparator))
	for i, part := range parts {
		current = filepath.Join(current, part)
		info, err = os.Lstat(current)
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return errors.Errorf("禁止备份符号链接: %s", current)
		}
		if i < len(parts)-1 && !info.IsDir() {
			return errors.Errorf("备份配置项父路径不是目录: %s", current)
		}
	}
	return nil
}

// backupSyncItem 将从库配置项复制到临时备份目录。
func backupSyncItem(src string, dst string) error {
	info, err := os.Lstat(src)
	if err != nil {
		return errors.WithStack(err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return errors.Errorf("禁止备份符号链接: %s", src)
	}
	if info.IsDir() {
		return backupSyncDir(src, dst)
	}
	if !info.Mode().IsRegular() {
		return errors.Errorf("不支持的备份文件类型: %s", src)
	}
	return copyBackupFile(src, dst, info.Mode())
}

// backupSyncDir 递归复制目录，并拒绝符号链接和特殊文件。
func backupSyncDir(src string, dst string) error {
	return filepath.WalkDir(src, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		info, err := entry.Info()
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return errors.Errorf("禁止备份符号链接: %s", path)
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(dst, relPath)
		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode().Perm())
		}
		if !info.Mode().IsRegular() {
			return errors.Errorf("不支持的备份文件类型: %s", path)
		}
		return copyBackupFile(path, targetPath, info.Mode())
	})
}

// copyBackupFile 复制普通文件并禁止覆盖已有备份。
func copyBackupFile(src string, dst string, mode fs.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_EXCL|os.O_WRONLY, mode.Perm())
	if err != nil {
		return err
	}
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		dstFile.Close()
		return err
	}
	return dstFile.Close()
}

// writeBackupManifest 将 Manifest 写入临时备份目录。
func writeBackupManifest(tempPath string, manifest backupManifest) error {
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	file, err := os.OpenFile(
		filepath.Join(tempPath, "manifest.json"),
		os.O_CREATE|os.O_EXCL|os.O_WRONLY,
		0o644,
	)
	if err != nil {
		return err
	}
	if _, err := file.Write(data); err != nil {
		file.Close()
		return err
	}
	return file.Close()
}

// readValidBackupManifest 读取并校验完整备份 Manifest。
func readValidBackupManifest(backupPath string) (backupManifest, error) {
	manifestPath := filepath.Join(backupPath, "manifest.json")
	info, err := os.Lstat(manifestPath)
	if err != nil {
		return backupManifest{}, err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return backupManifest{}, errors.Errorf("无效的备份 Manifest: %s", manifestPath)
	}

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return backupManifest{}, err
	}
	var manifest backupManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return backupManifest{}, err
	}
	if _, err := time.Parse(time.RFC3339Nano, manifest.CreatedAt); err != nil {
		return backupManifest{}, err
	}
	if strings.TrimSpace(manifest.MainVaultPath) == "" || len(manifest.Targets) == 0 {
		return backupManifest{}, errors.New("备份 Manifest 缺少必要信息")
	}

	targetIDs := make(map[string]struct{}, len(manifest.Targets))
	for _, target := range manifest.Targets {
		if !backupTargetIDPattern.MatchString(target.ID) || strings.TrimSpace(target.VaultPath) == "" || len(target.Items) == 0 {
			return backupManifest{}, errors.New("备份 Manifest 包含无效目标")
		}
		if _, exists := targetIDs[target.ID]; exists {
			return backupManifest{}, errors.Errorf("备份 Manifest 目标重复: %s", target.ID)
		}
		targetIDs[target.ID] = struct{}{}

		targetPath := filepath.Join(backupPath, "targets", target.ID)
		targetInfo, err := os.Lstat(targetPath)
		if err != nil {
			return backupManifest{}, err
		}
		if targetInfo.Mode()&os.ModeSymlink != 0 || !targetInfo.IsDir() {
			return backupManifest{}, errors.Errorf("无效的备份目标目录: %s", targetPath)
		}
		for _, item := range target.Items {
			if item.Action != SyncPlanActionOverwrite {
				return backupManifest{}, errors.Errorf("无效的备份动作: %s", item.Action)
			}
			if err := precheckBackupSourcePath(targetPath, item.Path); err != nil {
				return backupManifest{}, err
			}
			itemPath := filepath.Join(targetPath, filepath.FromSlash(item.Path))
			itemInfo, err := os.Lstat(itemPath)
			if err != nil {
				return backupManifest{}, err
			}
			if !itemInfo.IsDir() && !itemInfo.Mode().IsRegular() {
				return backupManifest{}, errors.Errorf("无效的备份配置项: %s", itemPath)
			}
		}
	}
	return manifest, nil
}

// validateSyncBackupDir 校验待打开的目录是直属完整备份。
func validateSyncBackupDir(root string, input string) (string, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	root = filepath.Clean(root)

	target, err := filepath.Abs(input)
	if err != nil {
		return "", err
	}
	target = filepath.Clean(target)

	ok, err := isDirectChild(root, target)
	if err != nil {
		return "", err
	}
	if !ok || !backupDirNamePattern.MatchString(filepath.Base(target)) {
		return "", errors.New("无效的备份目录")
	}
	info, err := os.Lstat(target)
	if err != nil {
		return "", err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return "", errors.New("无效的备份目录")
	}
	if _, err := readValidBackupManifest(target); err != nil {
		return "", errors.Wrap(err, "备份目录不完整")
	}
	return target, nil
}

// isDirectChild 判断目标是否为根目录的直属子路径。
func isDirectChild(root string, target string) (bool, error) {
	rel, err := filepath.Rel(filepath.Clean(root), filepath.Clean(target))
	if err != nil {
		return false, err
	}
	if rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return false, nil
	}
	return filepath.Dir(rel) == ".", nil
}

// shortUUID 生成 4 位小写十六进制随机标识。
func shortUUID() (string, error) {
	var value [2]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", errors.WithStack(err)
	}
	return hex.EncodeToString(value[:]), nil
}
