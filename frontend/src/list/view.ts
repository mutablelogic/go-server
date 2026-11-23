// The gs-table element shows a list controller's page as a Carbon data
// table, with its columns, and pagination beneath it, which loads the page,
// or the page's size, which is changed, or, when it scrolls, the items which
// are loaded, which loads more as its end is scrolled to, beneath a header
// which stays at the top of what scrolls it. Its columns can have fixed
// widths, whose cells are cut short, of one line. A column which has a sort key sorts
// the list, by the server, when its header is clicked. Clicking a row opens
// its item in the controller, as does its first cell, which is a button, from
// the keyboard, and the item which is opened is selected. It
// shows a skeleton until it has items, and the items as unavailable while
// the server can't be reached. It can have a legend, to the right of the
// table, or beneath it, when the element is narrow, which explains what's in
// it, such as its tags. Its rows can be expanded, to show what's beneath an
// item, such as its history, when expand is set, and selected, for its batch
// actions, such as Delete, when actions are set. It has a toolbar, with a
// search, and a menu of settings, such as Show disabled, when they're set. It needs lit and
// @carbon/web-components, which are the package's optional peer
// dependencies.

import { LitElement, css, html, nothing } from 'lit';
import iconLoader from '@carbon/web-components/es/globals/internal/icon-loader.js';
import type { CarbonIconDescriptor } from '@carbon/web-components/es/globals/internal/icon-loader-utils.js';
import { checkmarkIcon, overflowIcon, settingsIcon } from '../internal/icons.ts';
import { runAction } from '../page/action.ts';
import type { ListController, ListState, Sort } from './controller.ts';

// Carbon components
import '@carbon/web-components/es/components/button/index.js';
import '@carbon/web-components/es/components/checkbox/index.js';
import '@carbon/web-components/es/components/data-table/index.js';
import '@carbon/web-components/es/components/overflow-menu/index.js';
import '@carbon/web-components/es/components/pagination/index.js';
import '@carbon/web-components/es/components/skeleton-text/index.js';

// Column is a column of the table: its header, its cell of an item, and
// the key the list is sorted by, by the server, when it can be, and which
// direction it's sorted in first, which is descending, unless it's set
// A column can have a width, such as 8rem, which makes the table's columns
// fixed, of their widths, where those without one share what's left, and
// its cells can be aligned to its end, such as of sizes.
export interface Column<T> {
    header: string;
    cell: (item: T) => unknown;
    sort?: string;
    sortFirst?: Sort['direction'];
    width?: string;
    align?: 'start' | 'end';
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

// TableRowAction is an action of a row's item, in its menu of actions, such
// as Delete: its label, whether it's dangerous, which is red, as Carbon's
// are, whether a divider is above it, whether it's disabled, and what it
// runs, with the item. When run throws, or its promise rejects,
// TableActionErrorEvent is fired.
export interface TableRowAction<T> {
    label: string;
    danger?: boolean;
    divider?: boolean;
    disabled?: boolean;
    run: (item: T) => void | PromiseLike<unknown>;
}

// TableToolbarAction is a button in a table's toolbar, such as Create: its
// label, its kind, which is primary by default, its icon, a 16px Carbon icon
// descriptor, or a rendered icon, whether it's disabled, and what it runs,
// when it's clicked. When run returns a promise, its button is disabled
// until it settles.
export interface TableToolbarAction {
    label: string;
    kind?: 'primary' | 'secondary' | 'tertiary' | 'ghost' | 'danger';
    icon?: unknown;
    disabled?: boolean;
    run: () => void | PromiseLike<unknown>;
}

// TableActionErrorEvent is fired, with the action and its error, when a
// batch action's, or a toolbar action's, run throws, or its promise rejects.
// A batch action's selection is kept.
export const TableActionErrorEvent = 'gs-table-action-error';

// TableActionError is the detail of a TableActionErrorEvent
export interface TableActionError<T = unknown> {
    action: TableAction<T> | TableToolbarAction | TableRowAction<T>;
    error: unknown;
}

// TableSelectionEvent is fired, with the selected items, when the
// selection changes
export const TableSelectionEvent = 'gs-table-selection';

// TableSelection is the detail of a TableSelectionEvent
export interface TableSelection<T = unknown> {
    items: T[];
}

// TableSetting is a setting in a table's toolbar's menu, such as Show
// disabled, which is checked when it's on, and can't be chosen when it's
// disabled, such as the last of the columns which are shown
export interface TableSetting {
    key: string;
    label: string;
    checked: boolean;
    disabled?: boolean;
}

// TableSearchEvent is fired, with what's typed in the toolbar's search,
// once typing pauses, so the app can filter the list
export const TableSearchEvent = 'gs-table-search';

// TableSearch is the detail of a TableSearchEvent
export interface TableSearch {
    value: string;
}

// TableSettingEvent is fired, with a setting's key and whether it's now
// checked, when it's chosen, so the app can set the settings again
export const TableSettingEvent = 'gs-table-setting';

// TableSettingChange is the detail of a TableSettingEvent
export interface TableSettingChange {
    key: string;
    checked: boolean;
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
        rowActions: { attribute: false },
        search: { type: String },
        settings: { attribute: false },
        toolbarActions: { attribute: false },
        scrolls: { type: Boolean, reflect: true },
        size: { type: String },
        toolbarSize: { type: String, attribute: 'toolbar-size' },
        countLabel: { attribute: false },
        running: { state: true },
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

