import { test } from 'node:test';
import assert from 'node:assert/strict';
import { ApiError } from '../src/api/index.ts';
import { ListController, type ListState } from '../src/list/index.ts';
import { waitFor } from './fake.ts';

interface Item { name: string; n?: number }

// pages returns a list function of the items, a page at a time
function pages(items: Item[]) {
    return async (offset: number, limit: number) => ({ count: items.length, body: items.slice(offset, offset + limit) });
}

test('ListController', async (t) => {
    await t.test('load', async () => {
        // A page is loaded, and the views are told, while it loads, and
        // once it has
        const items = Array.from({ length: 30 }, (_, i) => ({ name: `item${i}` }));
        const list = new ListController<Item>({ name: 'Items', list: pages(items), same: (a, b) => a.name === b.name, pageSize: 10 });
        const states: ListState<Item>[] = [];
        list.subscribe((state) => states.push(state));
        await list.load();
        assert.equal(list.current.items?.length, 10);
        assert.equal(list.current.total, 30);
        assert.equal(states.at(-2)?.items, undefined);
        list.setPage(3);
        await list.refresh();
        assert.equal(list.current.items?.[0].name, 'item20');
    });

    await t.test('reset', async () => {
        // Another list starts at its first page, loading, with nothing open
        let items = Array.from({ length: 30 }, (_, i) => ({ name: `a${i}` }));
        const list = new ListController<Item>({ name: 'Items', list: (offset, limit) => pages(items)(offset, limit), same: (a, b) => a.name === b.name, pageSize: 10 });
        list.setPage(2);
        await list.refresh();
        list.open(items[12]);
        items = [{ name: 'b0' }];
        const states: ListState<Item>[] = [];
        list.subscribe((state) => states.push(state));
        await list.reset();
        assert.equal(states[1].items, undefined);
        assert.deepEqual(list.current.items, [{ name: 'b0' }]);
        assert.equal(list.current.page, 1);
        assert.equal(list.current.selected, undefined);
    });

    await t.test('sort', async () => {
        // The list is loaded with its sort, and from the first page, when
        // it's sorted by another
        const sorts: (string | undefined)[] = [];
        const list = new ListController<Item>({
            name: 'Items',
            list: async (offset, limit, sort) => {
                sorts.push(sort ? `${sort.key} ${sort.direction}` : undefined);
                return { count: 30, body: [{ name: `at${offset}` }] };
            },
            same: (a, b) => a.name === b.name,
            pageSize: 10,
            sort: { key: 'calls', direction: 'descending' },
        });
        list.setPage(2);
        await list.refresh();
        list.setSort({ key: 'rows', direction: 'ascending' });
        await list.refresh();
        assert.equal(list.current.page, 1);
        assert.deepEqual(list.current.sort, { key: 'rows', direction: 'ascending' });
        assert.deepEqual(sorts.at(0), 'calls descending');
        assert.deepEqual(sorts.at(-1), 'rows ascending');
        assert.equal(list.current.items?.[0].name, 'at0');
    });

    await t.test('last page', async () => {
        // A page beyond the last, as items were deleted, is the last
        const items = Array.from({ length: 30 }, (_, i) => ({ name: `item${i}` }));
        const list = new ListController<Item>({ name: 'Items', list: pages(items), same: (a, b) => a.name === b.name, pageSize: 10 });
        list.setPage(3);
        await list.refresh();
        items.splice(15);
        await list.refresh();
        assert.equal(list.current.page, 2);
        assert.equal(list.current.items?.[0].name, 'item10');
    });

    await t.test('selected', async () => {
        // The item which is opened is kept current, from the page, or by
        // itself, and is closed once it's been deleted
        const items: Item[] = [{ name: 'a', n: 1 }];
        let other: Item | undefined = { name: 'b', n: 1 };
        const list = new ListController<Item>({
            name: 'Items',
            list: pages(items),
            get: async () => {
                if (!other) {
                    throw new ApiError(404, 'Not Found');
                }
                return other;
            },
            same: (a, b) => a.name === b.name,
        });
        await list.load();
        list.open(items[0]);
        items[0] = { name: 'a', n: 2 };
        await list.refresh();
        assert.deepEqual(list.current.selected, { name: 'a', n: 2 });
        list.open({ name: 'b' });
        await list.refresh();
        assert.deepEqual(list.current.selected, { name: 'b', n: 1 });
        other = undefined;
        await list.refresh();
        assert.equal(list.current.selected, undefined);
    });

    await t.test('stale', async () => {
        // A response to a request before the list was reset, of the page it
        // was on then, isn't shown, nor does it change the page, and the
        // list is loaded again, from its first page
        const requests: { offset: number; resolve: (page: { count: number; body: Item[] }) => void }[] = [];
        const list = new ListController<Item>({
            name: 'Items',
            list: (offset) => new Promise((resolve) => requests.push({ offset, resolve })),
            same: (a, b) => a.name === b.name,
            pageSize: 10,
        });
        list.setPage(3);
        await waitFor(() => requests.length === 1);
        const done = list.reset();
        requests[0].resolve({ count: 15, body: [{ name: 'old' }] });
        await waitFor(() => requests.length === 2);
        assert.equal(list.current.page, 1);
        assert.equal(list.current.items, undefined);
        assert.equal(requests[1].offset, 0);
        requests[1].resolve({ count: 1, body: [{ name: 'new' }] });
        await done;
        assert.deepEqual(list.current.items, [{ name: 'new' }]);
        assert.equal(list.current.page, 1);
    });

    await t.test('selected while loading', async () => {
        // An item which is opened, or closed, while the one which was open is
        // loaded again, stays as it is
        let resolve: ((item: Item) => void) | undefined;
        const list = new ListController<Item>({
            name: 'Items',
            list: pages([{ name: 'a' }]),
            get: () => new Promise((r) => { resolve = r; }),
            same: (a, b) => a.name === b.name,
        });
        await list.load();
        list.open({ name: 'b' });
        const done = list.refresh();
        await waitFor(() => resolve !== undefined);
        list.close();
        resolve!({ name: 'b', n: 2 });
        await done;
        assert.equal(list.current.selected, undefined);
    });

    await t.test('unavailable', () => {
        // A list made while the changes can't be watched is unavailable from
        // the start, and one made before they're started isn't
        const changes = (started: boolean, connected: boolean) => ({
            started, connected,
            on() { return this; }, onConnect() { return this; }, onDisconnect() { return this; },
        }) as never;
        const make = (c: never) => new ListController<Item>({ name: 'Items', list: pages([]), same: () => true, changes: c });
        assert.equal(make(changes(true, false)).current.unavailable, true);
        assert.equal(make(changes(true, true)).current.unavailable, false);
        assert.equal(make(changes(false, false)).current.unavailable, false);
    });

    await t.test('error', async () => {
        // A page which can't be loaded keeps what's shown, and says why
        const errors: string[] = [];
        const list = new ListController<Item>({
            name: 'Roles',
            list: async () => { throw new Error('refused'); },
            same: (a, b) => a.name === b.name,
            notifications: { error: (title: string) => { errors.push(title); return 0; } } as never,
        });
        await list.load();
        assert.deepEqual(list.current.items, []);
        assert.deepEqual(errors, ["Roles couldn't be loaded"]);
    });
});
