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

const API_BASE_URL = 'http://localhost:8080/api/v1';

// Helper function for making API requests
async function apiRequest<T>(
  endpoint: string,
  options: RequestInit = {}
): Promise<T> {
  const url = `${API_BASE_URL}${endpoint}`;

  const defaultOptions: RequestInit = {
    headers: {
      'Content-Type': 'application/json',
      ...options.headers
    },
    ...options
  };

  try {
    const response = await fetch(url, defaultOptions);

    if (!response.ok) {
      const error = await response.json().catch(() => ({}));
      throw new Error(error.error || `HTTP ${response.status}: ${response.statusText}`);
    }

    return await response.json();
  } catch (error) {
    console.error(`API request failed: ${endpoint}`, error);
    throw error;
  }
}

// Discovery API
export const discoveryAPI = {
  async listServers(filters?: { status?: string; source?: string }): Promise<{
    servers: MCPServer[];
    count: number;
    lastDiscovery: string;
  }> {
    const params = new URLSearchParams();
    if (filters?.status) params.append('status', filters.status);
    if (filters?.source) params.append('source', filters.source);

    const query = params.toString() ? `?${params.toString()}` : '';
    return apiRequest<{
      servers: MCPServer[];
      count: number;
      lastDiscovery: string;
    }>(`/servers${query}`);
  },

  async discoverServers(): Promise<{ message: string; scanId: string }> {
    return apiRequest<{ message: string; scanId: string }>('/servers/discover', {
      method: 'POST'
    });
  },

  async getServer(serverId: string): Promise<MCPServer> {
    return apiRequest<MCPServer>(`/servers/${serverId}`);
  }
};

// Lifecycle API
export const lifecycleAPI = {
  async startServer(serverId: string): Promise<{
    message: string;
    serverId: string;
    status: string;
  }> {
    return apiRequest<{
      message: string;
      serverId: string;
      status: string;
    }>(`/servers/${serverId}/start`, {
      method: 'POST'
    });
  },

  async stopServer(
    serverId: string,
    options?: { force?: boolean; timeout?: number }
  ): Promise<{ message: string; serverId: string }> {
    return apiRequest<{ message: string; serverId: string }>(
      `/servers/${serverId}/stop`,
      {
        method: 'POST',
        body: JSON.stringify(options || {})
      }
    );
  },

  async restartServer(serverId: string): Promise<{
    message: string;
    serverId: string;
  }> {
    return apiRequest<{ message: string; serverId: string }>(
      `/servers/${serverId}/restart`,
      {
        method: 'POST'
      }
    );
  },

  async getServerStatus(serverId: string): Promise<ServerStatus> {
    return apiRequest<ServerStatus>(`/servers/${serverId}/status`);
  }
};

// Configuration API
export const configAPI = {
  async getConfiguration(serverId: string): Promise<ServerConfiguration> {
    return apiRequest<ServerConfiguration>(`/servers/${serverId}/configuration`);
  },

  async updateConfiguration(
    serverId: string,
    config: ServerConfiguration
  ): Promise<ServerConfiguration> {
    return apiRequest<ServerConfiguration>(
      `/servers/${serverId}/configuration`,
      {
        method: 'PUT',
        body: JSON.stringify(config)
      }
    );
  }
};

// Monitoring API
export const monitoringAPI = {
  async getServerLogs(
    serverId: string,
    options?: { severity?: string; limit?: number; offset?: number }
  ): Promise<{ logs: LogEntry[]; total: number; hasMore: boolean }> {
    const params = new URLSearchParams();
    if (options?.severity) params.append('severity', options.severity);
    if (options?.limit) params.append('limit', options.limit.toString());
    if (options?.offset) params.append('offset', options.offset.toString());

    const query = params.toString() ? `?${params.toString()}` : '';
    return apiRequest<{ logs: LogEntry[]; total: number; hasMore: boolean }>(
      `/servers/${serverId}/logs${query}`
    );
  },

  async getAllLogs(options?: {
    serverId?: string;
    severity?: string;
    search?: string;
    limit?: number;
  }): Promise<{ logs: LogEntry[]; total: number }> {
    const params = new URLSearchParams();
    if (options?.serverId) params.append('serverId', options.serverId);
    if (options?.severity) params.append('severity', options.severity);
    if (options?.search) params.append('search', options.search);
    if (options?.limit) params.append('limit', options.limit.toString());

    const query = params.toString() ? `?${params.toString()}` : '';
    return apiRequest<{ logs: LogEntry[]; total: number }>(`/logs${query}`);
  },

  async getServerMetrics(serverId: string): Promise<ServerMetrics> {
    return apiRequest<ServerMetrics>(`/servers/${serverId}/metrics`);
  }
};

// Dependencies API
export const dependenciesAPI = {
  async getDependencies(serverId: string): Promise<{
    dependencies: Dependency[];
    allSatisfied: boolean;
  }> {
    return apiRequest<{ dependencies: Dependency[]; allSatisfied: boolean }>(
      `/servers/${serverId}/dependencies`
    );
  },

  async getUpdates(serverId: string): Promise<UpdateInfo> {
    return apiRequest<UpdateInfo>(`/servers/${serverId}/updates`);
  }
};

// Application State API
export const appStateAPI = {
  async getApplicationState(): Promise<ApplicationState> {
    return apiRequest<ApplicationState>('/application/state');
  },

  async updateApplicationState(
    state: ApplicationState
  ): Promise<{ message: string }> {
    return apiRequest<{ message: string }>('/application/state', {
      method: 'PUT',
      body: JSON.stringify(state)
    });
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
