<script lang="ts">
    import { createEventDispatcher, onMount } from 'svelte';
    import { Send, Link, Link2Off, Settings, AlertCircle, Play, Pause, Plus, Trash2, RefreshCw } from 'lucide-svelte';
    import { KafkaConnect, KafkaDisconnect, KafkaListTopics, KafkaStartConsumer, KafkaStopConsumer, KafkaProduceMessage } from '../../../wailsjs/go/main/App';
    import { tabsStore, activeTab } from '../stores/tabs';

    type AuthMechanism = 'none' | 'plain' | 'scram-sha-256' | 'scram-sha-512';
    type OffsetStrategy = 'latest' | 'earliest' | 'custom';
    type CompressionType = 'none' | 'gzip' | 'snappy' | 'lz4' | 'zstd';

    const dispatch = createEventDispatcher();

    // Connection state
    let bootstrapServers = '';
    let isConnected = false;
    let isConnecting = false;
    let isDisconnecting = false;
    let connectionError = '';
    let connectionId = '';
    let currentTabId = '';
    let hasLoadedInitialValues = false;

    // Settings
    let showSettings = false;
    let clientId = 'pulse-kafka-client';
    let authMechanism: AuthMechanism = 'none';
    let saslUsername = '';
    let saslPassword = '';
    let useTLS = false;
    let tlsSkipVerify = false;
    let connectionTimeout = 10000;

    // Topics
    let topics: Array<{name: string, partitions: number}> = [];
    let selectedTopic = '';
    let isLoadingTopics = false;

    // Consumer
    let showConsumer = false;
    let isConsuming = false;
    let selectedPartitions: string[] = ['all'];
    let consumerGroup = '';
    let offsetStrategy: OffsetStrategy = 'latest';
    let customOffset = 0;
    let customTimestamp = '';
    let autoCommit = true;

    // Producer
    let showProducer = false;
    let produceTopic = '';
    let producePartition = 'auto';
    let messageKey = '';
    let messageValue = '';
    let messageHeaders: Array<{key: string, value: string, enabled: boolean}> = [];
    let compression: CompressionType = 'none';
    let acks: 0 | 1 | -1 = 1;

    $: dispatch('connectionChange', isConnected);

    $: if ($activeTab && $activeTab.id !== currentTabId) {
        currentTabId = $activeTab.id;
        loadTabState();
    }

    function loadTabState() {
        if (!$activeTab || $activeTab.protocol !== 'kafka') return;

        bootstrapServers = $activeTab.streamingUrl || '';
        isConnected = $activeTab.isStreamConnected || false;
        connectionId = $activeTab.connectionId || '';

        if ($activeTab.streamingConfig) {
            const config = $activeTab.streamingConfig;
            clientId = config.clientId || 'pulse-kafka-client';
            authMechanism = config.authMechanism || 'none';
            saslUsername = config.saslUsername || '';
            saslPassword = config.saslPassword || '';
            useTLS = config.useTLS ?? false;
            tlsSkipVerify = config.tlsSkipVerify ?? false;
            connectionTimeout = config.connectionTimeout || 10000;
            consumerGroup = config.consumerGroup || '';
            topics = config.topics || [];
        } else {
            // Reset to defaults
            clientId = 'pulse-kafka-client';
            authMechanism = 'none';
            saslUsername = '';
            saslPassword = '';
            useTLS = false;
            tlsSkipVerify = false;
            connectionTimeout = 10000;
            consumerGroup = '';
            topics = [];
        }

        hasLoadedInitialValues = true;
    }

    onMount(() => {
        loadTabState();
    });

    let urlUpdateTimeout: number;
    function handleConfigChange() {
        if (!hasLoadedInitialValues || !$activeTab) return;

        clearTimeout(urlUpdateTimeout);
        urlUpdateTimeout = setTimeout(() => {
            tabsStore.updateTab($activeTab.id, {
                streamingUrl: bootstrapServers,
                streamingConfig: {
                    clientId,
                    authMechanism,
                    saslUsername,
                    saslPassword,
                    useTLS,
                    tlsSkipVerify,
                    connectionTimeout,
                    consumerGroup,
                    topics
                }
            });
        }, 300);
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
            topics = [];
            isConsuming = false;

            if (connToDisconnect) {
                KafkaDisconnect(connToDisconnect).catch(error => {
                    console.error('[Kafka] Disconnect error:', error);
                }).finally(() => {
                    isDisconnecting = false;
                });
            } else {
                isDisconnecting = false;
            }
            return;
        }

        if (!bootstrapServers.trim()) {
            connectionError = 'Bootstrap servers are required';
            return;
        }

        connectionError = '';
        isConnecting = true;

        try {
            connectionId = await KafkaConnect({
                bootstrapServers: bootstrapServers.split(',').map(s => s.trim()),
                clientId,
                authMechanism,
                saslUsername,
                saslPassword,
                useTLS,
                tlsSkipVerify,
                connectionTimeout
            });

            isConnected = true;

            if ($activeTab) {
                tabsStore.setConnectionState($activeTab.id, true, connectionId);
            }

            connectionError = '';

            // Auto-load topics
            await loadTopics();
        } catch (error) {
            connectionError = `Connection failed: ${error}`;
            isConnected = false;
        } finally {
            isConnecting = false;
        }
    }

    async function loadTopics() {
        if (!connectionId) return;

        isLoadingTopics = true;
        connectionError = '';

        try {
            const result = await KafkaListTopics(connectionId);
            topics = result.map((t: any) => ({
                name: t.name,
                partitions: t.partitions
            }));
            handleConfigChange();
        } catch (error) {
            connectionError = `Failed to load topics: ${error}`;
        } finally {
            isLoadingTopics = false;
        }
    }

    async function handleStartConsumer() {
        if (!connectionId || !selectedTopic) {
            connectionError = 'Select a topic to consume';
            return;
        }

        connectionError = '';

        try {
            const partitionNums = selectedPartitions.includes('all')
                ? []
                : selectedPartitions.map(p => parseInt(p));

            await KafkaStartConsumer({
                connectionId,
                topic: selectedTopic,
                partitions: partitionNums,
                consumerGroup,
                offsetStrategy,
                customOffset: offsetStrategy === 'custom' ? customOffset : 0,
                autoCommit
            });

            isConsuming = true;
            showConsumer = true;
        } catch (error) {
            connectionError = `Failed to start consumer: ${error}`;
        }
    }

    async function handleStopConsumer() {
        if (!connectionId) return;

        try {
            await KafkaStopConsumer(connectionId);
            isConsuming = false;
        } catch (error) {
            connectionError = `Failed to stop consumer: ${error}`;
        }
    }

    async function handleProduceMessage() {
        if (!connectionId || !produceTopic || !messageValue.trim()) {
            connectionError = 'Topic and message value are required';
            return;
        }

        connectionError = '';

        try {
            const headersObj: Record<string, string> = {};
            messageHeaders.filter(h => h.enabled && h.key).forEach(h => {
                headersObj[h.key] = h.value;
            });

            await KafkaProduceMessage({
                connectionId,
                topic: produceTopic,
                partition: producePartition === 'auto' ? -1 : parseInt(producePartition),
                key: messageKey,
                value: messageValue,
                headers: headersObj,
                compression,
                acks
            });

            // Clear message after successful send
            messageValue = '';
            messageKey = '';
        } catch (error) {
            connectionError = `Failed to produce message: ${error}`;
        }
    }

    function addHeader() {
        messageHeaders = [...messageHeaders, { key: '', value: '', enabled: true }];
    }

    function removeHeader(index: number) {
        messageHeaders = messageHeaders.filter((_, i) => i !== index);
    }

    function getPartitionOptions(topic: string): number {
        const t = topics.find(t => t.name === topic);
        return t ? t.partitions : 1;
    }
