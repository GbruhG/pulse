// Type definitions for Pulse API testing tool
// Defines the interfaces for all data structures used throughout the app

export interface Workspace {
    id: string;
    name: string;
    collections: Collection[];
    environments: Environment[];
    createdAt: Date;
}

export interface Environment {
    id: string;
    name: string;
    variables: Record<string, string>;
    workspaceId: string;
}

export interface Collection {
    id: string;
    name: string;
    requests: CollectionRequest[];
    workspaceId: string;
    createdAt: Date;
}

export interface CollectionRequest {
    id: string;
    name: string;
    collectionId: string;
    request: RequestData;
}

export interface RequestData {
    method: string;
    url: string;
    params: KeyValue[];
    headers: KeyValue[];
    body: string;
    bodyType: 'none' | 'json' | 'xml' | 'text' | 'form-data' | 'x-www-form-urlencoded';
    auth: RequestAuth | null;
}

export interface RequestAuth {
    type: 'none' | 'basic' | 'bearer' | 'api-key' | 'oauth2';
    username?: string;
    password?: string;
    token?: string;
    key?: string;
    value?: string;
}

export interface KeyValue {
    id: string;
    key: string;
    value: string;
    enabled: boolean;
    description?: string;
}

export interface ResponseData {
    statusCode: number;
    statusText: string;
    headers: Record<string, string>;
    body: string;
    time: string;
    size: string;
}

export interface HistoryItem {
    id: string;
    request: RequestData;
    response: ResponseData | null;
    timestamp: Date;
    workspaceId: string;
}

export type HTTPMethod = 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE' | 'HEAD' | 'OPTIONS';

export const HTTP_METHODS: HTTPMethod[] = ['GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'HEAD', 'OPTIONS'];