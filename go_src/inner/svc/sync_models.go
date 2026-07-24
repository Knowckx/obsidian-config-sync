package svc

// SyncRequest 描述一次同步计划请求。
type SyncRequest struct {
	MainVaultPath string `json:"mainVaultPath"`
	// 从库
	TargetVaultPaths []string `json:"targetVaultPaths"`
	// 选择要同步的配置
	SelectedPaths []string `json:"selectedPaths"`
}

// SyncPlan 描述一次同步计划。
type SyncPlan struct {
	MainVaultPath string           `json:"mainVaultPath"`
	Targets       []TargetSyncPlan `json:"targets"`
}

// SyncPlanAction 描述同步项将在目标库执行的动作。
type SyncPlanAction string

const (
	// SyncPlanActionCreate 表示目标库中不存在该配置项。
	SyncPlanActionCreate SyncPlanAction = "create"
	// SyncPlanActionOverwrite 表示目标库中已经存在该配置项。
	SyncPlanActionOverwrite SyncPlanAction = "overwrite"
)

// SyncPlanItem 描述单个配置项的同步计划。
type SyncPlanItem struct {
	Path   string         `json:"path"`
	Action SyncPlanAction `json:"action"`
}

// TargetSyncPlan 描述单个目标库的同步计划。
type TargetSyncPlan struct {
	// 当前目标库的绝对路径
	VaultPath string `json:"vaultPath"`
	// 按主库选择顺序排列的同步项
	Items []SyncPlanItem `json:"items"`
}

// SyncResult 描述一次同步的执行结果。
type SyncResult struct {
	BackupPath string             `json:"backupPath"`
	Targets    []TargetSyncResult `json:"targets"`
}

// SyncResultStatus 描述单个同步项的最终结果。
type SyncResultStatus string

const (
	// SyncResultStatusCreated 表示新增配置成功。
	SyncResultStatusCreated SyncResultStatus = "created"
	// SyncResultStatusOverwrote 表示覆盖配置成功。
	SyncResultStatusOverwrote SyncResultStatus = "overwrote"
	// SyncResultStatusFailed 表示同步失败。
	SyncResultStatusFailed SyncResultStatus = "failed"
)

// SyncResultItem 描述单个同步项的执行结果。
type SyncResultItem struct {
	Path   string           `json:"path"`
	Status SyncResultStatus `json:"status"`
	Error  string           `json:"error"`
}

// TargetSyncResult 描述单个目标库的执行结果。
type TargetSyncResult struct {
	VaultPath string           `json:"vaultPath"`
	Items     []SyncResultItem `json:"items"`
}
