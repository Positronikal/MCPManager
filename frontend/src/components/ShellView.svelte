<script lang="ts">
  import { addNotification } from '../stores/stores';
  import { LaunchShell } from '../../wailsjs/go/main/App';

  let launching = false;

  async function launchShell() {
    launching = true;

    try {
      const response = await LaunchShell();
      if (response.Success) {
        addNotification('success', response.Message || 'Shell launched successfully');
      } else {
        addNotification('error', response.Message || 'Failed to launch shell');
      }
    } catch (err: any) {
      addNotification('error', err.message || 'Failed to launch shell');
    } finally {
      launching = false;
    }
  }

  // Keyboard shortcuts
  function handleKeydown(event: KeyboardEvent) {
    if (event.ctrlKey && event.key === 'Enter') {
      event.preventDefault();
      launchShell();
    }
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<div class="shell-view">
  <div class="view-header">
    <div class="header-content">
      <h2>Quick Shell Access</h2>
      <p class="subtitle text-secondary">
        Launch a platform-appropriate terminal for command-line access
      </p>
    </div>
  </div>

  <div class="view-content">
    <div class="shell-card">
      <div class="shell-icon">💻</div>

      <div class="shell-info">
        <h3>Platform Terminal</h3>
        <p class="text-secondary">
          Launches your system's default terminal application:
        </p>
        <ul class="platform-list">
          <li><strong>Windows:</strong> Command Prompt (cmd.exe) or PowerShell</li>
          <li><strong>macOS:</strong> Terminal.app</li>
          <li><strong>Linux:</strong> xterm or default terminal emulator</li>
        </ul>
      </div>

      <div class="shell-actions">
        <button
          class="btn-primary btn-lg"
          on:click={launchShell}
          disabled={launching}
        >
          {launching ? 'Launching...' : '🚀 Open Shell'}
        </button>
        <p class="hint text-muted">
          Keyboard shortcut: <kbd>Ctrl</kbd> + <kbd>Enter</kbd>
        </p>
      </div>

      <div class="shell-tips">
        <h4>Quick Tips:</h4>
        <ul>
          <li>Use shell to manually test MCP server commands</li>
          <li>Navigate to server installation directories for debugging</li>
          <li>Run system diagnostics or monitoring commands</li>
          <li>Execute server-specific CLI tools</li>
        </ul>
      </div>
    </div>
  </div>

</div>

<style>
  .shell-view {
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

  .header-content h2 {
    margin: 0 0 var(--spacing-xs) 0;
    font-size: var(--font-size-xl);
    color: var(--text-primary);
  }

  .subtitle {
    margin: 0;
    font-size: var(--font-size-sm);
  }

  .view-content {
    flex: 1;
    overflow-y: auto;
    overflow-x: hidden;
    padding: var(--spacing-xl);
    display: flex;
    align-items: flex-start;
    justify-content: center;
    min-height: 0;
  }

  .shell-card {
    max-width: 600px;
    width: 100%;
    background-color: var(--bg-primary);
    border: 1px solid var(--border-color);
    border-radius: var(--radius-lg);
    padding: var(--spacing-xl);
    display: flex;
    flex-direction: column;
    gap: var(--spacing-lg);
    box-shadow: var(--shadow-md);
  }

  .shell-icon {
    font-size: 4rem;
    text-align: center;
  }

  .shell-info h3 {
    margin: 0 0 var(--spacing-sm) 0;
    font-size: var(--font-size-lg);
    color: var(--text-primary);
  }

  .shell-info p {
    margin: 0 0 var(--spacing-md) 0;
  }

  .platform-list {
    list-style: none;
    padding: 0;
    margin: 0;
    display: flex;
    flex-direction: column;
    gap: var(--spacing-sm);
  }

  .platform-list li {
    padding: var(--spacing-sm) var(--spacing-md);
    background-color: var(--bg-secondary);
    border-radius: var(--radius-sm);
    border-left: 3px solid var(--accent-primary);
    font-size: var(--font-size-sm);
    color: var(--text-secondary);
  }

  .platform-list strong {
    color: var(--text-primary);
  }

  .shell-actions {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--spacing-sm);
    padding: var(--spacing-lg) 0;
  }

  .btn-lg {
    padding: var(--spacing-md) var(--spacing-xl);
    font-size: var(--font-size-md);
    min-width: 200px;
  }

  .hint {
    font-size: var(--font-size-xs);
    text-align: center;
  }

  kbd {
    display: inline-block;
    padding: 2px 6px;
    font-size: var(--font-size-xs);
    font-family: var(--font-mono);
    background-color: var(--bg-tertiary);
    border: 1px solid var(--border-color);
    border-radius: var(--radius-sm);
    box-shadow: 0 1px 0 var(--border-color);
  }

  .shell-tips {
    padding: var(--spacing-md);
    background-color: var(--bg-secondary);
    border-radius: var(--radius-md);
    border: 1px solid var(--border-color);
  }

  .shell-tips h4 {
    margin: 0 0 var(--spacing-sm) 0;
    font-size: var(--font-size-sm);
    color: var(--text-primary);
  }

  .shell-tips ul {
    margin: 0;
    padding-left: var(--spacing-lg);
    display: flex;
    flex-direction: column;
    gap: var(--spacing-xs);
  }

  .shell-tips li {
    font-size: var(--font-size-sm);
    color: var(--text-secondary);
  }

  /* Responsive */
  @media (max-width: 768px) {
    .view-content {
      padding: var(--spacing-md);
    }

    .shell-card {
      padding: var(--spacing-md);
    }

    .platform-list {
      font-size: var(--font-size-xs);
    }
  }
</style>
