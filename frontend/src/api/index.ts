// Clients of go-server's APIs: their errors, their JSON requests, and
// their streams of server-sent events

export { ApiError, errorResponse } from './error.ts';
export { request, type RequestOptions } from './request.ts';
export { requestStream, url, type Client, type StreamOptions } from './stream.ts';
