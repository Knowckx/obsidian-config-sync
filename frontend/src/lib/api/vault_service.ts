import { DevService, VaultService } from '@bindings/obsi-conf-sync/go_src/inner/svc';
import {
  SyncPlanAction,
  SyncPlan as SyncPlanModel,
  SyncRequest as SyncRequestModel,
  SyncResultStatus,
} from '@bindings/obsi-conf-sync/go_src/inner/svc/models';

export { SyncPlanAction, SyncResultStatus };

export type VaultInfo = {
  path: string;
  name: string;
};

export type ConfigItem = {
  path: string;
  name: string;
  version: string;
  isDir: boolean;
  description: string;
  defaultSelected: boolean;
  isPlugin: boolean;
};

export type SyncRequest = {
  mainVaultPath: string;
  targetVaultPaths: string[];
  selectedPaths: string[];
};

export type SyncPlanItem = {
  path: string;
  action: SyncPlanAction;
};

export type TargetSyncPlan = {
  vaultPath: string;
  items: SyncPlanItem[];
};

export type SyncPlan = {
  mainVaultPath: string;
  targets: TargetSyncPlan[];
};

export type SyncResultItem = {
  path: string;
  status: SyncResultStatus;
  error: string;
};

export type TargetSyncResult = {
  vaultPath: string;
  items: SyncResultItem[];
};

export type SyncResult = {
  backupPath: string;
  targets: TargetSyncResult[];
};

export const scanVaults = (root: string): Promise<VaultInfo[]> => {
  return VaultService.ScanVaults(root);
};

export const getLastVaultRoot = (): Promise<string> => {
  return VaultService.GetLastVaultRoot();
};

export const saveLastVaultRoot = (root: string): Promise<void> => {
  return VaultService.SaveLastVaultRoot(root);
};

export const listConfigItems = (vaultPath: string): Promise<ConfigItem[]> => {
  return VaultService.ListConfigItems(vaultPath);
};

export const openVaultConfigDir = (vaultPath: string): Promise<void> => {
  return VaultService.OpenVaultConfigDir(vaultPath);
};

// 2026-08-30 备份功能已冻结，暂时停止实际调用。
// export const openSyncBackupDir = (backupPath: string): Promise<void> => {
//   return VaultService.OpenSyncBackupDir(backupPath);
// };

export const resetTestCases = (): Promise<void> => {
  return DevService.ResetTestCases();
};

export const removeDirectory = (): Promise<void> => {
  return DevService.RemoveDirectory();
};

export const buildSyncPlan = (request: SyncRequest): Promise<SyncPlan> => {
  return VaultService.BuildSyncPlan(new SyncRequestModel(request));
};

export const executeSyncPlan = (plan: SyncPlan): Promise<SyncResult> => {
  return VaultService.ExecuteSyncPlan(new SyncPlanModel(plan));
};
