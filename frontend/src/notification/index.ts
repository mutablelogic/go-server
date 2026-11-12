// Notifications, such as errors, which a controller keeps until they time
// out, or are closed, and which views, such as gs-notifications, show

export { NotificationController, type NotificationListener, type NotificationOptions } from './controller.ts';
export type { Notification, NotificationKind, ShownNotification } from './schema.ts';
