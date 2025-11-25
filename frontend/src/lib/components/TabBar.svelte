<script lang="ts">
    import { Plus, X, MoreVertical, Copy, Edit2, Circle } from 'lucide-svelte';
    import { tabsStore, activeTab } from '../stores/tabs';
    import type { TabProtocol } from '../stores/tabs';

    let showTabMenu: string | null = null;
    let showRenameModal = false;
    let renameTabId = '';
    let renameValue = '';

    function handleNewTab() {
        tabsStore.createTab('http');
    }

    function handleCloseTab(tabId: string, e: Event) {
        e.stopPropagation();
        tabsStore.closeTab(tabId);
    }

    function handleTabClick(tabId: string) {
        tabsStore.setActiveTab(tabId);
    }

    function toggleTabMenu(tabId: string, e: Event) {
        e.stopPropagation();
        showTabMenu = showTabMenu === tabId ? null : tabId;
    }

    function handleDuplicate(tabId: string) {
        tabsStore.duplicateTab(tabId);
        showTabMenu = null;
    }

    function handleRename(tabId: string, currentName: string) {
        renameTabId = tabId;
        renameValue = currentName;
        showRenameModal = true;
        showTabMenu = null;
    }

    function confirmRename() {
        if (renameValue.trim()) {
            tabsStore.renameTab(renameTabId, renameValue.trim());
        }
        showRenameModal = false;
    }

    function getProtocolColor(protocol: TabProtocol): string {
        switch (protocol) {
            case 'http': return '#3b82f6';
            case 'websocket': return '#06b6d4';
            case 'sse': return '#14b8a6';
            case 'grpc-stream': return '#a855f7';
            case 'kafka': return '#f97316';
            case 'mqtt': return '#ec4899';
            case 'grpc': return '#8b5cf6';
            default: return '#6b7280';
        }
    }

    function getProtocolLabel(protocol: TabProtocol): string {
        switch (protocol) {
            case 'http': return 'HTTP';
            case 'websocket': return 'WS';
            case 'sse': return 'SSE';
            case 'grpc-stream': return 'gRPC';
            case 'kafka': return 'Kafka';
            case 'mqtt': return 'MQTT';
            case 'grpc': return 'gRPC';
            default: return protocol.toUpperCase();
        }
    }
</script>

