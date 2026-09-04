package svc_test

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	"obsi-conf-sync/go_src/inner/svc"
)

var coreSyncSelectedPaths = []string{
	"app.json",
	"snippets/",
	"themes/",
	"plugins/open-in-new-tab/",
	"plugins/open-tab-settings/",
}

// Test_CoreSyncRegression 验证同步 MVP 的核心后端流程和固定预期结果。
func Test_CoreSyncRegression(t *testing.T) {
	fixturesRoot := coreTestCasesRoot(t)
	runRoot := t.TempDir()
	for _, name := range []string{"vault1", "vault2", "vault3"} {
		copyTree(t, filepath.Join(fixturesRoot, name), filepath.Join(runRoot, name))
	}

	mainVault := filepath.Join(runRoot, "vault1")
	targetVault2 := filepath.Join(runRoot, "vault2")
	targetVault3 := filepath.Join(runRoot, "vault3")
	service := &svc.VaultService{}

	items, err := service.ListConfigItems(mainVault)
	if err != nil {
		t.Fatalf("ListConfigItems 失败: %v", err)
	}
	assertContainsPaths(t, configItemPaths(items), coreSyncSelectedPaths)
	assertNotContainsPath(t, configItemPaths(items), "community-plugins.json")

	plan, err := service.BuildSyncPlan(svc.SyncRequest{
		MainVaultPath:    mainVault,
		TargetVaultPaths: []string{targetVault2, targetVault3},
		SelectedPaths:    coreSyncSelectedPaths,
	})
	if err != nil {
		t.Fatalf("BuildSyncPlan 失败: %v", err)
	}
	if len(plan.Targets) != 2 {
		t.Fatalf("目标库数量不符: want 2, got %d", len(plan.Targets))
	}

	assertTargetPlan(t, plan.Targets[0], []svc.SyncPlanAction{
		svc.SyncPlanActionOverwrite,
		svc.SyncPlanActionCreate,
		svc.SyncPlanActionCreate,
		svc.SyncPlanActionCreate,
		svc.SyncPlanActionCreate,
	})
	assertTargetPlan(t, plan.Targets[1], []svc.SyncPlanAction{
		svc.SyncPlanActionOverwrite,
		svc.SyncPlanActionOverwrite,
		svc.SyncPlanActionOverwrite,
		svc.SyncPlanActionOverwrite,
		svc.SyncPlanActionCreate,
	})

	result, err := service.ExecuteSyncPlan(plan)
	if err != nil {
		t.Fatalf("ExecuteSyncPlan 失败: %v", err)
	}
	if len(result.Targets) != 2 {
		t.Fatalf("执行结果目标库数量不符: want 2, got %d", len(result.Targets))
	}

	vault2Result := result.Targets[0]
	assertTargetResult(t, vault2Result, []svc.SyncResultStatus{
		svc.SyncResultStatusOverwrote,
		svc.SyncResultStatusCreated,
		svc.SyncResultStatusCreated,
		svc.SyncResultStatusCreated,
		svc.SyncResultStatusCreated,
	}, nil)

	vault3Result := result.Targets[1]
	assertTargetResult(t, vault3Result, []svc.SyncResultStatus{
		svc.SyncResultStatusOverwrote,
		svc.SyncResultStatusOverwrote,
		svc.SyncResultStatusFailed,
		svc.SyncResultStatusOverwrote,
		svc.SyncResultStatusCreated,
	}, []string{"", "", "themes", "", ""})

	assertFileEqual(t, filepath.Join(mainVault, ".obsidian", "snippets", "table-spacing-fix.css"), filepath.Join(targetVault3, ".obsidian", "snippets", "table-spacing-fix.css"))
	assertFileExists(t, filepath.Join(targetVault3, ".obsidian", "snippets", "vscode_light.css"))
	assertFileExists(t, filepath.Join(targetVault3, ".obsidian", "snippets", "target-only.css"))
	assertRegularFile(t, filepath.Join(targetVault3, ".obsidian", "themes"))
	assertStringList(t, readStringList(t, filepath.Join(targetVault2, ".obsidian", "community-plugins.json")), []string{"open-in-new-tab", "open-tab-settings"})
	assertStringList(t, readStringList(t, filepath.Join(targetVault3, ".obsidian", "community-plugins.json")), []string{"solve", "open-tab-settings", "open-in-new-tab"})
}

// Test_InvalidCommunityPluginsPreserved 验证无效启用列表不会被覆盖，也不会复制插件目录。
func Test_InvalidCommunityPluginsPreserved(t *testing.T) {
	fixturesRoot := coreTestCasesRoot(t)
	runRoot := t.TempDir()
	mainVault := filepath.Join(runRoot, "vault1")
	targetVault := filepath.Join(runRoot, "vault2")
	copyTree(t, filepath.Join(fixturesRoot, "vault1"), mainVault)
	copyTree(t, filepath.Join(fixturesRoot, "vault2"), targetVault)

	communityPluginsPath := filepath.Join(targetVault, ".obsidian", "community-plugins.json")
	invalidContent := []byte(`{"invalid": true}`)
	if err := os.WriteFile(communityPluginsPath, invalidContent, 0o644); err != nil {
		t.Fatalf("写入无效社区插件列表失败: %v", err)
	}

	service := &svc.VaultService{}
	plan, err := service.BuildSyncPlan(svc.SyncRequest{
		MainVaultPath:    mainVault,
		TargetVaultPaths: []string{targetVault},
		SelectedPaths:    []string{"plugins/open-in-new-tab/"},
	})
	if err != nil {
		t.Fatalf("BuildSyncPlan 失败: %v", err)
	}
	result, err := service.ExecuteSyncPlan(plan)
	if err != nil {
		t.Fatalf("ExecuteSyncPlan 失败: %v", err)
	}
	if result.Targets[0].Items[0].Status != svc.SyncResultStatusFailed {
		t.Errorf("无效社区插件列表应导致插件同步失败: %s", result.Targets[0].Items[0].Status)
	}
	gotContent, err := os.ReadFile(communityPluginsPath)
	if err != nil {
		t.Fatalf("读取无效社区插件列表失败: %v", err)
	}
	if string(gotContent) != string(invalidContent) {
		t.Errorf("无效社区插件列表不应被覆盖: %s", gotContent)
	}
	pluginPath := filepath.Join(targetVault, ".obsidian", "plugins", "open-in-new-tab")
	if _, err := os.Stat(pluginPath); !os.IsNotExist(err) {
		t.Errorf("目标插件目录不应被创建: %s", pluginPath)
	}
}

func coreTestCasesRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("无法定位核心测试文件")
	}
	return filepath.Join(filepath.Dir(file), "..", "..", "..", "test_cases")
}

func copyTree(t *testing.T, source string, target string) {
	t.Helper()
	err := filepath.WalkDir(source, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relPath, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(target, relPath)
		if entry.IsDir() {
			return os.MkdirAll(targetPath, 0o755)
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			return err
		}
		return os.WriteFile(targetPath, data, info.Mode().Perm())
	})
	if err != nil {
		t.Fatalf("复制测试 vault 失败: %v", err)
	}
}

func configItemPaths(items []svc.ConfigItem) []string {
	paths := make([]string, 0, len(items))
	for _, item := range items {
		paths = append(paths, item.Path)
	}
	return paths
}

func assertContainsPaths(t *testing.T, got []string, want []string) {
	t.Helper()
	gotSet := make(map[string]struct{}, len(got))
	for _, path := range got {
		gotSet[path] = struct{}{}
	}
	for _, path := range want {
		if _, ok := gotSet[path]; !ok {
			t.Errorf("缺少配置路径: %s\n实际路径: %v", path, got)
		}
	}
}

func assertNotContainsPath(t *testing.T, got []string, unwanted string) {
	t.Helper()
	for _, path := range got {
		if path == unwanted {
			t.Errorf("不应包含配置路径: %s", unwanted)
		}
	}
}

func assertTargetPlan(t *testing.T, target svc.TargetSyncPlan, wantActions []svc.SyncPlanAction) {
	t.Helper()
	if len(target.Items) != len(coreSyncSelectedPaths) {
		t.Fatalf("%s 同步项数量不符: want %d, got %d", target.VaultPath, len(coreSyncSelectedPaths), len(target.Items))
	}
	for i, item := range target.Items {
		if item.Path != coreSyncSelectedPaths[i] || item.Action != wantActions[i] {
			t.Errorf("%s 第 %d 项不符: want %s/%s, got %s/%s", target.VaultPath, i, coreSyncSelectedPaths[i], wantActions[i], item.Path, item.Action)
		}
	}
}

func assertTargetResult(t *testing.T, target svc.TargetSyncResult, wantStatuses []svc.SyncResultStatus, wantErrorParts []string) {
	t.Helper()
	if len(target.Items) != len(coreSyncSelectedPaths) {
		t.Fatalf("%s 执行项数量不符: want %d, got %d", target.VaultPath, len(coreSyncSelectedPaths), len(target.Items))
	}
	for i, item := range target.Items {
		if item.Path != coreSyncSelectedPaths[i] || item.Status != wantStatuses[i] {
			t.Errorf("%s 第 %d 项不符: want %s/%s, got %s/%s", target.VaultPath, i, coreSyncSelectedPaths[i], wantStatuses[i], item.Path, item.Status)
		}
		if len(wantErrorParts) > 0 && !strings.Contains(item.Error, wantErrorParts[i]) {
			t.Errorf("%s 第 %d 项错误不符: want 包含 %q, got %q", target.VaultPath, i, wantErrorParts[i], item.Error)
		}
		if len(wantErrorParts) == 0 && item.Error != "" {
			t.Errorf("%s 第 %d 项不应有错误: %q", target.VaultPath, i, item.Error)
		}
	}
}

func assertPaths(t *testing.T, name string, got []string, want []string) {
	t.Helper()
	gotCopy := append([]string(nil), got...)
	wantCopy := append([]string(nil), want...)
	sort.Strings(gotCopy)
	sort.Strings(wantCopy)
	if strings.Join(gotCopy, "\n") != strings.Join(wantCopy, "\n") {
		t.Errorf("%s 不符:\nwant: %v\ngot:  %v", name, wantCopy, gotCopy)
	}
}

func readStringList(t *testing.T, path string) []string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("读取字符串列表失败: %s: %v", path, err)
	}
	var items []string
	if err := json.Unmarshal(data, &items); err != nil {
		t.Fatalf("解析字符串列表失败: %s: %v", path, err)
	}
	return items
}

func assertStringList(t *testing.T, got []string, want []string) {
	t.Helper()
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("字符串列表不符: want %v, got %v", want, got)
	}
}

func assertFileEqual(t *testing.T, wantPath string, gotPath string) {
	t.Helper()
	want, err := os.ReadFile(wantPath)
	if err != nil {
		t.Fatalf("读取期望文件失败: %v", err)
	}
	got, err := os.ReadFile(gotPath)
	if err != nil {
		t.Fatalf("读取实际文件失败: %v", err)
	}
	if string(got) != string(want) {
		t.Errorf("文件内容不符: %s", gotPath)
	}
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Errorf("文件不存在: %s: %v", path, err)
	}
}

func assertRegularFile(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("目标路径不存在: %s: %v", path, err)
	}
	if info.IsDir() {
		t.Errorf("目标路径应为文件: %s", path)
	}
}
