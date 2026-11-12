// The gs-tiles element shows a list controller's page as a row of tiles,
// such as one for each database, with its size, after one for all of them,
// of which one is chosen, to show what's in it below. Clicking a tile fires
// gs-tile-select, with the tile's value. It shows a skeleton until it has
// items, and the items as unavailable while the server can't be reached. It
// needs lit and @carbon/web-components, which are the package's optional
// peer dependencies.

import { LitElement, css, html, nothing } from 'lit';
import type { ListController, ListState } from './controller.ts';

// Carbon components
import '@carbon/web-components/es/components/skeleton-placeholder/index.js';
import '@carbon/web-components/es/components/tag/tag.js';

// Tile is an item's tile: its value, which is chosen, its label, a badge
// beside it, such as its owner, its text, such as its size, which is
// prominent, and what it is, which is shown when it's hovered
export interface Tile {
    value: string;
    label: string;
    badge?: string;
    text?: string;
    title?: string;
}

// TilesTag is the element's tag, and TileSelectEvent the name of the event
// fired when a tile is clicked, whose detail is its value
export const TilesTag = 'gs-tiles';
export const TileSelectEvent = 'gs-tile-select';

export class TilesElement<T = unknown> extends LitElement {
    static override properties = {
        controller: { attribute: false },
        tile: { attribute: false },
        all: { attribute: false },
        selected: { type: String },
        label: { type: String },
        state: { state: true },
    };

    // controller is the list controller whose page is shown
    declare controller?: ListController<T>;

    // tile returns an item's tile
    declare tile?: (item: T) => Tile;

    // all is a tile before the items, such as for all of them, whose value
    // is chosen to show all of them
    declare all?: Tile;

    // selected is the value of the tile which is chosen
    declare selected?: string;

    // label is what the tiles choose, such as Databases, for assistive
    // technology
    declare label: string;
    declare private state?: ListState<T>;
    private unsubscribe?: () => void;

    static override styles = css`
        :host {
            display: block;
            margin: 0 0 2rem;
        }

        .tiles {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(12rem, 1fr));
            gap: 1px;
        }

        button {
            all: unset;
            box-sizing: border-box;
            display: flex;
            flex-direction: column;
            gap: 0.25rem;
            min-width: 0;
            padding: 0.75rem 1rem;
            background: var(--cds-layer-01, #f4f4f4);
            border-bottom: 2px solid transparent;
            cursor: pointer;
        }

        button:hover {
            background: var(--cds-layer-hover-01, #e8e8e8);
        }

        button:focus-visible {
            outline: 2px solid var(--cds-focus, #0f62fe);
            outline-offset: -2px;
        }

        button[aria-pressed='true'] {
            background: var(--cds-layer-selected-01, #e0e0e0);
            border-bottom-color: var(--cds-border-interactive, #0f62fe);
        }

        /* A tile's label, with its badge beside it */
        .label {
            display: flex;
            align-items: center;
            gap: 0.25rem;
            min-height: 1.5rem;
            color: var(--cds-text-secondary, #525252);
            font-size: 0.75rem;
        }

        .name {
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
        }

        .label cds-tag {
            margin: 0;
        }

        .text {
            font-size: 1.25rem;
            font-variant-numeric: tabular-nums;
        }

        .empty {
            color: var(--cds-text-secondary, #525252);
        }

        /* The items, while the server can't be reached */
        :host([unavailable]) .tiles {
            opacity: 0.5;
        }

        cds-skeleton-placeholder {
            width: 100%;
            height: 4.5rem;
        }
    `;

    constructor() {
        super();
        this.label = 'Choose';
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

    protected override willUpdate(changed: Map<PropertyKey, unknown>) {
        if (changed.has('controller')) {
            this.watch();
        }
        this.toggleAttribute('unavailable', this.state?.unavailable ?? false);
    }

    protected override render() {
        const items = this.state?.items;
        if (!items || !this.tile) {
            return html`<cds-skeleton-placeholder></cds-skeleton-placeholder>`;
        }
        return html`
            <div class="tiles" role="group" aria-label=${this.label}>
                ${this.all ? this.renderTile(this.all) : nothing}
                ${items.map((item) => this.renderTile(this.tile!(item)))}
            </div>
        `;
    }

    // The badge of the tile which is chosen is blue, as its border is, as
    // gray is the tile's colour then
    private renderTile(tile: Tile) {
        const pressed = tile.value === this.selected;
        return html`
            <button type="button" aria-pressed=${pressed ? 'true' : 'false'} title=${tile.title ?? nothing}
                @click=${() => !pressed && this.dispatchEvent(new CustomEvent(TileSelectEvent, { detail: tile.value, bubbles: true, composed: true }))}>
                <span class="label">
                    <span class="name">${tile.label}</span>
                    ${tile.badge ? html`<cds-tag size="sm" type=${pressed ? 'blue' : 'cool-gray'}>${tile.badge}</cds-tag>` : nothing}
                </span>
                <span class="text ${tile.text ? '' : 'empty'}">${tile.text || '–'}</span>
            </button>
        `;
    }

    // watch shows the controller's page, while the element is in the page
    private watch() {
        this.unsubscribe?.();
        this.unsubscribe = this.isConnected ? this.controller?.subscribe((state) => { this.state = state; }) : undefined;
        if (!this.unsubscribe) {
            this.state = undefined;
        }
    }
}

if (!customElements.get(TilesTag)) {
    customElements.define(TilesTag, TilesElement);
}

declare global {
    interface HTMLElementTagNameMap {
        'gs-tiles': TilesElement;
    }
}
