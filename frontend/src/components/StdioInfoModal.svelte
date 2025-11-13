<script lang="ts">
  import type { MCPServer } from '../stores/stores';

  // Props
  export let server: MCPServer;
  export let onClose: () => void;
  export let onOpenConfigEditor: () => void;

  // Handle keyboard shortcuts
  function handleKeydown(event: KeyboardEvent) {
    if (event.key === 'Escape') {
      onClose();
    }
  }

  function handleConfigEditor() {
    onClose();
    onOpenConfigEditor();
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<!-- Modal backdrop -->
<div class="modal-backdrop" on:click={onClose} role="presentation">
  <!-- Modal content -->
  <div class="modal-content" on:click|stopPropagation role="dialog" aria-modal="true" aria-labelledby="modal-title">
    <!-- Header -->
    <div class="modal-header">
      <h2 id="modal-title" class="modal-title">
        Stdio Server: {server.name}
      </h2>
      <button class="btn-close" on:click={onClose} aria-label="Close">&times;</button>
    </div>

    <!-- Body -->
    <div class="modal-body">
      <div class="info-section">
        <div class="info-icon">ℹ️</div>
        <div class="info-content">
          <h3>This server requires an MCP client</h3>
          <p>
            <strong>{server.name}</strong> uses <strong>stdio transport</strong>, which means it can only
            run when connected to an MCP client like Claude Desktop or Cursor.
          </p>

          <h4>How to start this server:</h4>
          <ol>
            <li>Add it to your MCP client's configuration file</li>
            <li>Restart your MCP client</li>
            <li>The client will launch and connect to the server automatically</li>
          </ol>

          <div class="note">
            <strong>Tip:</strong> Use the "Configure Client" button below to add this server
            to your client's configuration automatically.
          </div>
        </div>
      </div>
    </div>

    <!-- Footer -->
    <div class="modal-footer">
      <button class="btn btn-secondary" on:click={onClose}>
        Got it
      </button>
      <button class="btn btn-primary" on:click={handleConfigEditor}>
        Configure Client
      </button>
    </div>
  </div>
</div>

<style>
  .modal-backdrop {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: rgba(0, 0, 0, 0.6);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
    padding: var(--spacing-lg);
  }

  .modal-content {
    background-color: var(--bg-primary);
    border-radius: var(--radius-lg);
    box-shadow: 0 10px 40px rgba(0, 0, 0, 0.3);
    max-width: 600px;
    width: 100%;
    max-height: 90vh;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .modal-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: var(--spacing-lg);
    border-bottom: 1px solid var(--border-color);
  }

  .modal-title {
    margin: 0;
    font-size: var(--font-size-lg);
    color: var(--text-primary);
  }

  .btn-close {
    background: none;
    border: none;
    font-size: 1.5rem;
    color: var(--text-secondary);
    cursor: pointer;
    padding: 0;
    width: 2rem;
    height: 2rem;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: var(--radius-sm);
    transition: background-color var(--transition-fast);
  }

  .btn-close:hover {
    background-color: var(--bg-hover);
  }

  .modal-body {
    flex: 1;
    overflow-y: auto;
    padding: var(--spacing-lg);
  }

  .info-section {
    display: flex;
    gap: var(--spacing-md);
  }

  .info-icon {
    font-size: 2.5rem;
    flex-shrink: 0;
  }

  .info-content {
    flex: 1;
  }

  .info-content h3 {
    margin: 0 0 var(--spacing-md) 0;
    color: var(--text-primary);
    font-size: var(--font-size-md);
  }

  .info-content h4 {
    margin: var(--spacing-lg) 0 var(--spacing-sm) 0;
    color: var(--text-primary);
    font-size: var(--font-size-sm);
  }

  .info-content p {
    margin: 0 0 var(--spacing-md) 0;
    color: var(--text-secondary);
    line-height: 1.6;
  }

  .info-content ol {
    margin: var(--spacing-sm) 0;
    padding-left: var(--spacing-lg);
    color: var(--text-secondary);
    line-height: 1.8;
  }

  .info-content ol li {
    margin-bottom: var(--spacing-xs);
  }

  .note {
    margin-top: var(--spacing-lg);
    padding: var(--spacing-md);
    background-color: rgba(33, 150, 243, 0.1);
    border-left: 3px solid var(--accent-primary);
    border-radius: var(--radius-sm);
    color: var(--text-secondary);
    font-size: var(--font-size-sm);
  }

  .modal-footer {
    display: flex;
    justify-content: flex-end;
    gap: var(--spacing-sm);
    padding: var(--spacing-lg);
    border-top: 1px solid var(--border-color);
  }

  .btn, .btn-secondary, .btn-primary {
    padding: var(--spacing-sm) var(--spacing-lg);
    border-radius: var(--radius-md);
    font-size: var(--font-size-sm);
    font-weight: 500;
    cursor: pointer;
    transition: all var(--transition-fast);
  }

  .btn-secondary {
    background-color: var(--button-bg);
    border: 1px solid var(--border-color);
    color: var(--text-primary);
  }

  .btn-secondary:hover {
    background-color: var(--button-hover);
  }

  .btn-primary {
    background-color: var(--accent-primary);
    color: white;
    border: 1px solid var(--accent-primary);
  }

  .btn-primary:hover {
    background-color: var(--accent-hover);
    border-color: var(--accent-hover);
  }
</style>
