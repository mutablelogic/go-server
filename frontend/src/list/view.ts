// The gs-table element shows a list controller's page as a Carbon data
// table, with its columns, and pagination beneath it, which loads the page,
// or the page's size, which is changed. A column which has a sort key sorts
// the list, by the server, when its header is clicked. Clicking a row opens
// its item in the controller, as does its first cell, which is a button, from
// the keyboard, and the item which is opened is selected. It
// shows a skeleton until it has items, and the items as unavailable while
// the server can't be reached. It can have a legend, to the right of the
// table, or beneath it, when the element is narrow, which explains what's in
// it, such as its tags. It needs lit and @carbon/web-components, which are
// the package's optional peer dependencies.

import { LitElement, css, html, nothing } from 'lit';
import type { ListController, ListState, Sort } from './controller.ts';

// Carbon components
import '@carbon/web-components/es/components/data-table/index.js';
import '@carbon/web-components/es/components/pagination/index.js';

// Column is a column of the table: its header, its cell of an item, and
// the key the list is sorted by, by the server, when it can be, and which
// direction it's sorted in first, which is descending, unless it's set
export interface Column<T> {
    header: string;
    cell: (item: T) => unknown;
    sort?: string;
    sortFirst?: Sort['direction'];
}

// LegendItem is what a legend explains, such as a tag, as it's shown in the
// table, and what it means
export interface LegendItem {
    term: unknown;
    description: string;
}

// Legend explains what's in a table, with a title, such as Context
export interface Legend {
    title?: string;
    items: LegendItem[];
}

// TableTag is the element's tag
export const TableTag = 'gs-table';

export class TableElement<T = unknown> extends LitElement {
    static override properties = {
        controller: { attribute: false },
        columns: { attribute: false },
        legend: { attribute: false },
        empty: { type: String },
        state: { state: true },
    };

    // controller is the list controller whose page is shown
    declare controller?: ListController<T>;

    // columns are the table's columns
    declare columns: Column<T>[];

    // legend explains what's in the table, beside it, when it's set
    declare legend?: Legend;

    // empty is what's shown when there are no items, such as No roles
    declare empty: string;
    declare private state?: ListState<T>;
    private unsubscribe?: () => void;

    // The sizes of a page which can be chosen
    pageSizes = [10, 25, 50, 100];

    static override styles = css`
        :host {
            display: block;
            container-type: inline-size;
        }

        /* The table scrolls when it's too wide, beside its legend, which is
           beneath it, when the element is narrow */
        .layout {
            display: grid;
            grid-template-columns: minmax(0, 1fr);
            gap: 1rem;
        }

        .layout.legend {
            grid-template-columns: minmax(0, 1fr) var(--gs-table-legend-width, 16rem);
        }

        @container (max-width: 48rem) {
            .layout.legend {
                grid-template-columns: minmax(0, 1fr);
            }
        }

        .table {
            overflow-x: auto;
            min-width: 0;
        }

        aside {
            align-self: start;
            padding: 1rem;
            background: var(--cds-layer-01, #f4f4f4);
            color: var(--cds-text-primary, #161616);
            font-size: 0.875rem;
            line-height: 1.28572;
        }

        aside h2 {
            margin: 0 0 0.75rem;
            font-size: 0.875rem;
            font-weight: 600;
        }

        aside dl {
            margin: 0;
        }

        aside dt {
            margin-top: 0.75rem;
        }

        aside dt:first-child {
            margin-top: 0;
        }

        aside dd {
            margin: 0.25rem 0 0;
            color: var(--cds-text-secondary, #525252);
        }

        /* A cell is one line, and the table scrolls when it's too wide */
        cds-table-cell {
            white-space: nowrap;
        }

        cds-table-row {
            cursor: pointer;
        }

        /* The item which is opened */
        cds-table-row[data-selected] cds-table-cell {
            background: var(--cds-layer-selected, #e0e0e0);
        }

        /* The first cell of a row is a button, which opens its item from the
           keyboard, and looks like the other cells. Its click is the row's. */
        button.open {
            all: unset;
            cursor: pointer;
        }

        button.open:focus-visible {
            outline: 2px solid var(--cds-focus, #0f62fe);
            outline-offset: 2px;
        }

        /* What's shown, across the table, when there are no items */
        .empty {
            padding: 1rem;
            background: var(--cds-layer-01, #f4f4f4);
            border-bottom: 1px solid var(--cds-border-subtle-01, #c6c6c6);
            color: var(--cds-text-secondary, #525252);
            font-size: 0.875rem;
        }

        /* The items, while the server can't be reached */
        :host([unavailable]) cds-table-body {
            opacity: 0.5;
        }
    `;

