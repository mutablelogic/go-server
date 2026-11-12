import { test } from 'node:test';
import assert from 'node:assert/strict';
import { ApiError } from '../src/api/index.ts';
import { configClient, getConfig, type Config } from '../src/config/index.ts';

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
