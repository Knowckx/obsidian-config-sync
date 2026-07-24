<script lang="ts">
import { Button } from 'infa-s5';
import {
  openSyncBackupDir,
  SyncResultStatus,
  type SyncResult,
  type SyncResultItem,
} from '@/lib/api/vault_service';

type Props = {
  result?: SyncResult | null;
};

let { result = null }: Props = $props();
let openingBackup = $state(false);
let backupError = $state('');

const countStatus = (items: SyncResultItem[], status: SyncResultStatus) => {
  return items.filter((item) => item.status === status).length;
};

const totalStatus = (status: SyncResultStatus) => {
  return result?.targets.reduce((total, target) => total + countStatus(target.items, status), 0) ?? 0;
};

const statusLabel = (status: SyncResultStatus) => {
  if (status === SyncResultStatus.SyncResultStatusCreated) {
    return '新增成功';
  }
  if (status === SyncResultStatus.SyncResultStatusOverwrote) {
    return '覆盖成功';
  }
  return '失败';
};

const openBackup = async () => {
  if (!result?.backupPath) {
    return;
  }

  openingBackup = true;
  backupError = '';
  try {
    await openSyncBackupDir(result.backupPath);
  } catch (err) {
    backupError = err instanceof Error ? err.message : String(err);
  } finally {
    openingBackup = false;
  }
};
</script>

