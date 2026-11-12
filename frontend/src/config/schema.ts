// The server's configuration, which mirrors go-server's config package

// Config is a server's configuration, from /config.json
export interface Config {
    // The path prefix of the API, such as /api
    prefix: string;

    // The name of the server, such as its command
    name?: string;

    // The version of the server
    version?: string;

    // The path of the API's documentation, when the server serves it
    openapi?: string;

    // The server's public endpoint, such as https://example.com/api, when
    // it was given its public name, such as that of a proxy it's behind
    endpoint?: string;
}
