<script lang="ts">
import { onMount } from 'svelte';
import { Button } from '@knowckx/infa-s5';
import {
  buildSyncPlan,
  executeSyncPlan,
  getLastVaultRoot,
  removeDirectory,
  resetTestCases,
  scanVaults,
  type ConfigItem,
  type SyncPlan,
  type SyncResult,
  type VaultInfo,
} from '@/lib/api/vault_service';
import Card from '@/lib/infa_s5_candidates/card.svelte';
import Page from '@/lib/infa_s5_candidates/page.svelte';
import Section from '@/lib/infa_s5_candidates/section.svelte';
import ScanVaults from './scan_vaults.svelte';
import SelectScopeStep from './steps/select_scope_step.svelte';
import SelectVaultsStep from './steps/select_vaults_step.svelte';
import SyncResultStep from './steps/sync_result_step.svelte';
import SyncPlanStep from './steps/sync_plan_step.svelte';
import StepNav from '@/lib/components/step_nav.svelte';

type StepKey = 'scan' | 'vaults' | 'scope' | 'plan' | 'result';

type WizardStep = {
  key: StepKey;
  label: string;
};

const steps: WizardStep[] = [
  { key: 'scan', label: '扫描库' },
  { key: 'vaults', label: '选择库' },
  { key: 'scope', label: '同步范围' },
  { key: 'plan', label: '同步计划' },
  { key: 'result', label: '执行结果' },
];

const devSelectedPaths = [
  'app.json',
  'snippets/',
  'themes/',
  'plugins/open-in-new-tab/',
  'plugins/open-tab-settings/',
];
const devTargetRoot = 'temp/test_cases_1';

let stepIndex = $state(0);
let root = $state('');
let vaults = $state<VaultInfo[]>([]);
let mainVault = $state<VaultInfo | null>(null);
let targetVaults = $state<VaultInfo[]>([]);
let configItems = $state<ConfigItem[]>([]);
let selectedPaths = $state<string[]>([]);
let syncPlan = $state<SyncPlan | null>(null);
let planLoading = $state(false);
let planError = $state('');
let syncResult = $state<SyncResult | null>(null);
let syncing = $state(false);
let executeError = $state('');
let devPresetError = $state('');
let devTestRoot = $state<string | null>(null);
let initializing = $state(true);
let restoreError = $state('');

let currentStep = $derived(steps[stepIndex]!);
let canBack = $derived(stepIndex > 0 && !syncing);
let canNext = $derived(getCanNext());
let nextLabel = $derived(
  currentStep.key === 'plan'
    ? syncing
      ? '正在同步…'
      : '确认同步'
    : currentStep.key === 'result'
      ? '完成'
      : '下一步',
);

const setScannedVaults = (selectedRoot: string, foundVaults: VaultInfo[]) => {
  root = selectedRoot;
  const vaultMap = new Map(vaults.map((vault) => [vault.path, vault]));
  for (const vault of foundVaults) {
    vaultMap.set(vault.path, vault);
  }
  vaults = [...vaultMap.values()];
};

// 启动时重新扫描上次加载成功的目录，避免使用可能过期的 Vault 列表。
const restoreLastVaultRoot = async () => {
  try {
    const lastRoot = await getLastVaultRoot();
    if (!lastRoot) {
      return;
    }

    root = lastRoot;
    const foundVaults = await scanVaults(lastRoot);
    if (foundVaults.length === 0) {
      restoreError = '上次选择的目录中未发现 Vault，请重新选择。';
      return;
    }
    setScannedVaults(lastRoot, foundVaults);
  } catch (err) {
    restoreError = `上次选择的目录不可用，请重新选择：${err instanceof Error ? err.message : String(err)}`;
  } finally {
    initializing = false;
  }
};

const handleScannedVaults = (selectedRoot: string, foundVaults: VaultInfo[]) => {
  restoreError = '';
  setScannedVaults(selectedRoot, foundVaults);
};

