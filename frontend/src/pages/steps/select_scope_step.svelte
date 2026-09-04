<script lang="ts">
import { Button } from 'infa-s5';
import {
  listConfigItems,
  openVaultConfigDir,
  type ConfigItem,
  type VaultInfo,
} from '@/lib/api/vault_service';

type Props = {
  mainVault?: VaultInfo | null;
  configItems?: ConfigItem[];
  selectedPaths?: string[];
  onConfigItemsChange?: (items: ConfigItem[]) => void;
  onSelectedPathsChange?: (paths: string[]) => void;
};

let {
  mainVault = null,
  configItems = [],
  selectedPaths = [],
  onConfigItemsChange = () => {},
  onSelectedPathsChange = () => {},
}: Props = $props();

let error = $state('');
let loading = $state(false);
let loadedVaultPath = $state('');
let regularConfigItems = $derived(configItems.filter((item) => !item.isPlugin));
let selectedConfigItems = $derived(
  regularConfigItems.filter((item) => selectedPaths.includes(item.path)),
);
let skippedConfigItems = $derived(
  regularConfigItems.filter((item) => !selectedPaths.includes(item.path)),
);
let pluginConfigItems = $derived(configItems.filter((item) => item.isPlugin));
let selectedPluginCount = $derived(
  pluginConfigItems.filter((item) => selectedPaths.includes(item.path)).length,
);

$effect(() => {
  if (!mainVault || configItems.length > 0 || loadedVaultPath === mainVault.path) {
    return;
  }
  loadConfigItems();
});

const loadConfigItems = async () => {
  if (!mainVault) {
    return;
  }

  error = '';
  loading = true;
  const vaultPath = mainVault.path;
  loadedVaultPath = vaultPath;

  try {
    const items = await listConfigItems(vaultPath);
    onConfigItemsChange(items);
    if (selectedPaths.length === 0) {
      onSelectedPathsChange(items.filter((item) => item.defaultSelected).map((item) => item.path));
    }
  } catch (err) {
    error = getErrMsg(err);
  } finally {
    loading = false;
  }
};

const togglePath = (path: string) => {
  const exists = selectedPaths.includes(path);
  onSelectedPathsChange(
    exists ? selectedPaths.filter((item) => item !== path) : [...selectedPaths, path],
  );
};

const clearSelection = () => {
  onSelectedPathsChange([]);
};

const selectDefault = () => {
  onSelectedPathsChange(
    configItems.filter((item) => item.defaultSelected).map((item) => item.path),
  );
};

const openConfigDir = async () => {
  if (!mainVault) {
    return;
  }

  error = '';
  try {
    await openVaultConfigDir(mainVault.path);
  } catch (err) {
    error = getErrMsg(err);
  }
};

const getErrMsg = (err: unknown): string => {
  return err instanceof Error ? err.message : String(err);
};

const getItemDescription = (item: ConfigItem): string => {
  if (!item.isPlugin || !item.isDir) {
    return item.description;
  }
  return item.version
    ? `社区插件 · ${item.name} · v${item.version}`
    : `社区插件 · ${item.name}`;
};
</script>

