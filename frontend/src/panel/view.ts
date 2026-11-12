// The gs-panel element shows a panel controller's panel, on the right of
// the page, over the page beneath the header, which is darkened, as a modal
// dialog: its optional header, with its icon, title, subtitle and close
// button, its body, which is the element's content, and which scrolls, and
// its optional footer, of its actions, such as Save and Cancel, which show
// they're running, until they're done, and close it. It's open while the controller says so, and closes it by its
// close button, the escape key, or clicking the page beside it, when it can
// be closed. It needs lit and @carbon/web-components, which are the
// package's optional peer dependencies.
//
// The header's title and subtitle are the controller's, or the content of
// the element's title and subtitle slots. Its position is set by
// --gs-panel-top, below a 3rem header by default, --gs-panel-width and
// --gs-panel-z-index.

import { LitElement, css, html, nothing, type PropertyValues } from 'lit';
import { closeIcon } from '../internal/icons.ts';
import type { PanelController } from './controller.ts';
import type { PanelActionState, PanelState } from './schema.ts';

// Carbon components
import '@carbon/web-components/es/components/button/index.js';
import '@carbon/web-components/es/components/icon-button/index.js';
import '@carbon/web-components/es/components/inline-loading/index.js';

// PanelTag is the element's tag
export const PanelTag = 'gs-panel';

export class PanelElement extends LitElement {
    static override properties = {
        controller: { attribute: false },
        open: { type: Boolean, reflect: true },
        state: { state: true },
    };

    // controller is the panel controller whose panel is shown
    declare controller?: PanelController<unknown>;

    // open is whether the panel is open, which is reflected, so it's styled
    declare open: boolean;
    declare private state?: PanelState<unknown>;
    private unsubscribe?: () => void;

    // Whether the panel has been shown, as it isn't rendered until it's first
    // opened, and then is, so it can close with its transition
    private shown = false;

    // Whether the panel is opening for the first time, when it's rendered
    // closed, and then opened, once it's rendered, so it slides in
    private opening = false;

    // What was focused when the panel opened, which is focused again when it
    // closes
    private returnFocus?: HTMLElement;

    // The elements which are made inert while the panel is open, as it's a
    // modal dialog, so focus, and clicks, stay in it
    private inerted: HTMLElement[] = [];

