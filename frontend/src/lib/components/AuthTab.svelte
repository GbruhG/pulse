<script lang="ts">
    import { requestStore } from '../stores/request';

    $: auth = $requestStore.current.auth;
    $: authType = auth?.type || 'none';

    let showPassword = false;

    function setAuthType(type: 'none' | 'basic' | 'bearer' | 'api-key' | 'oauth2') {
        if (type === 'none') {
            requestStore.updateRequest({ auth: null });
        } else {
            requestStore.updateRequest({
                auth: {
                    type,
                    username: '',
                    password: '',
                    token: '',
                    key: '',
                    value: ''
                }
            });
        }
    }

    function updateAuth(field: string, value: string) {
        if (!auth) return;
        requestStore.updateRequest({
            auth: { ...auth, [field]: value }
        });
    }
</script>

<div class="auth-container">
    <div class="auth-header">
        <h3>Authentication</h3>
        <p>Configure how requests are authenticated</p>
    </div>

    <div class="auth-type-selector">
        <label class="section-label">Auth Type</label>
        <div class="auth-types">
            <button
                    class="auth-type-btn"
                    class:active={authType === 'none'}
                    on:click={() => setAuthType('none')}
            >
                No Auth
            </button>
            <button
                    class="auth-type-btn"
                    class:active={authType === 'basic'}
                    on:click={() => setAuthType('basic')}
            >
                Basic Auth
            </button>
            <button
                    class="auth-type-btn"
                    class:active={authType === 'bearer'}
                    on:click={() => setAuthType('bearer')}
            >
                Bearer Token
            </button>
            <button
                    class="auth-type-btn"
                    class:active={authType === 'api-key'}
                    on:click={() => setAuthType('api-key')}
            >
                API Key
            </button>
            <button
                    class="auth-type-btn"
                    class:active={authType === 'oauth2'}
                    on:click={() => setAuthType('oauth2')}
            >
                OAuth 2.0
            </button>
        </div>
    </div>

    {#if authType === 'basic'}
        <div class="auth-fields">
            <div class="field-group">
                <label>Username</label>
                <input
                        type="text"
                        value={auth?.username || ''}
                        on:input={(e) => updateAuth('username', e.currentTarget.value)}
                        placeholder="Enter username"
                        class="auth-input"
                />
            </div>
            <div class="field-group">
                <label>Password</label>
                <div class="password-wrapper">
                    <input
                            type={showPassword ? 'text' : 'password'}
                            value={auth?.password || ''}
                            on:input={(e) => updateAuth('password', e.currentTarget.value)}
                            placeholder="Enter password"
                            class="auth-input"
                    />
                </div>
            </div>
            <label class="checkbox-label">
                <input type="checkbox" bind:checked={showPassword} />
                <span>Show password</span>
            </label>
        </div>
    {:else if authType === 'bearer'}
        <div class="auth-fields">
            <div class="field-group">
                <label>Token</label>
                <input
                        type="text"
                        value={auth?.token || ''}
                        on:input={(e) => updateAuth('token', e.currentTarget.value)}
                        placeholder="Enter bearer token"
                        class="auth-input"
                />
                <span class="field-hint">Will be sent as: Authorization: Bearer {auth?.token || '<token>'}</span>
            </div>
        </div>
    {:else if authType === 'api-key'}
        <div class="auth-fields">
            <div class="field-group">
                <label>Key</label>
                <input
                        type="text"
                        value={auth?.key || ''}
                        on:input={(e) => updateAuth('key', e.currentTarget.value)}
                        placeholder="e.g., X-API-Key"
                        class="auth-input"
                />
            </div>
            <div class="field-group">
                <label>Value</label>
                <input
                        type="text"
                        value={auth?.value || ''}
                        on:input={(e) => updateAuth('value', e.currentTarget.value)}
                        placeholder="Enter API key value"
                        class="auth-input"
                />
            </div>
        </div>
    {:else if authType === 'oauth2'}
        <div class="coming-soon-box">
            <p>OAuth 2.0 configuration coming soon!</p>
            <span>For now, use Bearer Token with your OAuth token</span>
        </div>
    {:else}
        <div class="no-auth-box">
            <p>This request does not use any authentication.</p>
            <span>Select an auth type above to configure authentication</span>
        </div>
    {/if}
</div>

<style>
    .auth-container {
        padding: 1rem;
        max-width: 48rem;
    }

    .auth-header {
        margin-bottom: 1.5rem;
    }

    .auth-header h3 {
        margin: 0 0 0.25rem 0;
        font-size: 1rem;
        font-weight: 600;
        color: #e4e4e7;
    }

    .auth-header p {
        margin: 0;
        font-size: 0.875rem;
        color: #9ca3af;
    }

    .auth-type-selector {
        margin-bottom: 1.5rem;
    }

    .section-label {
        display: block;
        font-size: 0.875rem;
        font-weight: 600;
        color: #d1d5db;
        margin-bottom: 0.5rem;
    }

    .auth-types {
        display: flex;
        gap: 0.5rem;
        flex-wrap: wrap;
    }

    .auth-type-btn {
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

    .auth-type-btn:hover {
        background: rgba(255, 255, 255, 0.05);
        border-color: rgba(255, 255, 255, 0.2);
        color: #e4e4e7;
    }

    .auth-type-btn.active {
        background: rgba(239, 68, 68, 0.1);
        border-color: #ef4444;
        color: #ef4444;
    }

    .auth-fields {
        display: flex;
        flex-direction: column;
        gap: 1rem;
    }

    .field-group {
        display: flex;
        flex-direction: column;
        gap: 0.25rem;
    }

    .field-group label {
        font-size: 0.875rem;
        font-weight: 500;
        color: #d1d5db;
    }

    .auth-input {
        width: 100%;
        background: #0f0f0f;
        border: 1px solid rgba(255, 255, 255, 0.1);
        border-radius: 4px;
        padding: 0.5rem 0.75rem;
        color: #e4e4e7;
        font-size: 0.875rem;
        outline: none;
        transition: all 0.2s;
    }

    .auth-input:focus {
        border-color: rgba(239, 68, 68, 0.5);
    }

    .auth-input::placeholder {
        color: #52525b;
    }

    .field-hint {
        font-size: 0.75rem;
        color: #71717a;
        font-family: 'Monaco', 'Menlo', monospace;
    }

    .password-wrapper {
        position: relative;
    }

    .checkbox-label {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        font-size: 0.875rem;
        color: #9ca3af;
        cursor: pointer;
    }

    .checkbox-label input[type='checkbox'] {
        width: 16px;
        height: 16px;
        cursor: pointer;
        accent-color: #ef4444;
    }

    .no-auth-box,
    .coming-soon-box {
        padding: 1rem;
        background: #0f0f0f;
        border: 1px dashed rgba(255, 255, 255, 0.1);
        border-radius: 4px;
        text-align: center;
    }

    .no-auth-box p,
    .coming-soon-box p {
        margin: 0 0 0.25rem 0;
        font-size: 0.875rem;
        font-weight: 500;
        color: #d1d5db;
    }

    .no-auth-box span,
    .coming-soon-box span {
        font-size: 0.75rem;
        color: #6b7280;
    }
</style>