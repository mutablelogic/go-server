import { test } from 'node:test';
import assert from 'node:assert/strict';
import { ApiError, request } from '../src/api/index.ts';

test('request', async () => {
    // A request is JSON, beneath the base, with the query, and its body
    let seen: { url: string; method?: string; body?: unknown; type?: string | null } | undefined;
    const fetch = (async (input: RequestInfo | URL, init?: RequestInit) => {
        seen = { url: String(input), method: init?.method, body: init?.body, type: new Headers(init?.headers).get('Content-Type') };
        return new Response(JSON.stringify({ count: 1, body: [{ name: 'postgres' }] }), { status: 200 });
    }) as typeof globalThis.fetch;
    const list = await request<{ count: number }>({ base: '/api', fetch }, 'GET', 'role', { query: { offset: 0, limit: 10 } });
    assert.equal(list.count, 1);
    assert.deepEqual(seen, { url: '/api/role?offset=0&limit=10', method: 'GET', body: undefined, type: null });
    await request({ base: '/api', fetch }, 'POST', 'role', { body: { name: 'x' } });
    assert.deepEqual(seen, { url: '/api/role', method: 'POST', body: '{"name":"x"}', type: 'application/json' });

    // A response which isn't ok is an ApiError, and an empty body is
    // undefined
    const failing = (async () => new Response(JSON.stringify({ reason: 'Not Found' }), { status: 404 })) as typeof globalThis.fetch;
    await assert.rejects(request({ base: '/api', fetch: failing }, 'GET', 'role/x'), (error: unknown) => error instanceof ApiError && error.status === 404);
    const empty = (async () => new Response(null, { status: 204 })) as typeof globalThis.fetch;
    assert.equal(await request({ base: '/api', fetch: empty }, 'DELETE', 'role/x'), undefined);
});
