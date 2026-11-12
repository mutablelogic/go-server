// The broadcaster's HTTP client, of its API, which streams its changes at
// changes

import { requestStream, type Client } from '../api/stream.ts';
import { ChangeName, ChangeShutdown, type Change, type ChangeFilter } from './schema.ts';

// WatchOptions are the options of watching the changes
export interface WatchOptions {
    // filter is which changes are streamed, which is all of them, when it
    // isn't set
    filter?: ChangeFilter;

    // signal stops watching
    signal?: AbortSignal;

    // onOpen is called once the stream has started, which is once the
    // server is watching the changes, so what's loaded then is up to date
    onOpen?: () => void;

    // onShutdown is called as soon as the server says it's stopping
    onShutdown?: () => void;
}

// watchChanges calls the function with each change which matches the
// filter, as it happens, until the stream ends, or the signal aborts it, and
// returns true when the server said it's stopping, even when the stream then
// fails, as it can, as the server closes it. Throws an ApiError when the
// changes can't be watched, such as with a 503 when the server is stopping.
export async function watchChanges(client: Client, fn: (change: Change) => void, options: WatchOptions = {}): Promise<boolean> {
    let shutdown = false;
    try {
        await requestStream(client, 'changes', (name, data) => {
            if (name === ChangeName) {
                fn(data as Change);
            } else if (name === ChangeShutdown && !shutdown) {
                shutdown = true;
                options.onShutdown?.();
            }
        }, { query: { ...options.filter }, signal: options.signal, onOpen: options.onOpen });
    } catch (error) {
        if (!shutdown) {
            throw error;
        }
    }
    return shutdown;
}
