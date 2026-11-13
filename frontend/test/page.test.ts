import { test } from 'node:test';
import assert from 'node:assert/strict';
import { runAction, type PageAction } from '../src/page/action.ts';

test('runAction', async (t) => {
    const record = () => {
        const busy: boolean[] = [];
        const errors: unknown[] = [];
        return { busy, errors, onBusy: (b: boolean) => busy.push(b), onError: (e: unknown) => errors.push(e) };
    };

    await t.test('runs, and is never busy, without a promise', async () => {
        let ran = false;
        const r = record();
        await runAction({ label: 'Go', run: () => { ran = true; } }, r.onBusy, r.onError);
        assert.equal(ran, true);
        assert.deepEqual(r.busy, []);
        assert.deepEqual(r.errors, []);
    });

    await t.test('is busy while its promise is pending', async () => {
        const r = record();
        let resolve!: () => void;
        const done = runAction({ label: 'Go', run: () => new Promise<void>((res) => { resolve = res; }) }, r.onBusy, r.onError);
        assert.deepEqual(r.busy, [true]);
        resolve();
        await done;
        assert.deepEqual(r.busy, [true, false]);
        assert.deepEqual(r.errors, []);
    });

    await t.test('reports a rejection, and is then not busy', async () => {
        const r = record();
        const error = new Error('no');
        await runAction({ label: 'Go', run: () => Promise.reject(error) }, r.onBusy, r.onError);
        assert.deepEqual(r.busy, [true, false]);
        assert.deepEqual(r.errors, [error]);
    });

    await t.test('a promise of another realm, which has a then, is one too', async () => {
        const r = record();
        const error = new Error('elsewhere');
        const thenable: PromiseLike<unknown> = { then: (_ok, fail) => Promise.reject(error).then(undefined, fail) };
        await runAction({ label: 'Go', run: () => thenable }, r.onBusy, r.onError);
        assert.deepEqual(r.busy, [true, false]);
        assert.deepEqual(r.errors, [error]);
    });

    await t.test('reports what it throws', async () => {
        const r = record();
        const error = new Error('thrown');
        const action: PageAction = { label: 'Go', run: () => { throw error; } };
        await runAction(action, r.onBusy, r.onError);
        assert.deepEqual(r.busy, []);
        assert.deepEqual(r.errors, [error]);
    });
});
