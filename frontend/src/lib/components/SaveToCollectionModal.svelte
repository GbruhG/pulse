<script lang="ts">
    import { createEventDispatcher } from 'svelte';
    import { Save, Plus, FolderOpen } from 'lucide-svelte';
    import { collectionStore } from '../stores/collection';
    import { workspaceStore } from '../stores/workspace';
    import { requestStore } from '../stores/request';
    import { streamingStore } from '../stores/streaming';
    import type { Collection, CollectionRequest } from '../types';

    export let show = false;
    export let isStreamingMode = false; // New prop to detect if saving streaming request

    const dispatch = createEventDispatcher();

    let requestName = '';
    let selectedCollectionId = '';
    let showNewCollection = false;
    let newCollectionName = '';

    $: workspaceCollections = $collectionStore.filter(
        c => c.workspaceId === $workspaceStore.activeWorkspaceId
    );

    $: if (workspaceCollections.length > 0 && !selectedCollectionId) {
        selectedCollectionId = workspaceCollections[0].id;
    }

    function createNewCollection() {
        if (!newCollectionName.trim() || !$workspaceStore.activeWorkspaceId) return;

        const collection: Collection = {
            id: crypto.randomUUID(),
            name: newCollectionName,
            requests: [],
            workspaceId: $workspaceStore.activeWorkspaceId,
            createdAt: new Date()
        };

        collectionStore.addCollection(collection);
        selectedCollectionId = collection.id;
        newCollectionName = '';
        showNewCollection = false;
    }

    function getProtocolMethodName(protocol: string): string {
        switch (protocol) {
            case 'websocket': return 'WSS';
            case 'sse': return 'SSE';
            case 'grpc-stream': return 'GRPC';
            case 'kafka': return 'KAFKA';
            case 'mqtt': return 'MQTT';
            default: return 'WSS';
        }
    }

    function saveRequest() {
        if (!requestName.trim() || !selectedCollectionId) return;

        let requestData;

        if (isStreamingMode) {
            // Save streaming request
            const protocol = $streamingStore.protocol;
            requestData = {
                method: getProtocolMethodName(protocol),
                url: $streamingStore.url,
                params: [],
                headers: $streamingStore.config?.headers || [],
                body: '',
                bodyType: 'streaming',
                auth: null,
                streamingConfig: {
                    protocol: protocol,
                    config: $streamingStore.config
                }
            };
        } else {
            // Save HTTP request
            requestData = $requestStore.current;
        }

        const collectionRequest: CollectionRequest = {
            id: crypto.randomUUID(),
            name: requestName,
            collectionId: selectedCollectionId,
            request: requestData
        };

        const collection = $collectionStore.find(c => c.id === selectedCollectionId);
        if (collection) {
            collectionStore.updateCollection(selectedCollectionId, {
                requests: [...collection.requests, collectionRequest]
            });
        }

        dispatch('saved');
        close();
    }

    function close() {
        show = false;
        requestName = '';
        showNewCollection = false;
    }

    // Auto-generate name from URL
    $: if (show && !requestName) {
        if (isStreamingMode && $streamingStore.url) {
            try {
                const url = new URL($streamingStore.url);
                const protocol = getProtocolMethodName($streamingStore.protocol);
                const path = url.pathname.split('/').filter(Boolean).pop() || url.hostname;
                requestName = `${protocol} ${path}`;
            } catch {
                requestName = `${getProtocolMethodName($streamingStore.protocol)} Connection`;
            }
        } else if ($requestStore.current.url) {
            try {
                const url = new URL($requestStore.current.url);
                const path = url.pathname.split('/').filter(Boolean).pop() || 'request';
                requestName = `${$requestStore.current.method} ${path}`;
            } catch {
                requestName = `${$requestStore.current.method} Request`;
            }
        }
    }
</script>

