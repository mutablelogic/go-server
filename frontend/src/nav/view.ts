// The app's shell, of a navigation controller. gs-header is the menu bar at the top: a
// Carbon header, with its menu button, which opens the sidebar on small
// screens, the app's name, and a menu of the sections, which link to their
// first pages, of which the active page's is selected, so each module of an
// app has its item, even when there's only one. gs-side-nav is the
// sidebar: a Carbon side navigation, of the active section's pages, which
// the module which owns the section, such as Credentials, sets, of which
// the active page is selected, and the element's own content, such as
// gs-server-status, at its bottom.
//
// They render without a shadow root, so Carbon's theme, such as the class
// cds--g100 of the element they're in, and its ui-shell, which opens the
// side navigation from the header's button, work as they do in a page's
// HTML. They need lit and @carbon/web-components, which are the package's
// optional peer dependencies.

import { LitElement, html, nothing, type PropertyValues } from 'lit';
import type { NavController, NavState } from './controller.ts';

// Carbon components
import '@carbon/web-components/es/components/ui-shell/index.js';

// HeaderTag and SideNavTag are the elements' tags
export const HeaderTag = 'gs-header';
export const SideNavTag = 'gs-side-nav';

// ShellElement shows a navigation controller's state, without a shadow root
abstract class ShellElement extends LitElement {
    static override properties = {
        controller: { attribute: false },
        state: { state: true },
    };

    // controller is the navigation controller which is shown
    declare controller?: NavController;
    declare protected state?: NavState;
    private unsubscribe?: () => void;

    protected override createRenderRoot(): HTMLElement | DocumentFragment {
        return this;
    }

    override connectedCallback() {
        super.connectedCallback();
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

    private watch() {
        this.unsubscribe?.();
        this.unsubscribe = this.isConnected ? this.controller?.subscribe((state) => { this.state = state; }) : undefined;
        if (!this.unsubscribe) {
            this.state = undefined;
        }
    }
}

export class HeaderElement extends ShellElement {
    // The active section's item is current, as true, rather than page, as
    // Carbon's header item only shows it's current, with is-active, when
    // aria-current isn't page, which it doesn't pass to its link, whose
    // styles show it then
    protected override render() {
        const state = this.state;
        if (!state) {
            return nothing;
        }
        const sections = this.controller?.sections ?? [];
        return html`
            <cds-header aria-label=${state.name}>
                <cds-header-menu-button button-label-active="Close menu" button-label-inactive="Open menu"></cds-header-menu-button>
                <cds-header-name href=${state.href} prefix=${state.prefix ?? ''}>${state.name}</cds-header-name>
                ${sections.length ? html`
                    <cds-header-nav menu-bar-label=${state.name}>
                        ${sections.map((section) => html`
                            <cds-header-nav-item href=${section.href} ?is-active=${section.active} aria-current=${section.active ? 'true' : nothing}>${section.label}</cds-header-nav-item>
                        `)}
                    </cds-header-nav>
                ` : nothing}
            </cds-header>
        `;
    }
}

export class SideNavElement extends ShellElement {
    // The side navigation renders into Carbon's side navigation, before the
    // element's own content, such as gs-server-status, which is moved into
    // it, at the bottom. It's moved, not taken out of the page, so the page
    // still finds it, and its content stays connected.
    protected override createRenderRoot(): HTMLElement {
        const nav = document.createElement('cds-side-nav');
        nav.setAttribute('aria-label', 'Side navigation');
        nav.setAttribute('collapse-mode', 'responsive');
        nav.append(...Array.from(this.childNodes));
        this.append(nav);
        this.renderOptions.renderBefore = nav.firstChild;
        return nav;
    }

    protected override render() {
        const state = this.state;
        if (!state) {
            return nothing;
        }
        const section = this.controller?.activeSection;
        return html`
            <cds-side-nav-items>
                ${(section?.pages ?? []).map((page) => html`
                    <cds-side-nav-link href=${page.href} ?active=${page.id === state.active} aria-current=${page.id === state.active ? 'page' : 'false'}>${page.label}</cds-side-nav-link>
                `)}
            </cds-side-nav-items>
        `;
    }
}

if (!customElements.get(HeaderTag)) {
    customElements.define(HeaderTag, HeaderElement);
}
if (!customElements.get(SideNavTag)) {
    customElements.define(SideNavTag, SideNavElement);
}

declare global {
    interface HTMLElementTagNameMap {
        [HeaderTag]: HeaderElement;
        [SideNavTag]: SideNavElement;
    }
}