    static override styles = css`
        /* The panel covers the page beneath the header, which is hidden once
           the panel has closed */
        :host {
            position: fixed;
            inset: var(--gs-panel-top, 3rem) 0 0 0;
            z-index: var(--gs-panel-z-index, 8000);
            visibility: hidden;
            pointer-events: none;
            transition: visibility 0s 110ms;
        }

        :host([open]) {
            visibility: visible;
            pointer-events: auto;
            transition: none;
        }

        /* The page beside the panel is darkened, and closes the panel when
           it's clicked */
        .backdrop {
            position: absolute;
            inset: 0;
            background: var(--cds-overlay, rgba(22, 22, 22, 0.5));
            opacity: 0;
            transition: opacity 110ms cubic-bezier(0.2, 0, 0.38, 0.9);
        }

        :host([open]) .backdrop {
            opacity: 1;
            transition: opacity 240ms cubic-bezier(0, 0, 0.38, 0.9);
        }

        .panel {
            position: absolute;
            top: 0;
            right: 0;
            bottom: 0;
            display: flex;
            flex-direction: column;
            width: min(var(--gs-panel-width, 28rem), 100vw);
            background: var(--cds-layer-01, #f4f4f4);
            box-shadow: -2px 0 6px rgba(0, 0, 0, 0.2);
            transform: translateX(100%);
            transition: transform 110ms cubic-bezier(0.2, 0, 0.38, 0.9);
        }

        /* The panel is focused when it opens, but isn't outlined, as it isn't
           clicked, or tabbed to */
        .panel:focus {
            outline: none;
        }

        :host([open]) .panel {
            transform: none;
            transition: transform 240ms cubic-bezier(0, 0, 0.38, 0.9);
        }

        header {
            flex-shrink: 0;
            display: flex;
            align-items: flex-start;
            gap: 0.5rem;
            padding: 1rem 0.5rem 1rem 1rem;
            border-bottom: 1px solid var(--cds-border-subtle-01, #c6c6c6);
        }

        .title {
            flex: 1;
            min-width: 0;
        }

        h2 {
            margin: 0;
            font-size: var(--cds-heading-03-font-size, 1.25rem);
            font-weight: 600;
            line-height: var(--cds-heading-03-line-height, 1.4);
            overflow-wrap: anywhere;
        }

        .subtitle {
            margin-top: 0.25rem;
            color: var(--cds-text-secondary, #525252);
            font-size: 0.75rem;
            overflow-wrap: anywhere;
        }

        /* The icon beside the title, and its subtitle, as tall as they are
           together, on white, as brands' logos are drawn for it */
        .icon {
            flex: none;
            align-self: center;
            box-sizing: border-box;
            width: 2.5rem;
            height: 2.5rem;
            margin-right: 0.25rem;
            padding: 3px;
            border-radius: 4px;
            background: #fff;
            object-fit: contain;
        }

        /* The body scrolls, between the header, and the footer, which
           don't */
        .body {
            flex: 1;
            min-height: 0;
            overflow-y: auto;
            padding: 1rem;
        }

        /* The actions are at the bottom of the panel, side by side, and
           wrap */
        footer {
            flex-shrink: 0;
            display: flex;
            flex-wrap: wrap;
            align-items: center;
            gap: 1px;
            padding: 1rem;
            border-top: 1px solid var(--cds-border-subtle-01, #c6c6c6);
            background: var(--cds-layer-01, #f4f4f4);
        }

        footer cds-inline-loading {
            width: auto;
            padding: 0 1rem;
        }
    `;

    constructor() {
        super();
        this.open = false;
    }

    override connectedCallback() {
        super.connectedCallback();
        document.addEventListener('keydown', this.handleKeydown);
        this.watch();
    }

    override disconnectedCallback() {
        document.removeEventListener('keydown', this.handleKeydown);
        this.unsubscribe?.();
        this.unsubscribe = undefined;
        this.restoreBackground();
        super.disconnectedCallback();
    }

    protected override willUpdate(changed: PropertyValues<this>) {
        if (changed.has('controller')) {
            this.watch();
        }

        // The panel which opens for the first time is rendered closed, and
        // then opened, once it's rendered, so it slides in
        const open = this.state?.open ?? false;
        if (open && !this.shown) {
            this.shown = true;
            this.opening = true;
            this.open = false;
        } else if (!this.opening) {
            this.open = open;
        }
    }

    // updated makes the page beside the panel inert, and focuses the panel,
    // when it opens, as it's a modal dialog, and restores the page, and
    // focuses what was focused before, when it closes
    protected override updated(changed: PropertyValues<this>) {
        if (this.opening) {
            this.opening = false;
            this.renderRoot.querySelector('.panel')?.getBoundingClientRect();
            this.requestUpdate();
            return;
        }
        if (!changed.has('open') || (changed.get('open') === undefined && !this.open)) {
            return;
        }
        if (this.open) {
            this.returnFocus = deepActiveElement();
            this.inertBackground();
            this.renderRoot.querySelector<HTMLElement>('.panel')?.focus();
        } else {
            this.restoreBackground();
            this.returnFocus?.focus();
            this.returnFocus = undefined;
        }
    }

    protected override render() {
        const state = this.state;
        if (!state || !this.shown) {
            return nothing;
        }
        return html`
            <div class="backdrop" @click=${this.close}></div>
            <div class="panel" role="dialog" aria-modal="true" aria-labelledby="title" tabindex="-1">
                ${this.renderHeader(state)}
                <div class="body"><slot></slot></div>
                ${state.actions.length > 0 ? html`<footer>${state.actions.map((action) => this.renderAction(action))}</footer>` : nothing}
            </div>
        `;
    }

