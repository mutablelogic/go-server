# @mutablelogic/go-server

TypeScript clients of go-server's APIs, for the frontends of servers built on it.

- Its clients and controllers are plain TypeScript, with no framework, so they work with Lit, or anything else.
- Its views are Lit elements, with Carbon's web components. `lit` and `@carbon/web-components` are optional peer dependencies, which only an app which uses a view installs. Carbon's types need `skipLibCheck` in the app's `tsconfig.json`.
- It's published to GitHub Packages with each go-server release, at the release's version.

| Module | What it has |
|---|---|
| `@mutablelogic/go-server/api` | `request`, a JSON request, `requestStream`, which reads a stream of server-sent events, `ApiError`, and `Client`, which is where an API is |
| `@mutablelogic/go-server/broadcaster` | The broadcaster's `Change` schema, `watchChanges`, and `ChangesController`, which other controllers subscribe to |
| `@mutablelogic/go-server/notification` | `NotificationController`, which keeps notifications, such as errors, until they time out, or are closed, and tells the views which subscribe to it |
| `@mutablelogic/go-server/notification/view` | The `gs-notifications` element, which shows a controller's notifications as Carbon toasts |
| `@mutablelogic/go-server/panel` | `PanelController`, which keeps a panel's item, header and actions, such as Save and Cancel, which close it, and runs them |
| `@mutablelogic/go-server/panel/view` | The `gs-panel` element, which shows a controller's panel on the right of the page, with its header, its body, which is the element's content, and its actions |
| `@mutablelogic/go-server/config` | `getConfig`, which reads the server's `/config.json`, `configClient`, which is the client of the API it says is at its prefix, and `ConfigController`, which follows the server, as it goes down, comes back, and is upgraded |
| `@mutablelogic/go-server/config/view` | The `gs-server-status` element, for a sidebar, which shows the server's version, that it's down, or upgraded, and links to its documentation |
| `@mutablelogic/go-server/list` | `ListController`, which keeps a list an API returns a page at a time, sorted by the server, and follows its changes |
| `@mutablelogic/go-server/list/view` | The `gs-table` element, a Carbon data table of a list's page, with sortable columns and pagination, whose rows open their items |
| `@mutablelogic/go-server/list/tiles` | The `gs-tiles` element, a row of tiles of a list, after one for all of them, of which one is chosen |
| `@mutablelogic/go-server/router` | `Router`, which shows the page of the location's hash, such as `#database/shop` |

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

## Following the server

A `ConfigController` follows the server with the `ChangesController`. Its status is:

| Status | When |
|---|---|
| `loading` | until the configuration is read, unless it's given, as `config`, when the page has read it |
| `up` | once it's read, or the server is back, with the page's version |
| `down` | when the server says it's stopping, or the changes can't be watched |
| `upgraded` | when the server is back, with another version than the page's, or the same version built again, until the page is refreshed |

A server says when it was built, as `build`, from `pkg/version`'s `BuildDate`, which its Makefile sets with `-ldflags "-X github.com/mutablelogic/go-server/pkg/version.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)"`. Without it, a server rebuilt with the same version, such as while it's developed, isn't known to be upgraded.

With a `NotificationController`, it says when the server is down, until it's back, that it's back, and that it was upgraded, with a Refresh action, which reloads the page. `gs-server-status` shows the same, for a sidebar:

```ts
import { ChangesController } from '@mutablelogic/go-server/broadcaster';
import { ConfigController, configClient, getConfig } from '@mutablelogic/go-server/config';
import { ServerStatusTag } from '@mutablelogic/go-server/config/view';

const page = await getConfig();
const changes = new ChangesController(configClient(page));
const config = new ConfigController(changes, { config: page, notifications })
    .onUp(() => void reloadEverything())
    .onUpgrade((latest) => console.log('the server is now', latest.version));
changes.start();

const status = document.createElement(ServerStatusTag);
status.controller = config;
status.repository = 'https://github.com/mutablelogic/go-server';
document.querySelector('cds-side-nav')?.append(status);
```

- **The view:** the page's version, a warning while the server is down, and, when it was upgraded, a button which refreshes the page. It links to the OpenAPI documentation, when the server serves it, and to the `repository`, when it's set.
- **Callbacks:** `onDown`, `onUp` and `onUpgrade`, and `subscribe`, for other views.
- **Testing:** the controller takes anything with `onConnect`, `onDisconnect` and `onShutdown`, so a test can follow a fake server.

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
- **Fields:** a `subtitle`, a `caption`, such as when it happened, a `link`, and an `action`, such as `{ text: 'Refresh', run }`, each of which closes it when it's followed.
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

## Lists

A `ListController` keeps a list which an API returns a page at a time, as `{ count, body }`. It loads a page, sorted when it has a sort, and keeps what's shown until the next page is loaded. With a `ChangesController`, it loads the page again when a change matches its filter, and marks the list unavailable while the server can't be reached.

```ts
import { request } from '@mutablelogic/go-server/api';
import { ListController } from '@mutablelogic/go-server/list';
import { TableTag, type Column } from '@mutablelogic/go-server/list/view';

const roles = new ListController<Role>({
    name: 'Roles',
    list: (offset, limit, sort) => request(client, 'GET', 'role', { query: { offset, limit, sort: sort?.key } }),
    get: (role) => request(client, 'GET', `role/${role.name}`),
    same: (a, b) => a.name === b.name,
    notifications, changes, filter: { table: 'role' },
});

const columns: Column<Role>[] = [
    { header: 'Role', cell: (role) => role.name, sort: 'name', sortFirst: 'ascending' },
    { header: 'Connections', cell: (role) => role.connections, sort: 'connections' },
];
const table = document.createElement(TableTag);
table.controller = roles;
table.columns = columns;
table.empty = 'No roles';

// Show the role which is opened in a panel
roles.subscribe(({ selected }) => selected && panel.open(selected));
```

- **Pages:** `setPage`, `setPageSize`, and `refresh`, which loads the page again. A page beyond the last, as items were deleted, becomes the last.
- **Sorting:** `setSort({ key, direction })` loads the first page, and `list` gets the sort. A column with a `sort` key sorts the list when its header is clicked, descending first unless its `sortFirst` is ascending. Carbon's own sorting, which only sorts the page, is turned off.
- **Another list:** `reset` loads the first page, showing the views loading, and closes the open item, such as when a filter which `list` reads changes.
- **The open item:** `open` and `close`. It's kept current as the page loads, or loaded with `get` when it's left the page, and closed when it's gone (a 404).
- **Tiles:** `gs-tiles` shows a list as tiles, with a `tile` function, and an `all` tile before them. It fires `gs-tile-select` with the value of a tile which is clicked, and the app sets `selected`.

## Routing

`Router` shows the page of the location's hash and hides the others. `#database/shop` is the page `database` with the parameter `shop`. A hash which isn't a page shows the default.

```ts
new Router('databases')
    .add('databases', document.getElementById('databases'))
    .add('database', document.getElementById('database'), (route) => showDatabase(route.params[0]))
    .onRoute((route) => nav.setActive(route.page))
    .start();
```

## Developing

| Command | Description |
|---|---|
| `make frontend` | Installs from `package-lock.json`, type checks, and builds into `frontend/dist` |
| `make frontend-test` | Type checks, and runs the tests |
| `npm test` | Runs the tests, from the TypeScript, with Node's test runner |

The tests need Node 22.18 or later, which runs TypeScript.
