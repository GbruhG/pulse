<script lang="ts">
import { onMount, onDestroy } from 'svelte';
import { ArrowDown, ArrowUp, AlertCircle, Info, Trash2, Pause, Play, Download, ArrowDownToLine, Zap } from 'lucide-svelte';
import * as runtime from '../../../wailsjs/runtime/runtime';
import { streamMessageStore, filteredMessages } from '../stores/streamMessages';

type MessageDirection = 'inbound' | 'outbound' | 'error' | 'system';

interface StreamMessage {
    id: string;
    direction: MessageDirection;
    protocol: string;
    payload: string;
    timestamp: Date;
}

export let isConnected = false;

let messagesContainer: HTMLDivElement;
let shouldAutoScroll = true;
let showScrollButton = false;
let messageCount = 0;
let messagesPerSecond = 0;
let lastCountUpdate = Date.now();
let messageCountInterval: number;
let autoScrollInterval: number;

// Force scroll every 200ms when auto-scroll is enabled
function startAutoScrollLoop() {
    if (autoScrollInterval) return;

    autoScrollInterval = setInterval(() => {
        if (shouldAutoScroll && messagesContainer) {
            const { scrollTop, scrollHeight, clientHeight } = messagesContainer;
            const isAtBottom = scrollHeight - scrollTop - clientHeight < 10;

            if (!isAtBottom) {
                messagesContainer.scrollTop = messagesContainer.scrollHeight;
            }
        }
    }, 200);
}

onMount(() => {
    console.log('[MessageViewer] 🟢 Component mounted');

    startAutoScrollLoop();

    messageCountInterval = setInterval(() => {
        const now = Date.now();
        const timeDiff = (now - lastCountUpdate) / 1000;
        const newCount = $streamMessageStore.messages.length;
        const msgDiff = newCount - messageCount;
        messagesPerSecond = Math.round(msgDiff / timeDiff);
        messageCount = newCount;
        lastCountUpdate = now;
    }, 1000);

    runtime.EventsOn('stream-message', (data: any) => {
        const message: StreamMessage = {
            id: data.id,
            direction: data.direction,
            protocol: data.protocol,
            payload: data.payload,
            timestamp: new Date(data.timestamp)
        };

        streamMessageStore.addMessage(message);
    });

    console.log('[MessageViewer] ✅ Listener registered');
});

onDestroy(() => {
    console.log('[MessageViewer] 🔴 Cleanup');
    runtime.EventsOff('stream-message');
    streamMessageStore.reset();
    if (messageCountInterval) clearInterval(messageCountInterval);
    if (autoScrollInterval) clearInterval(autoScrollInterval);
});

function checkScrollPosition() {
    if (!messagesContainer) return;
    const { scrollTop, scrollHeight, clientHeight } = messagesContainer;
    const distanceFromBottom = scrollHeight - scrollTop - clientHeight;
    shouldAutoScroll = distanceFromBottom < 50;
    showScrollButton = distanceFromBottom > 100;
}

function forceScrollToBottom() {
    if (!messagesContainer) return;
    shouldAutoScroll = true;
    showScrollButton = false;
    messagesContainer.scrollTop = messagesContainer.scrollHeight;
}

function formatTime(date: Date): string {
    return date.toLocaleTimeString('en-US', {
        hour12: false,
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
        fractionalSecondDigits: 3
    });
}

function getDirectionIcon(direction: MessageDirection) {
    switch (direction) {
        case 'inbound': return ArrowDown;
        case 'outbound': return ArrowUp;
        case 'error': return AlertCircle;
        case 'system': return Info;
    }
}

function getDirectionLabel(direction: MessageDirection): string {
    switch (direction) {
        case 'inbound': return 'Received';
        case 'outbound': return 'Sent';
        case 'error': return 'Error';
        case 'system': return 'System';
    }
}

function handleClear() {
    if (confirm('Clear all messages?')) {
        streamMessageStore.clear();
        messageCount = 0;
        messagesPerSecond = 0;
    }
}

