<script lang="ts">
  import { servers, serverLogs, selectedServerId } from '../stores/stores';
  import type { LogEntry, LogSeverity } from '../stores/stores';
  import { onMount, afterUpdate } from 'svelte';

  // Filter states
  let filterServerId: string | null = null;
  let filterSeverity: LogSeverity | null = null;
  let searchQuery = '';
  let searchInput = ''; // Temporary input value for debouncing
  let searchTimeout: number | null = null;
  let autoScroll = true;

  // Log container reference for auto-scroll
  let logContainer: HTMLDivElement;
  let lastScrollHeight = 0;

  // Get all logs from all servers, flattened and sorted by timestamp
  $: allLogs = Object.entries($serverLogs).flatMap(([serverId, logs]) =>
    logs.map(log => ({ ...log, serverId }))
  ).sort((a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime());

  // Apply filters to logs
  $: filteredLogs = allLogs.filter(log => {
    // Filter by server
    if (filterServerId && log.serverId !== filterServerId) return false;

    // Filter by severity
    if (filterSeverity && log.severity !== filterSeverity) return false;

    // Filter by search query
    if (searchQuery && !log.message.toLowerCase().includes(searchQuery.toLowerCase())) {
      return false;
    }

    return true;
  });

  // Auto-scroll to bottom when new logs arrive
  afterUpdate(() => {
    if (autoScroll && logContainer) {
      const shouldScroll = logContainer.scrollHeight !== lastScrollHeight;
      if (shouldScroll) {
        logContainer.scrollTop = logContainer.scrollHeight;
        lastScrollHeight = logContainer.scrollHeight;
      }
    }
  });

  // Detect manual scroll away from bottom
  function handleScroll() {
    if (!logContainer) return;

    const isAtBottom = logContainer.scrollHeight - logContainer.scrollTop <= logContainer.clientHeight + 50;
    autoScroll = isAtBottom;
  }

  // Clear all logs
  function clearLogs() {
    if (confirm('Clear all logs?')) {
      Object.keys($serverLogs).forEach(serverId => {
        serverLogs.update(logs => {
          logs[serverId] = [];
          return logs;
        });
      });
    }
  }

  // Export logs to clipboard
  function exportLogs() {
    const logText = filteredLogs.map(log =>
      `[${formatTimestamp(log.timestamp)}] [${log.severity.toUpperCase()}] [${getServerName(log.serverId)}] ${log.message}`
    ).join('\n');

    navigator.clipboard.writeText(logText).then(() => {
      alert('Logs copied to clipboard!');
    }).catch(err => {
      console.error('Failed to copy logs:', err);
    });
  }

  // Format timestamp for display
  function formatTimestamp(timestamp: string): string {
    const date = new Date(timestamp);
    const hours = date.getHours().toString().padStart(2, '0');
    const minutes = date.getMinutes().toString().padStart(2, '0');
    const seconds = date.getSeconds().toString().padStart(2, '0');
    const ms = date.getMilliseconds().toString().padStart(3, '0');
    return `${hours}:${minutes}:${seconds}.${ms}`;
  }

  // Get server name by ID
  function getServerName(serverId: string): string {
    const server = $servers.find(s => s.id === serverId);
    return server ? server.name : serverId;
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

  // Select log entry (for future detail view)
  function selectLog(log: LogEntry & { serverId: string }) {
    selectedServerId.set(log.serverId);
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

<div class="log-viewer">
  <!-- Toolbar -->
  <div class="log-toolbar">
    <div class="toolbar-left">
      <h3 class="log-title">Log Viewer</h3>
      <span class="log-count text-muted">
        {filteredLogs.length} {filteredLogs.length === 1 ? 'entry' : 'entries'}
        {#if filteredLogs.length !== allLogs.length}
          <span class="text-secondary">/ {allLogs.length} total</span>
        {/if}
      </span>
    </div>

    <div class="toolbar-right">
      <!-- Server filter -->
      <select class="filter-select" bind:value={filterServerId}>
        <option value={null}>All Servers</option>
        {#each $servers as server}
          <option value={server.id}>{server.name}</option>
        {/each}
      </select>

      <!-- Severity filter -->
      <select class="filter-select" bind:value={filterSeverity}>
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

      <!-- Export button -->
      <button
        class="btn-icon"
        on:click={exportLogs}
        disabled={filteredLogs.length === 0}
        title="Copy logs to clipboard"
      >
        📋
      </button>

      <!-- Clear button -->
      <button
        class="btn-icon"
        on:click={clearLogs}
        disabled={allLogs.length === 0}
        title="Clear all logs"
      >
        🗑️
      </button>
    </div>
  </div>

  <!-- Log entries -->
  <div
    class="log-entries"
    bind:this={logContainer}
    on:scroll={handleScroll}
  >
    {#if filteredLogs.length === 0}
      <div class="empty-logs">
        <div class="empty-icon">📋</div>
        <p class="text-secondary">
          {#if allLogs.length === 0}
            No logs yet. Start a server to see logs.
          {:else}
            No logs match the current filters.
          {/if}
        </p>
      </div>
    {:else}
      {#each filteredLogs as log (log.timestamp + log.serverId)}
        <div
          class="log-entry log-{log.severity}"
          on:click={() => selectLog(log)}
          role="button"
          tabindex="0"
        >
          <span class="log-icon">{getSeverityIcon(log.severity)}</span>
          <span class="log-timestamp">{formatTimestamp(log.timestamp)}</span>
          <span class="log-severity">[{log.severity.toUpperCase()}]</span>
          <span class="log-server">[{getServerName(log.serverId)}]</span>
          <span class="log-message">{log.message}</span>
        </div>
      {/each}
    {/if}
  </div>
</div>

<style>
  .log-viewer {
    display: flex;
    flex-direction: column;
    height: 100%;
    background-color: var(--bg-secondary);
  }

  /* Toolbar */
  .log-toolbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: var(--spacing-sm) var(--spacing-md);
    border-bottom: 1px solid var(--border-color);
    background-color: var(--bg-tertiary);
    gap: var(--spacing-md);
    flex-wrap: wrap;
  }

  .toolbar-left {
    display: flex;
    align-items: center;
    gap: var(--spacing-md);
  }

  .toolbar-right {
    display: flex;
    align-items: center;
    gap: var(--spacing-sm);
    flex-wrap: wrap;
  }

  .log-title {
    margin: 0;
    font-size: var(--font-size-md);
    font-weight: 600;
    color: var(--text-primary);
  }

  .log-count {
    font-size: var(--font-size-sm);
    white-space: nowrap;
  }

  /* Filters */
  .filter-select {
    min-width: 120px;
    font-size: var(--font-size-sm);
  }

  .search-input {
    width: 200px;
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

  /* Log entries container */
  .log-entries {
    flex: 1;
    overflow-y: auto;
    overflow-x: hidden;
    padding: var(--spacing-sm);
    background-color: var(--bg-primary);
    font-family: var(--font-mono);
    font-size: var(--font-size-sm);
  }

  /* Empty state */
  .empty-logs {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100%;
    padding: var(--spacing-xl);
    text-align: center;
  }

  .empty-icon {
    font-size: 3rem;
    margin-bottom: var(--spacing-md);
    opacity: 0.5;
  }

  /* Log entry styling (FR-021) */
  .log-entry {
    display: flex;
    gap: var(--spacing-sm);
    padding: var(--spacing-xs) var(--spacing-sm);
    border-left: 3px solid transparent;
    cursor: pointer;
    transition: background-color var(--transition-fast);
    word-break: break-word;
    line-height: 1.6;
  }

  .log-entry:hover {
    background-color: var(--bg-hover);
  }

  /* Severity colors (FR-021) */
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

  /* Log entry parts */
  .log-icon {
    flex-shrink: 0;
    font-size: var(--font-size-md);
  }

  .log-timestamp {
    color: var(--text-muted);
    white-space: nowrap;
    flex-shrink: 0;
    font-family: var(--font-mono);
  }

  .log-severity {
    color: var(--text-secondary);
    font-weight: 600;
    white-space: nowrap;
    flex-shrink: 0;
    min-width: 70px;
  }

  .log-entry.log-info .log-severity {
    color: var(--log-info);
  }

  .log-entry.log-success .log-severity {
    color: var(--log-success);
  }

  .log-entry.log-warning .log-severity {
    color: var(--log-warning);
  }

  .log-entry.log-error .log-severity {
    color: var(--log-error);
  }

  .log-server {
    color: var(--text-secondary);
    white-space: nowrap;
    flex-shrink: 0;
    max-width: 150px;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .log-message {
    color: var(--text-primary);
    flex: 1;
    word-break: break-word;
  }

  /* Responsive */
  @media (max-width: 1024px) {
    .log-server {
      display: none;
    }
  }

  @media (max-width: 768px) {
    .log-toolbar {
      flex-direction: column;
      align-items: stretch;
    }

    .toolbar-left,
    .toolbar-right {
      width: 100%;
      justify-content: space-between;
    }

    .search-input {
      width: 100%;
    }

    .log-icon {
      display: none;
    }

    .log-timestamp {
      font-size: var(--font-size-xs);
    }
  }
</style>
