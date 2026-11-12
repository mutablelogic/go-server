// Clients of go-server's APIs: their errors, and their streams of
// server-sent events

export { ApiError, errorResponse } from './error.ts';
export { requestStream, url, type Client, type StreamOptions } from './stream.ts';
