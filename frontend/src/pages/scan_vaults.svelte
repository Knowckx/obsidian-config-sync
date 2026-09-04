<script lang="ts">
import { Dialogs } from '@wailsio/runtime';
import { Button, Input } from 'infa-s5';
import { saveLastVaultRoot, scanVaults, type VaultInfo } from '@/lib/api/vault_service';
import VaultList from '@/lib/components/vault_list.svelte';

type Props = {
  root?: string;
  vaults?: VaultInfo[];
  initializing?: boolean;
  onScanned?: (root: string, vaults: VaultInfo[]) => void;
  onDevPreset?: () => void | Promise<void>;
};

let {
  root = '',
  vaults = [],
  initializing = false,
  onScanned = () => {},
  onDevPreset = () => {},
}: Props = $props();
let error = $state('');
let scanning = $state(false);
let applyingDevPreset = $state(false);

const chooseAndScan = async () => {
  error = '';
  const selected = await Dialogs.OpenFile({
    Title: '选择 Obsidian 目录',
    ButtonText: '选择',
    CanChooseDirectories: true,
    CanChooseFiles: false,
  });

  if (!selected) {
    return;
  }

  scanning = true;
  try {
    const foundVaults = await scanVaults(selected);
    onScanned(selected, foundVaults);
    if (foundVaults.length > 0) {
      try {
        await saveLastVaultRoot(selected);
      } catch (err) {
        error = `Vault 已加载，但无法记忆目录：${getErrMsg(err)}`;
      }
    }
  } catch (err) {
    error = getErrMsg(err);
    onScanned(selected, []);
  } finally {
    scanning = false;
  }
};

const getErrMsg = (err: unknown): string => {
  return err instanceof Error ? err.message : String(err);
};

const applyDevPreset = async () => {
  applyingDevPreset = true;
  try {
    await onDevPreset();
  } finally {
    applyingDevPreset = false;
  }
};
</script>

<div class="step-content">
  <div class="toolbar">
    <div class="actions">
      <Button onclick={chooseAndScan} disabled={scanning || initializing}>
        {initializing ? '正在恢复' : scanning ? '扫描中' : '选择目录'}
      </Button>
      {#if import.meta.env.DEV}
        <Button onclick={applyDevPreset} disabled={applyingDevPreset || initializing}>
          {applyingDevPreset ? '正在进入 M3…' : '开发：快速进入 M3'}
        </Button>
      {/if}
    </div>
    <Input value={root} readonly placeholder="未选择目录" />
  </div>

  {#if error}
    <p class="status-error">{error}</p>
  {/if}

  <VaultList {vaults} />
</div>

<style>
  .toolbar {
    display: grid;
    grid-template-columns: max-content minmax(0, 1fr);
    gap: var(--space-3);
    align-items: center;
  }

  .actions {
    display: flex;
    gap: var(--space-2);
  }

  @media (max-width: 640px) {
    .toolbar {
      grid-template-columns: 1fr;
    }

    .actions {
      flex-wrap: wrap;
    }
  }
</style>
