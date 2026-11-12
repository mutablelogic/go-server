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
