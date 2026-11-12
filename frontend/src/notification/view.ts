// The gs-notifications element shows a notification controller's
// notifications as Carbon toasts, stacked at the top right of the page, and
// closes them in the controller when they're closed. It needs lit and
// @carbon/web-components, which are the package's optional peer
// dependencies.
//
// Its position is set by --gs-notifications-top, below a 3rem header by
// default, --gs-notifications-right, and --gs-notifications-z-index.

import { LitElement, css, html, nothing, type PropertyValues } from 'lit';
import { repeat } from 'lit/directives/repeat.js';
import type { NotificationController } from './controller.ts';
import type { ShownNotification } from './schema.ts';

// Carbon components
import '@carbon/web-components/es/components/notification/toast-notification.js';

// NotificationsTag is the element's tag
export const NotificationsTag = 'gs-notifications';

export class NotificationsElement extends LitElement {
    static override properties = {
        controller: { attribute: false },
        notifications: { state: true },
    };

    // controller is the notification controller whose notifications are shown
    declare controller?: NotificationController;
    declare private notifications: readonly ShownNotification[];
    private unsubscribe?: () => void;

    static override styles = css`
        :host {
            position: fixed;
            top: var(--gs-notifications-top, calc(3rem + 1rem));
            right: var(--gs-notifications-right, 1rem);
            z-index: var(--gs-notifications-z-index, 9000);
            display: block;
        }

        /* A notification's link, which is on the toast's dark background */
        a {
            display: block;
            margin-bottom: 1rem;
            color: var(--cds-link-inverse, #78a9ff);
        }
    `;

    constructor() {
        super();
        this.notifications = [];
    }

    override connectedCallback() {
        super.connectedCallback();
        this.setAttribute('aria-live', 'polite');
        this.watch();
    }

    override disconnectedCallback() {
        this.unsubscribe?.();
        this.unsubscribe = undefined;
        super.disconnectedCallback();
    }

    protected override willUpdate(changed: PropertyValues<this>) {
        if (changed.has('controller')) {
            this.watch();
        }
    }

    protected override render() {
        return html`${repeat(this.notifications, (notification) => notification.id, (notification) => html`
            <cds-toast-notification kind=${notification.kind} title=${notification.title}
                subtitle=${notification.subtitle ?? ''} caption=${notification.caption ?? ''}
                @cds-notification-closed=${() => this.controller?.close(notification.id)}>
                ${notification.link ? html`<a href=${notification.link.href}
                    @click=${() => this.controller?.close(notification.id)}>${notification.link.text}</a>` : nothing}
            </cds-toast-notification>
        `)}`;
    }

    // watch shows the controller's notifications, while the element is in
    // the page
    private watch() {
        this.unsubscribe?.();
        this.unsubscribe = this.isConnected ? this.controller?.subscribe((notifications) => { this.notifications = notifications; }) : undefined;
        if (!this.unsubscribe) {
            this.notifications = [];
        }
    }
}

// The element is defined once, so a page which imports the view more than
// once, such as from two bundles, doesn't fail
if (!customElements.get(NotificationsTag)) {
    customElements.define(NotificationsTag, NotificationsElement);
}

declare global {
    interface HTMLElementTagNameMap {
        [NotificationsTag]: NotificationsElement;
    }
}