    // rowActions returns the actions of a row's item, in a menu of them, of
    // Carbon's overflow menu, at the end of its row, which shows a column of
    // them, when it's set, and of none, of an item which has none
    declare rowActions?: (item: T) => TableRowAction<T>[];

    // search is the placeholder of a search in the toolbar, such as Search
    // certificates, which shows it, when it's set, and fires
    // TableSearchEvent, once typing pauses
    declare search?: string;

    // settings are the toolbar's menu of settings, which shows it, when
    // they're set, and fires TableSettingEvent when one is chosen
    declare settings?: TableSetting[];

    // toolbarActions are buttons at the toolbar's end, such as Create, which
    // show it, when they're set
    declare toolbarActions?: TableToolbarAction[];

    // scrolls is true of a table of a list controller which scrolls, which
    // has no pagination, and loads more as its end is scrolled to, in the
    // element which scrolls it, beneath a header which stays at its top, at
    // --gs-table-sticky-top
    declare scrolls: boolean;

    // size is the size of the table's rows, Carbon's, such as sm, or lg, by
    // default
    declare size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl';

    // toolbarSize is the size of the toolbar, and its search, settings,
    // and buttons, and of the batch actions bar, of Carbon's sizes of a
    // toolbar, which are xs, sm and lg, of 24, 32 and 48 pixels, which is of
    // the table's size by default, as Carbon's table's is: sm, or xs, of
    // those, and otherwise lg
    declare toolbarSize?: 'xs' | 'sm' | 'lg';

    // countLabel says how many items there are, beneath a table which
    // scrolls, once they're all loaded, such as 12 files
    declare countLabel?: (total: number) => string;

    // observer watches the end of a table which scrolls, to load more
    private observer?: IntersectionObserver;

    // running are the toolbar's actions whose promises are pending
    declare private running: Set<TableToolbarAction>;

    // searching waits for typing to pause, before a search is fired
    private searching?: ReturnType<typeof setTimeout>;

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

        /* The skeleton's rows, of more which are loading, aren't chosen */
        cds-table-row.skeleton {
            cursor: default;
            pointer-events: none;
        }

        cds-table-row.skeleton cds-skeleton-text {
            display: block;
            max-width: 75%;
        }

        /* The item which is opened */
        /* The checkboxes which select rows, for batch actions, are a narrow
           column */
        cds-table-header-cell.select,
        cds-table-cell.select {
            width: 3rem;
            padding-right: 0;
        }

        /* A row's menu of actions is a narrow column, at its end, whose
           menu is as tall as the row */
        cds-table-header-cell.row-actions,
        cds-table-cell.row-actions {
            box-sizing: border-box;
            width: 3rem;
            padding: 0;
            text-align: end;
        }

