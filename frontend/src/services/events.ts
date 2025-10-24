import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';
import {
  updateServer,
  updateServerStatus,
  addLog,
  updateMetrics,
  addNotification,
  type MCPServer,
  type ServerStatus,
  type LogEntry,
  type ServerMetrics
} from '../stores/stores';

// Event types matching the backend
export type WailsEventType =
  | 'server:discovered'
  | 'server:status:changed'
  | 'server:log:entry'
  | 'server:config:updated'
  | 'server:metrics:updated'
  | 'servers:initial'
  | 'servers:discovered';

/**
 * Initialize Wails event listeners
 */
export function setupWailsEvents() {
  // Server discovered event
  EventsOn('server:discovered', (data: any) => {
    console.log('Server discovered:', data);
    handleServerDiscovered(data);
  });

  // Server status changed event
  EventsOn('server:status:changed', (data: any) => {
    console.log('Server status changed:', data);
    handleServerStatusChanged(data);
  });

  // Server log entry event
  EventsOn('server:log:entry', (data: any) => {
    console.log('Server log entry:', data);
    handleServerLogEntry(data);
  });

  // Server metrics updated event
  EventsOn('server:metrics:updated', (data: any) => {
    console.log('Server metrics updated:', data);
    handleServerMetricsUpdated(data);
  });

  // Server config updated event
  EventsOn('server:config:updated', (data: any) => {
    console.log('Server config updated:', data);
    handleServerConfigUpdated(data);
  });

  // Initial servers event (sent on startup)
  EventsOn('servers:initial', (servers: MCPServer[]) => {
    console.log('Initial servers received:', servers);
    if (servers && servers.length > 0) {
      addNotification('success', `Loaded ${servers.length} servers`);
    }
  });

  // Servers discovered event (sent after manual discovery)
  EventsOn('servers:discovered', (servers: MCPServer[]) => {
    console.log('Servers discovered:', servers);
    if (servers && servers.length > 0) {
      addNotification('info', `Discovered ${servers.length} servers`);
    }
  });

  console.log('Wails event listeners initialized');
}

/**
 * Clean up Wails event listeners
 */
export function cleanupWailsEvents() {
  EventsOff('server:discovered');
  EventsOff('server:status:changed');
  EventsOff('server:log:entry');
  EventsOff('server:metrics:updated');
  EventsOff('server:config:updated');
  EventsOff('servers:initial');
  EventsOff('servers:discovered');
  console.log('Wails event listeners cleaned up');
}

/**
 * Handle server discovered event
 */
function handleServerDiscovered(data: any) {
  // Backend sends: { serverID, name, source }
  if (data && data.serverID) {
    const serverName = data.name || data.serverID;
    addNotification('info', `Server discovered: ${serverName}`);
  }
}

/**
 * Handle server status changed event
 */
async function handleServerStatusChanged(data: any) {
  // Backend sends: { serverID, oldState, newState }
  if (data && data.serverID) {
    const serverId = data.serverID;
    const oldState = data.oldState;
    const newState = data.newState;

    if (oldState !== newState) {
      addNotification('info', `Server ${serverId}: ${oldState} → ${newState}`);

      // FR-005/FR-047: Fetch updated server and update store for real-time UI update
      try {
        const { GetServer } = await import('../../wailsjs/go/main/App');
        const updatedServer = await GetServer(serverId);
        if (updatedServer) {
          updateServer(updatedServer);
        }
      } catch (error) {
        console.error('Failed to fetch updated server:', error);
      }
    }
  }
}

/**
 * Handle server log entry event
 */
function handleServerLogEntry(data: any) {
  // Backend sends: { serverID, severity, message }
  if (data && data.serverID) {
    const serverId = data.serverID;
    const logEntry: LogEntry = {
      timestamp: new Date().toISOString(),
      severity: data.severity,
      message: data.message,
      serverId: serverId
    };

    addLog(serverId, logEntry);

    // Show notification for error logs
    if (logEntry.severity === 'error') {
      addNotification('error', `${serverId}: ${logEntry.message}`);
    }
  }
}

/**
 * Handle server metrics updated event
 */
function handleServerMetricsUpdated(data: any) {
  // Backend sends: { serverID, metrics }
  if (data && data.serverID && data.metrics) {
    const serverId = data.serverID;
    const metrics: ServerMetrics = data.metrics;
    updateMetrics(serverId, metrics);
    // Don't show notification for metrics updates (too frequent)
  }
}

/**
 * Handle server config updated event
 * FR-050: Notify user of external config changes and prompt to refresh
 */
function handleServerConfigUpdated(data: any) {
  // Backend sends: { filePath }
  if (data && data.filePath) {
    const fileName = data.filePath.split(/[/\\]/).pop() || data.filePath;
    // FR-050: Provide option to refresh discovery results
    addNotification('warning', `Configuration file changed (${fileName}). Click Refresh to update server list.`, 10000);
  }
}
