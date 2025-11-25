<script lang="ts">
    import { Upload, FileCode, Server, Settings, Send, X, Link, Link2Off, AlertCircle } from 'lucide-svelte';
    import { GrpcParseProtoFiles, GrpcUseReflection, GrpcConnect, GrpcSendMessage, GrpcDisconnect } from '../../../wailsjs/go/main/App';

    type StreamType = 'server' | 'client' | 'bidi' | 'unary';

    let serverUrl = '';
    let isConnected = false;
    let isConnecting = false;
    let connectionError = '';
    let connectionId = '';

    // Proto Management
    let protoFiles: File[] = [];
    let protoText = '';
    let showProtoInput = false;
    let showReflectionInput = false;
    let reflectionUrl = '';
    let reflectionUseTLS = false;
    let isLoadingReflection = false;

    // Parsed Schema
    interface ServiceInfo {
        name: string;
        methods: Array<{name: string, type: StreamType, inputType: string, outputType: string}>;
    }

    let services: ServiceInfo[] = [];
    let selectedService = '';
    let selectedMethod = '';

    // Connection Settings
    let showSettings = false;
    let useTLS = false;
    let deadline = 30000; // ms
    let compression: 'none' | 'gzip' | 'deflate' = 'none';

    // Metadata (Headers)
    let metadata: Array<{key: string, value: string, enabled: boolean}> = [];

    // Dynamic Message Builder
    let messageBody = '{\n  "field": "value"\n}';

    // Stream Control
    let streamType: StreamType = 'unary';
    let canSendMultiple = false;

    $: canSendMultiple = streamType === 'client' || streamType === 'bidi';
    $: currentMethod = services
        .find(s => s.name === selectedService)
        ?.methods.find(m => m.name === selectedMethod);

    $: if (currentMethod) {
        streamType = currentMethod.type;
    }

    function validateGrpcUrl(url: string): boolean {
        if (!url) return false;
        try {
            const urlObj = new URL(url);
            return urlObj.protocol === 'grpc:' || urlObj.protocol === 'grpcs:';
        } catch {
            // Allow bare hostname:port format
            return /^[a-zA-Z0-9.-]+:\d+$/.test(url);
        }
    }

    async function handleProtoUpload(event: Event) {
        const input = event.target as HTMLInputElement;
        if (!input.files || input.files.length === 0) return;

        protoFiles = Array.from(input.files);

        try {
            // Read file contents
            const filePromises = Array.from(protoFiles).map(file => {
                return new Promise<{name: string, content: string}>((resolve, reject) => {
                    const reader = new FileReader();
                    reader.onload = (e) => {
                        resolve({
                            name: file.name,
                            content: e.target?.result as string
                        });
                    };
                    reader.onerror = reject;
                    reader.readAsText(file);
                });
            });

            const fileContents = await Promise.all(filePromises);

            // Send to backend
            const response = await GrpcParseProtoFiles({ files: fileContents });
            services = response.services;
            connectionError = '';

            console.log('Parsed proto files:', services);
        } catch (error) {
            connectionError = `Failed to parse proto files: ${error}`;
            console.error(error);
        }
    }

    async function handleProtoTextSubmit() {
        if (!protoText.trim()) return;

        try {
            const response = await GrpcParseProtoFiles({
                files: [{ name: 'inline.proto', content: protoText }]
            });
            services = response.services;
            connectionError = '';
            showProtoInput = false;

            console.log('Parsed proto text:', services);
        } catch (error) {
            connectionError = `Failed to parse proto: ${error}`;
            console.error(error);
        }
    }

    async function handleUseReflection() {
        if (!reflectionUrl) {
            connectionError = 'Enter server URL first';
            return;
        }

        if (!validateGrpcUrl(reflectionUrl)) {
            connectionError = 'Invalid gRPC URL format';
            return;
        }

        isLoadingReflection = true;
        connectionError = '';

        try {
            const response = await GrpcUseReflection(reflectionUrl, reflectionUseTLS);
            services = response.services;
            showReflectionInput = false;

            console.log('Fetched reflection from:', reflectionUrl, services);
        } catch (error) {
            connectionError = `Reflection failed: ${error}`;
            console.error(error);
        } finally {
            isLoadingReflection = false;
        }
    }

    async function handleConnect() {
        if (isConnected) {
            // Disconnect
            try {
                if (connectionId) {
                    await GrpcDisconnect(connectionId);
                }
            } catch (error) {
                console.error('Disconnect error:', error);
            }

            isConnected = false;
            connectionId = '';
            connectionError = '';
            return;
        }

        if (!selectedService || !selectedMethod) {
            connectionError = 'Select a service and method first';
            return;
        }

        if (!validateGrpcUrl(serverUrl)) {
            connectionError = 'Invalid gRPC URL. Use grpc://host:port or grpcs://host:port';
            return;
        }

        connectionError = '';
        isConnecting = true;

        try {
            // Prepare metadata
            const metadataObj: Record<string, string> = {};
            metadata.filter(m => m.enabled && m.key).forEach(m => {
                metadataObj[m.key] = m.value;
            });

            // Connect via Wails
            connectionId = await GrpcConnect({
                serverUrl: serverUrl,
                service: selectedService,
                method: selectedMethod,
                useTLS: useTLS,
                deadline: deadline,
                compression: compression,
                metadata: metadataObj
            });

            isConnected = true;
            console.log('Connected with ID:', connectionId);
        } catch (error) {
            connectionError = `Connection failed: ${error}`;
            console.error(error);
        } finally {
            isConnecting = false;
        }
    }

    async function handleSendMessage() {
        if (!messageBody.trim()) return;
        if (!connectionId) {
            connectionError = 'Not connected';
            return;
        }

        try {
            // Validate JSON
            JSON.parse(messageBody);

            // Send via Wails
            await GrpcSendMessage({
                connectionId: connectionId,
                message: messageBody
            });

            console.log('Message sent');
        } catch (error) {
            if (error instanceof SyntaxError) {
                connectionError = 'Invalid JSON in message body';
            } else {
                connectionError = `Failed to send: ${error}`;
            }
            console.error(error);
        }
    }

    async function handleFinishStream() {
        if (!connectionId) return;

        try {
            // Call backend to half-close the stream
            await GrpcDisconnect(connectionId);
            console.log('Stream finished');
        } catch (error) {
            connectionError = `Failed to finish stream: ${error}`;
            console.error(error);
        }
    }

    function addMetadata() {
        metadata = [...metadata, { key: '', value: '', enabled: true }];
    }

    function removeMetadata(index: number) {
        metadata = metadata.filter((_, i) => i !== index);
    }

    function clearProtoFiles() {
        protoFiles = [];
        services = [];
        selectedService = '';
        selectedMethod = '';
    }
