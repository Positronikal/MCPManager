// T-E026: Unit tests for Svelte stores
import { describe, it, expect, beforeEach } from 'vitest';
import { get } from 'svelte/store';
import {
  servers,
  logs,
  selectedServerId,
  selectedSeverity,
  filteredLogs,
  runningServers,
  isConnected,
  isDiscovering,
  notifications,
  applicationState,
  addNotification,
  type MCPServer,
  type LogEntry,
} from './stores';

describe('Stores', () => {
  beforeEach(() => {
    // Reset stores before each test
    servers.set([]);
    logs.set([]);
    selectedServerId.set(null);
    selectedSeverity.set(null);
    isConnected.set(false);
    isDiscovering.set(false);
    notifications.set([]);
    applicationState.set({
      windowLayout: {
        width: 1024,
        height: 768,
        x: 0,
        y: 0,
        maximized: false,
        logPanelHeight: 200,
      },
      userPreferences: {
        theme: 'dark',
        autoRefresh: true,
        refreshInterval: 30,
      },
      serverFilters: {
        searchQuery: '',
        selectedSource: null,
        selectedStatus: null,
      },
      lastSyncedAt: new Date().toISOString(),
    });
  });

  describe('Basic stores', () => {
    it('should initialize servers as empty array', () => {
      expect(get(servers)).toEqual([]);
    });

    it('should update servers', () => {
      const mockServers: MCPServer[] = [
        {
          id: 'server-1',
          name: 'Test Server',
          source: 'client_config',
          installationPath: '/path/to/server',
          version: '1.0.0',
          status: {
            state: 'running',
            pid: 12345,
            port: 8080,
            startedAt: new Date().toISOString(),
            uptime: 100,
            lastChecked: new Date().toISOString(),
          },
          configuration: {
            command: 'node',
            args: ['server.js'],
            env: {},
            autoStart: false,
            restartOnFailure: false,
          },
        },
      ];
      servers.set(mockServers);
      expect(get(servers)).toEqual(mockServers);
    });

    it('should update logs', () => {
      const mockLogs: LogEntry[] = [
        {
          serverId: 'server-1',
          timestamp: new Date().toISOString(),
          severity: 'info',
          message: 'Test log message',
        },
      ];
      logs.set(mockLogs);
      expect(get(logs)).toEqual(mockLogs);
    });

    it('should update isConnected', () => {
      isConnected.set(true);
      expect(get(isConnected)).toBe(true);
    });

    it('should update isDiscovering', () => {
      isDiscovering.set(true);
      expect(get(isDiscovering)).toBe(true);
    });
  });

  describe('Derived stores', () => {
    it('should filter logs by serverId', () => {
      logs.set([
        {
          serverId: 'server-1',
          timestamp: new Date().toISOString(),
          severity: 'info',
          message: 'Server 1 log',
        },
        {
          serverId: 'server-2',
          timestamp: new Date().toISOString(),
          severity: 'info',
          message: 'Server 2 log',
        },
      ]);

      selectedServerId.set('server-1');
      const filtered = get(filteredLogs);
      expect(filtered).toHaveLength(1);
      expect(filtered[0].serverId).toBe('server-1');
    });

    it('should filter logs by severity', () => {
      logs.set([
        {
          serverId: 'server-1',
          timestamp: new Date().toISOString(),
          severity: 'info',
          message: 'Info log',
        },
        {
          serverId: 'server-1',
          timestamp: new Date().toISOString(),
          severity: 'error',
          message: 'Error log',
        },
      ]);

      selectedSeverity.set('error');
      const filtered = get(filteredLogs);
      expect(filtered).toHaveLength(1);
      expect(filtered[0].severity).toBe('error');
    });

    it('should filter logs by search query', () => {
      logs.set([
        {
          serverId: 'server-1',
          timestamp: new Date().toISOString(),
          severity: 'info',
          message: 'Connection established',
        },
        {
          serverId: 'server-1',
          timestamp: new Date().toISOString(),
          severity: 'info',
          message: 'Server started',
        },
      ]);

      applicationState.update((state) => ({
        ...state,
        serverFilters: {
          ...state.serverFilters,
          searchQuery: 'connection',
        },
      }));

      const filtered = get(filteredLogs);
      expect(filtered).toHaveLength(1);
      expect(filtered[0].message).toContain('Connection');
    });

    it('should return only running servers', () => {
      servers.set([
        {
          id: 'server-1',
          name: 'Running Server',
          source: 'client_config',
          installationPath: '/path/1',
          status: {
            state: 'running',
            pid: 12345,
            uptime: 100,
            lastChecked: new Date().toISOString(),
          },
          configuration: {
            command: 'node',
            args: [],
            env: {},
            autoStart: false,
            restartOnFailure: false,
          },
        },
        {
          id: 'server-2',
          name: 'Stopped Server',
          source: 'client_config',
          installationPath: '/path/2',
          status: {
            state: 'stopped',
            uptime: 0,
            lastChecked: new Date().toISOString(),
          },
          configuration: {
            command: 'node',
            args: [],
            env: {},
            autoStart: false,
            restartOnFailure: false,
          },
        },
      ]);

      const running = get(runningServers);
      expect(running).toHaveLength(1);
      expect(running[0].id).toBe('server-1');
      expect(running[0].status.state).toBe('running');
    });
  });

  describe('Notifications', () => {
    it('should add notification with auto-generated ID', () => {
      addNotification('info', 'Test notification');
      const notifs = get(notifications);
      expect(notifs).toHaveLength(1);
      expect(notifs[0].type).toBe('info');
      expect(notifs[0].message).toBe('Test notification');
      expect(notifs[0].id).toBeDefined();
    });

    it('should auto-dismiss notifications after 5 seconds', async () => {
      addNotification('info', 'Auto-dismiss test');
      expect(get(notifications)).toHaveLength(1);

      // Wait for auto-dismiss timeout
      await new Promise((resolve) => setTimeout(resolve, 5100));
      expect(get(notifications)).toHaveLength(0);
    }, 6000);

    it('should support multiple notification types', () => {
      addNotification('info', 'Info message');
      addNotification('success', 'Success message');
      addNotification('warning', 'Warning message');
      addNotification('error', 'Error message');

      const notifs = get(notifications);
      expect(notifs).toHaveLength(4);
      expect(notifs.map((n) => n.type)).toEqual(['info', 'success', 'warning', 'error']);
    });
  });

  describe('Application state', () => {
    it('should update window layout', () => {
      applicationState.update((state) => ({
        ...state,
        windowLayout: {
          ...state.windowLayout,
          width: 1920,
          height: 1080,
          maximized: true,
        },
      }));

      const state = get(applicationState);
      expect(state.windowLayout.width).toBe(1920);
      expect(state.windowLayout.height).toBe(1080);
      expect(state.windowLayout.maximized).toBe(true);
    });

    it('should update user preferences', () => {
      applicationState.update((state) => ({
        ...state,
        userPreferences: {
          ...state.userPreferences,
          theme: 'light',
          autoRefresh: false,
        },
      }));

      const state = get(applicationState);
      expect(state.userPreferences.theme).toBe('light');
      expect(state.userPreferences.autoRefresh).toBe(false);
    });

    it('should update server filters', () => {
      applicationState.update((state) => ({
        ...state,
        serverFilters: {
          searchQuery: 'test',
          selectedSource: 'client_config',
          selectedStatus: 'running',
        },
      }));

      const state = get(applicationState);
      expect(state.serverFilters.searchQuery).toBe('test');
      expect(state.serverFilters.selectedSource).toBe('client_config');
      expect(state.serverFilters.selectedStatus).toBe('running');
    });
  });
});
