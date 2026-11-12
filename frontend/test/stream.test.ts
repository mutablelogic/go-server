import { test } from 'node:test';
import assert from 'node:assert/strict';
import { ApiError, requestStream, url } from '../src/api/index.ts';
import { fakeFetch } from './fake.ts';

test('url', () => {
    // The path is beneath the base, with the query, without undefined values
    assert.equal(url('/api', 'changes'), '/api/changes');
    assert.equal(url('http://localhost:8084/api/', '/changes', { table: 'session', action: undefined }), 'http://localhost:8084/api/changes?table=session');
});

test('requestStream', async () => {
    // Events are parsed across chunks, with data on more than one line, and
    // those without data, such as pings, are ignored
    const fake = fakeFetch({ chunks: ['event: ping\n\nevent: one\nda', 'ta: {"a":1}\n\r\nevent: two\ndata: [1,\ndata: 2]\n\n'] });
    const events: [string, unknown][] = [];
    let opened = false;
    await requestStream({ base: '/api', fetch: fake.fetch }, 'changes', (name, data) => events.push([name, data]), { onOpen: () => { opened = true; } });
    assert.ok(opened);
    assert.deepEqual(events, [['one', { a: 1 }], ['two', [1, 2]]]);
    assert.equal(fake.requests[0].accept, 'text/event-stream');
});

test('requestStream error', async () => {
    // A response which isn't ok is an ApiError, with its reason and status
    const fake = fakeFetch({ status: 503, body: { code: 503, reason: 'Service Unavailable: broadcaster stopped' } });
    await assert.rejects(
        requestStream({ base: '/api', fetch: fake.fetch }, 'changes', () => {}),
        (error: unknown) => error instanceof ApiError && error.status === 503 && error.message.includes('broadcaster stopped'),
    );
});

test('requestStream CRLF', async () => {
    // A CRLF which is split across chunks is one line's end, and a CR at
    // the end of the stream ends its line
    const fake = fakeFetch({ chunks: ['event: change\r', '\ndata: {"a":1}\r\n\r', '\nevent: last\rdata: 2\r\r'] });
    const events: [string, unknown][] = [];
    await requestStream({ base: '/api', fetch: fake.fetch }, 'changes', (name, data) => events.push([name, data]));
    assert.deepEqual(events, [['change', { a: 1 }], ['last', 2]]);
});

test('requestStream cancels', async () => {
    // When an event can't be parsed, or the function throws, the stream is
    // cancelled, with the error, which is thrown
    for (const [chunks, fn] of [
        [['event: change\ndata: {not json\n\n'], () => {}],
        [['event: change\ndata: 1\n\n'], () => { throw new Error('listener failed'); }],
    ] as [string[], () => void][]) {
        let cancelled: unknown;
        const fetch = (async () => new Response(new ReadableStream<Uint8Array>({
            start(controller) {
                chunks.forEach((chunk) => controller.enqueue(new TextEncoder().encode(chunk)));
            },
            cancel(reason) {
                cancelled = reason;
            },
        }), { status: 200 })) as typeof globalThis.fetch;
        await assert.rejects(requestStream({ base: '/api', fetch }, 'changes', fn), (error: unknown) => error === cancelled);
    }
});
