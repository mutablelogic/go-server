// A JSON request of an API, beside requestStream, which every app's API
// client needs

import { errorResponse } from './error.ts';
import { url, type Client, type StreamOptions } from './stream.ts';

// RequestOptions are the request's query, its body, which is sent as JSON,
// and its signal, which aborts it
export interface RequestOptions {
    query?: StreamOptions['query'];
    body?: unknown;
    signal?: AbortSignal;
}

// request requests the path, beneath the client's base, and returns the
// response, parsed as JSON, or undefined when it has no body. Throws an
// ApiError when the response isn't ok.
export async function request<T>(client: Client, method: string, path: string, options: RequestOptions = {}): Promise<T> {
    const headers: Record<string, string> = { Accept: 'application/json' };
    if (options.body !== undefined) {
        headers['Content-Type'] = 'application/json';
    }
    const response = await (client.fetch ?? fetch)(url(client.base, path, options.query), {
        method,
        headers,
        body: options.body === undefined ? undefined : JSON.stringify(options.body),
        signal: options.signal,
    });
    if (!response.ok) {
        throw await errorResponse(response);
    }
    const text = await response.text();
    return (text ? JSON.parse(text) : undefined) as T;
}
