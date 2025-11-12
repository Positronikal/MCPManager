## Phase E: Frontend Implementation (Tasks E001-E030)

**Objective**: Build Svelte UI components and integrate with backend API/SSE

### T-E001 Setup Svelte TypeScript and stores
**File**: `frontend/src/stores/stores.ts`
**Effort**: M (1-2 hours)
**Dependencies**: T-A003

**Description**:
Setup Svelte stores for state management and TypeScript configuration.

**Steps**:
1. Configure TypeScript in `frontend/tsconfig.json`:
   - Enable strict mode
   - Add type definitions for Svelte, Wails bindings
2. Create Svelte stores in `stores/stores.ts`:
   ```typescript
   import { writable } from 'svelte/store';

   export const servers = writable<MCPServer[]>([]);
   export const logs = writable<LogEntry[]>([]);
   export const appState = writable<ApplicationState>(defaultState);
   export const selectedServer = writable<string | null>(null);
   export const selectedSeverity = writable<LogSeverity | null>(null);
   ```
3. Define TypeScript interfaces matching Go models:
   - MCPServer, ServerStatus, LogEntry, ApplicationState, etc.

**Acceptance**:
- TypeScript compiles without errors
- Stores defined and exportable
- Type definitions match backend models
- `npm run check` passes (Svelte TypeScript check)

---

### T-E002 Implement API service client (REST wrapper)
**File**: `frontend/src/services/api.ts`
**Effort**: M (1-2 hours)
**Dependencies**: T-E001

**Description**:
Create TypeScript API client for calling backend REST endpoints.

**Steps**:
1. Create `api.ts`:
   ```typescript
   class APIClient {
       private baseURL = 'http://localhost:8080/api/v1';

       async listServers(statusFilter?: string): Promise<{servers: MCPServer[], count: number, lastDiscovery: string}> {
           const response = await fetch(`${this.baseURL}/servers?status=${statusFilter || ''}`);
           return response.json();
       }

       async startServer(serverId: string): Promise<void> {
           await fetch(`${this.baseURL}/servers/${serverId}/start`, { method: 'POST' });
       }

       // ... all other endpoints
   }

   export const apiClient = new APIClient();
   ```
2. Handle errors: Throw descriptive errors on non-200 responses
3. Add request/response logging for debugging

**Acceptance**:
- API client compiles and exports
- All endpoints from api-spec.yaml covered
- Error handling works (network errors, 4xx/5xx responses)
- Unit test (mock fetch): listServers() returns typed data

---

### T-E003 Implement SSE client with auto-reconnect
**File**: `frontend/src/services/sse.ts`
**Effort**: M (1-2 hours)
**Dependencies**: T-E001

**Description**:
Create SSE client for real-time events per api-spec.yaml reconnection strategy.

