import { test } from 'node:test';
import assert from 'node:assert/strict';
import { ApiError } from '../src/api/index.ts';
import { ChangesController, changeMatches, watchChanges, type Change } from '../src/broadcaster/index.ts';
import { fakeFetch, sse, waitFor } from './fake.ts';

const created: Change = { action: 'create', schema: 'llm', table: 'provider', row: 'ollama' };
const updated: Change = { action: 'update', schema: 'llm', table: 'session', row: '1', data: { serial: 2 } };

test('changeMatches', () => {
    assert.ok(changeMatches({}, created));
    assert.ok(changeMatches({ schema: 'llm', table: 'provider', action: 'create' }, created));
    assert.ok(!changeMatches({ table: 'session' }, created));
    assert.ok(!changeMatches({ action: 'delete' }, created));
});

test('watchChanges', async () => {
    // The changes are passed to the function, with the filter as the query,
    // and a shutdown is returned, and said
    const fake = fakeFetch({ chunks: [sse('change', created), sse('change', updated), sse('shutdown', {})] });
    const changes: Change[] = [];
    let shutdowns = 0;
    const shutdown = await watchChanges({ base: '/api', fetch: fake.fetch }, (change) => changes.push(change), {
        filter: { schema: 'llm', action: 'update' },
        onShutdown: () => shutdowns++,
    });
    assert.equal(shutdown, true);
    assert.equal(shutdowns, 1);
    assert.deepEqual(changes, [created, updated]);
    assert.equal(fake.requests[0].url, '/api/changes?schema=llm&action=update');
});

test('watchChanges unavailable', async () => {
    // A server which is stopping, before the stream starts, is an ApiError
    const fake = fakeFetch({ status: 503, body: { reason: 'Service Unavailable' } });
    await assert.rejects(watchChanges({ base: '/api', fetch: fake.fetch }, () => {}), (error: unknown) => error instanceof ApiError && error.status === 503);
});

test('ChangesController', async (t) => {
    await t.test('subscribe', async () => {
        // Every listener is called once the stream starts, and then with the
        // last change of a burst which matches its filters, and not once
        // it's unsubscribed
        const fake = fakeFetch({ open: true, chunks: [sse('change', created), sse('change', updated), sse('change', created)] });
        const controller = new ChangesController({ base: '/api', fetch: fake.fetch, interval: 20 });
        const providers: (Change | undefined)[] = [];
        const sessions: (Change | undefined)[] = [];
        let connects = 0;
        controller.onConnect(() => connects++);
        controller.on({ table: 'provider' }, (change) => providers.push(change));
        const unsubscribe = controller.subscribe([{ table: 'session' }, { table: 'nothing' }], (change) => sessions.push(change));
        controller.start();
        await waitFor(() => providers.length > 0 && sessions.length > 0);
        assert.equal(connects, 1);
        assert.ok(controller.connected);

        // The stream started, so changes may have been missed, which is no
        // change, in the same interval as the changes
        assert.deepEqual(providers, [undefined]);
        assert.deepEqual(sessions, [undefined]);
        unsubscribe();
        controller.stop();
        assert.ok(!controller.connected);
    });

    await t.test('burst', async () => {
        // A burst of changes, after the stream has started, is one call,
        // with the last change
        const last = { ...updated, data: { serial: 3 } };
        const fake = fakeFetch({ open: true, later: { delay: 50, chunks: [sse('change', updated), sse('change', last)] } });
        const controller = new ChangesController({ base: '/api', fetch: fake.fetch, interval: 20 });
        const sessions: (Change | undefined)[] = [];
        controller.on({ table: 'session' }, (change) => sessions.push(change));
        controller.start();
        await waitFor(() => sessions.length === 2);
        await new Promise((resolve) => setTimeout(resolve, 40));
        controller.stop();
        assert.deepEqual(sessions, [undefined, last]);
    });

    await t.test('late subscriber', async () => {
        // A listener which subscribes once the stream has started is called,
        // so it loads what it shows
        const fake = fakeFetch({ open: true });
        const controller = new ChangesController({ base: '/api', fetch: fake.fetch, interval: 10 });
        controller.start();
        await waitFor(() => controller.connected);
        const calls: (Change | undefined)[] = [];
        const unsubscribe = controller.subscribe({ table: 'session' }, (change) => calls.push(change));
        await waitFor(() => calls.length === 1);
        assert.deepEqual(calls, [undefined]);
        unsubscribe();
        controller.stop();
    });

    await t.test('reconnect', async () => {
        // A stream which fails is an error, said once, and a disconnect,
        // and it's watched again, after a delay, until it starts
        const fake = fakeFetch(
            { status: 500, body: { reason: 'boom' } },
            { status: 500, body: { reason: 'boom' } },
            { open: true, chunks: [sse('change', created)] },
        );
        const controller = new ChangesController({ base: '/api', fetch: fake.fetch, minDelay: 5, maxDelay: 10, interval: 10 });
        const errors: unknown[] = [];
        let disconnects = 0;
        const changes: (Change | undefined)[] = [];
        controller.onError((error) => errors.push(error)).onDisconnect(() => disconnects++).on({}, (change) => changes.push(change));
        controller.start();
        await waitFor(() => controller.connected && changes.length > 0);
        controller.stop();
        assert.equal(fake.requests.length, 3);
        assert.equal(errors.length, 1);
        assert.ok(errors[0] instanceof ApiError);
        assert.equal(disconnects, 1);
    });

    await t.test('shutdown', async () => {
        // A server which stops says so, and is a disconnect, and it's
        // watched again
        const fake = fakeFetch({ chunks: [sse('shutdown', {})] }, { open: true });
        const controller = new ChangesController({ base: '/api', fetch: fake.fetch, minDelay: 5, interval: 10 });
        let shutdowns = 0;
        let disconnects = 0;
        controller.onShutdown(() => shutdowns++).onDisconnect(() => disconnects++);
        controller.start();
        await waitFor(() => fake.requests.length === 2 && controller.connected);
        controller.stop();
        assert.equal(shutdowns, 1);
        assert.equal(disconnects, 1);
    });
});
