import { writable } from 'svelte/store';
import type { Environment } from '../types';

function createEnvironmentStore() {
    const { subscribe, set, update } = writable<{
        environments: Environment[];
        activeEnvironmentId: string | null;
    }>({
        environments: [],
        activeEnvironmentId: null
    });

    return {
        subscribe,
        setEnvironments: (environments: Environment[]) => update(state => ({ ...state, environments })),
        addEnvironment: (environment: Environment) => update(state => ({
            ...state,
            environments: [...state.environments, environment]
        })),
        setActive: (id: string | null) => update(state => ({ ...state, activeEnvironmentId: id })),
        getActive: () => {
            let active: Environment | undefined;
            subscribe(state => {
                active = state.environments.find(e => e.id === state.activeEnvironmentId);
            })();
            return active;
        }
    };
}

export const environmentStore = createEnvironmentStore();