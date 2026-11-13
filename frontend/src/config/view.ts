// The gs-server-status element shows the server's status, as a config
// controller follows it, for a sidebar, such as the footer of a side
// navigation: the page's version, that the server is down, or that it was
// upgraded, which refreshes the page when it's clicked, and links to the
// API's documentation, when the server serves it, and to a repository, when
// it's given one. It needs lit and @carbon/web-components, which are the
// package's optional peer dependencies.

import { LitElement, css, html, nothing } from 'lit';
import { apiIcon, githubIcon, renewIcon, warningIcon } from '../internal/icons.ts';
import type { ConfigController, ConfigState } from './controller.ts';

// Carbon components
import '@carbon/web-components/es/components/icon-button/index.js';

// ServerStatusTag is the element's tag
export const ServerStatusTag = 'gs-server-status';

export class ServerStatusElement extends LitElement {
    static override properties = {
        controller: { attribute: false },
        repository: { type: String },
        state: { state: true },
    };

    // controller is the config controller whose server is shown
    declare controller?: ConfigController;

    // repository is the URL of the project's repository, such as on GitHub,
    // which is linked to, when it's set
    declare repository?: string;
    declare private state?: ConfigState;
    private unsubscribe?: () => void;

    static override styles = css`
        :host {
            display: flex;
            align-items: center;
            justify-content: space-between;
            min-height: 3rem;
            padding: 0 1rem;
            color: var(--cds-text-secondary, #525252);
            font-size: 0.75rem;
            letter-spacing: 0.32px;
        }

        .version,
        .links {
            display: flex;
            align-items: center;
            gap: 0.25rem;
        }

        .links {
            margin-right: -0.5rem;
        }

        .down {
            color: var(--cds-support-error, #da1e28);
        }

        .upgraded {
            color: var(--cds-support-warning, #f1c21b);
        }

        .icon {
            display: flex;
        }

        /* The name of an icon, for screen readers, as Carbon only names an
           icon button which isn't a link */
        .visually-hidden {
            position: absolute;
            width: 1px;
            height: 1px;
            overflow: hidden;
            clip: rect(0 0 0 0);
            white-space: nowrap;
        }
    `;

    override connectedCallback() {
        super.connectedCallback();
        this.watch();
    }

    override disconnectedCallback() {
        this.unsubscribe?.();
        this.unsubscribe = undefined;
        super.disconnectedCallback();
    }

    protected override willUpdate(changed: Map<PropertyKey, unknown>) {
        if (changed.has('controller')) {
            this.watch();
        }
    }

    protected override render() {
        const state = this.state;
        if (!state?.config) {
            return nothing;
        }
        return html`
            <span class="version">
                <span>${state.config.version ?? ''}</span>
                ${this.renderStatus(state)}
            </span>
            <span class="links">
                ${state.config.openapi ? html`
                    <cds-icon-button kind="ghost" size="sm" align="top-end" href=${state.config.openapi} target="_blank" rel="noopener">
                        <span slot="icon" class="icon">${apiIcon}<span class="visually-hidden">OpenAPI documentation</span></span>
                        <span slot="tooltip-content">OpenAPI documentation</span>
                    </cds-icon-button>
                ` : nothing}
                ${this.repository ? html`
                    <cds-icon-button kind="ghost" size="sm" align="top-end" href=${this.repository} target="_blank" rel="noopener">
                        <span slot="icon" class="icon">${githubIcon}<span class="visually-hidden">Repository</span></span>
                        <span slot="tooltip-content">Repository</span>
                    </cds-icon-button>
                ` : nothing}
            </span>
        `;
    }

    // renderStatus returns that the server is down, or was upgraded, which
    // refreshes the page, or nothing, when it's up
    private renderStatus(state: ConfigState) {
        switch (state.status) {
            case 'down':
                return html`
                    <cds-icon-button kind="ghost" size="sm" align="top-start">
                        <span slot="icon" class="icon down">${warningIcon}<span class="visually-hidden">The server is down</span></span>
                        <span slot="tooltip-content">The server is down. Changes are shown again once it's back.</span>
                    </cds-icon-button>
                `;
            case 'upgraded':
                return html`
                    <cds-icon-button kind="ghost" size="sm" align="top-start" @click=${() => this.controller?.refresh()}>
                        <span slot="icon" class="icon upgraded">${renewIcon}<span class="visually-hidden">The application was upgraded</span></span>
                        <span slot="tooltip-content">${state.latest?.version === state.config?.version
                            ? "The server's been rebuilt"
                            : `The server is now ${state.latest?.version || 'another version'}`}. Refresh the page to use it.</span>
                    </cds-icon-button>
                `;
            default:
                return nothing;
        }
    }

    // watch shows the controller's server, while the element is in the page
    private watch() {
        this.unsubscribe?.();
        this.unsubscribe = this.isConnected ? this.controller?.subscribe((state) => { this.state = state; }) : undefined;
        if (!this.unsubscribe) {
            this.state = undefined;
        }
    }
}

// The element is defined once, so a page which imports the view more than
// once, such as from two bundles, doesn't fail
if (!customElements.get(ServerStatusTag)) {
    customElements.define(ServerStatusTag, ServerStatusElement);
}

declare global {
    interface HTMLElementTagNameMap {
        [ServerStatusTag]: ServerStatusElement;
    }
}
