// The gs-table element shows a list controller's page as a Carbon data
// table, with its columns, and pagination beneath it, which loads the page,
// or the page's size, which is changed. A column which has a sort key sorts
// the list, by the server, when its header is clicked. Clicking a row opens
// its item in the controller, as does its first cell, which is a button, from
// the keyboard, and the item which is opened is selected. It
// shows a skeleton until it has items, and the items as unavailable while
// the server can't be reached. It can have a legend, to the right of the
// table, or beneath it, when the element is narrow, which explains what's in
// it, such as its tags. Its rows can be expanded, to show what's beneath an
// item, such as its history, when expand is set, and selected, for its batch
// actions, such as Delete, when actions are set. It needs lit and
// @carbon/web-components, which are the package's optional peer
// dependencies.

import { LitElement, css, html, nothing } from 'lit';
import iconLoader from '@carbon/web-components/es/globals/internal/icon-loader.js';
import type { CarbonIconDescriptor } from '@carbon/web-components/es/globals/internal/icon-loader-utils.js';
import { runAction } from '../page/action.ts';
import type { ListController, ListState, Sort } from './controller.ts';

// Carbon components
import '@carbon/web-components/es/components/button/index.js';
import '@carbon/web-components/es/components/checkbox/index.js';
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

// TableAction is a batch action of a table's selected items, such as
// Delete: its label, its kind, which is primary by default, its icon, a
// 16px Carbon icon descriptor, or a rendered icon, and what it runs, with
// the items. When run returns a promise, the actions are disabled until it
// settles. The selection is cleared when it succeeds.
export interface TableAction<T> {
    label: string;
    kind?: 'primary' | 'secondary' | 'tertiary' | 'ghost' | 'danger';
    icon?: unknown;
    run: (items: T[]) => void | PromiseLike<unknown>;
}

// TableActionErrorEvent is fired, with the action and its error, when a
// batch action's run throws, or its promise rejects, and the selection is
// kept
export const TableActionErrorEvent = 'gs-table-action-error';

// TableActionError is the detail of a TableActionErrorEvent
export interface TableActionError<T = unknown> {
    action: TableAction<T>;
    error: unknown;
}

// TableSelectionEvent is fired, with the selected items, when the
// selection changes
export const TableSelectionEvent = 'gs-table-selection';

// TableSelection is the detail of a TableSelectionEvent
export interface TableSelection<T = unknown> {
    items: T[];
}

// TableTag is the element's tag
export const TableTag = 'gs-table';

export class TableElement<T = unknown> extends LitElement {
    static override properties = {
        controller: { attribute: false },
        columns: { attribute: false },
        legend: { attribute: false },
        expand: { attribute: false },
        itemKey: { attribute: false },
        actions: { attribute: false },
        empty: { type: String },
        state: { state: true },
        expanded: { state: true },
        selected: { state: true },
        busy: { state: true },
    };

    // controller is the list controller whose page is shown
    declare controller?: ListController<T>;

    // columns are the table's columns
    declare columns: Column<T>[];

    // legend explains what's in the table, beside it, when it's set
    declare legend?: Legend;

    // expand returns what's shown beneath an item's row, when it's expanded,
    // such as its history, and makes the rows expandable, when it's set. It's
    // only called for a row which is expanded, so what it shows can be loaded
    // then, such as with lit's until directive.
    declare expand?: (item: T) => unknown;

    // itemKey identifies an item, so its row stays expanded, or selected,
    // when the page is loaded again, and the selection is kept across pages.
    // Without it, an item is itself, and only what's on the page is expanded,
    // or selected, as a page which is loaded again has new items.
    declare itemKey?: (item: T) => string;

    // actions are the batch actions of the selected items, which make the
    // rows selectable, with a checkbox, when they're set. The selection is
    // kept across pages, and shown, with the actions, above the table.
    declare actions?: TableAction<T>[];

    // empty is what's shown when there are no items, such as No roles
    declare empty: string;
    declare private state?: ListState<T>;

    // expanded are the expanded rows' items, or their keys
    declare private expanded: Set<unknown>;

