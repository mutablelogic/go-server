// The list controller keeps a list which an API returns a page at a time,
// whose response is a page of its items, in body, with their count. It
// loads the page, sorted, when it's sorted, by the server, again when the
// changes it follows say what's in it changed, keeping what's shown until
// it's loaded, keeps the item which is opened current, and tells the views
// which subscribe to it, such as gs-table and gs-tiles, whenever it changes.
// Errors are said by the notification controller.

import { ApiError } from '../api/index.ts';
import type { ChangeFilter, ChangesController } from '../broadcaster/index.ts';
import type { NotificationController } from '../notification/index.ts';

// Page is a page of a list, with the number of items in the list
export interface Page<T> {
    count: number;
    body?: T[];
}

// Sort is what a list is sorted by, such as a column's key, and its
// direction
export interface Sort {
    key: string;
    direction: 'ascending' | 'descending';
}

// ListState is what the views show: the items on the page, which are
// undefined while they're first loading, the page, from 1, its size, the
// number of items, what they're sorted by, when they are, the item which is
// opened, and whether the server can't be reached
export interface ListState<T> {
    items?: readonly T[];
    page: number;
    pageSize: number;
    total: number;
    sort?: Sort;
    selected?: T;
    unavailable: boolean;
}

// ListOptions are how the list is loaded, and followed
export interface ListOptions<T> {
    // name is the name of the items, such as Roles, which says they can't be
    // loaded
    name: string;

    // list returns the page of the items, at the offset, of at most the
    // limit, sorted, when there's a sort
    list: (offset: number, limit: number, sort?: Sort) => Promise<Page<T>>;

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
}

export class ListController<T> {
    private readonly options: ListOptions<T>;
    private readonly listeners = new Set<(state: ListState<T>) => void>();
    private state: ListState<T>;

    // The refresh in progress, and whether to refresh again once it's done,
    // as the items changed while it loaded them
    private refreshing?: Promise<void>;
    private again = false;

    constructor(options: ListOptions<T>) {
        this.options = options;
        this.state = { page: 1, pageSize: options.pageSize ?? 25, total: 0, sort: options.sort, unavailable: false };
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
        this.update({ items: undefined });
        return this.refresh();
    }

    // reset loads the first page, showing the views loading, and closes the
    // item which is opened, as what's listed is another list, such as the
    // objects of another database
    reset(): Promise<void> {
        this.update({ items: undefined, page: 1, total: 0, selected: undefined });
        return this.refresh();
    }

    // refresh loads the page again, keeping the items which are shown until
    // it's loaded. A page beyond the last, as items were deleted, is the
    // last. While a refresh is in progress, it's refreshed once more after.
    refresh(): Promise<void> {
        if (this.refreshing) {
            this.again = true;
            return this.refreshing;
        }
        this.refreshing = (async () => {
            do {
                this.again = false;
                const { page, pageSize, sort } = this.state;
                try {
                    const list = await this.options.list((page - 1) * pageSize, pageSize, sort);
                    const last = Math.max(1, Math.ceil(list.count / pageSize));
                    if (page > last) {
                        this.update({ page: last });
                        this.again = true;
                        continue;
                    }
                    const items = list.body ?? [];
                    this.update({ items, total: list.count, selected: await this.reloadSelected(items) });
                } catch (error) {
                    this.update({ items: this.state.items ?? [] });
                    this.options.notifications?.error(`${this.options.name} couldn't be loaded`, error);
                }
            } while (this.again);
            this.refreshing = undefined;
        })();
        return this.refreshing;
    }

    // setPage loads the page, from 1
    setPage(page: number) {
        this.update({ page });
        void this.refresh();
    }

    // setSort loads the first page, sorted by the sort, or unsorted
    setSort(sort?: Sort) {
        this.update({ page: 1, sort });
        void this.refresh();
    }

    // setPageSize loads the first page of the size
    setPageSize(pageSize: number) {
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
