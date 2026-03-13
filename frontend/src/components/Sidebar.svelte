<script lang="ts">
  import { onMount } from 'svelte';
  import { selectedServerId, servers, addNotification, activeView, hasNetworkTransportServers } from '../stores/stores';
  import { BrowserOpenURL } from '../../wailsjs/runtime/runtime';
  import { GetVersion } from '../../wailsjs/go/main/App';

  let appVersion = '';

  onMount(async () => {
    try {
      appVersion = await GetVersion();
    } catch (err) {
      appVersion = 'unknown';
    }
  });

  function setView(view: string) {
    activeView.set(view);
  }

  // Open server installation directory in system file explorer
  async function openExplorer() {
    const selectedServer = $servers.find(s => s.id === $selectedServerId);
    if (selectedServer) {
      try {
        // Get the directory containing the server executable
        const installPath = selectedServer.installationPath;
        const directory = installPath.substring(0, installPath.lastIndexOf(/[/\\]/));
        await BrowserOpenURL(directory);
        addNotification('success', `Opened ${selectedServer.name} directory`);
      } catch (error) {
        addNotification('error', `Failed to open directory: ${error}`);
      }
    } else {
      addNotification('warning', 'Please select a server first');
    }
  }

  // Open system shell/terminal
  function openShell() {
    addNotification('info', 'Shell utility: Use your system terminal');
    // Note: Opening a system terminal varies by OS and requires platform-specific commands
    // This would be better implemented as a backend API method
  }

  function showNetstat() {
    if (!$hasNetworkTransportServers) {
      addNotification('info', 'Network monitoring applies to HTTP/SSE transport servers. All current servers use stdio transport.');
      return;
    }
    setView('netstat');
  }

  function showShell() {
    setView('shell');
  }

  function showExplorer() {
    setView('explorer');
  }

  function showHelp() {
    setView('help');
  }
</script>

<nav class="sidebar">
  <div class="sidebar-header">
    <h2>MCP Manager</h2>
  </div>

  <ul class="nav-menu">
    <li class:active={$activeView === 'servers'}>
      <button on:click={() => setView('servers')}>
        <span class="nav-icon">🖥️</span>
        <span class="nav-label">Servers</span>
      </button>
    </li>
  </ul>

  <!-- Utility functions -->
  <div class="utilities-section">
    <div class="section-label">Utilities</div>
    <ul class="nav-menu">
      <li class:active={$activeView === 'netstat'} class:disabled={!$hasNetworkTransportServers}>
        <button
          on:click={showNetstat}
          title={$hasNetworkTransportServers
            ? 'View HTTP/SSE network connections'
            : 'Network monitoring unavailable — no HTTP/SSE servers discovered'}
        >
          <span class="nav-icon">🌐</span>
          <span class="nav-label">Netstat</span>
        </button>
      </li>
      <li class:active={$activeView === 'shell'}>
        <button on:click={showShell} title="Shell launcher">
          <span class="nav-icon">💻</span>
          <span class="nav-label">Shell</span>
        </button>
      </li>
      <li class:active={$activeView === 'explorer'}>
        <button on:click={showExplorer} title="Open server directories">
          <span class="nav-icon">📁</span>
          <span class="nav-label">Explorer</span>
        </button>
      </li>
      <li class:active={$activeView === 'help'}>
        <button on:click={showHelp} title="Help and documentation">
          <span class="nav-icon">❓</span>
          <span class="nav-label">Help</span>
        </button>
      </li>
    </ul>
  </div>

  <div class="sidebar-footer">
    <p class="text-muted" style="font-size: var(--font-size-xs);">{appVersion || '...'}</p>
  </div>
</nav>

<style>
  .sidebar {
    display: flex;
    flex-direction: column;
    height: 100%;
    background-color: var(--bg-secondary);
    border-right: 1px solid var(--border-color);
    overflow-y: auto;
  }

  .sidebar-header {
    padding: var(--spacing-lg);
    border-bottom: 1px solid var(--border-color);
  }

  .sidebar-header h2 {
    margin: 0;
    font-size: var(--font-size-lg);
    color: var(--text-primary);
  }

  .nav-menu {
    flex: 0 0 auto;
    list-style: none;
    padding: var(--spacing-md) 0;
    margin: 0;
  }

  .nav-menu li {
    margin: 0;
  }

  .nav-menu button {
    display: flex;
    align-items: center;
    gap: var(--spacing-sm);
    width: 100%;
    padding: var(--spacing-md) var(--spacing-lg);
    background: transparent;
    border: none;
    border-left: 3px solid transparent;
    color: var(--text-secondary);
    text-align: left;
    cursor: pointer;
    transition: all var(--transition-fast);
  }

  .nav-menu button:hover {
    background-color: var(--bg-hover);
    color: var(--text-primary);
  }

  .nav-menu li.active button {
    background-color: var(--bg-tertiary);
    border-left-color: var(--accent-primary);
    color: var(--text-primary);
  }

  .nav-menu li.disabled button {
    opacity: 0.4;
    cursor: not-allowed;
  }

  .nav-menu li.disabled button:hover {
    background-color: transparent;
    color: var(--text-secondary);
  }

  .nav-icon {
    font-size: var(--font-size-lg);
  }

  .nav-label {
    font-size: var(--font-size-sm);
    font-weight: 500;
  }

  .utilities-section {
    flex: 0 0 auto;
    margin-top: auto;
    padding-top: var(--spacing-md);
    border-top: 1px solid var(--border-color);
  }

  .section-label {
    padding: var(--spacing-xs) var(--spacing-lg);
    font-size: var(--font-size-xs);
    text-transform: uppercase;
    color: var(--text-muted);
    font-weight: 600;
    letter-spacing: 0.5px;
  }

  .sidebar-footer {
    flex: 0 0 auto;
    padding: var(--spacing-md);
    border-top: 1px solid var(--border-color);
    text-align: center;
  }
</style>
