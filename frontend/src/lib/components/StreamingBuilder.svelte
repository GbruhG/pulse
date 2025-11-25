<script lang="ts">
    import { Save } from 'lucide-svelte';
    import WebSocketHandler2 from './WebSocketHandler2.svelte';
    import SseHandler2 from './SSEHandler2.svelte';
    import GrpcStreamHandler2 from './GrpcStreamHandler2.svelte';
    import KafkaHandler2 from './KafkaHandler2.svelte';
    import MqttHandler2 from './MQTTHandler2.svelte';

    type StreamProtocol = 'websocket' | 'sse' | 'grpc-stream' | 'kafka' | 'mqtt';

    export let onSaveToCollection: () => void;

    let activeProtocol: StreamProtocol = 'websocket';

    const protocols = [
        { id: 'websocket' as const, label: 'WebSocket' },
        { id: 'sse' as const, label: 'SSE' },
        { id: 'grpc-stream' as const, label: 'gRPC Stream' },
        { id: 'kafka' as const, label: 'Kafka' },
        { id: 'mqtt' as const, label: 'MQTT' }
    ];
</script>

<div class="streaming-builder">
    <!-- Protocol Tabs -->
    <div class="protocol-tabs-container">
        <div class="protocol-tabs">
            {#each protocols as protocol}
                <button
                        class="protocol-tab"
                        class:active={activeProtocol === protocol.id}
                        on:click={() => activeProtocol = protocol.id}
                >
                    {protocol.label}
                </button>
            {/each}
        </div>

        <button class="save-btn" on:click={onSaveToCollection} title="Save to collection">
            <Save size={16} />
        </button>
    </div>

    <!-- Protocol-Specific Handler -->
    <div class="protocol-content">
        {#if activeProtocol === 'websocket'}
            <WebSocketHandler2 />
        {:else if activeProtocol === 'sse'}
            <SseHandler2 />
        {:else if activeProtocol === 'grpc-stream'}
            <GrpcStreamHandler2 />
        {:else if activeProtocol === 'kafka'}
            <KafkaHandler2 />
        {:else if activeProtocol === 'mqtt'}
            <MqttHandler2 />
        {/if}
    </div>
</div>

<style>
    .streaming-builder {
        display: flex;
        flex-direction: column;
        background: #0a0a0a;
        border-bottom: 1px solid rgba(255, 255, 255, 0.08);
        height: 100%; /* Add this */
    }

    .protocol-tabs-container {
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: 12px 16px 0 16px;
        gap: 12px;
    }

    .protocol-tabs {
        display: flex;
        gap: 2px;
        background: #0a0a0a;
        padding: 2px;
        border-radius: 4px;
        border: 1px solid rgba(255, 255, 255, 0.08);
        flex: 1;
    }

    .protocol-tab {
        padding: 0.5rem 1rem;
        background: transparent;
        border: none;
        color: #71717a;
        font-size: 0.875rem;
        font-weight: 500;
        cursor: pointer;
        border-radius: 2px;
        transition: all 0.2s;
        white-space: nowrap;
    }

    .protocol-tab:hover {
        color: #e4e4e7;
        background: rgba(255, 255, 255, 0.05);
    }

    .protocol-tab.active {
        color: #e4e4e7;
        background: rgba(255, 255, 255, 0.1);
    }

    .save-btn {
        display: flex;
        align-items: center;
        justify-content: center;
        padding: 0.5rem;
        background: transparent;
        border: 1px solid rgba(255, 255, 255, 0.1);
        border-radius: 4px;
        color: #71717a;
        cursor: pointer;
        transition: all 0.2s;
    }

    .save-btn:hover {
        background: rgba(255, 255, 255, 0.05);
        border-color: rgba(255, 255, 255, 0.2);
        color: #e4e4e7;
    }

    .protocol-content {
        flex: 1;
        overflow: hidden; /* Add this */
        min-height: 0; /* Add this - critical for flex children scrolling */
    }
</style>