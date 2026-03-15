<script lang="ts">
  import './app.css';
  import ServerTable from './components/ServerTable.svelte';
  import LogViewer from './components/LogViewer.svelte';
  import Sidebar from './components/Sidebar.svelte';
  import NetstatView from './components/NetstatView.svelte';
  import ShellView from './components/ShellView.svelte';
  import ExplorerView from './components/ExplorerView.svelte';
  import HelpView from './components/HelpView.svelte';
  import { isConnected, isDiscovering, notifications, activeView, selectedServerId, servers, applicationState, hasNetworkTransportServers, addNotification, requestConfigEditor } from './stores/stores';
  import { get } from 'svelte/store';
  import { api } from './services/api';
  import { setupWailsEvents, cleanupWailsEvents } from './services/events';
  import { EventsOn } from '../wailsjs/runtime/runtime';
  import { onMount, onDestroy } from 'svelte';

  let logPanelHeight = 200; // Default 200px per FR-020
  let logPanelVisible = true;
  let savedLogPanelHeight = 200;
  let isResizing = false;
  let startY = 0;
  let startHeight = 0;
  let saveStateTimeout: number | null = null;

  // Setup Wails event listeners on mount
  onMount(() => {
    setupWailsEvents();
    // Mark as connected (Wails IPC is always available when running)
    isConnected.set(true);
    refreshDiscovery();

    // Register menu:action event from native menu bar
    EventsOn('menu:action', (action: string) => {
      handleMenuAction(action);
    });

    // FR-046: Setup keyboard shortcuts (simplified — menu accelerators cover most keys)
    window.addEventListener('keydown', handleKeyboardShortcut);

    // T-E022: Load persisted window state
    loadWindowState();

    // T-E022: Save window state on resize
    window.addEventListener('resize', handleWindowResize);
  });

  // Cleanup Wails event listeners on unmount
  onDestroy(() => {
    cleanupWailsEvents();
    // FR-046: Cleanup keyboard shortcuts
    window.removeEventListener('keydown', handleKeyboardShortcut);
    // T-E022: Cleanup window resize listener
    window.removeEventListener('resize', handleWindowResize);
    // Save any pending state changes
    if (saveStateTimeout) {
      clearTimeout(saveStateTimeout);
      saveWindowState();
    }
  });

  // Handle native menu bar actions forwarded via "menu:action" event
  function handleMenuAction(action: string) {
    switch (action) {
      case 'refresh-discovery':
        refreshDiscovery();
        break;
      case 'view-servers':
        activeView.set('servers');
        break;
      case 'view-netstat':
        if (!get(hasNetworkTransportServers)) {
          addNotification('info', 'Network monitoring applies to HTTP/SSE transport servers. All current servers use stdio transport.');
          return;
        }
        activeView.set('netstat');
        break;
      case 'view-shell':
        activeView.set('shell');
        break;
      case 'view-explorer':
        activeView.set('explorer');
        break;
      case 'view-help':
        activeView.set('help');
        break;
      case 'toggle-log-panel':
        toggleLogPanel();
        break;
      case 'server-start':
      case 'server-stop':
      case 'server-restart':
      case 'server-view-config':
      case 'server-open-directory':
        handleServerMenuAction(action);
        break;
    }
  }

  // Toggle log panel visibility, restoring previous height on re-open
  function toggleLogPanel() {
    if (logPanelVisible) {
      savedLogPanelHeight = logPanelHeight;
      logPanelVisible = false;
    } else {
      logPanelHeight = savedLogPanelHeight;
      logPanelVisible = true;
    }
  }

  // Handle server-scoped menu actions; requires a server to be selected
  function handleServerMenuAction(action: string) {
    const currentServerId = get(selectedServerId);
    if (!currentServerId) {
      addNotification('warning', 'No server selected. Select a server from the table first.');
      return;
    }

    const selectedSrv = get(servers).find(s => s.id === currentServerId);
    if (!selectedSrv) return;

    switch (action) {
      case 'server-start':
        api.lifecycle.startServer(currentServerId)
          .then(() => addNotification('success', `Starting ${selectedSrv.name}`))
          .catch((err: any) => addNotification('error', `Failed to start: ${err}`));
        break;
      case 'server-stop':
        api.lifecycle.stopServer(currentServerId, { force: false, timeout: 30 })
          .then(() => addNotification('success', `Stopping ${selectedSrv.name}`))
          .catch((err: any) => addNotification('error', `Failed to stop: ${err}`));
        break;
      case 'server-restart':
        api.lifecycle.restartServer(currentServerId)
          .then(() => addNotification('success', `Restarting ${selectedSrv.name}`))
          .catch((err: any) => addNotification('error', `Failed to restart: ${err}`));
        break;
      case 'server-view-config':
        // Switch to servers view so ServerTable is mounted, then request editor
        activeView.set('servers');
        requestConfigEditor.set(currentServerId);
        break;
      case 'server-open-directory':
        api.utilities.openExplorer(currentServerId)
          .then(() => addNotification('success', `Opened directory for ${selectedSrv.name}`))
          .catch((err: any) => addNotification('error', `Failed to open directory: ${err}`));
        break;
    }
  }

  // FR-046: Keyboard shortcuts — menu accelerators handle most keys;
  // Ctrl+R is kept as a secondary refresh shortcut (not in menu, no conflict).
  function handleKeyboardShortcut(event: KeyboardEvent) {
    if (event.ctrlKey && event.key === 'r') {
      event.preventDefault();
      refreshDiscovery();
    }
  }

  // Refresh discovery - triggers server scan
  async function refreshDiscovery() {
    try {
      $isDiscovering = true;
      await api.discovery.discoverServers();
    } catch (error) {
      console.error('Discovery failed:', error);
    } finally {
      $isDiscovering = false;
    }
  }

  // Resizable log panel handlers
  function startResize(event: MouseEvent) {
    isResizing = true;
    startY = event.clientY;
    startHeight = logPanelHeight;
    document.addEventListener('mousemove', handleResize);
    document.addEventListener('mouseup', stopResize);
  }

  function handleResize(event: MouseEvent) {
    if (!isResizing) return;
    const deltaY = startY - event.clientY; // Inverted because we're resizing from bottom
    logPanelHeight = Math.max(100, Math.min(600, startHeight + deltaY)); // Min 100px, max 600px
  }

  function stopResize() {
    isResizing = false;
    document.removeEventListener('mousemove', handleResize);
    document.removeEventListener('mouseup', stopResize);

    // T-E022: Save log panel height when done resizing
    saveLogPanelHeight();
  }

  // T-E022: Load window state from backend
  async function loadWindowState() {
    try {
      const state = await api.appState.getApplicationState();
      if (state && state.windowLayout) {
        logPanelHeight = state.windowLayout.logPanelHeight || 200;
      }
    } catch (error) {
      console.error('Failed to load window state:', error);
      // Use defaults if load fails
    }
  }

  // T-E022: Handle window resize (debounced)
  function handleWindowResize() {
    if (saveStateTimeout) {
      clearTimeout(saveStateTimeout);
    }
    saveStateTimeout = window.setTimeout(() => {
      saveWindowState();
    }, 1000); // Debounce 1 second
  }

  // T-E022: Save window state to backend
  async function saveWindowState() {
    try {
      const updatedState = {
        ...$applicationState,
        windowLayout: {
          ...$applicationState.windowLayout,
          width: window.innerWidth,
          height: window.innerHeight,
          x: window.screenX,
          y: window.screenY,
          maximized: window.outerWidth === screen.availWidth && window.outerHeight === screen.availHeight,
          logPanelHeight: logPanelHeight
        }
      };
      applicationState.set(updatedState);
      await api.appState.updateApplicationState(updatedState);
    } catch (error) {
      console.error('Failed to save window state:', error);
    }
  }

  // T-E022: Save log panel height
  function saveLogPanelHeight() {
    if (saveStateTimeout) {
      clearTimeout(saveStateTimeout);
    }
    saveStateTimeout = window.setTimeout(() => {
      saveWindowState();
    }, 500); // Debounce 500ms for log panel
  }

  // Remove notification
  function dismissNotification(id: string) {
    notifications.update(n => n.filter(notif => notif.id !== id));
  }
