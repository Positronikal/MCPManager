// TypeScript type definitions matching Go backend models
// Phase E: Frontend Implementation - T-E001

export type StatusState = 'stopped' | 'starting' | 'running' | 'error';
export type LogSeverity = 'info' | 'success' | 'warning' | 'error';
export type DiscoverySource = 'client_config' | 'extension' | 'filesystem' | 'process';
export type DependencyType = 'runtime' | 'library' | 'tool' | 'environment';

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
  discoveredAt: string; // ISO 8601
  lastSeenAt: string;   // ISO 8601
  source: DiscoverySource;
}

export interface ServerStatus {
  state: StatusState;
  message: string;
  lastStateChange: string; // ISO 8601
  startupAttempts: number;
}

export interface ServerConfiguration {
  environmentVariables: Record<string, string>;
  commandLineArguments: string[];
  workingDirectory?: string;
  autoStart: boolean;
  restartOnCrash: boolean;
  maxRestartAttempts: number;
  restartDelay: number;
  clientConfigPath?: string;
  lastModified?: string; // ISO 8601
}

export interface LogEntry {
  id: string;
  timestamp: string; // ISO 8601
  severity: LogSeverity;
  source: string; // server ID
  message: string;
  metadata: Record<string, any>;
}

export interface Dependency {
  name: string;
  type: DependencyType;
  requiredVersion: string;
  detectedVersion?: string;
  installed: boolean;
  installationInstructions?: string;
}

export interface ApplicationState {
  discoveredServers: string[]; // UUIDs
  monitoredConfigPaths: string[];
  userPreferences: UserPreferences;
  windowLayout: WindowLayout;
  filters: Filters;
}

export interface UserPreferences {
  theme: string; // "dark" | "light"
  logRetentionPerServer: number;
  enableNotifications: boolean;
}

export interface WindowLayout {
  width: number;
  height: number;
  x: number;
  y: number;
  maximized: boolean;
  logPanelHeight: number;
}

export interface Filters {
  selectedServer?: string;
  selectedSeverity?: LogSeverity;
  searchQuery: string;
}

export interface ServerMetrics {
  uptimeSeconds?: number;
  memoryUsageMB?: number;
  requestCount?: number;
  cpuPercent?: number;
}

export interface UpdateInfo {
  updateAvailable: boolean;
  currentVersion: string;
  latestVersion: string;
  releaseNotes?: string;
}

export interface Notification {
  id: string;
  type: 'info' | 'success' | 'warning' | 'error';
  message: string;
  timestamp: number;
  duration?: number;
}

export interface ServerFilters {
  status?: string;
  source?: string;
  searchQuery?: string;
}
