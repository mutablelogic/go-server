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

// PanelActionState is an action, and whether it's running, as it was clicked,
// and what it returned hasn't settled
export interface PanelActionState extends PanelAction {
    running: boolean;
}

// PanelState is what a view of the panel shows: whether it's open, its
// item, which is kept while it closes, its header, whether it can be closed,
// by its close button, the escape key, or clicking the page beside it, and
// the actions in its footer
export interface PanelState<T> {
    open: boolean;
    item?: T;
    header: PanelHeader;
    closable: boolean;
    actions: readonly PanelActionState[];
}
