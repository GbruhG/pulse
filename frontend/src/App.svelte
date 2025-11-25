<script lang="ts">
    import { onMount } from 'svelte';
    import { SendRequest, SaveWorkspaces, LoadWorkspaces, SaveCollections, LoadCollections, SaveEnvironments, LoadEnvironments, SaveHistory, LoadHistory, SaveSettings, LoadSettings } from '../wailsjs/go/main/App';
    import TopBar from './lib/components/TopBar.svelte';
    import TabRenderer from './lib/components/TabRenderer.svelte';
    import HistoryPanel from './lib/components/HistoryPanel.svelte';
    import CollectionsPanel from './lib/components/CollectionsPanel.svelte';
    import SaveToCollectionModal from './lib/components/SaveToCollectionModal.svelte';
    import EnvironmentEditor from './lib/components/EnvironmentEditor.svelte';
    import SettingsModal from './lib/components/SettingsModal.svelte';
    import { workspaceStore } from './lib/stores/workspace';
    import { environmentStore } from './lib/stores/environment';
    import { collectionStore } from './lib/stores/collection';
    import { historyStore } from './lib/stores/history';
    import { settingsStore } from './lib/stores/settings';
    import { tabsStore, activeTab } from './lib/stores/tabs';
    import type { Workspace, CollectionRequest } from './lib/types';
    import { FolderOpen, Clock } from 'lucide-svelte';

    let activeSection = 'collections';
    let showSaveModal = false;
    let showEnvEditor = false;
    let showSettings = false;
    let layoutMode: 'horizontal' | 'vertical' = 'horizontal';

    let isResizingMain = false;
    let isResizingSidebar = false;
    let responseSize = 50;
    let sidebarWidth = 256;
    let minResponseSize = 300;

    let saveTimeout: number;
    const SAVE_DEBOUNCE_MS = 1000;
    let unsubscribers: Array<() => void> = [];

    onMount(async () => {
        window.addEventListener('loadRequest', (e: CustomEvent) => {
            loadRequestFromCollection(e.detail);
        });

        try {
            const settings = await LoadSettings();
            if (settings) {
                settingsStore.set(settings);
                layoutMode = settings.layoutMode || 'horizontal';
            }

            const [workspaces, collections, environments, history, savedTabs] = await Promise.all([
                LoadWorkspaces(),
                LoadCollections(),
                LoadEnvironments(),
                LoadHistory(),
                loadTabs()
            ]);

            if (workspaces && workspaces.length > 0) {
                workspaceStore.setWorkspaces(workspaces);
                workspaceStore.setActive(workspaces[0].id);
            } else {
                const defaultWorkspace: Workspace = {
                    id: crypto.randomUUID(),
                    name: 'My Workspace',
                    collections: [],
                    environments: [],
                    createdAt: new Date()
                };
                workspaceStore.addWorkspace(defaultWorkspace);
                workspaceStore.setActive(defaultWorkspace.id);
                await SaveWorkspaces([defaultWorkspace]);
            }

            if (collections) collectionStore.setCollections(collections);
            if (environments) environmentStore.setEnvironments(environments);
            if (history) {
                historyStore.clearHistory();
                history.forEach(item => historyStore.addItem(item));
            }

            // Load saved tabs or create default
            if (savedTabs && savedTabs.tabs && savedTabs.tabs.length > 0) {
                savedTabs.tabs.forEach(tab => {
                    const newTabId = tabsStore.createTab(tab.protocol);
                    tabsStore.updateTab(newTabId, {
                        name: tab.name,
                        httpRequest: tab.httpRequest,
                        streamingUrl: tab.streamingUrl,
                        streamingConfig: tab.streamingConfig
                    });
                });
                if (savedTabs.activeTabId) {
                    const tabExists = savedTabs.tabs.find(t => t.id === savedTabs.activeTabId);
                    if (tabExists) {
                        tabsStore.setActiveTab(savedTabs.activeTabId);
                    }
                }
            } else {
                // Create initial tab
                tabsStore.createTab('http');
            }

            setupStoreSubscriptions();
        } catch (error) {
            console.error('Failed to load data:', error);
            tabsStore.createTab('http');
        }
    });

    async function loadTabs() {
        try {
            const data = localStorage.getItem('pulse-tabs');
            return data ? JSON.parse(data) : null;
        } catch {
            return null;
        }
    }

    function setupStoreSubscriptions() {
        let isInitialLoad = true;
        setTimeout(() => { isInitialLoad = false; }, 100);

        unsubscribers.push(
            workspaceStore.subscribe($store => {
                if (!isInitialLoad && $store.workspaces.length > 0) {
                    scheduleSave(() => SaveWorkspaces($store.workspaces));
                }
            })
        );

        unsubscribers.push(
            collectionStore.subscribe($store => {
                if (!isInitialLoad && $store.length > 0) {
                    scheduleSave(() => SaveCollections($store));
                }
            })
        );

        unsubscribers.push(
            environmentStore.subscribe($store => {
                if (!isInitialLoad && $store.environments.length > 0) {
                    scheduleSave(() => SaveEnvironments($store.environments));
                }
            })
        );

        unsubscribers.push(
            historyStore.subscribe($store => {
                if (!isInitialLoad && $store.length > 0) {
                    scheduleSave(() => SaveHistory($store));
                }
            })
        );

        unsubscribers.push(
            settingsStore.subscribe($store => {
                if (!isInitialLoad && $store) {
                    scheduleSave(() => SaveSettings({
                        ...$store,
                        layoutMode
                    }));
                }
            })
        );

        // Save tabs on change
        unsubscribers.push(
            tabsStore.subscribe($store => {
                if (!isInitialLoad) {
                    saveTabs($store);
                }
            })
        );
    }

    function saveTabs(store: any) {
        try {
            localStorage.setItem('pulse-tabs', JSON.stringify({
                tabs: store.tabs.map(t => ({
                    id: t.id,
                    name: t.name,
                    protocol: t.protocol,
                    httpRequest: t.httpRequest,
                    streamingUrl: t.streamingUrl,
                    streamingConfig: t.streamingConfig
                })),
                activeTabId: store.activeTabId
            }));
        } catch (e) {
            console.error('Failed to save tabs:', e);
        }
    }

    function scheduleSave(saveFn: () => Promise<void>) {
        clearTimeout(saveTimeout);
        saveTimeout = setTimeout(() => {
            saveFn().catch(console.error);
        }, SAVE_DEBOUNCE_MS);
    }

    function loadRequestFromCollection(collectionRequest: CollectionRequest) {
        const method = collectionRequest.request.method.toUpperCase();

        if (method === 'WSS' || method === 'SSE' || method === 'GRPC' || method === 'KAFKA' || method === 'MQTT') {
            const protocolMap = {
                'WSS': 'websocket',
                'SSE': 'sse',
                'GRPC': 'grpc-stream',
                'KAFKA': 'kafka',
                'MQTT': 'mqtt'
            };

            const protocol = protocolMap[method] || 'websocket';
            const newTabId = tabsStore.createTab(protocol as any);

            tabsStore.updateTab(newTabId, {
                name: collectionRequest.name,
                streamingUrl: collectionRequest.request.url,
                streamingConfig: collectionRequest.request.streamingConfig?.config || {}
            });
        } else {
            const newTabId = tabsStore.createTab('http');
            tabsStore.updateTab(newTabId, {
                name: collectionRequest.name,
                httpRequest: collectionRequest.request
            });
        }
    }

    function handleSaveToCollection() {
        showSaveModal = true;
    }

    function startResizeMain(e: MouseEvent) {
        isResizingMain = true;
        e.preventDefault();
    }

    function startResizeSidebar(e: MouseEvent) {
        isResizingSidebar = true;
        e.preventDefault();
    }

    function handleMouseMove(e: MouseEvent) {
        if (isResizingMain) {
            const container = document.querySelector('.work-area');
            if (!container) return;

            const rect = container.getBoundingClientRect();

            if (layoutMode === 'horizontal') {
                const availableWidth = rect.width;
                const responseWidth = availableWidth - (e.clientX - rect.left);
                const percentage = (responseWidth / availableWidth) * 100;
                const minPercentage = (minResponseSize / availableWidth) * 100;
                responseSize = Math.max(minPercentage, Math.min(80, percentage));
            } else {
                const availableHeight = rect.height;
                const responseHeight = availableHeight - (e.clientY - rect.top);
                const percentage = (responseHeight / availableHeight) * 100;
                const minPercentage = (minResponseSize / availableHeight) * 100;
                responseSize = Math.max(minPercentage, Math.min(80, percentage));
            }
        }

        if (isResizingSidebar) {
            sidebarWidth = Math.max(200, Math.min(500, e.clientX));
        }
    }

    function stopResize() {
        isResizingMain = false;
        isResizingSidebar = false;
    }

    function handleLayoutChange(newMode: 'horizontal' | 'vertical') {
        layoutMode = newMode;
        responseSize = newMode === 'horizontal' ? 50 : 45;
    }
