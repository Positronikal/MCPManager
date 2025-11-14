import { writable, derived, type Writable } from 'svelte/store';

// Type definitions matching backend models
export interface MCPServer {
  id: string;
  name: string;
  version?: string;
  installationPath: string;
  status: ServerStatus;
  pid?: number;
  capabilities?: string[];
  tools?: string[];
  configuration: ServerConfiguration;
  dependencies?: Dependency[];
  discoveredAt: string;
  lastSeenAt: string;
  source: 'client_config' | 'extension' | 'filesystem' | 'process';
}

export interface ServerStatus {
  state: 'stopped' | 'starting' | 'running' | 'error';
  uptime: number;
  lastChecked: string;
  lastStateChange: string;
  errorMessage?: string;
}

export interface ServerConfiguration {
  environmentVariables?: Record<string, string>;
  commandLineArguments?: string[];
  workingDirectory?: string;
  autoStart: boolean;
  restartOnCrash: boolean;
  maxRestartAttempts: number;
  startupTimeout: number;
  shutdownTimeout: number;
  healthCheckInterval?: number;
  healthCheckEndpoint?: string;
}

// Type alias for backward compatibility
export type LogSeverity = 'info' | 'success' | 'warning' | 'error';

export interface Dependency {
  name: string;
  type: 'runtime' | 'tool' | 'environment' | 'library';
  requiredVersion?: string;
  installedVersion?: string;
  satisfied: boolean;
  checkCommand?: string;
}

export interface LogEntry {
  timestamp: string;
  severity: LogSeverity;
  message: string;
  serverId: string;
}

export interface ServerMetrics {
  uptimeSeconds?: number;
  memoryUsageMB?: number;
  requestCount?: number;
  cpuPercent?: number;
}

export interface ApplicationState {
  userPreferences: UserPreferences;
  windowLayout: WindowLayout;
  serverFilters: ServerFilters;
  selectedServerId?: string;
  lastSyncedAt: string;
}

export interface UserPreferences {
  theme: 'light' | 'dark' | 'system';
  autoStartServers: boolean;
  showNotifications: boolean;
  logLevel: 'info' | 'warning' | 'error';
  refreshInterval: number;
  enableAutoDiscovery: boolean;
}

export interface WindowLayout {
  width: number;
  height: number;
  x: number;
  y: number;
  maximized: boolean;
  fullscreen: boolean;
}

export interface ServerFilters {
  status?: string;
  source?: string;
  searchQuery?: string;
}

export interface UpdateInfo {
  updateAvailable: boolean;
  currentVersion: string;
  latestVersion: string;
  releaseNotes?: string;
}

// Notification store
export interface Notification {
  id: string;
  type: 'info' | 'success' | 'warning' | 'error';
  message: string;
  timestamp: number;
  duration?: number;
}

// Store definitions
export const servers: Writable<MCPServer[]> = writable([]);
export const selectedServerId: Writable<string | null> = writable(null);
export const serverFilters: Writable<ServerFilters> = writable({
  status: undefined,
  source: undefined,
  searchQuery: ''
});

// Logs store (per-server logs)
export const serverLogs: Writable<Record<string, LogEntry[]>> = writable({});

// Global logs store (all logs aggregated)
export const logs: Writable<LogEntry[]> = writable([]);

// Selected severity for log filtering
export const selectedSeverity: Writable<LogSeverity | null> = writable(null);

// Metrics store (per-server metrics)
export const serverMetrics: Writable<Record<string, ServerMetrics>> = writable({});

// Dependencies store (per-server dependencies)
export const serverDependencies: Writable<Record<string, Dependency[]>> = writable({});

// Update info store (per-server update info)
export const serverUpdates: Writable<Record<string, UpdateInfo>> = writable({});

// Application state store
export const applicationState: Writable<ApplicationState> = writable({
  userPreferences: {
    theme: 'system',
    autoStartServers: false,
    showNotifications: true,
    logLevel: 'info',
    refreshInterval: 5,
    enableAutoDiscovery: true
  },
  windowLayout: {
    width: 1280,
    height: 800,
    x: 0,
    y: 0,
    maximized: false,
    fullscreen: false
  },
  serverFilters: {
    status: undefined,
    source: undefined,
    searchQuery: ''
  },
  selectedServerId: undefined,
  lastSyncedAt: new Date().toISOString()
});

// Notifications store
export const notifications: Writable<Notification[]> = writable([]);

// Loading states
export const isLoading: Writable<boolean> = writable(false);
export const isDiscovering: Writable<boolean> = writable(false);