        cds-table-cell.row-actions cds-overflow-menu {
            block-size: 100%;
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

        /* A table which scrolls is scrolled by what it's in, rather than
           itself, or Carbon's table, so its header stays at the top */
        :host([scrolls]) .table {
            overflow: visible;
        }

        :host([scrolls]) cds-table::part(inner-container),
        :host([scrolls]) cds-table::part(content) {
            overflow: visible;
        }

        :host([scrolls]) cds-table-head {
            position: sticky;
            top: var(--gs-table-sticky-top, 0);
            z-index: 1;
        }

        /* A table of fixed columns, of their widths, whose cells are of one
           line, cut short, and which is at least as wide as they are, with
           --gs-table-column-width of each column without a width */
        cds-table.fixed::part(content) {
            table-layout: fixed;
            min-width: var(--gs-table-min-width, 0);
        }

        cds-table.fixed cds-table-cell,
        cds-table.fixed cds-table-header-cell {
            box-sizing: border-box;
            overflow: hidden;
            text-overflow: ellipsis;
        }

        /* A button is as wide as what's in it, even as a block, so it's
           as wide as its cell, to cut what's in it short */
        cds-table.fixed button.open {
            display: block;
            box-sizing: border-box;
            width: 100%;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
        }

        .end-aligned {
            text-align: end;
        }

        /* Beneath a table which scrolls: more which are loading, which
           couldn't be, or how many there are */
        .more.end {
            min-height: 0;
        }

        .more {
            display: flex;
            align-items: center;
            gap: 1rem;
            min-height: 2.5rem;
            padding: 0 1rem;
            color: var(--cds-text-secondary, #525252);
            font-size: 0.875rem;
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
        this.scrolls = false;
        this.expanded = new Set();
        this.selected = new Map();
        this.busy = false;
        this.running = new Set();
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
        clearTimeout(this.searching);
        this.observer?.disconnect();
        this.observer = undefined;
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

    // An expandable table's rows' chevrons point right, when they're
    // collapsed, and down, when they're expanded, and the end of a table
    // which scrolls is watched
    protected override updated() {
        if (this.expand) {
            this.renderRoot.querySelectorAll('cds-table-row, cds-table-header-row').forEach((row) => disclose(row));
        }
        this.watchEnd();
    }

    protected override render() {
        if (!this.legend?.items.length) {
            return html`<div class="layout"><div class="table">${this.renderToolbar()}${this.renderTable()}</div></div>`;
        }
        return html`
            <div class="layout legend">
                <div class="table">${this.renderToolbar()}${this.renderTable()}</div>
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
        const fixed = this.columns.some((column) => column.width);
        return html`
            <cds-table size=${this.size ?? 'lg'} ?expandable=${!!expand} class=${fixed ? 'fixed' : ''} style=${fixed ? `--gs-table-min-width: ${minWidth(this.columns, selectable, !!this.rowActions)}` : ''}>
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
                        ${this.rowActions ? html`<cds-table-header-cell class="row-actions"><span style=${visuallyHiddenStyle}>Actions</span></cds-table-header-cell>` : nothing}
                    </cds-table-header-row>
                </cds-table-head>
                <cds-table-body>
                    ${state.items.map((item) => {
                        const expanded = !!expand && this.expanded.has(this.keyOf(item));
                        return html`
                            <cds-table-row size=${this.size ?? 'lg'} ?data-selected=${state.selected === item} ?selected=${selectable && this.selected.has(this.keyOf(item))}
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
                                ${this.columns.map((column, index) => html`<cds-table-cell class=${column.align === 'end' ? 'end-aligned' : ''}>${index === 0 ? html`<button type="button" class="open">${column.cell(item)}</button>` : column.cell(item)}</cds-table-cell>`)}
                                ${this.rowActions ? this.renderRowActions(item) : nothing}
                            </cds-table-row>
                            ${expand ? html`
                                <cds-table-expanded-row ?expanded=${expanded} colspan=${this.columns.length + 1 + (selectable ? 1 : 0) + (this.rowActions ? 1 : 0)}>
                                    ${expanded ? expand(item) : nothing}
                                </cds-table-expanded-row>
                            ` : nothing}
                        `;
                    })}
                    ${this.scrolls && state.more && !state.failed ? this.renderSkeletonRows(selectable) : nothing}
                </cds-table-body>
            </cds-table>
            ${state.items.length === 0 && !state.failed ? html`<div class="empty">${this.empty}</div>` : nothing}
            ${this.scrolls ? this.renderMore(state) : state.total > 0 ? html`
                <cds-pagination .page=${state.page} .pageSize=${state.pageSize} .start=${(state.page - 1) * state.pageSize}
                    .totalItems=${state.total}
                    @cds-pagination-changed-current=${(event: CustomEvent<{ page: number }>) => this.controller?.setPage(event.detail.page)}
                    @cds-page-sizes-select-changed=${(event: CustomEvent<{ pageSize: number }>) => this.controller?.setPageSize(event.detail.pageSize)}>
                    ${this.pageSizes.map((size) => html`<cds-select-item value=${size}>${size}</cds-select-item>`)}
                </cds-pagination>
            ` : nothing}
        `;
    }

    // renderSkeletonRows returns rows of Carbon's skeleton, whose grey bars
    // are animated, at the end of a table which scrolls, while there are
    // more, which are loaded as they're scrolled to
    private renderSkeletonRows(selectable: boolean) {
        return [0, 1, 2].map(() => html`
            <cds-table-row size=${this.size ?? 'lg'} class="skeleton" aria-hidden="true">
                ${selectable ? html`<cds-table-cell class="select"></cds-table-cell>` : nothing}
                ${this.columns.map(() => html`<cds-table-cell><cds-skeleton-text></cds-skeleton-text></cds-table-cell>`)}
                ${this.rowActions ? html`<cds-table-cell class="row-actions"></cds-table-cell>` : nothing}
            </cds-table-row>
        `);
    }

    // renderMore returns what's beneath a table which scrolls: the end,
    // which loads more when it's scrolled to, while there are more, beneath
    // the skeleton's rows, or that they couldn't be, with Try again, or else
    // how many there are
    private renderMore(state: ListState<T>) {
        if (state.failed) {
            return html`
                <div class="more">
                    <span>${state.items?.length ? "More couldn't be loaded" : "They couldn't be loaded"}</span>
                    <cds-button kind="ghost" size="sm" @click=${() => void this.controller?.loadMore()}>Try again</cds-button>
                </div>
            `;
        } else if (state.more) {
            return html`<div class="more end" aria-busy=${state.loading ? 'true' : 'false'}></div>`;
        } else if (this.countLabel && state.items?.length) {
            return html`<div class="more">${this.countLabel(state.total)}</div>`;
        }
        return nothing;
    }

    // watchEnd watches the end of a table which scrolls, once it's
    // rendered, and loads more when it's scrolled to, within 200px
    private watchEnd() {
        this.observer?.disconnect();
        this.observer = undefined;
        const end = this.renderRoot.querySelector('.more.end');
        if (this.scrolls && end && typeof IntersectionObserver !== 'undefined') {
            this.observer = new IntersectionObserver((entries) => {
                if (entries.some((entry) => entry.isIntersecting)) {
                    void this.controller?.loadMore();
                }
            }, { rootMargin: '200px' });
            this.observer.observe(end);
        }
    }

    // renderToolbar returns the toolbar, with its search and settings, and
    // the batch actions bar over it, while items are selected. It's apart
    // from the table, so it stays while the table is loading, as when the
    // list is filtered by its search, which keeps what's typed, and focus.
    private renderToolbar() {
        const selectable = !!this.actions?.length;
        if (!(selectable && this.selected.size) && !this.search && !this.settings?.length && !this.toolbarActions?.length) {
            return nothing;
        }
        return html`
            <cds-table-toolbar size=${this.toolbarSizeOf()}>
                ${this.renderToolbarContent()}
                ${selectable ? html`
                    <cds-table-batch-actions size=${this.toolbarSizeOf()} ?active=${this.selected.size > 0} selected-rows-count=${this.selected.size}
                        @cds-table-batch-actions-cancel-clicked=${() => this.clearSelection()}>
                        ${this.actions!.map((action) => html`
                            <cds-button size=${this.toolbarSizeOf()} kind=${action.kind ?? 'primary'} ?disabled=${this.busy} @click=${() => this.run(action)}>
                                ${action.label}${action.icon ? renderIcon(action.icon) : nothing}
                            </cds-button>
                        `)}
                    </cds-table-batch-actions>
                ` : nothing}
            </cds-table-toolbar>
        `;
    }

    // renderToolbarContent returns the toolbar's search and settings, which
    // are outside Carbon's table, which would otherwise filter its rows
    // itself, rather than the app filtering the list
    private renderToolbarContent() {
        if (!this.search && !this.settings?.length && !this.toolbarActions?.length) {
            return nothing;
        }
        return html`
            <cds-table-toolbar-content>
                ${this.search ? html`
                    <cds-table-toolbar-search size=${this.toolbarSizeOf()} persistent placeholder=${this.search}
                        @cds-search-input=${(event: CustomEvent<{ value: string }>) => this.searched(event.detail.value)}></cds-table-toolbar-search>
                ` : nothing}
                ${this.settings?.length ? html`
                    <cds-overflow-menu size=${this.toolbarSizeOf()} toolbar-action>
                        ${settingsIcon}
                        <span slot="tooltip-content">Settings</span>
                        <cds-overflow-menu-body flipped>
                            ${this.settings.map((setting) => html`
                                <cds-overflow-menu-item ?disabled=${setting.disabled} @click=${() => {
                                    if (!setting.disabled) {
                                        this.dispatchEvent(new CustomEvent<TableSettingChange>(TableSettingEvent, {
                                            detail: { key: setting.key, checked: !setting.checked }, bubbles: true, composed: true,
                                        }));
                                    }
                                }}>
                                    <span style=${settingStyle}>
                                        <span style=${checkStyle}>${setting.checked ? checkmarkIcon : nothing}</span>${setting.label}<span style=${visuallyHiddenStyle}>, ${setting.checked ? 'on' : 'off'}</span>
                                    </span>
                                </cds-overflow-menu-item>
                            `)}
                        </cds-overflow-menu-body>
                    </cds-overflow-menu>
                ` : nothing}
                ${(this.toolbarActions ?? []).map((action) => html`
                    <cds-button size=${this.toolbarSizeOf()} kind=${action.kind ?? 'primary'} ?disabled=${action.disabled || this.running.has(action)} @click=${() => this.runToolbar(action)}>
                        ${action.label}${action.icon ? renderIcon(action.icon) : nothing}
                    </cds-button>
                `)}
            </cds-table-toolbar-content>
        `;
    }

    // toolbarSizeOf returns the size of the toolbar: its own, or of the
    // table's, sm, or xs, of those, and otherwise lg, as Carbon's table's is
    private toolbarSizeOf(): string {
        if (this.toolbarSize) {
            return this.toolbarSize;
        }
        const size = this.size ?? 'lg';
        return size === 'xs' || size === 'sm' ? size : 'lg';
    }

    // renderRowActions returns the menu of the actions of a row's item, of
    // Carbon's overflow menu, whose clicks don't open, or select, its row,
    // or nothing, of an item which has none
    private renderRowActions(item: T) {
        const actions = this.rowActions?.(item) ?? [];
        return html`
            <cds-table-cell class="row-actions" @click=${(event: Event) => event.stopPropagation()}>
                ${actions.length ? html`
                    <cds-overflow-menu size=${this.size ?? 'lg'} data-table>
                        ${overflowIcon}
                        <span slot="tooltip-content">Actions</span>
                        <cds-overflow-menu-body flipped>
                            ${actions.map((action) => html`
                                <cds-overflow-menu-item ?danger=${action.danger} ?divider=${action.divider} ?disabled=${action.disabled}
                                    @click=${() => this.runRow(action, item)}>${action.label}</cds-overflow-menu-item>
                            `)}
                        </cds-overflow-menu-body>
                    </cds-overflow-menu>
                ` : nothing}
            </cds-table-cell>
        `;
    }

    // runRow runs a row's action, with its item, and fires
    // TableActionErrorEvent when it fails, unless it's disabled
    private runRow(action: TableRowAction<T>, item: T) {
        if (action.disabled) {
            return;
        }
        void (async () => action.run(item))().catch((error: unknown) => {
            this.dispatchEvent(new CustomEvent<TableActionError<T>>(TableActionErrorEvent, { detail: { action, error }, bubbles: true, composed: true }));
        });
    }

    // runToolbar runs a toolbar action, while its button is disabled, and
    // fires TableActionErrorEvent when it fails. One which is disabled, or
    // running, isn't run, even when it's clicked again before its button is
    // rendered disabled.
    private runToolbar(action: TableToolbarAction) {
        if (action.disabled || this.running.has(action)) {
            return;
        }
        void runAction(action, (busy) => {
            const next = new Set(this.running);
            if (busy) {
                next.add(action);
            } else {
                next.delete(action);
            }
            this.running = next;
        }, (error) => {
            this.dispatchEvent(new CustomEvent<TableActionError<T>>(TableActionErrorEvent, { detail: { action, error }, bubbles: true, composed: true }));
        });
    }

    // searched fires TableSearchEvent with what's typed, once typing pauses
    private searched(value: string) {
        clearTimeout(this.searching);
        this.searching = setTimeout(() => {
            this.dispatchEvent(new CustomEvent<TableSearch>(TableSearchEvent, { detail: { value: value.trim() }, bubbles: true, composed: true }));
        }, 300);
    }

    // renderHeader returns a column's header, which sorts the list, by the
    // server, when it's clicked, when the column can be sorted. Carbon's
    // table, which sorts the rows on the page itself, isn't told.
    private renderHeader(column: Column<T>, sort?: Sort) {
        const style = column.width ? `width: ${column.width}` : '';
        const align = column.align === 'end' ? 'end-aligned' : '';
        if (column.sort === undefined) {
            return html`<cds-table-header-cell class=${align} style=${style}>${column.header}</cds-table-header-cell>`;
        }
        const key = column.sort;
        const first = column.sortFirst ?? 'descending';
        const direction = sort?.key === key ? sort.direction : 'none';
        return html`
            <cds-table-header-cell class=${align} style=${style} is-sortable sort-direction=${direction} sort-cycle=${first === 'descending' ? 'bi-states-from-descending' : 'bi-states-from-ascending'}
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
// disclosure turns a row's expand chevron, which Carbon points down when
// it's collapsed, and up when it's expanded, to point right, then down, as
// a disclosure does. Carbon's row doesn't expose its chevron as a part, so
// it's added to the row's shadow root.
const disclosure = css`
    .cds--table-expand__svg {
        transform: rotate(0deg) !important;
    }

    :host([expanded]) .cds--table-expand__svg {
        transform: rotate(90deg) !important;
    }
`;

// disclose adds the disclosure to a row's shadow root, once it's rendered,
// when constructable stylesheets are supported
function disclose(row: Element) {
    const sheet = disclosure.styleSheet;
    const ready = (row as Element & { updateComplete?: Promise<unknown> }).updateComplete ?? Promise.resolve();
    void ready.then(() => {
        const root = row.shadowRoot;
        if (sheet && root && !root.adoptedStyleSheets.includes(sheet)) {
            root.adoptedStyleSheets = [...root.adoptedStyleSheets, sheet];
        }
    });
}

// minWidth returns the least width of a table of fixed columns: their
// widths, and --gs-table-column-width, 12rem by default, of each without
// one, and of the checkboxes which select rows, and the menus of their
// actions
function minWidth<T>(columns: readonly Column<T>[], selectable: boolean, rowActions = false): string {
    const widths = columns.map((column) => column.width ?? 'var(--gs-table-column-width, 12rem)');
    if (selectable) {
        widths.push('3rem');
    }
    if (rowActions) {
        widths.push('3rem');
    }
    return `calc(${widths.join(' + ')})`;
}

// The styles of a setting in the toolbar's menu, which are inline, as
// Carbon moves the menu out of the table's shadow root, whose styles don't
// reach it: its checkmark when it's on, and space for one when it isn't,
// and whether it's on, which is said to assistive technology, but not shown
const settingStyle = 'display: inline-flex; align-items: center; gap: 0.5rem';
const checkStyle = 'display: inline-flex; width: 1rem; height: 1rem';
const visuallyHiddenStyle = 'position: absolute; width: 1px; height: 1px; overflow: hidden; clip: rect(0 0 0 0); white-space: nowrap';

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
        [TableSearchEvent]: CustomEvent<TableSearch>;
        [TableSettingEvent]: CustomEvent<TableSettingChange>;
    }
}
