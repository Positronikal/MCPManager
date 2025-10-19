import type {
  MCPServer,
  ServerConfiguration,
  LogEntry,
  ServerMetrics,
  Dependency,
  UpdateInfo,
  ApplicationState,
  ServerStatus
} from '../stores/stores';

// Import Wails bindings
import * as WailsApp from '../../wailsjs/go/main/App';

// Discovery API
export const discoveryAPI = {
  async listServers(filters?: { status?: string; source?: string }): Promise<{
    servers: MCPServer[];
    count: number;
    lastDiscovery: string;
  }> {
    const response = await WailsApp.ListServers();
    // Note: Wails bindings don't support query params in the same way
    // Filtering would need to be done client-side or added as parameters to the Go method
    let servers = response.servers || [];

    // Apply client-side filtering if needed
    if (filters?.status) {
      servers = servers.filter(s => s.status?.state === filters.status);
    }
    if (filters?.source) {
      servers = servers.filter(s => s.source === filters.source);
    }

    return {
      servers,
      count: servers.length,
      lastDiscovery: response.lastDiscovery
    };
  },

  async discoverServers(): Promise<{ message: string; scanId: string }> {
    return await WailsApp.DiscoverServers();
  },

  async getServer(serverId: string): Promise<MCPServer> {
    return await WailsApp.GetServer(serverId);
  }
};

// Lifecycle API
export const lifecycleAPI = {
  async startServer(serverId: string): Promise<{
    message: string;
    serverId: string;
    status: string;
  }> {
    return await WailsApp.StartServer(serverId);
  },

  async stopServer(
    serverId: string,
    options?: { force?: boolean; timeout?: number }
  ): Promise<{ message: string; serverId: string }> {
    const force = options?.force || false;
    const timeout = options?.timeout || 30;
    return await WailsApp.StopServer(serverId, force, timeout);
  },

  async restartServer(serverId: string): Promise<{
    message: string;
    serverId: string;
  }> {
    return await WailsApp.RestartServer(serverId);
  },

  async getServerStatus(serverId: string): Promise<ServerStatus> {
    return await WailsApp.GetServerStatus(serverId);
  }
};

// Configuration API
export const configAPI = {
  async getConfiguration(serverId: string): Promise<ServerConfiguration> {
    return await WailsApp.GetConfiguration(serverId);
  },

  async updateConfiguration(
    serverId: string,
    config: ServerConfiguration
  ): Promise<ServerConfiguration> {
    return await WailsApp.UpdateConfiguration(serverId, config);
  }
};

// Monitoring API
export const monitoringAPI = {
  async getServerLogs(
    serverId: string,
    options?: { severity?: string; limit?: number; offset?: number }
  ): Promise<{ logs: LogEntry[]; total: number; hasMore: boolean }> {
    const severity = options?.severity || '';
    const limit = options?.limit || 100;
    const offset = options?.offset || 0;

    return await WailsApp.GetLogs(serverId, severity, limit, offset);
  },

  async getAllLogs(options?: {
    serverId?: string;
    severity?: string;
    search?: string;
    limit?: number;
  }): Promise<{ logs: LogEntry[]; total: number }> {
    const serverId = options?.serverId || '';
    const severity = options?.severity || '';
    const search = options?.search || '';
    const limit = options?.limit || 100;

    const response = await WailsApp.GetAllLogs(serverId, severity, search, limit);
    return {
      logs: response.logs || [],
      total: response.total || 0
    };
  },

  async getServerMetrics(serverId: string): Promise<ServerMetrics> {
    return await WailsApp.GetMetrics(serverId);
  }
};

// Dependencies API
export const dependenciesAPI = {
  async getDependencies(serverId: string): Promise<{
    dependencies: Dependency[];
    allSatisfied: boolean;
  }> {
    return await WailsApp.GetDependencies(serverId);
  },

  async getUpdates(serverId: string): Promise<UpdateInfo> {
    return await WailsApp.GetUpdates(serverId);
  }
};

// Application State API
export const appStateAPI = {
  async getApplicationState(): Promise<ApplicationState> {
    return await WailsApp.GetApplicationState();
  },

  async updateApplicationState(
    state: ApplicationState
  ): Promise<{ message: string }> {
    return await WailsApp.UpdateApplicationState(state);
  }
};

// Export all APIs
export const api = {
  discovery: discoveryAPI,
  lifecycle: lifecycleAPI,
  config: configAPI,
  monitoring: monitoringAPI,
  dependencies: dependenciesAPI,
  appState: appStateAPI
};

export default api;
