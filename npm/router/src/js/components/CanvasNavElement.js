import { html, css, nothing } from 'lit';
import { CanvasContentElement } from './CanvasContentElement.js';

/**
 * @class CanvasNavElement
 *
 * This class is a navigaton group used to contain content boxes which are stacked
 * vertically or horizontally within the canvas.
 *
 * @property {Boolean} hidden - Whether the navigation is hidden, default false
 *
 * @example
 * <c-canvas vertical>
 *  <c-canvas-nav>....</c-canvas-nav>
 *  <c-canvas-content>....</c-canvas-content>
 *  <c-canvas-nav>....</c-canvas-nav>
 * </c-canvas>
 */
export class CanvasNavElement extends CanvasContentElement {
  static get localName() {
    return 'c-canvas-nav';
  }

  constructor() {
    super();
    this.hidden = false;
  }

  static get properties() {
    return {
      hidden: { type: Boolean },
    };
  }

  static get styles() {
    return [
      CanvasContentElement.styles,
      css`
        :host {
          display: flex;
          align-items: start;
        }
    `];
  }

  render() {
    return html`
      <div class=${this.classes.join(' ') || nothing}>
        <slot></slot>
      </div>
    `;
  }

  get classes() {
    const classes = [];
    if (this.hidden) {
      classes.push('hidden');
    }
    return classes;
  }
}