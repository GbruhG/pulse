<script lang="ts">
    import { Settings, ChevronDown, Plus, Pencil } from 'lucide-svelte';
    import { workspaceStore } from '../stores/workspace';
    import { environmentStore } from '../stores/environment';
    import TabBar from './TabBar.svelte';
    import type { Workspace, Environment } from '../types';

    export let showEnvEditor = false;
    export let showSettings = false;

    let showWorkspaceMenu = false;
    let showEnvironmentMenu = false;
    let showNewWorkspaceModal = false;
    let showNewEnvironmentModal = false;
    let newWorkspaceName = '';
    let newEnvironmentName = '';

    $: activeWorkspace = $workspaceStore.workspaces.find(
        w => w.id === $workspaceStore.activeWorkspaceId
    );
    $: activeEnvironment = $environmentStore.environments.find(
        e => e.id === $environmentStore.activeEnvironmentId
    );

    function selectWorkspace(id: string) {
        workspaceStore.setActive(id);
        showWorkspaceMenu = false;
    }

    function selectEnvironment(id: string | null) {
        environmentStore.setActive(id);
        showEnvironmentMenu = false;
    }

    function createWorkspace() {
        if (!newWorkspaceName.trim()) return;

        const workspace: Workspace = {
            id: crypto.randomUUID(),
            name: newWorkspaceName,
            collections: [],
            environments: [],
            createdAt: new Date()
        };

        workspaceStore.addWorkspace(workspace);
        workspaceStore.setActive(workspace.id);
        newWorkspaceName = '';
        showNewWorkspaceModal = false;
    }

    function createEnvironment() {
        if (!newEnvironmentName.trim() || !activeWorkspace) return;

        const environment: Environment = {
            id: crypto.randomUUID(),
            name: newEnvironmentName,
            variables: {},
            workspaceId: activeWorkspace.id
        };

        environmentStore.addEnvironment(environment);
        newEnvironmentName = '';
        showNewEnvironmentModal = false;
    }

    function openEnvEditor() {
        showEnvironmentMenu = false;
        showEnvEditor = true;
    }
</script>

