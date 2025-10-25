<script lang="ts">
  import { servers, filteredServers, serverFilters, addNotification } from '../stores/stores';
  import { BrowserOpenURL } from '../../wailsjs/runtime/runtime';

  // NOTE: This component uses Wails BrowserOpenURL which works for opening directories
  // Alternative: Backend API endpoint POST /api/v1/explorer?path=<path>
  // Response: { success: boolean, message: string }
  // Backend would launch: explorer (Windows), open (macOS), xdg-open (Linux)

  let searchQuery = '';
  let searchInput = ''; // Temporary input value for debouncing
  let searchTimeout: number | null = null;
  let selectedSource: string | null = null;

  // Filter servers by search and source
  $: displayedServers = $servers.filter(server => {
    if (searchQuery && !server.name.toLowerCase().includes(searchQuery.toLowerCase()) &&
        !server.installationPath.toLowerCase().includes(searchQuery.toLowerCase())) {
      return false;
    }
    if (selectedSource && server.source !== selectedSource) {
      return false;
    }
    return true;
  });

  async function openDirectory(path: string, serverName: string) {
    try {
      // Extract directory from full path (remove filename)
      let directory = path;

      // If path ends with an executable or file, get its parent directory
      if (path.match(/\.(exe|sh|py|js|ts)$/i)) {
        const lastSlash = Math.max(path.lastIndexOf('/'), path.lastIndexOf('\\'));
        if (lastSlash > 0) {
          directory = path.substring(0, lastSlash);
        }
      }

      await BrowserOpenURL(directory);
      addNotification('success', `Opened ${serverName} directory`);
    } catch (error: any) {
      addNotification('error', `Failed to open directory: ${error.message || error}`);
    }
  }

  function getDirectoryFromPath(path: string): string {
    const lastSlash = Math.max(path.lastIndexOf('/'), path.lastIndexOf('\\'));
    if (lastSlash > 0 && path.match(/\.(exe|sh|py|js|ts)$/i)) {
      return path.substring(0, lastSlash);
    }
    return path;
  }

  function getFileName(path: string): string {
    const lastSlash = Math.max(path.lastIndexOf('/'), path.lastIndexOf('\\'));
    return lastSlash >= 0 ? path.substring(lastSlash + 1) : path;
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

<div class="explorer-view">
  <div class="view-header">
    <div class="header-left">
      <h2>Server Directories</h2>
      <p class="subtitle text-secondary">
        Open server installation directories in your file explorer
      </p>
    </div>
    <div class="header-actions">
      <select class="filter-select" bind:value={selectedSource}>
        <option value={null}>All Sources</option>
        <option value="client_config">Client Config</option>
        <option value="filesystem">Filesystem</option>
        <option value="process">Process</option>
      </select>
      <input
        type="search"
        class="search-input"
        value={searchInput}
        on:input={handleSearchInput}
        placeholder="Search servers..."
      />
    </div>
  </div>

  <div class="view-content">
    {#if displayedServers.length === 0}
      <div class="empty-state">
        <div class="empty-icon">📁</div>
        <h3>No servers found</h3>
        <p class="text-secondary">
          {#if $servers.length === 0}
            Click "Refresh" to discover MCP servers.
          {:else if searchQuery || selectedSource}
            No servers match your filters.
          {:else}
            No servers available.
          {/if}
        </p>
      </div>
    {:else}
      <div class="servers-grid">
        {#each displayedServers as server (server.id)}
          <div class="server-card">
            <div class="server-header">
              <div class="server-icon">
                {#if server.source === 'client_config'}
                  ⚙️
                {:else if server.source === 'filesystem'}
                  📂
                {:else if server.source === 'process'}
                  🔄
                {:else}
                  🖥️
                {/if}
              </div>
              <div class="server-info">
                <h3 class="server-name">{server.name}</h3>
                <span class="source-badge badge-{server.source}">
                  {server.source}
                </span>
              </div>
            </div>

            <div class="server-path">
              <div class="path-label text-muted">Installation Path:</div>
              <div class="path-value" title={server.installationPath}>
                <span class="directory">{getDirectoryFromPath(server.installationPath)}</span>
                {#if server.installationPath.match(/\.(exe|sh|py|js|ts)$/i)}
                  <span class="separator">/</span>
                  <span class="filename">{getFileName(server.installationPath)}</span>
                {/if}
              </div>
            </div>

            <div class="server-meta">
              {#if server.version}
                <div class="meta-item">
                  <span class="meta-label">Version:</span>
                  <span class="meta-value">{server.version}</span>
                </div>
              {/if}
              {#if server.pid}
                <div class="meta-item">
                  <span class="meta-label">PID:</span>
                  <span class="meta-value mono">{server.pid}</span>
                </div>
              {/if}
              <div class="meta-item">
                <span class="meta-label">Status:</span>
                <span class="status-badge status-{server.status.state}">
                  {server.status.state}
                </span>
              </div>
            </div>

            <button
              class="btn-primary btn-block"
              on:click={() => openDirectory(server.installationPath, server.name)}
            >
              📂 Open Directory
            </button>
          </div>
        {/each}
      </div>

      <div class="table-footer">
        <span class="text-secondary">
          Showing {displayedServers.length} of {$servers.length} server{$servers.length !== 1 ? 's' : ''}
        </span>
      </div>
    {/if}
  </div>
</div>

<style>
  .explorer-view {
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

  .filter-select {
    min-width: 140px;
    font-size: var(--font-size-sm);
  }

  .search-input {
    width: 200px;
    font-size: var(--font-size-sm);
  }

  .view-content {
    flex: 1;
    overflow: auto;
    padding: var(--spacing-lg);
  }

  /* Empty state */
  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: var(--spacing-xl);
    text-align: center;
    min-height: 400px;
  }

  .empty-icon {
    font-size: 4rem;
    margin-bottom: var(--spacing-md);
    opacity: 0.5;
  }

  .empty-state h3 {
    margin-bottom: var(--spacing-sm);
    color: var(--text-primary);
  }

  /* Servers grid */
  .servers-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
    gap: var(--spacing-lg);
  }

  .server-card {
    background-color: var(--bg-primary);
    border: 1px solid var(--border-color);
    border-radius: var(--radius-md);
    padding: var(--spacing-lg);
    display: flex;
    flex-direction: column;
    gap: var(--spacing-md);
    transition: all var(--transition-fast);
  }

  .server-card:hover {
    border-color: var(--accent-primary);
    box-shadow: var(--shadow-md);
  }

  .server-header {
    display: flex;
    gap: var(--spacing-md);
    align-items: flex-start;
  }

  .server-icon {
    font-size: 2rem;
    flex-shrink: 0;
  }

  .server-info {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: var(--spacing-xs);
  }

  .server-name {
    margin: 0;
    font-size: var(--font-size-md);
    color: var(--text-primary);
    font-weight: 600;
  }

  .source-badge {
    display: inline-block;
    padding: 2px 8px;
    border-radius: var(--radius-sm);
    font-size: var(--font-size-xs);
    font-weight: 500;
    text-transform: uppercase;
    align-self: flex-start;
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

  .server-path {
    display: flex;
    flex-direction: column;
    gap: var(--spacing-xs);
  }

  .path-label {
    font-size: var(--font-size-xs);
    font-weight: 600;
    text-transform: uppercase;
  }

  .path-value {
    font-family: var(--font-mono);
    font-size: var(--font-size-xs);
    color: var(--text-secondary);
    word-break: break-all;
    line-height: 1.5;
    padding: var(--spacing-sm);
    background-color: var(--bg-secondary);
    border-radius: var(--radius-sm);
    border: 1px solid var(--border-color);
  }

  .directory {
    color: var(--text-secondary);
  }

  .separator {
    color: var(--text-muted);
  }

  .filename {
    color: var(--text-primary);
    font-weight: 600;
  }

  .server-meta {
    display: flex;
    flex-direction: column;
    gap: var(--spacing-xs);
    padding: var(--spacing-sm);
    background-color: var(--bg-secondary);
    border-radius: var(--radius-sm);
  }

  .meta-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-size: var(--font-size-sm);
  }

  .meta-label {
    color: var(--text-muted);
    font-weight: 500;
  }

  .meta-value {
    color: var(--text-primary);
  }

  .mono {
    font-family: var(--font-mono);
  }

  .status-badge {
    display: inline-block;
    padding: 2px 8px;
    border-radius: var(--radius-sm);
    font-size: var(--font-size-xs);
    font-weight: 600;
    text-transform: uppercase;
  }

  .status-running {
    background-color: rgba(76, 175, 80, 0.2);
    color: var(--status-running);
  }

  .status-stopped {
    background-color: rgba(244, 67, 54, 0.2);
    color: var(--status-stopped);
  }

  .status-starting {
    background-color: rgba(33, 150, 243, 0.2);
    color: var(--status-starting);
  }

  .status-error {
    background-color: rgba(255, 152, 0, 0.2);
    color: var(--status-error);
  }

  .btn-block {
    width: 100%;
  }

  .table-footer {
    padding: var(--spacing-md) 0;
    text-align: center;
    font-size: var(--font-size-sm);
  }

  /* Responsive */
  @media (max-width: 1024px) {
    .servers-grid {
      grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
    }
  }

  @media (max-width: 768px) {
    .view-header {
      flex-direction: column;
      align-items: stretch;
    }

    .header-actions {
      width: 100%;
      flex-direction: column;
    }

    .filter-select,
    .search-input {
      width: 100%;
    }

    .servers-grid {
      grid-template-columns: 1fr;
    }
  }
</style>
