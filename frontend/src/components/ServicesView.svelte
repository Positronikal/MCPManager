<script lang="ts">
  import { onMount } from 'svelte';
  import { addNotification } from '../stores/stores';

  // NOTE: This component requires backend API endpoint:
  // GET /api/v1/services
  // Response: { services: Service[] }
  // Service: { name: string, status: string, description: string, pid?: number }
  // Backend should run: sc query (Windows), launchctl list (macOS), systemctl list-units (Linux)

  interface Service {
    name: string;
    status: string;
    description: string;
    pid?: number;
  }

  let services: Service[] = [];
  let filteredServices: Service[] = [];
  let loading = false;
  let error = '';
  let searchQuery = '';
  let searchInput = ''; // Temporary input value for debouncing
  let searchTimeout: number | null = null;
  let statusFilter: string | null = null;

  // Mock data for demonstration
  const mockServices: Service[] = [
    {
      name: 'mcp-server-example',
      status: 'running',
      description: 'Example MCP Server Service',
      pid: 12345
    },
    {
      name: 'docker',
      status: 'running',
      description: 'Docker Desktop Service',
      pid: 4567
    },
    {
      name: 'ssh-agent',
      status: 'stopped',
      description: 'OpenSSH Authentication Agent'
    }
  ];

  onMount(() => {
    loadServices();
  });

  async function loadServices() {
    loading = true;
    error = '';

    try {
      // TODO: Replace with actual API call when backend endpoint is ready
      // const response = await fetch('/api/v1/services');
      // if (!response.ok) throw new Error('Failed to fetch services');
      // const data = await response.json();
      // services = data.services;

      // For now, use mock data
      await new Promise(resolve => setTimeout(resolve, 500));
      services = mockServices;
      applyFilters();

      addNotification('warning', 'Backend API /api/v1/services not implemented yet - showing mock data');
    } catch (err: any) {
      error = err.message || 'Failed to load system services';
      services = [];
    } finally {
      loading = false;
    }
  }

  function applyFilters() {
    filteredServices = services.filter(service => {
      // Filter by search query
      if (searchQuery) {
        const query = searchQuery.toLowerCase();
        if (!service.name.toLowerCase().includes(query) &&
            !service.description.toLowerCase().includes(query)) {
          return false;
        }
      }

      // Filter by status
      if (statusFilter && service.status !== statusFilter) {
        return false;
      }

      return true;
    });
  }

  // Watch filters and reapply
  $: {
    searchQuery;
    statusFilter;
    applyFilters();
  }

  function getStatusClass(status: string): string {
    switch (status.toLowerCase()) {
      case 'running':
        return 'status-running';
      case 'stopped':
        return 'status-stopped';
      case 'starting':
        return 'status-starting';
      case 'error':
      case 'failed':
        return 'status-error';
      default:
        return 'status-other';
    }
  }

  function getStatusIcon(status: string): string {
    switch (status.toLowerCase()) {
      case 'running':
        return '✅';
      case 'stopped':
        return '⏹️';
      case 'starting':
        return '🔄';
      case 'error':
      case 'failed':
        return '❌';
      default:
        return '⚪';
    }
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

<div class="services-view">
  <div class="view-header">
    <div class="header-left">
      <h2>System Services</h2>
      <p class="subtitle text-secondary">
        View and monitor system services and daemons
      </p>
    </div>
    <div class="header-actions">
      <select class="filter-select" bind:value={statusFilter}>
        <option value={null}>All Statuses</option>
        <option value="running">Running</option>
        <option value="stopped">Stopped</option>
        <option value="starting">Starting</option>
      </select>
      <input
        type="search"
        class="search-input"
        value={searchInput}
        on:input={handleSearchInput}
        placeholder="Search services..."
      />
      <button
        class="btn-secondary"
        on:click={loadServices}
        disabled={loading}
      >
        {loading ? 'Loading...' : '🔄 Refresh'}
      </button>
    </div>
  </div>

  <div class="view-content">
    {#if loading && services.length === 0}
      <div class="loading-state">
        <div class="spinner"></div>
        <p>Loading system services...</p>
      </div>
    {:else if error}
      <div class="error-state">
        <div class="error-icon">⚠️</div>
        <h3>Unable to load services</h3>
        <p class="text-secondary">{error}</p>
        <button class="btn-secondary" on:click={loadServices}>Try Again</button>
      </div>
    {:else if filteredServices.length === 0}
      <div class="empty-state">
        <div class="empty-icon">🔧</div>
        <h3>No services found</h3>
        <p class="text-secondary">
          {#if services.length === 0}
            No system services available.
          {:else}
            No services match your filters.
          {/if}
        </p>
      </div>
    {:else}
      <div class="table-wrapper">
        <table class="services-table">
          <thead>
            <tr>
              <th class="col-icon"></th>
              <th class="col-name">Service Name</th>
              <th class="col-status">Status</th>
              <th class="col-description">Description</th>
              <th class="col-pid">PID</th>
            </tr>
          </thead>
          <tbody>
            {#each filteredServices as service (service.name)}
              <tr>
                <td class="col-icon">
                  <span class="service-icon">{getStatusIcon(service.status)}</span>
                </td>
                <td class="col-name">
                  <span class="service-name">{service.name}</span>
                </td>
                <td class="col-status">
                  <span class="status-badge {getStatusClass(service.status)}">
                    {service.status}
                  </span>
                </td>
                <td class="col-description">
                  <span class="text-secondary">{service.description}</span>
                </td>
                <td class="col-pid">
                  {#if service.pid}
                    <span class="pid mono">{service.pid}</span>
                  {:else}
                    <span class="text-muted">-</span>
                  {/if}
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>

      <div class="table-footer">
        <span class="text-secondary">
          Showing {filteredServices.length} of {services.length} service{services.length !== 1 ? 's' : ''}
        </span>
      </div>
    {/if}
  </div>

  <!-- Backend API notice -->
  <div class="api-notice">
    <strong>⚠️ Backend API Required:</strong>
    <code>GET /api/v1/services</code>
    <br />
    Response: <code>{'{ services: [{ name, status, description, pid? }] }'}</code>
    <br />
    <small>Should run: sc query (Windows), launchctl list (macOS), systemctl list-units (Linux)</small>
  </div>
</div>

<style>
  .services-view {
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
    min-width: 120px;
    font-size: var(--font-size-sm);
  }

  .search-input {
    width: 200px;
    font-size: var(--font-size-sm);
  }

  .view-content {
    flex: 1;
    overflow: auto;
    display: flex;
    flex-direction: column;
  }

  /* Loading, error, empty states */
  .loading-state,
  .error-state,
  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: var(--spacing-xl);
    text-align: center;
    flex: 1;
  }

  .spinner {
    width: 40px;
    height: 40px;
    border: 4px solid var(--border-color);
    border-top-color: var(--accent-primary);
    border-radius: 50%;
    animation: spin 1s linear infinite;
    margin-bottom: var(--spacing-md);
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  .error-icon,
  .empty-icon {
    font-size: 4rem;
    margin-bottom: var(--spacing-md);
    opacity: 0.5;
  }

  .error-state h3,
  .empty-state h3 {
    margin-bottom: var(--spacing-sm);
    color: var(--text-primary);
  }

  /* Table */
  .table-wrapper {
    flex: 1;
    overflow: auto;
  }

  .services-table {
    width: 100%;
    border-collapse: collapse;
  }

  .services-table thead {
    position: sticky;
    top: 0;
    background-color: var(--bg-tertiary);
    z-index: 10;
  }

  .services-table th {
    text-align: left;
    padding: var(--spacing-md);
    font-weight: 600;
    font-size: var(--font-size-sm);
    color: var(--text-primary);
    border-bottom: 2px solid var(--border-color);
    white-space: nowrap;
  }

  .services-table td {
    padding: var(--spacing-md);
    border-bottom: 1px solid var(--border-color);
    font-size: var(--font-size-sm);
  }

  .services-table tbody tr:hover {
    background-color: var(--bg-hover);
  }

  .col-icon {
    width: 40px;
    text-align: center;
  }

  .col-name {
    min-width: 200px;
  }

  .col-status {
    width: 120px;
  }

  .col-description {
    width: auto;
  }

  .col-pid {
    width: 80px;
  }

  .service-icon {
    font-size: var(--font-size-lg);
  }

  .service-name {
    font-weight: 500;
    color: var(--text-primary);
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

  .status-other {
    background-color: var(--bg-tertiary);
    color: var(--text-muted);
  }

  .pid {
    font-family: var(--font-mono);
    font-size: var(--font-size-xs);
    color: var(--text-secondary);
  }

  .table-footer {
    padding: var(--spacing-md);
    border-top: 1px solid var(--border-color);
    background-color: var(--bg-tertiary);
    text-align: center;
    font-size: var(--font-size-sm);
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

    .col-description {
      display: none;
    }
  }
</style>
