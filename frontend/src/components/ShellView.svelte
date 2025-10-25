<script lang="ts">
  import { addNotification } from '../stores/stores';

  // NOTE: This component requires backend API endpoint:
  // POST /api/v1/shell
  // Response: { success: boolean, message: string }
  // Backend should launch platform shell: cmd.exe (Windows), Terminal.app (macOS), xterm (Linux)

  let launching = false;

  async function launchShell() {
    launching = true;

    try {
      // TODO: Replace with actual API call when backend endpoint is ready
      // const response = await fetch('/api/v1/shell', { method: 'POST' });
      // if (!response.ok) throw new Error('Failed to launch shell');
      // const data = await response.json();
      // addNotification('success', data.message || 'Shell launched successfully');

      // Mock implementation for now
      await new Promise(resolve => setTimeout(resolve, 500));
      addNotification('warning', 'Backend API /api/v1/shell not implemented yet');
      addNotification('info', 'Would launch: cmd.exe (Windows), Terminal.app (macOS), or xterm (Linux)');
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

  <!-- Backend API notice -->
  <div class="api-notice">
    <strong>⚠️ Backend API Required:</strong>
    <code>POST /api/v1/shell</code>
    <br />
    Response: <code>{'{ success: boolean, message: string }'}</code>
    <br />
    <small>Should launch platform shell using os/exec: cmd.exe, Terminal.app, or xterm</small>
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
    overflow: auto;
    padding: var(--spacing-xl);
    display: flex;
    align-items: center;
    justify-content: center;
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

  .api-notice small {
    display: block;
    margin-top: var(--spacing-xs);
    color: var(--text-muted);
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
