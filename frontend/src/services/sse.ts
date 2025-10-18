import {
  updateServer,
  updateServerStatus,
  addLog,
  updateMetrics,
  isConnected,
  lastEventId,
  addNotification,
  type MCPServer,
  type ServerStatus,
  type LogEntry,
  type ServerMetrics
} from '../stores/stores';

const SSE_URL = 'http://localhost:8080/api/v1/events';

export interface ServerEvent {
  id: string;
  type: 'server.discovered' | 'server.status.changed' | 'server.log.entry' | 'config.file.changed' | 'server.metrics.updated';
  timestamp: string;
  metadata?: Record<string, string>;
  data?: any;
}

export class SSEClient {
  private eventSource: EventSource | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 10;
  private reconnectDelay = 1000; // Start with 1 second
  private maxReconnectDelay = 30000; // Max 30 seconds
  private serverIds: string[] = [];
  private isManuallyDisconnected = false;
  private lastReceivedEventId: string | null = null;

  constructor(serverIds: string[] = []) {
    this.serverIds = serverIds;
  }

  /**
   * Connect to the SSE stream
   */
  connect() {
    this.isManuallyDisconnected = false;
    this.createEventSource();
  }

  /**
   * Disconnect from the SSE stream
   */
  disconnect() {
    this.isManuallyDisconnected = true;
    if (this.eventSource) {
      this.eventSource.close();
      this.eventSource = null;
    }
    isConnected.set(false);
    this.reconnectAttempts = 0;
  }

  /**
   * Create and configure EventSource
   */
  private createEventSource() {
    // Build URL with query parameters
    const url = new URL(SSE_URL);
    if (this.serverIds.length > 0) {
      url.searchParams.set('serverIds', this.serverIds.join(','));
    }

    // Create EventSource with Last-Event-ID if available
    const init: EventSourceInit = {};

    // Note: EventSource doesn't support custom headers in the constructor
    // We'll need to handle reconnection with Last-Event-ID via query param
    // or use a polyfill that supports headers
    this.eventSource = new EventSource(url.toString());

    // Connection opened
    this.eventSource.onopen = () => {
      console.log('SSE connection established');
      isConnected.set(true);
      this.reconnectAttempts = 0;
      this.reconnectDelay = 1000;

      addNotification('success', 'Connected to server');
    };

    // Handle errors
    this.eventSource.onerror = (error) => {
      console.error('SSE connection error:', error);
      isConnected.set(false);

      if (!this.isManuallyDisconnected) {
        this.handleReconnect();
      }
    };

    // Listen for specific event types (using dot notation from backend)
    this.eventSource.addEventListener('server.discovered', (event: MessageEvent) => {
      this.handleEvent(event);
    });

    this.eventSource.addEventListener('server.status.changed', (event: MessageEvent) => {
      this.handleEvent(event);
    });

    this.eventSource.addEventListener('server.log.entry', (event: MessageEvent) => {
      this.handleEvent(event);
    });

    this.eventSource.addEventListener('config.file.changed', (event: MessageEvent) => {
      this.handleEvent(event);
    });

    this.eventSource.addEventListener('server.metrics.updated', (event: MessageEvent) => {
      this.handleEvent(event);
    });

    // Handle generic messages
    this.eventSource.onmessage = (event: MessageEvent) => {
      this.handleEvent(event);
    };
  }

  /**
   * Handle incoming SSE events
   */
  private handleEvent(event: MessageEvent) {
    try {
      // Store last event ID
      if (event.lastEventId) {
        this.lastReceivedEventId = event.lastEventId;
        lastEventId.set(event.lastEventId);
      }

      // Parse event data
      const eventData: ServerEvent = JSON.parse(event.data);

      console.log('Received SSE event:', eventData.type, eventData);

      // Route event to appropriate handler
      switch (eventData.type) {
        case 'server.discovered':
          this.handleServerDiscovered(eventData);
          break;
        case 'server.status.changed':
          this.handleServerStatusChanged(eventData);
          break;
        case 'server.log.entry':
          this.handleServerLogEntry(eventData);
          break;
        case 'config.file.changed':
          this.handleServerConfigChanged(eventData);
          break;
        case 'server.metrics.updated':
          this.handleServerMetricsUpdated(eventData);
          break;
        default:
          console.warn('Unknown event type:', eventData.type);
      }
    } catch (error) {
      console.error('Failed to handle SSE event:', error);
    }
  }

