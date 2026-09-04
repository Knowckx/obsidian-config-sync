<script lang="ts">
import type { VaultInfo } from '@/lib/api/vault_service';

type Props = {
  vaults?: VaultInfo[];
  mode?: 'read' | 'select';
  mainVault?: VaultInfo | null;
  targetVaults?: VaultInfo[];
  onMainChange?: (vault: VaultInfo) => void;
  onTargetToggle?: (vault: VaultInfo) => void;
  onRemove?: (vault: VaultInfo) => void;
};

let {
  vaults = [],
  mode = 'read',
  mainVault = null,
  targetVaults = [],
  onMainChange = () => {},
  onTargetToggle = () => {},
  onRemove,
}: Props = $props();

const isMainVault = (vault: VaultInfo) => mainVault?.path === vault.path;
const isTargetVault = (vault: VaultInfo) => targetVaults.some((item) => item.path === vault.path);
</script>

{#if vaults.length === 0}
  <p>未发现 Vault</p>
{:else}
  <ul class="vault-list">
    {#each vaults as vault (vault.path)}
      <li>
        <div class="vault-info">
          <strong>{vault.name}</strong>
          <span>{vault.path}</span>
        </div>

        {#if mode === 'select' || onRemove}
          <div class="actions">
            {#if mode === 'select'}
              <button class:active={isMainVault(vault)} onclick={() => onMainChange(vault)}>
                主库
              </button>
              <button
                class:active={isTargetVault(vault)}
                disabled={isMainVault(vault)}
                onclick={() => onTargetToggle(vault)}
              >
                从库
              </button>
            {/if}
            {#if onRemove}
              <button onclick={() => onRemove(vault)}>移除</button>
            {/if}
          </div>
        {/if}
      </li>
    {/each}
  </ul>
{/if}

<style>
  p {
    margin: 0;
  }

  .vault-list {
    display: grid;
    gap: var(--space-2);
    list-style: none;
    margin: 0;
    padding: 0;
  }

  li {
    display: grid;
    grid-template-columns: minmax(0, 1fr) max-content;
    gap: var(--space-3);
    align-items: center;
    padding: 10px var(--space-3);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-control);
  }

  .vault-info {
    display: grid;
    gap: var(--space-1);
    min-width: 0;
  }

  span {
    color: var(--color-text-subtle);
    overflow-wrap: anywhere;
  }

  .actions {
    display: flex;
    gap: var(--space-2);
  }

  button {
    min-width: 52px;
    height: 32px;
    border: 1px solid var(--color-border);
    border-radius: var(--radius-control);
    background: var(--color-surface);
    color: var(--color-text);
    cursor: pointer;
  }

  button.active {
    border-color: var(--color-primary);
    background: var(--color-primary-bg);
    color: var(--color-primary-text);
  }

  button:disabled {
    color: var(--color-text-disabled);
    cursor: not-allowed;
  }

  @media (max-width: 640px) {
    li {
      grid-template-columns: 1fr;
    }
  }
</style>
