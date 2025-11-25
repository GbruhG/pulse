import { writable } from 'svelte/store';
import type { Workspace } from '../types';

function createWorkspaceStore() {
    const { subscribe, set, update } = writable<{
        workspaces: Workspace[];
        activeWorkspaceId: string | null;
    }>({
        workspaces: [],
        activeWorkspaceId: null
    });

    return {
        subscribe,
        setWorkspaces: (workspaces: Workspace[]) => update(state => ({ ...state, workspaces })),
        addWorkspace: (workspace: Workspace) => update(state => ({
            ...state,
            workspaces: [...state.workspaces, workspace]
        })),
        setActive: (id: string) => update(state => ({ ...state, activeWorkspaceId: id })),
        getActive: () => {
            let active: Workspace | undefined;
            subscribe(state => {
                active = state.workspaces.find(w => w.id === state.activeWorkspaceId);
            })();
            return active;
        }
    };
}

export const workspaceStore = createWorkspaceStore();