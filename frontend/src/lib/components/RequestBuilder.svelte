<script lang="ts">
    import { createEventDispatcher } from 'svelte';
    import { Send, Save, ChevronDown, AlertTriangle } from 'lucide-svelte';
    import { requestStore } from '../stores/request';
    import { environmentStore } from '../stores/environment';
    import { extractVariables } from '../utils/variables';
    import { HTTP_METHODS } from '../types';

    export let activeTab = 'params';
    export let onSaveToCollection: () => void;

    const dispatch = createEventDispatcher();

    $: method = $requestStore.current.method;
    $: url = $requestStore.current.url;
    $: loading = $requestStore.loading;

    $: activeEnv = $environmentStore.environments.find(
        e => e.id === $environmentStore.activeEnvironmentId
    );
    $: variables = activeEnv?.variables || {};

    $: missingVariables = findMissingVariables();

    function findMissingVariables(): string[] {
        const current = $requestStore.current;
        const allText = [
            current.url,
            ...current.params.map(p => p.key + p.value),
            ...current.headers.map(h => h.key + h.value),
            current.body,
            current.auth?.username || '',
            current.auth?.password || '',
            current.auth?.token || '',
            current.auth?.value || ''
        ].join(' ');

        const usedVars = extractVariables(allText);
        return usedVars.filter(v => !(v in variables));
    }

    function updateMethod(newMethod: string) {
        requestStore.updateRequest({ method: newMethod });
    }

    function updateUrl(newUrl: string) {
        requestStore.updateRequest({ url: newUrl });
    }

    function handleSend() {
        dispatch('send');
    }

    const tabs = ['Params', 'Headers', 'Auth', 'Body'];
</script>

