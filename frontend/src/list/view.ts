// The gs-table element shows a list controller's page as a Carbon data
// table, with its columns, and pagination beneath it, which loads the page,
// or the page's size, which is changed. A column which has a sort key sorts
// the list, by the server, when its header is clicked. Clicking a row opens
// its item in the controller, and the item which is opened is selected. It
// shows a skeleton until it has items, and the items as unavailable while
// the server can't be reached. It needs lit and @carbon/web-components,
// which are the package's optional peer dependencies.

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

// TableTag is the element's tag
export const TableTag = 'gs-table';

export class TableElement<T = unknown> extends LitElement {
    static override properties = {
        controller: { attribute: false },
        columns: { attribute: false },
        empty: { type: String },
        state: { state: true },
    };

    // controller is the list controller whose page is shown
    declare controller?: ListController<T>;

    // columns are the table's columns
    declare columns: Column<T>[];

    // empty is what's shown when there are no items, such as No roles
    declare empty: string;
    declare private state?: ListState<T>;
    private unsubscribe?: () => void;

    // The sizes of a page which can be chosen
    pageSizes = [10, 25, 50, 100];

    static override styles = css`
        :host {
            display: block;
            overflow-x: auto;
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
                    ${state.items.length === 0
                        ? html`<cds-table-row><cds-table-cell>${this.empty}</cds-table-cell></cds-table-row>`
                        : state.items.map((item) => html`
                            <cds-table-row ?data-selected=${state.selected === item} @click=${() => this.controller?.open(item)}>
                                ${this.columns.map((column) => html`<cds-table-cell>${column.cell(item)}</cds-table-cell>`)}
                            </cds-table-row>
                        `)}
                </cds-table-body>
            </cds-table>
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
