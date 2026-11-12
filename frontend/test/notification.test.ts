import { test } from 'node:test';
import assert from 'node:assert/strict';
import { NotificationController, type ShownNotification } from '../src/notification/index.ts';

test('NotificationController', async (t) => {
    await t.test('push', async () => {
        // A notification is shown until its timeout, and a subscriber is
        // told of the notifications now, and whenever they change
        const controller = new NotificationController({ timeout: 20 });
        const seen: (readonly ShownNotification[])[] = [];
        const unsubscribe = controller.subscribe((notifications) => seen.push(notifications));
        const id = controller.push({ kind: 'info', title: 'Hello' });
        assert.deepEqual(controller.notifications, [{ kind: 'info', title: 'Hello', id }]);
        await new Promise((resolve) => setTimeout(resolve, 40));
        assert.deepEqual(controller.notifications, []);
        assert.deepEqual(seen.map((notifications) => notifications.length), [0, 1, 0]);
        unsubscribe();
        controller.push({ kind: 'info', title: 'Unseen', timeout: null });
        assert.equal(seen.length, 3);
    });

    await t.test('sticky', async () => {
        // A notification without a timeout stays until it's closed, and
        // one with its own timeout uses it
        const controller = new NotificationController({ timeout: 10 });
        const sticky = controller.push({ kind: 'warning', title: 'Stays', timeout: null });
        controller.push({ kind: 'info', title: 'Longer', timeout: 60 });
        await new Promise((resolve) => setTimeout(resolve, 30));
        assert.deepEqual(controller.notifications.map((n) => n.title), ['Stays', 'Longer']);
        controller.close(sticky);
        assert.deepEqual(controller.notifications.map((n) => n.title), ['Longer']);
        controller.clear();
        assert.deepEqual(controller.notifications, []);
    });

    await t.test('error', () => {
        // An error says what failed, and why
        const controller = new NotificationController();
        controller.error("It couldn't be saved", new Error('disk full'));
        controller.error('Something failed', 'just because');
        assert.deepEqual(controller.notifications.map(({ kind, title, subtitle }) => ({ kind, title, subtitle })), [
            { kind: 'error', title: "It couldn't be saved", subtitle: 'disk full' },
            { kind: 'error', title: 'Something failed', subtitle: 'just because' },
        ]);
        controller.clear();
    });

    await t.test('close', () => {
        // Closing a notification which isn't shown changes nothing
        const controller = new NotificationController();
        let changes = 0;
        controller.subscribe(() => changes++);
        controller.close(42);
        controller.clear();
        assert.equal(changes, 1);
    });
});
