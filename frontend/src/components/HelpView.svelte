<script lang="ts">
  let activeTab = 'quickstart'; // quickstart, shortcuts, about

  function setTab(tab: string) {
    activeTab = tab;
  }
</script>

<div class="help-view">
  <div class="view-header">
    <h2>Help & Documentation</h2>
    <p class="subtitle text-secondary">
      Learn how to use MCP Manager effectively
    </p>
  </div>

  <div class="help-tabs">
    <button
      class="tab-button"
      class:active={activeTab === 'quickstart'}
      on:click={() => setTab('quickstart')}
    >
      📖 Quickstart
    </button>
    <button
      class="tab-button"
      class:active={activeTab === 'shortcuts'}
      on:click={() => setTab('shortcuts')}
    >
      ⌨️ Keyboard Shortcuts
    </button>
    <button
      class="tab-button"
      class:active={activeTab === 'about'}
      on:click={() => setTab('about')}
    >
      ℹ️ About
    </button>
  </div>

  <div class="help-content">
    {#if activeTab === 'quickstart'}
      <div class="content-section">
        <h3>Welcome to MCP Manager</h3>
        <p>
          MCP Manager is a cross-platform application for discovering, managing, and monitoring
          Model Context Protocol (MCP) servers on your system.
        </p>

        <h4>Getting Started</h4>
        <ol>
          <li>
            <strong>Discover Servers:</strong> Click the "Refresh" button to scan your system
            for MCP servers. Servers can be discovered from Claude Desktop configuration,
            filesystem scans, or running processes.
          </li>
          <li>
            <strong>Start/Stop Servers:</strong> Use the action buttons in the server table
            to control server lifecycle. Green indicates running, red indicates stopped.
          </li>
          <li>
            <strong>View Logs:</strong> Monitor real-time logs in the bottom panel. Filter
            by server, severity, or search for specific messages.
          </li>
          <li>
            <strong>Configure Servers:</strong> Click the ⚙️ button to edit environment
            variables, command-line arguments, and behavior settings.
          </li>
        </ol>

        <h4>Server Discovery Sources</h4>
        <dl>
          <dt>Client Config</dt>
          <dd>Servers defined in Claude Desktop's configuration file</dd>

          <dt>Filesystem</dt>
          <dd>MCP servers found through filesystem scanning</dd>

          <dt>Process</dt>
          <dd>Running MCP server processes detected on your system</dd>
        </dl>

        <h4>Server Status Indicators</h4>
        <ul>
          <li><span class="status-dot running"></span> <strong>Running:</strong> Server is active and operational</li>
          <li><span class="status-dot stopped"></span> <strong>Stopped:</strong> Server is not running</li>
          <li><span class="status-dot starting"></span> <strong>Starting:</strong> Server is in the process of starting</li>
          <li><span class="status-dot error"></span> <strong>Error:</strong> Server encountered an error</li>
        </ul>

        <h4>Log Severity Levels</h4>
        <ul>
          <li><span class="log-badge info">INFO</span> General informational messages</li>
          <li><span class="log-badge success">SUCCESS</span> Successful operations</li>
          <li><span class="log-badge warning">WARNING</span> Warning messages (non-critical)</li>
          <li><span class="log-badge error">ERROR</span> Error messages (requires attention)</li>
        </ul>
      </div>
    {:else if activeTab === 'shortcuts'}
      <div class="content-section">
        <h3>Keyboard Shortcuts</h3>
        <p>Use these keyboard shortcuts to navigate and control MCP Manager more efficiently.</p>

        <h4>Global Shortcuts</h4>
        <table class="shortcuts-table">
          <tbody>
            <tr>
              <td><kbd>F5</kbd> or <kbd>Ctrl</kbd> + <kbd>R</kbd></td>
              <td>Refresh server discovery</td>
            </tr>
            <tr>
              <td><kbd>Ctrl</kbd> + <kbd>H</kbd></td>
              <td>Show keyboard shortcuts help</td>
            </tr>
            <tr>
              <td><kbd>Esc</kbd></td>
              <td>Close modals and dialogs</td>
            </tr>
          </tbody>
        </table>

        <h4>Configuration Editor</h4>
        <table class="shortcuts-table">
          <tbody>
            <tr>
              <td><kbd>Ctrl</kbd> + <kbd>S</kbd></td>
              <td>Save configuration</td>
            </tr>
            <tr>
              <td><kbd>Esc</kbd></td>
              <td>Close editor (with unsaved changes warning)</td>
            </tr>
          </tbody>
        </table>

        <h4>Detailed Logs View</h4>
        <table class="shortcuts-table">
          <tbody>
            <tr>
              <td><kbd>Ctrl</kbd> + <kbd>R</kbd></td>
              <td>Refresh logs</td>
            </tr>
            <tr>
              <td><kbd>Esc</kbd></td>
              <td>Close detailed logs view</td>
            </tr>
          </tbody>
        </table>

        <h4>Shell Launcher</h4>
        <table class="shortcuts-table">
          <tbody>
            <tr>
              <td><kbd>Ctrl</kbd> + <kbd>Enter</kbd></td>
              <td>Launch system shell</td>
            </tr>
          </tbody>
        </table>
      </div>
    {:else if activeTab === 'about'}
      <div class="content-section">
        <div class="about-header">
          <div class="app-logo">🖥️</div>
          <h3>MCP Manager</h3>
          <p class="version">Version 0.1.0</p>
        </div>

        <h4>About This Application</h4>
        <p>
          MCP Manager is a comprehensive management tool for Model Context Protocol servers.
          It provides a unified interface for discovering, monitoring, and controlling MCP
          servers across your development environment.
        </p>

        <h4>Key Features</h4>
        <ul>
          <li>Automatic server discovery from multiple sources</li>
          <li>Real-time server status monitoring</li>
          <li>Live log streaming with filtering</li>
          <li>Configuration management with validation</li>
          <li>System utilities (netstat, shell, explorer, services)</li>
          <li>Keyboard shortcuts for efficient workflow</li>
          <li>Cross-platform support (Windows, macOS, Linux)</li>
        </ul>

        <h4>Technology Stack</h4>
        <dl>
          <dt>Backend</dt>
          <dd>Go 1.21+ with Wails v2 framework</dd>

          <dt>Frontend</dt>
          <dd>Svelte 3.x with TypeScript</dd>

          <dt>Architecture</dt>
          <dd>Event-driven with real-time updates via Wails runtime events</dd>
        </dl>

        <h4>Project Links</h4>
        <div class="links-grid">
          <a href="https://modelcontextprotocol.io/introduction" target="_blank" rel="noopener" class="link-button">
            📚 MCP Documentation
          </a>
          <a href="https://wails.io" target="_blank" rel="noopener" class="link-button">
            🔧 Wails Framework
          </a>
          <a href="https://svelte.dev" target="_blank" rel="noopener" class="link-button">
            ⚡ Svelte
          </a>
          <a href="https://golang.org" target="_blank" rel="noopener" class="link-button">
            🐹 Go Language
          </a>
        </div>

        <div class="copyright">
          <p class="text-muted">
            Built with <span class="heart">❤️</span> using Wails and Svelte
          </p>
          <p class="text-muted">
            © 2025 MCP Manager. Open Source Project.
          </p>
        </div>
      </div>
    {/if}
  </div>
</div>

<style>
  .help-view {
    display: flex;
    flex-direction: column;
    height: 100%;
    background-color: var(--bg-secondary);
    border-radius: var(--radius-md);
    overflow: hidden;
  }

  .view-header {
    padding: var(--spacing-lg);
    border-bottom: 1px solid var(--border-color);
    background-color: var(--bg-tertiary);
  }

  .view-header h2 {
    margin: 0 0 var(--spacing-xs) 0;
    font-size: var(--font-size-xl);
    color: var(--text-primary);
  }

  .subtitle {
    margin: 0;
    font-size: var(--font-size-sm);
  }

  /* Tabs */
  .help-tabs {
    display: flex;
    gap: var(--spacing-xs);
    padding: var(--spacing-md) var(--spacing-lg);
    border-bottom: 1px solid var(--border-color);
    background-color: var(--bg-tertiary);
  }

  .tab-button {
    padding: var(--spacing-sm) var(--spacing-md);
    background: transparent;
    border: 1px solid transparent;
    border-radius: var(--radius-sm);
    color: var(--text-secondary);
    font-size: var(--font-size-sm);
    font-weight: 500;
    cursor: pointer;
    transition: all var(--transition-fast);
  }

  .tab-button:hover {
    background-color: var(--bg-hover);
    color: var(--text-primary);
  }

  .tab-button.active {
    background-color: var(--accent-primary);
    border-color: var(--accent-primary);
    color: white;
  }

  /* Content */
  .help-content {
    flex: 1;
    overflow: auto;
    padding: var(--spacing-xl);
  }

  .content-section {
    max-width: 800px;
    margin: 0 auto;
    text-align: left;
  }

  .content-section h3 {
    margin: 0 0 var(--spacing-md) 0;
    font-size: var(--font-size-2xl);
    color: var(--text-primary);
  }

  .content-section h4 {
    margin: var(--spacing-lg) 0 var(--spacing-md) 0;
    font-size: var(--font-size-lg);
    color: var(--text-primary);
    border-bottom: 2px solid var(--border-color);
    padding-bottom: var(--spacing-xs);
  }

  .content-section p {
    margin: 0 0 var(--spacing-md) 0;
    line-height: 1.6;
    color: var(--text-secondary);
  }

  .content-section ol,
  .content-section ul {
    margin: 0 0 var(--spacing-md) 0;
    padding-left: var(--spacing-xl);
    line-height: 1.8;
    color: var(--text-secondary);
  }

  .content-section li {
    margin-bottom: var(--spacing-sm);
  }

  .content-section li strong {
    color: var(--text-primary);
  }

  .content-section dl {
    display: grid;
    grid-template-columns: minmax(120px, 150px) 1fr;
    gap: var(--spacing-sm) var(--spacing-md);
    margin: 0 0 var(--spacing-md) 0;
    align-items: start;
  }

  .content-section dt {
    font-weight: 600;
    color: var(--text-primary);
    text-align: left;
  }

  .content-section dd {
    margin: 0;
    color: var(--text-secondary);
    text-align: left;
  }

  /* Status indicators */
  .status-dot {
    display: inline-block;
    width: 12px;
    height: 12px;
    border-radius: 50%;
    margin-right: var(--spacing-xs);
  }

  .status-dot.running {
    background-color: var(--status-running);
  }

  .status-dot.stopped {
    background-color: var(--status-stopped);
  }

  .status-dot.starting {
    background-color: var(--status-starting);
  }

  .status-dot.error {
    background-color: var(--status-error);
  }

  /* Log badges */
  .log-badge {
    display: inline-block;
    padding: 2px 8px;
    border-radius: var(--radius-sm);
    font-size: var(--font-size-xs);
    font-weight: 600;
    margin-right: var(--spacing-sm);
  }

  .log-badge.info {
    background-color: rgba(33, 150, 243, 0.2);
    color: var(--log-info);
  }

  .log-badge.success {
    background-color: rgba(76, 175, 80, 0.2);
    color: var(--log-success);
  }

  .log-badge.warning {
    background-color: rgba(255, 152, 0, 0.2);
    color: var(--log-warning);
  }

  .log-badge.error {
    background-color: rgba(244, 67, 54, 0.2);
    color: var(--log-error);
  }

  /* Shortcuts table */
  .shortcuts-table {
    width: 100%;
    margin: 0 0 var(--spacing-lg) 0;
    border-collapse: collapse;
  }

  .shortcuts-table td {
    padding: var(--spacing-sm) var(--spacing-md);
    border-bottom: 1px solid var(--border-color);
  }

  .shortcuts-table td:first-child {
    font-family: var(--font-mono);
    font-weight: 500;
    color: var(--text-primary);
    width: 250px;
  }

  .shortcuts-table td:last-child {
    color: var(--text-secondary);
  }

  kbd {
    display: inline-block;
    padding: 2px 6px;
    margin: 0 2px;
    font-size: var(--font-size-xs);
    font-family: var(--font-mono);
    background-color: var(--bg-tertiary);
    border: 1px solid var(--border-color);
    border-radius: var(--radius-sm);
    box-shadow: 0 1px 0 var(--border-color);
  }

  /* About section */
  .about-header {
    text-align: center;
    padding: var(--spacing-xl) 0;
    border-bottom: 2px solid var(--border-color);
    margin-bottom: var(--spacing-lg);
  }

  .app-logo {
    font-size: 4rem;
    margin-bottom: var(--spacing-md);
  }

  .about-header h3 {
    margin: 0 0 var(--spacing-xs) 0;
  }

  .version {
    margin: 0;
    font-size: var(--font-size-sm);
    color: var(--text-muted);
  }

  .links-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: var(--spacing-md);
    margin: var(--spacing-lg) 0;
  }

  .link-button {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: var(--spacing-md);
    background-color: var(--bg-tertiary);
    border: 1px solid var(--border-color);
    border-radius: var(--radius-md);
    color: var(--text-primary);
    text-decoration: none;
    font-weight: 500;
    transition: all var(--transition-fast);
  }

  .link-button:hover {
    background-color: var(--accent-primary);
    border-color: var(--accent-primary);
    color: white;
    transform: translateY(-2px);
    box-shadow: var(--shadow-md);
  }

  .copyright {
    text-align: center;
    margin-top: var(--spacing-xl);
    padding-top: var(--spacing-lg);
    border-top: 1px solid var(--border-color);
  }

  .copyright p {
    margin: var(--spacing-xs) 0;
    font-size: var(--font-size-sm);
  }

  .heart {
    color: #e74c3c;
    animation: heartbeat 1.5s ease-in-out infinite;
  }

  @keyframes heartbeat {
    0%, 100% {
      transform: scale(1);
    }
    50% {
      transform: scale(1.1);
    }
  }

  /* Responsive */
  @media (max-width: 768px) {
    .help-tabs {
      flex-direction: column;
    }

    .tab-button {
      width: 100%;
    }

    .help-content {
      padding: var(--spacing-md);
    }

    .content-section dl {
      grid-template-columns: 1fr;
    }

    .shortcuts-table td:first-child {
      width: auto;
    }

    .links-grid {
      grid-template-columns: 1fr;
    }
  }
</style>
