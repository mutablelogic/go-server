// The panel controller keeps the state of a panel: the item it shows, while
// it's open, its header, and the actions in its footer, which it runs, and
// tells the views which subscribe to it, such as the gs-panel element,
// whenever it changes

import type { PanelAction, PanelActionState, PanelHeader, PanelState, PanelTab } from './schema.ts';

// PanelListener is called with the panel's state, whenever it changes
export type PanelListener<T> = (state: PanelState<T>) => void;

// PanelOptions are the panel's options
export interface PanelOptions {
    // header is its header, which can be changed with setHeader
    header?: PanelHeader;

    // closable is whether it can be closed, by its close button, the escape
    // key, or clicking the page beside it, which it can be by default
    closable?: boolean;

    // actions are the actions in its footer, which can be changed with
    // setActions, when it has no tabs
    actions?: PanelAction[];

    // tabs are its tabs, each of its own actions, of which tab is chosen,
    // or else the first which isn't disabled, which can be changed with
    // setTabs
    tabs?: PanelTab[];
    tab?: string;
}

export class PanelController<T = unknown> {
    private readonly listeners = new Set<PanelListener<T>>();
    private readonly closes = new Set<(item: T) => void>();
    private readonly errors = new Set<(error: unknown, action: PanelAction) => void>();
    private state: PanelState<T>;

    // pending is the action which is running, until what it returns settles,
    // which is kept apart from the footer's actions, so it's running while
    // they're set, or its tab's are, or its tab is removed
    private pending?: PanelAction;

    constructor(options: PanelOptions = {}) {
        this.state = {
            open: false,
            header: { ...options.header },
            closable: options.closable ?? true,
            tabs: [],
            actions: (options.actions ?? []).map((action) => ({ ...action, running: false })),
        };
        if (options.tabs?.length) {
            this.setTabs(options.tabs, options.tab);
        }
    }

    // item is the item which is shown, while the panel is open, or the one
    // it showed, while it closes
    get item(): T | undefined {
        return this.state.item;
    }

    // isOpen is whether the panel is open
    get isOpen(): boolean {
        return this.state.open;
    }

    // current is the panel's state
    get current(): PanelState<T> {
        return this.state;
    }

    // open shows the item, and opens the panel, when it's closed
    open(item: T) {
        this.update({ open: true, item });
    }

    // close closes the panel, which keeps its item while it closes, and
    // calls the close functions with it
    close() {
        if (!this.state.open) {
            return;
        }
        this.update({ open: false });
        const item = this.state.item;
        if (item !== undefined) {
            this.closes.forEach((fn) => fn(item));
        }
    }

    // setHeader sets the panel's header, such as its title
    setHeader(header: PanelHeader) {
        this.update({ header: { ...header } });
    }

    // setClosable sets whether the panel can be closed
    setClosable(closable: boolean) {
        this.update({ closable });
    }

    // setActions sets the actions in the panel's footer, which is left out
    // when there are none, which are the chosen tab's, when it has tabs, so
    // they're its actions while it's chosen again, such as Delete, which is
    // enabled once the name of what's deleted is typed. An action which is
    // running keeps running.
    setActions(actions: PanelAction[]) {
        const tabs = this.state.tabs.map((tab) => tab.value === this.state.tab ? { ...tab, actions: [...actions] } : tab);
        this.update({ tabs, actions: this.states(actions) });
    }

    // tab is the value of the tab which is chosen, or undefined, when the
    // panel has no tabs
    get tab(): string | undefined {
        return this.state.tab;
    }

    // setTabs sets the panel's tabs, of which the tab of the value is chosen,
    // or else the one which is chosen now, when it's one of them, or else
    // the first which isn't disabled, and the footer's actions are its. No
    // tabs are shown when there are none.
    setTabs(tabs: PanelTab[], tab?: string) {
        const copies = tabs.map((t) => ({ ...t, actions: t.actions ? [...t.actions] : undefined }));
        const chosen = [tab, this.state.tab].find((value) => copies.some((t) => t.value === value && !t.disabled))
            ?? copies.find((t) => !t.disabled)?.value;
        const actions = copies.find((t) => t.value === chosen)?.actions ?? (copies.length ? [] : this.state.actions);
        this.update({ tabs: copies, tab: chosen, actions: this.states(actions) });
    }

    // selectTab chooses the tab of the value, whose actions are the footer's,
    // unless it's disabled, or isn't one of the tabs, or an action is
    // running, as it would be of another tab. Returns true when it's chosen.
    selectTab(value: string): boolean {
        const tab = this.state.tabs.find((t) => t.value === value);
        if (!tab || tab.disabled || this.running) {
            return false;
        }
        if (value !== this.state.tab) {
            this.update({ tab: value, actions: this.states(tab.actions ?? []) });
        }
        return true;
    }

    // run runs the action, as when it's clicked, unless it's disabled, or
    // an action is running, until what it returns settles, and then closes
    // the panel, unless close is false. When it fails, the panel stays open,
    // and the error functions are called. Returns true when it ran, and
    // didn't fail.
    async run(action: PanelAction): Promise<boolean> {
        const state = this.state.actions.find((state) => same(state, action));
        if (!state || state.disabled || this.running) {
            return false;
        }
        if (state.run) {
            this.pending = state;
            this.setRunning(action, true);
            try {
                await state.run();
            } catch (error) {
                this.errors.forEach((fn) => fn(error, action));
                return false;
            } finally {
                // It's done, so the panel isn't running, which its listeners
                // are told, even when the action isn't one of its actions any
                // more, such as when it replaced them, so their buttons aren't
                // left disabled
                this.pending = undefined;
                this.update({ actions: this.states(this.state.actions) });
            }
        }
        if (state.close !== false) {
            this.close();
        }
        return true;
    }

    // running is whether an action is running, while which the others can't
    // be run, such as Cancel while Save is running
    get running(): boolean {
        return this.pending !== undefined;
    }

    // subscribe calls the function with the panel's state, now, and whenever
    // it changes, until the function it returns is called, which
    // unsubscribes
    subscribe(fn: PanelListener<T>): () => void {
        this.listeners.add(fn);
        fn(this.state);
        return () => this.listeners.delete(fn);
    }

    // onClose calls the function with the item when the panel is closed
    onClose(fn: (item: T) => void): this {
        this.closes.add(fn);
        return this;
    }

    // onError calls the function when an action fails, with its error
    onError(fn: (error: unknown, action: PanelAction) => void): this {
        this.errors.add(fn);
        return this;
    }

    // states are the actions' states, of which the one which is pending is
    // running
    private states(actions: readonly PanelAction[]): PanelActionState[] {
        const pending = this.pending;
        return actions.map((action) => ({ ...action, running: pending !== undefined && same(pending, action) }));
    }

    // setRunning sets that the action is running, when it's still one of
    // the panel's actions
    private setRunning(action: PanelAction, running: boolean) {
        if (this.state.actions.some((state) => same(state, action) && state.running !== running)) {
            this.update({ actions: this.state.actions.map((state) => same(state, action) ? { ...state, running } : state) });
        }
    }

    private update(state: Partial<PanelState<T>>) {
        this.state = { ...this.state, ...state };
        this.listeners.forEach((fn) => fn(this.state));
    }
}

// same returns true when the actions are the same action, by what they run,
// and their label
function same(a: PanelAction, b: PanelAction): boolean {
    return a.run === b.run && a.label === b.label;
}
