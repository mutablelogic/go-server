// The server's configuration, which mirrors go-server's config package

// Config is a server's configuration, from /config.json
export interface Config {
    // The path prefix of the API, such as /api
    prefix: string;

    // The name of the server, such as its command
    name?: string;

    // The version of the server
    version?: string;

    // When the server was built, such as 2026-11-12T18:00:00Z, so a server
    // which was rebuilt is known, even with the same version
    build?: string;

    // The path of the API's documentation, when the server serves it
    openapi?: string;

    // The server's public endpoint, such as https://example.com/api, when
    // it was given its public name, such as that of a proxy it's behind
    endpoint?: string;
}