</script>

<div class="app-container">
  <!-- Header -->
  <header class="app-header">
    <div class="header-content">
      <h1>MCP Manager</h1>
      <div class="header-actions">
        <div class="connection-status">
          {#if $isConnected}
            <span class="status-indicator connected" title="Connected to backend"></span>
            <span class="text-secondary">Connected</span>
          {:else}
            <span class="status-indicator disconnected" title="Disconnected from backend"></span>
            <span class="text-secondary">Disconnected</span>
          {/if}
        </div>
        <button
          class="primary"
          on:click={refreshDiscovery}
          disabled={$isDiscovering}
        >
          {$isDiscovering ? 'Discovering...' : '🔄 Refresh'}
        </button>
      </div>
    </div>
  </header>

  <!-- Main content area with sidebar -->
  <div class="main-content">
    <Sidebar />
    <div class="content-area">
      {#if $activeView === 'servers'}
        <ServerTable />
      {:else if $activeView === 'netstat'}
        <NetstatView />
      {:else if $activeView === 'shell'}
        <ShellView />
      {:else if $activeView === 'explorer'}
        <ExplorerView />
      {:else if $activeView === 'help'}
        <HelpView />
      {:else}
        <ServerTable />
      {/if}
    </div>
  </div>

  <!-- Resizable log viewer -->
  {#if logPanelVisible}
  <div class="log-panel-container" style="height: {logPanelHeight}px;">
    <div
      class="resize-handle"
      on:mousedown={startResize}
      role="separator"
      aria-label="Resize log panel"
      tabindex="0"
    ></div>
    <LogViewer />
  </div>
  {/if}

  <!-- Notification toast container -->
  {#if $notifications.length > 0}
    <div class="notifications-container">
      {#each $notifications as notification (notification.id)}
        <div class="notification {notification.type}">
          <span class="notification-message">{notification.message}</span>
          <button
            class="notification-close"
            on:click={() => dismissNotification(notification.id)}
            aria-label="Dismiss notification"
          >
            ×
          </button>
        </div>
      {/each}
    </div>
  {/if}
</div>

<style>
  /* App container - CSS Grid layout (FR-045) */
  .app-container {
    display: grid;
    grid-template-rows: auto 1fr auto;
    height: 100vh;
    width: 100vw;
    overflow: hidden;
    background-color: var(--bg-primary);
  }

  /* Header */
  .app-header {
    background-color: var(--bg-secondary);
    border-bottom: 1px solid var(--border-color);
    box-shadow: var(--shadow-sm);
    z-index: 10;
  }

  .header-content {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: var(--spacing-md) var(--spacing-lg);
    max-width: 100%;
  }

  .app-header h1 {
    margin: 0;
    font-size: var(--font-size-xl);
    color: var(--text-primary);
  }

  .header-actions {
    display: flex;
    align-items: center;
    gap: var(--spacing-md);
  }

  .connection-status {
    display: flex;
    align-items: center;
    gap: var(--spacing-xs);
    padding: var(--spacing-xs) var(--spacing-sm);
    border-radius: var(--radius-sm);
    background-color: var(--bg-tertiary);
  }

  .status-indicator {
    display: inline-block;
    width: 8px;
    height: 8px;
    border-radius: 50%;
    animation: pulse 2s ease-in-out infinite;
  }

  .status-indicator.connected {
    background-color: var(--status-running);
  }

  .status-indicator.disconnected {
    background-color: var(--status-error);
  }

  @keyframes pulse {
    0%, 100% {
      opacity: 1;
    }
    50% {
      opacity: 0.5;
    }
  }

  /* Main content - CSS Grid with sidebar */
  .main-content {
    display: grid;
    grid-template-columns: 200px 1fr;
    overflow: hidden;
    height: 100%;
  }

  .content-area {
    overflow-y: auto;
    padding: var(--spacing-lg);
    background-color: var(--bg-primary);
  }

  /* Resizable log panel (FR-020) */
  .log-panel-container {
    position: relative;
    min-height: 100px;
    max-height: 600px;
    overflow: hidden;
  }

  .resize-handle {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    height: 4px;
    background-color: var(--border-color);
    cursor: ns-resize;
    z-index: 20;
    transition: background-color var(--transition-fast);
  }

  .resize-handle:hover,
  .resize-handle:focus {
    background-color: var(--accent-primary);
    outline: none;
  }

  /* Notifications container */
  .notifications-container {
    position: fixed;
    top: var(--spacing-lg);
    right: var(--spacing-lg);
    z-index: 1000;
    display: flex;
    flex-direction: column;
    gap: var(--spacing-sm);
    max-width: 400px;
    pointer-events: none;
  }

  .notification {
    pointer-events: auto;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--spacing-md);
    min-width: 300px;
  }

  .notification-message {
    flex: 1;
    font-size: var(--font-size-sm);
  }

  .notification-close {
    background: transparent;
    border: none;
    color: var(--text-primary);
    font-size: var(--font-size-xl);
    line-height: 1;
    padding: 0;
    cursor: pointer;
    opacity: 0.7;
    transition: opacity var(--transition-fast);
  }

  .notification-close:hover {
    opacity: 1;
  }

  /* Responsive layout (FR-045) */
  @media (max-width: 768px) {
    .main-content {
      grid-template-columns: 60px 1fr;
    }

    .header-content h1 {
      font-size: var(--font-size-lg);
    }

    .connection-status span:last-child {
      display: none;
    }
  }

  @media (max-width: 480px) {
    .main-content {
      grid-template-columns: 1fr;
    }

    .log-panel-container {
      height: 150px !important;
    }

    .notifications-container {
      left: var(--spacing-sm);
      right: var(--spacing-sm);
      max-width: none;
    }
  }
</style>
