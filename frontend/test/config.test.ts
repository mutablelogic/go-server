import { test } from 'node:test';
import assert from 'node:assert/strict';
import { ApiError } from '../src/api/index.ts';
import { configClient, configSection, getConfig, type Config } from '../src/config/index.ts';

const config: Config = { prefix: '/api', name: 'server', version: '1.2.3', openapi: '/api/openapi.html' };

test('getConfig', async () => {
    // The configuration is read from config.json, beside the page, without
    // the cache, and gives the API's client
    let request: { url: string; cache?: RequestCache } | undefined;
    const fetch = (async (input: RequestInfo | URL, init?: RequestInit) => {
        request = { url: String(input), cache: init?.cache };
        return new Response(JSON.stringify(config), { status: 200, headers: { 'Content-Type': 'application/json' } });
    }) as typeof globalThis.fetch;
    const got = await getConfig({ fetch });
    assert.deepEqual(got, config);
    assert.deepEqual(request, { url: 'config.json', cache: 'no-store' });
    assert.equal(configClient(got, fetch).base, '/api');
});

test('getConfig error', async () => {
    const fetch = (async () => new Response('', { status: 404, statusText: 'Not Found' })) as typeof globalThis.fetch;
    await assert.rejects(getConfig({ url: '/config.json', fetch }), (error: unknown) => error instanceof ApiError && error.status === 404);
});

test('configSection', () => {
    const config = { prefix: '/api', modules: { credential: { decrypt: true }, empty: null } };
    assert.deepEqual(configSection<{ decrypt: boolean }>(config, 'credential'), { decrypt: true });
    assert.equal(configSection(config, 'queue'), undefined);
    assert.equal(configSection(config, 'empty'), undefined);
    assert.equal(configSection({ prefix: '/api' }, 'credential'), undefined, 'an older server has no modules');
    assert.equal(configSection(undefined, 'credential'), undefined);

    // Only the modules' own sections, not what every object inherits, even
    // from JSON
    const parsed = JSON.parse('{"prefix":"/api","modules":{"credential":{"decrypt":true}}}') as Config;
    for (const name of ['constructor', 'toString', '__proto__', 'hasOwnProperty']) {
        assert.equal(configSection(parsed, name), undefined, name);
    }
    assert.deepEqual(configSection(parsed, 'credential'), { decrypt: true });
});