{#if show}
    <div class="modal-overlay" on:click={close}>
        <div class="modal" on:click|stopPropagation>
            <div class="modal-header">
                <div class="header-icon">
                    <Save size={24} />
                </div>
                <div>
                    <h2>Save Request</h2>
                    <p class="subtitle">Save this {isStreamingMode ? 'streaming connection' : 'request'} to a collection for later use</p>
                </div>
            </div>

            <div class="modal-body">
                <div class="form-group">
                    <label>Request Name</label>
                    <input
                            type="text"
                            bind:value={requestName}
                            placeholder="e.g., Binance BTC Stream"
                            class="form-input"
                            autofocus
                    />
                    <span class="hint">
                        {#if isStreamingMode}
                            {getProtocolMethodName($streamingStore.protocol)} {$streamingStore.url}
                        {:else}
                            {$requestStore.current.method} {$requestStore.current.url}
                        {/if}
                    </span>
                </div>

                <div class="form-group">
                    <label>Collection</label>
                    {#if !showNewCollection}
                        <div class="collection-select-wrapper">
                            <select bind:value={selectedCollectionId} class="form-select">
                                {#each workspaceCollections as collection}
                                    <option value={collection.id}>{collection.name}</option>
                                {/each}
                            </select>
                            <button class="new-collection-btn" on:click={() => showNewCollection = true}>
                                <Plus size={16} />
                                New Collection
                            </button>
                        </div>
                    {:else}
                        <div class="new-collection-input-group">
                            <input
                                    type="text"
                                    bind:value={newCollectionName}
                                    placeholder="Collection name"
                                    class="form-input"
                                    on:keydown={(e) => e.key === 'Enter' && createNewCollection()}
                            />
                            <button class="btn-small btn-primary" on:click={createNewCollection}>
                                Create
                            </button>
                            <button class="btn-small btn-secondary" on:click={() => showNewCollection = false}>
                                Cancel
                            </button>
                        </div>
                    {/if}
                </div>

                {#if workspaceCollections.length === 0 && !showNewCollection}
                    <div class="empty-collections">
                        <FolderOpen size={32} class="empty-icon" />
                        <p>No collections in this workspace</p>
                        <button class="create-first-btn" on:click={() => showNewCollection = true}>
                            <Plus size={16} />
                            Create Your First Collection
                        </button>
                    </div>
                {/if}
            </div>

            <div class="modal-actions">
                <button class="btn-secondary" on:click={close}>
                    Cancel
                </button>
                <button
                        class="btn-primary"
                        on:click={saveRequest}
                        disabled={!requestName.trim() || !selectedCollectionId}
                >
                    <Save size={16} />
                    Save Request
                </button>
            </div>
        </div>
    </div>
{/if}

<style>
    .modal-overlay {
        position: fixed;
        inset: 0;
        background: rgba(0, 0, 0, 0.7);
        display: flex;
        align-items: center;
        justify-content: center;
        z-index: 1000;
    }

    .modal {
        background: #0a0a0a;
        border: 1px solid rgba(255, 255, 255, 0.08);
        border-radius: 6px;
        width: 90%;
        max-width: 500px;
    }

    .modal-header {
        display: flex;
        gap: 0.75rem;
        padding: 1rem;
        border-bottom: 1px solid rgba(255, 255, 255, 0.08);
    }

    .header-icon {
        display: flex;
        align-items: center;
        justify-content: center;
        width: 36px;
        height: 36px;
        background: rgba(239, 68, 68, 0.1);
        border-radius: 4px;
        color: #ef4444;
    }

    .modal-header h2 {
        margin: 0 0 0.125rem 0;
        font-size: 1.1rem;
        font-weight: 600;
        color: #e4e4e7;
    }

    .subtitle {
        margin: 0;
        font-size: 0.875rem;
        color: #9ca3af;
    }

    .modal-body {
        padding: 1rem;
    }

    .form-group {
        margin-bottom: 1.25rem;
    }

    .form-group:last-child {
        margin-bottom: 0;
    }

    .form-group label {
        display: block;
        font-size: 0.875rem;
        font-weight: 500;
        color: #d1d5db;
        margin-bottom: 0.5rem;
    }

    .form-input,
    .form-select {
        width: 100%;
        background: #0f0f0f;
        border: 1px solid rgba(255, 255, 255, 0.1);
        border-radius: 4px;
        padding: 0.5rem 0.75rem;
        color: #e4e4e7;
        font-size: 0.875rem;
        transition: border-color 0.2s;
    }

    .form-input:focus,
    .form-select:focus {
        outline: none;
        border-color: rgba(239, 68, 68, 0.5);
    }

    .hint {
        display: block;
        margin-top: 0.25rem;
        font-size: 0.75rem;
        color: #71717a;
        font-family: 'Monaco', 'Menlo', monospace;
        word-break: break-all;
    }

    .collection-select-wrapper {
        display: flex;
        gap: 0.5rem;
    }

    .form-select {
        flex: 1;
    }

    .new-collection-btn {
        display: flex;
        align-items: center;
        gap: 0.25rem;
        padding: 0.5rem 0.75rem;
        background: transparent;
        border: 1px solid rgba(255, 255, 255, 0.1);
        border-radius: 4px;
        color: #9ca3af;
        font-size: 0.875rem;
        font-weight: 500;
        cursor: pointer;
        transition: all 0.2s;
        white-space: nowrap;
    }

    .new-collection-btn:hover {
        background: rgba(255, 255, 255, 0.05);
        border-color: rgba(255, 255, 255, 0.2);
        color: #e4e4e7;
    }

    .new-collection-input-group {
        display: flex;
        gap: 0.25rem;
    }

    .new-collection-input-group .form-input {
        flex: 1;
    }

    .btn-small {
        padding: 0.375rem 0.75rem;
        border-radius: 4px;
        font-weight: 500;
        cursor: pointer;
        transition: all 0.2s;
        border: none;
        font-size: 0.875rem;
    }

    .btn-small.btn-primary {
        background: #dc2626;
        color: white;
    }

    .btn-small.btn-primary:hover {
        background: #ef4444;
    }

    .btn-small.btn-secondary {
        background: transparent;
        border: 1px solid rgba(255, 255, 255, 0.1);
        color: #9ca3af;
    }

    .btn-small.btn-secondary:hover {
        background: rgba(255, 255, 255, 0.05);
        border-color: rgba(255, 255, 255, 0.2);
        color: #e4e4e7;
    }

    .empty-collections {
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 0.75rem;
        padding: 1rem;
        background: #0f0f0f;
        border: 1px dashed rgba(255, 255, 255, 0.08);
        border-radius: 4px;
        text-align: center;
    }

    .empty-collections :global(.empty-icon) {
        color: #52525b;
    }

    .empty-collections p {
        margin: 0;
        font-size: 0.875rem;
        color: #9ca3af;
    }

    .create-first-btn {
        display: flex;
        align-items: center;
        gap: 0.25rem;
        padding: 0.5rem 0.75rem;
        background: transparent;
        border: 1px solid rgba(239, 68, 68, 0.3);
        border-radius: 4px;
        color: #ef4444;
        font-weight: 500;
        cursor: pointer;
        transition: all 0.2s;
    }

    .create-first-btn:hover {
        background: rgba(239, 68, 68, 0.1);
    }

    .modal-actions {
        display: flex;
        gap: 0.5rem;
        justify-content: flex-end;
        padding: 1rem;
        border-top: 1px solid rgba(255, 255, 255, 0.08);
    }

    .btn-secondary,
    .btn-primary {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        padding: 0.5rem 1rem;
        border-radius: 4px;
        font-weight: 500;
        cursor: pointer;
        transition: all 0.2s;
        border: none;
    }

    .btn-secondary {
        background: transparent;
        border: 1px solid rgba(255, 255, 255, 0.1);
        color: #9ca3af;
    }

    .btn-secondary:hover {
        background: rgba(255, 255, 255, 0.05);
        border-color: rgba(255, 255, 255, 0.2);
        color: #e4e4e7;
    }

    .btn-primary {
        background: #dc2626;
        color: white;
    }

    .btn-primary:hover:not(:disabled) {
        background: #ef4444;
    }

    .btn-primary:disabled {
        opacity: 0.5;
        cursor: not-allowed;
    }
</style>