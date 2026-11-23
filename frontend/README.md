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
| `@mutablelogic/go-server/panel` | `PanelController`, which keeps a panel's item, header, tabs and actions, such as Save and Cancel, which close it, and runs them |
| `@mutablelogic/go-server/panel/view` | The `gs-panel` element, which shows a controller's panel on the right of the page, with its header, its tabs, its body, which is the element's content, and its actions |
| `@mutablelogic/go-server/player` | `Media`, what's played, `play`, which plays it in the page's player, of `PlayEvent`, and `playable`, whether a type is audio or video |
| `@mutablelogic/go-server/player/view` | The `gs-media-player` element, for a sidebar, which plays the media of each `PlayEvent`, until it's closed, and says when it can't be loaded |
| `@mutablelogic/go-server/config` | `getConfig`, which reads the server's `/config.json`, `configClient`, which is the client of the API it says is at its prefix, and `ConfigController`, which follows the server, as it goes down, comes back, and is upgraded |
| `@mutablelogic/go-server/config/view` | The `gs-server-status` element, for a sidebar, which shows the server's version, that it's down, or upgraded, and links to its documentation |
| `@mutablelogic/go-server/list` | `ListController`, which keeps a list an API returns a page at a time, or, when it scrolls, the pages loaded so far, sorted by the server, and follows its changes |
| `@mutablelogic/go-server/list/view` | The `gs-table` element, a Carbon data table of a list's page, with sortable columns and pagination, or infinite scroll with a sticky header, whose rows open their items, can be expanded, and selected for batch actions, with a toolbar of a search and settings |
| `@mutablelogic/go-server/list/tiles` | The `gs-tiles` element, a row of tiles of a list, after one for all of them, of which one is chosen |
| `@mutablelogic/go-server/nav` | `NavController`, which keeps an app's name, and its pages, in sections, and which is active. A module owns a section with `setSection`, or shares one, owning only its pages, with `setModule`. Pages with a `menu` are in a menu of the sidebar, such as Workspaces, with the pages next to them with the same menu, and pages with a `parent` are indented beneath it |
| `@mutablelogic/go-server/nav/view` | The `gs-header` and `gs-side-nav` elements, the app's menu bar of its sections, when there's more than one, and its sidebar of the active section's pages, above the element's own content, such as `gs-server-status` |
| `@mutablelogic/go-server/page` | `PageAction`, an action of a page's header, or a menu of `PageMenuItem`s, and `runAction`, which runs one, busy while its promise is pending |
| `@mutablelogic/go-server/page/view` | The `gs-page-header` element, the dark header at the top of a page, with an icon, a title, a subtitle or details, and actions on the right |
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

A module of the server, such as go-auth's credentials, can say what it can do, in its section of `modules`, which the server adds with `config.WithSection`, and documents by its type. `configSection` returns one, or nothing, when the server doesn't say, such as when it's older:

```go
config.RegisterHandler(router, config.New(ctx, router, openapi),
    config.WithSection("credential", Features{Decrypt: true, Rotate: true}),
)
```

```ts
// { prefix: '/api', ..., modules: { credential: { decrypt: true, rotate: true } } }
const features = configSection<{ decrypt: boolean; rotate: boolean }>(config, 'credential');
```

A section which changes while the server runs, such as a module's status, is added with `config.WithSectionFunc`, whose function returns its value each time `config.json` is served, as `getConfig` reads it again:

```go
config.WithSectionFunc("ldap", func() Status { return manager.Status() }),
```

## Following the server

A `ConfigController` follows the server with the `ChangesController`. Its status is:

| Status | When |
|---|---|
| `loading` | until the configuration is read, unless it's given, as `config`, when the page has read it |
| `up` | once it's read, or the server is back, with the page's version |
| `down` | when the server says it's stopping, or the changes can't be watched |
| `upgraded` | when the server is back, with another version than the page's, or the same version built again, until the page is refreshed |

A server says when it was built, as `build`, from `pkg/version`'s `BuildDate`, which its Makefile sets with `-ldflags "-X github.com/mutablelogic/go-server/pkg/version.BuildDate=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)"`, as go-server's own does. Without it, a server rebuilt with the same version, such as while it's developed, isn't known to be upgraded.

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
- **Tabs:** optional, beneath the header, such as View, Edit and Delete, each of its own actions, with `setTabs(tabs, tab?)`, or the `tabs` option. The chosen tab is `tab`, which the body is rendered from, on the state's change. It's the first which isn't disabled, unless it's chosen.
  - `selectTab(value)` chooses another, as clicking it does, unless it's disabled, or an action is running.
  - `setActions` sets the chosen tab's actions, which it keeps while another is chosen, such as Delete, once it's enabled.

