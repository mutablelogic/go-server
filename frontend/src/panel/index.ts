// A panel, which shows an item on the right of the page, with an optional
// header, optional tabs, such as View, Edit and Delete, a body, of the tab
// which is chosen, and an optional footer, of its actions, such as Save and
// Cancel, which close it, which a controller keeps, and views, such as
// gs-panel, show

export { PanelController, type PanelListener, type PanelOptions } from './controller.ts';
export type { PanelAction, PanelActionKind, PanelActionState, PanelHeader, PanelState, PanelTab } from './schema.ts';
