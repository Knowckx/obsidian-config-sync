package svc

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"

	"obsi-conf-sync/go_src/conf"
	"obsi-conf-sync/go_src/inner/app_settings"

	"github.com/cockroachdb/errors"
)

const communityPluginsFileName = "community-plugins.json"

// VaultService 向前端暴露 vault 扫描接口。
type VaultService struct {
	isSyncing atomic.Bool // 是否正在执行同步
}

// VaultInfo 表示扫描发现的 Obsidian vault。
type VaultInfo struct {
	Path string `json:"path"`
	Name string `json:"name"`
}

// ConfigItem 表示 .obsidian 下可同步的配置项。
type ConfigItem struct {
	Path            string `json:"path"`
	Name            string `json:"name"`
	Version         string `json:"version"`
	IsDir           bool   `json:"isDir"`
	Description     string `json:"description"`
	DefaultSelected bool   `json:"defaultSelected"`
	IsPlugin        bool   `json:"isPlugin"`
}

// String 返回 vault 的简短展示文本。
func (t VaultInfo) String() string {
	out := fmt.Sprintf("Name:%s", t.Name)
	return out
}

// ScanVaults 扫描 root 并返回包含 .obsidian/ 的目录。
func (s *VaultService) ScanVaults(root string) ([]VaultInfo, error) {
	rootPath, err := precheckScanRoot(root)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	vaults, err := scanVaultRoot(rootPath)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return vaults, nil
}

// GetLastVaultRoot 返回最近一次加载成功的 Vault 扫描根目录。
func (s *VaultService) GetLastVaultRoot() (string, error) {
	root, err := app_settings.GetLastVaultRoot()
	if err != nil {
		return "", errors.WithStack(err)
	}
	return root, nil
}

// SaveLastVaultRoot 保存最近一次加载成功的 Vault 扫描根目录。
func (s *VaultService) SaveLastVaultRoot(root string) error {
	rootPath, err := precheckScanRoot(root)
	if err != nil {
		return errors.WithStack(err)
	}
	if err := app_settings.SaveLastVaultRoot(rootPath); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// ListConfigItems 列出 vault 的 .obsidian 下可选择同步的配置项。
func (s *VaultService) ListConfigItems(vaultPath string) ([]ConfigItem, error) {
	rootPath, err := precheckVaultPath(vaultPath)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	items, err := listConfigItems(filepath.Join(rootPath, ".obsidian"))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return items, nil
}

// OpenVaultConfigDir 使用系统文件管理器打开 vault 的 .obsidian 配置目录。
func (s *VaultService) OpenVaultConfigDir(vaultPath string) error {
	rootPath, err := precheckVaultPath(vaultPath)
	if err != nil {
		return errors.WithStack(err)
	}

	err = openDir(filepath.Join(rootPath, ".obsidian"))
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// precheckScanRoot 检查扫描根目录并返回绝对路径。
func precheckScanRoot(root string) (string, error) {
	rootPath, err := absoluteDirectoryPath(root)
	if err != nil {
		return "", err
	}

	info, err := os.Stat(rootPath)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", errors.Errorf("root 不是目录: %s", rootPath)
	}
	return rootPath, nil
}

// absoluteDirectoryPath 返回非空路径的绝对路径。
func absoluteDirectoryPath(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", errors.New("目录路径不能为空")
	}
	return filepath.Abs(path)
}

// precheckVaultPath 检查 vault 路径并返回绝对路径。
func precheckVaultPath(vaultPath string) (string, error) {
	rootPath, err := precheckScanRoot(vaultPath)
	if err != nil {
		return "", err
	}

	isVault, err := isVaultDir(rootPath)
	if err != nil {
		return "", err
	}
	if !isVault {
		return "", errors.Errorf("不是 Obsidian vault: %s", rootPath)
	}
	return rootPath, nil
}

// scanVaultRoot 扫描 root 本身或 root 的最多 2 层子目录。
func scanVaultRoot(rootPath string) ([]VaultInfo, error) {
	isVault, err := isVaultDir(rootPath)
	if err != nil {
		return nil, err
	}
	if isVault {
		return []VaultInfo{{
			Path: rootPath,
			Name: filepath.Base(rootPath),
		}}, nil
	}

	vaults := make([]VaultInfo, 0)
	err = scanVaultChildren(rootPath, 2, &vaults)
	if err != nil {
		return nil, err
	}
	return vaults, nil
}

