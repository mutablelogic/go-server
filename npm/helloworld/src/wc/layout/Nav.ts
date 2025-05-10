import { LitElement, html, css, nothing } from 'lit';
import { customElement, property } from 'lit/decorators.js';

/**
 * @class Nav
 *
 * This class is a container for vertical or horizontal navigation.
 *
 * @example
 * <wc-canvas vertical>
 *  <wc-nav>....</wc-nav>
 *  <wc-content>....</wc-content>
 *  <wc-content>....</wc-content>
 * </wc-canvas>
 */
@customElement('wc-nav')
export class Nav extends LitElement {
  @property({ type: Boolean }) vertical: boolean = false;

  render() {
    return html`
      <ul class=${this.className || nothing}>
        <slot></slot>
      </ul>
    `;
  }

  static get styles() {
    return css`
      :host {
        background-color: var(--nav-background-color, #eee);
      }

      ul {
        flex: 1 0;
        display: flex;
        padding: 0;
        margin: 0;        
        list-style-type: none;

        &.vertical {
          flex-direction: column;
          height: 100%;
          border-right: var(--nav-border);
        }

        &:not(.vertical) {
          flex-direction: row;
          border-bottom: var(--nav-border);
        }
      }

      ::slotted(*) {
        cursor: pointer;
        user-select: none;
        padding: var(--space-default);
        background-color: var(--nav-background-color);
        color: var(--nav-text-color);
      }

      ::slotted(:not(wc-navspace):hover) {
          background-color: var(--nav-background-color-hover);
          color: var(--nav-text-color-hover);
      }


      ::slotted(wc-navspace) {
        cursor: default;
        flex: 1 0;
        background-color: inherit;
        color: inherit;
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
