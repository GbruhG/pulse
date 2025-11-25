<script lang="ts">
    import { createEventDispatcher, onMount } from 'svelte';
    import { Link, Link2Off, Settings, AlertCircle } from 'lucide-svelte';
    import { SSEConnect, SSEDisconnect } from '../../../wailsjs/go/main/App';
    import { tabsStore, activeTab } from '../stores/tabs';

    const dispatch = createEventDispatcher();

    let connectionUrl = '';
    let isConnected = false;
    let isConnecting = false;
    let isDisconnecting = false;
    let connectionError = '';
    let connectionId = '';
    let hasLoadedInitialValues = false;
    let currentTabId = '';

    // Emit connection state changes to parent
    $: dispatch('connectionChange', isConnected);

    // SSE specific settings
    let showSettings = false;
    let withCredentials = false;
    let retryTimeout = 3000;
    let lastEventId = '';
    let autoReconnect = true;
    let customHeaders: Array<{key: string, value: string, enabled: boolean}> = [];
    let eventTypeFilter: string[] = [];

    // Watch for tab changes and reload state
    $: if ($activeTab && $activeTab.id !== currentTabId) {
        currentTabId = $activeTab.id;
        loadTabState();
    }

    function loadTabState() {
        if (!$activeTab || $activeTab.protocol !== 'sse') return;

        connectionUrl = $activeTab.streamingUrl || '';
        isConnected = $activeTab.isStreamConnected || false;
        connectionId = $activeTab.connectionId || '';

        if ($activeTab.streamingConfig) {
            const config = $activeTab.streamingConfig;
            withCredentials = config.withCredentials ?? false;
            retryTimeout = config.retryTimeout || 3000;
            lastEventId = config.lastEventId || '';
            autoReconnect = config.autoReconnect ?? true;
            customHeaders = config.headers || [];
            eventTypeFilter = config.eventTypeFilter || [];
        } else {
            withCredentials = false;
            retryTimeout = 3000;
            lastEventId = '';
            autoReconnect = true;
            customHeaders = [];
            eventTypeFilter = [];
        }

        hasLoadedInitialValues = true;
    }

    onMount(() => {
        loadTabState();
    });

    let urlUpdateTimeout: number;
    function handleUrlChange() {
        if (!hasLoadedInitialValues || !$activeTab) return;

        clearTimeout(urlUpdateTimeout);
        urlUpdateTimeout = setTimeout(() => {
            tabsStore.updateTab($activeTab.id, {
                streamingUrl: connectionUrl,
                streamingConfig: {
                    withCredentials,
                    retryTimeout,
                    lastEventId,
                    autoReconnect,
                    headers: customHeaders,
                    eventTypeFilter
                }
            });
        }, 300);
    }

    function validateSSEUrl(url: string): boolean {
        try {
            const urlObj = new URL(url);
            return urlObj.protocol === 'http:' || urlObj.protocol === 'https:';
        } catch {
            return false;
        }
    }

    async function handleConnect() {
        if (isConnected) {
            isDisconnecting = true;
            const connToDisconnect = connectionId;
            isConnected = false;

            if ($activeTab) {
                tabsStore.setConnectionState($activeTab.id, false);
            }

            connectionId = '';
            connectionError = '';

            if (connToDisconnect) {
                SSEDisconnect(connToDisconnect).catch(error => {
                    console.error('[SSE] Disconnect error:', error);
                }).finally(() => {
                    isDisconnecting = false;
                });
            } else {
                isDisconnecting = false;
            }
            return;
        }

        if (!validateSSEUrl(connectionUrl)) {
            connectionError = 'Invalid URL. SSE requires http:// or https://';
            return;
        }

        connectionError = '';
        isConnecting = true;

        try {
            const headersObj: Record<string, string> = {};
            customHeaders.filter(h => h.enabled && h.key).forEach(h => {
                headersObj[h.key] = h.value;
            });

            connectionId = await SSEConnect({
                url: connectionUrl,
                withCredentials: withCredentials,
                retryTimeout: retryTimeout,
                lastEventId: lastEventId,
                autoReconnect: autoReconnect,
                customHeaders: headersObj,
                eventTypeFilter: eventTypeFilter.filter(f => f.trim() !== '')
            });

            isConnected = true;

            if ($activeTab) {
                tabsStore.setConnectionState($activeTab.id, true, connectionId);
            }

            connectionError = '';
            console.log('[SSE] Connected:', connectionId);
        } catch (error) {
            connectionError = `Connection failed: ${error}`;
            isConnected = false;
            console.error('[SSE] Connection error:', error);
        } finally {
            isConnecting = false;
        }
    }

    function addHeader() {
        customHeaders = [...customHeaders, { key: '', value: '', enabled: true }];
        handleUrlChange();
    }

    function removeHeader(index: number) {
        customHeaders = customHeaders.filter((_, i) => i !== index);
        handleUrlChange();
    }

    function addEventType() {
        eventTypeFilter = [...eventTypeFilter, ''];
        handleUrlChange();
    }

    function removeEventType(index: number) {
        eventTypeFilter = eventTypeFilter.filter((_, i) => i !== index);
        handleUrlChange();
    }
