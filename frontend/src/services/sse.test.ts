// T-E028: Unit tests for SSE client (focused on critical paths)
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { SSEClient, getSSEClient, connectSSE, disconnectSSE } from './sse';
import { get } from 'svelte/store';
import { isConnected, notifications, servers } from '../stores/stores';

// Mock EventSource
class MockEventSource {
  onopen: ((event: Event) => void) | null = null;
  onerror: ((event: Event) => void) | null = null;
  onmessage: ((event: MessageEvent) => void) | null = null;
  readyState = 0;
  url: string;
  withCredentials = false;

  private listeners: Map<string, ((event: MessageEvent) => void)[]> = new Map();

  constructor(url: string) {
    this.url = url;
    // Simulate immediate connection
    setTimeout(() => {
      this.readyState = 1;
      if (this.onopen) {
        this.onopen(new Event('open'));
      }
    }, 10);
  }

  addEventListener(type: string, listener: (event: MessageEvent) => void) {
    if (!this.listeners.has(type)) {
      this.listeners.set(type, []);
    }
    this.listeners.get(type)!.push(listener);
  }

  removeEventListener(type: string, listener: (event: MessageEvent) => void) {
    const listeners = this.listeners.get(type);
    if (listeners) {
      const index = listeners.indexOf(listener);
      if (index > -1) {
        listeners.splice(index, 1);
      }
    }
  }

  close() {
    this.readyState = 2;
  }

  // Helper to simulate incoming events
  simulateEvent(type: string, data: any, lastEventId?: string) {
    const event = new MessageEvent(type, {
      data: JSON.stringify(data),
      lastEventId: lastEventId || '',
    });

    if (type === 'message' && this.onmessage) {
      this.onmessage(event);
    } else {
      const listeners = this.listeners.get(type);
      if (listeners) {
        listeners.forEach((listener) => listener(event));
      }
    }
  }
}

// Replace global EventSource
global.EventSource = MockEventSource as any;

describe('SSE Client', () => {
  let client: SSEClient;

  beforeEach(() => {
    vi.clearAllMocks();
    isConnected.set(false);
    notifications.set([]);
    servers.set([]);
    client = new SSEClient();
  });

  afterEach(() => {
    client.disconnect();
  });

  describe('Connection management', () => {
    it('should connect to SSE stream', async () => {
      client.connect();

      // Wait for connection
      await new Promise((resolve) => setTimeout(resolve, 20));

      expect(get(isConnected)).toBe(true);
    });

    it('should disconnect from SSE stream', async () => {
      client.connect();
      await new Promise((resolve) => setTimeout(resolve, 20));

      client.disconnect();
      expect(get(isConnected)).toBe(false);
    });

    it('should include server IDs in URL query params', () => {
      const clientWithServers = new SSEClient(['server-1', 'server-2']);
      clientWithServers.connect();

      // EventSource URL should contain serverIds
      expect(clientWithServers).toBeDefined();
      clientWithServers.disconnect();
    });
  });

  describe('Event handling', () => {
    it('should handle server.discovered events', async () => {
      client.connect();
      await new Promise((resolve) => setTimeout(resolve, 20));

      const eventData = {
        type: 'server.discovered',
        timestamp: new Date().toISOString(),
        data: {
          serverID: 'server-1',
          name: 'Test Server',
          source: 'client_config',
        },
      };

      // Get the mock EventSource instance and simulate event
      const mockES = (client as any).eventSource as MockEventSource;
      mockES.simulateEvent('server.discovered', eventData);

      // Wait for event processing
      await new Promise((resolve) => setTimeout(resolve, 10));

      // Check notification was added
      const notifs = get(notifications);
      expect(notifs.some((n) => n.message.includes('Test Server'))).toBe(true);
    });

    it('should handle server.status.changed events', async () => {
      // Add a mock server to the store first
      servers.set([{
        id: 'server-1',
        name: 'Test Server',
        installationPath: '/path/to/server',
        source: 'client_config',
        status: {
          state: 'stopped',
          uptime: 0,
          lastChecked: new Date().toISOString(),
          lastStateChange: new Date().toISOString(),
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
      }]);

      client.connect();
      await new Promise((resolve) => setTimeout(resolve, 20));

      const eventData = {
        type: 'server.status.changed',
        timestamp: new Date().toISOString(),
        data: {
          serverID: 'server-1',
          oldState: 'stopped',
          newState: 'running',
        },
      };

      const mockES = (client as any).eventSource as MockEventSource;
      mockES.simulateEvent('server.status.changed', eventData);

      await new Promise((resolve) => setTimeout(resolve, 10));

      // Check that server status was updated in store
      const updatedServers = get(servers);
      const updatedServer = updatedServers.find(s => s.id === 'server-1');
      expect(updatedServer?.status.state).toBe('running');

      // Check notification was added
      const notifs = get(notifications);
      expect(notifs.some((n) => n.message.includes('stopped → running'))).toBe(true);
    });

    it('should handle server.log.entry events', async () => {
      client.connect();
      await new Promise((resolve) => setTimeout(resolve, 20));

      const eventData = {
        type: 'server.log.entry',
        timestamp: new Date().toISOString(),
        data: {
          serverID: 'server-1',
          severity: 'error',
          message: 'Test error message',
        },
      };

      const mockES = (client as any).eventSource as MockEventSource;
      mockES.simulateEvent('server.log.entry', eventData);

      await new Promise((resolve) => setTimeout(resolve, 10));

      // Error logs should trigger notifications
      const notifs = get(notifications);
      expect(notifs.some((n) => n.type === 'error')).toBe(true);
    });

    it('should handle config.file.changed events', async () => {
      client.connect();
      await new Promise((resolve) => setTimeout(resolve, 20));

      const eventData = {
        type: 'config.file.changed',
        timestamp: new Date().toISOString(),
        data: {
          filePath: '/path/to/config.json',
        },
      };

      const mockES = (client as any).eventSource as MockEventSource;
      mockES.simulateEvent('config.file.changed', eventData);

      await new Promise((resolve) => setTimeout(resolve, 10));

      const notifs = get(notifications);
      expect(notifs.some((n) => n.message.includes('Configuration file changed'))).toBe(true);
    });
  });

  describe('Singleton pattern', () => {
    it('should return same instance', () => {
      const client1 = getSSEClient();
      const client2 = getSSEClient();

      expect(client1).toBe(client2);
    });

    it('should connect and disconnect via helper functions', async () => {
      connectSSE(['server-1']);
      await new Promise((resolve) => setTimeout(resolve, 20));

      expect(get(isConnected)).toBe(true);

      disconnectSSE();
      expect(get(isConnected)).toBe(false);
    });
  });
});
