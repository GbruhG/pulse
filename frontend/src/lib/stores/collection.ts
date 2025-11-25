import { writable } from 'svelte/store';
import type { Collection } from '../types';

function createCollectionStore() {
    const { subscribe, set, update } = writable<Collection[]>([]);

    return {
        subscribe,
        setCollections: (collections: Collection[]) => set(collections),
        addCollection: (collection: Collection) => update(state => [...state, collection]),
        removeCollection: (id: string) => update(state => state.filter(c => c.id !== id)),
        updateCollection: (id: string, updates: Partial<Collection>) => update(state =>
            state.map(c => c.id === id ? { ...c, ...updates } : c)
        )
    };
}

export const collectionStore = createCollectionStore();