    // selected are the selected items, by their keys, or themselves, and
    // busy is true while a batch action runs
    declare private selected: Map<unknown, T>;
    declare private busy: boolean;
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
            background: var(--cds-layer-01, #f4f4f4);
            color: var(--cds-text-primary, #161616);
            font-size: 0.875rem;
            line-height: 1.28572;
        }

        /* The legend's title is a bar, as the table's header is */
        aside h2 {
            display: flex;
            align-items: center;
            min-height: 3rem;
            margin: 0;
            padding: 0 1rem;
            background: var(--cds-layer-accent-01, #e0e0e0);
            font-size: 0.875rem;
            font-weight: 600;
        }

        aside dl {
            margin: 0;
            padding: 1rem;
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
        /* The checkboxes which select rows, for batch actions, are a narrow
           column */
        cds-table-header-cell.select,
        cds-table-cell.select {
            width: 3rem;
            padding-right: 0;
        }

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
        this.expanded = new Set();
        this.selected = new Map();
        this.busy = false;
    }

    // selection is the selected items, which batch actions run with
    get selection(): T[] {
        return [...this.selected.values()];
    }

    // clearSelection deselects every item
    clearSelection() {
        this.select(new Map());
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
        // An expandable table's rows have a button which expands them, which
        // doesn't open the row's item, and each is followed by a row which
        // shows what's beneath it, when it's expanded. A table with batch
        // actions has a checkbox in each row, and one in the header, which
        // selects the page, and the actions above it, while there's a
        // selection.
        const expand = this.expand;
        const selectable = !!this.actions?.length;
        const items = state.items;
        const onPage = items.filter((item) => this.selected.has(this.keyOf(item))).length;
        return html`
            ${selectable && this.selected.size ? html`
                <cds-table-toolbar>
                    <cds-table-batch-actions active selected-rows-count=${this.selected.size}
                        @cds-table-batch-actions-cancel-clicked=${() => this.clearSelection()}>
                        ${this.actions!.map((action) => html`
                            <cds-button kind=${action.kind ?? 'primary'} ?disabled=${this.busy} @click=${() => this.run(action)}>
                                ${action.label}${action.icon ? renderIcon(action.icon) : nothing}
                            </cds-button>
                        `)}
                    </cds-table-batch-actions>
                </cds-table-toolbar>
            ` : nothing}
            <cds-table ?expandable=${!!expand}>
                <cds-table-head>
                    <cds-table-header-row ?expandable=${!!expand}>
                        ${selectable ? html`
                            <cds-table-header-cell class="select">
                                <cds-checkbox hide-label ?checked=${onPage > 0 && onPage === state.items.length}
                                    ?indeterminate=${onPage > 0 && onPage < state.items.length} ?disabled=${!state.items.length}
                                    @cds-checkbox-changed=${(event: CustomEvent<{ checked: boolean }>) => this.selectPage(items, event.detail.checked)}>Select all</cds-checkbox>
                            </cds-table-header-cell>
                        ` : nothing}
                        ${this.columns.map((column) => this.renderHeader(column, state.sort))}
                    </cds-table-header-row>
                </cds-table-head>
                <cds-table-body>
                    ${state.items.map((item) => {
                        const expanded = !!expand && this.expanded.has(this.keyOf(item));
                        return html`
                            <cds-table-row ?data-selected=${state.selected === item} ?selected=${selectable && this.selected.has(this.keyOf(item))}
                                ?expandable=${!!expand} ?expanded=${expanded}
                                @click=${(event: Event) => {
                                    if (!expanderClicked(event)) {
                                        this.controller?.open(item);
                                    }
                                }}
                                @cds-table-row-expando-toggled=${(event: CustomEvent<{ expanded: boolean }>) => this.toggle(item, event.detail.expanded)}>
                                ${selectable ? html`
                                    <cds-table-cell class="select" @click=${(event: Event) => event.stopPropagation()}>
                                        <cds-checkbox hide-label ?checked=${this.selected.has(this.keyOf(item))}
                                            @cds-checkbox-changed=${(event: CustomEvent<{ checked: boolean }>) => this.selectItem(item, event.detail.checked)}>Select</cds-checkbox>
                                    </cds-table-cell>
                                ` : nothing}
                                ${this.columns.map((column, index) => html`<cds-table-cell>${index === 0 ? html`<button type="button" class="open">${column.cell(item)}</button>` : column.cell(item)}</cds-table-cell>`)}
                            </cds-table-row>
                            ${expand ? html`
                                <cds-table-expanded-row ?expanded=${expanded} colspan=${this.columns.length + 1 + (selectable ? 1 : 0)}>
                                    ${expanded ? expand(item) : nothing}
                                </cds-table-expanded-row>
                            ` : nothing}
                        `;
                    })}
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

    // prune keeps only the items on the page expanded, or selected, and fires
    // TableSelectionEvent when the selection changes
    private prune(items: readonly T[]) {
        const present = new Set<unknown>(items);
        if ([...this.expanded].some((item) => !present.has(item))) {
            this.expanded = new Set([...this.expanded].filter((item) => present.has(item)));
        }
        if ([...this.selected.keys()].some((item) => !present.has(item))) {
            this.select(new Map([...this.selected].filter(([item]) => present.has(item))));
        }
    }

    // selectItem selects or deselects an item
    private selectItem(item: T, selected: boolean) {
        const next = new Map(this.selected);
        if (selected) {
            next.set(this.keyOf(item), item);
        } else {
            next.delete(this.keyOf(item));
        }
        this.select(next);
    }

    // selectPage selects or deselects the page's items, and keeps the
    // selection of the other pages
    private selectPage(items: readonly T[], selected: boolean) {
        const next = new Map(this.selected);
        for (const item of items) {
            if (selected) {
                next.set(this.keyOf(item), item);
            } else {
                next.delete(this.keyOf(item));
            }
        }
        this.select(next);
    }

    // select sets the selection, and fires TableSelectionEvent
    private select(selected: Map<unknown, T>) {
        this.selected = selected;
        this.dispatchEvent(new CustomEvent<TableSelection<T>>(TableSelectionEvent, { detail: { items: this.selection }, bubbles: true, composed: true }));
    }

    // run runs a batch action with the selected items, while the actions are
    // disabled, and clears the selection when it succeeds, or else fires
    // TableActionErrorEvent, and keeps it
    private run(action: TableAction<T>) {
        if (this.busy) {
            return;
        }
        const items = this.selection;
        let failed = false;
        void runAction({ label: action.label, run: () => action.run(items) }, (busy) => { this.busy = busy; }, (error) => {
            failed = true;
            this.dispatchEvent(new CustomEvent<TableActionError<T>>(TableActionErrorEvent, { detail: { action, error }, bubbles: true, composed: true }));
        }).then(() => {
            if (!failed) {
                this.clearSelection();
            }
        });
    }

    // keyOf is what identifies an item's row as expanded: its key, or else
    // the item
    private keyOf(item: T): unknown {
        return this.itemKey ? this.itemKey(item) : item;
    }

    // toggle expands or collapses an item's row
    private toggle(item: T, expanded: boolean) {
        const next = new Set(this.expanded);
        if (expanded) {
            next.add(this.keyOf(item));
        } else {
            next.delete(this.keyOf(item));
        }
        this.expanded = next;
    }

    // watch shows the controller's page, while the element is in the page,
    // and keeps the selected items which are on it current
    private watch() {
        this.unsubscribe?.();
        this.unsubscribe = this.isConnected ? this.controller?.subscribe((state) => {
            this.state = state;

            // Without keys, an item is itself, and a page which is loaded
            // again has new ones, so what's expanded, or selected, is only
            // what's on the page, rather than items which are gone, which
            // would otherwise be kept forever
            if (!this.itemKey && state.items) {
                this.prune(state.items);
                return;
            }
            if (this.selected.size && state.items?.some((item) => this.selected.has(this.keyOf(item)))) {
                const next = new Map(this.selected);
                for (const item of state.items) {
                    if (next.has(this.keyOf(item))) {
                        next.set(this.keyOf(item), item);
                    }
                }
                this.selected = next;
            }
        }) : undefined;
        if (!this.unsubscribe) {
            this.state = undefined;
        }
    }
}

// renderIcon renders a Carbon icon descriptor, in a button's icon slot, or
// else an icon which is rendered already
function renderIcon(icon: unknown) {
    if (typeof icon === 'object' && icon !== null && 'elem' in icon) {
        return iconLoader(icon as CarbonIconDescriptor, { slot: 'icon' });
    }
    return html`<span slot="icon">${icon}</span>`;
}

// expanderClicked is true when the event is a click on a row's expand
// button, which isn't a click on the row
function expanderClicked(event: Event): boolean {
    return event.composedPath().some((node) => node instanceof HTMLElement && node.classList.contains('cds--table-expand__button'));
}

if (!customElements.get(TableTag)) {
    customElements.define(TableTag, TableElement);
}

declare global {
    interface HTMLElementTagNameMap {
        [TableTag]: TableElement;
    }
    interface HTMLElementEventMap {
        [TableActionErrorEvent]: CustomEvent<TableActionError>;
        [TableSelectionEvent]: CustomEvent<TableSelection>;
    }
}
