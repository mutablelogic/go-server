// The config controller reads the server's configuration, and follows the
// server, with the broadcaster's changes controller: it's down when the
// server says it's stopping, or the changes can't be watched, and when it
// can be watched again, the configuration is read again, so it's up, or it
// was upgraded, when its version changed. With a notification controller,
// it says so, and that an upgraded server can be used by refreshing the
// page.

import type { ChangesController } from '../broadcaster/controller.ts';
import type { NotificationController } from '../notification/controller.ts';
import { getConfig, type ConfigOptions } from './httpclient.ts';
import type { Config } from './schema.ts';

// ServerStatus is the server's status, as the page sees it: loading, until
// its configuration is read, up, down, or upgraded, when it's up, with
// another version than the page's, until the page is refreshed
export type ServerStatus = 'loading' | 'up' | 'down' | 'upgraded';

// ConfigState is the configuration which the page was loaded with, the
// configuration as it is now, which is read when the server is up again,
// and the server's status
export interface ConfigState {
    config?: Config;
    latest?: Config;
    status: ServerStatus;
}

// ConfigListener is called with the state, whenever it changes
export type ConfigListener = (state: ConfigState) => void;

// ConfigServer is what the controller follows the server with, which the
// broadcaster's ChangesController is
export type ConfigServer = Pick<ChangesController, 'onConnect' | 'onDisconnect' | 'onShutdown'>;

// ConfigControllerOptions are where the configuration is, and the
// notification controller which is told when the server is down, up again,
// or upgraded, when there is one
export interface ConfigControllerOptions extends ConfigOptions {
    notifications?: NotificationController;
    reload?: () => void;
}

export class ConfigController {
    private readonly options: ConfigControllerOptions;
    private readonly listeners = new Set<ConfigListener>();
    private readonly downs: (() => void)[] = [];
    private readonly ups: ((config: Config) => void)[] = [];
    private readonly upgrades: ((config: Config) => void)[] = [];
    private state: ConfigState = { status: 'loading' };

    // The notification which says the server is down, which is closed when
    // it's up again, and the one which says it's back, which is closed when
    // it's down again, or upgraded
    private downNotification?: number;
    private upNotification?: number;

    constructor(server: ConfigServer, options: ConfigControllerOptions = {}) {
        this.options = options;
        server.onShutdown(() => this.down());
        server.onDisconnect(() => this.down());
        server.onConnect(() => void this.connected());
    }

    // config is the configuration which the page was loaded with, once it's
    // read
    get config(): Config | undefined {
        return this.state.config;
    }

    // status is the server's status
    get status(): ServerStatus {
        return this.state.status;
    }

    // current is the controller's state
    get current(): ConfigState {
        return this.state;
    }

    // load reads the configuration, which the page was loaded with, and
    // returns it, so the server is up. When it can't be read, the server is
    // down, and it's read again when the server is up again. Throws the
    // error reading it.
    async load(): Promise<Config> {
        try {
            const config = await getConfig(this.options);
            this.closeDown();
            this.update({ config, latest: config, status: 'up' });
            return config;
        } catch (error) {
            this.down();
            throw error;
        }
    }

    // refresh reloads the page, so it uses the server as it is now
    refresh() {
        (this.options.reload ?? (() => globalThis.location?.reload()))();
    }

    // subscribe calls the function with the state, now, and whenever it
    // changes, until the function it returns is called, which unsubscribes
    subscribe(fn: ConfigListener): () => void {
        this.listeners.add(fn);
        fn(this.state);
        return () => this.listeners.delete(fn);
    }

    // onDown calls the function when the server goes down
    onDown(fn: () => void): this {
        this.downs.push(fn);
        return this;
    }

    // onUp calls the function with the configuration when the server is up
    // again, with the page's version
    onUp(fn: (config: Config) => void): this {
        this.ups.push(fn);
        return this;
    }

    // onUpgrade calls the function with the configuration when the server is
    // up again, with another version than the page's
    onUpgrade(fn: (config: Config) => void): this {
        this.upgrades.push(fn);
        return this;
    }

    // down says the server is down, once until it's up again
    private down() {
        if (this.state.status === 'down') {
            return;
        }
        this.update({ status: 'down' });
        this.downs.forEach((fn) => fn());
        this.closeUp();
        if (this.options.notifications && this.downNotification === undefined) {
            this.downNotification = this.options.notifications.push({
                kind: 'warning',
                title: 'The server is down',
                subtitle: "Changes are shown again once it's back",
                timeout: null,
            });
        }
    }

    // connected reads the configuration again, as the server is up again,
    // or, when it wasn't read, reads it for the first time
    private async connected() {
        if (this.state.config === undefined) {
            await this.load().catch(() => undefined);
            return;
        } else if (this.state.status !== 'down') {
            return;
        }
        let latest: Config;
        try {
            latest = await getConfig(this.options);
        } catch {
            // It's still down, until the changes are watched again
            return;
        }
        const notifications = this.options.notifications;
        this.closeDown();
        if ((latest.version ?? '') !== (this.state.config.version ?? '')) {
            this.update({ latest, status: 'upgraded' });
            this.upgrades.forEach((fn) => fn(latest));
            notifications?.push({
                kind: 'info',
                title: 'The server was upgraded',
                subtitle: `It's now ${latest.version || 'another version'}. Refresh the page to use it.`,
                action: { text: 'Refresh', run: () => this.refresh() },
                timeout: null,
            });
        } else {
            this.update({ latest, status: 'up' });
            this.ups.forEach((fn) => fn(latest));
            this.upNotification = notifications?.push({ kind: 'success', title: 'The server is back' });
        }
    }

    // closeDown closes the notification which says the server is down
    private closeDown() {
        if (this.downNotification !== undefined) {
            this.options.notifications?.close(this.downNotification);
            this.downNotification = undefined;
        }
    }

    // closeUp closes the notification which says the server is back
    private closeUp() {
        if (this.upNotification !== undefined) {
            this.options.notifications?.close(this.upNotification);
            this.upNotification = undefined;
        }
    }

    private update(state: Partial<ConfigState>) {
        this.state = { ...this.state, ...state };
        this.listeners.forEach((fn) => fn(this.state));
    }
}