// SSE connection state
export const isConnected: Writable<boolean> = writable(false);
export const lastEventId: Writable<string | null> = writable(null);

// Active view for main content area
export const activeView: Writable<string> = writable('servers'); // servers, netstat, shell, explorer, services, help

// Derived store for selected server
export const selectedServer = derived(
  [servers, selectedServerId],
  ([$servers, $selectedServerId]) => {
    if (!$selectedServerId) return null;
    return $servers.find(s => s.id === $selectedServerId) || null;
  }
);

// Derived store for filtered servers
export const filteredServers = derived(
  [servers, serverFilters],
  ([$servers, $filters]) => {
    let result = $servers;

    // Filter by status
    if ($filters.status) {
      result = result.filter(s => s.status.state === $filters.status);
    }

    // Filter by source
    if ($filters.source) {
      result = result.filter(s => s.source === $filters.source);
    }

    // Filter by search query
    if ($filters.searchQuery) {
      const query = $filters.searchQuery.toLowerCase();
      result = result.filter(s =>
        s.name.toLowerCase().includes(query) ||
        s.installationPath.toLowerCase().includes(query)
      );
    }

    return result;
  }
);

// Derived store for filtered logs (per Phase E directive)
export const filteredLogs = derived(
  [logs, selectedServerId, selectedSeverity, applicationState],
  ([$logs, $selectedServerId, $selectedSeverity, $appState]) => {
    return $logs.filter(log => {
      // Filter by selected server
      if ($selectedServerId && log.serverId !== $selectedServerId) return false;

      // Filter by selected severity
      if ($selectedSeverity && log.severity !== $selectedSeverity) return false;

      // Filter by search query
      if ($appState.serverFilters.searchQuery) {
        const query = $appState.serverFilters.searchQuery.toLowerCase();
        if (!log.message.toLowerCase().includes(query)) return false;
      }

      return true;
    });
  }
);

// Derived store for running servers
export const runningServers = derived(
  servers,
  $servers => $servers.filter(s => s.status.state === 'running')
);

// Helper functions for notifications
export function addNotification(type: Notification['type'], message: string, duration = 5000) {
  const id = crypto.randomUUID();
  const notification: Notification = {
    id,
    type,
    message,
    timestamp: Date.now(),
    duration
  };

  notifications.update(n => [...n, notification]);

  // Auto-remove after duration
  if (duration > 0) {
    setTimeout(() => {
      notifications.update(n => n.filter(notif => notif.id !== id));
    }, duration);
  }

  return id;
}

export function removeNotification(id: string) {
  notifications.update(n => n.filter(notif => notif.id !== id));
}

// Helper functions for server operations
export function updateServer(server: MCPServer) {
  servers.update(s => {
    const index = s.findIndex(srv => srv.id === server.id);
    if (index !== -1) {
      // Create new array with updated server to trigger reactivity
      return [...s.slice(0, index), server, ...s.slice(index + 1)];
    } else {
      // Create new array with added server to trigger reactivity
      return [...s, server];
    }
  });
}

export function removeServer(serverId: string) {
  servers.update(s => s.filter(srv => srv.id !== serverId));
}

export function updateServerStatus(serverId: string, status: ServerStatus) {
  servers.update(s => {
    const index = s.findIndex(srv => srv.id === serverId);
    if (index !== -1) {
      // Create new server object with updated status
      const updatedServer = { ...s[index], status };
      // Create new array with updated server to trigger reactivity
      return [...s.slice(0, index), updatedServer, ...s.slice(index + 1)];
    }
    return s;
  });
}

// Helper functions for logs
export function addLog(serverId: string, log: LogEntry) {
  // Update per-server logs
  serverLogs.update(logsMap => {
    if (!logsMap[serverId]) {
      logsMap[serverId] = [];
    }
    logsMap[serverId].push(log);
    // Keep only last 1000 logs per server
    if (logsMap[serverId].length > 1000) {
      logsMap[serverId] = logsMap[serverId].slice(-1000);
    }
    return logsMap;
  });

  // Update global logs store
  logs.update(globalLogs => {
    const updatedLogs = [...globalLogs, log];
    // Keep only last 1000 logs globally
    if (updatedLogs.length > 1000) {
      return updatedLogs.slice(-1000);
    }
    return updatedLogs;
  });
}

// Helper functions for metrics
export function updateMetrics(serverId: string, metrics: ServerMetrics) {
  serverMetrics.update(m => {
    m[serverId] = metrics;
    return m;
  });
}
