<script lang="ts">
    import { Plus, Trash2 } from 'lucide-svelte';
    import { requestStore } from '../stores/request';
    import type { KeyValue } from '../types';

    $: headers = $requestStore.current.headers;

    const commonHeaders = [
        { key: 'Content-Type', value: 'application/json' },
        { key: 'Accept', value: 'application/json' },
        { key: 'Authorization', value: 'Bearer {{token}}' },
        { key: 'User-Agent', value: 'Pulse/1.0' },
        { key: 'Accept-Language', value: 'en-US,en;q=0.9' },
    ];

    function addHeader(preset?: { key: string; value: string }) {
        const newHeader: KeyValue = {
            id: crypto.randomUUID(),
            key: preset?.key || '',
            value: preset?.value || '',
            enabled: true,
            description: ''
        };
        requestStore.updateRequest({
            headers: [...headers, newHeader]
        });
    }

    function removeHeader(id: string) {
        requestStore.updateRequest({
            headers: headers.filter(h => h.id !== id)
        });
    }

    function updateHeader(id: string, field: keyof KeyValue, value: any) {
        requestStore.updateRequest({
            headers: headers.map(h => h.id === id ? { ...h, [field]: value } : h)
        });
    }

    $: if (headers.length === 0) {
        addHeader();
    }
</script>

<div class="headers-container">
    <div class="headers-header">
        <div>
            <span class="header-title">Request Headers</span>
            <p class="header-subtitle">Headers are sent with your HTTP request</p>
        </div>
        <button class="add-button" on:click={() => addHeader()}>
            <Plus size={16} />
            Add Header
        </button>
    </div>

    <div class="common-headers">
        <span class="common-label">Quick Add:</span>
        <div class="common-buttons">
            {#each commonHeaders as header}
                <button class="common-btn" on:click={() => addHeader(header)}>
                    {header.key}
                </button>
            {/each}
        </div>
    </div>

    <div class="headers-table">
        <div class="table-header">
            <div class="col-checkbox"></div>
            <div class="col-key">Key</div>
            <div class="col-value">Value</div>
            <div class="col-description">Description</div>
            <div class="col-actions"></div>
        </div>

        {#each headers as header (header.id)}
            <div class="table-row" class:disabled={!header.enabled}>
                <div class="col-checkbox">
                    <input
                            type="checkbox"
                            checked={header.enabled}
                            on:change={(e) => updateHeader(header.id, 'enabled', e.currentTarget.checked)}
                    />
                </div>
                <div class="col-key">
                    <input
                            type="text"
                            placeholder="Header name"
                            value={header.key}
                            on:input={(e) => updateHeader(header.id, 'key', e.currentTarget.value)}
                            class="header-input"
                            list="header-suggestions"
                    />
                </div>
                <div class="col-value">
                    <input
                            type="text"
                            placeholder="Header value"
                            value={header.value}
                            on:input={(e) => updateHeader(header.id, 'value', e.currentTarget.value)}
                            class="header-input"
                    />
                </div>
                <div class="col-description">
                    <input
                            type="text"
                            placeholder="Description (optional)"
                            value={header.description || ''}
                            on:input={(e) => updateHeader(header.id, 'description', e.currentTarget.value)}
                            class="header-input"
                    />
                </div>
                <div class="col-actions">
                    <button class="delete-btn" on:click={() => removeHeader(header.id)}>
                        <Trash2 size={16} />
                    </button>
                </div>
            </div>
        {/each}
    </div>
</div>

<datalist id="header-suggestions">
    {#each commonHeaders as header}
        <option value={header.key}>{header.value}</option>
    {/each}
</datalist>

<style>
    .headers-container {
        padding: 1rem;
        max-width: 100%;
    }

    .headers-header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        margin-bottom: 0.75rem;
    }

    .header-title {
        font-size: 0.875rem;
        font-weight: 600;
        color: #9ca3af;
        text-transform: uppercase;
        letter-spacing: 0.05em;
    }

    .header-subtitle {
        margin: 0;
        font-size: 0.75rem;
        color: #6b7280;
    }

    .add-button {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        padding: 0.375rem 0.75rem;
        background: transparent;
        border: 1px solid rgba(255, 255, 255, 0.1);
        border-radius: 4px;
        color: #9ca3af;
        font-size: 0.875rem;
        font-weight: 500;
        cursor: pointer;
        transition: all 0.2s;
    }

    .add-button:hover {
        background: rgba(255, 255, 255, 0.05);
        border-color: rgba(255, 255, 255, 0.2);
        color: #e4e4e7;
    }

    .common-headers {
        display: flex;
        align-items: center;
        gap: 0.75rem;
        margin-bottom: 0.75rem;
        padding: 0.5rem;
        background: #0a0a0a;
        border: 1px solid rgba(255, 255, 255, 0.08);
        border-radius: 4px;
    }

    .common-label {
        font-size: 0.75rem;
        font-weight: 600;
        color: #6b7280;
        text-transform: uppercase;
        letter-spacing: 0.05em;
        white-space: nowrap;
    }

    .common-buttons {
        display: flex;
        gap: 0.25rem;
        flex-wrap: wrap;
    }

    .common-btn {
        padding: 0.25rem 0.5rem;
        background: transparent;
        border: 1px solid rgba(255, 255, 255, 0.1);
        border-radius: 4px;
        color: #9ca3af;
        font-size: 0.75rem;
        font-weight: 500;
        cursor: pointer;
        transition: all 0.2s;
    }

    .common-btn:hover {
        background: rgba(255, 255, 255, 0.05);
        border-color: rgba(239, 68, 68, 0.3);
        color: #e4e4e7;
    }

    .headers-table {
        background: #0a0a0a;
        border: 1px solid rgba(255, 255, 255, 0.08);
        border-radius: 4px;
        overflow: hidden;
    }

    .table-header {
        display: grid;
        grid-template-columns: 32px 1fr 1fr 1.5fr 32px;
        gap: 0.25rem;
        padding: 0.5rem;
        background: #0a0a0a;
        border-bottom: 1px solid rgba(255, 255, 255, 0.08);
        font-size: 0.75rem;
        font-weight: 600;
        color: #6b7280;
        text-transform: uppercase;
        letter-spacing: 0.05em;
    }

    .table-row {
        display: grid;
        grid-template-columns: 32px 1fr 1fr 1.5fr 32px;
        gap: 0.25rem;
        padding: 0.5rem;
        border-bottom: 1px solid rgba(255, 255, 255, 0.05);
        transition: background 0.1s;
    }

    .table-row:hover {
        background: rgba(255, 255, 255, 0.02);
    }

    .table-row.disabled {
        opacity: 0.6;
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

    .header-input {
        width: 100%;
        background: transparent;
        border: none;
        color: #e4e4e7;
        font-size: 0.875rem;
        padding: 0.25rem 0.25rem;
        border-radius: 2px;
        transition: background 0.2s;
    }

    .header-input:focus {
        outline: none;
        background: rgba(255, 255, 255, 0.05);
    }

    .header-input::placeholder {
        color: #52525b;
    }

    .delete-btn {
        padding: 0.25rem;
        background: transparent;
        border: none;
        color: #6b7280;
        cursor: pointer;
        border-radius: 2px;
        transition: all 0.2s;
    }

    .delete-btn:hover {
        background: rgba(255, 255, 255, 0.05);
        color: #ef4444;
    }
</style>