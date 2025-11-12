<script lang="ts">
  import { onMount } from 'svelte';
  import { servers, runningServers, addNotification } from '../stores/stores';
  import { GetNetstat } from '../../wailsjs/go/main/App';

  interface NetstatEntry {
    protocol: string;
    localAddress: string;
    remoteAddress: string;
    state: string;
    pid: number;
  }

  let connections: NetstatEntry[] = [];
  let loading = false;
  let error = '';
  let autoRefresh = false;
  let refreshInterval: number | null = null;

  onMount(() => {
    loadConnections();
  });

  async function loadConnections() {
    loading = true;
    error = '';

    try {
      // Get PIDs of running servers
      const pidsList = $runningServers
        .filter(s => s.pid)
        .map(s => s.pid as number);

      if (pidsList.length === 0) {
        connections = [];
        error = 'No running servers to monitor';
        loading = false;
        return;
      }

      // Call Wails backend method
      const response = await GetNetstat(pidsList);
      connections = response.connections || [];
    } catch (err: any) {
      error = err.message || 'Failed to load network connections';
      connections = [];
    } finally {
      loading = false;
    }
  }

  function toggleAutoRefresh() {
    autoRefresh = !autoRefresh;

    if (autoRefresh) {
      refreshInterval = window.setInterval(() => {
        loadConnections();
      }, 5000); // Refresh every 5 seconds
      addNotification('info', 'Auto-refresh enabled (5s)');
    } else {
      if (refreshInterval) {
        clearInterval(refreshInterval);
        refreshInterval = null;
      }
      addNotification('info', 'Auto-refresh disabled');
    }
  }

  function getServerName(pid: number): string {
    const server = $servers.find(s => s.pid === pid);
    return server ? server.name : `PID ${pid}`;
  }

  function getStateClass(state: string): string {
    switch (state.toUpperCase()) {
      case 'ESTABLISHED':
        return 'state-established';
      case 'LISTENING':
        return 'state-listening';
      case 'TIME_WAIT':
      case 'CLOSE_WAIT':
        return 'state-closing';
      default:
        return 'state-other';
    }
  }

  // Cleanup on unmount
  onMount(() => {
    return () => {
      if (refreshInterval) {
        clearInterval(refreshInterval);
      }
    };
  });
</script>

