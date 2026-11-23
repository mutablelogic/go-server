// The list controller keeps a list which an API returns a page at a time,
// whose response is a page of its items, in body, with their count. It
// shows a page at a time, or, when it scrolls, the pages which are loaded,
// one after another, as more are loaded, such as when a table is scrolled
// to its end. It loads the page, sorted, when it's sorted, by the server, again when the
// changes it follows say what's in it changed, keeping what's shown until
// it's loaded, keeps the item which is opened current, and tells the views
// which subscribe to it, such as gs-table and gs-tiles, whenever it changes.
// Errors are said by the notification controller.

import { ApiError } from '../api/index.ts';
import type { ChangeFilter, ChangesController } from '../broadcaster/index.ts';
import type { NotificationController } from '../notification/index.ts';

// Page is a page of a list, with the number of items in the list, which
// a list which scrolls needn't know, and, of one which scrolls, whether
// there are more after it, which is otherwise of its count, or, without
// one, of whether the page is full
export interface Page<T> {
    count?: number;
    body?: T[];
    more?: boolean;
}

// Sort is what a list is sorted by, such as a column's key, and its
// direction
export interface Sort {
    key: string;
    direction: 'ascending' | 'descending';
}

// ListState is what the views show: the items on the page, or, of a list
// which scrolls, those which are loaded, which are undefined while they're
// first loading, the page, from 1, its size, the number of items, which is
// those which are loaded, of a list which scrolls, whose count isn't known,
// what they're sorted by, when they are, the item which is opened, and
// whether the server can't be reached. A list which scrolls also says
// whether there are more, whether more are loading, and whether they
// couldn't be.
export interface ListState<T> {
    items?: readonly T[];
    page: number;
    pageSize: number;
    total: number;
    sort?: Sort;
    selected?: T;
    unavailable: boolean;
    more?: boolean;
    loading?: boolean;
    failed?: boolean;
}

// ListOptions are how the list is loaded, and followed
export interface ListOptions<T> {
    // name is the name of the items, such as Roles, which says they can't be
    // loaded
    name: string;

    // list returns the page of the items, at the offset, of at most the
    // limit, sorted, when there's a sort, and after the item before it, of
    // a list which scrolls, so an API of a cursor, rather than an offset,
    // can start after it
    list: (offset: number, limit: number, sort?: Sort, after?: T) => Promise<Page<T>>;

    // get returns an item again, by itself, when it isn't on the page
    get?: (item: T) => Promise<T>;

    // same returns true when the items are the same, such as by their names
    same: (a: T, b: T) => boolean;

    // notifications say why the items can't be loaded
    notifications?: NotificationController;

    // changes, and the filter of the changes to the items, which load them
    // again, and say when the server can't be reached
    changes?: ChangesController;
    filter?: ChangeFilter | ChangeFilter[];

    // pageSize is the size of a page, which is 25 by default
    pageSize?: number;

    // sort is what the items are sorted by, until another is set
    sort?: Sort;

    // scroll is true of a list which scrolls, whose pages are shown one
    // after another, as more are loaded (see loadMore), rather than a page
    // at a time
    scroll?: boolean;
}

export class ListController<T> {
    private readonly options: ListOptions<T>;
    private readonly listeners = new Set<(state: ListState<T>) => void>();
    private state: ListState<T>;

    // The refresh in progress, and whether to refresh again once it's done,
    // as the items changed while it loaded them
    private refreshing?: Promise<void>;
    private again = false;

    // The list's generation, which changes when what's listed does, such as
    // its page, or its sort, so a response to a request of an earlier one
    // is ignored, and the list is loaded again
    private generation = 0;

    constructor(options: ListOptions<T>) {
        this.options = options;
        // It's unavailable from the start when the changes have started, and
        // can't be watched, as it's told when they can be, but not again that
        // they can't be
        const unavailable = options.changes ? options.changes.started && !options.changes.connected : false;
        this.state = { page: 1, pageSize: options.pageSize ?? 25, total: 0, sort: options.sort, unavailable, ...(options.scroll ? { more: false, loading: false, failed: false } : {}) };
        if (options.changes) {
            options.changes.on(options.filter ?? {}, () => void this.refresh());
            options.changes.onDisconnect(() => this.update({ unavailable: true }));
            options.changes.onConnect(() => this.update({ unavailable: false }));
        }
    }

    // current is the list's state
    get current(): ListState<T> {
        return this.state;
    }

    // subscribe calls the function with the state, now, and whenever it
    // changes, until the function it returns is called, which unsubscribes
    subscribe(fn: (state: ListState<T>) => void): () => void {
        this.listeners.add(fn);
        fn(this.state);
        return () => this.listeners.delete(fn);
    }

    // load shows the views loading, and loads the page
    load(): Promise<void> {
        this.generation++;
        this.update({ items: undefined });
        return this.refresh();
    }

    // reset loads the first page, showing the views loading, and closes the
    // item which is opened, as what's listed is another list, such as the
    // objects of another database
    reset(): Promise<void> {
        this.generation++;
        this.update({ items: undefined, page: 1, total: 0, selected: undefined });
        return this.refresh();
    }