```ts
panel.setTabs([
    { value: 'view', label: 'View', actions: [{ label: 'Unmount', kind: 'secondary', run: unmount }] },
    { value: 'edit', label: 'Edit', actions: [{ label: 'Save', run: save }] },
    { value: 'delete', label: 'Delete', actions: [{ label: 'Delete', kind: 'danger', disabled: true, run: remove }] },
]);
panel.subscribe(({ tab }) => render(tab));
```

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
- **Legend:** a table's `legend`, `{ title, items: [{ term, description }] }`, explains what's in it, such as its tags, to the right of the table, or beneath it when the element is narrower than 48rem. A `term` is any template, such as a tag, as it's shown in the table. Its width is `--gs-table-legend-width`, 16rem by default.
- **Expandable rows:** a table's `expand`, `(item) => content`, makes its rows expandable, with a button which doesn't open the item, and shows the content beneath a row when it's expanded, such as an item's history. It's only called for an expanded row, so the content can load then, such as with lit's `until`. `itemKey`, `(item) => string`, keeps a row expanded when the page loads again, whose items are new. Without it, only rows on the page stay expanded, or selected, as a page which loads again has new items.
- **Batch actions:** a table's `actions`, `[{ label, kind, icon, run(items) }]`, makes its rows selectable, with a checkbox, and one in the header, which selects the page. While items are selected, Carbon's batch actions bar is above the table, with how many, the actions, and Cancel. The selection is kept across pages, and when the page loads again, by `itemKey`. An action's buttons are disabled while its promise is pending, and the selection is cleared when it succeeds. When it fails, `gs-table-action-error` is fired, with the action and the error, and the selection is kept. `selection` is the selected items, `clearSelection()` clears it, and `gs-table-selection` is fired, with the items, when it changes.
- **Row actions:** a table's `rowActions`, `(item) => [{ label, icon, danger, divider, disabled, run(item) }]`, gives each row Carbon's overflow menu, a column at its end, of its item's actions, of which those which are `danger` are red, and a click in it doesn't open, or select, the row. When one fails, `gs-table-action-error` is fired, with the action and the error. An item which has none has no menu.
- **Toolbar:** a table's `search`, a placeholder such as Search roles, and `settings`, `[{ key, label, checked, disabled }]`, show Carbon's toolbar above it, with a search, and a menu of settings, such as Show disabled, which has a checkmark when it's on. The table doesn't filter its rows: `gs-table-search` is fired, with `value`, once typing pauses, and `gs-table-setting`, with a setting's `key`, and whether it's now `checked`, when it's chosen, and the app filters the list, and sets `settings` again. The batch actions bar is over the toolbar, while items are selected. `toolbarActions`, `[{ label, kind, icon, disabled, run }]`, are buttons at the toolbar's end, such as Create, which show it too. A button is disabled while its promise is pending, and `gs-table-action-error` is fired when it fails.
- **Scrolling:** a `ListController` with `scroll: true` adds pages as they're loaded instead of replacing them. `loadMore()` loads the next one, and `list` gets the last loaded item as `after`, so a cursor API can start after it. A page's `count` is optional: without it, `more` says whether there are more, or else a full page means there are. `refresh` reloads as many as are loaded, and `setSort` starts again from the top. The state adds `more`, `loading`, and `failed`, which `loadMore` clears.
- **Infinite scroll:** a table with `scrolls` has no pagination. It loads more when its end is within 200px of view, beneath rows of Carbon's skeleton while there are more, or says that more couldn't be loaded with Try again, and `countLabel(total)` beneath once all are loaded. Its header stays at the top of what scrolls it, at `--gs-table-sticky-top`, as the table's own scrolling is turned off.
- **Column widths:** a column's `width`, such as `8rem`, makes the columns fixed, and those without one share the rest, at least `--gs-table-column-width` (12rem) each. Cells are one line, cut short with an ellipsis. `align: 'end'` aligns a column's cells to its end, such as sizes. A table's `size`, Carbon's `xs` to `xl`, is its rows' height, `lg` by default. Its toolbar, with its search, settings, buttons and batch actions bar, is of `toolbarSize`, Carbon's `xs`, `sm` or `lg` (24, 32 or 48 pixels), or of the table's size, as Carbon's is: `xs` or `sm`, of those, and otherwise `lg`.
- **Tiles:** `gs-tiles` shows a list as tiles, with a `tile` function, and an `all` tile before them. It fires `gs-tile-select` with the value of a tile which is clicked, and the app sets `selected`. A tile's `badge`, such as an owner, can have a `badgeIcon`, such as Carbon's 16px user icon, which is shown at 12px, and a `badgeTitle`, such as Owner, which is shown when it's hovered, and read by screen readers.