function handleExport() {
    const dataStr = JSON.stringify($streamMessageStore.messages, null, 2);
    const dataBlob = new Blob([dataStr], { type: 'application/json' });
    const url = URL.createObjectURL(dataBlob);
    const link = document.createElement('a');
    link.href = url;
    link.download = `stream-messages-${Date.now()}.json`;
    link.click();
    URL.revokeObjectURL(url);
}
</script>

<div class="stream-viewer">
    <!-- Toolbar -->
    <div class="viewer-toolbar">
        <div class="filter-section">
            <select
                    value={$streamMessageStore.filterDirection}
                    on:change={(e) => streamMessageStore.setFilter(e.currentTarget.value)}
                    class="filter-select"
            >
                <option value="all">All Messages</option>
                <option value="inbound">Received</option>
                <option value="outbound">Sent</option>
                <option value="system">System</option>
                <option value="error">Errors</option>
            </select>

            <input
                    type="text"
                    value={$streamMessageStore.searchTerm}
                    on:input={(e) => streamMessageStore.setSearch(e.currentTarget.value)}
                    placeholder="Search messages..."
                    class="search-input"
            />
        </div>

        <div class="toolbar-right">
            <span class="message-count">
                {$filteredMessages.length} message{$filteredMessages.length !== 1 ? 's' : ''}
            </span>

            {#if $streamMessageStore.messages.length > 0}
                <button
                        class="tool-btn"
                        on:click={() => streamMessageStore.togglePause()}
                        title={$streamMessageStore.isPaused ? 'Resume' : 'Pause'}
                >
                    {#if $streamMessageStore.isPaused}
                        <Play size={14} />
                    {:else}
                        <Pause size={14} />
                    {/if}
                </button>

                <button class="tool-btn" on:click={handleExport} title="Export messages">
                    <Download size={14} />
                </button>

                <button class="tool-btn danger" on:click={handleClear} title="Clear all">
                    <Trash2 size={14} />
                </button>
            {/if}
        </div>
    </div>

    <!-- Messages List -->
    <div
            class="messages-container"
            bind:this={messagesContainer}
            on:scroll={checkScrollPosition}
    >
        {#if !isConnected && $streamMessageStore.messages.length === 0}
            <div class="empty-state">
                <svg width="64" height="64" viewBox="0 0 64 64" fill="none">
                    <path d="M16 32h32M32 16v32" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
                    <circle cx="32" cy="32" r="24" stroke="currentColor" stroke-width="2" fill="none"/>
                </svg>
                <p>Not Connected</p>
                <span>Connect to a streaming endpoint to see messages</span>
            </div>
        {:else if isConnected && $filteredMessages.length === 0 && $streamMessageStore.messages.length === 0}
            <div class="empty-state">
                <svg width="64" height="64" viewBox="0 0 64 64" fill="none">
                    <circle cx="32" cy="32" r="24" stroke="currentColor" stroke-width="2" fill="none"/>
                    <path d="M32 20v16l8 8" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
                </svg>
                <p>Listening...</p>
                <span>Waiting for messages from the stream</span>
            </div>
        {:else if $filteredMessages.length === 0}
            <div class="empty-state">
                <svg width="64" height="64" viewBox="0 0 64 64" fill="none">
                    <circle cx="32" cy="32" r="20" stroke="currentColor" stroke-width="2" fill="none"/>
                    <path d="M26 26l12 12M38 26l-12 12" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
                </svg>
                <p>No matches found</p>
                <span>Try adjusting your filters or search term</span>
            </div>
        {:else}
            {#if $streamMessageStore.isPaused}
                <div class="paused-banner">
                    <Pause size={14} />
                    Stream paused - new messages won't appear until resumed
                </div>
            {/if}

            {#each $filteredMessages as message (message.id)}
                <div class="message-item" class:inbound={message.direction === 'inbound'} class:outbound={message.direction === 'outbound'} class:error={message.direction === 'error'} class:system={message.direction === 'system'}>
                    <div class="message-header">
                        <div class="message-meta">
                            <span class="message-time">{formatTime(message.timestamp)}</span>
                            <span class="message-direction">
                                <svelte:component this={getDirectionIcon(message.direction)} size={12} />
                                {getDirectionLabel(message.direction)}
                            </span>
                            <span class="message-protocol">{message.protocol}</span>
                        </div>
                    </div>
                    <div class="message-body">
                        <pre>{message.payload}</pre>
                    </div>
                </div>
            {/each}
        {/if}
    </div>

    <!-- Scroll to Bottom Button -->
    {#if showScrollButton}
        <button class="scroll-to-bottom" on:click={forceScrollToBottom} title="Scroll to bottom">
            <ArrowDownToLine size={20} />
        </button>
    {/if}
</div>

<style>
    .stream-viewer {
        display: flex;
        flex-direction: column;
        height: 100%;
        background: #0a0a0a;
        color: #e4e4e7;
    }

    .viewer-toolbar {
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: 12px 16px;
        background: #111111;
        border-bottom: 1px solid #222222;
        gap: 12px;
        flex-shrink: 0;
    }

    .filter-section {
        display: flex;
        gap: 10px;
        flex: 1;
    }

    .filter-select {
        padding: 7px 12px;
        background: #0a0a0a;
        border: 1px solid #27272a;
        border-radius: 6px;
        color: #e4e4e7;
        font-size: 12px;
        font-weight: 600;
        cursor: pointer;
        outline: none;
        transition: all 0.2s;
    }

    .filter-select:focus {
        border-color: #3f3f46;
    }

    .search-input {
        flex: 1;
        max-width: 300px;
        padding: 7px 12px;
        background: #0a0a0a;
        border: 1px solid #27272a;
        border-radius: 6px;
        color: #e4e4e7;
        font-size: 12px;
        font-family: 'SF Mono', Monaco, monospace;
        outline: none;
        transition: all 0.2s;
    }

    .search-input:focus {
        border-color: #3f3f46;
        background: #0f0f0f;
    }

    .search-input::placeholder {
        color: #52525b;
    }

    .toolbar-right {
        display: flex;
        align-items: center;
        gap: 10px;
    }

    .message-count {
        font-size: 11px;
        font-weight: 600;
        color: #71717a;
        text-transform: uppercase;
        letter-spacing: 0.5px;
    }

    .tool-btn {
        display: flex;
        align-items: center;
        justify-content: center;
        width: 28px;
        height: 28px;
        background: #18181b;
        border: 1px solid #27272a;
        border-radius: 6px;
        color: #a1a1aa;
        cursor: pointer;
        transition: all 0.2s;
    }

    .tool-btn:hover {
        background: #27272a;
        border-color: #3f3f46;
        color: #e4e4e7;
    }

    .tool-btn.danger:hover {
        background: rgba(239, 68, 68, 0.1);
        border-color: #ef4444;
        color: #ef4444;
    }

    .messages-container {
        flex: 1;
        overflow-y: auto;
        overflow-x: hidden;
        padding: 12px;
        display: flex;
        flex-direction: column;
        gap: 8px;
    }

    .empty-state {
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        height: 100%;
        color: #52525b;
        text-align: center;
    }

    .empty-state svg {
        margin-bottom: 1rem;
        opacity: 0.3;
    }

    .empty-state p {
        margin: 0;
        font-size: 14px;
        font-weight: 600;
        color: #71717a;
    }

    .empty-state span {
        margin-top: 4px;
        font-size: 12px;
    }

    .paused-banner {
        display: flex;
        align-items: center;
        justify-content: center;
        gap: 8px;
        padding: 10px;
        background: rgba(251, 191, 36, 0.1);
        border: 1px solid rgba(251, 191, 36, 0.3);
        border-radius: 6px;
        color: #fbbf24;
        font-size: 12px;
        font-weight: 600;
        margin-bottom: 8px;
    }

    .message-item {
        background: #111111;
        border: 1px solid #27272a;
        border-radius: 8px;
        padding: 12px;
        transition: all 0.2s;
    }

    .message-item:hover {
        background: #141414;
        border-color: #3f3f46;
    }

    .message-item.inbound {
        border-left: 3px solid #10b981;
    }

    .message-item.outbound {
        border-left: 3px solid #3b82f6;
    }

    .message-item.error {
        border-left: 3px solid #ef4444;
        background: rgba(239, 68, 68, 0.05);
    }

    .message-item.system {
        border-left: 3px solid #71717a;
        background: #0f0f0f;
    }

    .message-header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        margin-bottom: 8px;
    }

    .message-meta {
        display: flex;
        align-items: center;
        gap: 12px;
        font-size: 11px;
        font-weight: 600;
    }

    .message-time {
        color: #71717a;
        font-family: 'SF Mono', Monaco, monospace;
    }

    .message-direction {
        display: flex;
        align-items: center;
        gap: 4px;
        padding: 3px 8px;
        border-radius: 4px;
        font-size: 10px;
        text-transform: uppercase;
        letter-spacing: 0.5px;
    }

    .message-item.inbound .message-direction {
        background: rgba(16, 185, 129, 0.15);
        color: #10b981;
    }

    .message-item.outbound .message-direction {
        background: rgba(59, 130, 246, 0.15);
        color: #3b82f6;
    }

    .message-item.error .message-direction {
        background: rgba(239, 68, 68, 0.15);
        color: #ef4444;
    }

    .message-item.system .message-direction {
        background: rgba(113, 113, 122, 0.15);
        color: #a1a1aa;
    }

    .message-protocol {
        color: #a1a1aa;
        font-size: 10px;
        text-transform: uppercase;
        letter-spacing: 0.5px;
    }

    .message-body {
        background: #0a0a0a;
        border: 1px solid #222222;
        border-radius: 6px;
        padding: 10px;
    }

    .message-body pre {
        margin: 0;
        font-family: 'SF Mono', Monaco, monospace;
        font-size: 12px;
        line-height: 1.6;
        color: #e4e4e7;
        white-space: pre-wrap;
        word-wrap: break-word;
    }

    .messages-container::-webkit-scrollbar {
        width: 12px;
    }

    .messages-container::-webkit-scrollbar-track {
        background: #0a0a0a;
    }

    .messages-container::-webkit-scrollbar-thumb {
        background: #27272a;
        border-radius: 6px;
        border: 3px solid #0a0a0a;
    }

    .messages-container::-webkit-scrollbar-thumb:hover {
        background: #3f3f46;
    }

    .scroll-to-bottom {
        position: absolute;
        bottom: 24px;
        right: 24px;
        width: 48px;
        height: 48px;
        background: #3b82f6;
        border: 2px solid #2563eb;
        border-radius: 50%;
        color: white;
        cursor: pointer;
        display: flex;
        align-items: center;
        justify-content: center;
        box-shadow: 0 4px 12px rgba(59, 130, 246, 0.4);
        transition: all 0.2s;
        z-index: 10;
    }

    .scroll-to-bottom:hover {
        background: #2563eb;
        border-color: #1d4ed8;
        transform: scale(1.1);
        box-shadow: 0 6px 16px rgba(59, 130, 246, 0.6);
    }

    .scroll-to-bottom:active {
        transform: scale(0.95);
    }

    .performance-badge {
        display: flex;
        align-items: center;
        gap: 4px;
        padding: 4px 8px;
        background: rgba(34, 197, 94, 0.15);
        border: 1px solid rgba(34, 197, 94, 0.3);
        border-radius: 4px;
        color: #22c55e;
        font-size: 11px;
        font-weight: 600;
        font-family: 'SF Mono', Monaco, monospace;
    }

    .performance-badge.high-rate {
        background: rgba(251, 191, 36, 0.15);
        border-color: rgba(251, 191, 36, 0.3);
        color: #fbbf24;
        animation: pulse 2s ease-in-out infinite;
    }

    @keyframes pulse {
        0%, 100% {
            opacity: 1;
        }
        50% {
            opacity: 0.7;
        }
    }
</style>