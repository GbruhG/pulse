<script lang="ts">
    import { Copy, Check } from 'lucide-svelte';
    import { writable } from 'svelte/store';
    import JsonNode from './JsonNode.svelte';
    import type { ResponseData } from '../types';

    export let response: ResponseData | null = null;

    let viewMode: 'pretty' | 'raw' = 'pretty';
    let activeTab: 'body' | 'headers' = 'body';
    let copied = false;
    let searchTerm = '';
    let expandedStore = writable(new Set<string>());

    $: parsedBody = response ? parseBody(response.body) : null;
    $: isJSON = response ? isJSONResponse(response) : false;
    $: filteredHeaders = response && searchTerm
        ? Object.entries(response.headers).filter(
            ([key, value]) =>
                key.toLowerCase().includes(searchTerm.toLowerCase()) ||
                value.toLowerCase().includes(searchTerm.toLowerCase())
        )
        : response ? Object.entries(response.headers) : [];

    function isJSONResponse(resp: ResponseData): boolean {
        const contentType = resp.headers['Content-Type'] || resp.headers['content-type'] || '';
        return contentType.includes('application/json');
    }

    function parseBody(body: string) {
        if (!body) return null;
        try {
            return JSON.parse(body);
        } catch {
            return null;
        }
    }

    async function copyToClipboard() {
        if (!response) return;
        try {
            await navigator.clipboard.writeText(response.body);
            copied = true;
            setTimeout(() => copied = false, 2000);
        } catch (err) {
            console.error('Failed to copy:', err);
        }
    }

    function getStatusColor(status: number): string {
        if (status >= 200 && status < 300) return '#10b981';
        if (status >= 300 && status < 400) return '#3b82f6';
        if (status >= 400 && status < 500) return '#f59e0b';
        return '#ef4444';
    }
</script>

