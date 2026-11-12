import { test } from 'node:test';
import assert from 'node:assert/strict';
import { ConfigController, type Config, type ConfigServer, type ServerStatus } from '../src/config/index.ts';
import { NotificationController } from '../src/notification/index.ts';
import { waitFor } from './fake.ts';

// fakeServer is what the controller follows the server with, whose events
// the test fires
function fakeServer() {
    const fns = { connect: [] as (() => void)[], disconnect: [] as (() => void)[], shutdown: [] as (() => void)[] };
    const server: ConfigServer & { fire: (event: keyof typeof fns) => void } = {
        onConnect(fn) { fns.connect.push(fn); return this as never; },
        onDisconnect(fn) { fns.disconnect.push(fn); return this as never; },
        onShutdown(fn) { fns.shutdown.push(fn); return this as never; },
        fire(event) { fns[event].forEach((fn) => fn()); },
    };
    return server;
}

// fakeConfig returns a fetch of the configurations, in turn, or which fails,
// when one is an error
function fakeConfig(...configs: (Config | Error)[]): typeof fetch {
    return (async () => {
        const config = configs.shift();
        if (!config || config instanceof Error) {
            throw config ?? new Error('no more configurations');
        }
        return new Response(JSON.stringify(config), { status: 200, headers: { 'Content-Type': 'application/json' } });
    }) as typeof fetch;
}

const v1: Config = { prefix: '/api', name: 'server', version: 'v1.0.0', openapi: '/api/openapi.html' };
const v2: Config = { ...v1, version: 'v1.1.0' };

test('ConfigController', async (t) => {
    await t.test('down and up', async () => {
        // The server is down when it stops, which is said once, until it's
        // up again, with the page's version, which is said too
        const server = fakeServer();
        const notifications = new NotificationController({ timeout: 1000 });
        const config = new ConfigController(server, { fetch: fakeConfig(v1, v1), notifications });
        const statuses: ServerStatus[] = [];
        config.subscribe((state) => statuses.push(state.status));
        let ups = 0;
        config.onUp(() => ups++);
        assert.deepEqual(await config.load(), v1);
        server.fire('shutdown');
        server.fire('disconnect');
        assert.equal(config.status, 'down');
        assert.deepEqual(notifications.notifications.map((n) => [n.kind, n.title, n.timeout]), [['warning', 'The server is down', null]]);
        server.fire('connect');
        await waitFor(() => config.status === 'up');
        assert.deepEqual(statuses, ['loading', 'up', 'down', 'up']);
        assert.equal(ups, 1);
        assert.deepEqual(notifications.notifications.map((n) => n.title), ['The server is back']);

        // Going down again closes the notification which said it was back
        server.fire('disconnect');
        assert.deepEqual(notifications.notifications.map((n) => n.title), ['The server is down']);
        notifications.clear();
    });

    await t.test('config', async () => {
        // A configuration the page has read is up, and isn't read again when
        // the server is connected
        const server = fakeServer();
        const config = new ConfigController(server, { config: v1, fetch: fakeConfig() });
        assert.equal(config.status, 'up');
        assert.deepEqual(config.config, v1);
        server.fire('connect');
        await new Promise((resolve) => setTimeout(resolve, 10));
        assert.equal(config.status, 'up');
    });

    await t.test('upgraded', async () => {
        // A server which is up again, with another version, was upgraded,
        // which is said, with an action which refreshes the page
        const server = fakeServer();
        const notifications = new NotificationController();
        let reloads = 0;
        const config = new ConfigController(server, { fetch: fakeConfig(v1, v2), notifications, reload: () => reloads++ });
        const upgrades: Config[] = [];
        config.onUpgrade((latest) => upgrades.push(latest));
        await config.load();
        server.fire('disconnect');
        server.fire('connect');
        await waitFor(() => config.status === 'upgraded');
        assert.deepEqual(config.current.latest, v2);
        assert.deepEqual(config.config, v1);
        assert.deepEqual(upgrades, [v2]);
        const upgraded = notifications.notifications.find((n) => n.title === 'The server was upgraded');
        assert.ok(upgraded?.action);
        assert.equal(upgraded.timeout, null);
        upgraded.action.run();
        assert.equal(reloads, 1);
        notifications.clear();
    });

    await t.test('rebuilt', async () => {
        // A server which is up again, with the same version, but built again,
        // was upgraded, which says it's been rebuilt
        const server = fakeServer();
        const notifications = new NotificationController();
        const built = { ...v1, build: '2026-11-12T18:00:00Z' };
        const rebuilt = { ...v1, build: '2026-11-12T19:00:00Z' };
        const config = new ConfigController(server, { fetch: fakeConfig(built, rebuilt), notifications });
        await config.load();
        server.fire('disconnect');
        server.fire('connect');
        await waitFor(() => config.status === 'upgraded');
        const upgraded = notifications.notifications.find((n) => n.title === 'The server was upgraded');
        assert.equal(upgraded?.subtitle, "It's been rebuilt. Refresh the page to use it.");
        notifications.clear();
    });

    await t.test('connected', async () => {
        // A stream which starts again, when the server wasn't down, doesn't
        // read the configuration again
        const server = fakeServer();
        const config = new ConfigController(server, { fetch: fakeConfig(v1) });
        await config.load();
        server.fire('connect');
        await new Promise((resolve) => setTimeout(resolve, 20));
        assert.equal(config.status, 'up');
    });

    await t.test('load fails', async () => {
        // A configuration which can't be read is that the server is down,
        // and it's read when the server is up
        const server = fakeServer();
        const notifications = new NotificationController();
        const config = new ConfigController(server, { fetch: fakeConfig(new Error('refused'), v1), notifications });
        await assert.rejects(config.load(), /refused/);
        assert.equal(config.status, 'down');
        assert.equal(notifications.notifications.length, 1);
        server.fire('connect');
        await waitFor(() => config.status === 'up');
        assert.deepEqual(config.config, v1);
        assert.deepEqual(notifications.notifications, []);
    });
});