<div class="request-builder">
    <!-- Request Line -->
    <div class="request-line">
        <div class="request-input-group">
            <div class="method-selector-wrapper">
                <select
                        value={method}
                        on:change={(e) => updateMethod(e.currentTarget.value)}
                        class="method-selector"
                        class:get={method === 'GET'}
                        class:post={method === 'POST'}
                        class:put={method === 'PUT'}
                        class:patch={method === 'PATCH'}
                        class:delete={method === 'DELETE'}
                >
                    {#each HTTP_METHODS as m}
                        <option value={m}>{m}</option>
                    {/each}
                </select>
                <ChevronDown size={14} class="dropdown-icon" />
            </div>

            <input
                    type="text"
                    value={url}
                    on:input={(e) => updateUrl(e.currentTarget.value)}
                    class="url-input"
                    placeholder="https://api.example.com/endpoint"
            />
        </div>

        <button class="send-btn" on:click={handleSend} disabled={loading || !url}>
            <Send size={16} />
            <span>{loading ? 'Sending...' : 'Send'}</span>
        </button>

        <button class="save-btn" on:click={onSaveToCollection} title="Save to collection">
            <Save size={16} />
        </button>
    </div>

    <!-- Variable Warning -->
    {#if missingVariables.length > 0}
        <div class="variable-warning">
            <AlertTriangle size={14} />
            <div class="warning-content">
                <span class="warning-title">Missing variables:</span>
                <span class="warning-vars">{missingVariables.map(v => `{{${v}}}`).join(', ')}</span>
            </div>
        </div>
    {/if}

    <!-- Request Config Tabs -->
    <div class="tabs">
        {#each tabs as tab}
            <button
                    class="tab"
                    class:active={activeTab === tab.toLowerCase()}
                    on:click={() => activeTab = tab.toLowerCase()}
            >
                {tab}
                {#if tab === 'Params' && $requestStore.current.params.filter(p => p.enabled && p.key).length > 0}
                    <span class="badge">{$requestStore.current.params.filter(p => p.enabled && p.key).length}</span>
                {:else if tab === 'Headers' && $requestStore.current.headers.filter(h => h.enabled && h.key).length > 0}
                    <span class="badge">{$requestStore.current.headers.filter(h => h.enabled && h.key).length}</span>
                {/if}
            </button>
        {/each}
    </div>
</div>

<style>
    .request-builder {
        display: flex;
        flex-direction: column;
        gap: 12px;
        background: #0a0a0a;
        padding: 1rem;
        border-bottom: 1px solid rgba(255, 255, 255, 0.08);
    }

    /* Request Line */
    .request-line {
        display: flex;
        gap: 8px;
        align-items: center;
    }

    .request-input-group {
        flex: 1;
        display: flex;
        background: #0f0f0f;
        border: 1px solid rgba(255, 255, 255, 0.1);
        border-radius: 4px;
        overflow: hidden;
        transition: all 0.2s;
    }

    .request-input-group:focus-within {
        border-color: rgba(239, 68, 68, 0.5);
    }

    .method-selector-wrapper {
        position: relative;
        display: flex;
        align-items: center;
        background: #0a0a0a;
        border-right: 1px solid rgba(255, 255, 255, 0.1);
    }

    .method-selector {
        appearance: none;
        background: transparent;
        border: none;
        color: white;
        font-weight: 600;
        font-size: 0.875rem;
        padding: 0.625rem 2rem 0.625rem 0.75rem;
        cursor: pointer;
        outline: none;
        letter-spacing: 0.3px;
        min-width: 80px;
    }

    .method-selector.get {
        color: #10b981;
    }

    .method-selector.post {
        color: #3b82f6;
    }

    .method-selector.put {
        color: #f59e0b;
    }

    .method-selector.patch {
        color: #8b5cf6;
    }

    .method-selector.delete {
        color: #ef4444;
    }

    .dropdown-icon {
        position: absolute;
        right: 0.5rem;
        color: #52525b;
        pointer-events: none;
    }

    .url-input {
        flex: 1;
        background: transparent;
        border: none;
        color: #e4e4e7;
        padding: 0.625rem 0.75rem;
        font-size: 0.875rem;
        outline: none;
        font-family: 'SF Mono', Monaco, monospace;
    }

    .url-input::placeholder {
        color: #52525b;
    }

    .send-btn {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        background: #dc2626;
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

    .send-btn:hover:not(:disabled) {
        background: #ef4444;
    }

    .send-btn:disabled {
        opacity: 0.5;
        cursor: not-allowed;
    }

    .save-btn {
        display: flex;
        align-items: center;
        justify-content: center;
        padding: 0.625rem;
        background: transparent;
        border: 1px solid rgba(255, 255, 255, 0.1);
        color: #71717a;
        cursor: pointer;
        border-radius: 4px;
        transition: all 0.2s;
    }

    .save-btn:hover {
        background: rgba(255, 255, 255, 0.1);
        border-color: rgba(255, 255, 255, 0.2);
        color: #e4e4e7;
    }

    /* Variable Warning */
    .variable-warning {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        padding: 0.5rem 0.75rem;
        background: rgba(245, 158, 11, 0.1);
        border: 1px solid rgba(245, 158, 11, 0.2);
        border-radius: 4px;
        color: #fbbf24;
        font-size: 0.75rem;
    }

    .warning-content {
        display: flex;
        flex-direction: column;
        gap: 0.25rem;
        font-size: 0.75rem;
        line-height: 1.4;
    }

    .warning-title {
        font-weight: 600;
        color: #f59e0b;
    }

    .warning-vars {
        color: #fcd34d;
        font-family: 'SF Mono', Monaco, monospace;
        font-size: 0.7rem;
    }

    /* Tabs */
    .tabs {
        display: flex;
        gap: 2px;
        background: #0f0f0f;
        padding: 2px;
        border-radius: 4px;
        border: 1px solid rgba(255, 255, 255, 0.08);
    }

    .tab {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        background: transparent;
        border: none;
        color: #71717a;
        padding: 0.5rem 1rem;
        font-weight: 500;
        font-size: 0.875rem;
        cursor: pointer;
        transition: all 0.2s;
        border-radius: 2px;
        white-space: nowrap;
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
        display: flex;
        align-items: center;
        justify-content: center;
        min-width: 18px;
        height: 18px;
        padding: 0 0.25rem;
        background: rgba(255, 255, 255, 0.1);
        border-radius: 9px;
        font-size: 0.6rem;
        font-weight: 600;
        color: #71717a;
    }

    .tab.active .badge {
        background: rgba(255, 255, 255, 0.2);
        color: #a1a1aa;
    }
</style>