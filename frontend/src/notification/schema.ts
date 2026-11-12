// The notifications which a notification controller shows

// NotificationKind is the kind of a notification, which sets its colour and
// icon
export type NotificationKind = 'error' | 'warning' | 'info' | 'success';

// Notification is a notification which is pushed, which closes after its
// timeout, in milliseconds, or the controller's, or stays until it's closed,
// when it's null, with a caption, such as when it happened, and a link, such
// as to what it's about
export interface Notification {
    kind: NotificationKind;
    title: string;
    subtitle?: string;
    caption?: string;
    link?: { href: string; text: string };
    timeout?: number | null;
}

// ShownNotification is a notification which is shown, with its id, which
// closes it
export interface ShownNotification extends Notification {
    id: number;
}
