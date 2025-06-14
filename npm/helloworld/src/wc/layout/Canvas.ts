import { LitElement, html, css, nothing } from 'lit';
import { customElement, property } from 'lit/decorators.js';

/**
 * @class Canvas
 *
 * This class is used to contain  content boxes which are stacked
 * vertically or horizontally within the canvas.
 *
 * @property {Boolean} vertical - Fill the canvas vertically rather than horizontally, default false
 *
 * @example
 * <wc-canvas vertical>
 *  <wc-nav>....</wc-nav>
 *  <wc-content>....</wc-content>
 *  <wc-content>....</wc-content>
 * </wc-canvas>
 */
@customElement('wc-canvas')
export class Canvas extends LitElement {
  @property({ type: Boolean }) vertical: boolean = false;

  render() {
    return html`
      <div class=${this.className || nothing}>
        <slot></slot>
      </div>
    `;
  }

  static get styles() {
    return css`
      div {
        position: absolute;
        top: 0;
        bottom: 0;
        left: 0;
        right: 0;
        display: flex;
      }

      div.vertical {
        flex-direction: column;
      }

      div:not(.vertical) {
        flex-direction: row;
      }

      /* Content is in flexbox mode with a fixed size */
      ::slotted(wc-content), ::slotted(wc-nav) {
        display: flex;
        justify-content: start;
      }
    `;
  }

  get className() {
    const classes = [];
    if (this.vertical) {
      classes.push('vertical');
    }
    return classes.join(' ');
  }
}
