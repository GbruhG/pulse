import { writable } from 'svelte/store';
import type { HistoryItem } from '../types';

function createHistoryStore() {
    const { subscribe, set, update } = writable<HistoryItem[]>([]);

    return {
        subscribe,
        addItem: (item: HistoryItem) => update(state => [item, ...state].slice(0, 100)), // Keep last 100
        clearHistory: () => set([]),
        removeItem: (id: string) => update(state => state.filter(h => h.id !== id))
    };
}

export const historyStore = createHistoryStore();