  /**
   * Handle server discovered event
   */
  private handleServerDiscovered(event: ServerEvent) {
    // Backend sends: { type: "server.discovered", data: { serverID, name, source } }
    // We need to fetch full server details from API
    if (event.data && event.data.serverID) {
      const serverName = event.data.name || event.data.serverID;
      addNotification('info', `Server discovered: ${serverName}`);
      // Full server data will come from polling or API call
      // For now, just notify user
    }
  }

  /**
   * Handle server status changed event
   */
  private handleServerStatusChanged(event: ServerEvent) {
    // Backend sends: { type: "server.status.changed", data: { serverID, oldState, newState } }
    if (event.data && event.data.serverID) {
      const serverId = event.data.serverID;
      const oldState = event.data.oldState;
      const newState = event.data.newState;

      // Update server status in store
      // Note: We don't have full ServerStatus object, so we'll need to fetch it
      // For now, just show notification
      if (oldState !== newState) {
        addNotification('info', `Server ${serverId}: ${oldState} → ${newState}`);
      }
    }
  }

  /**
   * Handle server log entry event
   */
  private handleServerLogEntry(event: ServerEvent) {
    // Backend sends: { type: "server.log.entry", data: { serverID, severity, message } }
    if (event.data && event.data.serverID) {
      const serverId = event.data.serverID;
      const logEntry: LogEntry = {
        timestamp: event.timestamp,
        severity: event.data.severity,
        message: event.data.message,
        serverId: serverId
      };

      addLog(serverId, logEntry);

      // Show notification for error logs
      if (logEntry.severity === 'error') {
        const serverName = event.metadata?.serverName || serverId;
        addNotification('error', `${serverName}: ${logEntry.message}`);
      }
    }
  }

  /**
   * Handle config file changed event
   */
  private handleServerConfigChanged(event: ServerEvent) {
    // Backend sends: { type: "config.file.changed", data: { filePath } }
    if (event.data && event.data.filePath) {
      const filePath = event.data.filePath;
      addNotification('info', `Configuration file changed: ${filePath}`);
      // Could trigger a re-fetch of servers from the config file
    }
  }

  /**
   * Handle server metrics updated event
   */
  private handleServerMetricsUpdated(event: ServerEvent) {
    // Backend sends: { type: "server.metrics.updated", data: { serverID, metrics } }
    if (event.data && event.data.serverID && event.data.metrics) {
      const serverId = event.data.serverID;
      const metrics: ServerMetrics = event.data.metrics;
      updateMetrics(serverId, metrics);
      // Don't show notification for metrics updates (too frequent)
    }
  }

  /**
   * Handle reconnection with exponential backoff
   */
  private handleReconnect() {
    if (this.isManuallyDisconnected) {
      return;
    }

    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error('Max reconnection attempts reached');
      addNotification('error', 'Lost connection to server. Please refresh.');
      return;
    }

    this.reconnectAttempts++;
    const delay = Math.min(
      this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1),
      this.maxReconnectDelay
    );

    console.log(`Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts})`);

    setTimeout(() => {
      if (!this.isManuallyDisconnected) {
        console.log('Attempting to reconnect...');
        this.createEventSource();
      }
    }, delay);
  }

  /**
   * Update server filter
   */
  setServerFilter(serverIds: string[]) {
    this.serverIds = serverIds;
    // Reconnect with new filter
    if (this.eventSource) {
      this.disconnect();
      this.connect();
    }
  }
}

// Create and export singleton instance
let sseClient: SSEClient | null = null;

export function getSSEClient(): SSEClient {
  if (!sseClient) {
    sseClient = new SSEClient();
  }
  return sseClient;
}

export function connectSSE(serverIds: string[] = []): SSEClient {
  const client = getSSEClient();
  if (serverIds.length > 0) {
    client.setServerFilter(serverIds);
  }
  client.connect();
  return client;
}

export function disconnectSSE() {
  if (sseClient) {
    sseClient.disconnect();
  }
}
