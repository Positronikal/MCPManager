export namespace dependencies {
	
	export class UpdateInfo {
	    updateAvailable: boolean;
	    status: string;
	    currentVersion: string;
	    latestVersion: string;
	    releaseNotes?: string;
	    packageName: string;
	    packageType: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.updateAvailable = source["updateAvailable"];
	        this.status = source["status"];
	        this.currentVersion = source["currentVersion"];
	        this.latestVersion = source["latestVersion"];
	        this.releaseNotes = source["releaseNotes"];
	        this.packageName = source["packageName"];
	        this.packageType = source["packageType"];
	    }
	}

}

export namespace main {
	
	export class DiscoverServersResponse {
	    message: string;
	    scanId: string;
	
	    static createFrom(source: any = {}) {
	        return new DiscoverServersResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.message = source["message"];
	        this.scanId = source["scanId"];
	    }
	}
	export class GetDependenciesResponse {
	    dependencies: models.Dependency[];
	    allSatisfied: boolean;
	
	    static createFrom(source: any = {}) {
	        return new GetDependenciesResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.dependencies = this.convertValues(source["dependencies"], models.Dependency);
	        this.allSatisfied = source["allSatisfied"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class GetLogsResponse {
	    logs: models.LogEntry[];
	    total: number;
	    hasMore: boolean;
	
	    static createFrom(source: any = {}) {
	        return new GetLogsResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.logs = this.convertValues(source["logs"], models.LogEntry);
	        this.total = source["total"];
	        this.hasMore = source["hasMore"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ListServersResponse {
	    servers: models.MCPServer[];
	    count: number;
	    lastDiscovery: string;
	
	    static createFrom(source: any = {}) {
	        return new ListServersResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.servers = this.convertValues(source["servers"], models.MCPServer);
	        this.count = source["count"];
	        this.lastDiscovery = source["lastDiscovery"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ServerOperationResponse {
	    message: string;
	    serverId: string;
	    status?: string;
	
	    static createFrom(source: any = {}) {
	        return new ServerOperationResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.message = source["message"];
	        this.serverId = source["serverId"];
	        this.status = source["status"];
	    }
	}
	export class UpdateApplicationStateResponse {
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateApplicationStateResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.message = source["message"];
	    }
	}

}

export namespace models {
	
	export class Filters {
	    selectedServer?: string;
	    selectedSeverity?: string;
	    searchQuery?: string;
	
	    static createFrom(source: any = {}) {
	        return new Filters(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.selectedServer = source["selectedServer"];
	        this.selectedSeverity = source["selectedSeverity"];
	        this.searchQuery = source["searchQuery"];
	    }
	}
	export class WindowLayout {
	    width: number;
	    height: number;
	    x: number;
	    y: number;
	    maximized: boolean;
	    logPanelHeight: number;
	
	    static createFrom(source: any = {}) {
	        return new WindowLayout(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.width = source["width"];
	        this.height = source["height"];
	        this.x = source["x"];
	        this.y = source["y"];
	        this.maximized = source["maximized"];
	        this.logPanelHeight = source["logPanelHeight"];
	    }
	}
	export class UserPreferences {
	    theme: string;
	    logRetentionPerServer: number;
	    autoStartServers: boolean;
	    minimizeToTray: boolean;
	    showNotifications: boolean;
	
	    static createFrom(source: any = {}) {
	        return new UserPreferences(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.theme = source["theme"];
	        this.logRetentionPerServer = source["logRetentionPerServer"];
	        this.autoStartServers = source["autoStartServers"];
	        this.minimizeToTray = source["minimizeToTray"];
	        this.showNotifications = source["showNotifications"];
	    }
	}
	export class ApplicationState {
	    version: string;
	    // Go type: time
	    lastSaved: any;
	    preferences: UserPreferences;
	    windowLayout: WindowLayout;
	    filters: Filters;
	    discoveredServers: string[];
	    monitoredConfigPaths: string[];
	    // Go type: time
	    lastDiscoveryScan: any;
	
	    static createFrom(source: any = {}) {
	        return new ApplicationState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.lastSaved = this.convertValues(source["lastSaved"], null);
	        this.preferences = this.convertValues(source["preferences"], UserPreferences);
	        this.windowLayout = this.convertValues(source["windowLayout"], WindowLayout);
	        this.filters = this.convertValues(source["filters"], Filters);
	        this.discoveredServers = source["discoveredServers"];
	        this.monitoredConfigPaths = source["monitoredConfigPaths"];
	        this.lastDiscoveryScan = this.convertValues(source["lastDiscoveryScan"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Dependency {
	    name: string;
	    type: string;
	    requiredVersion?: string;
	    detectedVersion?: string;
	    installationInstructions?: string;
	
	    static createFrom(source: any = {}) {
	        return new Dependency(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.type = source["type"];
	        this.requiredVersion = source["requiredVersion"];
	        this.detectedVersion = source["detectedVersion"];
	        this.installationInstructions = source["installationInstructions"];
	    }
	}
	
	export class LogEntry {
	    id: string;
	    // Go type: time
	    timestamp: any;
	    severity: string;
	    source: string;
	    message: string;
	    metadata?: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new LogEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.timestamp = this.convertValues(source["timestamp"], null);
	        this.severity = source["severity"];
	        this.source = source["source"];
	        this.message = source["message"];
	        this.metadata = source["metadata"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ServerConfiguration {
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
	
	    static createFrom(source: any = {}) {
	        return new ServerConfiguration(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.environmentVariables = source["environmentVariables"];
	        this.commandLineArguments = source["commandLineArguments"];
	        this.workingDirectory = source["workingDirectory"];
	        this.autoStart = source["autoStart"];
	        this.restartOnCrash = source["restartOnCrash"];
	        this.maxRestartAttempts = source["maxRestartAttempts"];
	        this.startupTimeout = source["startupTimeout"];
	        this.shutdownTimeout = source["shutdownTimeout"];
	        this.healthCheckInterval = source["healthCheckInterval"];
	        this.healthCheckEndpoint = source["healthCheckEndpoint"];
	    }
	}
	export class ServerStatus {
	    state: string;
	    startupAttempts: number;
	    // Go type: time
	    lastStateChange: any;
	    errorMessage?: string;
	    crashRecoverable: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ServerStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.state = source["state"];
	        this.startupAttempts = source["startupAttempts"];
	        this.lastStateChange = this.convertValues(source["lastStateChange"], null);
	        this.errorMessage = source["errorMessage"];
	        this.crashRecoverable = source["crashRecoverable"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class MCPServer {
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
	    // Go type: time
	    discoveredAt: any;
	    // Go type: time
	    lastSeenAt: any;
	    source: string;
	
	    static createFrom(source: any = {}) {
	        return new MCPServer(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.version = source["version"];
	        this.installationPath = source["installationPath"];
	        this.status = this.convertValues(source["status"], ServerStatus);
	        this.pid = source["pid"];
	        this.capabilities = source["capabilities"];
	        this.tools = source["tools"];
	        this.configuration = this.convertValues(source["configuration"], ServerConfiguration);
	        this.dependencies = this.convertValues(source["dependencies"], Dependency);
	        this.discoveredAt = this.convertValues(source["discoveredAt"], null);
	        this.lastSeenAt = this.convertValues(source["lastSeenAt"], null);
	        this.source = source["source"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class ServerMetrics {
	    serverId: string;
	    uptime: number;
	    memoryBytes?: number;
	    requestCount?: number;
	    // Go type: time
	    timestamp: any;
	
	    static createFrom(source: any = {}) {
	        return new ServerMetrics(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.serverId = source["serverId"];
	        this.uptime = source["uptime"];
	        this.memoryBytes = source["memoryBytes"];
	        this.requestCount = source["requestCount"];
	        this.timestamp = this.convertValues(source["timestamp"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	

}

