// The HTTP client of the server's configuration, which its frontend reads,
// to know where its API is

import { errorResponse } from '../api/error.ts';
import type { Client } from '../api/stream.ts';
import type { Config } from './schema.ts';

// ConfigOptions are where the configuration is, and how it's requested
export interface ConfigOptions {
    // url is where the configuration is, which is config.json, beside the
    // page, when it isn't set
    url?: string;

    // fetch requests the configuration, which is the global fetch, when it
    // isn't set
    fetch?: typeof fetch;
}

// getConfig returns the server's configuration, as it is now, which isn't
// cached, as the server can change, such as when it's updated. Throws an
// ApiError when it can't be read.
export async function getConfig(options: ConfigOptions = {}): Promise<Config> {
    const response = await (options.fetch ?? fetch)(options.url ?? 'config.json', {
        headers: { Accept: 'application/json' },
        cache: 'no-store',
    });
    if (!response.ok) {
        throw await errorResponse(response);
    }
    return await response.json() as Config;
}

// configClient returns the client of the API which the configuration says
// is at its prefix, beside the page, with the fetch
export function configClient(config: Config, fetch?: typeof globalThis.fetch): Client {
    return { base: config.prefix, fetch };
}
