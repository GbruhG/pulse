<script lang="ts">
    import { FolderOpen, ChevronRight, ChevronDown, Plus, Trash2, FileText, X } from 'lucide-svelte';
    import { collectionStore } from '../stores/collection';
    import { workspaceStore } from '../stores/workspace';
    import { requestStore } from '../stores/request';
    import type { Collection, CollectionRequest } from '../types';

    let expandedCollections = new Set<string>();
    let showNewCollectionModal = false;
    let newCollectionName = '';

    function toggleCollection(id: string) {
        if (expandedCollections.has(id)) {
            expandedCollections.delete(id);
        } else {
            expandedCollections.add(id);
        }
        expandedCollections = expandedCollections;
    }

    function createCollection() {
        if (!newCollectionName.trim() || !$workspaceStore.activeWorkspaceId) return;

        const collection: Collection = {
            id: crypto.randomUUID(),
            name: newCollectionName,
            requests: [],
            workspaceId: $workspaceStore.activeWorkspaceId,
            createdAt: new Date()
        };

        collectionStore.addCollection(collection);
        newCollectionName = '';
        showNewCollectionModal = false;
    }

    function deleteCollection(id: string, e: Event) {
        e.stopPropagation();
        if (confirm('Delete this collection and all its requests?')) {
            collectionStore.removeCollection(id);
        }
    }

    function loadRequest(collectionRequest: CollectionRequest, e: Event) {
        e.stopPropagation();

        // Dispatch event to parent (App.svelte) to handle protocol switching
        window.dispatchEvent(new CustomEvent('loadRequest', {
            detail: collectionRequest
        }));
    }

    function deleteRequest(collectionId: string, requestId: string, e: Event) {
        e.stopPropagation();
        const collection = $collectionStore.find(c => c.id === collectionId);
        if (!collection) return;

        const updatedRequests = collection.requests.filter(r => r.id !== requestId);
        collectionStore.updateCollection(collectionId, { requests: updatedRequests });
    }

    function getMethodColor(method: string): string {
        switch (method.toUpperCase()) {
            case 'GET': return '#10b981';
            case 'POST': return '#3b82f6';
            case 'PUT': return '#f59e0b';
            case 'PATCH': return '#8b5cf6';
            case 'DELETE': return '#ef4444';
            case 'WSS': return '#06b6d4';
            case 'SSE': return '#14b8a6';
            case 'GRPC': return '#a855f7';
            case 'KAFKA': return '#f97316';
            case 'MQTT': return '#ec4899';
            default: return '#6b7280';
        }
    }

    $: workspaceCollections = $collectionStore.filter(
        c => c.workspaceId === $workspaceStore.activeWorkspaceId
    );
</script>