    // renderHeader returns the header, with its icon, title, subtitle and
    // close button, or nothing, when it has none of them
    private renderHeader(state: PanelState<unknown>) {
        const { title, subtitle, icon } = state.header;
        const slotted = (name: string) => this.querySelector(`:scope > [slot="${name}"]`) !== null;
        if (!title && !subtitle && !icon && !state.closable && !slotted('title') && !slotted('subtitle')) {
            return nothing;
        }
        return html`
            <header>
                ${icon ? html`<img class="icon" src=${icon} alt="" aria-hidden="true">` : nothing}
                <div class="title">
                    <h2 id="title"><slot name="title">${title ?? ''}</slot></h2>
                    <div class="subtitle"><slot name="subtitle">${subtitle ?? ''}</slot></div>
                </div>
                ${state.closable ? html`
                    <cds-icon-button kind="ghost" size="md" align="left" @click=${this.close}>
                        ${closeIcon}
                        <span slot="tooltip-content">Close</span>
                    </cds-icon-button>
                ` : nothing}
            </header>
        `;
    }

    // renderAction returns an action's button, which is disabled while
    // another action runs, or that it's running
    private renderAction(action: PanelActionState) {
        if (action.running) {
            return html`<cds-inline-loading status="active">${action.label}</cds-inline-loading>`;
        }
        return html`<cds-button kind=${action.kind ?? 'primary'} ?disabled=${action.disabled || this.controller?.running}
            @click=${() => void this.controller?.run(action)}>${action.label}</cds-button>`;
    }

    // watch shows the controller's panel, while the element is in the page
    private watch() {
        this.unsubscribe?.();
        this.unsubscribe = this.isConnected ? this.controller?.subscribe((state) => { this.state = state; }) : undefined;
        if (!this.unsubscribe) {
            this.state = undefined;
        }
    }

    private readonly close = () => {
        if (this.state?.closable) {
            this.controller?.close();
        }
    };

    // handleKeydown closes the panel when the escape key is pressed, while
    // it's open
    private readonly handleKeydown = (event: KeyboardEvent) => {
        if (event.key === 'Escape' && this.open) {
            this.close();
        }
    };

    // inertBackground makes the elements beside the panel, and beside each
    // of the elements it's in, up to the body, inert, which aren't already,
    // so only the panel can be focused, or clicked, but not the regions which
    // announce notifications, such as gs-notifications, which are read
    private inertBackground() {
        this.restoreBackground();
        for (let element: HTMLElement = this; element.parentElement && element !== document.body; element = element.parentElement) {
            for (const sibling of Array.from(element.parentElement.children)) {
                if (sibling !== element && sibling instanceof HTMLElement && !sibling.inert && !sibling.hasAttribute('aria-live')) {
                    sibling.inert = true;
                    this.inerted.push(sibling);
                }
            }
        }
    }

    // restoreBackground makes the elements which were made inert no longer
    // inert
    private restoreBackground() {
        for (const element of this.inerted) {
            element.inert = false;
        }
        this.inerted = [];
    }
}

// deepActiveElement returns the element which is focused, in the shadow roots
// of the elements which are focused
function deepActiveElement(): HTMLElement | undefined {
    let element = document.activeElement;
    while (element?.shadowRoot?.activeElement) {
        element = element.shadowRoot.activeElement;
    }
    return element instanceof HTMLElement ? element : undefined;
}

// The element is defined once, so a page which imports the view more than
// once, such as from two bundles, doesn't fail
if (!customElements.get(PanelTag)) {
    customElements.define(PanelTag, PanelElement);
}

declare global {
    interface HTMLElementTagNameMap {
        [PanelTag]: PanelElement;
    }
}
