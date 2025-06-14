import { LitElement, html, css, nothing } from 'lit';
import { customElement, property } from 'lit/decorators.js';

/**
 * @class Content
 *
 * This class is a container for content, within a canvas or content element.
 *
 * @example
 * <wc-canvas vertical>
 *  <wc-nav>....</wc-nav>
 *  <wc-content>....</wc-content>
 *  <wc-content>....</wc-content>
 * </wc-canvas>
 */
@customElement('wc-content')
export class Content extends LitElement {
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
      :host {
          flex: 1 0;
      }

      div {
        display: flex;
        flex: 1 0;
      }

      div.vertical {
        flex-direction: column;
      }

      div:not(.vertical) {
        flex-direction: row;
      }

      /* Content is in flexbox mode with a fixed size */
      ::slotted(wc-content) {
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
