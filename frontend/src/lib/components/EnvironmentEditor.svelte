<script lang="ts">
    import { Plus, Trash2, X } from 'lucide-svelte';
    import { environmentStore } from '../stores/environment';

    export let show = false;

    let selectedEnvId: string | null = null;
    let variables: { key: string; value: string }[] = [];

    $: selectedEnv = $environmentStore.environments.find(e => e.id === selectedEnvId);

    $: if (show && !selectedEnvId && $environmentStore.environments.length > 0) {
        selectedEnvId = $environmentStore.environments[0].id;
        loadVariables();
    }

    function loadVariables() {
        if (!selectedEnv) {
            variables = [];
            return;
        }
        variables = Object.entries(selectedEnv.variables).map(([key, value]) => ({ key, value }));
        if (variables.length === 0) {
            addVariable();
        }
    }

    function addVariable() {
        variables = [...variables, { key: '', value: '' }];
    }

    function removeVariable(index: number) {
        variables = variables.filter((_, i) => i !== index);
        if (variables.length === 0) {
            addVariable();
        }
    }

    function saveVariables() {
        if (!selectedEnvId) return;

        const variablesObj: Record<string, string> = {};
        variables.filter(v => v.key.trim()).forEach(v => {
            variablesObj[v.key.trim()] = v.value;
        });

        // Update the environment in the store
        const envs = $environmentStore.environments.map(e =>
            e.id === selectedEnvId ? { ...e, variables: variablesObj } : e
        );

        environmentStore.setEnvironments(envs);

        close();
    }

    function close() {
        show = false;
        variables = [];
        selectedEnvId = null;
    }

    function selectEnvironment(id: string) {
        selectedEnvId = id;
        loadVariables();
    }
</script>

