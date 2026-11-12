// ApiError is an error response of an API, with its HTTP status, its reason,
// and its detail, when it has any, as go-server's httpresponse writes them
export class ApiError extends Error {
    readonly status: number;
    readonly detail?: unknown;

    constructor(status: number, reason: string, detail?: unknown) {
        super(reason);
        this.name = 'ApiError';
        this.status = status;
        this.detail = detail;
    }
}

// errorResponse returns the ApiError of a response which isn't ok, from its
// JSON body, which is go-server's error, or else its status text
export async function errorResponse(response: Response): Promise<ApiError> {
    const error = await response.json().catch(() => undefined) as { reason?: string; detail?: unknown } | undefined;
    return new ApiError(response.status, error?.reason || response.statusText, error?.detail);
}