</script>

<div class="grpc-handler">
    <!-- Proto Management Section -->
    <div class="proto-section">
        <div class="section-header">
            <span class="section-title">Proto Definition</span>
            <div class="proto-actions">
                {#if protoFiles.length > 0}
                    <span class="proto-count">{protoFiles.length} file{protoFiles.length > 1 ? 's' : ''} loaded</span>
                    <button class="action-btn" on:click={clearProtoFiles}>
                        <X size={14} />
                    </button>
                {/if}
                <label class="upload-btn">
                    <Upload size={14} />
                    <span>Upload Proto</span>
                    <input type="file" accept=".proto" multiple on:change={handleProtoUpload} style="display: none;" />
                </label>
                <button class="action-btn" on:click={() => showProtoInput = !showProtoInput}>
                    <FileCode size={14} />
                    <span>Paste Proto</span>
                </button>
                <button class="action-btn reflection" on:click={() => showReflectionInput = !showReflectionInput}>
                    <Server size={14} />
                    <span>Use Reflection</span>
                </button>
            </div>
        </div>

        {#if showProtoInput}
            <div class="proto-input-modal">
                <textarea
                        bind:value={protoText}
                        class="proto-textarea"
                        placeholder={'syntax = "proto3";\n\nservice ExampleService {\n  rpc StreamData (Request) returns (stream Response);\n}'}
                />
                <div class="modal-actions">
                    <button class="cancel-btn" on:click={() => showProtoInput = false}>Cancel</button>
                    <button class="submit-btn" on:click={handleProtoTextSubmit}>Parse Proto</button>
                </div>
            </div>
        {/if}

        {#if showReflectionInput}
            <div class="reflection-modal">
                <div class="reflection-inputs">
                    <input
                            type="text"
                            bind:value={reflectionUrl}
                            class="reflection-url-input"
                            placeholder="localhost:50051 or grpc://localhost:50051"
                    />
                    <label class="tls-checkbox-label">
                        <input type="checkbox" bind:checked={reflectionUseTLS} class="tls-checkbox" />
                        <span>Use TLS</span>
                    </label>
                </div>
                <div class="modal-actions">
                    <button class="cancel-btn" on:click={() => showReflectionInput = false}>Cancel</button>
                    <button
                            class="submit-btn"
                            on:click={handleUseReflection}
                            disabled={isLoadingReflection}
                    >
                        {#if isLoadingReflection}
                            <span class="spinner-small"></span>
                            Fetching...
                        {:else}
                            Fetch Schema
                        {/if}
                    </button>
                </div>
            </div>
        {/if}

        {#if protoFiles.length > 0}
            <div class="proto-files-list">
                {#each protoFiles as file}
                    <div class="proto-file-item">
                        <FileCode size={14} />
                        <span>{file.name}</span>
                    </div>
                {/each}
            </div>
        {/if}
    </div>

    <!-- Service & Method Selection -->
    {#if services.length > 0}
        <div class="selection-section">
            <div class="selection-row">
                <div class="select-group">
                    <label class="select-label">Service</label>
                    <select bind:value={selectedService} class="select-input">
                        <option value="">Select service...</option>
                        {#each services as service}
                            <option value={service.name}>{service.name}</option>
                        {/each}
                    </select>
                </div>

                {#if selectedService}
                    <div class="select-group">
                        <label class="select-label">Method</label>
                        <select bind:value={selectedMethod} class="select-input">
                            <option value="">Select method...</option>
                            {#each services.find(s => s.name === selectedService)?.methods || [] as method}
                                <option value={method.name}>
                                    {method.name} ({method.type})
                                </option>
                            {/each}
                        </select>
                    </div>
                {/if}
            </div>

            {#if streamType && selectedMethod}
                <div class="stream-type-badge" class:server={streamType === 'server'} class:client={streamType === 'client'} class:bidi={streamType === 'bidi'} class:unary={streamType === 'unary'}>
                    {#if streamType === 'server'}
                        ← Server Streaming
                    {:else if streamType === 'client'}
                        → Client Streaming
                    {:else if streamType === 'bidi'}
                        ↔ Bidirectional Streaming
                    {:else}
                        • Unary Call
                    {/if}
                </div>
            {/if}
        </div>
    {/if}

    <!-- Connection Bar -->
    {#if selectedService && selectedMethod}
        <div class="connection-section">
            <div class="connection-bar">
                <div class="url-input-group">
                    <input
                            type="text"
                            bind:value={serverUrl}
                            class="url-input"
                            placeholder="grpc://localhost:50051 or localhost:50051"
                            disabled={isConnected}
                    />

                    {#if useTLS}
                        <div class="tls-badge">TLS</div>
                    {/if}
                </div>

                <button
                        class="settings-btn"
                        class:active={showSettings}
                        on:click={() => showSettings = !showSettings}
                        title="Connection settings"
                >
                    <Settings size={16} />
                </button>

                <button
                        class="connect-btn"
                        class:connected={isConnected}
                        class:connecting={isConnecting}
                        on:click={handleConnect}
                        disabled={isConnecting}
                >
                    {#if isConnecting}
                        <span class="spinner"></span>
                        <span>Connecting...</span>
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
                    <!-- TLS -->
                    <div class="setting-item">
                        <label class="setting-label">
                            <input type="checkbox" bind:checked={useTLS} class="setting-checkbox" />
                            Use TLS (grpcs://)
                        </label>
                    </div>

                    <!-- Deadline -->
                    <div class="setting-item">
                        <label class="setting-label">Deadline</label>
                        <div class="input-with-unit">
                            <input type="number" bind:value={deadline} class="setting-input" />
                            <span class="unit">ms</span>
                        </div>
                    </div>

                    <!-- Compression -->
                    <div class="setting-item">
                        <label class="setting-label">Compression</label>
                        <select bind:value={compression} class="setting-select">
                            <option value="none">None</option>
                            <option value="gzip">gzip</option>
                            <option value="deflate">deflate</option>
                        </select>
                    </div>
                </div>

                <!-- Metadata (Headers) -->
                <div class="metadata-section">
                    <div class="metadata-header">
                        <span class="setting-label">Metadata (Headers)</span>
                        <button class="add-btn" on:click={addMetadata}>+ Add</button>
                    </div>
                    {#if metadata.length > 0}
                        <div class="metadata-list">
                            {#each metadata as meta, i}
                                <div class="metadata-row">
                                    <input type="checkbox" bind:checked={meta.enabled} class="metadata-checkbox" />
                                    <input type="text" bind:value={meta.key} placeholder="Key" class="metadata-input" />
                                    <input type="text" bind:value={meta.value} placeholder="Value" class="metadata-input" />
                                    <button class="remove-btn" on:click={() => removeMetadata(i)}>×</button>
                                </div>
                            {/each}
                        </div>
                    {/if}
                </div>
            </div>
        {/if}

        <!-- Message Builder -->
        {#if isConnected}
            <div class="message-section">
                <div class="message-header">
                    <span class="section-title">Message Body (JSON)</span>
                    <div class="message-info">
                        {#if streamType === 'client' || streamType === 'bidi'}
                            <span class="info-badge">You can send multiple messages</span>
                        {/if}
                    </div>
                </div>

                <div class="message-input-wrapper">
                    <textarea
                            bind:value={messageBody}
                            class="message-input"
                            placeholder={'{\n  "field1": "value",\n  "field2": 123\n}'}
                    />
                    <div class="message-actions">
                        <button
                                class="send-btn"
                                on:click={handleSendMessage}
                                disabled={!messageBody.trim()}
                        >
                            <Send size={16} />
                            <span>Send Message</span>
                        </button>

                        {#if canSendMultiple}
                            <button class="finish-btn" on:click={handleFinishStream}>
                                <X size={16} />
                                <span>Finish Stream</span>
                            </button>
                        {/if}
                    </div>
                </div>
            </div>
        {/if}
    {/if}
</div>

<style>
    .grpc-handler {
        display: flex;
        flex-direction: column;
        gap: 12px;
        padding: 1rem;
    }

    /* Proto Section */
    .proto-section {
        background: #0f0f0f;
        border: 1px solid rgba(255, 255, 255, 0.08);
        border-radius: 4px;
        padding: 0.75rem;
    }

    .section-header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        margin-bottom: 0.5rem;
    }

    .section-title {
        font-size: 0.75rem;
        font-weight: 600;
        color: #a1a1aa;
        text-transform: uppercase;
        letter-spacing: 0.5px;
    }

    .proto-actions {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        flex-wrap: wrap;
    }

    .proto-count {
        font-size: 0.75rem;
        color: #71717a;
        font-family: 'SF Mono', Monaco, monospace;
    }

    .upload-btn,
    .action-btn {
        display: flex;
        align-items: center;
        gap: 0.375rem;
        padding: 0.375rem 0.75rem;
        background: transparent;
        border: 1px solid rgba(255, 255, 255, 0.1);
        border-radius: 4px;
        color: #71717a;
        font-size: 0.75rem;
        font-weight: 500;
        cursor: pointer;
        transition: all 0.2s;
    }

    .upload-btn:hover,
    .action-btn:hover {
        background: rgba(255, 255, 255, 0.05);
        border-color: rgba(255, 255, 255, 0.2);
        color: #e4e4e7;
    }

    .action-btn.reflection {
        color: #60a5fa;
        border-color: rgba(96, 165, 250, 0.3);
    }

    .action-btn.reflection:hover {
        background: rgba(96, 165, 250, 0.1);
        border-color: #60a5fa;
    }

    .proto-input-modal,
    .reflection-modal {
        margin-top: 10px;
        display: flex;
        flex-direction: column;
        gap: 10px;
        padding: 12px;
        background: #0a0a0a;
        border: 1px solid #222222;
        border-radius: 6px;
    }

    .reflection-inputs {
        display: flex;
        gap: 12px;
        align-items: center;
    }

    .reflection-url-input {
        flex: 1;
        padding: 8px 12px;
        background: #111111;
        border: 1px solid #27272a;
        border-radius: 6px;
        color: #e4e4e7;
        font-size: 12px;
        font-family: 'SF Mono', Monaco, monospace;
        outline: none;
    }

    .reflection-url-input:focus {
        border-color: #3f3f46;
    }

    .tls-checkbox-label {
        display: flex;
        align-items: center;
        gap: 6px;
        font-size: 12px;
        color: #a1a1aa;
        font-weight: 600;
        white-space: nowrap;
    }

    .tls-checkbox {
        width: 16px;
        height: 16px;
        cursor: pointer;
    }

    .proto-textarea {
        width: 100%;
        min-height: 200px;
        background: #111111;
        border: 1px solid #27272a;
        border-radius: 6px;
        color: #e4e4e7;
        padding: 12px;
        font-size: 12px;
        font-family: 'SF Mono', Monaco, monospace;
        resize: vertical;
        outline: none;
    }

    .proto-textarea:focus {
        border-color: #3f3f46;
    }

    .modal-actions {
        display: flex;
        justify-content: flex-end;
        gap: 8px;
    }

    .cancel-btn,
    .submit-btn {
        display: flex;
        align-items: center;
        gap: 6px;
        padding: 8px 16px;
        border-radius: 6px;
        font-size: 12px;
        font-weight: 600;
        cursor: pointer;
        transition: all 0.2s;
    }

    .cancel-btn {
        background: transparent;
        border: 1px solid #27272a;
        color: #a1a1aa;
    }

    .cancel-btn:hover {
        background: #18181b;
        color: #e4e4e7;
    }

    .submit-btn {
        background: #3b82f6;
        border: none;
        color: white;
    }

    .submit-btn:hover:not(:disabled) {
        background: #2563eb;
    }

    .submit-btn:disabled {
        opacity: 0.6;
        cursor: not-allowed;
    }

    .spinner-small {
        width: 12px;
        height: 12px;
        border: 2px solid rgba(255,255,255,0.3);
        border-top-color: white;
        border-radius: 50%;
        animation: spin 0.8s linear infinite;
    }

    .proto-files-list {
        display: flex;
        flex-direction: column;
        gap: 6px;
        margin-top: 10px;
    }

    .proto-file-item {
        display: flex;
        align-items: center;
        gap: 8px;
        padding: 8px 12px;
        background: #0a0a0a;
        border: 1px solid #222222;
        border-radius: 6px;
        color: #a1a1aa;
        font-size: 12px;
        font-family: 'SF Mono', Monaco, monospace;
    }

    /* Selection Section */
    .selection-section {
        display: flex;
        flex-direction: column;
        gap: 10px;
        background: #111111;
        border: 1px solid #27272a;
        border-radius: 8px;
        padding: 14px;
    }

    .selection-row {
        display: grid;
        grid-template-columns: 1fr 1fr;
        gap: 12px;
    }

    .select-group {
        display: flex;
        flex-direction: column;
        gap: 6px;
    }

    .select-label {
        font-size: 11px;
        font-weight: 600;
        color: #71717a;
        text-transform: uppercase;
        letter-spacing: 0.5px;
    }

    .select-input {
        padding: 8px 12px;
        background: #0a0a0a;
        border: 1px solid #27272a;
        border-radius: 6px;
        color: #e4e4e7;
        font-size: 12px;
        font-family: 'SF Mono', Monaco, monospace;
        cursor: pointer;
        outline: none;
    }

    .select-input:focus {
        border-color: #3f3f46;
    }

    .stream-type-badge {
        padding: 8px 12px;
        background: #0a0a0a;
        border-radius: 6px;
        font-size: 12px;
        font-weight: 600;
        text-align: center;
        border: 2px solid;
    }

    .stream-type-badge.server {
        border-color: #10b981;
        color: #10b981;
    }

    .stream-type-badge.client {
        border-color: #3b82f6;
        color: #3b82f6;
    }

    .stream-type-badge.bidi {
        border-color: #8b5cf6;
        color: #8b5cf6;
    }

    .stream-type-badge.unary {
        border-color: #71717a;
        color: #a1a1aa;
    }

    /* Connection Section */
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
        gap: 0.5rem;
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

    .tls-badge {
        font-size: 0.75rem;
        font-weight: 600;
        color: #10b981;
        background: rgba(16, 185, 129, 0.1);
        padding: 0.25rem 0.5rem;
        border-radius: 4px;
        border: 1px solid rgba(16, 185, 129, 0.2);
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
        gap: 0.5rem;
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
        gap: 0.5rem;
        padding: 0.5rem 0.75rem;
        background: rgba(239, 68, 68, 0.1);
        border: 1px solid rgba(239, 68, 68, 0.2);
        border-radius: 4px;
        color: #ef4444;
        font-size: 0.875rem;
        font-weight: 500;
    }

    /* Settings Panel */
    .settings-panel {
        display: flex;
        flex-direction: column;
        gap: 1rem;
        background: #0f0f0f;
        border: 1px solid rgba(255, 255, 255, 0.08);
        border-radius: 4px;
        padding: 1rem;
    }

    .settings-grid {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
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

    .input-with-unit {
        display: flex;
        align-items: center;
        gap: 8px;
    }

    .unit {
        font-size: 11px;
        color: #71717a;
    }

    /* Metadata Section */
    .metadata-section {
        display: flex;
        flex-direction: column;
        gap: 10px;
    }

    .metadata-header {
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

    .metadata-list {
        display: flex;
        flex-direction: column;
        gap: 6px;
    }

    .metadata-row {
        display: flex;
        align-items: center;
        gap: 8px;
    }

    .metadata-checkbox {
        width: 16px;
        height: 16px;
        cursor: pointer;
    }

    .metadata-input {
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

    .metadata-input:focus {
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

    /* Message Section */
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
    }

    .info-badge {
        font-size: 11px;
        color: #60a5fa;
        background: rgba(96, 165, 250, 0.1);
        padding: 4px 8px;
        border-radius: 4px;
        border: 1px solid rgba(96, 165, 250, 0.3);
    }

    .message-input-wrapper {
        display: flex;
        flex-direction: column;
        gap: 10px;
    }

    .message-input {
        width: 100%;
        background: #0a0a0a;
        border: 1px solid #27272a;
        border-radius: 6px;
        color: #e4e4e7;
        padding: 10px 12px;
        font-size: 13px;
        font-family: 'SF Mono', Monaco, monospace;
        outline: none;
        resize: vertical;
        min-height: 120px;
        max-height: 300px;
        transition: all 0.2s;
    }

    .message-input:focus {
        border-color: #3f3f46;
        background: #0f0f0f;
    }

    .message-input::placeholder {
        color: #52525b;
    }

    .message-actions {
        display: flex;
        gap: 10px;
    }

    .send-btn,
    .finish-btn {
        display: flex;
        align-items: center;
        gap: 6px;
        padding: 10px 20px;
        border-radius: 6px;
        font-weight: 700;
        font-size: 12px;
        cursor: pointer;
        transition: all 0.2s;
        white-space: nowrap;
        border: none;
    }

    .send-btn {
        background: #3b82f6;
        color: white;
    }

    .send-btn:hover:not(:disabled) {
        background: #2563eb;
        transform: translateY(-1px);
    }

    .send-btn:disabled {
        opacity: 0.5;
        cursor: not-allowed;
    }

    .finish-btn {
        background: #f59e0b;
        color: white;
    }

    .finish-btn:hover {
        background: #d97706;
        transform: translateY(-1px);
    }
</style>