## Page headers

`gs-page-header` is the dark header at the top of a page, in Carbon's g90 theme, with an icon, a title, a subtitle beneath it, and actions on the right.

```ts
import AddIcon from '@carbon/icons/es/add/16.js';
import DataBaseIcon from '@carbon/icons/es/data-base/32.js';
import { PageActionErrorEvent, PageHeaderTag } from '@mutablelogic/go-server/page/view';

const header = document.createElement(PageHeaderTag);
header.heading = 'Databases';
header.subtitle = 'The databases on the server';
header.icon = DataBaseIcon;
header.actions = [{ label: 'Create a database', icon: AddIcon, run: () => createDatabase() }];
header.addEventListener(PageActionErrorEvent, (event) => notifications.error("It couldn't be done", event.detail.error));
```

- **Details:** `details`, a template, replaces the subtitle, such as a description, an owner and a size.
- **Icons:** `icon` (32px) and an action's `icon` (16px) are Carbon icon descriptors, or icons which are rendered already.
- **Actions:**
  - `kind` is Carbon's button kind, primary by default, and `disabled` disables one.
  - When `run` returns a promise, its button is disabled until it settles.
  - When it throws, or rejects, the header fires `gs-page-action-error`, with the action and the error.
  - `menu`, of items, each a `label`, `icon` (16px, as an action's), `run` and `disabled`, makes it a Carbon menu button, such as Add, of what's added. The item which is chosen runs as an action does, with the button disabled until it settles, and its error is the action's. An action of a menu needs no `run`.
- **Controls:** what's in its default slot, such as a search, sits beside the title.
- **Spacing:** `--gs-page-header-margin` sets its margin, such as the negative of the page's padding, so it spans the page.

## Navigation

A `NavController` keeps an app's name, and its pages, each with an `id`, a `label`, a link, and an optional `section`. `gs-header` is the menu bar at the top, with the app's name, and a menu of the sections, so each module has its item, even when there's only one. `gs-side-nav` is the sidebar of the active section's pages. Both show the active page, which the router sets.

An app can be composed of modules, such as Credentials, which each own a section: its item in the menu bar, and the sidebar, while it's active.

```ts
import { NavController } from '@mutablelogic/go-server/nav';
import '@mutablelogic/go-server/nav/view';

const nav = new NavController({
    name: 'manager',
    prefix: 'pg',
    href: '#',
    pages: [
        { id: 'databases', label: 'Databases', href: '#databases', section: 'Data' },
        { id: 'roles', label: 'Roles', href: '#roles', section: 'Access' },
    ],
});
document.querySelector('gs-header')!.controller = nav;
document.querySelector('gs-side-nav')!.controller = nav;
```

- **Sections:** a section links to its first page. A page without a section is a section of its own, even when its label matches another's. `activeSection` is the active page's section, or the first, when no page is active.
- **Modules:** `setSection(section, pages)` replaces a section's pages, keeping where it is in the menu bar, or adds it at the end, so a module can change its own pages without the others'. No pages removes it.
- **Pages:** `setPages` replaces them all, and `setActive` makes one active.
- **Titles:** a page's `title` is shown when its link is pointed at, such as its whole label, which the sidebar cuts short, with an ellipsis, when it's long.
- **Pages beneath others:** a page with a `parent`, the id of a page on its own, is indented beneath it in the sidebar, in order, and is always shown, as it isn't collapsed, such as each volume beneath Volumes. They're a list of their own, named for their parent, for screen readers. A page and its parent must be in the same section, as the sidebar is of one section's pages: give them the same `section`, or add them with `setSection` or `setModule`, which do. A page with a parent is never in a menu: one whose parent isn't a page on its own is on its own.
- **Menus:** pages with a `menu`, next to each other, are in a collapsible menu of the sidebar.
- **Theme:** the elements render without a shadow root, so a Carbon theme class on the element they're in, such as `cds--g100`, applies, and Carbon's ui-shell opens the sidebar from the header's menu button.
- **The sidebar's content:** what's inside `gs-side-nav`, such as `gs-server-status`, is kept at the bottom of the sidebar.

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