<div class="top-bar-container">
    <div class="top-bar">
        <div class="top-bar-left">
            <div class="logo-container">
                <svg width="32" height="32" viewBox="0 0 32 32" class="logo">
                    <path
                            d="M16 4C16 4 8 8 8 14C8 18 10 21 12 23L16 28L20 23C22 21 24 18 24 14C24 8 16 4 16 4Z"
                            fill="currentColor"
                            stroke="currentColor"
                            stroke-width="1.5"
                            stroke-linejoin="round"
                    />
                    <path
                            d="M10 14L14 16L16 12L18 16L22 14"
                            stroke="#0a0a0a"
                            stroke-width="2"
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            fill="none"
                    />
                </svg>
                <h1 class="app-title">PULSE</h1>
            </div>

            <div class="dropdown-container">
                <button class="dropdown-trigger" on:click={() => showWorkspaceMenu = !showWorkspaceMenu}>
                    <span>{activeWorkspace?.name || 'Select Workspace'}</span>
                    <ChevronDown size={16} />
                </button>

                {#if showWorkspaceMenu}
                    <div class="dropdown-menu">
                        <div class="dropdown-header">
                            <span>Workspaces</span>
                            <button class="add-btn" on:click={() => { showNewWorkspaceModal = true; showWorkspaceMenu = false; }}>
                                <Plus size={14} />
                            </button>
                        </div>
                        {#each $workspaceStore.workspaces as workspace}
                            <button
                                    class="dropdown-item"
                                    class:active={workspace.id === $workspaceStore.activeWorkspaceId}
                                    on:click={() => selectWorkspace(workspace.id)}
                            >
                                {workspace.name}
                            </button>
                        {/each}
                        {#if $workspaceStore.workspaces.length === 0}
                            <div class="dropdown-empty">No workspaces yet</div>
                        {/if}
                    </div>
                {/if}
            </div>

            <div class="dropdown-container">
                <button class="dropdown-trigger environment" on:click={() => showEnvironmentMenu = !showEnvironmentMenu}>
                    <div class="env-indicator" class:active={activeEnvironment} />
                    <span>{activeEnvironment?.name || 'No Environment'}</span>
                    <ChevronDown size={16} />
                </button>

                {#if showEnvironmentMenu}
                    <div class="dropdown-menu">
                        <div class="dropdown-header">
                            <span>Environments</span>
                            <div class="header-actions">
                                <button class="icon-btn-small" on:click={openEnvEditor} title="Manage variables">
                                    <Pencil size={12} />
                                </button>
                                <button class="add-btn" on:click={() => { showNewEnvironmentModal = true; showEnvironmentMenu = false; }}>
                                    <Plus size={14} />
                                </button>
                            </div>
                        </div>
                        <button
                                class="dropdown-item"
                                class:active={!$environmentStore.activeEnvironmentId}
                                on:click={() => selectEnvironment(null)}
                        >
                            <div class="env-indicator" />
                            No Environment
                        </button>
                        {#each $environmentStore.environments.filter(e => e.workspaceId === $workspaceStore.activeWorkspaceId) as environment}
                            <button
                                    class="dropdown-item"
                                    class:active={environment.id === $environmentStore.activeEnvironmentId}
                                    on:click={() => selectEnvironment(environment.id)}
                            >
                                <div class="env-indicator active" />
                                {environment.name}
                                <span class="var-count">{Object.keys(environment.variables).length}</span>
                            </button>
                        {/each}
                    </div>
                {/if}
            </div>
        </div>

        <div class="top-bar-right">
            <button class="icon-btn" on:click={() => showSettings = true} title="Settings">
                <Settings size={20} />
            </button>
        </div>
    </div>

    <!-- Tab Bar -->
    <TabBar />
</div>

{#if showNewWorkspaceModal}
    <div class="modal-overlay" on:click={() => showNewWorkspaceModal = false}>
        <div class="modal" on:click|stopPropagation>
            <h2>Create New Workspace</h2>
            <input
                    type="text"
                    bind:value={newWorkspaceName}
                    placeholder="Workspace name"
                    class="modal-input"
                    on:keydown={(e) => e.key === 'Enter' && createWorkspace()}
            />
            <div class="modal-actions">
                <button class="btn-secondary" on:click={() => showNewWorkspaceModal = false}>
                    Cancel
                </button>
                <button class="btn-primary" on:click={createWorkspace}>
                    Create
                </button>
            </div>
        </div>
    </div>
{/if}

{#if showNewEnvironmentModal}
    <div class="modal-overlay" on:click={() => showNewEnvironmentModal = false}>
        <div class="modal" on:click|stopPropagation>
            <h2>Create New Environment</h2>
            <input
                    type="text"
                    bind:value={newEnvironmentName}
                    placeholder="Environment name (e.g., Development, Production)"
                    class="modal-input"
                    on:keydown={(e) => e.key === 'Enter' && createEnvironment()}
            />
            <div class="modal-actions">
                <button class="btn-secondary" on:click={() => showNewEnvironmentModal = false}>
                    Cancel
                </button>
                <button class="btn-primary" on:click={createEnvironment}>
                    Create
                </button>
            </div>
        </div>
    </div>
{/if}

<style>
    .top-bar-container {
        background: #0a0a0a;
        border-bottom: 1px solid rgba(255, 255, 255, 0.08);
    }

    .top-bar {
        padding: 0.75rem 1.25rem;
        display: flex;
        align-items: center;
        justify-content: space-between;
    }

    .top-bar-left {
        display: flex;
        align-items: center;
        gap: 1rem;
    }

    .logo-container {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        margin-right: 0.5rem;
    }

    .logo {
        color: #ef4444;
    }

    .app-title {
        font-size: 1.4rem;
        font-weight: 700;
        margin: 0;
        letter-spacing: -0.02em;
        color: #ef4444;
    }

    .dropdown-container {
        position: relative;
    }

    .dropdown-trigger {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        background: #111111;
        border: 1px solid rgba(255, 255, 255, 0.1);
        border-radius: 4px;
        padding: 0.5rem 0.75rem;
        font-size: 0.875rem;
        color: #e4e4e7;
        cursor: pointer;
        transition: all 0.2s;
        min-width: 150px;
    }

    .dropdown-trigger:hover {
        background: #1a1a1a;
        border-color: rgba(255, 255, 255, 0.2);
    }

    .dropdown-trigger.environment {
        gap: 0.5rem;
    }

    .env-indicator {
        width: 8px;
        height: 8px;
        border-radius: 50%;
        background: #52525b;
    }

    .env-indicator.active {
        background: #10b981;
    }

    .dropdown-menu {
        position: absolute;
        top: calc(100% + 0.25rem);
        left: 0;
        background: #0f0f0f;
        border: 1px solid rgba(255, 255, 255, 0.1);
        border-radius: 6px;
        min-width: 200px;
        box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
        z-index: 100;
        overflow: hidden;
    }

    .dropdown-header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: 0.5rem 0.75rem;
        border-bottom: 1px solid rgba(255, 255, 255, 0.08);
        font-size: 0.75rem;
        font-weight: 600;
        color: #71717a;
        text-transform: uppercase;
        letter-spacing: 0.05em;
    }

    .header-actions {
        display: flex;
        gap: 0.25rem;
    }

    .icon-btn-small {
        display: flex;
        align-items: center;
        justify-content: center;
        width: 20px;
        height: 20px;
        padding: 0;
        background: transparent;
        border: none;
        color: #71717a;
        cursor: pointer;
        border-radius: 4px;
        transition: all 0.2s;
    }

    .icon-btn-small:hover {
        background: rgba(255, 255, 255, 0.1);
        color: #e4e4e7;
    }

    .add-btn {
        display: flex;
        align-items: center;
        justify-content: center;
        width: 20px;
        height: 20px;
        padding: 0;
        background: transparent;
        border: none;
        color: #71717a;
        cursor: pointer;
        border-radius: 4px;
        transition: all 0.2s;
    }

    .add-btn:hover {
        background: rgba(255, 255, 255, 0.1);
        color: #e4e4e7;
    }

    .dropdown-item {
        width: 100%;
        display: flex;
        align-items: center;
        gap: 0.5rem;
        padding: 0.5rem 0.75rem;
        background: transparent;
        border: none;
        color: #d4d4d4;
        cursor: pointer;
        transition: all 0.2s;
        text-align: left;
    }

    .dropdown-item:hover {
        background: rgba(255, 255, 255, 0.1);
    }

    .dropdown-item.active {
        background: rgba(239, 68, 68, 0.1);
        color: #f87171;
    }

    .var-count {
        margin-left: auto;
        padding: 0.125rem 0.25rem;
        background: rgba(255, 255, 255, 0.1);
        border-radius: 4px;
        font-size: 0.7rem;
        color: #71717a;
    }

    .dropdown-empty {
        padding: 0.75rem 0.75rem;
        text-align: center;
        color: #52525b;
        font-size: 0.875rem;
    }

    .top-bar-right {
        display: flex;
        gap: 0.5rem;
    }

    .icon-btn {
        padding: 0.5rem;
        background: transparent;
        border: 1px solid rgba(255, 255, 255, 0.1);
        color: #71717a;
        cursor: pointer;
        border-radius: 4px;
        transition: all 0.2s;
    }

    .icon-btn:hover {
        background: rgba(255, 255, 255, 0.1);
        border-color: rgba(239, 68, 68, 0.3);
        color: #e4e4e7;
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
        background: #0f0f0f;
        border: 1px solid rgba(255, 255, 255, 0.1);
        border-radius: 6px;
        padding: 1.5rem;
        width: 90%;
        max-width: 360px;
        box-shadow: 0 10px 30px rgba(0, 0, 0, 0.4);
    }

    .modal h2 {
        margin: 0 0 1rem 0;
        font-size: 1.2rem;
        font-weight: 600;
        color: #e4e4e7;
    }

    .modal-input {
        width: 100%;
        background: #0a0a0a;
        border: 1px solid rgba(255, 255, 255, 0.1);
        border-radius: 4px;
        padding: 0.5rem 0.75rem;
        color: white;
        font-size: 0.875rem;
        margin-bottom: 1rem;
        transition: all 0.2s;
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
        background: #dc2626;
        color: white;
    }

    .btn-primary:hover {
        background: #ef4444;
    }
</style>