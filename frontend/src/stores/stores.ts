import { writable, derived, type Writable } from 'svelte/store';

// Type definitions matching backend models
export interface MCPServer {
  id: string;
  name: string;
  version?: string;
  installationPath: string;
  transport: string;
  status: ServerStatus;
  pid?: number;
  capabilities?: string[];
  tools?: string[];
  configuration: ServerConfiguration;
  dependencies?: Dependency[];
  discoveredAt: string;
  lastSeenAt: string;
  source: string;
}

export interface ServerStatus {
  state: 'stopped' | 'starting' | 'running' | 'error';
  startupAttempts: number;
  lastStateChange: string;
  crashRecoverable: boolean;
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
  type: string;
  requiredVersion?: string;
  detectedVersion?: string;
  installationInstructions?: string;
}

export interface LogEntry {
  id: string;
  timestamp: string;
  severity: string;
  source: string;
  message: string;
  serverId: string;
  metadata?: Record<string, any>;
}

export interface ServerMetrics {
  serverId: string;
  uptime: number;
  memoryBytes?: number;
  requestCount?: number;
  timestamp: string;
}

export interface ApplicationState {
  version: string;
  lastSaved: string;
  preferences: UserPreferences;
  windowLayout: WindowLayout;
  filters: ServerFilters;
  discoveredServers: string[];
  monitoredConfigPaths: string[];
  lastDiscoveryScan: string;
  selectedServerId?: string;
  lastSyncedAt: string;
}

export interface UserPreferences {
  theme: string;
  logRetentionPerServer: number;
  autoStartServers: boolean;
  minimizeToTray: boolean;
  showNotifications: boolean;
}

export interface WindowLayout {
  width: number;
  height: number;
  x: number;
  y: number;
  maximized: boolean;
  logPanelHeight: number;
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
  version: '',
  lastSaved: '',
  preferences: {
    theme: 'system',
    logRetentionPerServer: 1000,
    autoStartServers: false,
    minimizeToTray: false,
    showNotifications: true,
  },
  windowLayout: {
    width: 1280,
    height: 800,
    x: 0,
    y: 0,
    maximized: false,
    logPanelHeight: 200,
  },
  filters: {
    status: undefined,
    source: undefined,
    searchQuery: ''
  },
  discoveredServers: [],
  monitoredConfigPaths: [],
  lastDiscoveryScan: '',
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
export const activeView: Writable<string> = writable('servers'); // servers, netstat, shell, explorer, help

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
      if ($appState.filters.searchQuery) {
        const query = $appState.filters.searchQuery.toLowerCase();
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

// Derived store: true when at least one HTTP/SSE transport server exists
export const hasNetworkTransportServers = derived(
  servers,
  $servers => $servers.some(s =>
    s.transport === 'sse' ||
    s.transport === 'http'
  )
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