    constructor() {
        super();
        this.columns = [];
        this.empty = 'Nothing';
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
        if (!this.legend?.items.length) {
            return html`<div class="layout"><div class="table">${this.renderTable()}</div></div>`;
        }
        return html`
            <div class="layout legend">
                <div class="table">${this.renderTable()}</div>
                <aside aria-label=${this.legend.title ?? 'Legend'}>
                    ${this.legend.title ? html`<h2>${this.legend.title}</h2>` : nothing}
                    <dl>
                        ${this.legend.items.map((item) => html`<dt>${item.term}</dt><dd>${item.description}</dd>`)}
                    </dl>
                </aside>
            </div>
        `;
    }

    // renderTable returns the table, with its pagination, or a skeleton,
    // until it has items
    private renderTable() {
        const state = this.state;
        if (!state?.items) {
            // Carbon's skeleton re-renders forever when it has no headers
            // and no columns, so it always gets one header at least
            const headers = this.columns.length ? this.columns.map((column) => column.header) : [''];
            return html`<cds-table-skeleton .headers=${headers} row-count="3"></cds-table-skeleton>`;
        }
        return html`
            <cds-table>
                <cds-table-head>
                    <cds-table-header-row>
                        ${this.columns.map((column) => this.renderHeader(column, state.sort))}
                    </cds-table-header-row>
                </cds-table-head>
                <cds-table-body>
                    ${state.items.map((item) => html`
                        <cds-table-row ?data-selected=${state.selected === item} @click=${() => this.controller?.open(item)}>
                            ${this.columns.map((column, index) => html`<cds-table-cell>${index === 0 ? html`<button type="button" class="open">${column.cell(item)}</button>` : column.cell(item)}</cds-table-cell>`)}
                        </cds-table-row>
                    `)}
                </cds-table-body>
            </cds-table>
            ${state.items.length === 0 ? html`<div class="empty">${this.empty}</div>` : nothing}
            ${state.total > 0 ? html`
                <cds-pagination .page=${state.page} .pageSize=${state.pageSize} .start=${(state.page - 1) * state.pageSize}
                    .totalItems=${state.total}
                    @cds-pagination-changed-current=${(event: CustomEvent<{ page: number }>) => this.controller?.setPage(event.detail.page)}
                    @cds-page-sizes-select-changed=${(event: CustomEvent<{ pageSize: number }>) => this.controller?.setPageSize(event.detail.pageSize)}>
                    ${this.pageSizes.map((size) => html`<cds-select-item value=${size}>${size}</cds-select-item>`)}
                </cds-pagination>
            ` : nothing}
        `;
    }

    // renderHeader returns a column's header, which sorts the list, by the
    // server, when it's clicked, when the column can be sorted. Carbon's
    // table, which sorts the rows on the page itself, isn't told.
    private renderHeader(column: Column<T>, sort?: Sort) {
        if (column.sort === undefined) {
            return html`<cds-table-header-cell>${column.header}</cds-table-header-cell>`;
        }
        const key = column.sort;
        const first = column.sortFirst ?? 'descending';
        const direction = sort?.key === key ? sort.direction : 'none';
        return html`
            <cds-table-header-cell is-sortable sort-direction=${direction} sort-cycle=${first === 'descending' ? 'bi-states-from-descending' : 'bi-states-from-ascending'}
                @cds-table-header-cell-sort=${(event: CustomEvent<{ sortDirection: string }>) => {
                    event.preventDefault();
                    event.stopPropagation();
                    const next = direction === 'none' ? first : direction === 'descending' ? 'ascending' : 'descending';
                    this.controller?.setSort({ key, direction: next });
                }}>${column.header}</cds-table-header-cell>
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

if (!customElements.get(TableTag)) {
    customElements.define(TableTag, TableElement);
}

declare global {
    interface HTMLElementTagNameMap {
        [TableTag]: TableElement;
    }
}
