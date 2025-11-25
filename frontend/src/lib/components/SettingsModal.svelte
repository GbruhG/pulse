<script lang="ts">
    import { X, RotateCcw } from 'lucide-svelte';
    import { settingsStore } from '../stores/settings';

    export let show = false;

    let localSettings = { ...$settingsStore };

    $: if (show) {
        localSettings = { ...$settingsStore };
    }

    function saveSettings() {
        settingsStore.set(localSettings);
        show = false;
    }

    function resetToDefaults() {
        if (confirm('Reset all settings to defaults?')) {
            settingsStore.reset();
            localSettings = { ...$settingsStore };
        }
    }

    function close() {
        show = false;
    }

    function handleScaleChange(e: Event) {
        const target = e.target as HTMLInputElement;
        localSettings.uiScale = parseInt(target.value);
    }

    function handleTimeoutChange(e: Event) {
        const target = e.target as HTMLInputElement;
        localSettings.defaultTimeout = parseInt(target.value);
    }

    function handleMaxHistoryChange(e: Event) {
        const target = e.target as HTMLInputElement;
        localSettings.maxHistoryItems = parseInt(target.value);
    }
</script>

{#if show}
    <div class="modal-overlay" on:click={close}>
        <div class="modal" on:click|stopPropagation>
            <div class="modal-header">
                <h2>Settings</h2>
                <button class="close-btn" on:click={close}>
                    <X size={20} />
                </button>
            </div>

            <div class="modal-body">
                <div class="settings-section">
                    <h3>Appearance</h3>

                    <div class="setting-row">
                        <div class="setting-info">
                            <label>UI Scale</label>
                            <span class="setting-desc">Adjust the size of all interface elements</span>
                        </div>
                        <div class="setting-control">
                            <input
                                    type="range"
                                    min="75"
                                    max="125"
                                    step="5"
                                    value={localSettings.uiScale}
                                    on:input={handleScaleChange}
                                    class="slider"
                            />
                            <span class="value-label">{localSettings.uiScale}%</span>
                        </div>
                    </div>

                    <div class="setting-row">
                        <div class="setting-info">
                            <label>Default View</label>
                            <span class="setting-desc">Default JSON response view</span>
                        </div>
                        <div class="setting-control">
                            <label class="checkbox-label">
                                <input
                                        type="checkbox"
                                        checked={localSettings.prettyPrintByDefault}
                                        on:change={(e) => localSettings.prettyPrintByDefault = e.currentTarget.checked}
                                />
                                <span>Pretty print by default</span>
                            </label>
                        </div>
                    </div>
                </div>

                <div class="settings-section">
                    <h3>Behavior</h3>

                    <div class="setting-row">
                        <div class="setting-info">
                            <label>Request Timeout</label>
                            <span class="setting-desc">Maximum time to wait for response</span>
                        </div>
                        <div class="setting-control">
                            <input
                                    type="number"
                                    min="5"
                                    max="300"
                                    value={localSettings.defaultTimeout}
                                    on:input={handleTimeoutChange}
                                    class="number-input"
                            />
                            <span class="unit-label">seconds</span>
                        </div>
                    </div>

                    <div class="setting-row">
                        <div class="setting-info">
                            <label>History</label>
                            <span class="setting-desc">Automatically save request history</span>
                        </div>
                        <div class="setting-control">
                            <label class="checkbox-label">
                                <input
                                        type="checkbox"
                                        checked={localSettings.autoSaveHistory}
                                        on:change={(e) => localSettings.autoSaveHistory = e.currentTarget.checked}
                                />
                                <span>Auto-save history</span>
                            </label>
                        </div>
                    </div>

                    <div class="setting-row">
                        <div class="setting-info">
                            <label>Max History Items</label>
                            <span class="setting-desc">Maximum number of history items to keep</span>
                        </div>
                        <div class="setting-control">
                            <input
                                    type="number"
                                    min="10"
                                    max="500"
                                    value={localSettings.maxHistoryItems}
                                    on:input={handleMaxHistoryChange}
                                    class="number-input"
                            />
                        </div>
                    </div>
                </div>

                <div class="info-box">
                    <strong>Note:</strong> UI scale changes take effect immediately. Some changes may require app restart for full effect.
                </div>
            </div>

            <div class="modal-actions">
                <button class="btn-reset" on:click={resetToDefaults}>
                    <RotateCcw size={16} />
                    Reset to Defaults
                </button>
                <div class="action-buttons">
                    <button class="btn-secondary" on:click={close}>
                        Cancel
                    </button>
                    <button class="btn-primary" on:click={saveSettings}>
                        Save Settings
                    </button>
                </div>
            </div>
        </div>
    </div>
{/if}

<style>
    .modal-overlay {
        position: fixed;
        inset: 0;
        background: rgba(0, 0, 0, 0.7);
        display: flex;
        align-items: center;
        justify-content: center;
        z-index: 2000;
    }

    .modal {
        background: #0a0a0a;
        border: 1px solid rgba(255, 255, 255, 0.08);
        border-radius: 6px;
        width: 90%;
        max-width: 600px;
        max-height: 80vh;
        display: flex;
        flex-direction: column;
    }

    .modal-header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: 1rem;
        border-bottom: 1px solid rgba(255, 255, 255, 0.08);
    }

    .modal-header h2 {
        margin: 0;
        font-size: 1.1rem;
        font-weight: 600;
        color: #e4e4e7;
    }

    .close-btn {
        padding: 0.25rem;
        background: transparent;
        border: none;
        color: #71717a;
        cursor: pointer;
        border-radius: 4px;
        transition: all 0.2s;
    }

    .close-btn:hover {
        background: rgba(255, 255, 255, 0.1);
        color: #e4e4e7;
    }

    .modal-body {
        flex: 1;
        overflow-y: auto;
        padding: 1rem;
    }

    .settings-section {
        margin-bottom: 1.5rem;
    }

    .settings-section:last-of-type {
        margin-bottom: 0;
    }

    .settings-section h3 {
        margin: 0 0 0.75rem 0;
        font-size: 0.875rem;
        font-weight: 600;
        color: #ef4444;
        text-transform: uppercase;
        letter-spacing: 0.05em;
    }

    .setting-row {
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: 0.75rem;
        background: #0f0f0f;
        border: 1px solid rgba(255, 255, 255, 0.08);
        border-radius: 4px;
        margin-bottom: 0.5rem;
        gap: 1rem;
    }

    .setting-info {
        flex: 1;
        display: flex;
        flex-direction: column;
        gap: 0.25rem;
    }

    .setting-info label {
        font-size: 0.875rem;
        font-weight: 500;
        color: #e4e4e7;
    }

    .setting-desc {
        font-size: 0.75rem;
        color: #71717a;
        line-height: 1.4;
    }

    .setting-control {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        flex-shrink: 0;
    }

    .slider {
        width: 150px;
        height: 4px;
        background: #27272a;
        border-radius: 2px;
        outline: none;
        -webkit-appearance: none;
        cursor: pointer;
    }

    .slider::-webkit-slider-thumb {
        -webkit-appearance: none;
        appearance: none;
        width: 14px;
        height: 14px;
        background: #ef4444;
        border-radius: 50%;
        cursor: pointer;
    }

    .slider::-moz-range-thumb {
        width: 14px;
        height: 14px;
        background: #ef4444;
        border-radius: 50%;
        cursor: pointer;
        border: none;
    }

    .value-label {
        min-width: 40px;
        font-size: 0.875rem;
        font-weight: 500;
        color: #e4e4e7;
        text-align: right;
    }

    .number-input {
        width: 70px;
        padding: 0.375rem 0.5rem;
        background: #0f0f0f;
        border: 1px solid rgba(255, 255, 255, 0.1);
        border-radius: 4px;
        color: #e4e4e7;
        font-size: 0.875rem;
        font-weight: 500;
        text-align: center;
    }

    .number-input:focus {
        outline: none;
        border-color: rgba(239, 68, 68, 0.5);
    }

    .unit-label {
        font-size: 0.75rem;
        color: #71717a;
    }

    .checkbox-label {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        cursor: pointer;
    }

    .checkbox-label input[type="checkbox"] {
        width: 16px;
        height: 16px;
        cursor: pointer;
        accent-color: #ef4444;
    }

    .checkbox-label span {
        font-size: 0.875rem;
        font-weight: 500;
        color: #e4e4e7;
    }

    .info-box {
        margin-top: 1rem;
        padding: 0.75rem;
        background: rgba(245, 158, 11, 0.1);
        border: 1px solid rgba(245, 158, 11, 0.2);
        border-radius: 4px;
        color: #f59e0b;
        font-size: 0.75rem;
        line-height: 1.4;
    }

    .modal-actions {
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: 1rem;
        border-top: 1px solid rgba(255, 255, 255, 0.08);
    }

    .btn-reset {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        padding: 0.5rem 0.75rem;
        background: transparent;
        border: 1px solid rgba(255, 255, 255, 0.1);
        border-radius: 4px;
        color: #9ca3af;
        font-size: 0.875rem;
        font-weight: 500;
        cursor: pointer;
        transition: all 0.2s;
    }

    .btn-reset:hover {
        background: rgba(255, 255, 255, 0.05);
        color: #e4e4e7;
        border-color: rgba(255, 255, 255, 0.2);
    }

    .action-buttons {
        display: flex;
        gap: 0.5rem;
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
        background: transparent;
        border: 1px solid rgba(255, 255, 255, 0.1);
        color: #9ca3af;
    }

    .btn-secondary:hover {
        background: rgba(255, 255, 255, 0.05);
        color: #e4e4e7;
        border-color: rgba(255, 255, 255, 0.2);
    }

    .btn-primary {
        background: #dc2626;
        color: white;
    }

    .btn-primary:hover {
        background: #ef4444;
    }
</style>