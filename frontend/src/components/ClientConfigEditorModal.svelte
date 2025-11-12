<script lang="ts">
  import { onMount } from 'svelte';
  import type { MCPServer } from '../stores/stores';
  import { addNotification } from '../stores/stores';
  import * as WailsApp from '../../wailsjs/go/main/App';

  // Props
  export let server: MCPServer;
  export let onClose: () => void;

  // Client info type
  interface ClientInfo {
    type: string;
    name: string;
    configPath: string;
    installed: boolean;
  }

  // State
  let clients: ClientInfo[] = [];
  let selectedClient: ClientInfo | null = null;
  let loading = true;
  let saving = false;
  let errorMessage = '';

  // Server entry fields
  let serverName = server.name;
  let command = 'node';
  let args: string[] = [];
  let envVars: Record<string, string> = {};
  let newArg = '';
  let newEnvKey = '';
  let newEnvValue = '';

  onMount(async () => {
    await detectClients();
  });

  async function detectClients() {
    loading = true;
    errorMessage = '';
    try {
      clients = await WailsApp.DetectClients();
      // Select first installed client
      selectedClient = clients.find(c => c.installed) || clients[0];

      // Try to parse command from server's installation path
      if (server.installationPath) {
        const parts = server.installationPath.split(' ');
        if (parts.length > 0) {
          command = parts[0];
          args = parts.slice(1);
        }
      }

      loading = false;
    } catch (error: any) {
      errorMessage = `Failed to detect clients: ${error.message}`;
      loading = false;
    }
  }

  async function addToClient() {
    if (!selectedClient) {
      errorMessage = 'No client selected';
      return;
    }

    if (!serverName.trim()) {
      errorMessage = 'Server name cannot be empty';
      return;
    }

    saving = true;
    errorMessage = '';

    try {
      await WailsApp.AddServerToClientConfig(
        selectedClient.configPath,
        serverName,
        command,
        args,
        envVars
      );

      addNotification('success', `Added ${serverName} to ${selectedClient.name} config`);
      onClose();
    } catch (error: any) {
      errorMessage = error.message || 'Failed to add server to config';
    } finally {
      saving = false;
    }
  }

  function addArg() {
    if (newArg.trim()) {
      args = [...args, newArg.trim()];
      newArg = '';
    }
  }

  function removeArg(index: number) {
    args = args.filter((_, i) => i !== index);
  }

  function addEnvVar() {
    if (newEnvKey.trim()) {
      envVars = { ...envVars, [newEnvKey.trim()]: newEnvValue };
      newEnvKey = '';
      newEnvValue = '';
    }
  }

  function removeEnvVar(key: string) {
    const updated = { ...envVars };
    delete updated[key];
    envVars = updated;
  }

  function handleKeydown(event: KeyboardEvent) {
    if (event.key === 'Escape') {
      onClose();
    }
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<div class="modal-backdrop" on:click={onClose} role="presentation">
  <div class="modal-content" on:click|stopPropagation role="dialog">
    <div class="modal-header">
      <h2>Add {server.name} to MCP Client</h2>
      <button class="btn-close" on:click={onClose}>&times;</button>
    </div>

    <div class="modal-body">
      {#if loading}
        <div class="loading">Loading...</div>
      {:else if errorMessage}
        <div class="error">{errorMessage}</div>
      {:else}
        <!-- Client selection -->
        <div class="form-group">
          <label>MCP Client</label>
          <select bind:value={selectedClient}>
            {#each clients as client}
              <option value={client} disabled={!client.installed}>
                {client.name} {!client.installed ? '(Not Installed)' : ''}
              </option>
            {/each}
          </select>
        </div>

        <!-- Server name -->
        <div class="form-group">
          <label>Server Name</label>
          <input type="text" bind:value={serverName} placeholder="e.g., filesystem" />
        </div>

        <!-- Command -->
        <div class="form-group">
          <label>Command</label>
          <input type="text" bind:value={command} placeholder="e.g., node" />
        </div>

        <!-- Arguments -->
        <div class="form-group">
          <label>Arguments</label>
          <div class="list-items">
            {#each args as arg, i}
              <div class="list-item">
                <span>{arg}</span>
                <button class="btn-small" on:click={() => removeArg(i)}>×</button>
              </div>
            {/each}
          </div>
          <div class="input-group">
            <input type="text" bind:value={newArg} placeholder="Add argument" />
            <button class="btn-secondary" on:click={addArg}>Add</button>
          </div>
        </div>

        <!-- Environment Variables -->
        <div class="form-group">
          <label>Environment Variables</label>
          <div class="list-items">
            {#each Object.entries(envVars) as [key, value]}
              <div class="list-item">
                <span>{key}={value}</span>
                <button class="btn-small" on:click={() => removeEnvVar(key)}>×</button>
              </div>
            {/each}
          </div>
          <div class="input-group">
            <input type="text" bind:value={newEnvKey} placeholder="Key" />
            <input type="text" bind:value={newEnvValue} placeholder="Value" />
            <button class="btn-secondary" on:click={addEnvVar}>Add</button>
          </div>
        </div>

        <!-- Instructions -->
        <div class="info-box">
          <strong>After adding:</strong>
          <ol>
            <li>Close {selectedClient?.name || 'the client'} completely</li>
            <li>Restart {selectedClient?.name || 'the client'}</li>
            <li>The server will start automatically</li>
          </ol>
        </div>
      {/if}
    </div>

    <div class="modal-footer">
      <button class="btn-secondary" on:click={onClose} disabled={saving}>Cancel</button>
      <button class="btn-primary" on:click={addToClient} disabled={saving || loading}>
        {saving ? 'Adding...' : 'Add to Client'}
      </button>
    </div>
  </div>
</div>

<style>
  .modal-backdrop {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: rgba(0, 0, 0, 0.6);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
    padding: var(--spacing-lg);
  }

  .modal-content {
    background-color: var(--bg-primary);
    border-radius: var(--radius-lg);
    box-shadow: 0 10px 40px rgba(0, 0, 0, 0.3);
    max-width: 600px;
    width: 100%;
    max-height: 90vh;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .modal-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: var(--spacing-lg);
    border-bottom: 1px solid var(--border-color);
  }

  .modal-header h2 {
    margin: 0;
    font-size: var(--font-size-lg);
    color: var(--text-primary);
  }

  .btn-close {
    background: none;
    border: none;
    font-size: 1.5rem;
    color: var(--text-secondary);
    cursor: pointer;
    padding: 0;
    width: 2rem;
    height: 2rem;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: var(--radius-sm);
  }

  .btn-close:hover {
    background-color: var(--bg-hover);
  }

  .modal-body {
    flex: 1;
    overflow-y: auto;
    padding: var(--spacing-lg);
  }

  .loading, .error {
    text-align: center;
    padding: var(--spacing-xl);
    color: var(--text-secondary);
  }

  .error {
    color: var(--status-error);
  }

  .form-group {
    margin-bottom: var(--spacing-lg);
  }

  .form-group label {
    display: block;
    margin-bottom: var(--spacing-xs);
    color: var(--text-primary);
    font-weight: 500;
  }

  .form-group input, .form-group select {
    width: 100%;
    padding: var(--spacing-sm);
    border: 1px solid var(--border-color);
    border-radius: var(--radius-sm);
    background-color: var(--bg-secondary);
    color: var(--text-primary);
  }

  .input-group {
    display: flex;
    gap: var(--spacing-xs);
  }

  .input-group input {
    flex: 1;
  }

  .list-items {
    display: flex;
    flex-direction: column;
    gap: var(--spacing-xs);
    margin-bottom: var(--spacing-sm);
  }

  .list-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: var(--spacing-xs) var(--spacing-sm);
    background-color: var(--bg-tertiary);
    border-radius: var(--radius-sm);
  }

  .list-item span {
    font-family: var(--font-mono);
    font-size: var(--font-size-xs);
    color: var(--text-secondary);
  }

  .btn-small {
    background: none;
    border: none;
    color: var(--status-error);
    font-size: 1.2rem;
    cursor: pointer;
    padding: 0 var(--spacing-xs);
  }

  .info-box {
    padding: var(--spacing-md);
    background-color: rgba(33, 150, 243, 0.1);
    border-left: 3px solid var(--accent-primary);
    border-radius: var(--radius-sm);
    margin-top: var(--spacing-md);
  }

  .info-box strong {
    color: var(--text-primary);
  }

  .info-box ol {
    margin: var(--spacing-xs) 0 0 var(--spacing-lg);
    padding: 0;
    color: var(--text-secondary);
  }

  .modal-footer {
    display: flex;
    justify-content: flex-end;
    gap: var(--spacing-sm);
    padding: var(--spacing-lg);
    border-top: 1px solid var(--border-color);
  }

  .btn-secondary, .btn-primary {
    padding: var(--spacing-sm) var(--spacing-lg);
    border-radius: var(--radius-md);
    font-size: var(--font-size-sm);
    font-weight: 500;
    cursor: pointer;
    transition: all var(--transition-fast);
  }

  .btn-secondary {
    background-color: var(--button-bg);
    border: 1px solid var(--border-color);
    color: var(--text-primary);
  }

  .btn-secondary:hover:not(:disabled) {
    background-color: var(--button-hover);
  }

  .btn-primary {
    background-color: var(--accent-primary);
    border: 1px solid var(--accent-primary);
    color: white;
  }

  .btn-primary:hover:not(:disabled) {
    background-color: var(--accent-hover);
  }

  button:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
</style>
