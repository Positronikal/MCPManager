<script lang="ts">
  import { servers, filteredServers, selectedServerId, addNotification } from '../stores/stores';
  import { api } from '../services/api';
  import type { MCPServer } from '../stores/stores';
  import ConfigurationEditor from './ConfigurationEditor.svelte';
  import DetailedLogsView from './DetailedLogsView.svelte';
  import StdioInfoModal from './StdioInfoModal.svelte';
  import ClientConfigEditorModal from './ClientConfigEditorModal.svelte';

  // Loading states for individual servers
  let loadingServers = new Map<string, string>(); // serverId -> action type

  // Configuration editor state
  let showConfigEditor = false;
  let configEditorServerId: string | null = null;
  let configEditorServerName: string | null = null;

  // Detailed logs view state
  let showDetailedLogs = false;
  let detailedLogsServerId: string | null = null;

  // Stdio info modal state
  let showStdioInfoModal = false;
  let stdioInfoServer: MCPServer | null = null;

  // Client config editor state
  let showClientConfigEditor = false;
  let clientConfigEditorServer: MCPServer | null = null;

  // Note: No need to fetch servers on mount - App.svelte triggers refreshDiscovery()
  // on startup, and the backend emits servers:initial event which populates the store
  // via events.ts handlers. This avoids race condition where listServers() is called
  // before backend completes initial discovery.

  // Handle Start server action
  async function handleStart(server: MCPServer) {
    // Check transport type (Option D: stdio servers need client configuration)
    if (server.transport === 'stdio') {
      stdioInfoServer = server;
      showStdioInfoModal = true;
      return;
    }

    // Standalone servers (http/sse/unknown) can be started directly
    try {
      loadingServers.set(server.id, 'starting');
      loadingServers = loadingServers; // Trigger reactivity

      await api.lifecycle.startServer(server.id);
      addNotification('success', `Starting ${server.name}...`);
    } catch (error) {
      console.error('Failed to start server:', error);
      addNotification('error', `Failed to start ${server.name}: ${error.message}`);
    } finally {
      loadingServers.delete(server.id);
      loadingServers = loadingServers;
    }
  }

  // Handle Stop server action
  async function handleStop(server: MCPServer, force: boolean = false) {
    console.log('[UI] handleStop called', { serverId: server.id, name: server.name, force, timeout: 10 });
    try {
      loadingServers.set(server.id, 'stopping');
      loadingServers = loadingServers;

      console.log('[UI] Calling api.lifecycle.stopServer...');
      const result = await api.lifecycle.stopServer(server.id, { force, timeout: 10 });
      console.log('[UI] api.lifecycle.stopServer returned:', result);
      addNotification('success', `${force ? 'Force stopping' : 'Stopping'} ${server.name}...`);
    } catch (error) {
      console.error('[UI] Failed to stop server - Full error object:', error);
      console.error('[UI] Error type:', typeof error);
      console.error('[UI] Error message:', error?.message);
      console.error('[UI] Error string:', String(error));
      const errorMsg = error?.message || String(error) || 'undefined';
      addNotification('error', `Failed to stop ${server.name}: ${errorMsg}`);
    } finally {
      loadingServers.delete(server.id);
      loadingServers = loadingServers;
    }
  }

  // Handle Restart server action
  async function handleRestart(server: MCPServer) {
    try {
      loadingServers.set(server.id, 'restarting');
      loadingServers = loadingServers;

      await api.lifecycle.restartServer(server.id);
      addNotification('success', `Restarting ${server.name}...`);
    } catch (error) {
      console.error('Failed to restart server:', error);
      addNotification('error', `Failed to restart ${server.name}: ${error.message}`);
    } finally {
      loadingServers.delete(server.id);
      loadingServers = loadingServers;
    }
  }

  // Handle opening configuration panel
  function openConfig(server: MCPServer) {
    selectedServerId.set(server.id);
    configEditorServerId = server.id;
    configEditorServerName = server.name;
    showConfigEditor = true;
  }

  // Handle closing configuration editor
  function closeConfigEditor() {
    showConfigEditor = false;
    configEditorServerId = null;
    configEditorServerName = null;
  }

  // Handle opening detailed logs view
  function openLogs(server: MCPServer) {
    selectedServerId.set(server.id);
    detailedLogsServerId = server.id;
    showDetailedLogs = true;
  }

  // Handle closing detailed logs view
  function closeDetailedLogs() {
    showDetailedLogs = false;
    detailedLogsServerId = null;
  }

  // Handle closing stdio info modal
  function closeStdioInfoModal() {
    showStdioInfoModal = false;
    stdioInfoServer = null;
  }

  // Handle opening client config editor from stdio modal
  function openClientConfigEditor() {
    if (stdioInfoServer) {
      clientConfigEditorServer = stdioInfoServer;
      showClientConfigEditor = true;
    }
  }

  // Handle closing client config editor
  function closeClientConfigEditor() {
    showClientConfigEditor = false;
    clientConfigEditorServer = null;
  }

  // Get button text based on loading state and transport type
  function getButtonText(serverId: string, action: string, defaultText: string): string {
    const loadingAction = loadingServers.get(serverId);
    if (loadingAction === action) {
      return '...';
    }
    return defaultText;
  }

  // Get start button label based on transport type
  function getStartButtonLabel(server: MCPServer): string {
    if (server.transport === 'stdio') {
      return 'ℹ️ Info';
    }
    return '▶️ Start';
  }

  // Get start button tooltip based on transport type
  function getStartButtonTooltip(server: MCPServer): string {
    if (server.transport === 'stdio') {
      return 'Stdio server - requires MCP client';
    }
    return 'Start server';
  }

  // Check if server is currently loading
  function isServerLoading(serverId: string): boolean {
    return loadingServers.has(serverId);
  }

  // Format uptime from seconds
  function formatUptime(seconds: number): string {
    if (!seconds || seconds === 0) return '-';

    const days = Math.floor(seconds / 86400);
    const hours = Math.floor((seconds % 86400) / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);

    if (days > 0) {
      return `${days}d ${hours}h`;
    } else if (hours > 0) {
      return `${hours}h ${minutes}m`;
    } else {
      return `${minutes}m`;
    }
  }

  // Get status text with proper capitalization
  function getStatusText(state: string): string {
    return state.charAt(0).toUpperCase() + state.slice(1);
  }