{#if show}
    <div class="modal-overlay" on:click={close}>
        <div class="modal large" on:click|stopPropagation>
            <div class="modal-header">
                <h2>Manage Environments</h2>
                <button class="close-btn" on:click={close}>
                    <X size={20} />
                </button>
            </div>

            <div class="modal-content">
                <div class="sidebar">
                    <div class="sidebar-header">
                        <span>Environments</span>
                    </div>
                    <div class="env-list">
                        {#each $environmentStore.environments as env}
                            <button
                                    class="env-item"
                                    class:active={selectedEnvId === env.id}
                                    on:click={() => selectEnvironment(env.id)}
                            >
                                <div class="env-indicator" class:active={env.id === $environmentStore.activeEnvironmentId} />
                                <span class="env-name">{env.name}</span>
                                <span class="var-count">{Object.keys(env.variables).length}</span>
                            </button>
                        {/each}
                        {#if $environmentStore.environments.length === 0}
                            <div class="no-envs">
                                <p>No environments yet</p>
                                <span>Create one from the top bar</span>
                            </div>
                        {/if}
                    </div>
                </div>

                <div class="editor">
                    {#if selectedEnv}
                        <div class="editor-header">
                            <h3>{selectedEnv.name}</h3>
                            <p>Define variables that can be used in requests with <code>{`{{variable_name}}`}</code></p>
                        </div>

                        <div class="variables-section">
                            <div class="section-header">
                                <span>Variables</span>
                                <button class="add-var-btn" on:click={addVariable}>
                                    <Plus size={16} />
                                    Add Variable
                                </button>
                            </div>

                            <div class="variables-table">
                                <div class="table-header">
                                    <div class="col-key">Variable Name</div>
                                    <div class="col-value">Value</div>
                                    <div class="col-actions"></div>
                                </div>

                                {#each variables as variable, i (i)}
                                    <div class="table-row">
                                        <input
                                                type="text"
                                                bind:value={variable.key}
                                                placeholder="e.g., base_url, auth_token, api_key"
                                                class="var-input"
                                        />
                                        <input
                                                type="text"
                                                bind:value={variable.value}
                                                placeholder="e.g., https://api.example.com"
                                                class="var-input"
                                        />
                                        <button class="delete-btn" on:click={() => removeVariable(i)}>
                                            <Trash2 size={16} />
                                        </button>
                                    </div>
                                {/each}
                            </div>

                            <div class="usage-hint">
                                <div class="hint-row">
                                    <strong>Usage:</strong> Use variables in your requests with <code>{`{{variable_name}}`}</code>
                                </div>
                                <div class="hint-row">
                                    <strong>Example:</strong> <code>{`{{base_url}}`}/users</code> → <code>https://api.example.com/users</code>
                                </div>
                                <div class="hint-row">
                                    <strong>Note:</strong> Variables work in URL, params, headers, body, and auth fields
                                </div>
                            </div>
                        </div>
                    {:else}
                        <div class="no-selection">
                            <p>Select an environment to edit variables</p>
                            <span>or create a new environment from the top bar</span>
                        </div>
                    {/if}
                </div>
            </div>

            <div class="modal-actions">
                <button class="btn-secondary" on:click={close}>
                    Cancel
                </button>
                <button class="btn-primary" on:click={saveVariables} disabled={!selectedEnvId}>
                    Save Changes
                </button>
            </div>
        </div>
    </div>
{/if}

<style>
    .modal-overlay {
        position: fixed;
        inset: 0;
        background: rgba(0, 0, 0, 0.8);
        display: flex;
        align-items: center;
        justify-content: center;
        z-index: 1000;
        backdrop-filter: blur(4px);
    }

    .modal {
        background: #1a1a1a;
        border: 1px solid #2a2a2a;
        border-radius: 0.75rem;
        width: 90%;
        max-width: 950px;
        max-height: 85vh;
        display: flex;
        flex-direction: column;
        box-shadow: 0 20px 60px rgba(0, 0, 0, 0.8);
    }

    .modal-header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: 1.5rem;
        border-bottom: 1px solid #2a2a2a;
    }

    .modal-header h2 {
        margin: 0;
        font-size: 1.25rem;
        font-weight: 600;
    }

    .close-btn {
        padding: 0.5rem;
        background: transparent;
        border: none;
        color: #9ca3af;
        cursor: pointer;
        border-radius: 0.375rem;
        transition: all 0.2s;
    }

    .close-btn:hover {
        background: #2a2a2a;
        color: white;
    }

    .modal-content {
        flex: 1;
        display: flex;
        overflow: hidden;
    }

    .sidebar {
        width: 240px;
        background: #0f0f0f;
        border-right: 1px solid #2a2a2a;
        display: flex;
        flex-direction: column;
    }

    .sidebar-header {
        padding: 1rem 1.25rem;
        font-size: 0.75rem;
        font-weight: 600;
        color: #6b7280;
        text-transform: uppercase;
        letter-spacing: 0.05em;
        border-bottom: 1px solid #2a2a2a;
    }

    .env-list {
        flex: 1;
        overflow-y: auto;
        padding: 0.5rem;
    }

    .env-item {
        width: 100%;
        display: flex;
        align-items: center;
        gap: 0.75rem;
        padding: 0.875rem 1rem;
        background: transparent;
        border: none;
        border-radius: 0.5rem;
        color: #d1d5db;
        text-align: left;
        cursor: pointer;
        transition: all 0.2s;
        font-size: 0.875rem;
        margin-bottom: 0.25rem;
    }

    .env-item:hover {
        background: #1a1a1a;
    }

    .env-item.active {
        background: #1a1a1a;
        color: white;
    }

    .env-indicator {
        width: 8px;
        height: 8px;
        border-radius: 50%;
        background: #4b5563;
        flex-shrink: 0;
    }

    .env-indicator.active {
        background: #10b981;
        box-shadow: 0 0 8px rgba(16, 185, 129, 0.4);
    }

    .env-name {
        flex: 1;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
    }

    .var-count {
        padding: 0.125rem 0.5rem;
        background: #2a2a2a;
        border-radius: 0.25rem;
        font-size: 0.7rem;
        color: #9ca3af;
        font-weight: 600;
    }

    .no-envs {
        padding: 2rem 1rem;
        text-align: center;
    }

    .no-envs p {
        margin: 0 0 0.25rem 0;
        font-size: 0.875rem;
        color: #9ca3af;
    }

    .no-envs span {
        font-size: 0.75rem;
        color: #6b7280;
    }

    .editor {
        flex: 1;
        display: flex;
        flex-direction: column;
        overflow-y: auto;
    }

    .editor-header {
        padding: 1.5rem;
        border-bottom: 1px solid #2a2a2a;
    }

    .editor-header h3 {
        margin: 0 0 0.5rem 0;
        font-size: 1.125rem;
        font-weight: 600;
    }

    .editor-header p {
        margin: 0;
        font-size: 0.875rem;
        color: #9ca3af;
        line-height: 1.5;
    }

    .editor-header code {
        padding: 0.125rem 0.375rem;
        background: #0f0f0f;
        border-radius: 0.25rem;
        font-family: 'Monaco', 'Menlo', monospace;
        color: #10b981;
    }

    .variables-section {
        flex: 1;
        padding: 1.5rem;
    }

    .section-header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        margin-bottom: 1rem;
    }

    .section-header span {
        font-size: 0.875rem;
        font-weight: 600;
        color: #d1d5db;
    }

    .add-var-btn {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        padding: 0.5rem 1rem;
        background: #2a2a2a;
        border: none;
        border-radius: 0.5rem;
        color: white;
        font-size: 0.875rem;
        font-weight: 500;
        cursor: pointer;
        transition: all 0.2s;
    }

    .add-var-btn:hover {
        background: #3a3a3a;
    }

    .variables-table {
        background: #0f0f0f;
        border: 1px solid #2a2a2a;
        border-radius: 0.5rem;
        overflow: hidden;
        margin-bottom: 1.5rem;
    }

    .table-header {
        display: grid;
        grid-template-columns: 1fr 1fr 48px;
        gap: 0.5rem;
        padding: 0.875rem 1rem;
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
        grid-template-columns: 1fr 1fr 48px;
        gap: 0.5rem;
        padding: 0.875rem 1rem;
        border-bottom: 1px solid #1a1a1a;
        align-items: center;
    }

    .table-row:last-child {
        border-bottom: none;
    }

    .var-input {
        background: transparent;
        border: 1px solid transparent;
        color: white;
        font-size: 0.875rem;
        padding: 0.5rem 0.75rem;
        border-radius: 0.375rem;
        transition: all 0.2s;
        font-family: 'Monaco', 'Menlo', monospace;
    }

    .var-input:focus {
        outline: none;
        background: #1a1a1a;
        border-color: #ef4444;
    }

    .var-input:hover {
        background: #1a1a1a;
    }

    .var-input::placeholder {
        color: #4b5563;
    }

    .delete-btn {
        padding: 0.5rem;
        background: transparent;
        border: none;
        color: #6b7280;
        cursor: pointer;
        border-radius: 0.375rem;
        transition: all 0.2s;
    }

    .delete-btn:hover {
        background: #1a1a1a;
        color: #ef4444;
    }

    .usage-hint {
        padding: 1.25rem;
        background: #0f0f0f;
        border: 1px solid #2a2a2a;
        border-radius: 0.5rem;
        font-size: 0.875rem;
        color: #9ca3af;
        line-height: 1.8;
    }

    .hint-row {
        margin-bottom: 0.75rem;
    }

    .hint-row:last-child {
        margin-bottom: 0;
    }

    .usage-hint code {
        padding: 0.125rem 0.375rem;
        background: #1a1a1a;
        border-radius: 0.25rem;
        font-family: 'Monaco', 'Menlo', monospace;
        color: #10b981;
        font-size: 0.85rem;
    }

    .no-selection {
        flex: 1;
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        padding: 3rem;
        gap: 0.5rem;
    }

    .no-selection p {
        margin: 0;
        font-size: 0.875rem;
        color: #d1d5db;
    }

    .no-selection span {
        font-size: 0.75rem;
        color: #6b7280;
    }

    .modal-actions {
        display: flex;
        gap: 0.75rem;
        justify-content: flex-end;
        padding: 1rem 1.5rem;
        border-top: 1px solid #2a2a2a;
    }

    .btn-secondary,
    .btn-primary {
        padding: 0.625rem 1.25rem;
        border-radius: 0.5rem;
        font-weight: 500;
        cursor: pointer;
        transition: all 0.2s;
        border: none;
    }

    .btn-secondary {
        background: #2a2a2a;
        color: white;
    }

    .btn-secondary:hover {
        background: #3a3a3a;
    }

    .btn-primary {
        background: linear-gradient(to right, #dc2626, #ef4444);
        color: white;
    }

    .btn-primary:hover:not(:disabled) {
        background: linear-gradient(to right, #ef4444, #f87171);
    }

    .btn-primary:disabled {
        opacity: 0.5;
        cursor: not-allowed;
    }
</style>