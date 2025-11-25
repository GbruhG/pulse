import { writable } from 'svelte/store';

export interface Settings {
    uiScale: number; // 80-150
    theme: 'dark' | 'light';
    layoutMode: 'horizontal' | 'vertical';
    autoSaveHistory: boolean;
    maxHistoryItems: number;
    defaultTimeout: number; // seconds
    prettyPrintByDefault: boolean;
}

const defaultSettings: Settings = {
    uiScale: 100,
    theme: 'dark',
    layoutMode: 'horizontal',
    autoSaveHistory: true,
    maxHistoryItems: 100,
    defaultTimeout: 30,
    prettyPrintByDefault: true
};

function createSettingsStore() {
    const { subscribe, set, update } = writable<Settings>(defaultSettings);

    return {
        subscribe,
        set,
        update,
        reset: () => set(defaultSettings)
    };
}

export const settingsStore = createSettingsStore();