</script>

<div class="server-table-container">
  {#if $filteredServers.length === 0}
    <div class="empty-state">
      <div class="empty-state-icon">🖥️</div>
      <h3>No servers found</h3>
      <p class="text-secondary">
        {#if $servers.length === 0}
          Click "Refresh" to discover MCP servers on your system.
        {:else}
          No servers match the current filters.
        {/if}
      </p>
    </div>
  {:else}
    <div class="table-wrapper">
      <table class="server-table">
        <thead>
          <tr>
            <th class="col-status">Status</th>
            <th class="col-name">Name</th>
            <th class="col-version">Version</th>
            <th class="col-transport">Transport</th>
            <th class="col-capabilities">Capabilities</th>
            <th class="col-uptime">Uptime</th>
            <th class="col-pid">PID</th>
            <th class="col-actions">Actions</th>
          </tr>
        </thead>
        <tbody>
          {#each $filteredServers as server (server.id)}
            <tr class:selected={$selectedServerId === server.id}>
              <!-- Status Indicator -->
              <td class="col-status">
                <div class="status-cell">
                  <span
                    class="status-indicator status-{server.status.state}"
                    title="{getStatusText(server.status.state)}"
                  ></span>
                  <span class="status-text">{getStatusText(server.status.state)}</span>
                </div>
              </td>

              <!-- Server Name -->
              <td class="col-name">
                <div class="server-name">
                  <span class="name-text">{server.name}</span>
                  {#if server.source}
                    <span class="source-badge badge-{server.source}">{server.source}</span>
                  {/if}
                </div>
              </td>

              <!-- Version -->
              <td class="col-version">
                <span class="text-secondary">{server.version || 'N/A'}</span>
              </td>

              <!-- Transport -->
              <td class="col-transport">
                <span class="transport-badge transport-{server.transport || 'unknown'}">
                  {server.transport || 'unknown'}
                </span>
              </td>

              <!-- Capabilities -->
              <td class="col-capabilities">
                {#if server.capabilities && server.capabilities.length > 0}
                  <div class="capabilities-list">
                    {#each server.capabilities.slice(0, 3) as capability}
                      <span class="capability-badge">{capability}</span>
                    {/each}
                    {#if server.capabilities.length > 3}
                      <span class="capability-more text-muted">+{server.capabilities.length - 3}</span>
                    {/if}
                  </div>
                {:else}
                  <span class="text-muted">N/A</span>
                {/if}
              </td>

              <!-- Uptime -->
              <td class="col-uptime">
                <span class="text-secondary">
                  {formatUptime(server.status.uptime || 0)}
                </span>
              </td>

              <!-- PID -->
              <td class="col-pid">
                <span class="text-secondary mono">{server.pid || '-'}</span>
              </td>

              <!-- Actions -->
              <td class="col-actions">
                <div class="action-buttons">
                  <!-- Start button (only when stopped or error) -->
                  {#if server.status.state === 'stopped' || server.status.state === 'error'}
                    <button
                      class="btn-action {server.transport === 'stdio' ? 'btn-info' : 'btn-start'}"
                      on:click={() => handleStart(server)}
                      disabled={isServerLoading(server.id)}
                      title={getStartButtonTooltip(server)}
                    >
                      {getButtonText(server.id, 'starting', getStartButtonLabel(server))}
                    </button>
                  {/if}

                  <!-- Stop and Restart buttons (only when running) -->
                  {#if server.status.state === 'running'}
                    <button
                      class="btn-action btn-stop"
                      on:click={() => handleStop(server, false)}
                      disabled={isServerLoading(server.id)}
                      title="Stop server gracefully"
                    >
                      {getButtonText(server.id, 'stopping', '⏹️ Stop')}
                    </button>
                    <button
                      class="btn-action btn-restart"
                      on:click={() => handleRestart(server)}
                      disabled={isServerLoading(server.id)}
                      title="Restart server"
                    >
                      {getButtonText(server.id, 'restarting', '🔄 Restart')}
                    </button>
                  {/if}

                  <!-- Config and Logs buttons (always available) -->
                  <button
                    class="btn-action btn-config"
                    on:click={() => openConfig(server)}
                    disabled={isServerLoading(server.id)}
                    title="Open configuration"
                  >
                    ⚙️
                  </button>
                  <button
                    class="btn-action btn-logs"
                    on:click={() => openLogs(server)}
                    title="View logs"
                  >
                    📋
                  </button>
                </div>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>

    <!-- Server count footer -->
    <div class="table-footer">
      <span class="text-secondary">
        Showing {$filteredServers.length} of {$servers.length} server{$servers.length !== 1 ? 's' : ''}
      </span>
    </div>
  {/if}
</div>

<!-- Configuration Editor Modal -->
{#if showConfigEditor && configEditorServerId && configEditorServerName}
  <ConfigurationEditor
    serverId={configEditorServerId}
    serverName={configEditorServerName}
    onClose={closeConfigEditor}
  />
{/if}

<!-- Detailed Logs View Modal -->
{#if showDetailedLogs && detailedLogsServerId}
  <DetailedLogsView
    serverId={detailedLogsServerId}
    onClose={closeDetailedLogs}
  />
{/if}

<!-- Stdio Info Modal -->
{#if showStdioInfoModal && stdioInfoServer}
  <StdioInfoModal
    server={stdioInfoServer}
    onClose={closeStdioInfoModal}
    onOpenConfigEditor={openClientConfigEditor}
  />
{/if}

<!-- Client Config Editor Modal -->
{#if showClientConfigEditor && clientConfigEditorServer}
  <ClientConfigEditorModal
    server={clientConfigEditorServer}
    onClose={closeClientConfigEditor}
  />
{/if}

<style>
  .server-table-container {
    display: flex;
    flex-direction: column;
    height: 100%;
    background-color: var(--bg-secondary);
    border-radius: var(--radius-md);
    overflow: hidden;
  }

  /* Empty state */
  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: var(--spacing-xl);
    text-align: center;
    min-height: 300px;
  }

  .empty-state-icon {
    font-size: 4rem;
    margin-bottom: var(--spacing-md);
    opacity: 0.5;
  }

  .empty-state h3 {
    margin-bottom: var(--spacing-sm);
    color: var(--text-primary);
  }

  /* Table wrapper */
  .table-wrapper {
    flex: 1;
    overflow: auto;
  }

  /* Table styling */
  .server-table {
    width: 100%;
    border-collapse: collapse;
  }

  .server-table thead {
    position: sticky;
    top: 0;
    background-color: var(--bg-tertiary);
    z-index: 10;
  }

  .server-table th {
    text-align: left;
    padding: var(--spacing-md);
    font-weight: 600;
    font-size: var(--font-size-sm);
    color: var(--text-primary);
    border-bottom: 2px solid var(--border-color);
    white-space: nowrap;
  }

  .server-table td {
    padding: var(--spacing-md);
    border-bottom: 1px solid var(--border-color);
    vertical-align: middle;
  }

  .server-table tbody tr {
    transition: background-color var(--transition-fast);
  }

  .server-table tbody tr:hover {
    background-color: var(--bg-hover);
  }

  .server-table tbody tr.selected {
    background-color: rgba(33, 150, 243, 0.1);
    border-left: 3px solid var(--accent-primary);
  }

  /* Column widths */
  .col-status { width: 120px; }
  .col-name { width: auto; min-width: 200px; }
  .col-version { width: 100px; }
  .col-transport { width: 90px; }
  .col-capabilities { width: 200px; }
  .col-uptime { width: 100px; }
  .col-pid { width: 80px; }
  .col-actions { width: 280px; }

  /* Status cell */
  .status-cell {
    display: flex;
    align-items: center;
    gap: var(--spacing-sm);
  }

  .status-indicator {
    display: inline-block;
    width: 12px;
    height: 12px;
    border-radius: 50%;
    flex-shrink: 0;
  }

  .status-running {
    background-color: var(--status-running);
    box-shadow: 0 0 8px var(--status-running);
  }

  .status-stopped {
    background-color: var(--status-stopped);
  }

  .status-starting {
    background-color: var(--status-starting);
    animation: pulse 1.5s ease-in-out infinite;
  }

  .status-error {
    background-color: var(--status-error);
    animation: pulse 1s ease-in-out infinite;
  }

  @keyframes pulse {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.5; }
  }

  .status-text {
    font-size: var(--font-size-sm);
    font-weight: 500;
    color: var(--text-secondary);
  }

  /* Server name cell */
  .server-name {
    display: flex;
    flex-direction: column;
    gap: var(--spacing-xs);
  }

  .name-text {
    font-weight: 500;
    color: var(--text-primary);
  }

  .source-badge {
    display: inline-block;
    padding: 2px 6px;
    border-radius: var(--radius-sm);
    font-size: var(--font-size-xs);
    font-weight: 500;
    background-color: var(--bg-tertiary);
    color: var(--text-muted);
    text-transform: uppercase;
  }

  .badge-client_config {
    background-color: rgba(33, 150, 243, 0.2);
    color: var(--accent-primary);
  }

  .badge-filesystem {
    background-color: rgba(76, 175, 80, 0.2);
    color: var(--status-running);
  }

  .badge-process {
    background-color: rgba(255, 152, 0, 0.2);
    color: var(--status-error);
  }

  /* Capabilities */
  .capabilities-list {
    display: flex;
    flex-wrap: wrap;
    gap: var(--spacing-xs);
  }

  .capability-badge {
    padding: 2px 6px;
    border-radius: var(--radius-sm);
    font-size: var(--font-size-xs);
    background-color: var(--bg-tertiary);
    color: var(--text-secondary);
    border: 1px solid var(--border-color);
  }

  .capability-more {
    font-size: var(--font-size-xs);
    padding: 2px 6px;
  }

  /* Transport badge */
  .transport-badge {
    display: inline-block;
    padding: 2px 8px;
    border-radius: var(--radius-sm);
    font-size: var(--font-size-xs);
    font-weight: 500;
    text-transform: uppercase;
    border: 1px solid;
  }

  .transport-stdio {
    background-color: rgba(33, 150, 243, 0.2);
    border-color: var(--accent-primary);
    color: var(--accent-primary);
  }

  .transport-http {
    background-color: rgba(76, 175, 80, 0.2);
    border-color: var(--status-running);
    color: var(--status-running);
  }

  .transport-sse {
    background-color: rgba(156, 39, 176, 0.2);
    border-color: #9c27b0;
    color: #9c27b0;
  }

  .transport-unknown {
    background-color: var(--bg-tertiary);
    border-color: var(--border-color);
    color: var(--text-muted);
  }

  /* Monospace text */
  .mono {
    font-family: var(--font-mono);
    font-size: var(--font-size-xs);
  }

  /* Action buttons */
  .action-buttons {
    display: flex;
    gap: var(--spacing-xs);
    flex-wrap: wrap;
  }

  .btn-action {
    padding: var(--spacing-xs) var(--spacing-sm);
    font-size: var(--font-size-xs);
    border-radius: var(--radius-sm);
    white-space: nowrap;
  }

  .btn-start {
    background-color: rgba(76, 175, 80, 0.2);
    border-color: var(--status-running);
    color: var(--status-running);
  }

  .btn-start:hover:not(:disabled) {
    background-color: rgba(76, 175, 80, 0.3);
  }

  .btn-info {
    background-color: rgba(33, 150, 243, 0.2);
    border-color: var(--accent-primary);
    color: var(--accent-primary);
  }

  .btn-info:hover:not(:disabled) {
    background-color: rgba(33, 150, 243, 0.3);
  }

  .btn-stop {
    background-color: rgba(244, 67, 54, 0.2);
    border-color: var(--status-stopped);
    color: var(--status-stopped);
  }

  .btn-stop:hover:not(:disabled) {
    background-color: rgba(244, 67, 54, 0.3);
  }

  .btn-restart {
    background-color: rgba(33, 150, 243, 0.2);
    border-color: var(--status-starting);
    color: var(--status-starting);
  }

  .btn-restart:hover:not(:disabled) {
    background-color: rgba(33, 150, 243, 0.3);
  }

  .btn-config,
  .btn-logs {
    background-color: var(--button-bg);
    border-color: var(--border-color);
    color: var(--text-secondary);
  }

  .btn-config:hover:not(:disabled),
  .btn-logs:hover:not(:disabled) {
    background-color: var(--button-hover);
  }

  /* Table footer */
  .table-footer {
    padding: var(--spacing-md);
    border-top: 1px solid var(--border-color);
    background-color: var(--bg-tertiary);
    text-align: center;
    font-size: var(--font-size-sm);
  }

  /* Responsive */
  @media (max-width: 1024px) {
    .col-capabilities {
      display: none;
    }
  }

  @media (max-width: 768px) {
    .col-version,
    .col-transport,
    .col-uptime {
      display: none;
    }

    .status-text {
      display: none;
    }
  }
</style>