{#if response}
    <div class="response-viewer">
        <!-- Header Bar -->
        <div class="header-bar">
            <div class="status-section">
                <div
                        class="status-badge"
                        style="
                        background-color: {getStatusColor(response.statusCode)}15;
                        color: {getStatusColor(response.statusCode)};
                        border: 1px solid {getStatusColor(response.statusCode)}30;
                    "
                >
                    <div class="status-dot" style="background-color: {getStatusColor(response.statusCode)}"></div>
                    {response.statusCode} {response.statusText}
                </div>
                <div class="metrics">
                    <div class="metric">
                        <span class="metric-icon">⚡</span>
                        <span>{response.time}</span>
                    </div>
                    <div class="metric">
                        <span class="metric-icon">📦</span>
                        <span>{response.size}</span>
                    </div>
                </div>
            </div>

            <div class="controls-section">
                <div class="tab-group">
                    <button
                            class="tab"
                            class:active={activeTab === 'body'}
                            on:click={() => activeTab = 'body'}
                    >
                        Body
                    </button>
                    <button
                            class="tab"
                            class:active={activeTab === 'headers'}
                            on:click={() => activeTab = 'headers'}
                    >
                        Headers
                        <span class="badge">{Object.keys(response.headers).length}</span>
                    </button>
                </div>

                {#if activeTab === 'body' && isJSON && parsedBody}
                    <div class="view-mode-toggle">
                        <button
                                class:active={viewMode === 'pretty'}
                                on:click={() => viewMode = 'pretty'}
                        >
                            Pretty
                        </button>
                        <button
                                class:active={viewMode === 'raw'}
                                on:click={() => viewMode = 'raw'}
                        >
                            Raw
                        </button>
                    </div>
                {/if}

                <button class="icon-btn" on:click={copyToClipboard} title="Copy to clipboard">
                    {#if copied}
                        <Check size={16} />
                    {:else}
                        <Copy size={16} />
                    {/if}
                </button>
            </div>
        </div>

        <!-- Content Area -->
        <div class="content-area">
            {#if activeTab === 'body'}
                {#if isJSON && parsedBody && viewMode === 'pretty'}
                    <div class="json-pretty">
                        <JsonNode data={parsedBody} level={0} path="root" {expandedStore} />
                    </div>
                {:else}
                    <pre class="json-raw">{response.body}</pre>
                {/if}
            {:else}
                <div class="headers-view">
                    <div class="headers-search">
                        <input
                                type="text"
                                placeholder="Filter headers..."
                                bind:value={searchTerm}
                                class="search-input"
                        />
                    </div>
                    <div class="headers-grid">
                        {#each filteredHeaders as [key, value]}
                            <div class="header-row">
                                <div class="header-key">{key}</div>
                                <div class="header-value">{value}</div>
                            </div>
                        {/each}
                    </div>
                </div>
            {/if}
        </div>
    </div>
{:else}
    <div class="response-viewer empty">
        <div class="empty-state">
            <svg width="64" height="64" viewBox="0 0 64 64" fill="none">
                <path d="M32 8L8 20V44L32 56L56 44V20L32 8Z" stroke="currentColor" stroke-width="2" fill="none"/>
                <path d="M32 32L32 56" stroke="currentColor" stroke-width="2"/>
                <path d="M32 32L56 20" stroke="currentColor" stroke-width="2"/>
                <path d="M32 32L8 20" stroke="currentColor" stroke-width="2"/>
            </svg>
            <p>Send a request to see the response</p>
        </div>
    </div>
{/if}

<style>
    .response-viewer {
        display: flex;
        flex-direction: column;
        height: 100%;
        background: #0a0a0a;
        color: #e4e4e7;
        font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', system-ui, sans-serif;
    }

    .response-viewer.empty {
        align-items: center;
        justify-content: center;
        background: #0a0a0a;
    }

    .empty-state {
        text-align: center;
        color: #52525b;
    }

    .empty-state svg {
        margin-bottom: 1rem;
        opacity: 0.3;
    }

    .empty-state p {
        margin: 0;
        font-size: 0.875rem;
        color: #52525b;
    }

    .header-bar {
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: 0.75rem 1rem;
        background: #0a0a0a;
        border-bottom: 1px solid rgba(255, 255, 255, 0.08);
        gap: 1rem;
        flex-wrap: wrap;
    }

    .status-section {
        display: flex;
        align-items: center;
        gap: 1rem;
    }

    .status-badge {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        padding: 0.25rem 0.75rem;
        border-radius: 0.25rem;
        font-size: 0.875rem;
        font-weight: 600;
        background: rgba(0, 0, 0, 0.2);
        border: 1px solid rgba(255, 255, 255, 0.08);
    }

    .status-dot {
        width: 6px;
        height: 6px;
        border-radius: 50%;
    }

    .metrics {
        display: flex;
        gap: 1rem;
    }

    .metric {
        display: flex;
        align-items: center;
        gap: 0.25rem;
        font-size: 0.75rem;
        color: #71717a;
        font-weight: 500;
        font-family: 'SF Mono', Monaco, 'Cascadia Code', monospace;
    }

    .metric-icon {
        font-size: 0.875rem;
    }

    .controls-section {
        display: flex;
        align-items: center;
        gap: 0.75rem;
    }

    .tab-group {
        display: flex;
        gap: 2px;
        background: #0a0a0a;
        padding: 2px;
        border-radius: 4px;
        border: 1px solid rgba(255, 255, 255, 0.08);
    }

    .tab {
        padding: 0.25rem 0.75rem;
        background: transparent;
        border: none;
        color: #71717a;
        font-size: 0.875rem;
        font-weight: 500;
        cursor: pointer;
        border-radius: 2px;
        transition: all 0.2s;
        display: flex;
        align-items: center;
        gap: 0.25rem;
    }

    .tab:hover {
        color: #e4e4e7;
        background: rgba(255, 255, 255, 0.05);
    }

    .tab.active {
        color: #e4e4e7;
        background: rgba(255, 255, 255, 0.1);
    }

    .badge {
        background: rgba(255, 255, 255, 0.1);
        color: #71717a;
        padding: 0.125rem 0.375rem;
        border-radius: 0.25rem;
        font-size: 0.65rem;
        font-weight: 600;
    }

    .tab.active .badge {
        background: rgba(255, 255, 255, 0.2);
        color: #a1a1aa;
    }

    .view-mode-toggle {
        display: flex;
        gap: 2px;
        background: #0a0a0a;
        padding: 2px;
        border-radius: 4px;
        border: 1px solid rgba(255, 255, 255, 0.08);
    }

    .view-mode-toggle button {
        padding: 0.25rem 0.75rem;
        background: transparent;
        border: none;
        color: #71717a;
        font-size: 0.75rem;
        font-weight: 500;
        cursor: pointer;
        border-radius: 2px;
        transition: all 0.2s;
    }

    .view-mode-toggle button:hover {
        color: #e4e4e7;
        background: rgba(255, 255, 255, 0.05);
    }

    .view-mode-toggle button.active {
        background: rgba(255, 255, 255, 0.1);
        color: #e4e4e7;
    }

    .icon-btn {
        padding: 0.5rem;
        background: transparent;
        border: 1px solid rgba(255, 255, 255, 0.1);
        color: #71717a;
        cursor: pointer;
        border-radius: 4px;
        transition: all 0.2s;
        display: flex;
        align-items: center;
        justify-content: center;
    }

    .icon-btn:hover {
        background: rgba(255, 255, 255, 0.1);
        color: #e4e4e7;
        border-color: rgba(255, 255, 255, 0.2);
    }

    .content-area {
        flex: 1;
        overflow: auto;
        background: #0a0a0a;
    }

    /* JSON Pretty View */
    .json-pretty {
        padding: 1rem;
        font-family: 'SF Mono', Monaco, 'Cascadia Code', 'Courier New', monospace;
        font-size: 0.875rem;
        line-height: 1.6;
    }

    /* JSON Raw View */
    .json-raw {
        padding: 1rem;
        margin: 0;
        font-family: 'SF Mono', Monaco, 'Cascadia Code', monospace;
        font-size: 0.875rem;
        line-height: 1.6;
        color: #e4e4e7;
        white-space: pre-wrap;
        word-wrap: break-word;
    }

    /* Headers View */
    .headers-view {
        display: flex;
        flex-direction: column;
        height: 100%;
    }

    .headers-search {
        padding: 0.75rem 1rem;
        border-bottom: 1px solid rgba(255, 255, 255, 0.08);
        background: #0a0a0a;
        position: sticky;
        top: 0;
        z-index: 10;
    }

    .search-input {
        width: 100%;
        padding: 0.375rem 0.5rem;
        background: #0a0a0a;
        border: 1px solid rgba(255, 255, 255, 0.1);
        border-radius: 4px;
        color: #e4e4e7;
        font-size: 0.75rem;
        font-family: 'SF Mono', Monaco, monospace;
        transition: all 0.2s;
    }

    .search-input:focus {
        outline: none;
        border-color: rgba(59, 130, 246, 0.5);
    }

    .search-input::placeholder {
        color: #52525b;
    }

    .headers-grid {
        flex: 1;
        overflow: auto;
    }

    .header-row {
        display: grid;
        grid-template-columns: 200px 1fr;
        gap: 1rem;
        padding: 0.5rem 1rem;
        border-bottom: 1px solid rgba(255, 255, 255, 0.05);
        transition: background 0.1s;
        align-items: start;
    }

    .header-row:hover {
        background: rgba(255, 255, 255, 0.02);
    }

    .header-key {
        color: #60a5fa;
        font-weight: 500;
        font-size: 0.75rem;
        font-family: 'SF Mono', Monaco, monospace;
        word-break: break-word;
        line-height: 1.4;
    }

    .header-value {
        color: #a1a1aa;
        font-size: 0.75rem;
        font-family: 'SF Mono', Monaco, monospace;
        word-break: break-all;
        line-height: 1.4;
    }

    /* Scrollbar */
    .content-area::-webkit-scrollbar,
    .headers-grid::-webkit-scrollbar {
        width: 6px;
        height: 6px;
    }

    .content-area::-webkit-scrollbar-track,
    .headers-grid::-webkit-scrollbar-track {
        background: #0a0a0a;
    }

    .content-area::-webkit-scrollbar-thumb,
    .headers-grid::-webkit-scrollbar-thumb {
        background: #27272a;
        border-radius: 3px;
    }

    .content-area::-webkit-scrollbar-thumb:hover,
    .headers-grid::-webkit-scrollbar-thumb:hover {
        background: #3f3f46;
    }
</style>

<!-- Separate JsonNode Component -->
<script lang="ts" context="module">
    export interface JsonNodeProps {
        data: any;
        level: number;
        path: string;
    }
</script>