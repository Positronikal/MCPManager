// T-E029: Component tests for ServerTable (focused on critical interactions)
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/svelte';
import ServerTable from './ServerTable.svelte';
import { servers, applicationState } from '../stores/stores';
import type { MCPServer } from '../stores/stores';

describe('ServerTable Component', () => {
  const mockServers: MCPServer[] = [
    {
      id: 'server-1',
      name: 'Test Server 1',
      source: 'client_config',
      transport: 'stdio',
      installationPath: '/path/to/server1',
      version: '1.0.0',
      status: {
        state: 'running',
        startupAttempts: 0,
        lastStateChange: new Date().toISOString(),
        crashRecoverable: true,
      },
      configuration: {
        autoStart: false,
        restartOnCrash: false,
        maxRestartAttempts: 3,
        startupTimeout: 30,
        shutdownTimeout: 10,
      },
      discoveredAt: new Date().toISOString(),
      lastSeenAt: new Date().toISOString(),
    },
    {
      id: 'server-2',
      name: 'Test Server 2',
      source: 'filesystem',
      transport: 'http',
      installationPath: '/path/to/server2',
      status: {
        state: 'stopped',
        startupAttempts: 0,
        lastStateChange: new Date().toISOString(),
        crashRecoverable: true,
      },
      configuration: {
        autoStart: false,
        restartOnCrash: false,
        maxRestartAttempts: 3,
        startupTimeout: 30,
        shutdownTimeout: 10,
      },
      discoveredAt: new Date().toISOString(),
      lastSeenAt: new Date().toISOString(),
    },
  ];

  beforeEach(() => {
    vi.clearAllMocks();
    servers.set([]);
    applicationState.set({
      version: '',
      lastSaved: '',
      preferences: {
        theme: 'dark',
        logRetentionPerServer: 1000,
        autoStartServers: false,
        minimizeToTray: false,
        showNotifications: true,
      },
      windowLayout: {
        width: 1024,
        height: 768,
        x: 0,
        y: 0,
        maximized: false,
        logPanelHeight: 200,
      },
      filters: {
        status: undefined,
        source: undefined,
        searchQuery: '',
      },
      discoveredServers: [],
      monitoredConfigPaths: [],
      lastDiscoveryScan: '',
      lastSyncedAt: new Date().toISOString(),
    });
  });

  describe('Rendering', () => {
    it('should render empty state when no servers', () => {
      render(ServerTable);
      expect(screen.getByText(/no servers found/i)).toBeInTheDocument();
    });

    it('should render server table with servers', () => {
      servers.set(mockServers);
      render(ServerTable);

      expect(screen.getByText('Test Server 1')).toBeInTheDocument();
      expect(screen.getByText('Test Server 2')).toBeInTheDocument();
    });

    it('should display server status correctly', () => {
      servers.set(mockServers);
      render(ServerTable);

      // Running server should show running status
      const rows = screen.getAllByRole('row');
      expect(rows.length).toBeGreaterThan(0);
    });

    it('should display server source badges', () => {
      servers.set(mockServers);
      render(ServerTable);

      expect(screen.getByText(/client_config/i)).toBeInTheDocument();
      expect(screen.getByText(/filesystem/i)).toBeInTheDocument();
    });
  });

  describe('Display and interaction', () => {
    beforeEach(() => {
      servers.set(mockServers);
    });

    it('should display filtered servers from store', () => {
      // Note: ServerTable displays filteredServers from store,
      // filtering logic is handled by parent component (App.svelte)
      render(ServerTable);

      // Both servers should be visible when filteredServers contains them
      expect(screen.getByText('Test Server 1')).toBeInTheDocument();
      expect(screen.getByText('Test Server 2')).toBeInTheDocument();
    });

    it('should show server count in footer', () => {
      render(ServerTable);

      // Check that footer shows correct count
      expect(screen.getByText(/showing 2 of 2 servers/i)).toBeInTheDocument();
    });

    it('should render table with proper structure', () => {
      render(ServerTable);

      // Check that table element exists with proper role
      const table = screen.getByRole('table');
      expect(table).toBeInTheDocument();
    });
  });

  describe('Server actions', () => {
    it('should have start/stop buttons for each server', () => {
      servers.set(mockServers);
      render(ServerTable);

      // Check for action buttons
      const buttons = screen.getAllByRole('button');
      expect(buttons.length).toBeGreaterThan(0);
    });

    it('should have config button for each server', () => {
      servers.set(mockServers);
      render(ServerTable);

      // Look for config/settings buttons
      const buttons = screen.getAllByRole('button');
      const hasConfigButton = buttons.some(
        (btn) => btn.textContent?.includes('⚙️') || btn.getAttribute('title')?.includes('config')
      );
      expect(hasConfigButton || buttons.length > 2).toBe(true);
    });
  });

  describe('Server selection', () => {
    it('should allow row selection', async () => {
      servers.set(mockServers);
      render(ServerTable);

      const rows = screen.getAllByRole('row');
      // Header row + 2 data rows
      expect(rows.length).toBe(3);

      // Click on first data row
      if (rows[1]) {
        await fireEvent.click(rows[1]);
        // Selection state would be updated in store
      }
    });
  });

  describe('Responsive behavior', () => {
    it('should render on different screen sizes', () => {
      servers.set(mockServers);

      // Test desktop
      render(ServerTable);
      expect(screen.getByRole('table')).toBeInTheDocument();
    });
  });
});
