<script lang="ts">
    import { createEventDispatcher } from 'svelte';
    import { PanelRightOpen, PanelBottomOpen } from 'lucide-svelte';
    import ProtocolSelector from './ProtocolSelector.svelte';
    import RequestBuilder from './RequestBuilder.svelte';
    import ParamsTab from './ParamsTab.svelte';
    import HeadersTab from './HeadersTab.svelte';
    import AuthTab from './AuthTab.svelte';
    import BodyTab from './BodyTab.svelte';
    import ResponseViewer from './ResponseViewer.svelte';
    import StreamingBuilder from './StreamingBuilder.svelte';
    import StreamMessageViewer from './StreamMessageViewer.svelte';
    import { tabsStore } from '../stores/tabs';
    import { requestStore } from '../stores/request';
    import { streamingStore } from '../stores/streaming';
    import { SendRequest } from '../../../wailsjs/go/main/App';
    import { environmentStore } from '../stores/environment';
    import { historyStore } from '../stores/history';
    import { workspaceStore } from '../stores/workspace';
    import { substituteVariables } from '../utils/variables';
    import type { TabState } from '../stores/tabs';
    import type { HistoryItem } from '../types';

    export let tab: TabState;
    export let layoutMode: 'horizontal' | 'vertical';
    export let responseSize: number;

    const dispatch = createEventDispatcher();

    let activeProtocol: 'http' | 'streaming' | 'grpc' = tab.protocol === 'http' ? 'http' : 'streaming';
    let activeTab = 'params';
    let isLoading = false;
    let response: any = null;
    let isStreamConnected = false;

    // Update tab protocol when changed
    $: if (activeProtocol !== (tab.protocol === 'http' ? 'http' : 'streaming')) {
        const newProtocol = activeProtocol === 'http' ? 'http'
            : activeProtocol === 'grpc' ? 'grpc'
                : 'websocket';

        tabsStore.updateTab(tab.id, { protocol: newProtocol as any });
    }

    // Sync tab state with stores
    $: if (tab.protocol === 'http' && tab.httpRequest) {
        requestStore.setRequest(tab.httpRequest);
    }

    $: if (tab.protocol !== 'http' && tab.streamingUrl) {
        streamingStore.setRequest({
            protocol: tab.protocol as any,
            url: tab.streamingUrl,
            config: tab.streamingConfig || {}
        });
    }

    // Update connection state
    $: isStreamConnected = tab.isStreamConnected || false;

    async function handleSend() {
        if (activeProtocol !== 'http') return;

        isLoading = true;

        try {
            const current = tab.httpRequest!;
            const startTime = Date.now();

            const activeEnv = $environmentStore.environments.find(
                e => e.id === $environmentStore.activeEnvironmentId
            );
            const variables = activeEnv?.variables || {};

            let url = substituteVariables(current.url, variables);

            const enabledParams = current.params?.filter(p => p.enabled && p.key) || [];
            if (enabledParams.length > 0) {
                const params = new URLSearchParams();
                enabledParams.forEach(p => {
                    const key = substituteVariables(p.key, variables);
                    const value = substituteVariables(p.value, variables);
                    params.append(key, value);
                });
                url += (url.includes('?') ? '&' : '?') + params.toString();
            }

            const headers: Record<string, string> = {};
            (current.headers || []).filter(h => h.enabled && h.key).forEach(h => {
                const key = substituteVariables(h.key, variables);
                const value = substituteVariables(h.value, variables);
                headers[key] = value;
            });

            const body = substituteVariables(current.body || '', variables);

            let auth = current.auth;
            if (auth) {
                auth = {
                    ...auth,
                    username: substituteVariables(auth.username || '', variables),
                    password: substituteVariables(auth.password || '', variables),
                    token: substituteVariables(auth.token || '', variables),
                    value: substituteVariables(auth.value || '', variables),
                };
            }

            const result = await SendRequest({
                method: current.method,
                url,
                params: current.params || [],
                headers: current.headers || [],
                body,
                bodyType: current.bodyType || 'none',
                auth
            });

            const endTime = Date.now();
            const duration = endTime - startTime;

            response = {
                statusCode: result.statusCode,
                statusText: result.statusText,
                time: `${duration}ms`,
                size: formatBytes(new Blob([result.body]).size),
                headers: result.headers,
                body: result.body
            };

            tabsStore.updateTab(tab.id, { httpResponse: response });

            const historyItem: HistoryItem = {
                id: crypto.randomUUID(),
                request: current,
                response,
                timestamp: new Date(),
                workspaceId: $workspaceStore.activeWorkspaceId || ''
            };
            historyStore.addItem(historyItem);

        } catch (error) {
            response = {
                statusCode: 0,
                statusText: 'Error',
                time: '0ms',
                size: '0kb',
                headers: {},
                body: `Error: ${error}`
            };
            tabsStore.updateTab(tab.id, { httpResponse: response });
        } finally {
            isLoading = false;
        }
    }

    function formatBytes(bytes: number): string {
        if (bytes === 0) return '0 Bytes';
        const k = 1024;
        const sizes = ['Bytes', 'KB', 'MB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i];
    }

    function handleConnectionChange(connected: boolean) {
        tabsStore.setConnectionState(tab.id, connected);
    }

    function handleSaveToCollection() {
        dispatch('saveToCollection');
    }
</script>

<div class="work-area" class:horizontal={layoutMode === 'horizontal'} class:vertical={layoutMode === 'vertical'}>
    <div class="request-section" style="{layoutMode === 'horizontal' ? `width: ${100 - responseSize}%` : `height: ${100 - responseSize}%`}">
        <div class="layout-toggle">
            <button
                    class="toggle-btn"
                    class:active={layoutMode === 'horizontal'}
                    on:click={() => dispatch('layoutChange', 'horizontal')}
                    title="Response on right"
            >
                <PanelRightOpen size={16} />
            </button>
            <button
                    class="toggle-btn"
                    class:active={layoutMode === 'vertical'}
                    on:click={() => dispatch('layoutChange', 'vertical')}
                    title="Response on bottom"
            >
                <PanelBottomOpen size={16} />
            </button>
        </div>

        <div class="protocol-section">
            <ProtocolSelector bind:activeProtocol />

            {#if activeProtocol === 'http'}
                <RequestBuilder
                        on:send={handleSend}
                        bind:activeTab
                        onSaveToCollection={handleSaveToCollection}
                        loading={isLoading}
                />
            {:else if activeProtocol === 'streaming'}
                <StreamingBuilder onSaveToCollection={handleSaveToCollection} />
            {:else if activeProtocol === 'grpc'}
                <div class="coming-soon">gRPC - Coming Soon</div>
            {/if}
        </div>

        {#if activeProtocol === 'http'}
            <div class="tab-content">
                {#if activeTab === 'params'}
                    <ParamsTab />
                {:else if activeTab === 'headers'}
                    <HeadersTab />
                {:else if activeTab === 'auth'}
                    <AuthTab />
                {:else if activeTab === 'body'}
                    <BodyTab />
                {/if}
            </div>
        {/if}
    </div>

    <div class="main-resize-handle" class:horizontal={layoutMode === 'horizontal'} class:vertical={layoutMode === 'vertical'} on:mousedown={() => dispatch('resizeStart')}></div>

    <div class="response-section" style="{layoutMode === 'horizontal' ? `width: ${responseSize}%` : `height: ${responseSize}%`}">
        {#if activeProtocol === 'http'}
            <ResponseViewer response={response || tab.httpResponse} />
        {:else if activeProtocol === 'streaming'}
            <StreamMessageViewer
                    messages={tab.messages || []}
                    isConnected={isStreamConnected}
            />
        {/if}
    </div>
</div>

<style>
    .work-area {
        display: flex;
        flex: 1;
        overflow: hidden;
        position: relative;
    }

    .work-area.horizontal {
        flex-direction: row;
    }

    .work-area.vertical {
        flex-direction: column;
    }

    .request-section {
        display: flex;
        flex-direction: column;
        overflow: hidden;
        background: #0a0a0a;
        position: relative;
        min-width: 20%;
        min-height: 20%;
    }

    .layout-toggle {
        position: absolute;
        top: 1.5rem;
        right: 1.5rem;
        display: flex;
        gap: 0.25rem;
        background: #0a0a0a;
        border: 1px solid #2a2a2a;
        border-radius: 0.5rem;
        padding: 0.25rem;
        z-index: 10;
    }

    .toggle-btn {
        display: flex;
        align-items: center;
        justify-content: center;
        width: 32px;
        height: 32px;
        background: transparent;
        border: none;
        color: #6b7280;
        cursor: pointer;
        border-radius: 0.375rem;
        transition: all 0.2s;
    }

    .toggle-btn:hover {
        background: #1a1a1a;
        color: #d1d5db;
    }

    .toggle-btn.active {
        background: #1a1a1a;
        color: #ef4444;
    }

    .protocol-section {
        flex-shrink: 0;
    }

    .tab-content {
        flex: 1;
        overflow: auto;
        background: #0a0a0a;
    }

    .coming-soon {
        display: flex;
        align-items: center;
        justify-content: center;
        padding: 4rem;
        color: #71717a;
        font-size: 14px;
    }

    .main-resize-handle {
        background: #2a2a2a;
        transition: background 0.2s;
        position: relative;
        z-index: 10;
        flex-shrink: 0;
    }

    .main-resize-handle.horizontal {
        width: 4px;
        cursor: col-resize;
    }

    .main-resize-handle.vertical {
        height: 4px;
        cursor: row-resize;
    }

    .main-resize-handle:hover,
    .main-resize-handle:active {
        background: #ef4444;
    }

    .response-section {
        display: flex;
        flex-direction: column;
        overflow: hidden;
        background: #0f0f0f;
        min-width: 300px;
        min-height: 300px;
    }
</style>