// scanVaultChildren 扫描子目录，remainDepth 表示还允许向下检查的层数。
func scanVaultChildren(parentPath string, remainDepth int, vaults *[]VaultInfo) error {
	if remainDepth <= 0 {
		return nil
	}

	entries, err := os.ReadDir(parentPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if shouldSkipScanDir(entry.Name()) {
			continue
		}

		childPath := filepath.Join(parentPath, entry.Name())
		isVault, err := isVaultDir(childPath)
		if err != nil {
			return err
		}
		if isVault {
			*vaults = append(*vaults, VaultInfo{
				Path: childPath,
				Name: entry.Name(),
			})
			continue
		}

		if err := scanVaultChildren(childPath, remainDepth-1, vaults); err != nil {
			return err
		}
	}
	return nil
}

// listConfigItems 列出 .obsidian 根目录配置项和插件目录。
func listConfigItems(obsidianPath string) ([]ConfigItem, error) {
	entries, err := os.ReadDir(obsidianPath)
	if err != nil {
		return nil, err
	}

	items := make([]ConfigItem, 0)
	for _, entry := range entries {
		if entry.Name() == communityPluginsFileName {
			continue
		}
		if entry.Name() == "plugins" && entry.IsDir() {
			pluginItems, err := listPluginConfigItems(obsidianPath)
			if err != nil {
				return nil, err
			}
			items = append(items, pluginItems...)
			continue
		}

		path := entry.Name()
		if entry.IsDir() {
			path += "/"
		}
		items = append(items, newConfigItem(path, entry.Name(), entry.IsDir(), false))
	}
	return items, nil
}

// newConfigItem 创建配置项并应用已知规则。
func newConfigItem(path string, name string, isDir bool, isPlugin bool) ConfigItem {
	rule := conf.ConfigItemRules[path]
	if rule.IsPlugin {
		isPlugin = true
	}
	description := rule.Description
	if isPlugin && description == "" {
		description = "社区插件程序和设置"
	}
	return ConfigItem{
		Path:            path,
		Name:            name,
		IsDir:           isDir,
		Description:     description,
		DefaultSelected: rule.DefaultSelected,
		IsPlugin:        isPlugin,
	}
}

// listPluginConfigItems 列出 .obsidian/plugins 下的插件配置目录。
func listPluginConfigItems(obsidianPath string) ([]ConfigItem, error) {
	pluginsPath := filepath.Join(obsidianPath, "plugins")
	entries, err := os.ReadDir(pluginsPath)
	if err != nil {
		return nil, err
	}

	items := make([]ConfigItem, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		path := filepath.ToSlash(filepath.Join("plugins", entry.Name())) + "/"
		item := newConfigItem(path, entry.Name(), true, true)
		manifest := readPluginManifest(filepath.Join(pluginsPath, entry.Name(), "manifest.json"))
		if manifest.Name != "" {
			item.Name = manifest.Name
		}
		item.Version = manifest.Version
		items = append(items, item)
	}
	return items, nil
}

// pluginManifest 表示插件清单中用于展示的信息。
type pluginManifest struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Version string `json:"version"`
}

// readPluginManifest 读取插件清单，读取失败时返回零值。
func readPluginManifest(path string) pluginManifest {
	manifest, err := loadPluginManifest(path)
	if err != nil {
		return pluginManifest{}
	}
	return manifest
}

// loadPluginManifest 读取并解析插件清单。
func loadPluginManifest(path string) (pluginManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return pluginManifest{}, err
	}
	var manifest pluginManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return pluginManifest{}, err
	}
	return manifest, nil
}

// openDir 使用当前系统的文件管理器打开目录。
func openDir(dir string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", dir)
	case "darwin":
		cmd = exec.Command("open", dir)
	case "linux":
		cmd = exec.Command("xdg-open", dir)
	default:
		return errors.Errorf("不支持打开目录的系统: %s", runtime.GOOS)
	}
	return cmd.Start()
}

// isVaultDir 判断目录是否为 Obsidian vault。
func isVaultDir(dir string) (bool, error) {
	obsidianDir := filepath.Join(dir, ".obsidian")
	info, err := os.Stat(obsidianDir)
	if err == nil {
		return info.IsDir(), nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// shouldSkipScanDir 判断目录是否应跳过扫描。
func shouldSkipScanDir(name string) bool {
	if _, ok := conf.SkipVaultScanDirNames[name]; ok {
		return true
	}
	// 目录名以 . 开头，就认为是隐藏目录，跳过扫描
	return strings.HasPrefix(name, ".")
}
