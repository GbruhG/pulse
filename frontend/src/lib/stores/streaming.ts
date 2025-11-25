// Streaming protocol management store
import { writable } from 'svelte/store';

export type StreamingProtocol = 'websocket' | 'sse' | 'grpc-stream' | 'kafka' | 'mqtt';

export interface StreamingRequest {
    protocol: StreamingProtocol;
    url: string;
    config: Record<string, any>;
}

function createStreamingStore() {
    const { subscribe, set, update } = writable<StreamingRequest>({
        protocol: 'websocket',
        url: '',
        config: {}
    });

    return {
        subscribe,
        setRequest: (request: StreamingRequest) => set(request),
        updateProtocol: (protocol: StreamingProtocol) => update(state => ({ ...state, protocol })),
        updateUrl: (url: string) => update(state => ({ ...state, url })),
        updateConfig: (config: Record<string, any>) => update(state => ({ ...state, config })),
        reset: () => set({ protocol: 'websocket', url: '', config: {} })
    };
}

export const streamingStore = createStreamingStore();