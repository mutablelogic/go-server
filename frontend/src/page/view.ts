// The gs-page-header element is the dark header at the top of a page,
// across its width, with an icon, a title, a subtitle or details beneath
// it, and actions on the right, such as Create, which are Carbon buttons.
// Controls, such as a search, can be slotted beside the title. It's in
// Carbon's g90 theme, so its buttons, and what's slotted, are too. It needs
// lit and @carbon/web-components, which are the package's optional peer
// dependencies.

import { LitElement, css, html, nothing } from 'lit';
import iconLoader from '@carbon/web-components/es/globals/internal/icon-loader.js';
import type { CarbonIconDescriptor } from '@carbon/web-components/es/globals/internal/icon-loader-utils.js';
import { runAction, type PageAction } from './action.ts';

// Carbon components
import '@carbon/web-components/es/components/button/index.js';

export type { PageAction } from './action.ts';

// PageHeaderTag is the element's tag
export const PageHeaderTag = 'gs-page-header';

// PageActionErrorEvent is fired, with the action and its error, when an
// action's run throws, or its promise rejects
export const PageActionErrorEvent = 'gs-page-action-error';

// PageActionError is the detail of a PageActionErrorEvent
export interface PageActionError {
    action: PageAction;
    error: unknown;
}

export class PageHeaderElement extends LitElement {
    static override properties = {
        heading: { type: String },
        subtitle: { type: String },
        details: { attribute: false },
        icon: { attribute: false },
        actions: { attribute: false },
        busy: { state: true },
    };

    // heading is the page's title, and subtitle the text beneath it
    declare heading: string;
    declare subtitle: string;

    // details are beneath the title, in place of the subtitle, when they're
    // set, such as a template of a description, an owner, and a size
    declare details?: unknown;

    // icon is before the title: a 32px Carbon icon descriptor, or a
    // rendered icon
    declare icon?: unknown;

    // actions are the buttons on the right
    declare actions: PageAction[];

    // busy are the actions whose promises are pending, whose buttons are
    // disabled
    declare private busy: Set<PageAction>;

    static override styles = css`
        :host {
            display: block;
            margin: var(--gs-page-header-margin, 0 0 1.5rem);
            padding: 1.5rem 2rem;
            background: var(--cds-background, #262626);
            color: var(--cds-text-primary, #f4f4f4);
        }

        :host([hidden]) {
            display: none;
        }

        header {
            display: flex;
            align-items: center;
            gap: 1rem;
        }

        .icon {
            display: flex;
            flex: none;
        }

        .icon svg {
            fill: currentColor;
        }

        .text {
            min-width: 0;
        }

        h1 {
            margin: 0;
            font-size: 1.75rem;
            font-weight: 400;
            line-height: 1.28572;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
        }

        p {
            margin: 0.125rem 0 0;
            color: var(--cds-text-secondary, #c6c6c6);
            font-size: 0.875rem;
            line-height: 1.28572;
        }

        /* What's slotted beside the title fills the rest of the header */
        ::slotted(*) {
            flex: 1;
            min-width: 0;
        }

        .actions {
            display: flex;
            flex: none;
            gap: 1px;
            margin-left: auto;
        }
    `;

    constructor() {
        super();
        this.heading = '';
        this.subtitle = '';
        this.actions = [];
        this.busy = new Set();
    }

    // The header is in Carbon's g90 theme, whose tokens the shadow root, and
    // what's slotted, inherit
    override connectedCallback() {
        super.connectedCallback();
        this.classList.add('cds--g90');
    }

    protected override render() {
        return html`
            <header>
                ${this.icon ? html`<span class="icon">${renderIcon(this.icon)}</span>` : nothing}
                <div class="text">
                    <h1>${this.heading}</h1>
                    ${this.details ? html`<p>${this.details}</p>` : this.subtitle ? html`<p>${this.subtitle}</p>` : nothing}
                </div>
                <slot></slot>
                ${this.actions.length ? html`
                    <div class="actions">
                        ${this.actions.map((action) => html`
                            <cds-button kind=${action.kind ?? 'primary'} size="md" ?disabled=${action.disabled || this.busy.has(action)} @click=${() => this.run(action)}>
                                ${action.label}
                                ${action.icon ? renderIcon(action.icon, 'icon') : nothing}
                            </cds-button>
                        `)}
                    </div>
                ` : nothing}
            </header>
        `;
    }

    // run runs an action, with its button disabled while its promise is
    // pending, and fires PageActionErrorEvent when it fails. An action
    // which is disabled, or busy, isn't run, even when it's clicked again
    // before its button is rendered disabled.
    private run(action: PageAction) {
        if (action.disabled || this.busy.has(action)) {
            return;
        }
        void runAction(action, (busy) => {
            const next = new Set(this.busy);
            if (busy) {
                next.add(action);
            } else {
                next.delete(action);
            }
            this.busy = next;
        }, (error) => {
            this.dispatchEvent(new CustomEvent<PageActionError>(PageActionErrorEvent, { detail: { action, error }, bubbles: true, composed: true }));
        });
    }
}

// renderIcon renders a Carbon icon descriptor, in the slot, when it's
// given, or else an icon which is rendered already, as it is
function renderIcon(icon: unknown, slot?: string) {
    if (isDescriptor(icon)) {
        return iconLoader(icon, slot ? { slot } : undefined);
    }
    return slot ? html`<span slot=${slot}>${icon}</span>` : icon;
}

// isDescriptor is true for a Carbon icon descriptor, such as
// @carbon/icons/es/add/16.js's default export
function isDescriptor(icon: unknown): icon is CarbonIconDescriptor {
    return typeof icon === 'object' && icon !== null && 'elem' in icon && 'attrs' in icon;
}

if (!customElements.get(PageHeaderTag)) {
    customElements.define(PageHeaderTag, PageHeaderElement);
}

declare global {
    interface HTMLElementTagNameMap {
        [PageHeaderTag]: PageHeaderElement;
    }
    // It bubbles, so it's heard on the element, or the document
    interface HTMLElementEventMap {
        [PageActionErrorEvent]: CustomEvent<PageActionError>;
    }
    interface DocumentEventMap {
        [PageActionErrorEvent]: CustomEvent<PageActionError>;
    }
}
