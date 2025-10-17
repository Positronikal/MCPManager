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
  type: 'server_discovered' | 'server_status_changed' | 'server_log_entry' | 'server_config_changed' | 'server_metrics_updated';
  timestamp: string;
  metadata: Record<string, string>;
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

    // Listen for specific event types
    this.eventSource.addEventListener('server_discovered', (event: MessageEvent) => {
      this.handleEvent(event);
    });

    this.eventSource.addEventListener('server_status_changed', (event: MessageEvent) => {
      this.handleEvent(event);
    });

    this.eventSource.addEventListener('server_log_entry', (event: MessageEvent) => {
      this.handleEvent(event);
    });

    this.eventSource.addEventListener('server_config_changed', (event: MessageEvent) => {
      this.handleEvent(event);
    });

    this.eventSource.addEventListener('server_metrics_updated', (event: MessageEvent) => {
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
        case 'server_discovered':
          this.handleServerDiscovered(eventData);
          break;
        case 'server_status_changed':
          this.handleServerStatusChanged(eventData);
          break;
        case 'server_log_entry':
          this.handleServerLogEntry(eventData);
          break;
        case 'server_config_changed':
          this.handleServerConfigChanged(eventData);
          break;
        case 'server_metrics_updated':
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
    if (event.data && event.data.server) {
      const server: MCPServer = event.data.server;
      updateServer(server);
      addNotification('info', `Server discovered: ${server.name}`);
    }
  }

  /**
   * Handle server status changed event
   */
  private handleServerStatusChanged(event: ServerEvent) {
    const serverId = event.metadata.serverId;
    if (serverId && event.data && event.data.newStatus) {
      const status: ServerStatus = event.data.newStatus;
      updateServerStatus(serverId, status);

      // Show notification for significant status changes
      const oldState = event.data.oldStatus?.state;
      const newState = status.state;

      if (oldState !== newState) {
        const serverName = event.metadata.serverName || serverId;
        addNotification('info', `${serverName}: ${oldState} → ${newState}`);
      }
    }
  }

  /**
   * Handle server log entry event
   */
  private handleServerLogEntry(event: ServerEvent) {
    const serverId = event.metadata.serverId;
    if (serverId && event.data && event.data.logEntry) {
      const logEntry: LogEntry = event.data.logEntry;
      addLog(serverId, logEntry);

      // Show notification for error logs
      if (logEntry.severity === 'error') {
        addNotification('error', `${event.metadata.serverName || serverId}: ${logEntry.message}`);
      }
    }
  }

  /**
   * Handle server config changed event
   */
  private handleServerConfigChanged(event: ServerEvent) {
    const serverId = event.metadata.serverId;
    if (serverId) {
      const serverName = event.metadata.serverName || serverId;
      addNotification('info', `Configuration updated: ${serverName}`);
      // Could trigger a re-fetch of the server data here
    }
  }

  /**
   * Handle server metrics updated event
   */
  private handleServerMetricsUpdated(event: ServerEvent) {
    const serverId = event.metadata.serverId;
    if (serverId && event.data && event.data.metrics) {
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
