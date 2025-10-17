<script lang="ts">
  // Placeholder for sidebar navigation
  import { selectedServerId, serverFilters } from '../stores/stores';

  let activeView = 'servers';

  function setView(view: string) {
    activeView = view;
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

  .sidebar-footer {
    padding: var(--spacing-md);
    border-top: 1px solid var(--border-color);
    text-align: center;
  }
</style>