**Steps**:
1. Create `sse.ts`:
   ```typescript
   class SSEClient {
       private eventSource: EventSource | null = null;
       private lastEventId: string | null = null;
       private reconnectDelay = 1000; // Start at 1s, exponential backoff

       connect(onEvent: (event: Event) => void) {
           const url = `http://localhost:8080/api/v1/events`;
           this.eventSource = new EventSource(url);

           this.eventSource.onmessage = (e) => {
               this.lastEventId = e.lastEventId;
               this.reconnectDelay = 1000; // Reset backoff on success
               onEvent(JSON.parse(e.data));
           };

           this.eventSource.onerror = () => {
               this.eventSource?.close();
               setTimeout(() => this.reconnect(onEvent), this.reconnectDelay);
               this.reconnectDelay = Math.min(this.reconnectDelay * 2, 30000); // Exponential backoff, max 30s
           };
       }

       disconnect() {
           this.eventSource?.close();
       }
   }

   export const sseClient = new SSEClient();
   ```
2. Handle event types: ServerDiscovered, ServerStatusChanged, ServerLogEntry, ConfigFileChanged, ServerMetricsUpdated
3. Update Svelte stores on event received

**Acceptance**:
- SSE connection established
- Events received and parsed
- Auto-reconnect works (kill backend → restart → reconnects)
- Exponential backoff implemented
- Unit test (mock EventSource): Event triggers store update

---

### T-E004 Implement dark theme styling
**File**: `frontend/src/app.css`
**Effort**: M (1-2 hours)
**Dependencies**: T-A003

**Description**:
Implement dark theme per FR-043 using CSS variables.

**Steps**:
1. Create `app.css` with CSS variables:
   ```css
   :root {
       --bg-primary: #1e1e1e;
       --bg-secondary: #2d2d2d;
       --text-primary: #e0e0e0;
       --text-secondary: #a0a0a0;
       --border-color: #3d3d3d;
       --status-running: #4caf50;  /* Green */
       --status-stopped: #f44336;   /* Red */
       --status-starting: #2196f3;  /* Blue */
       --status-error: #ff9800;     /* Yellow/Orange */
       --log-info: #2196f3;         /* Blue */
       --log-success: #4caf50;      /* Green */
       --log-warning: #ff9800;      /* Yellow */
       --log-error: #f44336;        /* Red */
   }

   body {
       background-color: var(--bg-primary);
       color: var(--text-primary);
       font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
   }
   ```
2. Import in `App.svelte`
3. Style all components using variables

**Acceptance**:
- Dark theme applied globally
- Status colors match FR-004 (green/red/blue-gray/yellow)
- Log severity colors match FR-021
- Consistent spacing and alignment (FR-048)

---

### T-E005 Implement main application layout
**File**: `frontend/src/App.svelte`
**Effort**: M (1-2 hours)
**Dependencies**: T-E001, T-E004

**Description**:
Create main application layout: header, sidebar, content area, log viewer.

**Steps**:
1. Create `App.svelte` structure:
   ```svelte
   <script lang="ts">
       import ServerTable from './components/ServerTable.svelte';
       import LogViewer from './components/LogViewer.svelte';
       import Sidebar from './components/Sidebar.svelte';
   </script>

   <div class="app-container">
       <header>
           <h1>MCP Manager</h1>
           <button on:click={refreshDiscovery}>Refresh</button>
       </header>

       <div class="main-content">
           <Sidebar />
           <div class="content-area">
               <ServerTable />
           </div>
       </div>

       <LogViewer />
   </div>

   <style>
       .app-container { display: grid; grid-template-rows: auto 1fr auto; height: 100vh; }
       .main-content { display: grid; grid-template-columns: 200px 1fr; }
       /* ... responsive layout */
   </style>
   ```
2. Use CSS Grid for responsive layout (FR-045)
3. Log viewer fixed at bottom, resizable height (FR-020)

**Acceptance**:
- Layout renders correctly
- Responsive resizing works (FR-045)
- Log viewer at bottom (default 200px height)
- Header, sidebar, content area all visible

---

### T-E006 Implement ServerTable component
**File**: `frontend/src/components/ServerTable.svelte`
**Effort**: L (2+ hours)
**Dependencies**: T-E001, T-E002

**Description**:
Display servers in table per FR-003, with Start/Stop/Restart buttons.

**Steps**:
1. Create `ServerTable.svelte`:
   ```svelte
   <script lang="ts">
       import { servers } from '../stores/stores';
       import { apiClient } from '../services/api';

       async function handleStart(serverId: string) {
           await apiClient.startServer(serverId);
       }

       async function handleStop(serverId: string, force: boolean) {
           await apiClient.stopServer(serverId, force, 10);
       }
   </script>

   <table>
       <thead>
           <tr>
               <th>Status</th>
               <th>Name</th>
               <th>Version</th>
               <th>Capabilities</th>
               <th>PID</th>
               <th>Actions</th>
           </tr>
       </thead>
       <tbody>
           {#each $servers as server}
               <tr>
                   <td><span class="status-indicator status-{server.status.state}"></span></td>
                   <td>{server.name}</td>
                   <td>{server.version || 'N/A'}</td>
                   <td>{server.capabilities?.join(', ') || 'N/A'}</td>
                   <td>{server.pid || '-'}</td>
                   <td>
                       {#if server.status.state === 'stopped' || server.status.state === 'error'}
                           <button on:click={() => handleStart(server.id)}>Start</button>
                       {/if}
                       {#if server.status.state === 'running'}
                           <button on:click={() => handleStop(server.id, false)}>Stop</button>
                           <button on:click={() => handleRestart(server.id)}>Restart</button>
                       {/if}
                       <button on:click={() => openConfig(server.id)}>Config</button>
                       <button on:click={() => openLogs(server.id)}>Logs</button>
                   </td>
               </tr>
           {/each}
       </tbody>
   </table>

   <style>
       .status-indicator { display: inline-block; width: 12px; height: 12px; border-radius: 50%; }
       .status-running { background-color: var(--status-running); }
       .status-stopped { background-color: var(--status-stopped); }
       .status-starting { background-color: var(--status-starting); }
       .status-error { background-color: var(--status-error); }
   </style>
   ```
2. Bind to `servers` store (reactive updates)
3. Color-code status per FR-004

**Acceptance**:
- Table displays all servers from store
- Status indicators color-coded correctly
- Buttons enabled/disabled based on state
- Clicking Start calls API and updates UI

---

### T-E007 Implement real-time status updates via SSE
**File**: Modify `App.svelte` and stores
**Effort**: M (1-2 hours)
**Dependencies**: T-E003, T-E006

**Description**:
Connect SSE client to update server table in real-time per FR-005, FR-047.

**Steps**:
1. In `App.svelte` onMount:
   ```typescript
   import { sseClient } from './services/sse';
   import { servers } from './stores/stores';

   onMount(() => {
       sseClient.connect((event) => {
           if (event.type === 'ServerStatusChanged') {
               servers.update(list => {
                   const index = list.findIndex(s => s.id === event.data.serverId);
                   if (index !== -1) {
                       list[index].status.state = event.data.newState;
                       list[index].pid = event.data.pid;
                   }
                   return list;
               });
           }
           // Handle other event types
       });

       return () => sseClient.disconnect();
   });
   ```
2. Update stores on ServerDiscovered, ServerStatusChanged, ServerMetricsUpdated events

**Acceptance**:
- Server status updates in real-time (no manual refresh)
- Start button click → status changes to starting → running
- Backend crash → status updates to error automatically
- SSE reconnects if connection lost

---

### T-E008 Implement LogViewer component
**File**: `frontend/src/components/LogViewer.svelte`
**Effort**: L (2+ hours)
**Dependencies**: T-E001, T-E002

**Description**:
Display real-time logs at bottom of window per FR-020 through FR-023.

**Steps**:
1. Create `LogViewer.svelte`:
   ```svelte
   <script lang="ts">
       import { logs, selectedServer, selectedSeverity } from '../stores/stores';

       let searchQuery = '';

       $: filteredLogs = $logs.filter(log => {
           if ($selectedServer && log.source !== $selectedServer) return false;
           if ($selectedSeverity && log.severity !== $selectedSeverity) return false;
           if (searchQuery && !log.message.toLowerCase().includes(searchQuery.toLowerCase())) return false;
           return true;
       });
   </script>

   <div class="log-viewer">
       <div class="log-toolbar">
           <select bind:value={$selectedServer}>
               <option value={null}>All Servers</option>
               {#each $servers as server}
                   <option value={server.id}>{server.name}</option>
               {/each}
           </select>

           <select bind:value={$selectedSeverity}>
               <option value={null}>All Severities</option>
               <option value="info">INFO</option>
               <option value="success">SUCCESS</option>
               <option value="warning">WARNING</option>
               <option value="error">ERROR</option>
           </select>

           <input type="text" bind:value={searchQuery} placeholder="Search logs..." />
       </div>

       <div class="log-entries">
           {#each filteredLogs as log}
               <div class="log-entry log-{log.severity}">
                   <span class="log-timestamp">{formatTimestamp(log.timestamp)}</span>
                   <span class="log-source">[{log.source}]</span>
                   <span class="log-message">{log.message}</span>
               </div>
           {/each}
       </div>
   </div>

   <style>
       .log-viewer { height: 200px; border-top: 1px solid var(--border-color); overflow-y: auto; }
       .log-info { color: var(--log-info); }
       .log-success { color: var(--log-success); }
       .log-warning { color: var(--log-warning); }
       .log-error { color: var(--log-error); }
   </style>
   ```
2. Implement filtering: server, severity, search (FR-022, FR-023)
3. Color-code by severity (FR-021)
4. Auto-scroll to bottom on new log entry

**Acceptance**:
- Logs display color-coded by severity
- Filter by server dropdown works
- Filter by severity dropdown works
- Search box filters logs in real-time
- Auto-scrolls to newest entry

---

### T-E009 Connect LogViewer to SSE for real-time logs
**File**: Modify `App.svelte` SSE handler
**Effort**: M (1-2 hours)
**Dependencies**: T-E003, T-E008

**Description**:
Update logs store on ServerLogEntry SSE events for real-time log streaming.

**Steps**:
1. In SSE event handler:
   ```typescript
   if (event.type === 'ServerLogEntry') {
       logs.update(list => {
           list.push({
               id: event.id,
               timestamp: event.timestamp,
               severity: event.data.severity,
               source: event.data.serverId,
               message: event.data.message,
               metadata: {}
           });

           // Enforce 1000 entry limit per server (client-side circular buffer)
           // Count entries per source, remove oldest if > 1000
           const sourceEntries = list.filter(l => l.source === event.data.serverId);
           if (sourceEntries.length > 1000) {
               const oldestId = sourceEntries[0].id;
               list = list.filter(l => l.id !== oldestId);
           }

           return list;
       });
   }
   ```

**Acceptance**:
- New log entries appear in real-time
- No page refresh needed
- Client-side buffer enforces 1000 entry limit
- UI performance remains smooth with 50k entries

---

### T-E010 Implement ConfigurationEditor component
**File**: `frontend/src/components/ConfigurationEditor.svelte`
**Effort**: L (2+ hours)
**Dependencies**: T-E001, T-E002

**Description**:
Configuration editor modal/panel per FR-014 through FR-018.

**Steps**:
1. Create `ConfigurationEditor.svelte`:
   ```svelte
   <script lang="ts">
       import { apiClient } from '../services/api';

       export let serverId: string;
       export let onClose: () => void;

       let config: ServerConfiguration;
       let errors: string[] = [];

       onMount(async () => {
           config = await apiClient.getConfiguration(serverId);
       });

       async function saveConfig() {
           try {
               errors = [];
               await apiClient.updateConfiguration(serverId, config);
               alert('Configuration saved successfully');
               onClose();
           } catch (err) {
               errors = [err.message];
           }
       }
   </script>

   <div class="modal">
       <div class="modal-content">
           <h2>Server Configuration</h2>

           <label>
               Environment Variables:
               <table>
                   <thead><tr><th>Key</th><th>Value</th><th></th></tr></thead>
                   <tbody>
                       {#each Object.entries(config.environmentVariables || {}) as [key, value]}
                           <tr>
                               <td><input bind:value={key} /></td>
                               <td><input bind:value={value} /></td>
                               <td><button on:click={() => deleteEnvVar(key)}>Delete</button></td>
                           </tr>
                       {/each}
                   </tbody>
               </table>
               <button on:click={addEnvVar}>Add Variable</button>
           </label>

           <label>
               Command-Line Arguments:
               <textarea bind:value={argsText}></textarea>
           </label>

           <label>
               <input type="checkbox" bind:checked={config.autoStart} />
               Auto-start on launch
           </label>

           <label>
               <input type="checkbox" bind:checked={config.restartOnCrash} />
               Restart on crash
           </label>

           {#if errors.length > 0}
               <div class="errors">
                   {#each errors as error}
                       <p class="error">{error}</p>
                   {/each}
               </div>
           {/if}

           <button on:click={saveConfig}>Save</button>
           <button on:click={onClose}>Cancel</button>
       </div>
   </div>
   ```
2. Validate inputs client-side before submit
3. Display validation errors from API (400 responses)
4. Show read-only client config section (FR-019 - display only, no editing)

**Acceptance**:
- Modal opens when Config button clicked
- Loads current configuration
- Environment variables table editable
- Validation errors displayed
- Save persists changes (verified via reload)
- Client config section clearly marked read-only

---

### T-E011 Implement DetailedLogsView component
**File**: `frontend/src/components/DetailedLogsView.svelte`
**Effort**: M (1-2 hours)
**Dependencies**: T-E001, T-E002

**Description**:
Server-specific detailed log view modal per FR-024.

**Steps**:
1. Create `DetailedLogsView.svelte`:
   ```svelte
   <script lang="ts">
       import { apiClient } from '../services/api';

       export let serverId: string;
       export let onClose: () => void;

       let logs: LogEntry[] = [];
       let severityFilter: LogSeverity | null = null;
       let searchQuery = '';

       onMount(async () => {
           await loadLogs();
       });

       async function loadLogs() {
           const response = await apiClient.getServerLogs(serverId, severityFilter, 1000, 0);
           logs = response.logs;
       }

       $: filteredLogs = logs.filter(log => {
           if (severityFilter && log.severity !== severityFilter) return false;
           if (searchQuery && !log.message.toLowerCase().includes(searchQuery.toLowerCase())) return false;
           return true;
       });
   </script>

   <div class="modal">
       <div class="modal-content detailed-logs">
           <h2>Detailed Logs - {serverName}</h2>

           <div class="log-toolbar">
               <select bind:value={severityFilter} on:change={loadLogs}>
                   <option value={null}>All Severities</option>
                   <option value="info">INFO</option>
                   <option value="success">SUCCESS</option>
                   <option value="warning">WARNING</option>
                   <option value="error">ERROR</option>
               </select>

               <input type="text" bind:value={searchQuery} placeholder="Search logs..." />
           </div>

           <div class="log-entries">
               {#each filteredLogs as log}
                   <div class="log-entry log-{log.severity}">
                       <span class="log-timestamp">{log.timestamp}</span>
                       <span class="log-severity">[{log.severity.toUpperCase()}]</span>
                       <span class="log-message">{log.message}</span>
                   </div>
               {/each}
           </div>

           <button on:click={onClose}>Close</button>
       </div>
   </div>
   ```
2. Load up to 1000 most recent entries (FR-053 limit)
3. Client-side filtering and search

**Acceptance**:
- Modal opens when Logs button clicked
- Displays up to 1000 entries
- Color-coded by severity
- Search and filter work
- Scrollable list

---

### T-E012 Implement Sidebar component with utilities
**File**: `frontend/src/components/Sidebar.svelte`
**Effort**: M (1-2 hours)
**Dependencies**: T-E001

**Description**:
Sidebar with utility buttons per FR-030 through FR-034.

**Steps**:
1. Create `Sidebar.svelte`:
   ```svelte
   <script lang="ts">
       let activeView = 'servers'; // servers, netstat, shell, explorer, services, help
   </script>

   <aside class="sidebar">
       <button class:active={activeView === 'servers'} on:click={() => activeView = 'servers'}>
           Servers
       </button>
       <button class:active={activeView === 'netstat'} on:click={() => activeView = 'netstat'}>
           Netstat
       </button>
       <button class:active={activeView === 'shell'} on:click={() => activeView = 'shell'}>
           Shell
       </button>
       <button class:active={activeView === 'explorer'} on:click={() => activeView = 'explorer'}>
           Explorer
       </button>
       <button class:active={activeView === 'services'} on:click={() => activeView = 'services'}>
           Services
       </button>
       <button class:active={activeView === 'help'} on:click={() => activeView = 'help'}>
           Help
       </button>
   </aside>

   <style>
       .sidebar { display: flex; flex-direction: column; background-color: var(--bg-secondary); padding: 10px; }
       .sidebar button { margin-bottom: 10px; text-align: left; }
       .sidebar button.active { background-color: var(--status-running); }
   </style>
   ```
2. Clicking utility buttons switches main content area view

**Acceptance**:
- Sidebar displays all utility buttons
- Active button highlighted
- Clicking button changes content area (placeholder views OK for now)

---

### T-E013 Implement Netstat utility view
**File**: `frontend/src/components/NetstatView.svelte`
**Effort**: M (1-2 hours)
**Dependencies**: T-E001, T-E012

**Description**:
Display network connections per FR-030 by calling backend API.

**Steps**:
1. Add backend API method: `GET /api/v1/netstat?pids=<comma-separated>`
   - Backend: Parse `netstat` output, filter by PIDs of running MCP servers
   - Return: [{protocol, localAddress, remoteAddress, state, pid}]
2. Create `NetstatView.svelte`:
   ```svelte
   <script lang="ts">
       import { apiClient } from '../services/api';

       let connections: NetstatEntry[] = [];

       onMount(async () => {
           const pids = $servers.filter(s => s.pid).map(s => s.pid).join(',');
           connections = await apiClient.getNetstat(pids);
       });
   </script>

   <div class="netstat-view">
       <h2>Network Connections (MCP Servers)</h2>
       <table>
           <thead>
               <tr><th>Protocol</th><th>Local</th><th>Remote</th><th>State</th><th>PID</th></tr>
           </thead>
           <tbody>
               {#each connections as conn}
                   <tr>
                       <td>{conn.protocol}</td>
                       <td>{conn.localAddress}</td>
                       <td>{conn.remoteAddress}</td>
                       <td>{conn.state}</td>
                       <td>{conn.pid}</td>
                   </tr>
               {/each}
           </tbody>
       </table>
   </div>
   ```

**Acceptance**:
- Netstat view displays connections for running MCP servers
- Table shows protocol, addresses, state, PID
- Works on Windows, macOS, Linux

---

### T-E014 Implement Shell utility view
**File**: `frontend/src/components/ShellView.svelte`
**Effort**: S (<1 hour)
**Dependencies**: T-E001, T-E012

**Description**:
Launch platform shell per FR-031 via backend API.

**Steps**:
1. Add backend API method: `POST /api/v1/shell`
   - Backend: Launch platform shell (cmd.exe, Terminal.app, xterm) via os/exec
   - Return: {success: bool, message: string}
2. Create `ShellView.svelte`:
   ```svelte
   <script lang="ts">
       import { apiClient } from '../services/api';

       async function openShell() {
           await apiClient.launchShell();
           alert('Shell opened');
       }
   </script>

   <div class="shell-view">
       <h2>Quick Shell Access</h2>
       <p>Launch a platform-appropriate shell for quick terminal access.</p>
       <button on:click={openShell}>Open Shell</button>
   </div>
   ```

**Acceptance**:
- Button click launches shell externally
- Correct shell per platform (cmd.exe on Windows, etc.)

---

### T-E015 Implement Explorer utility view
**File**: `frontend/src/components/ExplorerView.svelte`
**Effort**: S (<1 hour)
**Dependencies**: T-E001, T-E012

**Description**:
Open server installation directories per FR-032 via backend API.

**Steps**:
1. Add backend API method: `POST /api/v1/explorer?path=<path>`
   - Backend: Launch file explorer (explorer, open, xdg-open) with path
2. Create `ExplorerView.svelte`:
   ```svelte
   <script lang="ts">
       import { servers } from '../stores/stores';
       import { apiClient } from '../services/api';

       async function openInExplorer(path: string) {
           await apiClient.openExplorer(path);
       }
   </script>

   <div class="explorer-view">
       <h2>Server Installation Directories</h2>
       <ul>
           {#each $servers as server}
               <li>
                   {server.name}: {server.installationPath}
                   <button on:click={() => openInExplorer(server.installationPath)}>Open</button>
               </li>
           {/each}
       </ul>
   </div>
   ```

**Acceptance**:
- Lists all server installation paths
- Open button launches file explorer at path
- Works on all platforms

---

### T-E016 Implement Services utility view
**File**: `frontend/src/components/ServicesView.svelte`
**Effort**: M (1-2 hours)
**Dependencies**: T-E001, T-E012

**Description**:
View system service management per FR-033 via backend API.

**Steps**:
1. Add backend API method: `GET /api/v1/services`
   - Backend: Run platform service commands (sc query, launchctl list, systemctl list-units)
   - Return: [{name, status, description}]
2. Create `ServicesView.svelte`:
   ```svelte
   <script lang="ts">
       import { apiClient } from '../services/api';

       let services: Service[] = [];

       onMount(async () => {
           services = await apiClient.getServices();
       });
   </script>

   <div class="services-view">
       <h2>System Services</h2>
       <table>
           <thead><tr><th>Name</th><th>Status</th><th>Description</th></tr></thead>
           <tbody>
               {#each services as service}
                   <tr>
                       <td>{service.name}</td>
                       <td>{service.status}</td>
                       <td>{service.description}</td>
                   </tr>
               {/each}
           </tbody>
       </table>
   </div>
   ```

**Acceptance**:
- Lists system services
- Shows status (running/stopped)
- Works on all platforms

---

### T-E017 Implement Help utility view
**File**: `frontend/src/components/HelpView.svelte`
**Effort**: M (1-2 hours)
**Dependencies**: T-E001, T-E012

**Description**:
Display embedded documentation per FR-034.

**Steps**:
1. Create `HelpView.svelte`:
   ```svelte
   <script lang="ts">
       let activeTab = 'quickstart'; // quickstart, keyboard-shortcuts, about
   </script>

   <div class="help-view">
       <div class="help-tabs">
           <button class:active={activeTab === 'quickstart'} on:click={() => activeTab = 'quickstart'}>
               Quickstart
           </button>
           <button class:active={activeTab === 'keyboard-shortcuts'} on:click={() => activeTab = 'keyboard-shortcuts'}>
               Keyboard Shortcuts
           </button>
           <button class:active={activeTab === 'about'} on:click={() => activeTab = 'about'}>
               About
           </button>
       </div>

       <div class="help-content">
           {#if activeTab === 'quickstart'}
               <h2>Quickstart Guide</h2>
               <p>Getting started with MCP Manager...</p>
               <!-- Embed quickstart.md content or simplified version -->
           {:else if activeTab === 'keyboard-shortcuts'}
               <h2>Keyboard Shortcuts</h2>
               <table>
                   <tr><td>Ctrl/Cmd+S</td><td>Start selected server</td></tr>
                   <tr><td>Ctrl/Cmd+T</td><td>Stop selected server</td></tr>
                   <tr><td>Ctrl/Cmd+R</td><td>Restart selected server</td></tr>
                   <tr><td>F5</td><td>Refresh discovery</td></tr>
                   <tr><td>Ctrl/Cmd+F</td><td>Focus search</td></tr>
                   <tr><td>Ctrl/Cmd+L</td><td>Toggle logs panel</td></tr>
               </table>
           {:else if activeTab === 'about'}
               <h2>About MCP Manager</h2>
               <p>Version 1.0.0</p>
               <p>Cross-platform desktop application for managing MCP servers.</p>
               <p>© 2025 Your Organization</p>
           {/if}
       </div>
   </div>
   ```

**Acceptance**:
- Help view displays documentation tabs
- Quickstart guide readable
- Keyboard shortcuts listed (per research.md §14)
- About section shows version

---

### T-E018 Implement keyboard shortcuts
**File**: `frontend/src/App.svelte` global key handler
**Effort**: M (1-2 hours)
**Dependencies**: T-E001

**Description**:
Global keyboard shortcuts per FR-046 and research.md §14.

**Steps**:
1. In `App.svelte`, add global keydown handler:
   ```typescript
   function handleKeydown(event: KeyboardEvent) {
       const isMac = navigator.platform.toUpperCase().indexOf('MAC') >= 0;
       const modKey = isMac ? event.metaKey : event.ctrlKey;

       if (!modKey) return;

       switch (event.key.toLowerCase()) {
           case 's': // Start server
               event.preventDefault();
               if ($selectedServer) handleStart($selectedServer);
               break;
           case 't': // Stop server
               event.preventDefault();
               if ($selectedServer) handleStop($selectedServer);
               break;
           case 'r': // Restart server
               event.preventDefault();
               if ($selectedServer) handleRestart($selectedServer);
               break;
           case 'f': // Focus search
               event.preventDefault();
               document.querySelector('input[type="text"]')?.focus();
               break;
           case 'l': // Toggle logs
               event.preventDefault();
               toggleLogPanel();
               break;
       }

       if (event.key === 'F5') { // Refresh discovery
           event.preventDefault();
           refreshDiscovery();
       }
   }

   onMount(() => {
       window.addEventListener('keydown', handleKeydown);
       return () => window.removeEventListener('keydown', handleKeydown);
   });
   ```
2. Disable shortcuts when input fields focused

**Acceptance**:
- Keyboard shortcuts work per research.md §14 table
- Platform-aware (Cmd on macOS, Ctrl on Windows/Linux)
- Shortcuts disabled when typing in input fields
- Visual indicators (tooltips show shortcuts)

---

### T-E019 Implement config file change notifications
**File**: `frontend/src/App.svelte` SSE handler
**Effort**: M (1-2 hours)
**Dependencies**: T-E003

**Description**:
Display notification when external config changes detected per clarification (hybrid approach).

**Steps**:
1. In SSE event handler:
   ```typescript
   if (event.type === 'ConfigFileChanged') {
       showNotification({
           message: `Configuration file changed: ${event.data.filePath}. Refresh discovery?`,
           actions: [
               { label: 'Refresh', callback: () => refreshDiscovery() },
               { label: 'Dismiss', callback: () => {} }
           ]
       });
   }
   ```
2. Create `Notification.svelte` component:
   ```svelte
   <script lang="ts">
       export let message: string;
       export let actions: {label: string, callback: () => void}[];
   </script>

   <div class="notification">
       <p>{message}</p>
       <div class="notification-actions">
           {#each actions as action}
               <button on:click={action.callback}>{action.label}</button>
           {/each}
       </div>
   </div>

   <style>
       .notification { position: fixed; top: 20px; right: 20px; background: var(--bg-secondary); border: 1px solid var(--border-color); padding: 15px; border-radius: 5px; z-index: 1000; }
   </style>
   ```

**Acceptance**:
- Notification appears when config file changed
- User can click Refresh or Dismiss
- Refresh triggers discovery scan
- Dismiss closes notification

---

### T-E020 Implement single-instance window activation
**File**: Backend `cmd/mcpmanager/main.go`, Wails config
**Effort**: M (1-2 hours)
**Dependencies**: T-A008, T-E001

**Description**:
Bring existing window to foreground when second instance launched per FR-051.

**Steps**:
1. In `main.go`:
   ```go
   func main() {
       singleton := platform.NewSingleInstance()
       acquired, err := singleton.Acquire()
       if !acquired {
           // Signal existing instance to show window
           // On Windows: Find window by title, call SetForegroundWindow
           // On Unix: Send signal to existing process
           return
       }
       defer singleton.Release()

       // ... start Wails app
   }
   ```
2. Add window show handler in Wails app (listen for signal)

**Acceptance**:
- Second launch brings first window to foreground
- No duplicate processes
- Verified on Windows, macOS, Linux

---

### T-E021 Implement responsive window resizing
**File**: `frontend/src/app.css`, component styles
**Effort**: M (1-2 hours)
**Dependencies**: T-E005

**Description**:
Responsive layout per FR-045.

**Steps**:
1. Use CSS Grid with fr units for flexible sizing:
   ```css
   .app-container {
       display: grid;
       grid-template-rows: auto 1fr auto;
       height: 100vh;
   }

   .main-content {
       display: grid;
       grid-template-columns: 200px 1fr;
   }

   @media (max-width: 1024px) {
       .main-content {
           grid-template-columns: 1fr; /* Stack sidebar on narrow screens */
       }
   }
   ```
2. Resizable log panel:
   - Add drag handle between content and log viewer
   - Save height to ApplicationState.windowLayout.logPanelHeight

**Acceptance**:
- Window resizes smoothly
- Layout reflows without breaking
- Log panel height adjustable via drag
- Saved log panel height persists across restarts

---

### T-E022 Implement window state persistence
**File**: `frontend/src/App.svelte`
**Effort**: M (1-2 hours)
**Dependencies**: T-E001, T-E002

**Description**:
Persist window size/position per FR-041 ApplicationState.windowLayout.

**Steps**:
1. On window resize/move:
   ```typescript
   import { appState } from './stores/stores';
   import { apiClient } from './services/api';

   function saveWindowLayout() {
       const layout = {
           width: window.innerWidth,
           height: window.innerHeight,
           x: window.screenX,
           y: window.screenY,
           maximized: window.outerWidth === screen.width && window.outerHeight === screen.height,
           logPanelHeight: $appState.windowLayout.logPanelHeight
       };

       appState.update(state => ({ ...state, windowLayout: layout }));
       apiClient.updateApplicationState($appState); // Debounced auto-save on backend
   }

   window.addEventListener('resize', saveWindowLayout);
   window.addEventListener('beforeunload', saveWindowLayout);
   ```
2. On app load, restore window layout from ApplicationState

**Acceptance**:
- Window size/position saved on close
- Restored on next launch
- Maximized state preserved

---

### T-E023 Implement log severity color coding
**File**: `frontend/src/components/LogViewer.svelte` styles
**Effort**: S (<1 hour)
**Dependencies**: T-E008

**Description**:
Color-code log entries per FR-021.

**Steps**:
1. Add CSS classes per severity:
   ```css
   .log-entry.log-info { color: var(--log-info); } /* Blue */
   .log-entry.log-success { color: var(--log-success); } /* Green */
   .log-entry.log-warning { color: var(--log-warning); } /* Yellow */
   .log-entry.log-error { color: var(--log-error); } /* Red */
   ```
2. Apply class dynamically: `<div class="log-entry log-{log.severity}">`

**Acceptance**:
- INFO logs blue
- SUCCESS logs green
- WARNING logs yellow
- ERROR logs red
- Colors consistent across LogViewer and DetailedLogsView

---

### T-E024 Implement status indicator color coding
**File**: `frontend/src/components/ServerTable.svelte` styles
**Effort**: S (<1 hour)
**Dependencies**: T-E006

**Description**:
Color-code server status per FR-004.

**Steps**:
1. Add CSS classes per status:
   ```css
   .status-indicator.status-stopped { background-color: var(--status-stopped); } /* Red */
   .status-indicator.status-starting { background-color: var(--status-starting); } /* Blue/Gray */
   .status-indicator.status-running { background-color: var(--status-running); } /* Green */
   .status-indicator.status-error { background-color: var(--status-error); } /* Yellow */
   ```
2. Apply class dynamically: `<span class="status-indicator status-{server.status.state}"></span>`

**Acceptance**:
- Stopped: red
- Starting: blue/gray
- Running: green
- Error: yellow
- Color mapping matches data-model.md §2

---

### T-E025 Implement UI responsiveness optimization
**File**: All Svelte components
**Effort**: M (1-2 hours)
**Dependencies**: T-E006 through T-E024

**Description**:
Optimize for <200ms UI response per FR-038, performance validation tests.

**Steps**:
1. Debounce search input (300ms delay):
   ```typescript
   let searchQuery = '';
   let debouncedSearch = '';

   $: {
       clearTimeout(searchTimeout);
       searchTimeout = setTimeout(() => {
           debouncedSearch = searchQuery;
       }, 300);
   }
   ```
2. Virtualize long log lists (render only visible entries):
   - Use `svelte-virtual-list` or similar for 1000+ entry lists
3. Memoize expensive computations:
   ```typescript
   import { derived } from 'svelte/store';

   const filteredLogs = derived([logs, selectedServer, selectedSeverity], ([$logs, $selectedServer, $selectedSeverity]) => {
       return $logs.filter(log => {
           // ... filtering logic
       });
   });
   ```
4. Profile with Chrome DevTools, optimize hot paths

**Acceptance**:
- Button clicks respond within 200ms (per quickstart.md performance test)
- Log filtering <50ms for 50k entries
- Search <300ms
- UI remains smooth during background operations

---

### T-E026 Write frontend unit tests for stores
**File**: `frontend/tests/stores.test.ts`
**Effort**: M (1-2 hours)
**Dependencies**: T-E001

**Description**:
Unit tests for Svelte stores using Vitest.

**Test Cases**:
1. servers store: Update works, reactive
2. logs store: Add entry, enforce 1000 limit per server
3. appState store: Defaults set correctly
4. selectedServer store: Null by default, updates
5. Derived stores compute correctly

**Acceptance**:
- All store tests pass
- Coverage > 80% for stores.ts
- Tests use Vitest or similar

---

### T-E027 Write frontend unit tests for API client
**File**: `frontend/tests/api.test.ts`
**Effort**: M (1-2 hours)
**Dependencies**: T-E002

**Description**:
Unit tests for API client with mocked fetch.

**Test Cases**:
1. listServers() returns typed data
2. startServer() sends POST with correct URL
3. Error handling: 404 throws descriptive error
4. Error handling: Network failure throws error
5. Query parameters encoded correctly

**Acceptance**:
- All API client tests pass
- Uses fetch mock (msw or vitest.mock)
- Coverage > 80% for api.ts

---

### T-E028 Write frontend unit tests for SSE client
**File**: `frontend/tests/sse.test.ts`
**Effort**: M (1-2 hours)
**Dependencies**: T-E003

**Description**:
Unit tests for SSE client with mocked EventSource.

**Test Cases**:
1. connect() establishes EventSource
2. onmessage handler calls callback with parsed event
3. onerror triggers reconnect with exponential backoff
4. disconnect() closes EventSource
5. Reconnect includes Last-Event-ID header

**Acceptance**:
- All SSE client tests pass
- Uses EventSource mock
- Coverage > 80% for sse.ts

---

### T-E029 Write frontend component tests for ServerTable
**File**: `frontend/tests/ServerTable.test.ts`
**Effort**: M (1-2 hours)
**Dependencies**: T-E006

**Description**:
Component tests for ServerTable using Svelte Testing Library.

**Test Cases**:
1. Renders list of servers from store
2. Start button enabled when status=stopped
3. Stop button enabled when status=running
4. Clicking Start calls apiClient.startServer()
5. Status color indicator matches server.status.state

**Acceptance**:
- All component tests pass
- Uses @testing-library/svelte
- Coverage > 70% for ServerTable.svelte

---

### T-E030 Write frontend component tests for LogViewer
**File**: `frontend/tests/LogViewer.test.ts`
**Effort**: M (1-2 hours)
**Dependencies**: T-E008

**Description**:
Component tests for LogViewer using Svelte Testing Library.

**Test Cases**:
1. Renders logs from store
2. Filter by server dropdown works
3. Filter by severity dropdown works
4. Search input filters logs
5. Color coding applied correctly

**Acceptance**:
- All component tests pass
- Uses @testing-library/svelte
- Coverage > 70% for LogViewer.svelte

---

