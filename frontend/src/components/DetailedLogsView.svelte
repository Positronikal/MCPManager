<script lang="ts">
  import { onMount } from 'svelte';
  import type { LogEntry, LogSeverity } from '../stores/stores';
  import { api } from '../services/api';
  import { servers, addNotification } from '../stores/stores';

  // Props
  export let serverId: string;
  export let onClose: () => void;

  // State
  let logs: LogEntry[] = [];
  let filteredLogs: LogEntry[] = [];
  let severityFilter: LogSeverity | null = null;
  let searchQuery = '';
  let searchInput = ''; // Temporary input value for debouncing
  let searchTimeout: number | null = null;
  let loading = true;
  let errorMessage = '';
  let limit = 100;
  let offset = 0;
  let hasMore = true;
  let total = 0;

  // Auto-scroll
  let logContainer: HTMLDivElement;
  let autoScroll = false; // Off by default for detailed view

  // Get server name
  $: serverName = $servers.find(s => s.id === serverId)?.name || serverId;

  // Load logs on mount
  onMount(async () => {
    await loadLogs();
  });

  // Load logs from API
  async function loadLogs() {
    loading = true;
    errorMessage = '';
    try {
      const response = await api.monitoring.getServerLogs(serverId, {
        severity: severityFilter || undefined,
        limit,
        offset
      });
      logs = response.logs;
      total = response.total;
      hasMore = response.hasMore;
      applyFilters();
      loading = false;
    } catch (error: any) {
      errorMessage = `Failed to load logs: ${error.message}`;
      loading = false;
    }
  }

  // Apply client-side filters
  function applyFilters() {
    filteredLogs = logs.filter(log => {
      // Filter by severity
      if (severityFilter && log.severity !== severityFilter) return false;

      // Filter by search query
      if (searchQuery && !log.message.toLowerCase().includes(searchQuery.toLowerCase())) {
        return false;
      }

      return true;
    });
  }

  // Watch filters and reapply
  $: {
    severityFilter;
    searchQuery;
    applyFilters();
  }

  // Load more logs (pagination)
  async function loadMore() {
    offset += limit;
    await loadLogs();
  }

  // Refresh logs
  async function refresh() {
    offset = 0;
    await loadLogs();
    addNotification('success', 'Logs refreshed');
  }

  // Export logs to clipboard
  function exportLogs() {
    const logText = filteredLogs.map(log =>
      `[${formatTimestamp(log.timestamp)}] [${log.severity.toUpperCase()}] ${log.message}`
    ).join('\n');

    navigator.clipboard.writeText(logText).then(() => {
      addNotification('success', 'Logs copied to clipboard');
    }).catch(err => {
      console.error('Failed to copy logs:', err);
      addNotification('error', 'Failed to copy logs');
    });
  }

  // Format timestamp for display
  function formatTimestamp(timestamp: string): string {
    const date = new Date(timestamp);
    const year = date.getFullYear();
    const month = (date.getMonth() + 1).toString().padStart(2, '0');
    const day = date.getDate().toString().padStart(2, '0');
    const hours = date.getHours().toString().padStart(2, '0');
    const minutes = date.getMinutes().toString().padStart(2, '0');
    const seconds = date.getSeconds().toString().padStart(2, '0');
    const ms = date.getMilliseconds().toString().padStart(3, '0');
    return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}.${ms}`;
  }

  // Get severity icon
  function getSeverityIcon(severity: LogSeverity): string {
    switch (severity) {
      case 'info': return 'ℹ️';
      case 'success': return '✅';
      case 'warning': return '⚠️';
      case 'error': return '❌';
      default: return '📝';
    }
  }

  // Handle keyboard shortcuts
  function handleKeydown(event: KeyboardEvent) {
    if (event.key === 'Escape') {
      onClose();
    } else if (event.ctrlKey && event.key === 'r') {
      event.preventDefault();
      refresh();
    }
  }

  // Scroll to bottom if auto-scroll enabled
  function scrollToBottom() {
    if (autoScroll && logContainer) {
      logContainer.scrollTop = logContainer.scrollHeight;
    }
  }

  // Watch for new logs and auto-scroll
  $: if (filteredLogs.length > 0) {
    setTimeout(scrollToBottom, 100);
  }

  // T-E025: Debounce search input (300ms) for UI responsiveness
  function handleSearchInput(event: Event) {
    const target = event.target as HTMLInputElement;
    searchInput = target.value;

    if (searchTimeout) {
      clearTimeout(searchTimeout);
    }

    searchTimeout = window.setTimeout(() => {
      searchQuery = searchInput;
    }, 300);
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<!-- Modal backdrop -->
<div class="modal-backdrop" on:click={onClose} role="presentation">
  <!-- Modal content -->
  <div class="modal-content" on:click|stopPropagation role="dialog" aria-modal="true" aria-labelledby="modal-title">
    <!-- Header -->
    <div class="modal-header">
      <div class="header-left">
        <h2 id="modal-title" class="modal-title">
          Detailed Logs - {serverName}
        </h2>
        <span class="log-count text-muted">
          {filteredLogs.length} / {total} entries
        </span>
      </div>
      <button class="btn-close" on:click={onClose} aria-label="Close">&times;</button>
    </div>

    <!-- Toolbar -->
    <div class="modal-toolbar">
      <!-- Severity filter -->
      <select class="filter-select" bind:value={severityFilter}>
        <option value={null}>All Severities</option>
        <option value="info">INFO</option>
        <option value="success">SUCCESS</option>
        <option value="warning">WARNING</option>
        <option value="error">ERROR</option>
      </select>

      <!-- Search input -->
      <input
        type="search"
        class="search-input"
        value={searchInput}
        on:input={handleSearchInput}
        placeholder="Search logs..."
      />

      <!-- Auto-scroll toggle -->
      <button
        class="btn-icon"
        class:active={autoScroll}
        on:click={() => autoScroll = !autoScroll}
        title={autoScroll ? 'Auto-scroll enabled' : 'Auto-scroll disabled'}
      >
        {autoScroll ? '🔽' : '⏸️'}
      </button>

      <!-- Refresh button -->
      <button
        class="btn-icon"
        on:click={refresh}
        disabled={loading}
        title="Refresh logs (Ctrl+R)"
      >
        🔄
      </button>

      <!-- Export button -->
      <button
        class="btn-icon"
        on:click={exportLogs}
        disabled={filteredLogs.length === 0}
        title="Copy logs to clipboard"
      >
        📋
      </button>
    </div>

    <!-- Body -->
    <div class="modal-body">
      {#if loading && logs.length === 0}
        <div class="loading-state">
          <div class="spinner"></div>
          <p>Loading logs...</p>
        </div>
      {:else if errorMessage}
        <div class="error-message" role="alert">
          <span class="error-icon">❌</span>
          <span>{errorMessage}</span>
        </div>
      {:else if filteredLogs.length === 0}
        <div class="empty-logs">
          <div class="empty-icon">📋</div>
          <p class="text-secondary">
            {#if logs.length === 0}
              No logs available for this server.
            {:else}
              No logs match the current filters.
            {/if}
          </p>
        </div>
      {:else}
        <div class="log-entries" bind:this={logContainer}>
          {#each filteredLogs as log (log.id || log.timestamp)}
            <div class="log-entry log-{log.severity}">
              <div class="log-header">
                <span class="log-icon">{getSeverityIcon(log.severity)}</span>
                <span class="log-timestamp">{formatTimestamp(log.timestamp)}</span>
                <span class="log-severity log-severity-{log.severity}">
                  [{log.severity.toUpperCase()}]
                </span>
              </div>
              <div class="log-message">{log.message}</div>
              {#if log.metadata && Object.keys(log.metadata).length > 0}
                <details class="log-metadata">
                  <summary class="metadata-summary">Metadata</summary>
                  <pre class="metadata-content">{JSON.stringify(log.metadata, null, 2)}</pre>
                </details>
              {/if}
            </div>
          {/each}
        </div>

        <!-- Load more button -->
        {#if hasMore && !loading}
          <div class="load-more-container">
            <button class="btn-secondary" on:click={loadMore}>
              Load More Logs
            </button>
          </div>
        {/if}
      {/if}
    </div>

    <!-- Footer -->
    <div class="modal-footer">
      <button class="btn-secondary" on:click={onClose}>Close</button>
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
    background-color: rgba(0, 0, 0, 0.7);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
    padding: var(--spacing-lg);
  }

  /* Modal content */
  .modal-content {
    background-color: var(--bg-secondary);
    border-radius: var(--radius-lg);
    box-shadow: var(--shadow-lg);
    width: 100%;
    max-width: 900px;
    max-height: 90vh;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  /* Modal header */
  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: var(--spacing-md) var(--spacing-lg);
    border-bottom: 1px solid var(--border-color);
    background-color: var(--bg-tertiary);
  }

  .header-left {
    display: flex;
    align-items: center;
    gap: var(--spacing-md);
    flex: 1;
  }

  .modal-title {
    margin: 0;
    font-size: var(--font-size-lg);
    color: var(--text-primary);
  }

  .log-count {
    font-size: var(--font-size-sm);
    white-space: nowrap;
  }

  .btn-close {
    background: transparent;
    border: none;
    color: var(--text-primary);
    font-size: 2rem;
    line-height: 1;
    cursor: pointer;
    padding: 0;
    width: 32px;
    height: 32px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: var(--radius-sm);
    transition: background-color var(--transition-fast);
  }

  .btn-close:hover {
    background-color: var(--bg-hover);
  }

  /* Toolbar */
  .modal-toolbar {
    display: flex;
    align-items: center;
    gap: var(--spacing-sm);
    padding: var(--spacing-sm) var(--spacing-lg);
    border-bottom: 1px solid var(--border-color);
    background-color: var(--bg-tertiary);
    flex-wrap: wrap;
  }

  .filter-select {
    min-width: 140px;
    font-size: var(--font-size-sm);
  }

  .search-input {
    flex: 1;
    min-width: 200px;
    font-size: var(--font-size-sm);
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

  /* Modal body */
  .modal-body {
    flex: 1;
    overflow: hidden;
    display: flex;
    flex-direction: column;
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
    margin: var(--spacing-md);
    background-color: var(--notif-error-bg);
    border: 1px solid var(--notif-error-border);
    border-radius: var(--radius-md);
    color: var(--text-primary);
  }

  .error-icon {
    font-size: var(--font-size-xl);
  }

  /* Empty state */
  .empty-logs {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: var(--spacing-xl);
    text-align: center;
  }

  .empty-icon {
    font-size: 3rem;
    margin-bottom: var(--spacing-md);
    opacity: 0.5;
  }

  /* Log entries */
  .log-entries {
    flex: 1;
    overflow-y: auto;
    overflow-x: hidden;
    padding: var(--spacing-md);
    background-color: var(--bg-primary);
  }

  .log-entry {
    display: flex;
    flex-direction: column;
    gap: var(--spacing-xs);
    padding: var(--spacing-sm) var(--spacing-md);
    margin-bottom: var(--spacing-sm);
    border-left: 3px solid transparent;
    border-radius: var(--radius-sm);
    background-color: var(--bg-secondary);
    font-family: var(--font-mono);
    font-size: var(--font-size-sm);
  }

  .log-entry.log-info {
    border-left-color: var(--log-info);
  }

  .log-entry.log-success {
    border-left-color: var(--log-success);
  }

  .log-entry.log-warning {
    border-left-color: var(--log-warning);
  }

  .log-entry.log-error {
    border-left-color: var(--log-error);
    background-color: rgba(244, 67, 54, 0.05);
  }

  .log-header {
    display: flex;
    align-items: center;
    gap: var(--spacing-sm);
  }

  .log-icon {
    font-size: var(--font-size-md);
  }

  .log-timestamp {
    color: var(--text-muted);
    white-space: nowrap;
    font-family: var(--font-mono);
    font-size: var(--font-size-xs);
  }

  .log-severity {
    font-weight: 600;
    white-space: nowrap;
    font-size: var(--font-size-xs);
  }

  .log-severity-info { color: var(--log-info); }
  .log-severity-success { color: var(--log-success); }
  .log-severity-warning { color: var(--log-warning); }
  .log-severity-error { color: var(--log-error); }

  .log-message {
    color: var(--text-primary);
    word-break: break-word;
    line-height: 1.5;
  }

  /* Metadata */
  .log-metadata {
    margin-top: var(--spacing-xs);
  }

  .metadata-summary {
    color: var(--text-secondary);
    font-size: var(--font-size-xs);
    cursor: pointer;
    user-select: none;
  }

  .metadata-summary:hover {
    color: var(--accent-primary);
  }

  .metadata-content {
    margin-top: var(--spacing-xs);
    padding: var(--spacing-sm);
    background-color: var(--bg-primary);
    border: 1px solid var(--border-color);
    border-radius: var(--radius-sm);
    color: var(--text-secondary);
    font-size: var(--font-size-xs);
    overflow-x: auto;
  }

  /* Load more */
  .load-more-container {
    display: flex;
    justify-content: center;
    padding: var(--spacing-md);
    border-top: 1px solid var(--border-color);
  }

  /* Modal footer */
  .modal-footer {
    display: flex;
    justify-content: flex-end;
    gap: var(--spacing-sm);
    padding: var(--spacing-md) var(--spacing-lg);
    border-top: 1px solid var(--border-color);
    background-color: var(--bg-tertiary);
  }

  /* Responsive */
  @media (max-width: 768px) {
    .modal-content {
      max-width: 100%;
      max-height: 100vh;
      border-radius: 0;
    }

    .modal-backdrop {
      padding: 0;
    }

    .modal-toolbar {
      flex-direction: column;
      align-items: stretch;
    }

    .search-input {
      width: 100%;
    }
  }
</style>