<div class="collections-panel">
    <div class="panel-header">
        <div class="header-left">
            <FolderOpen size={16} />
            <span>Collections</span>
        </div>
        <button class="add-btn" on:click={() => showNewCollectionModal = true}>
            <Plus size={14} />
        </button>
    </div>

    <div class="collections-list">
        {#if workspaceCollections.length === 0}
            <div class="empty-state">
                <FolderOpen size={32} class="empty-icon" />
                <p>No collections yet</p>
                <span>Create a collection to organize your requests</span>
                <button class="create-first-btn" on:click={() => showNewCollectionModal = true}>
                    <Plus size={16} />
                    Create Collection
                </button>
            </div>
        {:else}
            {#each workspaceCollections as collection (collection.id)}
                <div class="collection-item">
                    <div class="collection-header">
                        <button class="expand-btn" on:click={() => toggleCollection(collection.id)}>
                            {#if expandedCollections.has(collection.id)}
                                <ChevronDown size={16} />
                            {:else}
                                <ChevronRight size={16} />
                            {/if}
                        </button>
                        <button class="collection-name" on:click={() => toggleCollection(collection.id)}>
                            <FolderOpen size={16} />
                            <span>{collection.name}</span>
                            <span class="request-count">{collection.requests.length}</span>
                        </button>
                        <button class="delete-collection-btn" on:click={(e) => deleteCollection(collection.id, e)}>
                            <Trash2 size={14} />
                        </button>
                    </div>

                    {#if expandedCollections.has(collection.id)}
                        <div class="collection-requests">
                            {#if collection.requests.length === 0}
                                <div class="no-requests">No requests yet</div>
                            {:else}
                                {#each collection.requests as request (request.id)}
                                    <div class="request-item-container">
                                        <button class="request-item" on:click={(e) => loadRequest(request, e)}>
                                            <FileText size={14} />
                                            <span class="request-name">{request.name}</span>
                                            <span class="request-method" style="color: {getMethodColor(request.request.method)}">
                                                {request.request.method}
                                            </span>
                                        </button>
                                        <button
                                                class="delete-request-btn"
                                                on:click={(e) => deleteRequest(collection.id, request.id, e)}
                                                title="Delete request"
                                        >
                                            <X size={14} />
                                        </button>
                                    </div>
                                {/each}
                            {/if}
                        </div>
                    {/if}
                </div>
            {/each}
        {/if}
    </div>
</div>

{#if showNewCollectionModal}
    <div class="modal-overlay" on:click={() => showNewCollectionModal = false}>
        <div class="modal" on:click|stopPropagation>
            <h2>Create New Collection</h2>
            <input
                    type="text"
                    bind:value={newCollectionName}
                    placeholder="Collection name"
                    class="modal-input"
                    on:keydown={(e) => e.key === 'Enter' && createCollection()}
            />
            <div class="modal-actions">
                <button class="btn-secondary" on:click={() => showNewCollectionModal = false}>
                    Cancel
                </button>
                <button class="btn-primary" on:click={createCollection}>
                    Create
                </button>
            </div>
        </div>
    </div>
{/if}

<style>
    .collections-panel {
        display: flex;
        flex-direction: column;
        height: 100%;
    }

    .panel-header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: 0.75rem;
        border-bottom: 1px solid rgba(255, 255, 255, 0.08);
    }

    .header-left {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        font-size: 0.875rem;
        font-weight: 500;
        color: #d1d5db;
    }

    .add-btn {
        display: flex;
        align-items: center;
        justify-content: center;
        width: 24px;
        height: 24px;
        background: transparent;
        border: 1px solid rgba(255, 255, 255, 0.1);
        border-radius: 4px;
        color: #9ca3af;
        cursor: pointer;
        transition: all 0.2s;
    }

    .add-btn:hover {
        background: rgba(255, 255, 255, 0.05);
        border-color: rgba(239, 68, 68, 0.3);
        color: #e4e4e7;
    }

    .collections-list {
        flex: 1;
        overflow-y: auto;
    }

    .empty-state {
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        gap: 0.5rem;
        padding: 2rem 1rem;
        text-align: center;
    }

    .empty-state :global(.empty-icon) {
        color: #52525b;
    }

    .empty-state p {
        margin: 0;
        font-size: 0.875rem;
        font-weight: 500;
        color: #9ca3af;
    }

    .empty-state span {
        font-size: 0.75rem;
        color: #71717a;
    }

    .create-first-btn {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        padding: 0.5rem 1rem;
        margin-top: 0.5rem;
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

    .collection-item {
        border-bottom: 1px solid rgba(255, 255, 255, 0.05);
    }

    .collection-header {
        display: flex;
        align-items: center;
        gap: 0.25rem;
    }

    .expand-btn {
        display: flex;
        align-items: center;
        justify-content: center;
        width: 28px;
        height: 28px;
        background: transparent;
        border: none;
        color: #71717a;
        cursor: pointer;
        transition: all 0.2s;
    }

    .expand-btn:hover {
        color: #e4e4e7;
    }

    .collection-name {
        flex: 1;
        display: flex;
        align-items: center;
        gap: 0.5rem;
        padding: 0.5rem 0.25rem 0.5rem 0;
        background: transparent;
        border: none;
        color: #d1d5db;
        font-size: 0.875rem;
        font-weight: 500;
        text-align: left;
        cursor: pointer;
        transition: color 0.2s;
    }

    .collection-name:hover {
        color: #e4e4e7;
    }

    .request-count {
        margin-left: auto;
        padding: 0.125rem 0.375rem;
        background: rgba(255, 255, 255, 0.1);
        border-radius: 4px;
        font-size: 0.7rem;
        color: #9ca3af;
    }

    .delete-collection-btn {
        display: flex;
        align-items: center;
        justify-content: center;
        width: 28px;
        height: 28px;
        background: transparent;
        border: none;
        color: #71717a;
        cursor: pointer;
        transition: all 0.2s;
    }

    .delete-collection-btn:hover {
        color: #ef4444;
    }

    .collection-requests {
        padding-left: 2rem;
        padding-bottom: 0.25rem;
    }

    .no-requests {
        padding: 0.5rem 0;
        font-size: 0.75rem;
        color: #71717a;
        font-style: italic;
    }

    .request-item-container {
        display: flex;
        align-items: center;
        gap: 0.25rem;
        margin-bottom: 0.25rem;
    }

    .request-item {
        flex: 1;
        display: flex;
        align-items: center;
        gap: 0.5rem;
        padding: 0.5rem 0.75rem;
        background: transparent;
        border: none;
        border-radius: 4px;
        color: #9ca3af;
        font-size: 0.875rem;
        text-align: left;
        cursor: pointer;
        transition: all 0.2s;
    }

    .request-item:hover {
        background: rgba(255, 255, 255, 0.05);
        color: #e4e4e7;
    }

    .request-name {
        flex: 1;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
    }

    .request-method {
        font-size: 0.7rem;
        font-weight: 700;
        letter-spacing: 0.05em;
    }

    .delete-request-btn {
        display: flex;
        align-items: center;
        justify-content: center;
        width: 24px;
        background: transparent;
        border: none;
        color: #71717a;
        cursor: pointer;
        border-radius: 4px;
        transition: all 0.2s;
        opacity: 0;
    }

    .request-item-container:hover .delete-request-btn {
        opacity: 1;
    }

    .delete-request-btn:hover {
        background: rgba(255, 255, 255, 0.05);
        color: #ef4444;
    }

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
        padding: 1.5rem;
        width: 90%;
        max-width: 400px;
    }

    .modal h2 {
        margin: 0 0 1rem 0;
        font-size: 1.1rem;
        font-weight: 600;
        color: #e4e4e7;
    }

    .modal-input {
        width: 100%;
        background: #0f0f0f;
        border: 1px solid rgba(255, 255, 255, 0.1);
        border-radius: 4px;
        padding: 0.5rem 0.75rem;
        color: #e4e4e7;
        font-size: 0.875rem;
        margin-bottom: 1rem;
        transition: border-color 0.2s;
    }

    .modal-input:focus {
        outline: none;
        border-color: rgba(239, 68, 68, 0.5);
    }

    .modal-actions {
        display: flex;
        gap: 0.5rem;
        justify-content: flex-end;
    }

    .btn-secondary,
    .btn-primary {
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

    .btn-primary:hover {
        background: #ef4444;
    }
</style>