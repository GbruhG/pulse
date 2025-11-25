<script lang="ts">
    export let activeProtocol: 'http' | 'websocket' | 'grpc' = 'http';

    const protocols = [
        { id: 'http' as const, label: 'HTTP', disabled: false },
        { id: 'streaming' as const, label: 'Streaming', disabled: false },
        { id: 'grpc' as const, label: 'gRPC', disabled: true }
    ];

    function selectProtocol(protocolId: typeof activeProtocol) {
        if (protocols.find(p => p.id === protocolId)?.disabled) return;
        activeProtocol = protocolId;
    }
</script>

<div class="protocol-tabs">
    {#each protocols as protocol}
        <button
                class="protocol-tab"
                class:active={activeProtocol === protocol.id}
                class:disabled={protocol.disabled}
                on:click={() => selectProtocol(protocol.id)}
                disabled={protocol.disabled}
        >
            {protocol.label}
            {#if protocol.disabled}
                <span class="coming-soon-badge">Soon</span>
            {/if}
        </button>
    {/each}
</div>

<style>
    .protocol-tabs {
        display: flex;
        gap: 0.25rem;
        padding: 1rem 1.5rem 0 1.5rem;
        background: #0a0a0a;
    }

    .protocol-tab {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        padding: 0.5rem 1rem;
        background: transparent;
        border: 1px solid #2a2a2a;
        border-radius: 0.5rem 0.5rem 0 0;
        color: #6b7280;
        font-size: 0.75rem;
        font-weight: 600;
        text-transform: uppercase;
        letter-spacing: 0.05em;
        cursor: pointer;
        transition: all 0.2s;
        position: relative;
    }

    .protocol-tab:hover:not(.disabled) {
        background: #1a1a1a;
        color: white;
    }

    .protocol-tab.active {
        background: #141414;
        border-bottom-color: #141414;
        color: #ef4444;
    }

    .protocol-tab.disabled {
        opacity: 0.5;
        cursor: not-allowed;
    }

    .coming-soon-badge {
        padding: 0.125rem 0.375rem;
        background: #2a2a2a;
        border-radius: 0.25rem;
        font-size: 0.625rem;
        color: #9ca3af;
    }
</style>