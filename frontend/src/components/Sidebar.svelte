<script lang="ts">
  import { selectedServerId, serverFilters, servers, addNotification } from '../stores/stores';
  import { BrowserOpenURL } from '../../wailsjs/runtime/runtime';

  let activeView = 'servers';

  function setView(view: string) {
    activeView = view;
  }

  // FR-032: Open server installation directory in system file explorer
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

  // FR-031: Open system shell/terminal
  function openShell() {
    addNotification('info', 'Shell utility: Use your system terminal');
    // Note: Opening a system terminal varies by OS and requires platform-specific commands
    // This would be better implemented as a backend API method
  }

  // FR-030: Show network connections (would open a netstat view)
  function showNetstat() {
    addNotification('info', 'Netstat view coming soon');
    // This would show a modal with network connections for MCP servers
  }

  // FR-033: Show system services
  function showServices() {
    addNotification('info', 'Services view coming soon');
    // This would show system services related to MCP servers
  }

  // FR-034: Show help and documentation
  function showHelp() {
    BrowserOpenURL('https://github.com/hoytech/mcpmanager/blob/main/README.md');
  }
</script>

<nav class="sidebar">
  <div class="sidebar-header">
    <h2>MCP Manager</h2>
  </div>

  <ul class="nav-menu">
    <li class:active={activeView === 'servers'}>
      <button on:click={() => setView('servers')}>
        <span class="nav-icon">🖥️</span>
        <span class="nav-label">Servers</span>
      </button>
    </li>
    <li class:active={activeView === 'logs'}>
      <button on:click={() => setView('logs')}>
        <span class="nav-icon">📋</span>
        <span class="nav-label">Logs</span>
      </button>
    </li>
    <li class:active={activeView === 'settings'}>
      <button on:click={() => setView('settings')}>
        <span class="nav-icon">⚙️</span>
        <span class="nav-label">Settings</span>
      </button>
    </li>
  </ul>

  <!-- FR-030-034: Utility functions -->
  <div class="utilities-section">
    <div class="section-label">Utilities</div>
    <ul class="nav-menu">
      <li>
        <button on:click={showNetstat} title="FR-030: View network connections">
          <span class="nav-icon">🌐</span>
          <span class="nav-label">Netstat</span>
        </button>
      </li>
      <li>
        <button on:click={openShell} title="FR-031: Open terminal">
          <span class="nav-icon">💻</span>
          <span class="nav-label">Shell</span>
        </button>
      </li>
      <li>
        <button on:click={openExplorer} title="FR-032: Open server directory">
          <span class="nav-icon">📁</span>
          <span class="nav-label">Explorer</span>
        </button>
      </li>
      <li>
        <button on:click={showServices} title="FR-033: View system services">
          <span class="nav-icon">🔧</span>
          <span class="nav-label">Services</span>
        </button>
      </li>
      <li>
        <button on:click={showHelp} title="FR-034: Help and documentation">
          <span class="nav-icon">❓</span>
          <span class="nav-label">Help</span>
        </button>
      </li>
    </ul>
  </div>

  <div class="sidebar-footer">
    <p class="text-muted" style="font-size: var(--font-size-xs);">v0.1.0</p>
  </div>
</nav>

<style>
  .sidebar {
    display: flex;
    flex-direction: column;
    height: 100%;
    background-color: var(--bg-secondary);
    border-right: 1px solid var(--border-color);
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
    flex: 1;
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

  .nav-icon {
    font-size: var(--font-size-lg);
  }

  .nav-label {
    font-size: var(--font-size-sm);
    font-weight: 500;
  }

  .utilities-section {
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
    padding: var(--spacing-md);
    border-top: 1px solid var(--border-color);
    text-align: center;
  }
</style>
