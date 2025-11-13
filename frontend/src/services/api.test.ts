// T-E027: Unit tests for API client
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { api } from './api';
import * as AppBindings from '../../wailsjs/go/main/App';

// Mock the Wails bindings
vi.mock('../../wailsjs/go/main/App');

describe('API Client', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Discovery API', () => {
    it('should call DiscoverServers', async () => {
      const mockResponse = {
        message: 'Discovery started',
        scanId: 'scan-123',
      };

      vi.mocked(AppBindings.DiscoverServers).mockResolvedValue(mockResponse as any);

      const result = await api.discovery.discoverServers();
      expect(AppBindings.DiscoverServers).toHaveBeenCalledTimes(1);
      expect(result.scanId).toBe('scan-123');
    });

    it('should handle DiscoverServers errors', async () => {
      vi.mocked(AppBindings.DiscoverServers).mockRejectedValue(new Error('Discovery failed'));

      await expect(api.discovery.discoverServers()).rejects.toThrow('Discovery failed');
    });

    it('should get server by ID', async () => {
      const mockServer = {
        id: 'server-1',
        name: 'Test Server',
      };

      vi.mocked(AppBindings.GetServer).mockResolvedValue(mockServer as any);

      const result = await api.discovery.getServer('server-1');
      expect(AppBindings.GetServer).toHaveBeenCalledWith('server-1');
      expect(result).toEqual(mockServer);
    });

    it('should list servers with filters', async () => {
      const mockResponse = {
        servers: [{ id: 'server-1', name: 'Test' }],
        count: 1,
        lastDiscovery: new Date().toISOString(),
      };

      vi.mocked(AppBindings.ListServers).mockResolvedValue(mockResponse as any);

      const result = await api.discovery.listServers();
      expect(AppBindings.ListServers).toHaveBeenCalledTimes(1);
      expect(result.servers).toHaveLength(1);
    });
  });

  describe('Lifecycle API', () => {
    it('should start server', async () => {
      const mockResponse = {
        message: 'Server started',
        serverId: 'server-1',
        status: 'running',
      };

      vi.mocked(AppBindings.StartServer).mockResolvedValue(mockResponse as any);

      const result = await api.lifecycle.startServer('server-1');
      expect(AppBindings.StartServer).toHaveBeenCalledWith('server-1');
      expect(result.message).toBe('Server started');
    });

    it('should stop server', async () => {
      const mockResponse = {
        message: 'Server stopped',
        serverId: 'server-1',
      };

      vi.mocked(AppBindings.StopServer).mockResolvedValue(mockResponse as any);

      const result = await api.lifecycle.stopServer('server-1');
      expect(AppBindings.StopServer).toHaveBeenCalledWith('server-1', false, 30);
      expect(result.message).toBe('Server stopped');
    });

    it('should restart server', async () => {
      const mockResponse = {
        message: 'Server restarted',
        serverId: 'server-1',
        status: 'running',
      };

      vi.mocked(AppBindings.RestartServer).mockResolvedValue(mockResponse as any);

      const result = await api.lifecycle.restartServer('server-1');
      expect(AppBindings.RestartServer).toHaveBeenCalledWith('server-1');
      expect(result.message).toBe('Server restarted');
    });

    it('should get server status', async () => {
      const mockStatus = {
        state: 'running',
        pid: 12345,
        port: 8080,
      };

      vi.mocked(AppBindings.GetServerStatus).mockResolvedValue(mockStatus as any);

      const result = await api.lifecycle.getServerStatus('server-1');
      expect(AppBindings.GetServerStatus).toHaveBeenCalledWith('server-1');
      expect(result.state).toBe('running');
      expect(result.pid).toBe(12345);
    });
  });

  describe('Configuration API', () => {
    it('should get server configuration', async () => {
      const mockConfig = {
        command: 'node',
        args: ['server.js'],
        env: { NODE_ENV: 'production' },
        autoStart: true,
      };

      vi.mocked(AppBindings.GetConfiguration).mockResolvedValue(mockConfig as any);

      const result = await api.config.getConfiguration('server-1');
      expect(AppBindings.GetConfiguration).toHaveBeenCalledWith('server-1');
      expect(result).toEqual(mockConfig);
    });

    it('should update server configuration', async () => {
      const newConfig = {
        command: 'python',
        args: ['server.py'],
        env: {},
        autoStart: false,
        restartOnFailure: true,
      };

      vi.mocked(AppBindings.UpdateConfiguration).mockResolvedValue(newConfig as any);

      const result = await api.config.updateConfiguration('server-1', newConfig);
      expect(AppBindings.UpdateConfiguration).toHaveBeenCalledWith('server-1', newConfig);
      expect(result).toEqual(newConfig);
    });
  });

  describe('Monitoring API', () => {
    it('should get server logs with filters', async () => {
      const mockResponse = {
        logs: [
          {
            timestamp: new Date().toISOString(),
            severity: 'info',
            message: 'Test log',
          },
        ],
        total: 1,
        hasMore: false,
      };

      vi.mocked(AppBindings.GetLogs).mockResolvedValue(mockResponse as any);

      const result = await api.monitoring.getServerLogs('server-1', {
        severity: 'info',
        limit: 100,
        offset: 0,
      });

      expect(AppBindings.GetLogs).toHaveBeenCalledWith('server-1', 'info', 100, 0);
      expect(result.logs).toHaveLength(1);
      expect(result.total).toBe(1);
    });

    it('should get all server logs', async () => {
      const mockResponse = {
        logs: [
          {
            timestamp: new Date().toISOString(),
            severity: 'info',
            message: 'Log 1',
          },
        ],
        total: 1,
      };

      vi.mocked(AppBindings.GetAllLogs).mockResolvedValue(mockResponse as any);

      const result = await api.monitoring.getAllLogs();
      expect(AppBindings.GetAllLogs).toHaveBeenCalledWith('', '', '', 100);
      expect(result.logs).toHaveLength(1);
    });

    it('should get server metrics', async () => {
      const mockMetrics = {
        cpu: 25.5,
        memory: 128,
        uptime: 3600,
      };

      vi.mocked(AppBindings.GetMetrics).mockResolvedValue(mockMetrics as any);

      const result = await api.monitoring.getServerMetrics('server-1');
      expect(AppBindings.GetMetrics).toHaveBeenCalledWith('server-1');
      expect(result.cpu).toBe(25.5);
    });
  });

  describe('Dependencies API', () => {
    it('should get dependencies', async () => {
      const mockResponse = {
        dependencies: [
          {
            name: 'Node.js',
            type: 'runtime',
            required: true,
            version: '18.0.0',
            satisfied: true,
          },
        ],
        allSatisfied: true,
      };

      vi.mocked(AppBindings.GetDependencies).mockResolvedValue(mockResponse as any);

      const result = await api.dependencies.getDependencies('server-1');
      expect(AppBindings.GetDependencies).toHaveBeenCalledWith('server-1');
      expect(result.allSatisfied).toBe(true);
      expect(result.dependencies).toHaveLength(1);
    });

    it('should get updates', async () => {
      const mockUpdates = {
        available: true,
        latestVersion: '2.0.0',
        currentVersion: '1.0.0',
      };

      vi.mocked(AppBindings.GetUpdates).mockResolvedValue(mockUpdates as any);

      const result = await api.dependencies.getUpdates('server-1');
      expect(AppBindings.GetUpdates).toHaveBeenCalledWith('server-1');
      expect(result.available).toBe(true);
    });
  });

  describe('Application State API', () => {
    it('should get application state', async () => {
      const mockState = {
        windowLayout: {
          width: 1024,
          height: 768,
          x: 0,
          y: 0,
          maximized: false,
          logPanelHeight: 200,
        },
      };

      vi.mocked(AppBindings.GetApplicationState).mockResolvedValue(mockState as any);

      const result = await api.appState.getApplicationState();
      expect(AppBindings.GetApplicationState).toHaveBeenCalledTimes(1);
      expect(result.windowLayout.width).toBe(1024);
    });

    it('should update application state', async () => {
      const newState = {
        windowLayout: {
          width: 1920,
          height: 1080,
          x: 100,
          y: 100,
          maximized: true,
          logPanelHeight: 300,
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
      };

      vi.mocked(AppBindings.UpdateApplicationState).mockResolvedValue(undefined);

      await api.appState.updateApplicationState(newState);
      expect(AppBindings.UpdateApplicationState).toHaveBeenCalledWith(newState);
    });
  });

  describe('Error handling', () => {
    it('should propagate errors from bindings', async () => {
      vi.mocked(AppBindings.StartServer).mockRejectedValue(new Error('Server start failed'));

      await expect(api.lifecycle.startServer('server-1')).rejects.toThrow('Server start failed');
    });

    it('should handle GetServer errors', async () => {
      vi.mocked(AppBindings.GetServer).mockRejectedValue(new Error('Server not found'));

      await expect(api.discovery.getServer('missing-server')).rejects.toThrow('Server not found');
    });
  });
});