</script>

<svelte:window on:mousemove={handleMouseMove} on:mouseup={stopResize} />

<div class="app-container" style="font-size: {$settingsStore.uiScale}%">
    <TopBar bind:showEnvEditor={showEnvEditor} bind:showSettings={showSettings} />

    <div class="main-content">
        <div class="sidebar" style="width: {sidebarWidth}px">
            <div class="sidebar-nav">
                <button
                        class="nav-btn"
                        class:active={activeSection === 'collections'}
                        on:click={() => activeSection = 'collections'}
                >
                    <FolderOpen size={18} />
                    <span>Collections</span>
                </button>
                <button
                        class="nav-btn"
                        class:active={activeSection === 'history'}
                        on:click={() => activeSection = 'history'}
                >
                    <Clock size={18} />
                    <span>History</span>
                </button>
            </div>

            <div class="sidebar-content">
                {#if activeSection === 'collections'}
                    <CollectionsPanel />
                {:else if activeSection === 'history'}
                    <HistoryPanel />
                {/if}
            </div>
        </div>

        <div class="sidebar-resize-handle" on:mousedown={startResizeSidebar}></div>

        {#if $activeTab}
            <TabRenderer
                    tab={$activeTab}
                    {layoutMode}
                    {responseSize}
                    on:layoutChange={(e) => handleLayoutChange(e.detail)}
                    on:resizeStart={startResizeMain}
                    on:saveToCollection={handleSaveToCollection}
            />
        {:else}
            <div class="no-tabs">
                <p>No active tab</p>
                <span>Create a new tab to get started</span>
            </div>
        {/if}
    </div>
</div>

<SaveToCollectionModal
        bind:show={showSaveModal}
        isStreamingMode={$activeTab?.protocol !== 'http'}
        on:saved={() => showSaveModal = false}
/>
<EnvironmentEditor bind:show={showEnvEditor} />
<SettingsModal bind:show={showSettings} />

<style>
    :global(body) {
        margin: 0;
        padding: 0;
        font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen,
        Ubuntu, Cantarell, sans-serif;
        overflow: hidden;
        user-select: none;
        background: #0a0a0a;
    }

    .app-container {
        height: 100vh;
        background: #0a0a0a;
        color: white;
        display: flex;
        flex-direction: column;
        position: relative;
        overflow: hidden;
    }

    .main-content {
        display: flex;
        flex: 1;
        overflow: hidden;
        position: relative;
    }

    .sidebar {
        background: #0a0a0a;
        border-right: 1px solid rgba(255, 255, 255, 0.08);
        display: flex;
        flex-direction: column;
        width: 220px;
        transition: width 0.2s;
    }

    .sidebar-resize-handle {
        width: 4px;
        background: transparent;
        cursor: col-resize;
        transition: background 0.2s;
        position: relative;
    }

    .sidebar-resize-handle:hover,
    .sidebar-resize-handle:active {
        background: #ef4444;
    }

    .sidebar-nav {
        padding: 1rem;
        display: flex;
        flex-direction: column;
        gap: 0.25rem;
        border-bottom: 1px solid rgba(255, 255, 255, 0.08);
    }

    .nav-btn {
        display: flex;
        align-items: center;
        gap: 0.75rem;
        padding: 0.5rem 0.75rem;
        background: transparent;
        border: none;
        color: #9ca3af;
        cursor: pointer;
        border-radius: 4px;
        font-weight: 500;
        transition: all 0.2s;
        text-align: left;
        font-size: 0.875rem;
    }

    .nav-btn:hover {
        background: rgba(255, 255, 255, 0.05);
        color: #e4e4e7;
    }

    .nav-btn.active {
        background: #1a1a1a;
        color: #e4e4e7;
        border-left: 2px solid #ef4444;
    }

    .sidebar-content {
        flex: 1;
        overflow: hidden;
    }

    .no-tabs {
        flex: 1;
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        gap: 0.75rem;
        color: #71717a;
        background: #0a0a0a;
    }

    .no-tabs p {
        margin: 0;
        font-size: 1rem;
        font-weight: 500;
        color: #9ca3af;
    }

    .no-tabs span {
        font-size: 0.875rem;
        color: #52525b;
        text-align: center;
        max-width: 200px;
        line-height: 1.4;
    }

    :global(*::-webkit-scrollbar) {
        width: 6px;
        height: 6px;
    }

    :global(*::-webkit-scrollbar-track) {
        background: #0a0a0a;
    }

    :global(*::-webkit-scrollbar-thumb) {
        background: #27272a;
        border-radius: 3px;
    }

    :global(*::-webkit-scrollbar-thumb:hover) {
        background: #3f3f46;
    }

    :global(*::-webkit-scrollbar-corner) {
        background: transparent;
    }
</style>