</script>

<div class="sse-handler">
    <!-- Connection Bar -->
    <div class="connection-section">
        <div class="connection-bar">
            <div class="url-input-group">
                <input
                        type="text"
                        bind:value={connectionUrl}
                        on:input={handleUrlChange}
                        class="url-input"
                        placeholder="https://api.example.com/events"
                        disabled={isConnected || isDisconnecting}
                />

                {#if lastEventId}
                    <div class="event-id-badge">
                        Last ID: {lastEventId}
                    </div>
                {/if}
            </div>

            <button
                    class="settings-btn"
                    class:active={showSettings}
                    on:click={() => showSettings = !showSettings}
                    title="Connection settings"
                    disabled={isConnected || isDisconnecting}
            >
                <Settings size={16} />
            </button>

            <button
                    class="connect-btn"
                    class:connected={isConnected}
                    class:connecting={isConnecting}
                    class:disconnecting={isDisconnecting}
                    on:click={handleConnect}
                    disabled={isConnecting || isDisconnecting}
            >
                {#if isConnecting}
                    <span class="spinner"></span>
                    <span>Connecting...</span>
                {:else if isDisconnecting}
                    <span class="spinner"></span>
                    <span>Disconnecting...</span>
                {:else if isConnected}
                    <Link2Off size={16} />
                    <span>Disconnect</span>
                {:else}
                    <Link size={16} />
                    <span>Connect</span>
                {/if}
            </button>
        </div>

        {#if connectionError}
            <div class="error-message">
                <AlertCircle size={14} />
                {connectionError}
            </div>
        {/if}

        <div class="info-message">
            <span class="info-icon">ℹ️</span>
            Server-Sent Events is receive-only. You cannot send messages.
        </div>
    </div>

    <!-- Settings Panel -->
    {#if showSettings}
        <div class="settings-panel">
            <div class="settings-grid">
                <div class="setting-item">
                    <label class="setting-label">
                        <input type="checkbox" bind:checked={withCredentials} on:change={handleUrlChange} class="setting-checkbox" />
                        Include Credentials (CORS cookies)
                    </label>
                </div>

                <div class="setting-item">
                    <label class="setting-label">
                        <input type="checkbox" bind:checked={autoReconnect} on:change={handleUrlChange} class="setting-checkbox" />
                        Auto Reconnect
                    </label>
                </div>

                <div class="setting-item">
                    <label class="setting-label">Retry Timeout</label>
                    <div class="input-with-unit">
                        <input type="number" bind:value={retryTimeout} on:input={handleUrlChange} class="setting-input" placeholder="3000" />
                        <span class="unit">ms</span>
                    </div>
                </div>

                <div class="setting-item">
                    <label class="setting-label">Resume from Event ID</label>
                    <input type="text" bind:value={lastEventId} on:input={handleUrlChange} class="setting-input" placeholder="Optional" />
                </div>
            </div>

            <div class="filter-section">
                <div class="filter-header">
                    <span class="setting-label">Event Type Filter</span>
                    <button class="add-btn" on:click={addEventType}>+ Add</button>
                </div>
                <div class="filter-hint">Leave empty to receive all event types. Add specific types to filter.</div>
                {#if eventTypeFilter.length > 0}
                    <div class="filter-list">
                        {#each eventTypeFilter as eventType, i}
                            <div class="filter-row">
                                <input
                                        type="text"
                                        bind:value={eventTypeFilter[i]}
                                        on:input={handleUrlChange}
                                        placeholder="Event type (e.g., 'message', 'update')"
                                        class="filter-input"
                                />
                                <button class="remove-btn" on:click={() => removeEventType(i)}>×</button>
                            </div>
                        {/each}
                    </div>
                {/if}
            </div>

            <div class="headers-section">
                <div class="headers-header">
                    <span class="setting-label">Custom Headers</span>
                    <button class="add-btn" on:click={addHeader}>+ Add</button>
                </div>
                {#if customHeaders.length > 0}
                    <div class="headers-list">
                        {#each customHeaders as header, i}
                            <div class="header-row">
                                <input type="checkbox" bind:checked={header.enabled} on:change={handleUrlChange} class="header-checkbox" />
                                <input type="text" bind:value={header.key} on:input={handleUrlChange} placeholder="Header name" class="header-input" />
                                <input type="text" bind:value={header.value} on:input={handleUrlChange} placeholder="Value" class="header-input" />
                                <button class="remove-btn" on:click={() => removeHeader(i)}>×</button>
                            </div>
                        {/each}
                    </div>
                {/if}
            </div>
        </div>
    {/if}
</div>

<style>
    .sse-handler {
        display: flex;
        flex-direction: column;
        gap: 12px;
        padding: 1rem;
        background: #0a0a0a;
    }

    .connection-section {
        display: flex;
        flex-direction: column;
        gap: 8px;
    }

    .connection-bar {
        display: flex;
        align-items: center;
        gap: 8px;
    }

    .url-input-group {
        flex: 1;
        display: flex;
        align-items: center;
        gap: 0.5rem;
    }

    .url-input {
        flex: 1;
        padding: 0.625rem 0.75rem;
        background: #0f0f0f;
        border: 1px solid rgba(255, 255, 255, 0.1);
        border-radius: 4px;
        color: #e4e4e7;
        font-size: 0.875rem;
        font-family: 'SF Mono', Monaco, monospace;
        transition: all 0.2s;
    }

    .url-input:focus {
        outline: none;
        border-color: rgba(239, 68, 68, 0.5);
    }

    .url-input:disabled {
        opacity: 0.6;
        cursor: not-allowed;
    }

    .event-id-badge {
        padding: 0.25rem 0.5rem;
        background: rgba(59, 130, 246, 0.1);
        border: 1px solid rgba(59, 130, 246, 0.2);
        border-radius: 4px;
        color: #3b82f6;
        font-size: 0.75rem;
        font-weight: 500;
        white-space: nowrap;
    }

    .settings-btn {
        display: flex;
        align-items: center;
        justify-content: center;
        padding: 0.625rem;
        background: transparent;
        border: 1px solid rgba(255, 255, 255, 0.1);
        border-radius: 4px;
        color: #71717a;
        cursor: pointer;
        transition: all 0.2s;
    }

    .settings-btn:hover:not(:disabled) {
        background: rgba(255, 255, 255, 0.05);
        border-color: rgba(255, 255, 255, 0.2);
        color: #e4e4e7;
    }

    .settings-btn.active {
        background: rgba(255, 255, 255, 0.05);
        border-color: rgba(255, 255, 255, 0.2);
        color: #e4e4e7;
    }

    .settings-btn:disabled {
        opacity: 0.5;
        cursor: not-allowed;
    }

    .connect-btn {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        padding: 0.625rem 1.25rem;
        background: #10b981;
        color: white;
        border: none;
        border-radius: 4px;
        font-weight: 600;
        font-size: 0.875rem;
        cursor: pointer;
        transition: all 0.2s;
        white-space: nowrap;
    }

    .connect-btn:hover:not(:disabled) {
        background: #059669;
    }

    .connect-btn.connected {
        background: #dc2626;
    }

    .connect-btn.connected:hover:not(:disabled) {
        background: #ef4444;
    }

    .connect-btn.connecting,
    .connect-btn.disconnecting {
        background: #f59e0b;
        color: #78350f;
    }

    .connect-btn:disabled {
        opacity: 0.5;
        cursor: not-allowed;
    }

    .spinner {
        width: 14px;
        height: 14px;
        border: 2px solid transparent;
        border-top-color: currentColor;
        border-radius: 50%;
        animation: spin 0.6s linear infinite;
    }

    @keyframes spin {
        to { transform: rotate(360deg); }
    }

    .error-message {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        padding: 0.5rem 0.75rem;
        background: rgba(239, 68, 68, 0.1);
        border: 1px solid rgba(239, 68, 68, 0.2);
        border-radius: 4px;
        color: #ef4444;
        font-size: 0.875rem;
        font-weight: 500;
    }

    .info-message {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        padding: 0.5rem 0.75rem;
        background: rgba(59, 130, 246, 0.1);
        border: 1px solid rgba(59, 130, 246, 0.1);
        border-radius: 4px;
        color: #93c5fd;
        font-size: 0.875rem;
    }

    .info-icon {
        font-size: 1rem;
    }

    .settings-panel {
        display: flex;
        flex-direction: column;
        gap: 1.25rem;
        padding: 1rem;
        background: #0f0f0f;
        border: 1px solid rgba(255, 255, 255, 0.08);
        border-radius: 6px;
    }

    .settings-grid {
        display: grid;
        gap: 1rem;
    }

    .setting-item {
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
    }

    .setting-label {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        font-size: 0.875rem;
        font-weight: 500;
        color: #e4e4e7;
    }

    .setting-checkbox {
        width: 16px;
        height: 16px;
        cursor: pointer;
    }

    .setting-input {
        width: 100%;
        padding: 0.5rem 0.75rem;
        background: #0f0f0f;
        border: 1px solid rgba(255, 255, 255, 0.1);
        border-radius: 4px;
        color: #e4e4e7;
        font-size: 0.875rem;
    }

    .setting-input:focus {
        outline: none;
        border-color: rgba(239, 68, 68, 0.5);
    }

    .input-with-unit {
        display: flex;
        align-items: center;
        gap: 0.5rem;
    }

    .unit {
        font-size: 0.75rem;
        color: #71717a;
        font-weight: 500;
    }

    .filter-section,
    .headers-section {
        display: flex;
        flex-direction: column;
        gap: 0.75rem;
    }

    .filter-header,
    .headers-header {
        display: flex;
        align-items: center;
        justify-content: space-between;
    }

    .filter-hint {
        font-size: 0.75rem;
        color: #71717a;
        font-style: italic;
    }

    .add-btn {
        padding: 0.375rem 0.75rem;
        background: transparent;
        border: 1px solid rgba(255, 255, 255, 0.1);
        border-radius: 4px;
        color: #9ca3af;
        font-size: 0.75rem;
        font-weight: 500;
        cursor: pointer;
        transition: all 0.2s;
    }

    .add-btn:hover {
        background: rgba(255, 255, 255, 0.05);
        border-color: rgba(255, 255, 255, 0.2);
        color: #e4e4e7;
    }

    .filter-list,
    .headers-list {
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
    }

    .filter-row,
    .header-row {
        display: flex;
        align-items: center;
        gap: 0.5rem;
    }

    .filter-input,
    .header-input {
        flex: 1;
        padding: 0.5rem 0.75rem;
        background: #0f0f0f;
        border: 1px solid rgba(255, 255, 255, 0.1);
        border-radius: 4px;
        color: #e4e4e7;
        font-size: 0.875rem;
    }

    .filter-input:focus,
    .header-input:focus {
        outline: none;
        border-color: rgba(239, 68, 68, 0.5);
    }

    .header-checkbox {
        width: 16px;
        height: 16px;
        cursor: pointer;
    }

    .remove-btn {
        display: flex;
        align-items: center;
        justify-content: center;
        width: 1.75rem;
        height: 1.75rem;
        background: transparent;
        border: 1px solid rgba(255, 255, 255, 0.1);
        border-radius: 4px;
        color: #71717a;
        font-size: 1.125rem;
        cursor: pointer;
        transition: all 0.2s;
    }

    .remove-btn:hover {
        background: rgba(239, 68, 68, 0.1);
        border-color: rgba(239, 68, 68, 0.3);
        color: #ef4444;
    }
</style>