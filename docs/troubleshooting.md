Here are my testing results:

1. DevTools console output:

```
events.ts:74 Wails event listeners initialized
events.ts:30 Server discovered: Object
events.ts:30 Server discovered: Object
events.ts:30 Server discovered: Object
:34115/favicon.ico:1   Failed to load resource: the server responded with a status of 404 (Not Found)
events.ts:68 Servers discovered: Array(3)
ServerTable.svelte:47 [UI] handleStop called {serverId: '73054ce2-68a8-40ed-9bbf-ffb75e521d53', name: 'Filesystem', force: false, timeout: 10}
ServerTable.svelte:52 [UI] Calling api.lifecycle.stopServer...
ServerTable.svelte:57  [UI] Failed to stop server - Full error object: server not found: 73054ce2-68a8-40ed-9bbf-ffb75e521d53
handleStop @ ServerTable.svelte:57
await in handleStop
click_handler_1 @ ServerTable.svelte:248
click_handler_1 @ ServerTable.svelte:260
ServerTable.svelte:58  [UI] Error type: string
handleStop @ ServerTable.svelte:58
await in handleStop
click_handler_1 @ ServerTable.svelte:248
click_handler_1 @ ServerTable.svelte:260
ServerTable.svelte:59  [UI] Error message: undefined
handleStop @ ServerTable.svelte:59
await in handleStop
click_handler_1 @ ServerTable.svelte:248
click_handler_1 @ ServerTable.svelte:260
ServerTable.svelte:60  [UI] Error string: server not found: 73054ce2-68a8-40ed-9bbf-ffb75e521d53
handleStop @ ServerTable.svelte:60
await in handleStop
click_handler_1 @ ServerTable.svelte:248
click_handler_1 @ ServerTable.svelte:260
```

2. mcpmanager terminal output:

