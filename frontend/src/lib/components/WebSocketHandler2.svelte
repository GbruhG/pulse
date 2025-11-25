<script lang="ts">
    import { createEventDispatcher, onMount } from 'svelte';
    import { Send, Link, Link2Off, Settings, AlertCircle } from 'lucide-svelte';
    import { WebSocketConnect, WebSocketSendMessage, WebSocketDisconnect } from '../../../wailsjs/go/main/App';
    import { tabsStore, activeTab } from '../stores/tabs';

    type MessageFormat = 'text' | 'json' | 'binary';

    const dispatch = createEventDispatcher();

    let connectionUrl = '';
    let isConnected = false;
    let isConnecting = false;
    let isDisconnecting = false;
    let connectionError = '';
    let connectionId = '';
    let messageBody = '';
    let messageFormat: MessageFormat = 'text';
    let hasLoadedInitialValues = false;
    let currentTabId = '';

    // Emit connection state changes to parent
    $: dispatch('connectionChange', isConnected);

    // WebSocket specific settings
    let showSettings = false;
    let autoReconnect = true;
    let reconnectInterval = 3000;
    let enablePingPong = false;
    let pingInterval = 30000;
    let customHeaders: Array<{key: string, value: string, enabled: boolean}> = [];
    let selectedSubprotocol = '';
    let subprotocols = ['', 'soap', 'wamp', 'mqtt'];

    // Watch for tab changes and reload state
    $: if ($activeTab && $activeTab.id !== currentTabId) {
        currentTabId = $activeTab.id;
        loadTabState();
    }

    function loadTabState() {
        if (!$activeTab || $activeTab.protocol !== 'websocket') return;

        // Reset all state
        connectionUrl = $activeTab.streamingUrl || '';
        isConnected = $activeTab.isStreamConnected || false;
        connectionId = $activeTab.connectionId || '';

        if ($activeTab.streamingConfig) {
            const config = $activeTab.streamingConfig;
            selectedSubprotocol = config.subprotocol || '';
            autoReconnect = config.autoReconnect ?? true;
            reconnectInterval = config.reconnectInterval || 3000;
            enablePingPong = config.enablePingPong ?? false;
            pingInterval = config.pingInterval || 30000;
            customHeaders = config.headers || [];
        } else {
            // Reset to defaults
            selectedSubprotocol = '';
            autoReconnect = true;
            reconnectInterval = 3000;
            enablePingPong = false;
            pingInterval = 30000;
            customHeaders = [];
        }

        hasLoadedInitialValues = true;
    }

    // Load initial values on mount
    onMount(() => {
        loadTabState();
    });

    // Save to tab when URL changes (debounced, only after initial load)
    let urlUpdateTimeout: number;
    function handleUrlChange() {
        if (!hasLoadedInitialValues || !$activeTab) return;

        clearTimeout(urlUpdateTimeout);
        urlUpdateTimeout = setTimeout(() => {
            tabsStore.updateTab($activeTab.id, {
                streamingUrl: connectionUrl,
                streamingConfig: {
                    subprotocol: selectedSubprotocol,
                    autoReconnect,
                    reconnectInterval,
                    enablePingPong,
                    pingInterval,
                    headers: customHeaders
                }
            });
        }, 300);
    }

    function validateWebSocketUrl(url: string): boolean {
        try {
            const urlObj = new URL(url);
            return urlObj.protocol === 'ws:' || urlObj.protocol === 'wss:';
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
                WebSocketDisconnect(connToDisconnect).catch(error => {
                    console.error('[WS] Disconnect error:', error);
                }).finally(() => {
                    isDisconnecting = false;
                });
            } else {
                isDisconnecting = false;
            }
            return;
        }

        if (!validateWebSocketUrl(connectionUrl)) {
            connectionError = 'Invalid WebSocket URL. Must start with ws:// or wss://';
            return;
        }

        connectionError = '';
        isConnecting = true;

        try {
            const headersObj: Record<string, string> = {};
            customHeaders.filter(h => h.enabled && h.key).forEach(h => {
                headersObj[h.key] = h.value;
            });

            connectionId = await WebSocketConnect({
                url: connectionUrl,
                subprotocol: selectedSubprotocol,
                autoReconnect: autoReconnect,
                reconnectInterval: reconnectInterval,
                enablePingPong: enablePingPong,
                pingInterval: pingInterval,
                customHeaders: headersObj
            });

            isConnected = true;

            if ($activeTab) {
                tabsStore.setConnectionState($activeTab.id, true, connectionId);
            }

            connectionError = '';
        } catch (error) {
            connectionError = `Connection failed: ${error}`;
            isConnected = false;
        } finally {
            isConnecting = false;
        }
    }

    async function handleSendMessage() {
        if (!messageBody.trim() || !connectionId) {
            return;
        }

        try {
            if (messageFormat === 'json') {
                JSON.parse(messageBody);
            }

            await WebSocketSendMessage({
                connectionId: connectionId,
                message: messageBody,
                messageType: messageFormat
            });
        } catch (error) {
            if (error instanceof SyntaxError) {
                connectionError = 'Invalid JSON in message body';
            } else {
                connectionError = `Failed to send: ${error}`;
            }
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
</script>

<div class="websocket-handler">
    <!-- Connection Bar -->
    <div class="connection-section">
        <div class="connection-bar">
            <div class="url-input-group">
                <input
                        type="text"
                        bind:value={connectionUrl}
                        on:input={handleUrlChange}
                        class="url-input"
                        placeholder="wss://echo.websocket.org"
                        disabled={isConnected || isDisconnecting}
                />

                {#if selectedSubprotocol}
                    <div class="subprotocol-badge">
                        Subprotocol: {selectedSubprotocol}
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
    </div>

    <!-- Settings Panel -->
    {#if showSettings}
        <div class="settings-panel">
            <div class="settings-grid">
                <div class="setting-item">
                    <label class="setting-label">Subprotocol</label>
                    <select bind:value={selectedSubprotocol} on:change={handleUrlChange} class="setting-select">
                        <option value="">None</option>
                        {#each subprotocols.filter(s => s !== '') as subprotocol}
                            <option value={subprotocol}>{subprotocol}</option>
                        {/each}
                    </select>
                </div>

                <div class="setting-item">
                    <label class="setting-label">
                        <input type="checkbox" bind:checked={autoReconnect} on:change={handleUrlChange} class="setting-checkbox" />
                        Auto Reconnect
                    </label>
                    {#if autoReconnect}
                        <input type="number" bind:value={reconnectInterval} on:input={handleUrlChange} class="setting-input" placeholder="3000" />
                        <span class="setting-hint">ms</span>
                    {/if}
                </div>

                <div class="setting-item">
                    <label class="setting-label">
                        <input type="checkbox" bind:checked={enablePingPong} on:change={handleUrlChange} class="setting-checkbox" />
                        Enable Ping/Pong
                    </label>
                    {#if enablePingPong}
                        <input type="number" bind:value={pingInterval} on:input={handleUrlChange} class="setting-input" placeholder="30000" />
                        <span class="setting-hint">ms</span>
                    {/if}
                </div>
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

    <!-- Message Input -->
    {#if isConnected}
        <div class="message-section">
            <div class="message-header">
                <span class="section-title">Send Message</span>
                <div class="format-toggle">
                    <button
                            class="format-btn"
                            class:active={messageFormat === 'text'}
                            on:click={() => messageFormat = 'text'}
                    >
                        Text
                    </button>
                    <button
                            class="format-btn"
                            class:active={messageFormat === 'json'}
                            on:click={() => messageFormat = 'json'}
                    >
                        JSON
                    </button>
                    <button
                            class="format-btn"
                            class:active={messageFormat === 'binary'}
                            on:click={() => messageFormat = 'binary'}
                    >
                        Binary
                    </button>
                </div>
            </div>

            <div class="message-input-wrapper">
                <textarea
                        bind:value={messageBody}
                        class="message-input"
                        placeholder={messageFormat === 'json'
                        ? '{\n  "type": "message",\n  "content": "Hello"\n}'
                        : messageFormat === 'binary'
                        ? 'Base64 encoded data or hex (e.g., 0x48656C6C6F)'
                        : 'Enter your message...'}
                />
                <button
                        class="send-btn"
                        on:click={handleSendMessage}
                        disabled={!messageBody.trim()}
                >
                    <Send size={16} />
                    <span>Send</span>
                </button>
            </div>
        </div>
    {/if}
</div>

<style>
    .connect-btn.disconnecting {
        background: #f59e0b;
        color: #78350f;
    }

    .connect-btn.disconnecting:hover {
        background: #eab308;
    }
    .websocket-handler {
        display: flex;
        flex-direction: column;
        gap: 12px;
        padding: 1rem;
    }

    .connection-section {
        display: flex;
        flex-direction: column;
        gap: 8px;
    }

    .connection-bar {
        display: flex;
        gap: 8px;
        align-items: stretch;
    }

    .url-input-group {
        flex: 1;
        display: flex;
        align-items: center;
        gap: 10px;
        background: #0f0f0f;
        border: 1px solid rgba(255, 255, 255, 0.1);
        border-radius: 4px;
        padding: 0 0.75rem;
        transition: all 0.2s;
    }

    .url-input-group:focus-within {
        border-color: rgba(239, 68, 68, 0.5);
    }

    .url-input {
        flex: 1;
        background: transparent;
        border: none;
        color: #e4e4e7;
        padding: 0.625rem 0;
        font-size: 0.875rem;
        outline: none;
        font-family: 'SF Mono', Monaco, monospace;
    }

    .url-input:disabled {
        opacity: 0.6;
        cursor: not-allowed;
    }

    .url-input::placeholder {
        color: #52525b;
    }

    .subprotocol-badge {
        font-size: 11px;
        color: #a1a1aa;
        background: #18181b;
        padding: 4px 8px;
        border-radius: 4px;
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

    .settings-btn:hover,
    .settings-btn.active {
        background: rgba(255, 255, 255, 0.05);
        border-color: rgba(255, 255, 255, 0.2);
        color: #e4e4e7;
    }

    .connect-btn {
        display: flex;
        align-items: center;
        gap: 8px;
        background: #10b981;
        color: white;
        border: none;
        padding: 0.625rem 1.25rem;
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

    .connect-btn.connected:hover {
        background: #ef4444;
    }

    .connect-btn.connecting {
        opacity: 0.8;
        cursor: wait;
    }

    .spinner {
        width: 16px;
        height: 16px;
        border: 2px solid rgba(255,255,255,0.3);
        border-top-color: white;
        border-radius: 50%;
        animation: spin 0.8s linear infinite;
    }

    @keyframes spin {
        to { transform: rotate(360deg); }
    }

    .error-message {
        display: flex;
        align-items: center;
        gap: 8px;
        padding: 10px 14px;
        background: rgba(239, 68, 68, 0.1);
        border: 1px solid rgba(239, 68, 68, 0.3);
        border-radius: 6px;
        color: #ef4444;
        font-size: 12px;
        font-weight: 600;
    }

    .settings-panel {
        display: flex;
        flex-direction: column;
        gap: 16px;
        background: #111111;
        border: 1px solid #27272a;
        border-radius: 8px;
        padding: 16px;
    }

    .settings-grid {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
        gap: 16px;
    }

    .setting-item {
        display: flex;
        flex-direction: column;
        gap: 6px;
    }

    .setting-label {
        font-size: 12px;
        font-weight: 600;
        color: #a1a1aa;
        display: flex;
        align-items: center;
        gap: 8px;
    }

    .setting-checkbox {
        width: 16px;
        height: 16px;
        cursor: pointer;
    }

    .setting-select,
    .setting-input {
        padding: 8px 12px;
        background: #0a0a0a;
        border: 1px solid #27272a;
        border-radius: 6px;
        color: #e4e4e7;
        font-size: 12px;
        outline: none;
        font-family: 'SF Mono', Monaco, monospace;
    }

    .setting-select:focus,
    .setting-input:focus {
        border-color: #3f3f46;
    }

    .setting-hint {
        font-size: 11px;
        color: #71717a;
    }

    .headers-section {
        display: flex;
        flex-direction: column;
        gap: 10px;
    }

    .headers-header {
        display: flex;
        align-items: center;
        justify-content: space-between;
    }

    .add-btn {
        padding: 4px 10px;
        background: #18181b;
        border: 1px solid #27272a;
        border-radius: 4px;
        color: #a1a1aa;
        font-size: 11px;
        font-weight: 600;
        cursor: pointer;
        transition: all 0.2s;
    }

    .add-btn:hover {
        background: #27272a;
        color: #e4e4e7;
    }

    .headers-list {
        display: flex;
        flex-direction: column;
        gap: 6px;
    }

    .header-row {
        display: flex;
        align-items: center;
        gap: 8px;
    }

    .header-checkbox {
        width: 16px;
        height: 16px;
        cursor: pointer;
    }

    .header-input {
        flex: 1;
        padding: 6px 10px;
        background: #0a0a0a;
        border: 1px solid #27272a;
        border-radius: 4px;
        color: #e4e4e7;
        font-size: 12px;
        outline: none;
        font-family: 'SF Mono', Monaco, monospace;
    }

    .header-input:focus {
        border-color: #3f3f46;
    }

    .remove-btn {
        width: 24px;
        height: 24px;
        display: flex;
        align-items: center;
        justify-content: center;
        background: #18181b;
        border: 1px solid #27272a;
        border-radius: 4px;
        color: #ef4444;
        font-size: 18px;
        cursor: pointer;
        transition: all 0.2s;
    }

    .remove-btn:hover {
        background: rgba(239, 68, 68, 0.1);
        border-color: #ef4444;
    }

    .message-section {
        display: flex;
        flex-direction: column;
        gap: 10px;
        background: #111111;
        border: 1px solid #27272a;
        border-radius: 8px;
        padding: 14px;
    }

    .message-header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 12px;
    }

    .section-title {
        font-size: 12px;
        font-weight: 700;
        color: #a1a1aa;
        text-transform: uppercase;
        letter-spacing: 0.5px;
    }

    .format-toggle {
        display: flex;
        gap: 2px;
        background: #0a0a0a;
        padding: 2px;
        border-radius: 4px;
        border: 1px solid rgba(255, 255, 255, 0.08);
    }

    .format-btn {
        padding: 0.25rem 0.75rem;
        background: transparent;
        border: none;
        color: #71717a;
        font-size: 0.75rem;
        font-weight: 500;
        cursor: pointer;
        border-radius: 2px;
        transition: all 0.2s;
        white-space: nowrap;
    }

    .format-btn:hover {
        color: #e4e4e7;
        background: rgba(255, 255, 255, 0.05);
    }

    .format-btn.active {
        background: rgba(255, 255, 255, 0.1);
        color: #e4e4e7;
    }

    .message-input-wrapper {
        display: flex;
        gap: 10px;
        align-items: flex-end;
    }

    .message-input {
        flex: 1;
        background: #0a0a0a;
        border: 1px solid #27272a;
        border-radius: 6px;
        color: #e4e4e7;
        padding: 10px 12px;
        font-size: 13px;
        font-family: 'SF Mono', Monaco, monospace;
        outline: none;
        resize: vertical;
        min-height: 80px;
        max-height: 200px;
        transition: all 0.2s;
    }

    .message-input:focus {
        border-color: #3f3f46;
        background: #0f0f0f;
    }

    .message-input::placeholder {
        color: #52525b;
    }

    .send-btn {
        display: flex;
        align-items: center;
        gap: 6px;
        background: #3b82f6;
        color: white;
        border: none;
        padding: 0.5rem 1rem;
        border-radius: 4px;
        font-weight: 500;
        font-size: 0.875rem;
        cursor: pointer;
        transition: all 0.2s;
        white-space: nowrap;
    }

    .send-btn:hover:not(:disabled) {
        background: #2563eb;
    }

    .send-btn:disabled {
        opacity: 0.5;
        cursor: not-allowed;
    }
</style>