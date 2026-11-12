// A panel, which shows an item on the right of the page, with an optional
// header, a body, and an optional footer, of actions, such as Save and
// Cancel, which close it, which a controller
// keeps, and views, such as gs-panel, show

export { PanelController, type PanelListener, type PanelOptions } from './controller.ts';
export type { PanelAction, PanelActionKind, PanelActionState, PanelHeader, PanelState } from './schema.ts';
