import { writable, derived } from 'svelte/store';

type MessageDirection = 'inbound' | 'outbound' | 'error' | 'system';

interface StreamMessage {
    id: string;
    direction: MessageDirection;
    protocol: string;
    payload: string;
    timestamp: Date;
}

interface StreamStore {
    messages: StreamMessage[];
    isPaused: boolean;
    filterDirection: MessageDirection | 'all';
    searchTerm: string;
}

const MAX_MESSAGES = 1000;

function createStreamMessageStore() {
    const { subscribe, set, update } = writable<StreamStore>({
        messages: [],
        isPaused: false,
        filterDirection: 'all',
        searchTerm: ''
    });

    let messageQueue: StreamMessage[] = [];
    let isProcessing = false;

    async function processQueue() {
        if (isProcessing || messageQueue.length === 0) return;

        isProcessing = true;

        // Process all queued messages at once
        const batch = messageQueue.splice(0, messageQueue.length);

        update(store => {
            if (store.isPaused) return store;

            let newMessages = [...store.messages, ...batch];

            if (newMessages.length > MAX_MESSAGES) {
                newMessages = newMessages.slice(-MAX_MESSAGES);
            }

            return {
                ...store,
                messages: newMessages
            };
        });

        isProcessing = false;

        // If more messages arrived while processing, process them
        if (messageQueue.length > 0) {
            setTimeout(processQueue, 50);
        }
    }

    return {
        subscribe,
        addMessage: (message: StreamMessage) => {
            messageQueue.push(message);

            if (!isProcessing) {
                requestAnimationFrame(processQueue);
            }
        },
        setFilter: (filterDirection: MessageDirection | 'all') => {
            update(store => ({ ...store, filterDirection }));
        },
        setSearch: (searchTerm: string) => {
            update(store => ({ ...store, searchTerm }));
        },
        togglePause: () => {
            update(store => ({ ...store, isPaused: !store.isPaused }));
        },
        clear: () => {
            messageQueue = [];
            update(store => ({ ...store, messages: [] }));
        },
        reset: () => {
            messageQueue = [];
            set({
                messages: [],
                isPaused: false,
                filterDirection: 'all',
                searchTerm: ''
            });
        },
        getPendingCount: () => messageQueue.length
    };
}

export const streamMessageStore = createStreamMessageStore();

export const filteredMessages = derived(
    streamMessageStore,
    $store => {
        return $store.messages.filter(msg => {
            const matchesDirection = $store.filterDirection === 'all' || msg.direction === $store.filterDirection;
            const matchesSearch = !$store.searchTerm ||
                msg.payload.toLowerCase().includes($store.searchTerm.toLowerCase()) ||
                msg.protocol.toLowerCase().includes($store.searchTerm.toLowerCase());
            return matchesDirection && matchesSearch;
        });
    }
);