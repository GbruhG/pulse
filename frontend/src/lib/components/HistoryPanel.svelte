<script lang="ts">
    import { Clock, Trash2 } from 'lucide-svelte';
    import { historyStore } from '../stores/history';
    import { requestStore } from '../stores/request';
    import type { HistoryItem } from '../types';

    function loadRequest(item: HistoryItem, e: Event) {
        e.stopPropagation();
        requestStore.setRequest(item.request);
        if (item.response) {
            requestStore.setResponse(item.response);
        } else {
            requestStore.setResponse(null);
        }
    }

    function deleteItem(id: string, e: Event) {
        e.stopPropagation();
        historyStore.removeItem(id);
    }

    function clearAll() {
        if (confirm('Clear all history?')) {
            historyStore.clearHistory();
        }
    }

    function formatTime(date: Date): string {
        const now = new Date();
        const diff = now.getTime() - new Date(date).getTime();
        const minutes = Math.floor(diff / 60000);
        const hours = Math.floor(diff / 3600000);
        const days = Math.floor(diff / 86400000);

        if (minutes < 1) return 'Just now';
        if (minutes < 60) return `${minutes}m ago`;
        if (hours < 24) return `${hours}h ago`;
        return `${days}d ago`;
    }

    function getMethodColor(method: string): string {
        switch (method) {
            case 'GET': return '#10b981';
            case 'POST': return '#3b82f6';
            case 'PUT': return '#f59e0b';
            case 'PATCH': return '#8b5cf6';
            case 'DELETE': return '#ef4444';
            default: return '#6b7280';
        }
    }
</script>

<div class="history-panel">
    <div class="panel-header">
        <div class="header-left">
            <Clock size={16} />
            <span>History</span>
            <span class="count">{$historyStore.length}</span>
        </div>
        {#if $historyStore.length > 0}
            <button class="clear-btn" on:click={clearAll}>
                <Trash2 size={14} />
                Clear
            </button>
        {/if}
    </div>

    <div class="history-list">
        {#if $historyStore.length === 0}
            <div class="empty-state">
                <Clock size={32} class="empty-icon" />
                <p>No history yet</p>
                <span>Your recent requests will appear here</span>
            </div>
        {:else}
            {#each $historyStore as item (item.id)}
                <div class="history-item">
                    <button class="item-content" on:click={(e) => loadRequest(item, e)}>
                        <div class="item-header">
              <span class="method" style="color: {getMethodColor(item.request.method)}">
                {item.request.method}
              </span>
                            <span class="time">{formatTime(item.timestamp)}</span>
                        </div>
                        <div class="item-url" title={item.request.url}>
                            {item.request.url || 'No URL'}
                        </div>
                        {#if item.response}
                            <div
                                    class="item-status"
                                    class:success={item.response.statusCode >= 200 && item.response.statusCode < 300}
                                    class:warning={item.response.statusCode >= 300 && item.response.statusCode < 400}
                                    class:error={item.response.statusCode >= 400}
                            >
                                {item.response.statusCode} • {item.response.time}
                            </div>
                        {/if}
                    </button>
                    <button class="delete-item-btn" on:click={(e) => deleteItem(item.id, e)}>
                        <Trash2 size={14} />
                    </button>
                </div>
            {/each}
        {/if}
    </div>
</div>

<style>
    .history-panel {
        display: flex;
        flex-direction: column;
        height: 100%;
    }

    .panel-header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: 1rem;
        border-bottom: 1px solid #2a2a2a;
    }

    .header-left {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        font-size: 0.875rem;
        font-weight: 600;
        color: #d1d5db;
    }

    .count {
        padding: 0.125rem 0.5rem;
        background: #1a1a1a;
        border-radius: 0.375rem;
        font-size: 0.75rem;
        color: #9ca3af;
    }

    .clear-btn {
        display: flex;
        align-items: center;
        gap: 0.375rem;
        padding: 0.375rem 0.75rem;
        background: transparent;
        border: 1px solid #2a2a2a;
        border-radius: 0.375rem;
        color: #9ca3af;
        font-size: 0.75rem;
        cursor: pointer;
        transition: all 0.2s;
    }

    .clear-btn:hover {
        background: #1a1a1a;
        border-color: #ef4444;
        color: #ef4444;
    }

    .history-list {
        flex: 1;
        overflow-y: auto;
    }

    .empty-state {
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        gap: 0.75rem;
        padding: 4rem 2rem;
        text-align: center;
    }

    .empty-state :global(.empty-icon) {
        color: #4b5563;
    }

    .empty-state p {
        margin: 0;
        font-size: 0.875rem;
        font-weight: 500;
        color: #9ca3af;
    }

    .empty-state span {
        font-size: 0.75rem;
        color: #6b7280;
    }

    .history-item {
        display: flex;
        align-items: stretch;
        border-bottom: 1px solid #1a1a1a;
        transition: background 0.2s;
    }

    .history-item:hover {
        background: #0f0f0f;
    }

    .item-content {
        flex: 1;
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
        padding: 1rem;
        background: transparent;
        border: none;
        text-align: left;
        cursor: pointer;
        color: inherit;
    }

    .item-header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 0.5rem;
    }

    .method {
        font-size: 0.75rem;
        font-weight: 700;
        letter-spacing: 0.05em;
    }

    .time {
        font-size: 0.7rem;
        color: #6b7280;
    }

    .item-url {
        font-size: 0.875rem;
        color: #d1d5db;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
        line-height: 1.4;
    }

    .item-status {
        font-size: 0.75rem;
        color: #9ca3af;
        font-weight: 500;
    }

    .item-status.success {
        color: #10b981;
    }

    .item-status.warning {
        color: #3b82f6;
    }

    .item-status.error {
        color: #ef4444;
    }

    .delete-item-btn {
        display: flex;
        align-items: center;
        justify-content: center;
        width: 40px;
        background: transparent;
        border: none;
        color: #6b7280;
        cursor: pointer;
        transition: all 0.2s;
    }

    .delete-item-btn:hover {
        background: #1a1a1a;
        color: #ef4444;
    }
</style>