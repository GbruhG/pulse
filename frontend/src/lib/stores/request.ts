import { writable } from 'svelte/store';
import type { RequestData, ResponseData } from '../types';

function createRequestStore() {
    const { subscribe, set, update } = writable<{
        current: RequestData;
        response: ResponseData | null;
        loading: boolean;
    }>({
        current: {
            method: 'GET',
            url: '',
            params: [],
            headers: [],
            body: '',
            bodyType: 'none',
            auth: null
        },
        response: null,
        loading: false
    });

    return {
        subscribe,
        setRequest: (request: RequestData) => update(state => ({ ...state, current: request })),
        setResponse: (response: ResponseData | null) => update(state => ({ ...state, response })),
        setLoading: (loading: boolean) => update(state => ({ ...state, loading })),
        updateRequest: (updates: Partial<RequestData>) => update(state => ({
            ...state,
            current: { ...state.current, ...updates }
        }))
    };
}

export const requestStore = createRequestStore();