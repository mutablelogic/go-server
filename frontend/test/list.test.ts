import { test } from 'node:test';
import assert from 'node:assert/strict';
import { ApiError } from '../src/api/index.ts';
import { ListController, type ListState } from '../src/list/index.ts';

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