```
PS D:\dev\ARTIFICIAL_INTELLIGENCE\MCP\MCPManager> wails dev
Wails CLI v2.10.2

Executing: go mod tidy
  • Generating bindings: 2025/10/22 22:34:23 KnownStructs: dependencies.UpdateInfo      main.DiscoverServersResponse   main.GetDependenciesResponse     main.GetLogsResponse    main.ListServersResponse        main.ServerOperationResponse   main.UpdateApplicationStateResponse      models.ApplicationState models.Dependency       models.Filters  models.LogEntrymodels.MCPServer models.ServerConfiguration      models.ServerMetrics    models.ServerStatus     models.UserPreferences models.WindowLayout
Not found: time.Time
KnownStructs: dependencies.UpdateInfo   main.DiscoverServersResponse    main.GetDependenciesResponse    main.GetLogsResponse    main.ListServersResponse        main.ServerOperationResponse    main.UpdateApplicationStateResponse     models.ApplicationState models.Dependency       models.Filters  models.LogEntry models.MCPServer        models.ServerConfiguration      models.ServerMetrics    models.ServerStatus     models.UserPreferences  models.WindowLayout
Not found: time.Time
KnownStructs: dependencies.UpdateInfo   main.DiscoverServersResponse    main.GetDependenciesResponse    main.GetLogsResponse    main.ListServersResponse        main.ServerOperationResponse    main.UpdateApplicationStateResponse     models.ApplicationState models.Dependency       models.Filters  models.LogEntry models.MCPServer        models.ServerConfiguration      models.ServerMetrics    models.ServerStatus     models.UserPreferences  models.WindowLayout
Not found: time.Time
KnownStructs: dependencies.UpdateInfo   main.DiscoverServersResponse    main.GetDependenciesResponse    main.GetLogsResponse    main.ListServersResponse        main.ServerOperationResponse    main.UpdateApplicationStateResponse     models.ApplicationState models.Dependency       models.Filters  models.LogEntry models.MCPServer        models.ServerConfiguration      models.ServerMetrics    models.ServerStatus     models.UserPreferences  models.WindowLayout
Not found: time.Time
KnownStructs: dependencies.UpdateInfo   main.DiscoverServersResponse    main.GetDependenciesResponse    main.GetLogsResponse    main.ListServersResponse        main.ServerOperationResponse    main.UpdateApplicationStateResponse     models.ApplicationState models.Dependency       models.Filters  models.LogEntry models.MCPServer        models.ServerConfiguration      models.ServerMetrics    models.ServerStatus     models.UserPreferences  models.WindowLayout
Not found: time.Time
KnownStructs: dependencies.UpdateInfo   main.DiscoverServersResponse    main.GetDependenciesResponse    main.GetLogsResponse    main.ListServersResponse        main.ServerOperationResponse    main.UpdateApplicationStateResponse     models.ApplicationState models.Dependency       models.Filters  models.LogEntry models.MCPServer        models.ServerConfiguration      models.ServerMetrics    models.ServerStatus     models.UserPreferences  models.WindowLayout
Not found: time.Time
KnownStructs: dependencies.UpdateInfo   main.DiscoverServersResponse    main.GetDependenciesResponse    main.GetLogsResponse    main.ListServersResponse        main.ServerOperationResponse    main.UpdateApplicationStateResponse     models.ApplicationState models.Dependency       models.Filters  models.LogEntry models.MCPServer        models.ServerConfiguration      models.ServerMetrics    models.ServerStatus     models.UserPreferences  models.WindowLayout
Not found: time.Time

Done.
  • Installing frontend dependencies: Done.
  • Compiling frontend: Done.

> frontend@0.0.0 dev
> vite


  VITE v3.2.11  ready in 635 ms

Vite Server URL: http://localhost:5173/
  ➜  Local:   http://localhost:5173/
Running frontend DevWatcher command: 'npm run dev'  ➜  Network: use --host to expose

Building application for development...
  • Generating bindings: 2025/10/22 22:34:33 KnownStructs: dependencies.UpdateInfo      main.DiscoverServersResponse   main.GetDependenciesResponse     main.GetLogsResponse    main.ListServersResponse        main.ServerOperationResponse   main.UpdateApplicationStateResponse      models.ApplicationState models.Dependency       models.Filters  models.LogEntrymodels.MCPServer models.ServerConfiguration      models.ServerMetrics    models.ServerStatus     models.UserPreferences models.WindowLayout
Not found: time.Time
KnownStructs: dependencies.UpdateInfo   main.DiscoverServersResponse    main.GetDependenciesResponse    main.GetLogsResponse    main.ListServersResponse        main.ServerOperationResponse    main.UpdateApplicationStateResponse     models.ApplicationState models.Dependency       models.Filters  models.LogEntry models.MCPServer        models.ServerConfiguration      models.ServerMetrics    models.ServerStatus     models.UserPreferences  models.WindowLayout
Not found: time.Time
KnownStructs: dependencies.UpdateInfo   main.DiscoverServersResponse    main.GetDependenciesResponse    main.GetLogsResponse    main.ListServersResponse        main.ServerOperationResponse    main.UpdateApplicationStateResponse     models.ApplicationState models.Dependency       models.Filters  models.LogEntry models.MCPServer        models.ServerConfiguration      models.ServerMetrics    models.ServerStatus     models.UserPreferences  models.WindowLayout
Not found: time.Time
KnownStructs: dependencies.UpdateInfo   main.DiscoverServersResponse    main.GetDependenciesResponse    main.GetLogsResponse    main.ListServersResponse        main.ServerOperationResponse    main.UpdateApplicationStateResponse     models.ApplicationState models.Dependency       models.Filters  models.LogEntry models.MCPServer        models.ServerConfiguration      models.ServerMetrics    models.ServerStatus     models.UserPreferences  models.WindowLayout
Not found: time.Time
KnownStructs: dependencies.UpdateInfo   main.DiscoverServersResponse    main.GetDependenciesResponse    main.GetLogsResponse    main.ListServersResponse        main.ServerOperationResponse    main.UpdateApplicationStateResponse     models.ApplicationState models.Dependency       models.Filters  models.LogEntry models.MCPServer        models.ServerConfiguration      models.ServerMetrics    models.ServerStatus     models.UserPreferences  models.WindowLayout
Not found: time.Time
KnownStructs: dependencies.UpdateInfo   main.DiscoverServersResponse    main.GetDependenciesResponse    main.GetLogsResponse    main.ListServersResponse        main.ServerOperationResponse    main.UpdateApplicationStateResponse     models.ApplicationState models.Dependency       models.Filters  models.LogEntry models.MCPServer        models.ServerConfiguration      models.ServerMetrics    models.ServerStatus     models.UserPreferences  models.WindowLayout
Not found: time.Time
KnownStructs: dependencies.UpdateInfo   main.DiscoverServersResponse    main.GetDependenciesResponse    main.GetLogsResponse    main.ListServersResponse        main.ServerOperationResponse    main.UpdateApplicationStateResponse     models.ApplicationState models.Dependency       models.Filters  models.LogEntry models.MCPServer        models.ServerConfiguration      models.ServerMetrics    models.ServerStatus     models.UserPreferences  models.WindowLayout
Not found: time.Time

Done.
  • Generating application assets: Done.
  • Compiling application: Done.
 INFO  Wails is now using the new Go WebView2Loader. If you encounter any issues with it, please report them to https://github.com/wailsapp/wails/issues/2004. You could also use the old legacy loader with `-tags native_webview2loader`, but keep in mind this will be deprecated in the near future.

Using DevServer URL: http://localhost:34115
Using Frontend DevServer URL: http://localhost:5173/
Using reload debounce setting of 100 milliseconds
time=2025-10-22T22:34:36.364-05:00 level=INFO msg="Starting MCP Manager Desktop Application" version=0.1.0
INF | Serving assets from frontend DevServer URL: http://localhost:5173/
time=2025-10-22T22:34:36.419-05:00 level=INFO msg="[WebView2] Environment created successfully"
Watching (sub)/directory: D:\dev\ARTIFICIAL_INTELLIGENCE\MCP\MCPManager
time=2025-10-22T22:34:36.921-05:00 level=INFO msg="Starting MCP Manager Wails application"
time=2025-10-22T22:34:36.921-05:00 level=INFO msg="EventBus initialized"
time=2025-10-22T22:34:36.922-05:00 level=INFO msg="Event subscriptions configured"
time=2025-10-22T22:34:36.922-05:00 level=INFO msg="Storage service initialized"
time=2025-10-22T22:34:36.922-05:00 level=INFO msg="Discovery service initialized"
time=2025-10-22T22:34:36.922-05:00 level=INFO msg="Lifecycle service initialized"
time=2025-10-22T22:34:36.922-05:00 level=INFO msg="Config service initialized"
time=2025-10-22T22:34:36.922-05:00 level=INFO msg="Monitoring service initialized"
time=2025-10-22T22:34:36.922-05:00 level=INFO msg="Metrics collector initialized"
time=2025-10-22T22:34:36.922-05:00 level=INFO msg="Dependency service initialized"
time=2025-10-22T22:34:36.922-05:00 level=INFO msg="Update checker initialized"
time=2025-10-22T22:34:36.922-05:00 level=INFO msg="Running initial server discovery..."

=== MCP SERVER DISCOVERY START ===

[PHASE 1] Discovering from client configs...
  Config directory: C:\Users\hoyth\AppData\Roaming
  Checking Claude Desktop config: C:\Users\hoyth\AppData\Roaming\Claude\claude_desktop_config.json
    File exists, reading...
    Read 59 bytes
    Parsed JSON, found 0 server entries
    Found 0 servers
  Checking Cursor config: C:\Users\hoyth\AppData\Roaming\Cursor\mcp_config.json
    File does not exist
    Found 0 servers
[PHASE 1] Found 0 servers from client configs

[PHASE 1.5] Discovering from Claude Extensions...
  Config directory: C:\Users\hoyth\AppData\Roaming
  Extensions directory: C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions
  Settings directory: C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions Settings
  Found 3 extension entries
    Scanning extension: ant.dir.ant.anthropic.filesystem
      Found: Filesystem v0.1.6
      Extension enabled: true
      Added to server list (ID: 73054ce2-68a8-40ed-9bbf-ffb75e521d53)
    Scanning extension: ant.dir.ant.figma.figma
      Found: Figma Dev Mode v1.0.3
      Extension enabled: true
      Added to server list (ID: b14ec591-df5f-4b99-a4d8-f2a03b05a5e6)
    Scanning extension: ant.dir.cursortouch.windows-mcp
      Found:  v0.1.0
      Extension enabled: true
      Added to server list (ID: ab133625-6a61-4d84-a554-f5cf582a341b)
[PHASE 1.5] Found 3 servers from Claude Extensions
  [1] Filesystem (cmd: node, source: extension, version: 0.1.6)
  [2] Figma Dev Mode (cmd: node, source: extension, version: 1.0.3)
  [3] Windows-MCP (cmd: uv, source: extension, version: 0.1.0)

[PHASE 2] Discovering from filesystem...
  Scanning NPM global packages...
    Running: npm root -g
    NPM global root: C:\Users\hoyth\AppData\Roaming\npm\node_modules
    Scanning 16 entries in NPM directory...
    Found 0 NPM servers
  Scanning Python site-packages...
    Found 0 Python servers
  Scanning Go binaries...
    Found 0 Go servers
[PHASE 2] Found 0 servers from filesystem

[MERGE] Merging servers from all sources...
[MERGE] Total unique servers after merge: 3

[PHASE 3] Matching running processes to discovered servers...
  Getting running processes...
  Found 16 processes to match against
    PID 38780: claude.exe
      CMD: "C:\Users\hoyth\.local\bin\claude.exe"
    PID 5800: claude.exe
      CMD: "C:\Users\hoyth\AppData\Local\AnthropicClaude\app-0.14.4\claude.exe"
    PID 32692: claude.exe
      CMD: C:\Users\hoyth\AppData\Local\AnthropicClaude\app-0.14.4\claude.exe --type=crashpad-handler --user-data-dir=C:\Users\hoyth\AppData\Roaming\Claude /prefetch:4 --no-rate-limit --monitor-self-annotation=ptype=crashpad-handler --database=C:\Users\hoyth\AppData\Roaming\Claude\Crashpad --url=https://f.a.k/e --annotation=_productName=Claude --annotation=_version=0.14.4 --annotation=plat=Win64 --annotation=prod=Electron --annotation=ver=37.6.0 --initial-client-data=0x4c4,0x4c8,0x4cc,0x4c0,0x4d0,0x7ff6e92621f4,0x7ff6e9262200,0x7ff6e9262210
    PID 44652: claude.exe
      CMD: "C:\Users\hoyth\AppData\Local\AnthropicClaude\app-0.14.4\claude.exe" --type=gpu-process --user-data-dir="C:\Users\hoyth\AppData\Roaming\Claude" --gpu-preferences=UAAAAAAAAADgAAAEAAAAAAAAAAAAAAAAAABgAAEAAAAAAAAAAAAAAAAAAAACAAAAAAAAAAAAAAAAAAAAAAAAABAAAAAAAAAAEAAAAAAAAAAIAAAAAAAAAAgAAAAAAAAA --field-trial-handle=1800,i,18215422749562086063,15102240429383379184,262144 --enable-features=DocumentPolicyIncludeJSCallStacksInCrashReports,EnableTransparentHwndEnlargement,PdfUseShowSaveFilePicker --disable-features=ScreenAIOCREnabled,SpareRendererForSitePerProcess,WinDelaySpellcheckServiceInit --variations-seed-version --mojo-platform-channel-handle=1796 /prefetch:2
    PID 18260: claude.exe
      CMD: "C:\Users\hoyth\AppData\Local\AnthropicClaude\app-0.14.4\claude.exe" --type=utility --utility-sub-type=network.mojom.NetworkService --lang=en-US --service-sandbox-type=none --user-data-dir="C:\Users\hoyth\AppData\Roaming\Claude" --secure-schemes=sentry-ipc --bypasscsp-schemes=sentry-ipc --cors-schemes=sentry-ipc --fetch-schemes=sentry-ipc --field-trial-handle=1800,i,18215422749562086063,15102240429383379184,262144 --enable-features=DocumentPolicyIncludeJSCallStacksInCrashReports,EnableTransparentHwndEnlargement,PdfUseShowSaveFilePicker --disable-features=ScreenAIOCREnabled,SpareRendererForSitePerProcess,WinDelaySpellcheckServiceInit --variations-seed-version --mojo-platform-channel-handle=1932 /prefetch:11
    PID 19856: claude.exe
      CMD: "C:\Users\hoyth\AppData\Local\AnthropicClaude\app-0.14.4\claude.exe" --type=renderer --user-data-dir="C:\Users\hoyth\AppData\Roaming\Claude" --secure-schemes=sentry-ipc --bypasscsp-schemes=sentry-ipc --cors-schemes=sentry-ipc --fetch-schemes=sentry-ipc --app-user-model-id=com.squirrel.AnthropicClaude.claude --app-path="C:\Users\hoyth\AppData\Local\AnthropicClaude\app-0.14.4\resources\app.asar" --enable-sandbox --video-capture-use-gpu-memory-buffer --lang=en-US --device-scale-factor=1.5 --num-raster-threads=4 --enable-main-frame-before-activation --renderer-client-id=4 --time-ticks-at-unix-epoch=-1760518805885947 --launch-time-ticks=670306037456 --field-trial-handle=1800,i,18215422749562086063,15102240429383379184,262144 --enable-features=DocumentPolicyIncludeJSCallStacksInCrashReports,EnableTransparentHwndEnlargement,PdfUseShowSaveFilePicker --disable-features=ScreenAIOCREnabled,SpareRendererForSitePerProcess,WinDelaySpellcheckServiceInit --variations-seed-version --mojo-platform-channel-handle=2252 /prefetch:1
    PID 24160: claude.exe
      CMD: "C:\Users\hoyth\AppData\Local\AnthropicClaude\app-0.14.4\claude.exe" --type=renderer --user-data-dir="C:\Users\hoyth\AppData\Roaming\Claude" --secure-schemes=sentry-ipc --bypasscsp-schemes=sentry-ipc --cors-schemes=sentry-ipc --fetch-schemes=sentry-ipc --app-user-model-id=com.squirrel.AnthropicClaude.claude --app-path="C:\Users\hoyth\AppData\Local\AnthropicClaude\app-0.14.4\resources\app.asar" --enable-sandbox --video-capture-use-gpu-memory-buffer --lang=en-US --device-scale-factor=1.5 --num-raster-threads=4 --enable-main-frame-before-activation --renderer-client-id=6 --time-ticks-at-unix-epoch=-1760518805885947 --launch-time-ticks=670306067992 --field-trial-handle=1800,i,18215422749562086063,15102240429383379184,262144 --enable-features=DocumentPolicyIncludeJSCallStacksInCrashReports,EnableTransparentHwndEnlargement,PdfUseShowSaveFilePicker --disable-features=ScreenAIOCREnabled,SpareRendererForSitePerProcess,WinDelaySpellcheckServiceInit --variations-seed-version --mojo-platform-channel-handle=2880 /prefetch:1
    PID 11500: claude.exe
      CMD: "C:\Users\hoyth\AppData\Local\AnthropicClaude\app-0.14.4\claude.exe" --type=renderer --user-data-dir="C:\Users\hoyth\AppData\Roaming\Claude" --secure-schemes=sentry-ipc --bypasscsp-schemes=sentry-ipc --cors-schemes=sentry-ipc --fetch-schemes=sentry-ipc --app-user-model-id=com.squirrel.AnthropicClaude.claude --app-path="C:\Users\hoyth\AppData\Local\AnthropicClaude\app-0.14.4\resources\app.asar" --enable-sandbox --video-capture-use-gpu-memory-buffer --lang=en-US --device-scale-factor=1.5 --num-raster-threads=4 --enable-main-frame-before-activation --renderer-client-id=7 --time-ticks-at-unix-epoch=-1760518805885947 --launch-time-ticks=670306966351 --field-trial-handle=1800,i,18215422749562086063,15102240429383379184,262144 --enable-features=DocumentPolicyIncludeJSCallStacksInCrashReports,EnableTransparentHwndEnlargement,PdfUseShowSaveFilePicker --disable-features=ScreenAIOCREnabled,SpareRendererForSitePerProcess,WinDelaySpellcheckServiceInit --variations-seed-version --mojo-platform-channel-handle=2600 /prefetch:1
    PID 28664: uv.exe
      CMD: "C:\Program Files\Python313\Scripts\uv.exe" --directory "C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.cursortouch.windows-mcp" run main.py
    PID 22300: python.exe
      CMD: "C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.cursortouch.windows-mcp\.venv\Scripts\python.exe" main.py
    PID 34068: python.exe
      CMD: "C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.cursortouch.windows-mcp\.venv\Scripts\python.exe" main.py
    PID 10808: node.exe
      CMD: "C:\Program Files\nodejs\node.exe" "C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.ant.figma.figma/server/index.js"
    PID 17892: node.exe
      CMD: "C:\Program Files\nodejs\node.exe" "C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.ant.anthropic.filesystem/server/index.js" D:\dev D:\bin
    PID 36728: claude.exe
      CMD: "C:\Users\hoyth\AppData\Local\AnthropicClaude\app-0.14.4\claude.exe" --type=utility --utility-sub-type=audio.mojom.AudioService --lang=en-US --service-sandbox-type=audio --video-capture-use-gpu-memory-buffer --user-data-dir="C:\Users\hoyth\AppData\Roaming\Claude" --secure-schemes=sentry-ipc --bypasscsp-schemes=sentry-ipc --cors-schemes=sentry-ipc --fetch-schemes=sentry-ipc --field-trial-handle=1800,i,18215422749562086063,15102240429383379184,262144 --enable-features=DocumentPolicyIncludeJSCallStacksInCrashReports,EnableTransparentHwndEnlargement,PdfUseShowSaveFilePicker --disable-features=ScreenAIOCREnabled,SpareRendererForSitePerProcess,WinDelaySpellcheckServiceInit --variations-seed-version --mojo-platform-channel-handle=3028 /prefetch:12
    PID 43988: node.exe
      CMD: "C:\Program Files\nodejs\\node.exe"  "C:\Users\hoyth\AppData\Roaming\npm\node_modules\npm\bin\npm-cli.js" run dev
    PID 21592: node.exe
      CMD: "node"   "D:\dev\ARTIFICIAL_INTELLIGENCE\MCP\MCPManager\frontend\node_modules\.bin\\..\vite\bin\vite.js"
  Matching server: Filesystem (cmd: node)
      Checking process PID 38780:
        Process: claude.exe
        CmdLine: "c:\users\hoyth\.local\bin\claude.exe"
        Server cmd: node, args: [C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.ant.anthropic.filesystem/server/index.js D:\dev,D:\bin]
        ✗ Process name doesn't match server command
        ✗ No match
      Checking process PID 5800:
        Process: claude.exe
        CmdLine: "c:\users\hoyth\appdata\local\anthropicclaude\app-0.14.4\claude.exe"
        Server cmd: node, args: [C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.ant.anthropic.filesystem/server/index.js D:\dev,D:\bin]
        ✗ Process name doesn't match server command
        ✗ No match
      Checking process PID 32692:
        Process: claude.exe
        CmdLine: c:\users\hoyth\appdata\local\anthropicclaude\app-0.14.4\claude.exe --type=crashpad-handler --user-data-dir=c:\users\hoyth\appdata\roaming\claude /prefetch:4 --no-rate-limit --monitor-self-annotation=ptype=crashpad-handler --database=c:\users\hoyth\appdata\roaming\claude\crashpad --url=https://f.a.k/e --annotation=_productname=claude --annotation=_version=0.14.4 --annotation=plat=win64 --annotation=prod=electron --annotation=ver=37.6.0 --initial-client-data=0x4c4,0x4c8,0x4cc,0x4c0,0x4d0,0x7ff6e92621f4,0x7ff6e9262200,0x7ff6e9262210
        Server cmd: node, args: [C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.ant.anthropic.filesystem/server/index.js D:\dev,D:\bin]
        ✗ Process name doesn't match server command
        ✗ No match
      Checking process PID 44652:
        Process: claude.exe
        CmdLine: "c:\users\hoyth\appdata\local\anthropicclaude\app-0.14.4\claude.exe" --type=gpu-process --user-data-dir="c:\users\hoyth\appdata\roaming\claude" --gpu-preferences=uaaaaaaaaadgaaaeaaaaaaaaaaaaaaaaaabgaaeaaaaaaaaaaaaaaaaaaaacaaaaaaaaaaaaaaaaaaaaaaaaabaaaaaaaaaaeaaaaaaaaaaiaaaaaaaaaagaaaaaaaaa --field-trial-handle=1800,i,18215422749562086063,15102240429383379184,262144 --enable-features=documentpolicyincludejscallstacksincrashreports,enabletransparenthwndenlargement,pdfuseshowsavefilepicker --disable-features=screenaiocrenabled,sparerendererforsiteperprocess,windelayspellcheckserviceinit --variations-seed-version --mojo-platform-channel-handle=1796 /prefetch:2
        Server cmd: node, args: [C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.ant.anthropic.filesystem/server/index.js D:\dev,D:\bin]
        ✗ Process name doesn't match server command
        ✗ No match
      Checking process PID 18260:
        Process: claude.exe
        CmdLine: "c:\users\hoyth\appdata\local\anthropicclaude\app-0.14.4\claude.exe" --type=utility --utility-sub-type=network.mojom.networkservice --lang=en-us --service-sandbox-type=none --user-data-dir="c:\users\hoyth\appdata\roaming\claude" --secure-schemes=sentry-ipc --bypasscsp-schemes=sentry-ipc --cors-schemes=sentry-ipc --fetch-schemes=sentry-ipc --field-trial-handle=1800,i,18215422749562086063,15102240429383379184,262144 --enable-features=documentpolicyincludejscallstacksincrashreports,enabletransparenthwndenlargement,pdfuseshowsavefilepicker --disable-features=screenaiocrenabled,sparerendererforsiteperprocess,windelayspellcheckserviceinit --variations-seed-version --mojo-platform-channel-handle=1932 /prefetch:11
        Server cmd: node, args: [C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.ant.anthropic.filesystem/server/index.js D:\dev,D:\bin]
        ✗ Process name doesn't match server command
        ✗ No match
      Checking process PID 19856:
        Process: claude.exe
        CmdLine: "c:\users\hoyth\appdata\local\anthropicclaude\app-0.14.4\claude.exe" --type=renderer --user-data-dir="c:\users\hoyth\appdata\roaming\claude" --secure-schemes=sentry-ipc --bypasscsp-schemes=sentry-ipc --cors-schemes=sentry-ipc --fetch-schemes=sentry-ipc --app-user-model-id=com.squirrel.anthropicclaude.claude --app-path="c:\users\hoyth\appdata\local\anthropicclaude\app-0.14.4\resources\app.asar" --enable-sandbox --video-capture-use-gpu-memory-buffer --lang=en-us --device-scale-factor=1.5 --num-raster-threads=4 --enable-main-frame-before-activation --renderer-client-id=4 --time-ticks-at-unix-epoch=-1760518805885947 --launch-time-ticks=670306037456 --field-trial-handle=1800,i,18215422749562086063,15102240429383379184,262144 --enable-features=documentpolicyincludejscallstacksincrashreports,enabletransparenthwndenlargement,pdfuseshowsavefilepicker --disable-features=screenaiocrenabled,sparerendererforsiteperprocess,windelayspellcheckserviceinit --variations-seed-version --mojo-platform-channel-handle=2252 /prefetch:1
        Server cmd: node, args: [C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.ant.anthropic.filesystem/server/index.js D:\dev,D:\bin]
        ✗ Process name doesn't match server command
        ✗ No match
      Checking process PID 24160:
        Process: claude.exe
        CmdLine: "c:\users\hoyth\appdata\local\anthropicclaude\app-0.14.4\claude.exe" --type=renderer --user-data-dir="c:\users\hoyth\appdata\roaming\claude" --secure-schemes=sentry-ipc --bypasscsp-schemes=sentry-ipc --cors-schemes=sentry-ipc --fetch-schemes=sentry-ipc --app-user-model-id=com.squirrel.anthropicclaude.claude --app-path="c:\users\hoyth\appdata\local\anthropicclaude\app-0.14.4\resources\app.asar" --enable-sandbox --video-capture-use-gpu-memory-buffer --lang=en-us --device-scale-factor=1.5 --num-raster-threads=4 --enable-main-frame-before-activation --renderer-client-id=6 --time-ticks-at-unix-epoch=-1760518805885947 --launch-time-ticks=670306067992 --field-trial-handle=1800,i,18215422749562086063,15102240429383379184,262144 --enable-features=documentpolicyincludejscallstacksincrashreports,enabletransparenthwndenlargement,pdfuseshowsavefilepicker --disable-features=screenaiocrenabled,sparerendererforsiteperprocess,windelayspellcheckserviceinit --variations-seed-version --mojo-platform-channel-handle=2880 /prefetch:1
        Server cmd: node, args: [C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.ant.anthropic.filesystem/server/index.js D:\dev,D:\bin]
        ✗ Process name doesn't match server command
        ✗ No match
      Checking process PID 11500:
        Process: claude.exe
        CmdLine: "c:\users\hoyth\appdata\local\anthropicclaude\app-0.14.4\claude.exe" --type=renderer --user-data-dir="c:\users\hoyth\appdata\roaming\claude" --secure-schemes=sentry-ipc --bypasscsp-schemes=sentry-ipc --cors-schemes=sentry-ipc --fetch-schemes=sentry-ipc --app-user-model-id=com.squirrel.anthropicclaude.claude --app-path="c:\users\hoyth\appdata\local\anthropicclaude\app-0.14.4\resources\app.asar" --enable-sandbox --video-capture-use-gpu-memory-buffer --lang=en-us --device-scale-factor=1.5 --num-raster-threads=4 --enable-main-frame-before-activation --renderer-client-id=7 --time-ticks-at-unix-epoch=-1760518805885947 --launch-time-ticks=670306966351 --field-trial-handle=1800,i,18215422749562086063,15102240429383379184,262144 --enable-features=documentpolicyincludejscallstacksincrashreports,enabletransparenthwndenlargement,pdfuseshowsavefilepicker --disable-features=screenaiocrenabled,sparerendererforsiteperprocess,windelayspellcheckserviceinit --variations-seed-version --mojo-platform-channel-handle=2600 /prefetch:1
        Server cmd: node, args: [C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.ant.anthropic.filesystem/server/index.js D:\dev,D:\bin]
        ✗ Process name doesn't match server command
        ✗ No match
      Checking process PID 28664:
        Process: uv.exe
        CmdLine: "c:\program files\python313\scripts\uv.exe" --directory "c:\users\hoyth\appdata\roaming\claude\claude extensions\ant.dir.cursortouch.windows-mcp" run main.py
        Server cmd: node, args: [C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.ant.anthropic.filesystem/server/index.js D:\dev,D:\bin]
        ✗ Process name doesn't match server command
        ✗ No match
      Checking process PID 22300:
        Process: python.exe
        CmdLine: "c:\users\hoyth\appdata\roaming\claude\claude extensions\ant.dir.cursortouch.windows-mcp\.venv\scripts\python.exe" main.py
        Server cmd: node, args: [C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.ant.anthropic.filesystem/server/index.js D:\dev,D:\bin]
        ✗ Process name doesn't match server command
        ✗ No match
      Checking process PID 34068:
        Process: python.exe
        CmdLine: "c:\users\hoyth\appdata\roaming\claude\claude extensions\ant.dir.cursortouch.windows-mcp\.venv\scripts\python.exe" main.py
        Server cmd: node, args: [C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.ant.anthropic.filesystem/server/index.js D:\dev,D:\bin]
        ✗ Process name doesn't match server command
        ✗ No match
      Checking process PID 10808:
        Process: node.exe
        CmdLine: "c:\program files\nodejs\node.exe" "c:\users\hoyth\appdata\roaming\claude\claude extensions\ant.dir.ant.figma.figma/server/index.js"
        Server cmd: node, args: [C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.ant.anthropic.filesystem/server/index.js D:\dev,D:\bin]
        ✗ No match
      Checking process PID 17892:
        Process: node.exe
        CmdLine: "c:\program files\nodejs\node.exe" "c:\users\hoyth\appdata\roaming\claude\claude extensions\ant.dir.ant.anthropic.filesystem/server/index.js" d:\dev d:\bin
        Server cmd: node, args: [C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.ant.anthropic.filesystem/server/index.js D:\dev,D:\bin]
        ✓ Argument match: C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.ant.anthropic.filesystem/server/index.js
    ✓ MATCHED PID 17892
  Matching server: Figma Dev Mode (cmd: node)
      Checking process PID 38780:
        Process: claude.exe
        CmdLine: "c:\users\hoyth\.local\bin\claude.exe"
        Server cmd: node, args: [C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.ant.figma.figma/server/index.js]
        ✗ Process name doesn't match server command
        ✗ No match
      Checking process PID 5800:
        Process: claude.exe
        CmdLine: "c:\users\hoyth\appdata\local\anthropicclaude\app-0.14.4\claude.exe"
        Server cmd: node, args: [C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.ant.figma.figma/server/index.js]
        ✗ Process name doesn't match server command
        ✗ No match
      Checking process PID 32692:
        Process: claude.exe
        CmdLine: c:\users\hoyth\appdata\local\anthropicclaude\app-0.14.4\claude.exe --type=crashpad-handler --user-data-dir=c:\users\hoyth\appdata\roaming\claude /prefetch:4 --no-rate-limit --monitor-self-annotation=ptype=crashpad-handler --database=c:\users\hoyth\appdata\roaming\claude\crashpad --url=https://f.a.k/e --annotation=_productname=claude --annotation=_version=0.14.4 --annotation=plat=win64 --annotation=prod=electron --annotation=ver=37.6.0 --initial-client-data=0x4c4,0x4c8,0x4cc,0x4c0,0x4d0,0x7ff6e92621f4,0x7ff6e9262200,0x7ff6e9262210
        Server cmd: node, args: [C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.ant.figma.figma/server/index.js]
        ✗ Process name doesn't match server command
        ✗ No match
      Checking process PID 44652:
        Process: claude.exe
        CmdLine: "c:\users\hoyth\appdata\local\anthropicclaude\app-0.14.4\claude.exe" --type=gpu-process --user-data-dir="c:\users\hoyth\appdata\roaming\claude" --gpu-preferences=uaaaaaaaaadgaaaeaaaaaaaaaaaaaaaaaabgaaeaaaaaaaaaaaaaaaaaaaacaaaaaaaaaaaaaaaaaaaaaaaaabaaaaaaaaaaeaaaaaaaaaaiaaaaaaaaaagaaaaaaaaa --field-trial-handle=1800,i,18215422749562086063,15102240429383379184,262144 --enable-features=documentpolicyincludejscallstacksincrashreports,enabletransparenthwndenlargement,pdfuseshowsavefilepicker --disable-features=screenaiocrenabled,sparerendererforsiteperprocess,windelayspellcheckserviceinit --variations-seed-version --mojo-platform-channel-handle=1796 /prefetch:2
        Server cmd: node, args: [C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.ant.figma.figma/server/index.js]
        ✗ Process name doesn't match server command
        ✗ No match
      Checking process PID 18260:
        Process: claude.exe
        CmdLine: "c:\users\hoyth\appdata\local\anthropicclaude\app-0.14.4\claude.exe" --type=utility --utility-sub-type=network.mojom.networkservice --lang=en-us --service-sandbox-type=none --user-data-dir="c:\users\hoyth\appdata\roaming\claude" --secure-schemes=sentry-ipc --bypasscsp-schemes=sentry-ipc --cors-schemes=sentry-ipc --fetch-schemes=sentry-ipc --field-trial-handle=1800,i,18215422749562086063,15102240429383379184,262144 --enable-features=documentpolicyincludejscallstacksincrashreports,enabletransparenthwndenlargement,pdfuseshowsavefilepicker --disable-features=screenaiocrenabled,sparerendererforsiteperprocess,windelayspellcheckserviceinit --variations-seed-version --mojo-platform-channel-handle=1932 /prefetch:11
        Server cmd: node, args: [C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.ant.figma.figma/server/index.js]
        ✗ Process name doesn't match server command
        ✗ No match
      Checking process PID 19856:
        Process: claude.exe
        CmdLine: "c:\users\hoyth\appdata\local\anthropicclaude\app-0.14.4\claude.exe" --type=renderer --user-data-dir="c:\users\hoyth\appdata\roaming\claude" --secure-schemes=sentry-ipc --bypasscsp-schemes=sentry-ipc --cors-schemes=sentry-ipc --fetch-schemes=sentry-ipc --app-user-model-id=com.squirrel.anthropicclaude.claude --app-path="c:\users\hoyth\appdata\local\anthropicclaude\app-0.14.4\resources\app.asar" --enable-sandbox --video-capture-use-gpu-memory-buffer --lang=en-us --device-scale-factor=1.5 --num-raster-threads=4 --enable-main-frame-before-activation --renderer-client-id=4 --time-ticks-at-unix-epoch=-1760518805885947 --launch-time-ticks=670306037456 --field-trial-handle=1800,i,18215422749562086063,15102240429383379184,262144 --enable-features=documentpolicyincludejscallstacksincrashreports,enabletransparenthwndenlargement,pdfuseshowsavefilepicker --disable-features=screenaiocrenabled,sparerendererforsiteperprocess,windelayspellcheckserviceinit --variations-seed-version --mojo-platform-channel-handle=2252 /prefetch:1
        Server cmd: node, args: [C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.ant.figma.figma/server/index.js]
        ✗ Process name doesn't match server command
        ✗ No match
      Checking process PID 24160:
        Process: claude.exe
        CmdLine: "c:\users\hoyth\appdata\local\anthropicclaude\app-0.14.4\claude.exe" --type=renderer --user-data-dir="c:\users\hoyth\appdata\roaming\claude" --secure-schemes=sentry-ipc --bypasscsp-schemes=sentry-ipc --cors-schemes=sentry-ipc --fetch-schemes=sentry-ipc --app-user-model-id=com.squirrel.anthropicclaude.claude --app-path="c:\users\hoyth\appdata\local\anthropicclaude\app-0.14.4\resources\app.asar" --enable-sandbox --video-capture-use-gpu-memory-buffer --lang=en-us --device-scale-factor=1.5 --num-raster-threads=4 --enable-main-frame-before-activation --renderer-client-id=6 --time-ticks-at-unix-epoch=-1760518805885947 --launch-time-ticks=670306067992 --field-trial-handle=1800,i,18215422749562086063,15102240429383379184,262144 --enable-features=documentpolicyincludejscallstacksincrashreports,enabletransparenthwndenlargement,pdfuseshowsavefilepicker --disable-features=screenaiocrenabled,sparerendererforsiteperprocess,windelayspellcheckserviceinit --variations-seed-version --mojo-platform-channel-handle=2880 /prefetch:1
        Server cmd: node, args: [C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.ant.figma.figma/server/index.js]
        ✗ Process name doesn't match server command
        ✗ No match
      Checking process PID 11500:
        Process: claude.exe
        CmdLine: "c:\users\hoyth\appdata\local\anthropicclaude\app-0.14.4\claude.exe" --type=renderer --user-data-dir="c:\users\hoyth\appdata\roaming\claude" --secure-schemes=sentry-ipc --bypasscsp-schemes=sentry-ipc --cors-schemes=sentry-ipc --fetch-schemes=sentry-ipc --app-user-model-id=com.squirrel.anthropicclaude.claude --app-path="c:\users\hoyth\appdata\local\anthropicclaude\app-0.14.4\resources\app.asar" --enable-sandbox --video-capture-use-gpu-memory-buffer --lang=en-us --device-scale-factor=1.5 --num-raster-threads=4 --enable-main-frame-before-activation --renderer-client-id=7 --time-ticks-at-unix-epoch=-1760518805885947 --launch-time-ticks=670306966351 --field-trial-handle=1800,i,18215422749562086063,15102240429383379184,262144 --enable-features=documentpolicyincludejscallstacksincrashreports,enabletransparenthwndenlargement,pdfuseshowsavefilepicker --disable-features=screenaiocrenabled,sparerendererforsiteperprocess,windelayspellcheckserviceinit --variations-seed-version --mojo-platform-channel-handle=2600 /prefetch:1
        Server cmd: node, args: [C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.ant.figma.figma/server/index.js]
        ✗ Process name doesn't match server command
        ✗ No match
      Checking process PID 28664:
        Process: uv.exe
        CmdLine: "c:\program files\python313\scripts\uv.exe" --directory "c:\users\hoyth\appdata\roaming\claude\claude extensions\ant.dir.cursortouch.windows-mcp" run main.py
        Server cmd: node, args: [C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.ant.figma.figma/server/index.js]
        ✗ Process name doesn't match server command
        ✗ No match
      Checking process PID 22300:
        Process: python.exe
        CmdLine: "c:\users\hoyth\appdata\roaming\claude\claude extensions\ant.dir.cursortouch.windows-mcp\.venv\scripts\python.exe" main.py
        Server cmd: node, args: [C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.ant.figma.figma/server/index.js]
        ✗ Process name doesn't match server command
        ✗ No match
      Checking process PID 34068:
        Process: python.exe
        CmdLine: "c:\users\hoyth\appdata\roaming\claude\claude extensions\ant.dir.cursortouch.windows-mcp\.venv\scripts\python.exe" main.py
        Server cmd: node, args: [C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.ant.figma.figma/server/index.js]
        ✗ Process name doesn't match server command
        ✗ No match
      Checking process PID 10808:
        Process: node.exe
        CmdLine: "c:\program files\nodejs\node.exe" "c:\users\hoyth\appdata\roaming\claude\claude extensions\ant.dir.ant.figma.figma/server/index.js"
        Server cmd: node, args: [C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.ant.figma.figma/server/index.js]
        ✓ Argument match: C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.ant.figma.figma/server/index.js
    ✓ MATCHED PID 10808
  Matching server: Windows-MCP (cmd: uv)
      Checking process PID 38780:
        Process: claude.exe
        CmdLine: "c:\users\hoyth\.local\bin\claude.exe"
        Server cmd: uv, args: [--directory C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.cursortouch.windows-mcp run main.py]
        ✗ Process name doesn't match server command
        ✗ No match
      Checking process PID 5800:
        Process: claude.exe
        CmdLine: "c:\users\hoyth\appdata\local\anthropicclaude\app-0.14.4\claude.exe"
        Server cmd: uv, args: [--directory C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.cursortouch.windows-mcp run main.py]
        ✗ Process name doesn't match server command
        ✗ No match
      Checking process PID 32692:
        Process: claude.exe
        CmdLine: c:\users\hoyth\appdata\local\anthropicclaude\app-0.14.4\claude.exe --type=crashpad-handler --user-data-dir=c:\users\hoyth\appdata\roaming\claude /prefetch:4 --no-rate-limit --monitor-self-annotation=ptype=crashpad-handler --database=c:\users\hoyth\appdata\roaming\claude\crashpad --url=https://f.a.k/e --annotation=_productname=claude --annotation=_version=0.14.4 --annotation=plat=win64 --annotation=prod=electron --annotation=ver=37.6.0 --initial-client-data=0x4c4,0x4c8,0x4cc,0x4c0,0x4d0,0x7ff6e92621f4,0x7ff6e9262200,0x7ff6e9262210
        Server cmd: uv, args: [--directory C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.cursortouch.windows-mcp run main.py]
        ✗ Process name doesn't match server command
        ✗ No match
      Checking process PID 44652:
        Process: claude.exe
        CmdLine: "c:\users\hoyth\appdata\local\anthropicclaude\app-0.14.4\claude.exe" --type=gpu-process --user-data-dir="c:\users\hoyth\appdata\roaming\claude" --gpu-preferences=uaaaaaaaaadgaaaeaaaaaaaaaaaaaaaaaabgaaeaaaaaaaaaaaaaaaaaaaacaaaaaaaaaaaaaaaaaaaaaaaaabaaaaaaaaaaeaaaaaaaaaaiaaaaaaaaaagaaaaaaaaa --field-trial-handle=1800,i,18215422749562086063,15102240429383379184,262144 --enable-features=documentpolicyincludejscallstacksincrashreports,enabletransparenthwndenlargement,pdfuseshowsavefilepicker --disable-features=screenaiocrenabled,sparerendererforsiteperprocess,windelayspellcheckserviceinit --variations-seed-version --mojo-platform-channel-handle=1796 /prefetch:2
        Server cmd: uv, args: [--directory C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.cursortouch.windows-mcp run main.py]
        ✗ Process name doesn't match server command
        ✗ No match
      Checking process PID 18260:
        Process: claude.exe
        CmdLine: "c:\users\hoyth\appdata\local\anthropicclaude\app-0.14.4\claude.exe" --type=utility --utility-sub-type=network.mojom.networkservice --lang=en-us --service-sandbox-type=none --user-data-dir="c:\users\hoyth\appdata\roaming\claude" --secure-schemes=sentry-ipc --bypasscsp-schemes=sentry-ipc --cors-schemes=sentry-ipc --fetch-schemes=sentry-ipc --field-trial-handle=1800,i,18215422749562086063,15102240429383379184,262144 --enable-features=documentpolicyincludejscallstacksincrashreports,enabletransparenthwndenlargement,pdfuseshowsavefilepicker --disable-features=screenaiocrenabled,sparerendererforsiteperprocess,windelayspellcheckserviceinit --variations-seed-version --mojo-platform-channel-handle=1932 /prefetch:11
        Server cmd: uv, args: [--directory C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.cursortouch.windows-mcp run main.py]
        ✗ Process name doesn't match server command
        ✗ No match
      Checking process PID 19856:
        Process: claude.exe
        CmdLine: "c:\users\hoyth\appdata\local\anthropicclaude\app-0.14.4\claude.exe" --type=renderer --user-data-dir="c:\users\hoyth\appdata\roaming\claude" --secure-schemes=sentry-ipc --bypasscsp-schemes=sentry-ipc --cors-schemes=sentry-ipc --fetch-schemes=sentry-ipc --app-user-model-id=com.squirrel.anthropicclaude.claude --app-path="c:\users\hoyth\appdata\local\anthropicclaude\app-0.14.4\resources\app.asar" --enable-sandbox --video-capture-use-gpu-memory-buffer --lang=en-us --device-scale-factor=1.5 --num-raster-threads=4 --enable-main-frame-before-activation --renderer-client-id=4 --time-ticks-at-unix-epoch=-1760518805885947 --launch-time-ticks=670306037456 --field-trial-handle=1800,i,18215422749562086063,15102240429383379184,262144 --enable-features=documentpolicyincludejscallstacksincrashreports,enabletransparenthwndenlargement,pdfuseshowsavefilepicker --disable-features=screenaiocrenabled,sparerendererforsiteperprocess,windelayspellcheckserviceinit --variations-seed-version --mojo-platform-channel-handle=2252 /prefetch:1
        Server cmd: uv, args: [--directory C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.cursortouch.windows-mcp run main.py]
        ✗ Process name doesn't match server command
        ✗ No match
      Checking process PID 24160:
        Process: claude.exe
        CmdLine: "c:\users\hoyth\appdata\local\anthropicclaude\app-0.14.4\claude.exe" --type=renderer --user-data-dir="c:\users\hoyth\appdata\roaming\claude" --secure-schemes=sentry-ipc --bypasscsp-schemes=sentry-ipc --cors-schemes=sentry-ipc --fetch-schemes=sentry-ipc --app-user-model-id=com.squirrel.anthropicclaude.claude --app-path="c:\users\hoyth\appdata\local\anthropicclaude\app-0.14.4\resources\app.asar" --enable-sandbox --video-capture-use-gpu-memory-buffer --lang=en-us --device-scale-factor=1.5 --num-raster-threads=4 --enable-main-frame-before-activation --renderer-client-id=6 --time-ticks-at-unix-epoch=-1760518805885947 --launch-time-ticks=670306067992 --field-trial-handle=1800,i,18215422749562086063,15102240429383379184,262144 --enable-features=documentpolicyincludejscallstacksincrashreports,enabletransparenthwndenlargement,pdfuseshowsavefilepicker --disable-features=screenaiocrenabled,sparerendererforsiteperprocess,windelayspellcheckserviceinit --variations-seed-version --mojo-platform-channel-handle=2880 /prefetch:1
        Server cmd: uv, args: [--directory C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.cursortouch.windows-mcp run main.py]
        ✗ Process name doesn't match server command
        ✗ No match
      Checking process PID 11500:
        Process: claude.exe
        CmdLine: "c:\users\hoyth\appdata\local\anthropicclaude\app-0.14.4\claude.exe" --type=renderer --user-data-dir="c:\users\hoyth\appdata\roaming\claude" --secure-schemes=sentry-ipc --bypasscsp-schemes=sentry-ipc --cors-schemes=sentry-ipc --fetch-schemes=sentry-ipc --app-user-model-id=com.squirrel.anthropicclaude.claude --app-path="c:\users\hoyth\appdata\local\anthropicclaude\app-0.14.4\resources\app.asar" --enable-sandbox --video-capture-use-gpu-memory-buffer --lang=en-us --device-scale-factor=1.5 --num-raster-threads=4 --enable-main-frame-before-activation --renderer-client-id=7 --time-ticks-at-unix-epoch=-1760518805885947 --launch-time-ticks=670306966351 --field-trial-handle=1800,i,18215422749562086063,15102240429383379184,262144 --enable-features=documentpolicyincludejscallstacksincrashreports,enabletransparenthwndenlargement,pdfuseshowsavefilepicker --disable-features=screenaiocrenabled,sparerendererforsiteperprocess,windelayspellcheckserviceinit --variations-seed-version --mojo-platform-channel-handle=2600 /prefetch:1
        Server cmd: uv, args: [--directory C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.cursortouch.windows-mcp run main.py]
        ✗ Process name doesn't match server command
        ✗ No match
      Checking process PID 28664:
        Process: uv.exe
        CmdLine: "c:\program files\python313\scripts\uv.exe" --directory "c:\users\hoyth\appdata\roaming\claude\claude extensions\ant.dir.cursortouch.windows-mcp" run main.py
        Server cmd: uv, args: [--directory C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.cursortouch.windows-mcp run main.py]
        ✓ Argument match: C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.cursortouch.windows-mcp
    ✓ MATCHED PID 28664
[PHASE 3] Matched 3 running processes

=== DISCOVERY COMPLETE: 3 total servers ===

time=2025-10-22T22:34:38.031-05:00 level=INFO msg="Initial discovery complete" servers_found=3
10:34:38 PM [vite-plugin-svelte] /src/App.svelte:98:4 A11y: noninteractive element cannot have nonnegative tabIndex value
10:34:38 PM [vite-plugin-svelte] /src/components/LogViewer.svelte:184:8 A11y: visible, non-interactive elements with an on:click event must be accompanied by an on:keydown, on:keyup, or on:keypress event.
10:34:38 PM [vite-plugin-svelte] /src/components/ConfigurationEditor.svelte:241:12 A11y: A form label must be associated with a control.
10:34:38 PM [vite-plugin-svelte] /src/components/ConfigurationEditor.svelte:308:12 A11y: A form label must be associated with a control.
10:34:38 PM [vite-plugin-svelte] /src/components/ConfigurationEditor.svelte:209:2 A11y: visible, non-interactive elements with an on:click event must be accompanied by an on:keydown, on:keyup, or on:keypress event.
time=2025-10-22T22:34:39.288-05:00 level=INFO msg="ListServers called"


To develop in the browser and call your bound Go methods from Javascript, navigate to: http://localhost:34115
time=2025-10-22T22:34:39.422-05:00 level=INFO msg="DiscoverServers called"

=== MCP SERVER DISCOVERY START ===

[PHASE 1] Discovering from client configs...
  Config directory: C:\Users\hoyth\AppData\Roaming
  Checking Claude Desktop config: C:\Users\hoyth\AppData\Roaming\Claude\claude_desktop_config.json
    File exists, reading...
    Read 59 bytes
    Parsed JSON, found 0 server entries
    Found 0 servers
  Checking Cursor config: C:\Users\hoyth\AppData\Roaming\Cursor\mcp_config.json
    File does not exist
    Found 0 servers
[PHASE 1] Found 0 servers from client configs

[PHASE 1.5] Discovering from Claude Extensions...
  Config directory: C:\Users\hoyth\AppData\Roaming
  Extensions directory: C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions
  Settings directory: C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions Settings
  Found 3 extension entries
    Scanning extension: ant.dir.ant.anthropic.filesystem
      Found: Filesystem v0.1.6
      Extension enabled: true
      Added to server list (ID: d93fcdc3-84e7-4932-863e-f554cd60d2a9)
    Scanning extension: ant.dir.ant.figma.figma
      Found: Figma Dev Mode v1.0.3
      Extension enabled: true
      Added to server list (ID: 90e5a5d3-90ad-44fc-a30f-39d3b8bec5d8)
    Scanning extension: ant.dir.cursortouch.windows-mcp
      Found:  v0.1.0
      Extension enabled: true
      Added to server list (ID: b94d52e1-f2af-401e-9b58-db2a6fc82af2)
[PHASE 1.5] Found 3 servers from Claude Extensions
  [1] Filesystem (cmd: node, source: extension, version: 0.1.6)
  [2] Figma Dev Mode (cmd: node, source: extension, version: 1.0.3)
  [3] Windows-MCP (cmd: uv, source: extension, version: 0.1.0)

[PHASE 2] Discovering from filesystem...
  Scanning NPM global packages...
    Running: npm root -g
    NPM global root: C:\Users\hoyth\AppData\Roaming\npm\node_modules
    Scanning 16 entries in NPM directory...
    Found 0 NPM servers
  Scanning Python site-packages...
    Found 0 Python servers
  Scanning Go binaries...
    Found 0 Go servers
[PHASE 2] Found 0 servers from filesystem

[MERGE] Merging servers from all sources...
[MERGE] Total unique servers after merge: 3

[PHASE 3] Matching running processes to discovered servers...
  Getting running processes...
  Found 16 processes to match against
    PID 38780: claude.exe
      CMD: "C:\Users\hoyth\.local\bin\claude.exe"
    PID 5800: claude.exe
      CMD: "C:\Users\hoyth\AppData\Local\AnthropicClaude\app-0.14.4\claude.exe"
    PID 32692: claude.exe
      CMD: C:\Users\hoyth\AppData\Local\AnthropicClaude\app-0.14.4\claude.exe --type=crashpad-handler --user-data-dir=C:\Users\hoyth\AppData\Roaming\Claude /prefetch:4 --no-rate-limit --monitor-self-annotation=ptype=crashpad-handler --database=C:\Users\hoyth\AppData\Roaming\Claude\Crashpad --url=https://f.a.k/e --annotation=_productName=Claude --annotation=_version=0.14.4 --annotation=plat=Win64 --annotation=prod=Electron --annotation=ver=37.6.0 --initial-client-data=0x4c4,0x4c8,0x4cc,0x4c0,0x4d0,0x7ff6e92621f4,0x7ff6e9262200,0x7ff6e9262210
    PID 44652: claude.exe
      CMD: "C:\Users\hoyth\AppData\Local\AnthropicClaude\app-0.14.4\claude.exe" --type=gpu-process --user-data-dir="C:\Users\hoyth\AppData\Roaming\Claude" --gpu-preferences=UAAAAAAAAADgAAAEAAAAAAAAAAAAAAAAAABgAAEAAAAAAAAAAAAAAAAAAAACAAAAAAAAAAAAAAAAAAAAAAAAABAAAAAAAAAAEAAAAAAAAAAIAAAAAAAAAAgAAAAAAAAA --field-trial-handle=1800,i,18215422749562086063,15102240429383379184,262144 --enable-features=DocumentPolicyIncludeJSCallStacksInCrashReports,EnableTransparentHwndEnlargement,PdfUseShowSaveFilePicker --disable-features=ScreenAIOCREnabled,SpareRendererForSitePerProcess,WinDelaySpellcheckServiceInit --variations-seed-version --mojo-platform-channel-handle=1796 /prefetch:2
    PID 18260: claude.exe
      CMD: "C:\Users\hoyth\AppData\Local\AnthropicClaude\app-0.14.4\claude.exe" --type=utility --utility-sub-type=network.mojom.NetworkService --lang=en-US --service-sandbox-type=none --user-data-dir="C:\Users\hoyth\AppData\Roaming\Claude" --secure-schemes=sentry-ipc --bypasscsp-schemes=sentry-ipc --cors-schemes=sentry-ipc --fetch-schemes=sentry-ipc --field-trial-handle=1800,i,18215422749562086063,15102240429383379184,262144 --enable-features=DocumentPolicyIncludeJSCallStacksInCrashReports,EnableTransparentHwndEnlargement,PdfUseShowSaveFilePicker --disable-features=ScreenAIOCREnabled,SpareRendererForSitePerProcess,WinDelaySpellcheckServiceInit --variations-seed-version --mojo-platform-channel-handle=1932 /prefetch:11
    PID 19856: claude.exe
      CMD: "C:\Users\hoyth\AppData\Local\AnthropicClaude\app-0.14.4\claude.exe" --type=renderer --user-data-dir="C:\Users\hoyth\AppData\Roaming\Claude" --secure-schemes=sentry-ipc --bypasscsp-schemes=sentry-ipc --cors-schemes=sentry-ipc --fetch-schemes=sentry-ipc --app-user-model-id=com.squirrel.AnthropicClaude.claude --app-path="C:\Users\hoyth\AppData\Local\AnthropicClaude\app-0.14.4\resources\app.asar" --enable-sandbox --video-capture-use-gpu-memory-buffer --lang=en-US --device-scale-factor=1.5 --num-raster-threads=4 --enable-main-frame-before-activation --renderer-client-id=4 --time-ticks-at-unix-epoch=-1760518805885947 --launch-time-ticks=670306037456 --field-trial-handle=1800,i,18215422749562086063,15102240429383379184,262144 --enable-features=DocumentPolicyIncludeJSCallStacksInCrashReports,EnableTransparentHwndEnlargement,PdfUseShowSaveFilePicker --disable-features=ScreenAIOCREnabled,SpareRendererForSitePerProcess,WinDelaySpellcheckServiceInit --variations-seed-version --mojo-platform-channel-handle=2252 /prefetch:1
    PID 24160: claude.exe
      CMD: "C:\Users\hoyth\AppData\Local\AnthropicClaude\app-0.14.4\claude.exe" --type=renderer --user-data-dir="C:\Users\hoyth\AppData\Roaming\Claude" --secure-schemes=sentry-ipc --bypasscsp-schemes=sentry-ipc --cors-schemes=sentry-ipc --fetch-schemes=sentry-ipc --app-user-model-id=com.squirrel.AnthropicClaude.claude --app-path="C:\Users\hoyth\AppData\Local\AnthropicClaude\app-0.14.4\resources\app.asar" --enable-sandbox --video-capture-use-gpu-memory-buffer --lang=en-US --device-scale-factor=1.5 --num-raster-threads=4 --enable-main-frame-before-activation --renderer-client-id=6 --time-ticks-at-unix-epoch=-1760518805885947 --launch-time-ticks=670306067992 --field-trial-handle=1800,i,18215422749562086063,15102240429383379184,262144 --enable-features=DocumentPolicyIncludeJSCallStacksInCrashReports,EnableTransparentHwndEnlargement,PdfUseShowSaveFilePicker --disable-features=ScreenAIOCREnabled,SpareRendererForSitePerProcess,WinDelaySpellcheckServiceInit --variations-seed-version --mojo-platform-channel-handle=2880 /prefetch:1
    PID 11500: claude.exe
      CMD: "C:\Users\hoyth\AppData\Local\AnthropicClaude\app-0.14.4\claude.exe" --type=renderer --user-data-dir="C:\Users\hoyth\AppData\Roaming\Claude" --secure-schemes=sentry-ipc --bypasscsp-schemes=sentry-ipc --cors-schemes=sentry-ipc --fetch-schemes=sentry-ipc --app-user-model-id=com.squirrel.AnthropicClaude.claude --app-path="C:\Users\hoyth\AppData\Local\AnthropicClaude\app-0.14.4\resources\app.asar" --enable-sandbox --video-capture-use-gpu-memory-buffer --lang=en-US --device-scale-factor=1.5 --num-raster-threads=4 --enable-main-frame-before-activation --renderer-client-id=7 --time-ticks-at-unix-epoch=-1760518805885947 --launch-time-ticks=670306966351 --field-trial-handle=1800,i,18215422749562086063,15102240429383379184,262144 --enable-features=DocumentPolicyIncludeJSCallStacksInCrashReports,EnableTransparentHwndEnlargement,PdfUseShowSaveFilePicker --disable-features=ScreenAIOCREnabled,SpareRendererForSitePerProcess,WinDelaySpellcheckServiceInit --variations-seed-version --mojo-platform-channel-handle=2600 /prefetch:1
    PID 28664: uv.exe
      CMD: "C:\Program Files\Python313\Scripts\uv.exe" --directory "C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.cursortouch.windows-mcp" run main.py
    PID 22300: python.exe
      CMD: "C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.cursortouch.windows-mcp\.venv\Scripts\python.exe" main.py
    PID 34068: python.exe
      CMD: "C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.cursortouch.windows-mcp\.venv\Scripts\python.exe" main.py
    PID 10808: node.exe
      CMD: "C:\Program Files\nodejs\node.exe" "C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.ant.figma.figma/server/index.js"
    PID 17892: node.exe
      CMD: "C:\Program Files\nodejs\node.exe" "C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.ant.anthropic.filesystem/server/index.js" D:\dev D:\bin
    PID 36728: claude.exe
      CMD: "C:\Users\hoyth\AppData\Local\AnthropicClaude\app-0.14.4\claude.exe" --type=utility --utility-sub-type=audio.mojom.AudioService --lang=en-US --service-sandbox-type=audio --video-capture-use-gpu-memory-buffer --user-data-dir="C:\Users\hoyth\AppData\Roaming\Claude" --secure-schemes=sentry-ipc --bypasscsp-schemes=sentry-ipc --cors-schemes=sentry-ipc --fetch-schemes=sentry-ipc --field-trial-handle=1800,i,18215422749562086063,15102240429383379184,262144 --enable-features=DocumentPolicyIncludeJSCallStacksInCrashReports,EnableTransparentHwndEnlargement,PdfUseShowSaveFilePicker --disable-features=ScreenAIOCREnabled,SpareRendererForSitePerProcess,WinDelaySpellcheckServiceInit --variations-seed-version --mojo-platform-channel-handle=3028 /prefetch:12
    PID 43988: node.exe
      CMD: "C:\Program Files\nodejs\\node.exe"  "C:\Users\hoyth\AppData\Roaming\npm\node_modules\npm\bin\npm-cli.js" run dev
    PID 21592: node.exe
      CMD: "node"   "D:\dev\ARTIFICIAL_INTELLIGENCE\MCP\MCPManager\frontend\node_modules\.bin\\..\vite\bin\vite.js"
  Matching server: Filesystem (cmd: node)
      Checking process PID 38780:
        Process: claude.exe
        CmdLine: "c:\users\hoyth\.local\bin\claude.exe"
        Server cmd: node, args: [C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.ant.anthropic.filesystem/server/index.js D:\dev,D:\bin]
        ✗ Process name doesn't match server command
        ✗ No match
      Checking process PID 5800:
        Process: claude.exe
        CmdLine: "c:\users\hoyth\appdata\local\anthropicclaude\app-0.14.4\claude.exe"
        Server cmd: node, args: [C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.ant.anthropic.filesystem/server/index.js D:\dev,D:\bin]
        ✗ Process name doesn't match server command
        ✗ No match
      Checking process PID 32692:
        Process: claude.exe
        CmdLine: c:\users\hoyth\appdata\local\anthropicclaude\app-0.14.4\claude.exe --type=crashpad-handler --user-data-dir=c:\users\hoyth\appdata\roaming\claude /prefetch:4 --no-rate-limit --monitor-self-annotation=ptype=crashpad-handler --database=c:\users\hoyth\appdata\roaming\claude\crashpad --url=https://f.a.k/e --annotation=_productname=claude --annotation=_version=0.14.4 --annotation=plat=win64 --annotation=prod=electron --annotation=ver=37.6.0 --initial-client-data=0x4c4,0x4c8,0x4cc,0x4c0,0x4d0,0x7ff6e92621f4,0x7ff6e9262200,0x7ff6e9262210
        Server cmd: node, args: [C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.ant.anthropic.filesystem/server/index.js D:\dev,D:\bin]
        ✗ Process name doesn't match server command
        ✗ No match
      Checking process PID 44652:
        Process: claude.exe
        CmdLine: "c:\users\hoyth\appdata\local\anthropicclaude\app-0.14.4\claude.exe" --type=gpu-process --user-data-dir="c:\users\hoyth\appdata\roaming\claude" --gpu-preferences=uaaaaaaaaadgaaaeaaaaaaaaaaaaaaaaaabgaaeaaaaaaaaaaaaaaaaaaaacaaaaaaaaaaaaaaaaaaaaaaaaabaaaaaaaaaaeaaaaaaaaaaiaaaaaaaaaagaaaaaaaaa --field-trial-handle=1800,i,18215422749562086063,15102240429383379184,262144 --enable-features=documentpolicyincludejscallstacksincrashreports,enabletransparenthwndenlargement,pdfuseshowsavefilepicker --disable-features=screenaiocrenabled,sparerendererforsiteperprocess,windelayspellcheckserviceinit --variations-seed-version --mojo-platform-channel-handle=1796 /prefetch:2
        Server cmd: node, args: [C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.ant.anthropic.filesystem/server/index.js D:\dev,D:\bin]
        ✗ Process name doesn't match server command
        ✗ No match
      Checking process PID 18260:
        Process: claude.exe
        CmdLine: "c:\users\hoyth\appdata\local\anthropicclaude\app-0.14.4\claude.exe" --type=utility --utility-sub-type=network.mojom.networkservice --lang=en-us --service-sandbox-type=none --user-data-dir="c:\users\hoyth\appdata\roaming\claude" --secure-schemes=sentry-ipc --bypasscsp-schemes=sentry-ipc --cors-schemes=sentry-ipc --fetch-schemes=sentry-ipc --field-trial-handle=1800,i,18215422749562086063,15102240429383379184,262144 --enable-features=documentpolicyincludejscallstacksincrashreports,enabletransparenthwndenlargement,pdfuseshowsavefilepicker --disable-features=screenaiocrenabled,sparerendererforsiteperprocess,windelayspellcheckserviceinit --variations-seed-version --mojo-platform-channel-handle=1932 /prefetch:11
        Server cmd: node, args: [C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.ant.anthropic.filesystem/server/index.js D:\dev,D:\bin]
        ✗ Process name doesn't match server command
        ✗ No match
      Checking process PID 19856:
        Process: claude.exe
        CmdLine: "c:\users\hoyth\appdata\local\anthropicclaude\app-0.14.4\claude.exe" --type=renderer --user-data-dir="c:\users\hoyth\appdata\roaming\claude" --secure-schemes=sentry-ipc --bypasscsp-schemes=sentry-ipc --cors-schemes=sentry-ipc --fetch-schemes=sentry-ipc --app-user-model-id=com.squirrel.anthropicclaude.claude --app-path="c:\users\hoyth\appdata\local\anthropicclaude\app-0.14.4\resources\app.asar" --enable-sandbox --video-capture-use-gpu-memory-buffer --lang=en-us --device-scale-factor=1.5 --num-raster-threads=4 --enable-main-frame-before-activation --renderer-client-id=4 --time-ticks-at-unix-epoch=-1760518805885947 --launch-time-ticks=670306037456 --field-trial-handle=1800,i,18215422749562086063,15102240429383379184,262144 --enable-features=documentpolicyincludejscallstacksincrashreports,enabletransparenthwndenlargement,pdfuseshowsavefilepicker --disable-features=screenaiocrenabled,sparerendererforsiteperprocess,windelayspellcheckserviceinit --variations-seed-version --mojo-platform-channel-handle=2252 /prefetch:1
        Server cmd: node, args: [C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.ant.anthropic.filesystem/server/index.js D:\dev,D:\bin]
        ✗ Process name doesn't match server command
        ✗ No match
      Checking process PID 24160:
        Process: claude.exe
        CmdLine: "c:\users\hoyth\appdata\local\anthropicclaude\app-0.14.4\claude.exe" --type=renderer --user-data-dir="c:\users\hoyth\appdata\roaming\claude" --secure-schemes=sentry-ipc --bypasscsp-schemes=sentry-ipc --cors-schemes=sentry-ipc --fetch-schemes=sentry-ipc --app-user-model-id=com.squirrel.anthropicclaude.claude --app-path="c:\users\hoyth\appdata\local\anthropicclaude\app-0.14.4\resources\app.asar" --enable-sandbox --video-capture-use-gpu-memory-buffer --lang=en-us --device-scale-factor=1.5 --num-raster-threads=4 --enable-main-frame-before-activation --renderer-client-id=6 --time-ticks-at-unix-epoch=-1760518805885947 --launch-time-ticks=670306067992 --field-trial-handle=1800,i,18215422749562086063,15102240429383379184,262144 --enable-features=documentpolicyincludejscallstacksincrashreports,enabletransparenthwndenlargement,pdfuseshowsavefilepicker --disable-features=screenaiocrenabled,sparerendererforsiteperprocess,windelayspellcheckserviceinit --variations-seed-version --mojo-platform-channel-handle=2880 /prefetch:1
        Server cmd: node, args: [C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.ant.anthropic.filesystem/server/index.js D:\dev,D:\bin]
        ✗ Process name doesn't match server command
        ✗ No match
      Checking process PID 11500:
        Process: claude.exe
        CmdLine: "c:\users\hoyth\appdata\local\anthropicclaude\app-0.14.4\claude.exe" --type=renderer --user-data-dir="c:\users\hoyth\appdata\roaming\claude" --secure-schemes=sentry-ipc --bypasscsp-schemes=sentry-ipc --cors-schemes=sentry-ipc --fetch-schemes=sentry-ipc --app-user-model-id=com.squirrel.anthropicclaude.claude --app-path="c:\users\hoyth\appdata\local\anthropicclaude\app-0.14.4\resources\app.asar" --enable-sandbox --video-capture-use-gpu-memory-buffer --lang=en-us --device-scale-factor=1.5 --num-raster-threads=4 --enable-main-frame-before-activation --renderer-client-id=7 --time-ticks-at-unix-epoch=-1760518805885947 --launch-time-ticks=670306966351 --field-trial-handle=1800,i,18215422749562086063,15102240429383379184,262144 --enable-features=documentpolicyincludejscallstacksincrashreports,enabletransparenthwndenlargement,pdfuseshowsavefilepicker --disable-features=screenaiocrenabled,sparerendererforsiteperprocess,windelayspellcheckserviceinit --variations-seed-version --mojo-platform-channel-handle=2600 /prefetch:1
        Server cmd: node, args: [C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.ant.anthropic.filesystem/server/index.js D:\dev,D:\bin]
        ✗ Process name doesn't match server command
        ✗ No match
      Checking process PID 28664:
        Process: uv.exe
        CmdLine: "c:\program files\python313\scripts\uv.exe" --directory "c:\users\hoyth\appdata\roaming\claude\claude extensions\ant.dir.cursortouch.windows-mcp" run main.py
        Server cmd: node, args: [C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.ant.anthropic.filesystem/server/index.js D:\dev,D:\bin]
        ✗ Process name doesn't match server command
        ✗ No match
      Checking process PID 22300:
        Process: python.exe
        CmdLine: "c:\users\hoyth\appdata\roaming\claude\claude extensions\ant.dir.cursortouch.windows-mcp\.venv\scripts\python.exe" main.py
        Server cmd: node, args: [C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.ant.anthropic.filesystem/server/index.js D:\dev,D:\bin]
        ✗ Process name doesn't match server command
        ✗ No match
      Checking process PID 34068:
        Process: python.exe
        CmdLine: "c:\users\hoyth\appdata\roaming\claude\claude extensions\ant.dir.cursortouch.windows-mcp\.venv\scripts\python.exe" main.py
        Server cmd: node, args: [C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.ant.anthropic.filesystem/server/index.js D:\dev,D:\bin]
        ✗ Process name doesn't match server command
        ✗ No match
      Checking process PID 10808:
        Process: node.exe
        CmdLine: "c:\program files\nodejs\node.exe" "c:\users\hoyth\appdata\roaming\claude\claude extensions\ant.dir.ant.figma.figma/server/index.js"
        Server cmd: node, args: [C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.ant.anthropic.filesystem/server/index.js D:\dev,D:\bin]
        ✗ No match
      Checking process PID 17892:
        Process: node.exe
        CmdLine: "c:\program files\nodejs\node.exe" "c:\users\hoyth\appdata\roaming\claude\claude extensions\ant.dir.ant.anthropic.filesystem/server/index.js" d:\dev d:\bin
        Server cmd: node, args: [C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.ant.anthropic.filesystem/server/index.js D:\dev,D:\bin]
        ✓ Argument match: C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.ant.anthropic.filesystem/server/index.js
    ✓ MATCHED PID 17892
  Matching server: Figma Dev Mode (cmd: node)
      Checking process PID 38780:
        Process: claude.exe
        CmdLine: "c:\users\hoyth\.local\bin\claude.exe"
        Server cmd: node, args: [C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.ant.figma.figma/server/index.js]
        ✗ Process name doesn't match server command
        ✗ No match
      Checking process PID 5800:
        Process: claude.exe
        CmdLine: "c:\users\hoyth\appdata\local\anthropicclaude\app-0.14.4\claude.exe"
        Server cmd: node, args: [C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.ant.figma.figma/server/index.js]
        ✗ Process name doesn't match server command
        ✗ No match
      Checking process PID 32692:
        Process: claude.exe
        CmdLine: c:\users\hoyth\appdata\local\anthropicclaude\app-0.14.4\claude.exe --type=crashpad-handler --user-data-dir=c:\users\hoyth\appdata\roaming\claude /prefetch:4 --no-rate-limit --monitor-self-annotation=ptype=crashpad-handler --database=c:\users\hoyth\appdata\roaming\claude\crashpad --url=https://f.a.k/e --annotation=_productname=claude --annotation=_version=0.14.4 --annotation=plat=win64 --annotation=prod=electron --annotation=ver=37.6.0 --initial-client-data=0x4c4,0x4c8,0x4cc,0x4c0,0x4d0,0x7ff6e92621f4,0x7ff6e9262200,0x7ff6e9262210
        Server cmd: node, args: [C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.ant.figma.figma/server/index.js]
        ✗ Process name doesn't match server command
        ✗ No match
      Checking process PID 44652:
        Process: claude.exe
        CmdLine: "c:\users\hoyth\appdata\local\anthropicclaude\app-0.14.4\claude.exe" --type=gpu-process --user-data-dir="c:\users\hoyth\appdata\roaming\claude" --gpu-preferences=uaaaaaaaaadgaaaeaaaaaaaaaaaaaaaaaabgaaeaaaaaaaaaaaaaaaaaaaacaaaaaaaaaaaaaaaaaaaaaaaaabaaaaaaaaaaeaaaaaaaaaaiaaaaaaaaaagaaaaaaaaa --field-trial-handle=1800,i,18215422749562086063,15102240429383379184,262144 --enable-features=documentpolicyincludejscallstacksincrashreports,enabletransparenthwndenlargement,pdfuseshowsavefilepicker --disable-features=screenaiocrenabled,sparerendererforsiteperprocess,windelayspellcheckserviceinit --variations-seed-version --mojo-platform-channel-handle=1796 /prefetch:2
        Server cmd: node, args: [C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.ant.figma.figma/server/index.js]
        ✗ Process name doesn't match server command
        ✗ No match
      Checking process PID 18260:
        Process: claude.exe
        CmdLine: "c:\users\hoyth\appdata\local\anthropicclaude\app-0.14.4\claude.exe" --type=utility --utility-sub-type=network.mojom.networkservice --lang=en-us --service-sandbox-type=none --user-data-dir="c:\users\hoyth\appdata\roaming\claude" --secure-schemes=sentry-ipc --bypasscsp-schemes=sentry-ipc --cors-schemes=sentry-ipc --fetch-schemes=sentry-ipc --field-trial-handle=1800,i,18215422749562086063,15102240429383379184,262144 --enable-features=documentpolicyincludejscallstacksincrashreports,enabletransparenthwndenlargement,pdfuseshowsavefilepicker --disable-features=screenaiocrenabled,sparerendererforsiteperprocess,windelayspellcheckserviceinit --variations-seed-version --mojo-platform-channel-handle=1932 /prefetch:11
        Server cmd: node, args: [C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.ant.figma.figma/server/index.js]
        ✗ Process name doesn't match server command
        ✗ No match
      Checking process PID 19856:
        Process: claude.exe
        CmdLine: "c:\users\hoyth\appdata\local\anthropicclaude\app-0.14.4\claude.exe" --type=renderer --user-data-dir="c:\users\hoyth\appdata\roaming\claude" --secure-schemes=sentry-ipc --bypasscsp-schemes=sentry-ipc --cors-schemes=sentry-ipc --fetch-schemes=sentry-ipc --app-user-model-id=com.squirrel.anthropicclaude.claude --app-path="c:\users\hoyth\appdata\local\anthropicclaude\app-0.14.4\resources\app.asar" --enable-sandbox --video-capture-use-gpu-memory-buffer --lang=en-us --device-scale-factor=1.5 --num-raster-threads=4 --enable-main-frame-before-activation --renderer-client-id=4 --time-ticks-at-unix-epoch=-1760518805885947 --launch-time-ticks=670306037456 --field-trial-handle=1800,i,18215422749562086063,15102240429383379184,262144 --enable-features=documentpolicyincludejscallstacksincrashreports,enabletransparenthwndenlargement,pdfuseshowsavefilepicker --disable-features=screenaiocrenabled,sparerendererforsiteperprocess,windelayspellcheckserviceinit --variations-seed-version --mojo-platform-channel-handle=2252 /prefetch:1
        Server cmd: node, args: [C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.ant.figma.figma/server/index.js]
        ✗ Process name doesn't match server command
        ✗ No match
      Checking process PID 24160:
        Process: claude.exe
        CmdLine: "c:\users\hoyth\appdata\local\anthropicclaude\app-0.14.4\claude.exe" --type=renderer --user-data-dir="c:\users\hoyth\appdata\roaming\claude" --secure-schemes=sentry-ipc --bypasscsp-schemes=sentry-ipc --cors-schemes=sentry-ipc --fetch-schemes=sentry-ipc --app-user-model-id=com.squirrel.anthropicclaude.claude --app-path="c:\users\hoyth\appdata\local\anthropicclaude\app-0.14.4\resources\app.asar" --enable-sandbox --video-capture-use-gpu-memory-buffer --lang=en-us --device-scale-factor=1.5 --num-raster-threads=4 --enable-main-frame-before-activation --renderer-client-id=6 --time-ticks-at-unix-epoch=-1760518805885947 --launch-time-ticks=670306067992 --field-trial-handle=1800,i,18215422749562086063,15102240429383379184,262144 --enable-features=documentpolicyincludejscallstacksincrashreports,enabletransparenthwndenlargement,pdfuseshowsavefilepicker --disable-features=screenaiocrenabled,sparerendererforsiteperprocess,windelayspellcheckserviceinit --variations-seed-version --mojo-platform-channel-handle=2880 /prefetch:1
        Server cmd: node, args: [C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.ant.figma.figma/server/index.js]
        ✗ Process name doesn't match server command
        ✗ No match
      Checking process PID 11500:
        Process: claude.exe
        CmdLine: "c:\users\hoyth\appdata\local\anthropicclaude\app-0.14.4\claude.exe" --type=renderer --user-data-dir="c:\users\hoyth\appdata\roaming\claude" --secure-schemes=sentry-ipc --bypasscsp-schemes=sentry-ipc --cors-schemes=sentry-ipc --fetch-schemes=sentry-ipc --app-user-model-id=com.squirrel.anthropicclaude.claude --app-path="c:\users\hoyth\appdata\local\anthropicclaude\app-0.14.4\resources\app.asar" --enable-sandbox --video-capture-use-gpu-memory-buffer --lang=en-us --device-scale-factor=1.5 --num-raster-threads=4 --enable-main-frame-before-activation --renderer-client-id=7 --time-ticks-at-unix-epoch=-1760518805885947 --launch-time-ticks=670306966351 --field-trial-handle=1800,i,18215422749562086063,15102240429383379184,262144 --enable-features=documentpolicyincludejscallstacksincrashreports,enabletransparenthwndenlargement,pdfuseshowsavefilepicker --disable-features=screenaiocrenabled,sparerendererforsiteperprocess,windelayspellcheckserviceinit --variations-seed-version --mojo-platform-channel-handle=2600 /prefetch:1
        Server cmd: node, args: [C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.ant.figma.figma/server/index.js]
        ✗ Process name doesn't match server command
        ✗ No match
      Checking process PID 28664:
        Process: uv.exe
        CmdLine: "c:\program files\python313\scripts\uv.exe" --directory "c:\users\hoyth\appdata\roaming\claude\claude extensions\ant.dir.cursortouch.windows-mcp" run main.py
        Server cmd: node, args: [C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.ant.figma.figma/server/index.js]
        ✗ Process name doesn't match server command
        ✗ No match
      Checking process PID 22300:
        Process: python.exe
        CmdLine: "c:\users\hoyth\appdata\roaming\claude\claude extensions\ant.dir.cursortouch.windows-mcp\.venv\scripts\python.exe" main.py
        Server cmd: node, args: [C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.ant.figma.figma/server/index.js]
        ✗ Process name doesn't match server command
        ✗ No match
      Checking process PID 34068:
        Process: python.exe
        CmdLine: "c:\users\hoyth\appdata\roaming\claude\claude extensions\ant.dir.cursortouch.windows-mcp\.venv\scripts\python.exe" main.py
        Server cmd: node, args: [C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.ant.figma.figma/server/index.js]
        ✗ Process name doesn't match server command
        ✗ No match
      Checking process PID 10808:
        Process: node.exe
        CmdLine: "c:\program files\nodejs\node.exe" "c:\users\hoyth\appdata\roaming\claude\claude extensions\ant.dir.ant.figma.figma/server/index.js"
        Server cmd: node, args: [C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.ant.figma.figma/server/index.js]
        ✓ Argument match: C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.ant.figma.figma/server/index.js
    ✓ MATCHED PID 10808
  Matching server: Windows-MCP (cmd: uv)
      Checking process PID 38780:
        Process: claude.exe
        CmdLine: "c:\users\hoyth\.local\bin\claude.exe"
        Server cmd: uv, args: [--directory C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.cursortouch.windows-mcp run main.py]
        ✗ Process name doesn't match server command
        ✗ No match
      Checking process PID 5800:
        Process: claude.exe
        CmdLine: "c:\users\hoyth\appdata\local\anthropicclaude\app-0.14.4\claude.exe"
        Server cmd: uv, args: [--directory C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.cursortouch.windows-mcp run main.py]
        ✗ Process name doesn't match server command
        ✗ No match
      Checking process PID 32692:
        Process: claude.exe
        CmdLine: c:\users\hoyth\appdata\local\anthropicclaude\app-0.14.4\claude.exe --type=crashpad-handler --user-data-dir=c:\users\hoyth\appdata\roaming\claude /prefetch:4 --no-rate-limit --monitor-self-annotation=ptype=crashpad-handler --database=c:\users\hoyth\appdata\roaming\claude\crashpad --url=https://f.a.k/e --annotation=_productname=claude --annotation=_version=0.14.4 --annotation=plat=win64 --annotation=prod=electron --annotation=ver=37.6.0 --initial-client-data=0x4c4,0x4c8,0x4cc,0x4c0,0x4d0,0x7ff6e92621f4,0x7ff6e9262200,0x7ff6e9262210
        Server cmd: uv, args: [--directory C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.cursortouch.windows-mcp run main.py]
        ✗ Process name doesn't match server command
        ✗ No match
      Checking process PID 44652:
        Process: claude.exe
        CmdLine: "c:\users\hoyth\appdata\local\anthropicclaude\app-0.14.4\claude.exe" --type=gpu-process --user-data-dir="c:\users\hoyth\appdata\roaming\claude" --gpu-preferences=uaaaaaaaaadgaaaeaaaaaaaaaaaaaaaaaabgaaeaaaaaaaaaaaaaaaaaaaacaaaaaaaaaaaaaaaaaaaaaaaaabaaaaaaaaaaeaaaaaaaaaaiaaaaaaaaaagaaaaaaaaa --field-trial-handle=1800,i,18215422749562086063,15102240429383379184,262144 --enable-features=documentpolicyincludejscallstacksincrashreports,enabletransparenthwndenlargement,pdfuseshowsavefilepicker --disable-features=screenaiocrenabled,sparerendererforsiteperprocess,windelayspellcheckserviceinit --variations-seed-version --mojo-platform-channel-handle=1796 /prefetch:2
        Server cmd: uv, args: [--directory C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.cursortouch.windows-mcp run main.py]
        ✗ Process name doesn't match server command
        ✗ No match
      Checking process PID 18260:
        Process: claude.exe
        CmdLine: "c:\users\hoyth\appdata\local\anthropicclaude\app-0.14.4\claude.exe" --type=utility --utility-sub-type=network.mojom.networkservice --lang=en-us --service-sandbox-type=none --user-data-dir="c:\users\hoyth\appdata\roaming\claude" --secure-schemes=sentry-ipc --bypasscsp-schemes=sentry-ipc --cors-schemes=sentry-ipc --fetch-schemes=sentry-ipc --field-trial-handle=1800,i,18215422749562086063,15102240429383379184,262144 --enable-features=documentpolicyincludejscallstacksincrashreports,enabletransparenthwndenlargement,pdfuseshowsavefilepicker --disable-features=screenaiocrenabled,sparerendererforsiteperprocess,windelayspellcheckserviceinit --variations-seed-version --mojo-platform-channel-handle=1932 /prefetch:11
        Server cmd: uv, args: [--directory C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.cursortouch.windows-mcp run main.py]
        ✗ Process name doesn't match server command
        ✗ No match
      Checking process PID 19856:
        Process: claude.exe
        CmdLine: "c:\users\hoyth\appdata\local\anthropicclaude\app-0.14.4\claude.exe" --type=renderer --user-data-dir="c:\users\hoyth\appdata\roaming\claude" --secure-schemes=sentry-ipc --bypasscsp-schemes=sentry-ipc --cors-schemes=sentry-ipc --fetch-schemes=sentry-ipc --app-user-model-id=com.squirrel.anthropicclaude.claude --app-path="c:\users\hoyth\appdata\local\anthropicclaude\app-0.14.4\resources\app.asar" --enable-sandbox --video-capture-use-gpu-memory-buffer --lang=en-us --device-scale-factor=1.5 --num-raster-threads=4 --enable-main-frame-before-activation --renderer-client-id=4 --time-ticks-at-unix-epoch=-1760518805885947 --launch-time-ticks=670306037456 --field-trial-handle=1800,i,18215422749562086063,15102240429383379184,262144 --enable-features=documentpolicyincludejscallstacksincrashreports,enabletransparenthwndenlargement,pdfuseshowsavefilepicker --disable-features=screenaiocrenabled,sparerendererforsiteperprocess,windelayspellcheckserviceinit --variations-seed-version --mojo-platform-channel-handle=2252 /prefetch:1
        Server cmd: uv, args: [--directory C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.cursortouch.windows-mcp run main.py]
        ✗ Process name doesn't match server command
        ✗ No match
      Checking process PID 24160:
        Process: claude.exe
        CmdLine: "c:\users\hoyth\appdata\local\anthropicclaude\app-0.14.4\claude.exe" --type=renderer --user-data-dir="c:\users\hoyth\appdata\roaming\claude" --secure-schemes=sentry-ipc --bypasscsp-schemes=sentry-ipc --cors-schemes=sentry-ipc --fetch-schemes=sentry-ipc --app-user-model-id=com.squirrel.anthropicclaude.claude --app-path="c:\users\hoyth\appdata\local\anthropicclaude\app-0.14.4\resources\app.asar" --enable-sandbox --video-capture-use-gpu-memory-buffer --lang=en-us --device-scale-factor=1.5 --num-raster-threads=4 --enable-main-frame-before-activation --renderer-client-id=6 --time-ticks-at-unix-epoch=-1760518805885947 --launch-time-ticks=670306067992 --field-trial-handle=1800,i,18215422749562086063,15102240429383379184,262144 --enable-features=documentpolicyincludejscallstacksincrashreports,enabletransparenthwndenlargement,pdfuseshowsavefilepicker --disable-features=screenaiocrenabled,sparerendererforsiteperprocess,windelayspellcheckserviceinit --variations-seed-version --mojo-platform-channel-handle=2880 /prefetch:1
        Server cmd: uv, args: [--directory C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.cursortouch.windows-mcp run main.py]
        ✗ Process name doesn't match server command
        ✗ No match
      Checking process PID 11500:
        Process: claude.exe
        CmdLine: "c:\users\hoyth\appdata\local\anthropicclaude\app-0.14.4\claude.exe" --type=renderer --user-data-dir="c:\users\hoyth\appdata\roaming\claude" --secure-schemes=sentry-ipc --bypasscsp-schemes=sentry-ipc --cors-schemes=sentry-ipc --fetch-schemes=sentry-ipc --app-user-model-id=com.squirrel.anthropicclaude.claude --app-path="c:\users\hoyth\appdata\local\anthropicclaude\app-0.14.4\resources\app.asar" --enable-sandbox --video-capture-use-gpu-memory-buffer --lang=en-us --device-scale-factor=1.5 --num-raster-threads=4 --enable-main-frame-before-activation --renderer-client-id=7 --time-ticks-at-unix-epoch=-1760518805885947 --launch-time-ticks=670306966351 --field-trial-handle=1800,i,18215422749562086063,15102240429383379184,262144 --enable-features=documentpolicyincludejscallstacksincrashreports,enabletransparenthwndenlargement,pdfuseshowsavefilepicker --disable-features=screenaiocrenabled,sparerendererforsiteperprocess,windelayspellcheckserviceinit --variations-seed-version --mojo-platform-channel-handle=2600 /prefetch:1
        Server cmd: uv, args: [--directory C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.cursortouch.windows-mcp run main.py]
        ✗ Process name doesn't match server command
        ✗ No match
      Checking process PID 28664:
        Process: uv.exe
        CmdLine: "c:\program files\python313\scripts\uv.exe" --directory "c:\users\hoyth\appdata\roaming\claude\claude extensions\ant.dir.cursortouch.windows-mcp" run main.py
        Server cmd: uv, args: [--directory C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.cursortouch.windows-mcp run main.py]
        ✓ Argument match: C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.cursortouch.windows-mcp
    ✓ MATCHED PID 28664
[PHASE 3] Matched 3 running processes

=== DISCOVERY COMPLETE: 3 total servers ===

time=2025-10-22T22:35:55.035-05:00 level=INFO msg="StopServer called" serverId=73054ce2-68a8-40ed-9bbf-ffb75e521d53 force=false timeout=10
```

3. Task Manager rsults:
  - PID 38832 is not running.
  - PIDs 10808, 17892, 21592, and 43988 (node.exe) are running.
