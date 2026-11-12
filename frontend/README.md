# @mutablelogic/go-server

TypeScript clients of go-server's APIs, for the frontends of servers built on it.

- Its clients and controllers are plain TypeScript, with no framework, so they work with Lit, or anything else.
- Its views are Lit elements, with Carbon's web components. `lit` and `@carbon/web-components` are optional peer dependencies, which only an app which uses a view installs. Carbon's types need `skipLibCheck` in the app's `tsconfig.json`.
- It's published to GitHub Packages with each go-server release, at the release's version.

| Module | What it has |
|---|---|
| `@mutablelogic/go-server/api` | `requestStream`, which reads a stream of server-sent events, `ApiError`, and `Client`, which is where an API is |
| `@mutablelogic/go-server/broadcaster` | The broadcaster's `Change` schema, `watchChanges`, and `ChangesController`, which other controllers subscribe to |
| `@mutablelogic/go-server/notification` | `NotificationController`, which keeps notifications, such as errors, until they time out, or are closed, and tells the views which subscribe to it |
| `@mutablelogic/go-server/notification/view` | The `gs-notifications` element, which shows a controller's notifications as Carbon toasts |
| `@mutablelogic/go-server/panel` | `PanelController`, which keeps a panel's item, header and actions, such as Save and Cancel, which close it, and runs them |
| `@mutablelogic/go-server/panel/view` | The `gs-panel` element, which shows a controller's panel on the right of the page, with its header, its body, which is the element's content, and its actions |
| `@mutablelogic/go-server/config` | `getConfig`, which reads the server's `/config.json`, and `configClient`, which is the client of the API it says is at its prefix |

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

## Finding the API

A server with go-server's `config` package serves `/config.json`, which says where its API is, so a frontend doesn't assume its prefix:

```ts
import { configClient, getConfig } from '@mutablelogic/go-server/config';
import { ChangesController } from '@mutablelogic/go-server/broadcaster';

// { prefix: '/api', name: 'server', version: 'v1.7.21', openapi: '/api/openapi.html' }
const config = await getConfig();
const changes = new ChangesController(configClient(config));
```

`getConfig` reads `config.json` beside the page, without the cache, so a page sees a server which was updated. Its `endpoint` is the server's public URL, such as `https://example.com/api`, when it was given its public name.

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

## Notifications

A `NotificationController` keeps notifications until they time out, 8s by default, or are closed. The `gs-notifications` element shows them as toasts, stacked at the top right:

```ts
import { NotificationController } from '@mutablelogic/go-server/notification';
import { NotificationsTag } from '@mutablelogic/go-server/notification/view';

const notifications = new NotificationController();
const element = document.createElement(NotificationsTag);
element.controller = notifications;
document.body.append(element);

// Show the errors watching changes
changes.onError((error) => notifications.error("Changes couldn't be watched", error));

// A notification which stays until it's closed, with a link
notifications.push({ kind: 'info', title: 'A reminder fired', link: { href: '#chat/1', text: 'Open the chat' }, timeout: null });
```

- **Kinds:** `error`, `warning`, `info` or `success`, which set a toast's colour and icon.
- **Fields:** a `subtitle`, a `caption`, such as when it happened, and a `link`, which closes it when it's followed.
- **Closing:** closing a toast closes it in the controller. `push` returns an id, which `close` closes it by.
- **Position:** `--gs-notifications-top`, below a 3rem header by default, `--gs-notifications-right`, and `--gs-notifications-z-index`.
- **Other views:** `subscribe` calls a function with the notifications now, and whenever they change, until the function it returns is called.

## Panels

A `PanelController` keeps a panel: the item it shows, while it's open, its header, and the actions in its footer. The `gs-panel` element shows it on the right of the page, over the page, which is darkened, as a modal dialog. Its body is the element's content, which the app renders from the item.

```ts
import { PanelController } from '@mutablelogic/go-server/panel';
import { PanelTag } from '@mutablelogic/go-server/panel/view';

const panel = new PanelController<Agent>()
    .onClose((agent) => console.log('closed', agent.name))
    .onError((error, action) => notifications.error(`It couldn't ${action.label.toLowerCase()}`, error));

panel.setHeader({ title: 'Edit agent', subtitle: 'builtin:segment' });
panel.setActions([
    { label: 'Save', run: () => saveAgent(form) },
    { label: 'Cancel', kind: 'secondary' },
]);

const element = document.createElement(PanelTag);
element.controller = panel;
document.body.append(element);

// Render the agent into the element, and open the panel
render(html`<agent-form .agent=${agent}></agent-form>`, element);
panel.open(agent);
```

- **Header:** a `title`, a `subtitle`, an `icon`, which is a URL, such as of a logo, and a close button, when it's `closable`, which it is by default. It's left out when it has none of them. The element's `title` and `subtitle` slots take the place of the controller's text.
- **Actions:** each is a Carbon button, which runs its `run` when it's clicked, until what it returns settles, and then closes the panel. While it runs, it shows it's running, and the others are disabled.
  - An action without `run`, such as Cancel, closes the panel.
  - An action which fails leaves the panel open, and calls `onError`.
  - An action with `close: false` leaves it open.
- **Closing:** by its close button, the escape key, or clicking the page beside it, when it's closable. `onClose` is called with its item.
- **Focus:** the page beside it is inert while it's open, except for regions which announce, such as `gs-notifications`, and what was focused before is focused again when it closes.
- **Position:** `--gs-panel-top`, below a 3rem header by default, `--gs-panel-width`, 28rem by default, and `--gs-panel-z-index`.

## Developing

| Command | Description |
|---|---|
| `make frontend` | Installs from `package-lock.json`, type checks, and builds into `frontend/dist` |
| `make frontend-test` | Type checks, and runs the tests |
| `npm test` | Runs the tests, from the TypeScript, with Node's test runner |

The tests need Node 22.18 or later, which runs TypeScript.
