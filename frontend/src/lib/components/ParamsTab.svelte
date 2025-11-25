<script lang="ts">
    import { Plus, Trash2 } from 'lucide-svelte';
    import { requestStore } from '../stores/request';
    import { environmentStore } from '../stores/environment';
    import { substituteVariables } from '../utils/variables';
    import type { KeyValue } from '../types';

    $: params = $requestStore.current.params;
    $: url = $requestStore.current.url;

    $: activeEnv = $environmentStore.environments.find(
        e => e.id === $environmentStore.activeEnvironmentId
    );
    $: variables = activeEnv?.variables || {};

    $: previewUrl = buildPreviewUrl(url, params, variables);

    function buildPreviewUrl(baseUrl: string, params: KeyValue[], vars: Record<string, string>): string {
        if (!baseUrl) return 'https://api.example.com/endpoint';

        let resolvedUrl = substituteVariables(baseUrl, vars);
        const enabledParams = params.filter(p => p.enabled && p.key);

        if (enabledParams.length > 0) {
            const queryParams = enabledParams.map(p => {
                const key = substituteVariables(p.key, vars);
                const value = substituteVariables(p.value, vars);
                return `${encodeURIComponent(key)}=${encodeURIComponent(value)}`;
            }).join('&');

            resolvedUrl += (resolvedUrl.includes('?') ? '&' : '?') + queryParams;
        }

        return resolvedUrl;
    }

    function addParam() {
        const newParam: KeyValue = {
            id: crypto.randomUUID(),
            key: '',
            value: '',
            enabled: true,
            description: ''
        };
        requestStore.updateRequest({
            params: [...params, newParam]
        });
    }

    function removeParam(id: string) {
        requestStore.updateRequest({
            params: params.filter(p => p.id !== id)
        });
    }

    function updateParam(id: string, field: keyof KeyValue, value: any) {
        requestStore.updateRequest({
            params: params.map(p => p.id === id ? { ...p, [field]: value } : p)
        });
    }

    $: if (params.length === 0) {
        addParam();
    }
</script>

<div class="params-container">
    <div class="params-header">
        <span class="header-title">Query Parameters</span>
        <button class="add-button" on:click={addParam}>
            <Plus size={16} />
            Add Parameter
        </button>
    </div>

    <div class="params-table">
        <div class="table-header">
            <div class="col-checkbox"></div>
            <div class="col-key">Key</div>
            <div class="col-value">Value</div>
            <div class="col-description">Description</div>
            <div class="col-actions"></div>
        </div>

        {#each params as param (param.id)}
            <div class="table-row" class:disabled={!param.enabled}>
                <div class="col-checkbox">
                    <input
                            type="checkbox"
                            checked={param.enabled}
                            on:change={(e) => updateParam(param.id, 'enabled', e.currentTarget.checked)}
                    />
                </div>
                <div class="col-key">
                    <input
                            type="text"
                            placeholder="Key"
                            value={param.key}
                            on:input={(e) => updateParam(param.id, 'key', e.currentTarget.value)}
                            class="param-input"
                    />
                </div>
                <div class="col-value">
                    <input
                            type="text"
                            placeholder="Value"
                            value={param.value}
                            on:input={(e) => updateParam(param.id, 'value', e.currentTarget.value)}
                            class="param-input"
                    />
                </div>
                <div class="col-description">
                    <input
                            type="text"
                            placeholder="Description (optional)"
                            value={param.description || ''}
                            on:input={(e) => updateParam(param.id, 'description', e.currentTarget.value)}
                            class="param-input"
                    />
                </div>
                <div class="col-actions">
                    <button class="delete-btn" on:click={() => removeParam(param.id)}>
                        <Trash2 size={16} />
                    </button>
                </div>
            </div>
        {/each}
    </div>

    <div class="params-preview">
        <span class="preview-label">URL Preview:</span>
        <code class="preview-url">{previewUrl}</code>
    </div>
</div>

<style>
    .params-container {
        padding: 1.5rem;
        max-width: 100%;
    }

    .params-header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        margin-bottom: 1rem;
    }

    .header-title {
        font-size: 0.875rem;
        font-weight: 600;
        color: #9ca3af;
        text-transform: uppercase;
        letter-spacing: 0.05em;
    }

    .add-button {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        padding: 0.5rem 1rem;
        background: #1a1a1a;
        border: 1px solid #2a2a2a;
        border-radius: 0.5rem;
        color: white;
        font-size: 0.875rem;
        font-weight: 500;
        cursor: pointer;
        transition: all 0.2s;
    }

    .add-button:hover {
        background: #252525;
        border-color: #3a3a3a;
    }

    .params-table {
        background: #0f0f0f;
        border: 1px solid #2a2a2a;
        border-radius: 0.5rem;
        overflow: hidden;
    }

    .table-header {
        display: grid;
        grid-template-columns: 40px 1fr 1fr 1.5fr 40px;
        gap: 0.5rem;
        padding: 0.75rem 1rem;
        background: #141414;
        border-bottom: 1px solid #2a2a2a;
        font-size: 0.75rem;
        font-weight: 600;
        color: #6b7280;
        text-transform: uppercase;
        letter-spacing: 0.05em;
    }

    .table-row {
        display: grid;
        grid-template-columns: 40px 1fr 1fr 1.5fr 40px;
        gap: 0.5rem;
        padding: 0.75rem 1rem;
        border-bottom: 1px solid #1a1a1a;
        transition: background 0.2s;
    }

    .table-row:hover {
        background: #141414;
    }

    .table-row.disabled {
        opacity: 0.5;
    }

    .table-row:last-child {
        border-bottom: none;
    }

    .col-checkbox,
    .col-key,
    .col-value,
    .col-description,
    .col-actions {
        display: flex;
        align-items: center;
    }

    input[type="checkbox"] {
        width: 16px;
        height: 16px;
        cursor: pointer;
        accent-color: #ef4444;
    }

    .param-input {
        width: 100%;
        background: transparent;
        border: none;
        color: white;
        font-size: 0.875rem;
        padding: 0.375rem 0.5rem;
        border-radius: 0.25rem;
        transition: background 0.2s;
    }

    .param-input:focus {
        outline: none;
        background: #1a1a1a;
    }

    .param-input::placeholder {
        color: #4b5563;
    }

    .delete-btn {
        padding: 0.375rem;
        background: transparent;
        border: none;
        color: #6b7280;
        cursor: pointer;
        border-radius: 0.25rem;
        transition: all 0.2s;
    }

    .delete-btn:hover {
        background: #1a1a1a;
        color: #ef4444;
    }

    .params-preview {
        margin-top: 1.5rem;
        padding: 1rem;
        background: #0a0a0a;
        border: 1px solid #2a2a2a;
        border-radius: 0.5rem;
    }

    .preview-label {
        display: block;
        font-size: 0.75rem;
        font-weight: 600;
        color: #6b7280;
        margin-bottom: 0.5rem;
        text-transform: uppercase;
        letter-spacing: 0.05em;
    }

    .preview-url {
        display: block;
        font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
        font-size: 0.875rem;
        color: #10b981;
        word-break: break-all;
    }
</style>