<div class="netstat-view">
  <div class="view-header">
    <div class="header-left">
      <h2>Network Connections</h2>
      <p class="subtitle text-secondary">
        Active network connections for running MCP servers
      </p>
    </div>
    <div class="header-actions">
      <button
        class="btn-icon"
        class:active={autoRefresh}
        on:click={toggleAutoRefresh}
        title={autoRefresh ? 'Disable auto-refresh' : 'Enable auto-refresh (5s)'}
      >
        {autoRefresh ? '⏸️' : '▶️'}
      </button>
      <button
        class="btn-secondary"
        on:click={loadConnections}
        disabled={loading}
      >
        {loading ? 'Loading...' : '🔄 Refresh'}
      </button>
    </div>
  </div>

  <div class="view-content">
    {#if loading && connections.length === 0}
      <div class="loading-state">
        <div class="spinner"></div>
        <p>Loading network connections...</p>
      </div>
    {:else if error}
      <div class="error-state">
        <div class="error-icon">⚠️</div>
        <h3>Unable to load connections</h3>
        <p class="text-secondary">{error}</p>
        <button class="btn-secondary" on:click={loadConnections}>Try Again</button>
      </div>
    {:else if connections.length === 0}
      <div class="empty-state">
        <div class="empty-icon">🌐</div>
        <h3>No active connections</h3>
        <p class="text-secondary">
          {#if $runningServers.length === 0}
            Start a server to see its network connections.
          {:else}
            No network connections found for running servers.
          {/if}
        </p>
      </div>
    {:else}
      <div class="table-wrapper">
        <table class="connections-table">
          <thead>
            <tr>
              <th>Server</th>
              <th>Protocol</th>
              <th>Local Address</th>
              <th>Remote Address</th>
              <th>State</th>
              <th>PID</th>
            </tr>
          </thead>
          <tbody>
            {#each connections as conn (conn.localAddress + conn.remoteAddress + conn.pid)}
              <tr>
                <td class="server-name">{getServerName(conn.pid)}</td>
                <td class="protocol">{conn.protocol}</td>
                <td class="address">{conn.localAddress}</td>
                <td class="address">{conn.remoteAddress}</td>
                <td class="state">
                  <span class="state-badge {getStateClass(conn.state)}">
                    {conn.state}
                  </span>
                </td>
                <td class="pid mono">{conn.pid}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>

      <div class="table-footer">
        <span class="text-secondary">
          {connections.length} active connection{connections.length !== 1 ? 's' : ''}
        </span>
      </div>
    {/if}
  </div>

  <!-- Backend API notice -->
  <div class="api-notice">
    <strong>⚠️ Backend API Required:</strong>
    <code>GET /api/v1/netstat?pids=&lt;comma-separated&gt;</code>
    <br />
    Response: <code>{'{ connections: [{ protocol, localAddress, remoteAddress, state, pid }] }'}</code>
  </div>
</div>

<style>
  .netstat-view {
    display: flex;
    flex-direction: column;
    height: 100%;
    background-color: var(--bg-secondary);
    border-radius: var(--radius-md);
    overflow: hidden;
  }

  .view-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    padding: var(--spacing-lg);
    border-bottom: 1px solid var(--border-color);
    background-color: var(--bg-tertiary);
    gap: var(--spacing-md);
  }

  .header-left h2 {
    margin: 0 0 var(--spacing-xs) 0;
    font-size: var(--font-size-xl);
    color: var(--text-primary);
  }

  .subtitle {
    margin: 0;
    font-size: var(--font-size-sm);
  }

  .header-actions {
    display: flex;
    gap: var(--spacing-sm);
    align-items: center;
  }

  .btn-icon {
    padding: var(--spacing-xs);
    min-width: 32px;
    font-size: var(--font-size-md);
  }

  .btn-icon.active {
    background-color: var(--accent-primary);
    border-color: var(--accent-primary);
  }

  .view-content {
    flex: 1;
    overflow: auto;
    display: flex;
    flex-direction: column;
  }

  /* Loading, error, empty states */
  .loading-state,
  .error-state,
  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: var(--spacing-xl);
    text-align: center;
    flex: 1;
  }

  .spinner {
    width: 40px;
    height: 40px;
    border: 4px solid var(--border-color);
    border-top-color: var(--accent-primary);
    border-radius: 50%;
    animation: spin 1s linear infinite;
    margin-bottom: var(--spacing-md);
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  .error-icon,
  .empty-icon {
    font-size: 4rem;
    margin-bottom: var(--spacing-md);
    opacity: 0.5;
  }

  .error-state h3,
  .empty-state h3 {
    margin-bottom: var(--spacing-sm);
    color: var(--text-primary);
  }

  /* Table */
  .table-wrapper {
    flex: 1;
    overflow: auto;
  }

  .connections-table {
    width: 100%;
    border-collapse: collapse;
  }

  .connections-table thead {
    position: sticky;
    top: 0;
    background-color: var(--bg-tertiary);
    z-index: 10;
  }

  .connections-table th {
    text-align: left;
    padding: var(--spacing-md);
    font-weight: 600;
    font-size: var(--font-size-sm);
    color: var(--text-primary);
    border-bottom: 2px solid var(--border-color);
    white-space: nowrap;
  }

  .connections-table td {
    padding: var(--spacing-md);
    border-bottom: 1px solid var(--border-color);
    font-size: var(--font-size-sm);
  }

  .connections-table tbody tr:hover {
    background-color: var(--bg-hover);
  }

  .server-name {
    font-weight: 500;
    color: var(--text-primary);
  }

  .protocol {
    color: var(--text-secondary);
    font-weight: 500;
  }

  .address {
    font-family: var(--font-mono);
    font-size: var(--font-size-xs);
    color: var(--text-secondary);
  }

  .state-badge {
    display: inline-block;
    padding: 2px 8px;
    border-radius: var(--radius-sm);
    font-size: var(--font-size-xs);
    font-weight: 600;
    text-transform: uppercase;
  }

  .state-established {
    background-color: rgba(76, 175, 80, 0.2);
    color: var(--status-running);
  }

  .state-listening {
    background-color: rgba(33, 150, 243, 0.2);
    color: var(--status-starting);
  }

  .state-closing {
    background-color: rgba(255, 152, 0, 0.2);
    color: var(--status-error);
  }

  .state-other {
    background-color: var(--bg-tertiary);
    color: var(--text-muted);
  }

  .pid {
    font-family: var(--font-mono);
    font-size: var(--font-size-xs);
    color: var(--text-secondary);
  }

  .table-footer {
    padding: var(--spacing-md);
    border-top: 1px solid var(--border-color);
    background-color: var(--bg-tertiary);
    text-align: center;
    font-size: var(--font-size-sm);
  }

  /* API notice */
  .api-notice {
    padding: var(--spacing-md);
    background-color: rgba(255, 152, 0, 0.1);
    border-top: 2px solid var(--status-error);
    font-size: var(--font-size-xs);
    color: var(--text-secondary);
    line-height: 1.5;
  }

  .api-notice code {
    background-color: var(--bg-primary);
    padding: 2px 4px;
    border-radius: var(--radius-sm);
    font-family: var(--font-mono);
    color: var(--text-primary);
  }

  /* Responsive */
  @media (max-width: 768px) {
    .view-header {
      flex-direction: column;
      align-items: stretch;
    }

    .header-actions {
      width: 100%;
      justify-content: flex-end;
    }

    .address {
      font-size: 0.7rem;
    }
  }
</style>
