# @mutablelogic/go-server

TypeScript clients of go-server's APIs, for the frontends of servers built on it.

- It's plain TypeScript, with no framework, so it works with Lit, or anything else.
- It's published to GitHub Packages with each go-server release, at the release's version.

| Module | What it has |
|---|---|
| `@mutablelogic/go-server/api` | `requestStream`, which reads a stream of server-sent events, `ApiError`, and `Client`, which is where an API is |
| `@mutablelogic/go-server/broadcaster` | The broadcaster's `Change` schema, `watchChanges`, and `ChangesController`, which other controllers subscribe to |

## Installing

The package is on GitHub Packages, which needs a token with `read:packages` to install, even though it's public.

1. Add an `.npmrc` beside your `package.json`:

    ```ini
    @mutablelogic:registry=https://npm.pkg.github.com
    //npm.pkg.github.com/:_authToken=${GITHUB_TOKEN}
    ```

2. Install it:

    ```sh
    npm install @mutablelogic/go-server
    ```

Locally, `GITHUB_TOKEN` can be `$(gh auth token)`. In GitHub Actions, it's the workflow's `GITHUB_TOKEN`, with `packages: read`. A Docker build which runs `npm ci` needs it as a build secret.

## Watching changes

A `ChangesController` watches the server's `/changes`, and calls the listeners whose filters each change matches, so they load what changed.

```ts
import { ChangesController } from '@mutablelogic/go-server/broadcaster';

const changes = new ChangesController({ base: '/api' })
    .onConnect(() => console.log('connected'))
    .onDisconnect(() => console.log('the server can\'t be reached'))
    .onShutdown(() => console.log('the server is stopping'))
    .onError((error) => console.error('changes can\'t be watched', error));

// Load the sessions when one changes
changes.on({ schema: 'llm', table: 'session' }, (change) => void sessions.refresh());

changes.start();
```

- **The first call:** every listener is called, with no change, once the stream starts, as changes may have been missed before then.
- **Bursts:** a listener is called at most once in its interval, 500ms by default, with the last change of a burst.
- **Reconnecting:** when the server stops, or the stream fails, it's watched again, after a delay which doubles while it fails, from 1s to 30s.
- **Errors:** `onError` is called once, until the stream starts again.

## Subscribing from another controller

`subscribe` returns a function which unsubscribes, so a controller can follow the changes while it's in use, such as a Lit reactive controller while its host is connected:

```ts
import type { ReactiveController, ReactiveControllerHost } from 'lit';
import type { ChangesController } from '@mutablelogic/go-server/broadcaster';

export class SessionsController implements ReactiveController {
    private unsubscribe?: () => void;

    constructor(private host: ReactiveControllerHost, private changes: ChangesController) {
        host.addController(this);
    }

    hostConnected() {
        // Load the sessions, and again whenever one changes, or changes
        // may have been missed
        this.unsubscribe = this.changes.subscribe([{ table: 'session' }, { table: 'message' }], () => void this.load());
    }

    hostDisconnected() {
        this.unsubscribe?.();
    }

    private async load() { /* ... */ }
}
```

## Watching once

`watchChanges` watches the stream until it ends, without reconnecting. It returns `true` when the server said it's stopping.

```ts
import { watchChanges } from '@mutablelogic/go-server/broadcaster';

const abort = new AbortController();
const stopping = await watchChanges({ base: '/api' }, (change) => console.log(change), {
    filter: { table: 'session', action: 'update' },
    signal: abort.signal,
});
```

`filter` is sent as the stream's query, so the server only sends the changes which match it.

## Developing

| Command | Description |
|---|---|
| `make frontend` | Installs from `package-lock.json`, type checks, and builds into `frontend/dist` |
| `make frontend-test` | Type checks, and runs the tests |
| `npm test` | Runs the tests, from the TypeScript, with Node's test runner |

The tests need Node 22.18 or later, which runs TypeScript.
