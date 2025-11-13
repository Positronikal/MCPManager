// Test setup file for Vitest
import '@testing-library/jest-dom';
import { vi } from 'vitest';

// Mock Wails runtime functions
vi.mock('../../wailsjs/runtime/runtime', () => ({
  EventsOn: vi.fn(),
  EventsOff: vi.fn(),
  EventsEmit: vi.fn(),
  BrowserOpenURL: vi.fn(),
}));

// Mock Wails Go bindings
vi.mock('../../wailsjs/go/main/App', () => ({
  ListServers: vi.fn(),
  DiscoverServers: vi.fn(),
  GetServer: vi.fn(),
  StartServer: vi.fn(),
  StopServer: vi.fn(),
  RestartServer: vi.fn(),
  GetServerStatus: vi.fn(),
  GetConfiguration: vi.fn(),
  UpdateConfiguration: vi.fn(),
  GetLogs: vi.fn(),
  GetAllLogs: vi.fn(),
  GetMetrics: vi.fn(),
  GetDependencies: vi.fn(),
  GetUpdates: vi.fn(),
  GetApplicationState: vi.fn(),
  UpdateApplicationState: vi.fn(),
}));

// Mock window.navigator.clipboard
Object.assign(navigator, {
  clipboard: {
    writeText: vi.fn(() => Promise.resolve()),
  },
});

// Suppress console errors in tests unless debugging
if (!process.env.DEBUG) {
  global.console.error = vi.fn();
  global.console.warn = vi.fn();
}
