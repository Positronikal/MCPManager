<script lang="ts">
  import { onMount } from 'svelte';
  import type { ServerConfiguration, MCPServer } from '../stores/stores';
  import { api } from '../services/api';
  import { addNotification } from '../stores/stores';

  // Props
  export let serverId: string;
  export let serverName: string;
  export let onClose: () => void;

  // State
  let config: ServerConfiguration = {
    environmentVariables: {},
    commandLineArguments: [],
    workingDirectory: '',
    autoStart: false,
    restartOnCrash: false,
    maxRestartAttempts: 3,
    startupTimeout: 30,
    shutdownTimeout: 10,
    healthCheckInterval: 0,
    healthCheckEndpoint: ''
  };

  let server: MCPServer | null = null;
  let loading = true;
  let saving = false;
  let errorMessage = '';
  let hasUnsavedChanges = false;

  // Environment variables state
  let newEnvKey = '';
  let newEnvValue = '';
  let envVarError = '';

  // Command-line arguments state
  let newArg = '';

  // Environment variable regex validation
  const ENV_VAR_REGEX = /^[A-Z_][A-Z0-9_]*$/;

  // Load configuration on mount
  onMount(async () => {
    await loadConfiguration();
    await loadServerDetails();
  });

  async function loadConfiguration() {
    loading = true;
    errorMessage = '';
    try {
      config = await api.config.getConfiguration(serverId);

      // Ensure arrays and objects are initialized
      if (!config.environmentVariables) config.environmentVariables = {};
      if (!config.commandLineArguments) config.commandLineArguments = [];

      loading = false;
    } catch (error: any) {
      errorMessage = `Failed to load configuration: ${error.message}`;
      loading = false;
    }
  }

  async function loadServerDetails() {
    try {
      server = await api.discovery.getServer(serverId);
    } catch (error: any) {
      console.error('Failed to load server details:', error);
    }
  }

  // Environment variables management
  function addEnvironmentVariable() {
    envVarError = '';

    // Validate key
    if (!newEnvKey.trim()) {
      envVarError = 'Key cannot be empty';
      return;
    }

    if (!ENV_VAR_REGEX.test(newEnvKey)) {
      envVarError = 'Key must start with letter or underscore, contain only uppercase letters, digits, and underscores';
      return;
    }

    if (config.environmentVariables![newEnvKey]) {
      envVarError = 'Variable already exists';
      return;
    }

    // Add variable
    config.environmentVariables![newEnvKey] = newEnvValue;
    config.environmentVariables = { ...config.environmentVariables };

    // Clear inputs
    newEnvKey = '';
    newEnvValue = '';
    hasUnsavedChanges = true;
  }

  function deleteEnvironmentVariable(key: string) {
    delete config.environmentVariables![key];
    config.environmentVariables = { ...config.environmentVariables };
    hasUnsavedChanges = true;
  }

  // Command-line arguments management
  function addArgument() {
    if (!newArg.trim()) return;

    config.commandLineArguments!.push(newArg.trim());
    config.commandLineArguments = [...config.commandLineArguments!];
    newArg = '';
    hasUnsavedChanges = true;
  }

  function deleteArgument(index: number) {
    config.commandLineArguments!.splice(index, 1);
    config.commandLineArguments = [...config.commandLineArguments!];
    hasUnsavedChanges = true;
  }

  function moveArgumentUp(index: number) {
    if (index === 0) return;
    const args = [...config.commandLineArguments!];
    [args[index - 1], args[index]] = [args[index], args[index - 1]];
    config.commandLineArguments = args;
    hasUnsavedChanges = true;
  }

  function moveArgumentDown(index: number) {
    if (index === config.commandLineArguments!.length - 1) return;
    const args = [...config.commandLineArguments!];
    [args[index], args[index + 1]] = [args[index + 1], args[index]];
    config.commandLineArguments = args;
    hasUnsavedChanges = true;
  }

  // Client-side validation
  function validateConfiguration(): string | null {
    // Validate environment variable names
    for (const key of Object.keys(config.environmentVariables || {})) {
      if (!ENV_VAR_REGEX.test(key)) {
        return `Invalid environment variable name: ${key}`;
      }
    }

    // Validate max restart attempts range
    if (config.maxRestartAttempts < 0 || config.maxRestartAttempts > 10) {
      return 'Max restart attempts must be between 0 and 10';
    }

    // Validate timeouts are positive
    if (config.startupTimeout <= 0) {
      return 'Startup timeout must be positive';
    }

    if (config.shutdownTimeout <= 0) {
      return 'Shutdown timeout must be positive';
    }

    // Validate health check consistency
    if (config.healthCheckEndpoint && config.healthCheckInterval! <= 0) {
      return 'Health check interval must be positive when endpoint is set';
    }

    return null;
  }

  // Save configuration
  async function saveConfiguration() {
    errorMessage = '';

    // Client-side validation
    const validationError = validateConfiguration();
    if (validationError) {
      errorMessage = validationError;
      return;
    }

    saving = true;
    try {
      const updatedConfig = await api.config.updateConfiguration(serverId, config);
      config = updatedConfig;
      hasUnsavedChanges = false;
      addNotification('success', `Configuration saved for ${serverName}`);
      onClose();
    } catch (error: any) {
      // Handle validation errors from backend (400 responses)
      if (error.message.includes('Validation error')) {
        errorMessage = error.message;
      } else if (error.message.includes('404')) {
        errorMessage = 'Server not found';
      } else {
        errorMessage = `Failed to save configuration: ${error.message}`;
      }
      saving = false;
    }
  }

  // Handle cancel
  function handleCancel() {
    if (hasUnsavedChanges) {
      if (!confirm('You have unsaved changes. Are you sure you want to close?')) {
        return;
      }
    }
    onClose();
  }

  // Handle keyboard shortcuts
  function handleKeydown(event: KeyboardEvent) {
    if (event.key === 'Escape') {
      handleCancel();
    } else if (event.ctrlKey && event.key === 's') {
      event.preventDefault();
      saveConfiguration();
    }
  }

  // Track changes
  function markChanged() {
    hasUnsavedChanges = true;
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<!-- Modal backdrop -->
<div class="modal-backdrop" on:click={handleCancel} role="presentation">
  <!-- Modal content -->
  <div class="modal-content" on:click|stopPropagation role="dialog" aria-modal="true" aria-labelledby="modal-title">
    <!-- Header -->
    <div class="modal-header">
      <h2 id="modal-title" class="modal-title">
        Configuration - {serverName}
      </h2>
      <button class="btn-close" on:click={handleCancel} aria-label="Close">&times;</button>
    </div>

    <!-- Body -->
    <div class="modal-body">
      {#if loading}
        <div class="loading-state">
          <div class="spinner"></div>
          <p>Loading configuration...</p>
        </div>
      {:else}
        <!-- Error message -->
        {#if errorMessage}
          <div class="error-message" role="alert">
            <span class="error-icon">❌</span>
            <span>{errorMessage}</span>
          </div>
        {/if}

        <!-- MCP Manager Configuration (Editable) -->
        <section class="config-section">
          <h3 class="section-title">MCP Manager Configuration</h3>
          <p class="section-description">Configure how MCP Manager launches and manages this server.</p>

          <!-- Environment Variables -->
          <div class="form-group">
            <label class="form-label">
              Environment Variables
              <span class="label-hint">(Uppercase letters, digits, underscores only)</span>
            </label>

            {#if Object.keys(config.environmentVariables || {}).length > 0}
              <table class="env-table">
                <thead>
                  <tr>
                    <th>Key</th>
                    <th>Value</th>
                    <th>Action</th>
                  </tr>
                </thead>
                <tbody>
                  {#each Object.entries(config.environmentVariables || {}) as [key, value]}
                    <tr>
                      <td class="env-key">{key}</td>
                      <td class="env-value">{value}</td>
                      <td class="env-action">
                        <button
                          class="btn-small btn-danger"
                          on:click={() => deleteEnvironmentVariable(key)}
                          type="button"
                        >
                          Delete
                        </button>
                      </td>
                    </tr>
                  {/each}
                </tbody>
              </table>
            {:else}
              <p class="empty-text">No environment variables configured.</p>
            {/if}

            <!-- Add environment variable -->
            <div class="add-row">
              <input
                type="text"
                class="input-small"
                bind:value={newEnvKey}
                placeholder="KEY"
                maxlength="100"
              />
              <input
                type="text"
                class="input-medium"
                bind:value={newEnvValue}
                placeholder="value"
              />
              <button
                class="btn-small btn-primary"
                on:click={addEnvironmentVariable}
                type="button"
              >
                Add Variable
              </button>
            </div>

            {#if envVarError}
              <div class="field-error">{envVarError}</div>
            {/if}
          </div>

          <!-- Command-Line Arguments -->
          <div class="form-group">
            <label class="form-label">Command-Line Arguments</label>

            {#if config.commandLineArguments && config.commandLineArguments.length > 0}
              <div class="args-list">
                {#each config.commandLineArguments as arg, index}
                  <div class="arg-item">
                    <span class="arg-text">{arg}</span>
                    <div class="arg-actions">
                      <button
                        class="btn-icon-small"
                        on:click={() => moveArgumentUp(index)}
                        disabled={index === 0}
                        title="Move up"
                        type="button"
                      >
                        ↑
                      </button>
                      <button
                        class="btn-icon-small"
                        on:click={() => moveArgumentDown(index)}
                        disabled={index === (config.commandLineArguments || []).length - 1}
                        title="Move down"
                        type="button"
                      >
                        ↓
                      </button>
                      <button
                        class="btn-icon-small btn-danger"
                        on:click={() => deleteArgument(index)}
                        title="Remove"
                        type="button"
                      >
                        ×
                      </button>
                    </div>
                  </div>
                {/each}
              </div>
            {:else}
              <p class="empty-text">No command-line arguments configured.</p>
            {/if}

            <!-- Add argument -->
            <div class="add-row">
              <input
                type="text"
                class="input-large"
                bind:value={newArg}
                placeholder="--argument value"
                on:keydown={(e) => e.key === 'Enter' && addArgument()}
              />
              <button
                class="btn-small btn-primary"
                on:click={addArgument}
                type="button"
              >
                Add Argument
              </button>
            </div>
          </div>

          <!-- Working Directory -->
          <div class="form-group">
            <label class="form-label" for="working-dir">
              Working Directory
              <span class="label-hint">(Optional)</span>
            </label>
            <input
              id="working-dir"
              type="text"
              class="input-full"
              bind:value={config.workingDirectory}
              on:input={markChanged}
              placeholder="Leave empty to use default"
            />
          </div>

          <!-- Auto-start -->
          <div class="form-group">
            <label class="checkbox-label">
              <input
                type="checkbox"
                bind:checked={config.autoStart}
                on:change={markChanged}
              />
              <span>Auto-start server when MCP Manager launches</span>
            </label>
          </div>

          <!-- Restart on crash -->
          <div class="form-group">
            <label class="checkbox-label">
              <input
                type="checkbox"
                bind:checked={config.restartOnCrash}
                on:change={markChanged}
              />
              <span>Restart automatically on crash</span>
            </label>
          </div>

          <!-- Max restart attempts -->
          <div class="form-group">
            <label class="form-label" for="max-restarts">
              Max Restart Attempts (0-10)
            </label>
            <input
              id="max-restarts"
              type="number"
              class="input-small"
              bind:value={config.maxRestartAttempts}
              on:input={markChanged}
              min="0"
              max="10"
            />
            <span class="field-hint">0 = no automatic restarts, 10 = maximum</span>
          </div>

          <!-- Advanced settings (collapsible) -->
          <details class="advanced-settings">
            <summary class="advanced-toggle">Advanced Settings</summary>

            <div class="form-group">
              <label class="form-label" for="startup-timeout">
                Startup Timeout (seconds)
              </label>
              <input
                id="startup-timeout"
                type="number"
                class="input-small"
                bind:value={config.startupTimeout}
                on:input={markChanged}
                min="1"
              />
            </div>

            <div class="form-group">
              <label class="form-label" for="shutdown-timeout">
                Shutdown Timeout (seconds)
              </label>
              <input
                id="shutdown-timeout"
                type="number"
                class="input-small"
                bind:value={config.shutdownTimeout}
                on:input={markChanged}
                min="1"
              />
            </div>

            <div class="form-group">
              <label class="form-label" for="health-endpoint">
                Health Check Endpoint
                <span class="label-hint">(Optional)</span>
              </label>
              <input
                id="health-endpoint"
                type="text"
                class="input-full"
                bind:value={config.healthCheckEndpoint}
                on:input={markChanged}
                placeholder="http://localhost:3000/health"
              />
            </div>

            <div class="form-group">
              <label class="form-label" for="health-interval">
                Health Check Interval (seconds)
              </label>
              <input
                id="health-interval"
                type="number"
                class="input-small"
                bind:value={config.healthCheckInterval}
                on:input={markChanged}
                min="0"
              />
              <span class="field-hint">0 = disabled</span>
            </div>
          </details>
        </section>

        <!-- Client Configuration (Read-Only) -->
        {#if server}
          <section class="config-section readonly-section">
            <h3 class="section-title">Client Configuration (Read-Only)</h3>
            <p class="section-description">
              This server was discovered from: <strong>{server.source}</strong>
            </p>

            <div class="readonly-field">
              <span class="field-label">Installation Path:</span>
              <span class="field-value">{server.installationPath}</span>
            </div>

            {#if server.source === 'client_config'}
              <div class="readonly-notice">
                <span class="notice-icon">ℹ️</span>
                <span>
                  This server is configured in a client config file. MCP Manager will never modify that file.
                  The settings above control how MCP Manager manages the server process.
                </span>
              </div>
            {/if}
          </section>
        {/if}
      {/if}
    </div>

    <!-- Footer -->
    <div class="modal-footer">
      <button
        class="btn btn-secondary"
        on:click={handleCancel}
        disabled={saving}
      >
        Cancel
      </button>
      <button
        class="btn btn-primary"
        on:click={saveConfiguration}
        disabled={loading || saving}
      >
        {#if saving}
          Saving...
        {:else}
          Save Configuration
        {/if}
      </button>
    </div>
  </div>
</div>

<style>
  /* Modal backdrop */
  .modal-backdrop {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
    padding: var(--spacing-md);
  }

  /* Modal content */
  .modal-content {
    background-color: var(--bg-primary);
    border-radius: var(--border-radius);
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
    max-width: 800px;
    width: 100%;
    max-height: 90vh;
    display: flex;
    flex-direction: column;
  }

  /* Modal header */
  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: var(--spacing-md);
    border-bottom: 1px solid var(--border-color);
  }

  .modal-title {
    margin: 0;
    font-size: var(--font-size-lg);
    font-weight: 600;
    color: var(--text-primary);
  }

  .btn-close {
    background: none;
    border: none;
    font-size: 2rem;
    line-height: 1;
    color: var(--text-secondary);
    cursor: pointer;
    padding: 0;
    width: 32px;
    height: 32px;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: color var(--transition-fast);
  }

  .btn-close:hover {
    color: var(--text-primary);
  }

  /* Modal body */
  .modal-body {
    flex: 1;
    overflow-y: auto;
    padding: var(--spacing-md);
  }

  /* Loading state */
  .loading-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: var(--spacing-xl);
    gap: var(--spacing-md);
  }

  .spinner {
    width: 40px;
    height: 40px;
    border: 4px solid var(--border-color);
    border-top-color: var(--accent-primary);
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  /* Error message */
  .error-message {
    display: flex;
    align-items: center;
    gap: var(--spacing-sm);
    padding: var(--spacing-md);
    background-color: rgba(244, 67, 54, 0.1);
    border: 1px solid var(--status-error);
    border-radius: var(--border-radius);
    color: var(--status-error);
    margin-bottom: var(--spacing-md);
  }

  .error-icon {
    font-size: var(--font-size-lg);
  }

  /* Config sections */
  .config-section {
    margin-bottom: var(--spacing-lg);
    padding: var(--spacing-md);
    background-color: var(--bg-secondary);
    border-radius: var(--border-radius);
  }

  .section-title {
    margin: 0 0 var(--spacing-xs) 0;
    font-size: var(--font-size-md);
    font-weight: 600;
    color: var(--text-primary);
  }

  .section-description {
    margin: 0 0 var(--spacing-md) 0;
    font-size: var(--font-size-sm);
    color: var(--text-secondary);
  }

  /* Form groups */
  .form-group {
    margin-bottom: var(--spacing-md);
  }

  .form-label {
    display: block;
    margin-bottom: var(--spacing-xs);
    font-size: var(--font-size-sm);
    font-weight: 500;
    color: var(--text-primary);
  }

  .label-hint {
    font-weight: normal;
    color: var(--text-secondary);
    font-size: var(--font-size-xs);
  }

  .field-hint {
    display: block;
    margin-top: var(--spacing-xs);
    font-size: var(--font-size-xs);
    color: var(--text-secondary);
  }

  .field-error {
    margin-top: var(--spacing-xs);
    font-size: var(--font-size-sm);
    color: var(--status-error);
  }

  /* Input styles */
  input[type="text"],
  input[type="number"] {
    padding: var(--spacing-xs) var(--spacing-sm);
    border: 1px solid var(--border-color);
    border-radius: var(--border-radius);
    background-color: var(--bg-primary);
    color: var(--text-primary);
    font-size: var(--font-size-sm);
  }

  input[type="text"]:focus,
  input[type="number"]:focus {
    outline: none;
    border-color: var(--accent-primary);
  }

  .input-small {
    width: 150px;
  }

  .input-medium {
    width: 250px;
  }

  .input-large {
    width: 400px;
  }

  .input-full {
    width: 100%;
  }

  /* Checkbox */
  .checkbox-label {
    display: flex;
    align-items: center;
    gap: var(--spacing-sm);
    cursor: pointer;
    font-size: var(--font-size-sm);
    color: var(--text-primary);
  }

  input[type="checkbox"] {
    width: 18px;
    height: 18px;
    cursor: pointer;
  }

  /* Environment variables table */
  .env-table {
    width: 100%;
    border-collapse: collapse;
    margin-bottom: var(--spacing-sm);
    font-size: var(--font-size-sm);
  }

  .env-table th {
    text-align: left;
    padding: var(--spacing-xs) var(--spacing-sm);
    background-color: var(--bg-tertiary);
    border-bottom: 1px solid var(--border-color);
    font-weight: 600;
    color: var(--text-primary);
  }

  .env-table td {
    padding: var(--spacing-xs) var(--spacing-sm);
    border-bottom: 1px solid var(--border-color);
  }

  .env-key {
    font-family: var(--font-mono);
    color: var(--text-primary);
    font-weight: 500;
  }

  .env-value {
    font-family: var(--font-mono);
    color: var(--text-secondary);
    word-break: break-all;
  }

  .env-action {
    width: 100px;
    text-align: right;
  }

  .empty-text {
    color: var(--text-secondary);
    font-size: var(--font-size-sm);
    font-style: italic;
    margin: var(--spacing-sm) 0;
  }

  /* Add row */
  .add-row {
    display: flex;
    gap: var(--spacing-sm);
    align-items: center;
    flex-wrap: wrap;
  }

  /* Command-line arguments list */
  .args-list {
    margin-bottom: var(--spacing-sm);
  }

  .arg-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: var(--spacing-xs) var(--spacing-sm);
    border: 1px solid var(--border-color);
    border-radius: var(--border-radius);
    margin-bottom: var(--spacing-xs);
    background-color: var(--bg-primary);
  }

  .arg-text {
    font-family: var(--font-mono);
    font-size: var(--font-size-sm);
    color: var(--text-primary);
    flex: 1;
    word-break: break-all;
  }

  .arg-actions {
    display: flex;
    gap: var(--spacing-xs);
  }

  /* Buttons */
  .btn {
    padding: var(--spacing-sm) var(--spacing-md);
    border: 1px solid var(--border-color);
    border-radius: var(--border-radius);
    background-color: var(--bg-secondary);
    color: var(--text-primary);
    font-size: var(--font-size-sm);
    font-weight: 500;
    cursor: pointer;
    transition: all var(--transition-fast);
  }

  .btn:hover:not(:disabled) {
    background-color: var(--bg-hover);
  }

  .btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .btn-primary {
    background-color: var(--accent-primary);
    border-color: var(--accent-primary);
    color: white;
  }

  .btn-primary:hover:not(:disabled) {
    background-color: var(--accent-hover);
    border-color: var(--accent-hover);
  }

  .btn-secondary {
    background-color: var(--bg-tertiary);
  }

  .btn-small {
    padding: var(--spacing-xs) var(--spacing-sm);
    font-size: var(--font-size-xs);
  }

  .btn-danger {
    background-color: var(--status-error);
    border-color: var(--status-error);
    color: white;
  }

  .btn-danger:hover:not(:disabled) {
    background-color: #d32f2f;
    border-color: #d32f2f;
  }

  .btn-icon-small {
    padding: 2px 8px;
    font-size: var(--font-size-sm);
    min-width: 28px;
    border: 1px solid var(--border-color);
    border-radius: var(--border-radius);
    background-color: var(--bg-secondary);
    color: var(--text-primary);
    cursor: pointer;
    transition: all var(--transition-fast);
  }

  .btn-icon-small:hover:not(:disabled) {
    background-color: var(--bg-hover);
  }

  .btn-icon-small:disabled {
    opacity: 0.3;
    cursor: not-allowed;
  }

  /* Advanced settings */
  .advanced-settings {
    margin-top: var(--spacing-md);
    padding: var(--spacing-md);
    border: 1px solid var(--border-color);
    border-radius: var(--border-radius);
  }

  .advanced-toggle {
    cursor: pointer;
    font-size: var(--font-size-sm);
    font-weight: 600;
    color: var(--accent-primary);
    user-select: none;
  }

  .advanced-toggle:hover {
    color: var(--accent-hover);
  }

  /* Read-only section */
  .readonly-section {
    background-color: var(--bg-tertiary);
    border: 1px solid var(--border-color);
  }

  .readonly-field {
    display: flex;
    gap: var(--spacing-sm);
    margin-bottom: var(--spacing-sm);
    font-size: var(--font-size-sm);
  }

  .field-label {
    font-weight: 600;
    color: var(--text-primary);
  }

  .field-value {
    color: var(--text-secondary);
    word-break: break-all;
  }

  .readonly-notice {
    display: flex;
    align-items: flex-start;
    gap: var(--spacing-sm);
    padding: var(--spacing-sm);
    background-color: rgba(33, 150, 243, 0.1);
    border: 1px solid var(--log-info);
    border-radius: var(--border-radius);
    font-size: var(--font-size-sm);
    color: var(--text-secondary);
  }

  .notice-icon {
    font-size: var(--font-size-md);
  }

  /* Modal footer */
  .modal-footer {
    display: flex;
    justify-content: flex-end;
    gap: var(--spacing-sm);
    padding: var(--spacing-md);
    border-top: 1px solid var(--border-color);
  }

  /* Responsive */
  @media (max-width: 768px) {
    .modal-content {
      max-width: 100%;
      max-height: 100vh;
      border-radius: 0;
    }

    .input-large {
      width: 100%;
    }

    .add-row {
      flex-direction: column;
      align-items: stretch;
    }

    .input-small,
    .input-medium {
      width: 100%;
    }
  }
</style>
