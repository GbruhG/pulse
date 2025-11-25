<script lang="ts">
    import { ChevronRight, ChevronDown } from 'lucide-svelte';
    import { writable } from 'svelte/store';

    export let data: any;
    export let level: number = 0;
    export let path: string = 'root';
    export let expandedStore: any;

    // Create store on first mount if not provided
    if (!expandedStore) {
        expandedStore = writable(new Set<string>());
    }

    let isExpanded = false;

    // Initialize expansion state based on level
    $: {
        if (level < 2 && !$expandedStore.has(path)) {
            $expandedStore.add(path);
            expandedStore.set($expandedStore);
        }
        isExpanded = $expandedStore.has(path);
    }

    function toggleNode() {
        if ($expandedStore.has(path)) {
            $expandedStore.delete(path);
        } else {
            $expandedStore.add(path);
        }
        expandedStore.set($expandedStore);
    }

    function getValueType(value: any): string {
        if (value === null) return 'null';
        if (Array.isArray(value)) return 'array';
        return typeof value;
    }

    function renderPrimitiveValue(value: any, key: string) {
        const type = getValueType(value);
        const displayValue = type === 'string' ? `"${value}"` : String(value);
        return { key, value: displayValue, type };
    }
</script>

{#if typeof data === 'object' && data !== null}
    {#if Array.isArray(data)}
        {#each data as item, index}
            {@const itemPath = `${path}.${index}`}
            {@const type = getValueType(item)}
            {#if type === 'object' || type === 'array'}
                {@const isArray = type === 'array'}
                {@const isEmpty = isArray ? item.length === 0 : Object.keys(item).length === 0}
                {@const nodeExpanded = $expandedStore.has(itemPath)}
                <div class="json-node">
                    <div
                            class="json-line"
                            on:click={() => !isEmpty && (() => {
                            if ($expandedStore.has(itemPath)) {
                                $expandedStore.delete(itemPath);
                            } else {
                                $expandedStore.add(itemPath);
                            }
                            expandedStore.set($expandedStore);
                        })()}
                            on:keydown={(e) => e.key === 'Enter' && !isEmpty && (() => {
                            if ($expandedStore.has(itemPath)) {
                                $expandedStore.delete(itemPath);
                            } else {
                                $expandedStore.add(itemPath);
                            }
                            expandedStore.set($expandedStore);
                        })()}
                            role="button"
                            tabindex="0"
                            style="cursor: {isEmpty ? 'default' : 'pointer'}"
                    >
                        {#if !isEmpty}
                            <span class="toggle-icon">
                                {#if nodeExpanded}
                                    <ChevronDown size={14} />
                                {:else}
                                    <ChevronRight size={14} />
                                {/if}
                            </span>
                        {:else}
                            <span class="toggle-icon empty"></span>
                        {/if}
                        <span class="key">{index}:</span>
                        <span class="bracket">{isArray ? '[' : '{'}</span>
                        {#if !nodeExpanded && !isEmpty}
                            <span class="preview">{isArray ? `Array(${item.length})` : 'Object'}</span>
                        {/if}
                        {#if !nodeExpanded}
                            <span class="bracket">{isArray ? ']' : '}'}</span>
                        {/if}
                    </div>
                    {#if nodeExpanded && !isEmpty}
                        <div class="json-children">
                            <svelte:self data={item} level={level + 1} path={itemPath} {expandedStore} />
                            <div class="json-line">
                                <span class="bracket">{isArray ? ']' : '}'}</span>
                            </div>
                        </div>
                    {/if}
                </div>
            {:else}
                {@const primitive = renderPrimitiveValue(item, index.toString())}
                <div class="json-line">
                    <span class="toggle-icon empty"></span>
                    <span class="key">{primitive.key}:</span>
                    <span class="value {primitive.type}">
                        {primitive.value}
                    </span>
                </div>
            {/if}
        {/each}
    {:else}
        {#each Object.entries(data) as [key, value]}
            {@const itemPath = `${path}.${key}`}
            {@const type = getValueType(value)}
            {#if type === 'object' || type === 'array'}
                {@const isArray = type === 'array'}
                {@const isEmpty = isArray ? value.length === 0 : Object.keys(value).length === 0}
                {@const nodeExpanded = $expandedStore.has(itemPath)}
                <div class="json-node">
                    <div
                            class="json-line"
                            on:click={() => !isEmpty && (() => {
                            if ($expandedStore.has(itemPath)) {
                                $expandedStore.delete(itemPath);
                            } else {
                                $expandedStore.add(itemPath);
                            }
                            expandedStore.set($expandedStore);
                        })()}
                            on:keydown={(e) => e.key === 'Enter' && !isEmpty && (() => {
                            if ($expandedStore.has(itemPath)) {
                                $expandedStore.delete(itemPath);
                            } else {
                                $expandedStore.add(itemPath);
                            }
                            expandedStore.set($expandedStore);
                        })()}
                            role="button"
                            tabindex="0"
                            style="cursor: {isEmpty ? 'default' : 'pointer'}"
                    >
                        {#if !isEmpty}
                            <span class="toggle-icon">
                                {#if nodeExpanded}
                                    <ChevronDown size={14} />
                                {:else}
                                    <ChevronRight size={14} />
                                {/if}
                            </span>
                        {:else}
                            <span class="toggle-icon empty"></span>
                        {/if}
                        <span class="key">{key}:</span>
                        <span class="bracket">{isArray ? '[' : '{'}</span>
                        {#if !nodeExpanded && !isEmpty}
                            <span class="preview">{isArray ? `Array(${value.length})` : 'Object'}</span>
                        {/if}
                        {#if !nodeExpanded}
                            <span class="bracket">{isArray ? ']' : '}'}</span>
                        {/if}
                    </div>
                    {#if nodeExpanded && !isEmpty}
                        <div class="json-children">
                            <svelte:self data={value} level={level + 1} path={itemPath} {expandedStore} />
                            <div class="json-line">
                                <span class="bracket">{isArray ? ']' : '}'}</span>
                            </div>
                        </div>
                    {/if}
                </div>
            {:else}
                {@const primitive = renderPrimitiveValue(value, key)}
                <div class="json-line">
                    <span class="toggle-icon empty"></span>
                    <span class="key">{primitive.key}:</span>
                    <span class="value {primitive.type}">
                        {primitive.value}
                    </span>
                </div>
            {/if}
        {/each}
    {/if}
{/if}

<style>
    .json-node {
        margin-left: 0;
    }

    .json-children {
        margin-left: 20px;
        border-left: 1px solid #222222;
        padding-left: 12px;
    }

    .json-line {
        display: flex;
        align-items: center;
        gap: 6px;
        padding: 2px 0;
        transition: background 0.15s;
    }

    .json-line:hover {
        background: #111111;
        margin-left: -8px;
        padding-left: 8px;
        margin-right: -8px;
        padding-right: 8px;
        border-radius: 4px;
    }

    .toggle-icon {
        width: 16px;
        height: 16px;
        display: flex;
        align-items: center;
        justify-content: center;
        color: #52525b;
        flex-shrink: 0;
    }

    .toggle-icon.empty {
        width: 16px;
    }

    .toggle-icon:hover {
        color: #a1a1aa;
    }

    .key {
        color: #60a5fa;
        font-weight: 600;
        margin-right: 4px;
    }

    .bracket {
        color: #71717a;
        font-weight: 600;
    }

    .preview {
        color: #52525b;
        font-style: italic;
        margin: 0 4px;
    }

    .value {
        font-weight: 500;
    }

    .value.string {
        color: #34d399;
    }

    .value.number {
        color: #fbbf24;
    }

    .value.boolean {
        color: #c084fc;
        font-weight: 700;
    }

    .value.null {
        color: #f87171;
        font-weight: 700;
    }
</style>