// 从当前同步候选列表中移除 Vault，不操作磁盘目录。
const removeVault = (vault: VaultInfo) => {
  vaults = vaults.filter((item) => item.path !== vault.path);
};

onMount(() => {
  void restoreLastVaultRoot();
});

// 返回扫描步骤后，后续同步流程重新开始。
const resetFollowingSteps = () => {
  mainVault = null;
  targetVaults = [];
  configItems = [];
  selectedPaths = [];
  syncPlan = null;
  planLoading = false;
  planError = '';
  syncResult = null;
  syncing = false;
  executeError = '';
};

const setMainVault = (vault: VaultInfo) => {
  mainVault = vault;
  targetVaults = vaults.filter((item) => item.path !== vault.path);
  configItems = [];
  selectedPaths = [];
  syncPlan = null;
  planError = '';
  syncResult = null;
  executeError = '';
};

// 开发环境按本机预设进入同步范围。
const applyDevPreset = async () => {
  if (!import.meta.env.DEV) {
    return;
  }

  devPresetError = '';

  try {
    await resetTestCases();
    devTestRoot = devTargetRoot;

    const foundVaults = await scanVaults(devTargetRoot);
    setScannedVaults(devTargetRoot, foundVaults);

    const selectedMainVault = foundVaults.find((vault) => vault.name === 'vault1');
    if (!selectedMainVault) {
      throw new Error('开发预设未找到主库 vault1');
    }

    setMainVault(selectedMainVault);
    if (targetVaults.length === 0) {
      throw new Error('开发预设没有可用的从库');
    }

    selectedPaths = [...devSelectedPaths];
    stepIndex = 2;
  } catch (err) {
    await removeDirectory();
    devTestRoot = null;
    devPresetError = err instanceof Error ? err.message : String(err);
    stepIndex = 0;
  }
};

const toggleTargetVault = (vault: VaultInfo) => {
  if (mainVault?.path === vault.path) {
    return;
  }

  const exists = targetVaults.some((item) => item.path === vault.path);
  targetVaults = exists
    ? targetVaults.filter((item) => item.path !== vault.path)
    : [...targetVaults, vault];
};

const goBack = () => {
  if (!canBack) {
    return;
  }

  stepIndex -= 1;
  if (stepIndex === 0) {
    resetFollowingSteps();
  }
};

const goNext = async () => {
  if (!canNext) {
    return;
  }

  if (currentStep.key === 'result') {
    await finishSync();
    return;
  }

  if (currentStep.key === 'scope') {
    await loadSyncPlan();
    if (!syncPlan) {
      return;
    }
  }
  if (currentStep.key === 'plan') {
    await executePlan();
    if (!syncResult) {
      return;
    }
  }
  stepIndex += 1;
};

// 完成本轮同步，保留扫描结果并清空后续步骤状态。
const finishSync = async () => {
  if (devTestRoot) {
    await removeDirectory();
    devTestRoot = null;
  }
  resetFollowingSteps();
  stepIndex = 0;
};

const loadSyncPlan = async () => {
  if (!mainVault || targetVaults.length === 0) {
    return;
  }

  planLoading = true;
  planError = '';
  syncPlan = null;
  try {
    syncPlan = await buildSyncPlan({
      mainVaultPath: mainVault.path,
      targetVaultPaths: targetVaults.map((vault) => vault.path),
      selectedPaths,
    });
  } catch (err) {
    planError = err instanceof Error ? err.message : String(err);
  } finally {
    planLoading = false;
  }
};

const executePlan = async () => {
  if (!syncPlan) {
    return;
  }

  syncing = true;
  executeError = '';
  syncResult = null;
  try {
    syncResult = await executeSyncPlan(syncPlan);
  } catch (err) {
    executeError = err instanceof Error ? err.message : String(err);
  } finally {
    syncing = false;
  }
};