    // refresh loads the page again, keeping the items which are shown until
    // it's loaded. A page beyond the last, as items were deleted, is the
    // last. While a refresh is in progress, it's refreshed once more after.
    // A response to an earlier generation, or with the item which is opened
    // changed while it was loaded, isn't shown, as it's out of date.
    refresh(): Promise<void> {
        if (this.refreshing) {
            this.again = true;
            return this.refreshing;
        }
        this.refreshing = (async () => {
            do {
                this.again = false;
                const { page, pageSize, sort } = this.state;
                const generation = this.generation;
                if (this.options.scroll) {
                    await this.refreshScroll(generation);
                    continue;
                }
                try {
                    const list = await this.options.list((page - 1) * pageSize, pageSize, sort);
                    const count = list.count ?? list.body?.length ?? 0;
                    if (generation !== this.generation) {
                        this.again = true;
                        continue;
                    }
                    const last = Math.max(1, Math.ceil(count / pageSize));
                    if (page > last) {
                        this.update({ page: last });
                        this.again = true;
                        continue;
                    }
                    const items = list.body ?? [];
                    const selected = this.state.selected;
                    const reloaded = await this.reloadSelected(items);
                    if (generation !== this.generation) {
                        this.again = true;
                        continue;
                    }
                    this.update({ items, total: count, selected: this.state.selected === selected ? reloaded : this.state.selected });
                } catch (error) {
                    if (generation !== this.generation) {
                        this.again = true;
                        continue;
                    }
                    this.update({ items: this.state.items ?? [] });
                    this.options.notifications?.error(`${this.options.name} couldn't be loaded`, error);
                }
            } while (this.again);
            this.refreshing = undefined;
        })();
        return this.refreshing;
    }

    // loadMore loads the page after those which are loaded, of a list which
    // scrolls, unless there are no more, or they're loading, or the list is
    // loaded again, and adds it to them. When it can't be loaded, the list
    // says it failed, until it's loaded again, or more are.
    loadMore(): Promise<void> {
        const { items, more, loading, pageSize, sort } = this.state;
        if (!this.options.scroll || !items || !more || loading || this.refreshing) {
            return this.refreshing ?? Promise.resolve();
        }
        const generation = this.generation;
        this.update({ loading: true, failed: false });
        return (async () => {
            try {
                const list = await this.options.list(items.length, pageSize, sort, items.at(-1));
                if (generation !== this.generation || this.state.items !== items) {
                    return;
                }
                const body = list.body ?? [];
                const loaded = [...items, ...body];
                this.update({ items: loaded, total: list.count ?? loaded.length, more: hasMore(list, items.length, pageSize), loading: false });
            } catch {
                if (generation === this.generation) {
                    this.update({ loading: false, failed: true });
                }
            }
        })();
    }

    // setPage loads the page, from 1
    setPage(page: number) {
        this.generation++;
        this.update({ page });
        void this.refresh();
    }

    // setSort loads the first page, sorted by the sort, or unsorted, which,
    // of a list which scrolls, is loading, as what's loaded is of another
    // order
    setSort(sort?: Sort) {
        this.generation++;
        this.update(this.options.scroll ? { page: 1, sort, items: undefined } : { page: 1, sort });
        void this.refresh();
    }

    // setPageSize loads the first page of the size
    setPageSize(pageSize: number) {
        this.generation++;
        this.update({ page: 1, pageSize });
        void this.refresh();
    }

    // open selects the item, such as to show it in a panel
    open(item: T) {
        this.update({ selected: item });
    }

    // close unselects the item
    close() {
        if (this.state.selected !== undefined) {
            this.update({ selected: undefined });
        }
    }

    // refreshScroll loads a list which scrolls again, from its start, as
    // many as are loaded, or a page, of the generation, a page at a time,
    // keeping those which are shown until it's done, while it's loading.
    // When none can be loaded, it failed, and there are more, which
    // loadMore tries again.
    private async refreshScroll(generation: number) {
        const { pageSize, sort } = this.state;
        const want = Math.max(this.state.items?.length ?? 0, pageSize);
        if (this.state.items) {
            this.update({ loading: true });
        }
        const items: T[] = [];
        let more = true;
        let count: number | undefined;
        try {
            while (more && items.length < want) {
                const list = await this.options.list(items.length, pageSize, sort, items.at(-1));
                if (generation !== this.generation) {
                    this.again = true;
                    return;
                }
                const offset = items.length;
                items.push(...(list.body ?? []));
                count = list.count;
                more = hasMore(list, offset, pageSize) && (list.body?.length ?? 0) > 0;
            }
            const selected = this.state.selected;
            const reloaded = await this.reloadSelected(items);
            if (generation !== this.generation) {
                this.again = true;
                return;
            }
            this.update({ items, total: count ?? items.length, more, loading: false, failed: false, selected: this.state.selected === selected ? reloaded : this.state.selected });
        } catch (error) {
            if (generation !== this.generation) {
                this.again = true;
                return;
            }
            const items = this.state.items ?? [];
            this.update(items.length ? { loading: false } : { items, loading: false, more: true, failed: true });
            this.options.notifications?.error(`${this.options.name} couldn't be loaded`, error);
        }
    }

    // reloadSelected returns the selected item again, from the page, when
    // it's on it, or else by itself, or undefined, when it's been deleted
    private async reloadSelected(items: readonly T[]): Promise<T | undefined> {
        const selected = this.state.selected;
        if (selected === undefined) {
            return undefined;
        }
        const item = items.find((item) => this.options.same(item, selected));
        if (item || !this.options.get) {
            return item ?? selected;
        }
        try {
            return await this.options.get(selected);
        } catch (error) {
            if (error instanceof ApiError && error.status === 404) {
                return undefined;
            }
            throw error;
        }
    }

    private update(state: Partial<ListState<T>>) {
        this.state = { ...this.state, ...state };
        this.listeners.forEach((fn) => fn(this.state));
    }
}

// hasMore returns whether there are more items after a page, at the offset:
// what it says, or else of the count, or, without one, whether it's full
function hasMore<T>(page: Page<T>, offset: number, limit: number): boolean {
    if (page.more !== undefined) {
        return page.more;
    } else if (page.count !== undefined) {
        return offset + (page.body?.length ?? 0) < page.count;
    }
    return (page.body?.length ?? 0) >= limit;
}