</script>

<div class="kafka-handler-wrapper">
    <div class="kafka-handler">
        <!-- Connection Bar -->
        <div class="connection-section">
            <div class="connection-bar">
                <div class="url-input-group">
                    <input
                            type="text"
                            bind:value={bootstrapServers}
                            on:input={handleConfigChange}
                            class="url-input"
                            placeholder="localhost:9092,broker2:9092"
                            disabled={isConnected || isDisconnecting}
                    />
                    {#if authMechanism !== 'none'}
                        <div class="auth-badge">
                            Auth: {authMechanism.toUpperCase()}
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
                        <label class="setting-label">Client ID</label>
                        <input type="text" bind:value={clientId} on:input={handleConfigChange} class="setting-input" />
                    </div>

                    <div class="setting-item">
                        <label class="setting-label">Connection Timeout (ms)</label>
                        <input type="number" bind:value={connectionTimeout} on:input={handleConfigChange} class="setting-input" />
                    </div>

                    <div class="setting-item">
                        <label class="setting-label">SASL Mechanism</label>
                        <select bind:value={authMechanism} on:change={handleConfigChange} class="setting-select">
                            <option value="none">None</option>
                            <option value="plain">PLAIN</option>
                            <option value="scram-sha-256">SCRAM-SHA-256</option>
                            <option value="scram-sha-512">SCRAM-SHA-512</option>
                        </select>
                    </div>

                    {#if authMechanism !== 'none'}
                        <div class="setting-item">
                            <label class="setting-label">Username</label>
                            <input type="text" bind:value={saslUsername} on:input={handleConfigChange} class="setting-input" />
                        </div>

                        <div class="setting-item">
                            <label class="setting-label">Password</label>
                            <input type="password" bind:value={saslPassword} on:input={handleConfigChange} class="setting-input" />
                        </div>
                    {/if}

                    <div class="setting-item">
                        <label class="setting-label">
                            <input type="checkbox" bind:checked={useTLS} on:change={handleConfigChange} class="setting-checkbox" />
                            Enable TLS
                        </label>
                    </div>

                    {#if useTLS}
                        <div class="setting-item">
                            <label class="setting-label">
                                <input type="checkbox" bind:checked={tlsSkipVerify} on:change={handleConfigChange} class="setting-checkbox" />
                                Skip TLS Verification
                            </label>
                        </div>
                    {/if}
                </div>
            </div>
        {/if}

        <!-- Topics & Consumer/Producer -->
        {#if isConnected}
            <div class="topics-section">
                <div class="section-header">
                    <span class="section-title">Topics</span>
                    <button class="refresh-btn" on:click={loadTopics} disabled={isLoadingTopics}>
                        <RefreshCw size={14} class={isLoadingTopics ? 'spinning' : ''} />
                        {isLoadingTopics ? 'Loading...' : 'Refresh'}
                    </button>
                </div>

                {#if topics.length > 0}
                    <select bind:value={selectedTopic} class="topic-select">
                        <option value="">Select a topic...</option>
                        {#each topics as topic}
                            <option value={topic.name}>{topic.name} ({topic.partitions} partition{topic.partitions !== 1 ? 's' : ''})</option>
                        {/each}
                    </select>
                {:else}
                    <div class="empty-topics">No topics found</div>
                {/if}
            </div>

            <!-- Consumer Panel -->
            <div class="panel-section">
                <button class="panel-toggle" class:active={showConsumer} on:click={() => showConsumer = !showConsumer}>
                    Consumer
                </button>

                {#if showConsumer && selectedTopic}
                    <div class="panel-content">
                        <div class="consumer-controls">
                            <div class="control-row">
                                <div class="control-item">
                                    <label class="control-label">Consumer Group</label>
                                    <input type="text" bind:value={consumerGroup} placeholder="my-consumer-group" class="control-input" />
                                </div>

                                <div class="control-item">
                                    <label class="control-label">Offset Strategy</label>
                                    <select bind:value={offsetStrategy} class="control-select">
                                        <option value="latest">Latest</option>
                                        <option value="earliest">Earliest</option>
                                        <option value="custom">Custom</option>
                                    </select>
                                </div>

                                {#if offsetStrategy === 'custom'}
                                    <div class="control-item">
                                        <label class="control-label">Custom Offset</label>
                                        <input type="number" bind:value={customOffset} class="control-input" />
                                    </div>
                                {/if}
                            </div>

                            <div class="control-row">
                                <div class="control-item">
                                    <label class="control-label">
                                        <input type="checkbox" bind:checked={autoCommit} class="control-checkbox" />
                                        Auto Commit Offsets
                                    </label>
                                </div>
                            </div>

                            <div class="control-actions">
                                {#if !isConsuming}
                                    <button class="action-btn primary" on:click={handleStartConsumer}>
                                        <Play size={16} />
                                        Start Consumer
                                    </button>
                                {:else}
                                    <button class="action-btn danger" on:click={handleStopConsumer}>
                                        <Pause size={16} />
                                        Stop Consumer
                                    </button>
                                {/if}
                            </div>
                        </div>
                    </div>
                {/if}
            </div>

            <!-- Producer Panel -->
            <div class="panel-section">
                <button class="panel-toggle" class:active={showProducer} on:click={() => showProducer = !showProducer}>
                    Producer
                </button>

                {#if showProducer}
                    <div class="panel-content">
                        <div class="producer-controls">
                            <div class="control-row">
                                <div class="control-item">
                                    <label class="control-label">Topic</label>
                                    <select bind:value={produceTopic} class="control-select">
                                        <option value="">Select topic...</option>
                                        {#each topics as topic}
                                            <option value={topic.name}>{topic.name}</option>
                                        {/each}
                                    </select>
                                </div>

                                <div class="control-item">
                                    <label class="control-label">Partition</label>
                                    <select bind:value={producePartition} class="control-select">
                                        <option value="auto">Auto</option>
                                        {#if produceTopic}
                                            {#each Array(getPartitionOptions(produceTopic)) as _, i}
                                                <option value={i.toString()}>{i}</option>
                                            {/each}
                                        {/if}
                                    </select>
                                </div>

                                <div class="control-item">
                                    <label class="control-label">Compression</label>
                                    <select bind:value={compression} class="control-select">
                                        <option value="none">None</option>
                                        <option value="gzip">GZIP</option>
                                        <option value="snappy">Snappy</option>
                                        <option value="lz4">LZ4</option>
                                        <option value="zstd">ZSTD</option>
                                    </select>
                                </div>

                                <div class="control-item">
                                    <label class="control-label">Acks</label>
                                    <select bind:value={acks} class="control-select">
                                        <option value={0}>0 (No wait)</option>
                                        <option value={1}>1 (Leader)</option>
                                        <option value={-1}>All (All replicas)</option>
                                    </select>
                                </div>
                            </div>

                            <div class="control-row">
                                <div class="control-item full-width">
                                    <label class="control-label">Message Key (optional)</label>
                                    <input type="text" bind:value={messageKey} placeholder="key" class="control-input" />
                                </div>
                            </div>

                            <div class="control-row">
                                <div class="control-item full-width">
                                    <label class="control-label">Message Value</label>
                                    <textarea bind:value={messageValue} placeholder="Enter message value (JSON, text, etc.)" class="message-textarea"></textarea>
                                </div>
                            </div>

                            <div class="headers-section">
                                <div class="headers-header">
                                    <span class="control-label">Headers</span>
                                    <button class="add-btn" on:click={addHeader}>
                                        <Plus size={14} />
                                        Add
                                    </button>
                                </div>
                                {#if messageHeaders.length > 0}
                                    <div class="headers-list">
                                        {#each messageHeaders as header, i}
                                            <div class="header-row">
                                                <input type="checkbox" bind:checked={header.enabled} class="header-checkbox" />
                                                <input type="text" bind:value={header.key} placeholder="Header name" class="header-input" />
                                                <input type="text" bind:value={header.value} placeholder="Value" class="header-input" />
                                                <button class="remove-btn" on:click={() => removeHeader(i)}>
                                                    <Trash2 size={14} />
                                                </button>
                                            </div>
                                        {/each}
                                    </div>
                                {/if}
                            </div>

                            <div class="control-actions">
                                <button class="action-btn primary" on:click={handleProduceMessage} disabled={!produceTopic || !messageValue.trim()}>
                                    <Send size={16} />
                                    Produce Message
                                </button>
                            </div>
                        </div>
                    </div>
                {/if}
            </div>
        {/if}
    </div>
</div>

<style>
    .kafka-handler-wrapper {
        height: 100%;
        overflow-y: auto;
    }

    .kafka-handler-wrapper::-webkit-scrollbar {
        width: 12px;
    }

    .kafka-handler-wrapper::-webkit-scrollbar-track {
        background: #0a0a0a;
    }

    .kafka-handler-wrapper::-webkit-scrollbar-thumb {
        background: #27272a;
        border-radius: 6px;
        border: 3px solid #0a0a0a;
    }

    .kafka-handler-wrapper::-webkit-scrollbar-thumb:hover {
        background: #3f3f46;
    }

    .kafka-handler {
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

    .auth-badge {
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

    .connect-btn.connecting,
    .connect-btn.disconnecting {
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

    .setting-label,
    .control-label {
        font-size: 12px;
        font-weight: 600;
        color: #a1a1aa;
        display: flex;
        align-items: center;
        gap: 8px;
    }

    .setting-checkbox,
    .control-checkbox {
        width: 16px;
        height: 16px;
        cursor: pointer;
    }

    .setting-select,
    .setting-input,
    .control-select,
    .control-input {
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
    .setting-input:focus,
    .control-select:focus,
    .control-input:focus {
        border-color: #3f3f46;
    }

    .topics-section {
        display: flex;
        flex-direction: column;
        gap: 10px;
        background: #111111;
        border: 1px solid #27272a;
        border-radius: 8px;
        padding: 14px;
    }

    .section-header {
        display: flex;
        align-items: center;
        justify-content: space-between;
    }

    .section-title {
        font-size: 12px;
        font-weight: 700;
        color: #a1a1aa;
        text-transform: uppercase;
        letter-spacing: 0.5px;
    }

    .refresh-btn {
        display: flex;
        align-items: center;
        gap: 6px;
        padding: 6px 12px;
        background: #18181b;
        border: 1px solid #27272a;
        border-radius: 4px;
        color: #a1a1aa;
        font-size: 11px;
        font-weight: 600;
        cursor: pointer;
        transition: all 0.2s;
    }

    .refresh-btn:hover:not(:disabled) {
        background: #27272a;
        color: #e4e4e7;
    }

    .refresh-btn:disabled {
        opacity: 0.5;
        cursor: not-allowed;
    }

    :global(.spinning) {
        animation: spin 1s linear infinite;
    }

    .topic-select {
        padding: 10px 12px;
        background: #0a0a0a;
        border: 1px solid #27272a;
        border-radius: 6px;
        color: #e4e4e7;
        font-size: 13px;
        outline: none;
        font-family: 'SF Mono', Monaco, monospace;
    }

    .empty-topics {
        padding: 20px;
        text-align: center;
        color: #71717a;
        font-size: 12px;
    }

    .panel-section {
        display: flex;
        flex-direction: column;
        gap: 0;
        background: #111111;
        border: 1px solid #27272a;
        border-radius: 8px;
        overflow: hidden;
    }

    .panel-toggle {
        width: 100%;
        padding: 12px 14px;
        background: transparent;
        border: none;
        color: #a1a1aa;
        font-size: 12px;
        font-weight: 700;
        text-transform: uppercase;
        letter-spacing: 0.5px;
        text-align: left;
        cursor: pointer;
        transition: all 0.2s;
    }

    .panel-toggle:hover,
    .panel-toggle.active {
        background: rgba(255, 255, 255, 0.03);
        color: #e4e4e7;
    }

    .panel-content {
        padding: 14px;
        border-top: 1px solid #27272a;
    }

    .consumer-controls,
    .producer-controls {
        display: flex;
        flex-direction: column;
        gap: 12px;
    }

    .control-row {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
        gap: 12px;
    }

    .control-item {
        display: flex;
        flex-direction: column;
        gap: 6px;
    }

    .control-item.full-width {
        grid-column: 1 / -1;
    }

    .message-textarea {
        padding: 10px 12px;
        background: #0a0a0a;
        border: 1px solid #27272a;
        border-radius: 6px;
        color: #e4e4e7;
        font-size: 13px;
        font-family: 'SF Mono', Monaco, monospace;
        outline: none;
        resize: vertical;
        min-height: 100px;
        max-height: 300px;
    }

    .message-textarea:focus {
        border-color: #3f3f46;
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
        display: flex;
        align-items: center;
        gap: 4px;
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
        display: flex;
        align-items: center;
        justify-content: center;
        width: 24px;
        height: 24px;
        background: #18181b;
        border: 1px solid #27272a;
        border-radius: 4px;
        color: #ef4444;
        cursor: pointer;
        transition: all 0.2s;
    }

    .remove-btn:hover {
        background: rgba(239, 68, 68, 0.1);
        border-color: #ef4444;
    }

    .control-actions {
        display: flex;
        gap: 8px;
        margin-top: 8px;
    }

    .action-btn {
        display: flex;
        align-items: center;
        gap: 6px;
        padding: 10px 16px;
        border: none;
        border-radius: 6px;
        font-weight: 600;
        font-size: 13px;
        cursor: pointer;
        transition: all 0.2s;
    }

    .action-btn.primary {
        background: #3b82f6;
        color: white;
    }

    .action-btn.primary:hover:not(:disabled) {
        background: #2563eb;
    }

    .action-btn.danger {
        background: #dc2626;
        color: white;
    }

    .action-btn.danger:hover {
        background: #ef4444;
    }

    .action-btn:disabled {
        opacity: 0.5;
        cursor: not-allowed;
    }
</style>