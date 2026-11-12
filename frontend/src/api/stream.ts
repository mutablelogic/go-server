import { errorResponse, ApiError } from './error.ts';

// Client is where an API is, and how it's requested
export interface Client {
    // base is the API's prefix, such as /api, or its URL, such as
    // http://localhost:8084/api
    base: string;

    // fetch requests the API, which is the global fetch, when it isn't set
    fetch?: typeof fetch;
}

// StreamOptions are the options of a request of a stream of server-sent events
export interface StreamOptions {
    // query is the request's query, whose undefined values are left out
    query?: Record<string, string | number | boolean | undefined>;

    // signal aborts the request, and ends the stream
    signal?: AbortSignal;

    // onOpen is called once the stream has started
    onOpen?: () => void;
}

// requestStream requests the path, beneath the client's base, which accepts
// server-sent events, and calls the function with the name and data, parsed
// as JSON, of each event, as it's received, until the stream ends. An event
// without data, such as a ping, isn't passed to the function. Throws an
// ApiError when the response isn't ok.
export async function requestStream(client: Client, path: string, fn: (event: string, data: unknown) => void, options: StreamOptions = {}): Promise<void> {
    const response = await (client.fetch ?? fetch)(url(client.base, path, options.query), {
        method: 'GET',
        headers: { Accept: 'text/event-stream' },
        signal: options.signal,
    });
    if (!response.ok) {
        throw await errorResponse(response);
    } else if (!response.body) {
        throw new ApiError(response.status, 'missing response body');
    }
    options.onOpen?.();

    // Events are separated by blank lines, and each is lines of fields, of
    // which event is its name, and data its data, which is joined by newlines
    // when there's more than one
    const reader = response.body.pipeThrough(new TextDecoderStream()).getReader();
    let buffer = '';
    for (;;) {
        const { value, done } = await reader.read();
        if (done) {
            return;
        }
        buffer += value.replace(/\r\n?/g, '\n');
        let end: number;
        while ((end = buffer.indexOf('\n\n')) >= 0) {
            const block = buffer.slice(0, end);
            buffer = buffer.slice(end + 2);
            let event = 'message';
            const data: string[] = [];
            for (const line of block.split('\n')) {
                const colon = line.indexOf(':');
                const field = colon < 0 ? line : line.slice(0, colon);
                const text = colon < 0 ? '' : line.slice(colon + 1).replace(/^ /, '');
                if (field === 'event') {
                    event = text;
                } else if (field === 'data') {
                    data.push(text);
                }
            }
            if (data.length > 0) {
                fn(event, JSON.parse(data.join('\n')));
            }
        }
    }
}

// url returns the URL of the path, beneath the base, with the query
export function url(base: string, path: string, query: StreamOptions['query'] = {}): string {
    const params = new URLSearchParams();
    for (const [key, value] of Object.entries(query)) {
        if (value !== undefined && value !== '') {
            params.set(key, String(value));
        }
    }
    const search = params.toString();
    return base.replace(/\/+$/, '') + '/' + path.replace(/^\/+/, '') + (search ? '?' + search : '');
}
