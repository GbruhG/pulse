// Tab management store
import { writable, derived, get } from 'svelte/store';

export type TabProtocol = 'http' | 'websocket' | 'sse' | 'grpc-stream' | 'kafka' | 'mqtt' | 'grpc';

export interface TabState {
    id: string;
    name: string;
    protocol: TabProtocol;
    isDirty: boolean;
    lastActive: Date;

    // HTTP state
    httpRequest?: {
        method: string;
        url: string;
        params: any[];
        headers: any[];
        body: string;
        bodyType: string;
        auth: any;
    };
    httpResponse?: any;

    // Streaming state
    streamingUrl?: string;
    streamingConfig?: any;
    isStreamConnected?: boolean;
    connectionId?: string;

    // Messages for streaming protocols
    messages?: any[];
}

interface TabsStore {
    tabs: TabState[];
    activeTabId: string | null;
}

function createTabsStore() {
    const { subscribe, set, update } = writable<TabsStore>({
        tabs: [],
        activeTabId: null
    });

    function generateTabName(protocol: TabProtocol): string {
        const protocolNames = {
            'http': 'HTTP Request',
            'websocket': 'WebSocket',
            'sse': 'SSE Stream',
            'grpc-stream': 'gRPC Stream',
            'kafka': 'Kafka',
            'mqtt': 'MQTT',
            'grpc': 'gRPC'
        };

        const store = get({ subscribe });
        const existingCount = store.tabs.filter(t => t.protocol === protocol).length;
        const baseName = protocolNames[protocol] || 'Request';

        return existingCount > 0 ? `${baseName} ${existingCount + 1}` : baseName;
    }

    return {
        subscribe,

        createTab: (protocol: TabProtocol = 'http') => {
            const newTab: TabState = {
                id: crypto.randomUUID(),
                name: generateTabName(protocol),
                protocol,
                isDirty: false,
                lastActive: new Date(),
                httpRequest: protocol === 'http' ? {
                    method: 'GET',
                    url: '',
                    params: [],
                    headers: [],
                    body: '',
                    bodyType: 'none',
                    auth: null
                } : undefined,
                streamingUrl: protocol !== 'http' ? '' : undefined,
                streamingConfig: protocol !== 'http' ? {} : undefined,
                isStreamConnected: false,
                messages: []
            };

            update(store => ({
                tabs: [...store.tabs, newTab],
                activeTabId: newTab.id
            }));

            return newTab.id;
        },

        setActiveTab: (tabId: string) => {
            update(store => {
                const tab = store.tabs.find(t => t.id === tabId);
                if (tab) {
                    tab.lastActive = new Date();
                }
                return {
                    ...store,
                    activeTabId: tabId
                };
            });
        },

        closeTab: (tabId: string) => {
            update(store => {
                const index = store.tabs.findIndex(t => t.id === tabId);
                if (index === -1) return store;

                const newTabs = store.tabs.filter(t => t.id !== tabId);
                let newActiveId = store.activeTabId;

                // If closing active tab, switch to adjacent tab
                if (tabId === store.activeTabId) {
                    if (newTabs.length > 0) {
                        newActiveId = newTabs[Math.max(0, index - 1)].id;
                    } else {
                        newActiveId = null;
                    }
                }

                return {
                    tabs: newTabs,
                    activeTabId: newActiveId
                };
            });
        },

        updateTab: (tabId: string, updates: Partial<TabState>) => {
            update(store => ({
                ...store,
                tabs: store.tabs.map(tab =>
                    tab.id === tabId
                        ? { ...tab, ...updates, isDirty: true, lastActive: new Date() }
                        : tab
                )
            }));
        },

        renameTab: (tabId: string, name: string) => {
            update(store => ({
                ...store,
                tabs: store.tabs.map(tab =>
                    tab.id === tabId ? { ...tab, name } : tab
                )
            }));
        },

        setConnectionState: (tabId: string, connected: boolean, connectionId?: string) => {
            update(store => ({
                ...store,
                tabs: store.tabs.map(tab =>
                    tab.id === tabId
                        ? { ...tab, isStreamConnected: connected, connectionId }
                        : tab
                )
            }));
        },

        addMessage: (tabId: string, message: any) => {
            update(store => ({
                ...store,
                tabs: store.tabs.map(tab =>
                    tab.id === tabId
                        ? { ...tab, messages: [...(tab.messages || []), message] }
                        : tab
                )
            }));
        },

        clearMessages: (tabId: string) => {
            update(store => ({
                ...store,
                tabs: store.tabs.map(tab =>
                    tab.id === tabId ? { ...tab, messages: [] } : tab
                )
            }));
        },

        duplicateTab: (tabId: string) => {
            update(store => {
                const tab = store.tabs.find(t => t.id === tabId);
                if (!tab) return store;

                const newTab: TabState = {
                    ...JSON.parse(JSON.stringify(tab)),
                    id: crypto.randomUUID(),
                    name: `${tab.name} (Copy)`,
                    isDirty: false,
                    lastActive: new Date(),
                    isStreamConnected: false,
                    connectionId: undefined,
                    messages: []
                };

                return {
                    tabs: [...store.tabs, newTab],
                    activeTabId: newTab.id
                };
            });
        },

        reset: () => {
            set({ tabs: [], activeTabId: null });
        }
    };
}

export const tabsStore = createTabsStore();

// Derived store for active tab
export const activeTab = derived(
    tabsStore,
    $tabs => $tabs.tabs.find(t => t.id === $tabs.activeTabId) || null
);