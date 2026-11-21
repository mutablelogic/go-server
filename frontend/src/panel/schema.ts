// The state of a panel, which a panel controller keeps, and a view shows

// PanelHeader is the panel's header: its title, its subtitle, and its icon,
// which is a URL, such as of a logo, each of which is left out when it isn't
// set
export interface PanelHeader {
    title?: string;
    subtitle?: string;
    icon?: string;
}

// PanelActionKind is the kind of an action's button, as Carbon's
export type PanelActionKind = 'primary' | 'secondary' | 'tertiary' | 'ghost' | 'danger';

// PanelAction is an action in the panel's footer, such as Save, or Cancel,
// which runs when it's clicked, until what it returns settles, and then
// closes the panel, unless it fails, or close is false. An action without
// run, such as Cancel, closes the panel.
export interface PanelAction {
    label: string;
    kind?: PanelActionKind;
    disabled?: boolean;
    close?: boolean;
    run?: () => unknown;
}

// PanelTab is a tab of the panel, beneath its header, such as View, Edit or
// Delete, which is chosen by its value, whose body the panel's content
// shows, of the state's tab, and whose actions are the footer's, while it's
// chosen. A tab which is disabled can't be chosen.
export interface PanelTab {
    value: string;
    label: string;
    disabled?: boolean;
    actions?: PanelAction[];
}

// PanelActionState is an action, and whether it's running, as it was clicked,
// and what it returned hasn't settled
export interface PanelActionState extends PanelAction {
    running: boolean;
}

// PanelState is what a view of the panel shows: whether it's open, its
// item, which is kept while it closes, its header, whether it can be closed,
// by its close button, the escape key, or clicking the page beside it, its
// tabs, of which none are shown when there are none, the value of the one
// which is chosen, and the actions in its footer, which are the chosen
// tab's, when it has tabs
export interface PanelState<T> {
    open: boolean;
    item?: T;
    header: PanelHeader;
    closable: boolean;
    tabs: readonly PanelTab[];
    tab?: string;
    actions: readonly PanelActionState[];
}