function getCanNext(): boolean {
  if (currentStep.key === 'scan') {
    return vaults.length >= 2;
  }

  if (currentStep.key === 'vaults') {
    return Boolean(mainVault && targetVaults.length > 0);
  }

  if (currentStep.key === 'scope') {
    return selectedPaths.length > 0;
  }

  if (currentStep.key === 'plan') {
    return Boolean(syncPlan && !syncing);
  }

  return Boolean(syncResult);
}
</script>

<!-- <PanelBg>：应用级背景已由 AppLayout 负责。 -->
  <div class="panel-body">
    <!-- <ContentShell maxWidth="max-w-6xl">：页面容器已由 Page 负责。 -->
    <Page>
      <div class="layout">
        <StepNav {steps} currentKey={currentStep.key} />

        <Section>
          <!-- @knowckx/infa-s5: <Card> -->
          <Card>
            {#if devPresetError}
              <p class="status-error dev-preset-error">{devPresetError}</p>
            {/if}
            {#if restoreError}
              <p class="status-error restore-error">{restoreError}</p>
            {/if}

            <!-- <section class="step-body">：页面区块已由 Section 负责。 -->
            <div class="step-body">
              {#if currentStep.key === 'scan'}
                <ScanVaults
                  {root}
                  {vaults}
                  {initializing}
                  onScanned={handleScannedVaults}
                  onVaultRemove={removeVault}
                  onDevPreset={applyDevPreset}
                />
              {:else if currentStep.key === 'vaults'}
                <SelectVaultsStep
                  {vaults}
                  {mainVault}
                  {targetVaults}
                  onMainChange={setMainVault}
                  onTargetToggle={toggleTargetVault}
                />
              {:else if currentStep.key === 'scope'}
                <SelectScopeStep
                  {mainVault}
                  {configItems}
                  {selectedPaths}
                  onConfigItemsChange={(items) => (configItems = items)}
                  onSelectedPathsChange={(paths) => (selectedPaths = paths)}
                />
              {:else if currentStep.key === 'plan'}
                <SyncPlanStep plan={syncPlan} loading={planLoading} error={planError} />
                {#if executeError}
                  <p class="status-error execute-error">{executeError}</p>
                {/if}
              {:else if currentStep.key === 'result' && syncResult}
                <SyncResultStep result={syncResult} />
              {:else}
                <div class="pending-step">
                  <h2>{currentStep.label}</h2>
                  <p>等待对应后端接口补齐。</p>
                </div>
              {/if}
            </div>
            <!-- </section> -->

            <div class="footer">
              {#if currentStep.key !== 'result'}
                <Button onclick={goBack} disabled={!canBack}>上一步</Button>
              {/if}
              <Button onclick={goNext} disabled={!canNext}>{nextLabel}</Button>
            </div>
          </Card>
          <!-- @knowckx/infa-s5: </Card> -->
        </Section>
      </div>
    </Page>
    <!-- </ContentShell> -->
  </div>
<!-- </PanelBg> -->

<style>
  .panel-body {
    display: grid;
    min-height: calc(100vh - 5rem);
  }

  .layout {
    display: grid;
    flex: 1;
    grid-template-columns: 160px minmax(0, 1fr);
    gap: var(--space-4);
  }

  .step-body {
    min-height: 360px;
  }

  .dev-preset-error {
    margin-bottom: var(--space-4);
  }

  .restore-error {
    margin-bottom: var(--space-4);
  }

  .execute-error {
    margin-top: var(--space-4);
  }

  .pending-step {
    display: grid;
    gap: var(--space-2);
  }

  .pending-step h2,
  .pending-step p {
    margin: 0;
  }

  .footer {
    display: flex;
    justify-content: flex-end;
    gap: 10px;
    margin-top: var(--space-5);
  }

  @media (max-width: 760px) {
    .layout {
      grid-template-columns: 1fr;
    }
  }
</style>
