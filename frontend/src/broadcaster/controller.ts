// The changes controller watches the broadcaster's changes, and calls the
// listeners whose filters each change matches, so they load what changed.
//
// - Every listener is called once the stream has started, which is once the
//   server is watching, as changes may have been made before then, when
//   it's first watched, or while it wasn't.
// - When the server stops, or the stream fails, the disconnect functions are
//   called, as what's shown may no longer be so, and it's watched again,
//   after a delay which grows while it fails, until it starts. A stream which
//   ends otherwise, such as at the server's timeout, is watched again after
//   the shortest delay.
// - A listener is called at most once in an interval, after the first change
//   in it, with the last, as a burst of changes is one change to load.

import type { Client } from '../api/stream.ts';
import { watchChanges } from './httpclient.ts';
import { changeMatches, type Change, type ChangeFilter } from './schema.ts';

// ChangeListener is called with a change which matches its filters, or with
// no change when changes may have been missed, such as when the stream has
// started, again
export type ChangeListener = (change?: Change) => void;

// ChangesOptions are the client of the broadcaster's API, and the controller's
// timing, in milliseconds
export interface ChangesOptions extends Client {
    // minDelay is the delay before watching again, which doubles while it
    // fails, up to maxDelay
    minDelay?: number;
    maxDelay?: number;

    // interval is the interval in which a listener is called at most once
    interval?: number;
}

export class ChangesController {
    private readonly client: Client;
    private readonly minDelay: number;
    private readonly maxDelay: number;
    private readonly interval: number;
    private readonly listeners = new Set<Listener>();
    private readonly connects: (() => void)[] = [];
    private readonly disconnects: (() => void)[] = [];
    private readonly shutdowns: (() => void)[] = [];
    private readonly errors: ((error: unknown) => void)[] = [];

    // The stream which is watched, which is aborted when it's stopped
    private request?: AbortController;
    private isConnected = false;

    constructor(options: ChangesOptions) {
        this.client = { base: options.base, fetch: options.fetch };
        this.minDelay = options.minDelay ?? 1000;
        this.maxDelay = options.maxDelay ?? 30000;
        this.interval = options.interval ?? 500;
    }

    // connected is true while the stream is watched
    get connected(): boolean {
        return this.isConnected;
    }

    // subscribe calls the function with each change which matches the
    // filter, or any of the filters, or with no change when changes may have
    // been missed, until the function it returns is called, which
    // unsubscribes. It's called once for a burst of changes which match any
    // of them, such as when a change is to more than one table.
    subscribe(filter: ChangeFilter | ChangeFilter[], fn: ChangeListener): () => void {
        const listener = new Listener([filter].flat(), fn, this.interval);
        this.listeners.add(listener);
        return () => {
            listener.cancel();
            this.listeners.delete(listener);
        };
    }

    // on subscribes the function to the changes which match the filter, or
    // any of the filters, as subscribe does, until the controller is gone
    on(filter: ChangeFilter | ChangeFilter[], fn: ChangeListener): this {
        this.subscribe(filter, fn);
        return this;
    }

    // onConnect calls the function whenever the stream starts, as the server
    // can be reached again, before the listeners are called
    onConnect(fn: () => void): this {
        this.connects.push(fn);
        return this;
    }

    // onDisconnect calls the function when the server stops, or the stream
    // fails, as the server can't be reached until it's watched again
    onDisconnect(fn: () => void): this {
        this.disconnects.push(fn);
        return this;
    }

    // onShutdown calls the function when the server says it's stopping, as
    // it's told before the stream ends, so requests which would hold it up,
    // such as a reply which is streamed, can be stopped
    onShutdown(fn: () => void): this {
        this.shutdowns.push(fn);
        return this;
    }

    // onError calls the function with the error when the changes can't be
    // watched, once until it's watched again
    onError(fn: (error: unknown) => void): this {
        this.errors.push(fn);
        return this;
    }

    // start watches the changes, until it's stopped
    start() {
        this.stop();
        const request = new AbortController();
        this.request = request;
        void this.watch(request.signal);
    }

    // stop stops watching the changes
    stop() {
        this.request?.abort();
        this.request = undefined;
        this.isConnected = false;
    }

    // watch watches the changes, and watches them again when the stream ends,
    // until the signal aborts it
    private async watch(signal: AbortSignal) {
        let delay = this.minDelay;
        let failed = false;

        // Whether the disconnect functions have been called since the stream
        // last started, so they're called once, when the server stops, or the
        // stream fails, whether or not it had started
        let disconnected = false;
        const disconnect = () => {
            this.isConnected = false;
            if (!disconnected) {
                disconnected = true;
                this.disconnects.forEach((fn) => fn());
            }
        };

        while (!signal.aborted) {
            // Once the stream has started, every listener is called, as
            // changes may have been made before the server was watching, and
            // it's no longer failing
            const onOpen = () => {
                delay = this.minDelay;
                failed = false;
                disconnected = false;
                this.isConnected = true;
                this.connects.forEach((fn) => fn());
                this.listeners.forEach((listener) => listener.call());
            };
            const onShutdown = () => this.shutdowns.forEach((fn) => fn());
            try {
                const shutdown = await watchChanges(this.client, (change) => this.dispatch(change), { signal, onOpen, onShutdown });
                if (signal.aborted) {
                    return;
                }

                // A server which is stopping can't be reached until it's
                // back, but a stream which ended otherwise, such as at the
                // server's timeout, is watched again at once
                this.isConnected = false;
                if (shutdown) {
                    disconnect();
                    failed = true;
                }
            } catch (error) {
                if (signal.aborted) {
                    return;
                }
                disconnect();
                if (!failed) {
                    this.errors.forEach((fn) => fn(error));
                    failed = true;
                }
                delay = Math.min(delay * 2, this.maxDelay);
            }
            await sleep(delay, signal);
        }
    }

    // dispatch calls the listeners whose filters the change matches
    private dispatch(change: Change) {
        for (const listener of this.listeners) {
            if (listener.filters.some((filter) => changeMatches(filter, change))) {
                listener.call(change);
            }
        }
    }
}

// Listener calls its function at most once in the interval, after the first
// call in it, with the last change, or no change when any call had none
class Listener {
    readonly filters: ChangeFilter[];
    private readonly fn: ChangeListener;
    private readonly interval: number;
    private timer?: ReturnType<typeof setTimeout>;
    private change?: Change;
    private missed = false;

    constructor(filters: ChangeFilter[], fn: ChangeListener, interval: number) {
        this.filters = filters;
        this.fn = fn;
        this.interval = interval;
    }

    call(change?: Change) {
        this.change = change;
        this.missed ||= change === undefined;
        this.timer ??= setTimeout(() => {
            const change = this.missed ? undefined : this.change;
            this.timer = undefined;
            this.change = undefined;
            this.missed = false;
            this.fn(change);
        }, this.interval);
    }

    cancel() {
        clearTimeout(this.timer);
        this.timer = undefined;
    }
}

// sleep waits for the delay, or until the signal aborts
function sleep(delay: number, signal: AbortSignal): Promise<void> {
    return new Promise((resolve) => {
        const timer = setTimeout(done, delay);
        signal.addEventListener('abort', done, { once: true });
        function done() {
            clearTimeout(timer);
            signal.removeEventListener('abort', done);
            resolve();
        }
    });
}
