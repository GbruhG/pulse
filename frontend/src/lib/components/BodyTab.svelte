<script lang="ts">
    import { requestStore } from '../stores/request';

    $: bodyType = $requestStore.current.bodyType;
    $: body = $requestStore.current.body;

    function setBodyType(type: 'none' | 'json' | 'xml' | 'text' | 'form-data' | 'x-www-form-urlencoded') {
        requestStore.updateRequest({ bodyType: type });

        // Set appropriate content-type header
        const headers = $requestStore.current.headers;
        const contentTypeIndex = headers.findIndex(h => h.key.toLowerCase() === 'content-type');

        let contentType = '';
        switch (type) {
            case 'json':
                contentType = 'application/json';
                break;
            case 'xml':
                contentType = 'application/xml';
                break;
            case 'text':
                contentType = 'text/plain';
                break;
            case 'form-data':
                contentType = 'multipart/form-data';
                break;
            case 'x-www-form-urlencoded':
                contentType = 'application/x-www-form-urlencoded';
                break;
        }

        if (contentType && contentTypeIndex === -1) {
            requestStore.updateRequest({
                headers: [...headers, {
                    id: crypto.randomUUID(),
                    key: 'Content-Type',
                    value: contentType,
                    enabled: true
                }]
            });
        } else if (contentType && contentTypeIndex >= 0) {
            const newHeaders = [...headers];
            newHeaders[contentTypeIndex] = { ...newHeaders[contentTypeIndex], value: contentType };
            requestStore.updateRequest({ headers: newHeaders });
        }
    }

    function updateBody(value: string) {
        requestStore.updateRequest({ body: value });
    }

    function formatJSON() {
        try {
            const formatted = JSON.stringify(JSON.parse(body), null, 2);
            updateBody(formatted);
        } catch (e) {
            // Invalid JSON, do nothing
        }
    }

    function minifyJSON() {
        try {
            const minified = JSON.stringify(JSON.parse(body));
            updateBody(minified);
        } catch (e) {
            // Invalid JSON, do nothing
        }
    }
</script>

<div class="body-container">
    <div class="body-header">
        <div class="body-types">
            <button
                    class="type-btn"
                    class:active={bodyType === 'none'}
                    on:click={() => setBodyType('none')}
            >
                None
            </button>
            <button
                    class="type-btn"
                    class:active={bodyType === 'json'}
                    on:click={() => setBodyType('json')}
            >
                JSON
            </button>
            <button
                    class="type-btn"
                    class:active={bodyType === 'xml'}
                    on:click={() => setBodyType('xml')}
            >
                XML
            </button>
            <button
                    class="type-btn"
                    class:active={bodyType === 'text'}
                    on:click={() => setBodyType('text')}
            >
                Text
            </button>
            <button
                    class="type-btn"
                    class:active={bodyType === 'form-data'}
                    on:click={() => setBodyType('form-data')}
            >
                Form Data
            </button>
            <button
                    class="type-btn"
                    class:active={bodyType === 'x-www-form-urlencoded'}
                    on:click={() => setBodyType('x-www-form-urlencoded')}
            >
                URL Encoded
            </button>
        </div>

        {#if bodyType === 'json'}
            <div class="body-actions">
                <button class="action-btn" on:click={formatJSON}>
                    Prettify
                </button>
                <button class="action-btn" on:click={minifyJSON}>
                    Minify
                </button>
            </div>
        {/if}
    </div>

    {#if bodyType === 'none'}
        <div class="empty-state">
            <p>This request does not have a body</p>
            <span>Select a body type above to add content</span>
        </div>
    {:else if bodyType === 'json' || bodyType === 'xml' || bodyType === 'text'}
        <div class="editor-container">
      <textarea
              class="body-editor"
              placeholder={bodyType === 'json' ? '{\n  "key": "value"\n}' : bodyType === 'xml' ? '<?xml version="1.0"?>\n<root>\n  <element>value</element>\n</root>' : 'Enter text content'}
              value={body}
              on:input={(e) => updateBody(e.currentTarget.value)}
              spellcheck="false"
      />
            <div class="editor-info">
        <span class="info-item">
          Lines: {body.split('\n').length}
        </span>
                <span class="info-item">
          Size: {new Blob([body]).size} bytes
        </span>
            </div>
        </div>
    {:else if bodyType === 'form-data' || bodyType === 'x-www-form-urlencoded'}
        <div class="form-data-container">
            <p class="coming-soon">Form data editor coming soon!</p>
            <p class="suggestion">For now, use the Raw JSON option or Headers tab</p>
        </div>
    {/if}
</div>

<style>
    .body-container {
        display: flex;
        flex-direction: column;
        height: 100%;
    }

    .body-header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: 0.75rem 1rem;
        background: #0a0a0a;
        border-bottom: 1px solid rgba(255, 255, 255, 0.08);
    }

    .body-types {
        display: flex;
        gap: 0.25rem;
    }

    .type-btn {
        padding: 0.375rem 0.75rem;
        background: transparent;
        border: 1px solid rgba(255, 255, 255, 0.1);
        border-radius: 4px;
        color: #9ca3af;
        font-size: 0.875rem;
        font-weight: 500;
        cursor: pointer;
        transition: all 0.2s;
    }

    .type-btn:hover {
        background: rgba(255, 255, 255, 0.05);
        color: #e4e4e7;
    }

    .type-btn.active {
        background: rgba(239, 68, 68, 0.1);
        border-color: #ef4444;
        color: #ef4444;
    }

    .body-actions {
        display: flex;
        gap: 0.5rem;
    }

    .action-btn {
        padding: 0.375rem 0.75rem;
        background: transparent;
        border: 1px solid rgba(255, 255, 255, 0.1);
        border-radius: 4px;
        color: #9ca3af;
        font-size: 0.875rem;
        font-weight: 500;
        cursor: pointer;
        transition: all 0.2s;
    }

    .action-btn:hover {
        background: rgba(255, 255, 255, 0.05);
        color: #e4e4e7;
        border-color: rgba(255, 255, 255, 0.2);
    }

    .empty-state {
        flex: 1;
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        gap: 0.25rem;
        padding: 2rem 1rem;
        text-align: center;
    }

    .empty-state p {
        margin: 0;
        font-size: 0.875rem;
        color: #d1d5db;
    }

    .empty-state span {
        margin: 0;
        font-size: 0.75rem;
        color: #6b7280;
    }

    .editor-container {
        flex: 1;
        display: flex;
        flex-direction: column;
        background: #0a0a0a;
    }

    .body-editor {
        flex: 1;
        width: 100%;
        background: transparent;
        border: none;
        color: #d1d5db;
        font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
        font-size: 0.875rem;
        line-height: 1.6;
        padding: 1rem;
        resize: none;
        outline: none;
    }

    .body-editor::placeholder {
        color: #52525b;
    }

    .editor-info {
        display: flex;
        gap: 1rem;
        padding: 0.5rem 1rem;
        background: #0a0a0a;
        border-top: 1px solid rgba(255, 255, 255, 0.08);
        font-size: 0.75rem;
        color: #71717a;
    }

    .info-item {
        display: flex;
        align-items: center;
        gap: 0.25rem;
    }

    .form-data-container {
        flex: 1;
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        gap: 0.25rem;
        padding: 2rem 1rem;
        text-align: center;
    }

    .coming-soon {
        margin: 0;
        font-size: 0.875rem;
        color: #d1d5db;
        font-weight: 500;
    }

    .suggestion {
        margin: 0;
        font-size: 0.75rem;
        color: #6b7280;
    }
</style>