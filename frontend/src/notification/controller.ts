// The notification controller keeps the notifications which are pushed to
// it, such as errors, until they time out, or are closed, and tells the
// views which subscribe to it, such as the gs-notifications element, which
// shows them as toasts, whenever they change

import type { Notification, ShownNotification } from './schema.ts';

// NotificationListener is called with the notifications which are shown, in
// the order they were pushed, whenever they change
export type NotificationListener = (notifications: readonly ShownNotification[]) => void;

// NotificationOptions are the controller's options
export interface NotificationOptions {
    // timeout is how long a notification is shown, in milliseconds, unless
    // it says otherwise, which is 8s by default
    timeout?: number;
}

export class NotificationController {
    private readonly timeout: number;
    private readonly listeners = new Set<NotificationListener>();
    private readonly timers = new Map<number, ReturnType<typeof setTimeout>>();
    private shown: ShownNotification[] = [];
    private nextId = 1;

    constructor(options: NotificationOptions = {}) {
        this.timeout = options.timeout ?? 8000;
    }

    // notifications are those which are shown, in the order they were pushed
    get notifications(): readonly ShownNotification[] {
        return this.shown;
    }

    // push shows the notification, until it times out, or is closed, and
    // returns its id, which closes it
    push(notification: Notification): number {
        const id = this.nextId++;
        this.shown = [...this.shown, { ...notification, id }];
        const timeout = notification.timeout === undefined ? this.timeout : notification.timeout;
        if (timeout !== null) {
            this.timers.set(id, setTimeout(() => this.close(id), timeout));
        }
        this.changed();
        return id;
    }

    // error shows an error, with the title saying what failed, and the error
    // why, and returns its id
    error(title: string, error: unknown): number {
        return this.push({ kind: 'error', title, subtitle: error instanceof Error ? error.message : String(error) });
    }

    // close removes the notification, such as once it's closed in a view
    close(id: number) {
        clearTimeout(this.timers.get(id));
        this.timers.delete(id);
        const shown = this.shown.filter((notification) => notification.id !== id);
        if (shown.length !== this.shown.length) {
            this.shown = shown;
            this.changed();
        }
    }

    // clear removes every notification
    clear() {
        this.timers.forEach((timer) => clearTimeout(timer));
        this.timers.clear();
        if (this.shown.length > 0) {
            this.shown = [];
            this.changed();
        }
    }

    // subscribe calls the function with the notifications which are shown,
    // now, and whenever they change, until the function it returns is
    // called, which unsubscribes
    subscribe(fn: NotificationListener): () => void {
        this.listeners.add(fn);
        fn(this.shown);
        return () => this.listeners.delete(fn);
    }

    private changed() {
        this.listeners.forEach((fn) => fn(this.shown));
    }
}
