// A fake fetch, which responds to each request with the next of its
// responses, and records the requests

// FakeResponse is a response of the fake fetch: a stream of server-sent
// events, as chunks, and chunks which are sent later, after a delay, which
// stays open until the request is aborted when open is true, or an error
// status, with its body
export interface FakeResponse {
    chunks?: string[];
    later?: { delay: number; chunks: string[] };
    open?: boolean;
    status?: number;
    body?: unknown;
}

export interface FakeFetch {
    fetch: typeof fetch;
    requests: { url: string; accept: string | null }[];
}

export function fakeFetch(...responses: FakeResponse[]): FakeFetch {
    const requests: FakeFetch['requests'] = [];
    const fetch = async (input: RequestInfo | URL, init?: RequestInit): Promise<Response> => {
        requests.push({ url: String(input), accept: new Headers(init?.headers).get('Accept') });
        const response = responses.shift() ?? { status: 503, body: { reason: 'no more responses' } };
        if (response.status && response.status >= 300) {
            return new Response(JSON.stringify(response.body ?? {}), { status: response.status, headers: { 'Content-Type': 'application/json' } });
        }
        const signal = init?.signal;
        const body = new ReadableStream<Uint8Array>({
            start(controller) {
                const encoder = new TextEncoder();
                for (const chunk of response.chunks ?? []) {
                    controller.enqueue(encoder.encode(chunk));
                }
                if (response.later) {
                    const { delay, chunks } = response.later;
                    setTimeout(() => {
                        try {
                            chunks.forEach((chunk) => controller.enqueue(encoder.encode(chunk)));
                        } catch {
                            // The stream has ended
                        }
                    }, delay);
                }
                if (!response.open) {
                    controller.close();
                } else {
                    signal?.addEventListener('abort', () => controller.error(signal.reason), { once: true });
                }
            },
        });
        return new Response(body, { status: 200, headers: { 'Content-Type': 'text/event-stream' } });
    };
    return { fetch: fetch as typeof globalThis.fetch, requests };
}

// sse returns a server-sent event, with its name, and its data as JSON
export function sse(name: string, data: unknown): string {
    return `event: ${name}\ndata: ${JSON.stringify(data)}\n\n`;
}

// waitFor waits until the condition is true, or fails after the timeout
export async function waitFor(condition: () => boolean, timeout = 2000): Promise<void> {
    const end = Date.now() + timeout;
    while (!condition()) {
        if (Date.now() > end) {
            throw new Error('timed out');
        }
        await new Promise((resolve) => setTimeout(resolve, 5));
    }
}