<div class="tab-bar">
    <div class="tabs-container">
        {#each $tabsStore.tabs as tab (tab.id)}
            <div
                    class="tab"
                    class:active={tab.id === $tabsStore.activeTabId}
                    on:click={() => handleTabClick(tab.id)}
            >
                <div class="tab-indicator" style="background: {getProtocolColor(tab.protocol)}" />

                <div class="tab-content">
                    <div class="tab-info">
                        <span class="tab-protocol" style="color: {getProtocolColor(tab.protocol)}">
                            {getProtocolLabel(tab.protocol)}
                        </span>
                        <span class="tab-name">{tab.name}</span>
                        {#if tab.isDirty}
                            <Circle size={6} fill="currentColor" class="dirty-indicator" />
                        {/if}
                        {#if tab.isStreamConnected}
                            <div class="connection-indicator" title="Connected">
                                <Circle size={8} fill="#10b981" />
                            </div>
                        {/if}
                    </div>
                </div>

                <div class="tab-actions">
                    <button
                            class="tab-action-btn menu-btn"
                            on:click={(e) => toggleTabMenu(tab.id, e)}
                            title="Tab options"
                    >
                        <MoreVertical size={14} />
                    </button>

                    <button
                            class="tab-action-btn close-btn"
                            on:click={(e) => handleCloseTab(tab.id, e)}
                            title="Close tab"
                    >
                        <X size={14} />
                    </button>
                </div>

                {#if showTabMenu === tab.id}
                    <div class="tab-menu" on:click|stopPropagation>
                        <button class="menu-item" on:click={() => handleDuplicate(tab.id)}>
                            <Copy size={14} />
                            Duplicate
                        </button>
                        <button class="menu-item" on:click={() => handleRename(tab.id, tab.name)}>
                            <Edit2 size={14} />
                            Rename
                        </button>
                        <div class="menu-divider"></div>
                        <button class="menu-item danger" on:click={() => { handleCloseTab(tab.id, new Event('click')); showTabMenu = null; }}>
                            <X size={14} />
                            Close Tab
                        </button>
                    </div>
                {/if}
            </div>
        {/each}

        <button class="new-tab-btn" on:click={handleNewTab} title="New tab">
            <Plus size={16} />
        </button>
    </div>
</div>

{#if showRenameModal}
    <div class="modal-overlay" on:click={() => showRenameModal = false}>
        <div class="modal" on:click|stopPropagation>
            <h3>Rename Tab</h3>
            <input
                    type="text"
                    bind:value={renameValue}
                    class="rename-input"
                    placeholder="Tab name"
                    on:keydown={(e) => e.key === 'Enter' && confirmRename()}
                    autofocus
            />
            <div class="modal-actions">
                <button class="btn-secondary" on:click={() => showRenameModal = false}>Cancel</button>
                <button class="btn-primary" on:click={confirmRename}>Rename</button>
            </div>
        </div>
    </div>
{/if}

<svelte:window on:click={() => showTabMenu = null} />

<style>
    .tab-bar {
        background: #0a0a0a;
        border-bottom: 1px solid rgba(255, 255, 255, 0.08);
        padding: 0 1rem;
        overflow-x: auto;
        overflow-y: hidden;
    }

    .tabs-container {
        display: flex;
        align-items: center;
        gap: 2px;
        min-height: 40px;
    }

    .tab {
        position: relative;
        display: flex;
        align-items: center;
        gap: 6px;
        padding: 0.5rem 1rem;
        background: transparent;
        border: none;
        border-radius: 4px;
        min-width: 160px;
        max-width: 200px;
        cursor: pointer;
        transition: all 0.2s;
        user-select: none;
        color: #9ca3af;
        font-size: 0.875rem;
    }

    .tab:hover {
        background: rgba(255, 255, 255, 0.05);
        color: #e4e4e7;
    }

    .tab.active {
        background: #1a1a1a;
        color: #e4e4e7;
        border-bottom: 2px solid #ef4444;
    }

    .tab-indicator {
        width: 4px;
        height: 20px;
        border-radius: 2px;
        flex-shrink: 0;
    }

    .tab-content {
        flex: 1;
        min-width: 0;
    }

    .tab-info {
        display: flex;
        align-items: center;
        gap: 6px;
    }

    .tab-protocol {
        font-size: 0.6rem;
        font-weight: 600;
        letter-spacing: 0.5px;
        text-transform: uppercase;
        flex-shrink: 0;
        color: #71717a;
    }

    .tab-name {
        font-size: 0.875rem;
        color: #9ca3af;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
        flex: 1;
    }

    .tab.active .tab-name {
        color: #e4e4e7;
        font-weight: 500;
    }

    .dirty-indicator {
        color: #fbbf24;
        flex-shrink: 0;
    }

    .connection-indicator {
        display: flex;
        align-items: center;
        flex-shrink: 0;
    }

    .tab-actions {
        display: flex;
        gap: 2px;
        opacity: 0;
        transition: opacity 0.2s;
    }

    .tab:hover .tab-actions,
    .tab.active .tab-actions {
        opacity: 1;
    }

    .tab-action-btn {
        display: flex;
        align-items: center;
        justify-content: center;
        width: 18px;
        height: 18px;
        padding: 0;
        background: transparent;
        border: none;
        border-radius: 4px;
        color: #71717a;
        cursor: pointer;
        transition: all 0.2s;
    }

    .tab-action-btn:hover {
        background: rgba(255, 255, 255, 0.1);
        color: #e4e4e7;
    }

    .tab-action-btn.close-btn:hover {
        background: rgba(239, 68, 68, 0.1);
        color: #ef4444;
    }

    .tab-menu {
        position: absolute;
        top: 100%;
        right: 0;
        margin-top: 4px;
        background: #0f0f0f;
        border: 1px solid rgba(255, 255, 255, 0.1);
        border-radius: 6px;
        min-width: 160px;
        box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
        z-index: 1000;
        overflow: hidden;
    }

    .menu-item {
        width: 100%;
        display: flex;
        align-items: center;
        gap: 8px;
        padding: 0.5rem 1rem;
        background: transparent;
        border: none;
        color: #e4e4e7;
        font-size: 0.875rem;
        text-align: left;
        cursor: pointer;
        transition: all 0.2s;
    }

    .menu-item:hover {
        background: rgba(255, 255, 255, 0.1);
    }

    .menu-item.danger {
        color: #ef4444;
    }

    .menu-item.danger:hover {
        background: rgba(239, 68, 68, 0.1);
    }

    .menu-divider {
        height: 1px;
        background: rgba(255, 255, 255, 0.08);
        margin: 0.25rem 0;
    }

    .new-tab-btn {
        display: flex;
        align-items: center;
        justify-content: center;
        width: 32px;
        height: 32px;
        padding: 0;
        background: transparent;
        border: 1px solid rgba(255, 255, 255, 0.1);
        color: #71717a;
        cursor: pointer;
        border-radius: 4px;
        transition: all 0.2s;
        flex-shrink: 0;
    }

    .new-tab-btn:hover {
        background: rgba(255, 255, 255, 0.1);
        border-color: rgba(255, 255, 255, 0.2);
        color: #e4e4e7;
    }

    .modal-overlay {
        position: fixed;
        inset: 0;
        background: rgba(0, 0, 0, 0.7);
        display: flex;
        align-items: center;
        justify-content: center;
        z-index: 2000;
    }

    .modal {
        background: #0f0f0f;
        border: 1px solid rgba(255, 255, 255, 0.1);
        border-radius: 6px;
        padding: 1.5rem;
        width: 90%;
        max-width: 360px;
        box-shadow: 0 10px 30px rgba(0, 0, 0, 0.4);
    }

    .modal h3 {
        margin: 0 0 1rem 0;
        font-size: 1.1rem;
        font-weight: 600;
        color: #e4e4e7;
    }

    .rename-input {
        width: 100%;
        padding: 0.5rem 0.75rem;
        background: #0a0a0a;
        border: 1px solid rgba(255, 255, 255, 0.1);
        border-radius: 4px;
        color: white;
        font-size: 0.875rem;
        margin-bottom: 1rem;
        transition: all 0.2s;
        font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
    }

    .rename-input:focus {
        outline: none;
        border-color: rgba(59, 130, 246, 0.5);
    }

    .modal-actions {
        display: flex;
        gap: 0.5rem;
        justify-content: flex-end;
    }

    .btn-secondary,
    .btn-primary {
        padding: 0.5rem 1rem;
        border: none;
        border-radius: 4px;
        font-weight: 500;
        cursor: pointer;
        transition: all 0.2s;
        font-size: 0.875rem;
    }

    .btn-secondary {
        background: #2a2a2a;
        color: #e4e4e7;
    }

    .btn-secondary:hover {
        background: #3a3a3a;
    }

    .btn-primary {
        background: #2563eb;
        color: white;
    }

    .btn-primary:hover {
        background: #3b82f6;
    }

    .tab-bar::-webkit-scrollbar {
        height: 6px;
    }

    .tab-bar::-webkit-scrollbar-track {
        background: #0a0a0a;
    }

    .tab-bar::-webkit-scrollbar-thumb {
        background: #27272a;
        border-radius: 3px;
    }

    .tab-bar::-webkit-scrollbar-thumb:hover {
        background: #3f3f46;
    }
</style>