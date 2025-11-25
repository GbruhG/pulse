export namespace backend {
	
	export class RequestAuth {
	    type: string;
	    username: string;
	    password: string;
	    token: string;
	    key: string;
	    value: string;
	
	    static createFrom(source: any = {}) {
	        return new RequestAuth(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.username = source["username"];
	        this.password = source["password"];
	        this.token = source["token"];
	        this.key = source["key"];
	        this.value = source["value"];
	    }
	}
	export class KeyValue {
	    id: string;
	    key: string;
	    value: string;
	    enabled: boolean;
	    description: string;
	
	    static createFrom(source: any = {}) {
	        return new KeyValue(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.key = source["key"];
	        this.value = source["value"];
	        this.enabled = source["enabled"];
	        this.description = source["description"];
	    }
	}
	export class RequestData {
	    method: string;
	    url: string;
	    params: KeyValue[];
	    headers: KeyValue[];
	    body: string;
	    bodyType: string;
	    auth?: RequestAuth;
	
	    static createFrom(source: any = {}) {
	        return new RequestData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.method = source["method"];
	        this.url = source["url"];
	        this.params = this.convertValues(source["params"], KeyValue);
	        this.headers = this.convertValues(source["headers"], KeyValue);
	        this.body = source["body"];
	        this.bodyType = source["bodyType"];
	        this.auth = this.convertValues(source["auth"], RequestAuth);
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
	export class CollectionRequest {
	    id: string;
	    name: string;
	    collectionId: string;
	    request: RequestData;
	
	    static createFrom(source: any = {}) {
	        return new CollectionRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.collectionId = source["collectionId"];
	        this.request = this.convertValues(source["request"], RequestData);
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
	export class Collection {
	    id: string;
	    name: string;
	    workspaceId: string;
	    requests: CollectionRequest[];
	    // Go type: time
	    createdAt: any;
	
	    static createFrom(source: any = {}) {
	        return new Collection(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.workspaceId = source["workspaceId"];
	        this.requests = this.convertValues(source["requests"], CollectionRequest);
	        this.createdAt = this.convertValues(source["createdAt"], null);
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
	
	export class ConsumerConfig {
	    connectionId: string;
	    topic: string;
	    partitions: number[];
	    consumerGroup: string;
	    offsetStrategy: string;
	    customOffset: number;
	    autoCommit: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ConsumerConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.connectionId = source["connectionId"];
	        this.topic = source["topic"];
	        this.partitions = source["partitions"];
	        this.consumerGroup = source["consumerGroup"];
	        this.offsetStrategy = source["offsetStrategy"];
	        this.customOffset = source["customOffset"];
	        this.autoCommit = source["autoCommit"];
	    }
	}
	export class Environment {
	    id: string;
	    name: string;
	    variables: Record<string, string>;
	    workspaceId: string;
	
	    static createFrom(source: any = {}) {
	        return new Environment(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.variables = source["variables"];
	        this.workspaceId = source["workspaceId"];
	    }
	}
	export class GrpcConnectRequest {
	    serverUrl: string;
	    service: string;
	    method: string;
	    useTLS: boolean;
	    deadline: number;
	    compression: string;
	    metadata: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new GrpcConnectRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.serverUrl = source["serverUrl"];
	        this.service = source["service"];
	        this.method = source["method"];
	        this.useTLS = source["useTLS"];
	        this.deadline = source["deadline"];
	        this.compression = source["compression"];
	        this.metadata = source["metadata"];
	    }
	}
	export class GrpcSendMessageRequest {
	    connectionId: string;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new GrpcSendMessageRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.connectionId = source["connectionId"];
	        this.message = source["message"];
	    }
	}
	export class ResponseData {
	    statusCode: number;
	    statusText: string;
	    headers: Record<string, string>;
	    body: string;
	
	    static createFrom(source: any = {}) {
	        return new ResponseData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.statusCode = source["statusCode"];
	        this.statusText = source["statusText"];
	        this.headers = source["headers"];
	        this.body = source["body"];
	    }
	}
	export class HistoryItem {
	    id: string;
	    request: RequestData;
	    response?: ResponseData;
	    // Go type: time
	    timestamp: any;
	    workspaceId: string;
	
	    static createFrom(source: any = {}) {
	        return new HistoryItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.request = this.convertValues(source["request"], RequestData);
	        this.response = this.convertValues(source["response"], ResponseData);
	        this.timestamp = this.convertValues(source["timestamp"], null);
	        this.workspaceId = source["workspaceId"];
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
	export class KafkaConfig {
	    bootstrapServers: string[];
	    clientId: string;
	    authMechanism: string;
	    saslUsername: string;
	    saslPassword: string;
	    useTLS: boolean;
	    tlsSkipVerify: boolean;
	    connectionTimeout: number;
	
	    static createFrom(source: any = {}) {
	        return new KafkaConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.bootstrapServers = source["bootstrapServers"];
	        this.clientId = source["clientId"];
	        this.authMechanism = source["authMechanism"];
	        this.saslUsername = source["saslUsername"];
	        this.saslPassword = source["saslPassword"];
	        this.useTLS = source["useTLS"];
	        this.tlsSkipVerify = source["tlsSkipVerify"];
	        this.connectionTimeout = source["connectionTimeout"];
	    }
	}
	
	export class MethodInfo {
	    name: string;
	    type: string;
	    inputType: string;
	    outputType: string;
	
	    static createFrom(source: any = {}) {
	        return new MethodInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.type = source["type"];
	        this.inputType = source["inputType"];
	        this.outputType = source["outputType"];
	    }
	}
	export class ServiceInfo {
	    name: string;
	    methods: MethodInfo[];
	
	    static createFrom(source: any = {}) {
	        return new ServiceInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.methods = this.convertValues(source["methods"], MethodInfo);
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
	export class ParsedProtoResponse {
	    services: ServiceInfo[];
	
	    static createFrom(source: any = {}) {
	        return new ParsedProtoResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.services = this.convertValues(source["services"], ServiceInfo);
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
	export class ProducerConfig {
	    connectionId: string;
	    topic: string;
	    partition: number;
	    key: string;
	    value: string;
	    headers: Record<string, string>;
	    compression: string;
	    acks: number;
	
	    static createFrom(source: any = {}) {
	        return new ProducerConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.connectionId = source["connectionId"];
	        this.topic = source["topic"];
	        this.partition = source["partition"];
	        this.key = source["key"];
	        this.value = source["value"];
	        this.headers = source["headers"];
	        this.compression = source["compression"];
	        this.acks = source["acks"];
	    }
	}
	export class ProtoFile {
	    name: string;
	    content: string;
	
	    static createFrom(source: any = {}) {
	        return new ProtoFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.content = source["content"];
	    }
	}
	export class ProtoFileUploadRequest {
	    files: ProtoFile[];
	
	    static createFrom(source: any = {}) {
	        return new ProtoFileUploadRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.files = this.convertValues(source["files"], ProtoFile);
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
	
	
	
	export class SSEConnectRequest {
	    url: string;
	    withCredentials: boolean;
	    retryTimeout: number;
	    lastEventId: string;
	    autoReconnect: boolean;
	    customHeaders: Record<string, string>;
	    eventTypeFilter: string[];
	
	    static createFrom(source: any = {}) {
	        return new SSEConnectRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.url = source["url"];
	        this.withCredentials = source["withCredentials"];
	        this.retryTimeout = source["retryTimeout"];
	        this.lastEventId = source["lastEventId"];
	        this.autoReconnect = source["autoReconnect"];
	        this.customHeaders = source["customHeaders"];
	        this.eventTypeFilter = source["eventTypeFilter"];
	    }
	}
	
	export class Settings {
	    uiScale: number;
	    theme: string;
	    layoutMode: string;
	    autoSaveHistory: boolean;
	    maxHistoryItems: number;
	    defaultTimeout: number;
	    prettyPrintByDefault: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.uiScale = source["uiScale"];
	        this.theme = source["theme"];
	        this.layoutMode = source["layoutMode"];
	        this.autoSaveHistory = source["autoSaveHistory"];
	        this.maxHistoryItems = source["maxHistoryItems"];
	        this.defaultTimeout = source["defaultTimeout"];
	        this.prettyPrintByDefault = source["prettyPrintByDefault"];
	    }
	}
	export class TopicInfo {
	    name: string;
	    partitions: number;
	
	    static createFrom(source: any = {}) {
	        return new TopicInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.partitions = source["partitions"];
	    }
	}
	export class WebSocketConnectRequest {
	    url: string;
	    subprotocol: string;
	    autoReconnect: boolean;
	    reconnectInterval: number;
	    enablePingPong: boolean;
	    pingInterval: number;
	    customHeaders: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new WebSocketConnectRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.url = source["url"];
	        this.subprotocol = source["subprotocol"];
	        this.autoReconnect = source["autoReconnect"];
	        this.reconnectInterval = source["reconnectInterval"];
	        this.enablePingPong = source["enablePingPong"];
	        this.pingInterval = source["pingInterval"];
	        this.customHeaders = source["customHeaders"];
	    }
	}
	export class WebSocketSendRequest {
	    connectionId: string;
	    message: string;
	    messageType: string;
	
	    static createFrom(source: any = {}) {
	        return new WebSocketSendRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.connectionId = source["connectionId"];
	        this.message = source["message"];
	        this.messageType = source["messageType"];
	    }
	}
	export class Workspace {
	    id: string;
	    name: string;
	    // Go type: time
	    createdAt: any;
	
	    static createFrom(source: any = {}) {
	        return new Workspace(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
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