{#snippet configItem(item: ConfigItem)}
  <li>
    <label class="item-name">
      <input
        type="checkbox"
        checked={selectedPaths.includes(item.path)}
        onchange={() => togglePath(item.path)}
      />
      <span>{item.path}</span>
    </label>
    <span class="item-description">{getItemDescription(item)}</span>
    <span class="item-reserved" aria-hidden="true"></span>
  </li>
{/snippet}

<div class="step-content">
  <div class="header">
    <div>
      <h2>选择同步范围</h2>
      <p class="vault-context">
        <span class="vault-role">主库</span>
        <strong title={mainVault?.path ?? ''}>{mainVault?.name ?? ''}</strong>
        <span>的 .obsidian 配置项</span>
      </p>
    </div>
    <div class="actions">
      <Button onclick={openConfigDir} disabled={!mainVault}>打开配置文件夹</Button>
      <Button onclick={selectDefault} disabled={loading || configItems.length === 0}>默认</Button>
      <Button onclick={clearSelection} disabled={loading || selectedPaths.length === 0}>全不选</Button>
    </div>
  </div>

  {#if loading}
    <p class="muted">加载中</p>
  {:else if error}
    <p class="status-error">{error}</p>
  {:else if configItems.length === 0}
    <p class="muted">未发现配置项</p>
  {:else}
    <div class="config-groups">
      <section class="config-group">
        <h3>同步 <span>{selectedConfigItems.length}</span></h3>
        {#if selectedConfigItems.length === 0}
          <p class="muted">暂无同步项</p>
        {:else}
          <ul class="config-list">
            {#each selectedConfigItems as item (item.path)}
              {@render configItem(item)}
            {/each}
          </ul>
        {/if}
      </section>

      <section class="config-group skipped-group">
        <h3>跳过 <span>{skippedConfigItems.length}</span></h3>
        {#if skippedConfigItems.length === 0}
          <p class="muted">暂无跳过项</p>
        {:else}
          <ul class="config-list">
            {#each skippedConfigItems as item (item.path)}
              {@render configItem(item)}
            {/each}
          </ul>
        {/if}
      </section>

      {#if pluginConfigItems.length > 0}
        <section class="config-group plugin-group">
          <div>
            <h3>
              社区插件
              <span>已选择 {selectedPluginCount} / 共 {pluginConfigItems.length}</span>
            </h3>
            <p>同步插件会复制程序和设置，并在目标库中启用插件。</p>
          </div>
          <ul class="config-list">
            {#each pluginConfigItems as item (item.path)}
              {@render configItem(item)}
            {/each}
          </ul>
        </section>
      {/if}
    </div>
  {/if}
</div>

<style>
  .header {
    display: flex;
    justify-content: space-between;
    gap: var(--space-4);
  }

  .actions {
    display: flex;
    gap: var(--space-2);
  }

  h2,
  p {
    margin: 0;
  }

  p {
    color: var(--color-text-muted);
  }

  .vault-context {
    display: flex;
    gap: 6px;
    align-items: center;
  }

  .vault-context strong {
    color: var(--color-text);
  }

  .vault-role {
    padding: 3px var(--space-2);
    border-radius: var(--radius-sm);
    background: var(--color-surface-muted);
    color: var(--color-text-subtle);
    font-size: var(--font-size-sm);
    font-weight: 500;
  }

  .config-groups,
  .config-group {
    display: grid;
    gap: var(--space-3);
  }

  .skipped-group,
  .plugin-group {
    padding-top: var(--space-4);
    border-top: 1px solid var(--color-border);
  }

  h3 {
    display: flex;
    gap: var(--space-2);
    align-items: center;
    margin: 0;
    font-size: 15px;
  }

  h3 span {
    color: var(--color-text-muted);
    font-size: var(--font-size-sm);
    font-weight: 400;
  }

  .config-list {
    display: grid;
    gap: var(--space-2);
    list-style: none;
    margin: 0;
    padding: 0;
  }

  li {
    display: grid;
    grid-template-columns: minmax(180px, 1fr) minmax(240px, 2fr) minmax(120px, 1fr);
    gap: var(--space-3);
    align-items: center;
    padding: 10px var(--space-3);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-control);
  }

  .item-name {
    display: flex;
    gap: 10px;
    align-items: center;
    min-width: 0;
    overflow-wrap: anywhere;
  }

  input {
    width: 16px;
    height: 16px;
  }

  .item-description {
    color: var(--color-text-muted);
    overflow-wrap: anywhere;
  }

  .muted {
    color: var(--color-text-muted);
  }

  @media (max-width: 640px) {
    .header,
    li {
      grid-template-columns: 1fr;
    }

    .header {
      display: grid;
    }

    .item-reserved {
      display: none;
    }
  }
</style>