<div class="step-content">
  <div class="header">
    <div>
      <h2>{totalStatus(SyncResultStatus.SyncResultStatusFailed) > 0 ? '同步完成，但有失败项' : '同步完成'}</h2>
      <p>以下是本次同步的实际执行结果。</p>
    </div>
  </div>

  {#if !result}
    <p class="muted">暂无执行结果</p>
  {:else}
    <div class="summary">
      <div class="summary-item created-summary">
        <strong>{totalStatus(SyncResultStatus.SyncResultStatusCreated)}</strong>
        <span>新增成功</span>
      </div>
      <div class="summary-item overwrite-summary">
        <strong>{totalStatus(SyncResultStatus.SyncResultStatusOverwrote)}</strong>
        <span>覆盖成功</span>
      </div>
      <div class:error-summary={totalStatus(SyncResultStatus.SyncResultStatusFailed) > 0} class="summary-item">
        <strong>{totalStatus(SyncResultStatus.SyncResultStatusFailed)}</strong>
        <span>失败</span>
      </div>
    </div>

    {#if result.backupPath}
      <div class="backup-card">
        <div class="backup-context">
          <strong>同步前备份已创建</strong>
          <span title={result.backupPath}>{result.backupPath}</span>
        </div>
        <Button onclick={openBackup} disabled={openingBackup}>
          {openingBackup ? '正在打开…' : '打开备份目录'}
        </Button>
      </div>
      {#if backupError}
        <p class="backup-error">{backupError}</p>
      {/if}
    {/if}

    <div class="target-list">
      {#each result.targets as target (target.vaultPath)}
        <section class="target-card">
          <div class="target-header">
            <div class="target-context">
              <span class="target-role">目标库</span>
              <h3 title={target.vaultPath}>{target.vaultPath}</h3>
            </div>
            <div class="target-summary">
              <span><strong>{countStatus(target.items, SyncResultStatus.SyncResultStatusCreated)}</strong> 项新增成功</span>
              <span><strong>{countStatus(target.items, SyncResultStatus.SyncResultStatusOverwrote)}</strong> 项覆盖成功</span>
              <span><strong>{countStatus(target.items, SyncResultStatus.SyncResultStatusFailed)}</strong> 项失败</span>
            </div>
          </div>

          <ul class="result-list">
            {#each target.items as item (item.path)}
              <li class="result-item">
                <div class="item-content">
                  <span class="item-path">{item.path}</span>
                  {#if item.status === SyncResultStatus.SyncResultStatusFailed && item.error}
                    <span class="item-error">{item.error}</span>
                  {/if}
                </div>
                <span
                  class="status-badge"
                  class:created-status={item.status === SyncResultStatus.SyncResultStatusCreated}
                  class:overwrite-status={item.status === SyncResultStatus.SyncResultStatusOverwrote}
                  class:failed-status={item.status === SyncResultStatus.SyncResultStatusFailed}
                >
                  {statusLabel(item.status)}
                </span>
              </li>
            {/each}
          </ul>
        </section>
      {/each}
    </div>
  {/if}
</div>

<style>
  .header,
  .target-list,
  .target-card {
    display: grid;
    gap: var(--space-3);
  }

  h2,
  h3,
  p {
    margin: 0;
  }

  p,
  .muted {
    color: var(--color-text-muted);
  }

  .summary {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: var(--space-3);
  }

  .summary-item {
    display: grid;
    gap: var(--space-1);
    padding: var(--space-3);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-control);
    background: var(--color-surface-muted);
  }

  .summary-item strong {
    font-size: 22px;
    line-height: 1;
  }

  .summary-item span {
    color: var(--color-text-muted);
    font-size: var(--font-size-sm);
  }

  .created-summary strong {
    color: var(--color-success);
  }

  .overwrite-summary strong {
    color: var(--color-primary-text);
  }

  .error-summary strong {
    color: var(--color-danger);
  }

  .backup-card {
    display: flex;
    justify-content: space-between;
    gap: var(--space-4);
    align-items: center;
    padding: var(--space-4);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-control);
    background: var(--color-surface-muted);
  }

  .backup-context {
    display: grid;
    gap: var(--space-1);
    min-width: 0;
  }

  .backup-context span {
    color: var(--color-text-muted);
    font-size: var(--font-size-sm);
    overflow-wrap: anywhere;
  }

  .backup-error {
    color: var(--color-danger);
  }

  .target-card {
    overflow: hidden;
    border: 1px solid var(--color-border);
    border-radius: var(--radius-control);
    background: var(--color-surface);
  }

  .target-header {
    display: flex;
    justify-content: space-between;
    gap: var(--space-4);
    align-items: center;
    border-bottom: 1px solid var(--color-border-subtle);
    background: var(--color-surface-muted);
    padding: var(--space-4);
  }

  .target-context {
    display: flex;
    gap: var(--space-2);
    align-items: center;
    min-width: 0;
  }

  .target-role,
  .target-summary span {
    flex: none;
    padding: 3px var(--space-2);
    border-radius: var(--radius-round);
    font-size: var(--font-size-sm);
  }

  .target-role {
    background: var(--color-primary-bg);
    color: var(--color-primary-text);
    font-weight: 500;
  }

  h3 {
    min-width: 0;
    overflow-wrap: anywhere;
  }

  .target-summary {
    display: flex;
    gap: var(--space-2);
    flex: none;
    color: var(--color-text-muted);
    font-size: var(--font-size-sm);
  }

  .target-summary span {
    border: 1px solid var(--color-border-subtle);
    border-radius: var(--radius-control);
    background: var(--color-surface);
  }

  .target-summary strong {
    margin-right: var(--space-1);
    color: var(--color-text);
  }

  .result-list {
    display: grid;
    gap: var(--space-2);
    margin: 0;
    padding: var(--space-4);
    list-style: none;
  }

  .result-item {
    display: flex;
    justify-content: space-between;
    gap: var(--space-3);
    align-items: flex-start;
    padding: 9px var(--space-3);
    border: 1px solid var(--color-border-subtle);
    border-radius: var(--radius-control);
    background: var(--color-surface-muted);
  }

  .item-content {
    display: grid;
    gap: var(--space-1);
    min-width: 0;
  }

  .item-path {
    overflow-wrap: anywhere;
  }

  .item-error {
    color: var(--color-danger);
    font-size: var(--font-size-sm);
    overflow-wrap: anywhere;
  }

  .status-badge {
    flex: none;
    padding: 3px var(--space-2);
    border-radius: var(--radius-round);
    font-size: var(--font-size-sm);
    font-weight: 500;
  }

  .created-status {
    background: var(--color-success-bg);
    color: var(--color-success);
  }

  .overwrite-status {
    background: var(--color-primary-bg);
    color: var(--color-primary-text);
  }

  .failed-status {
    background: color-mix(in srgb, var(--color-danger) 8%, var(--color-surface));
    color: var(--color-danger);
  }

  @media (max-width: 640px) {
    .summary {
      grid-template-columns: 1fr;
    }

    .result-item {
      display: grid;
    }

    .target-header {
      display: grid;
    }

    .backup-card {
      display: grid;
    }

    .target-summary {
      flex-wrap: wrap;
    }
  }
</style>
