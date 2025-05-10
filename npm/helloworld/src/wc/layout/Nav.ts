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

      /* Set cursor and flex for slotted elements */
      ul ::slotted(*) {
        cursor: pointer;
        user-select: none;
        padding: var(--space-default);
        align-items: center;
        justify-content: center;
      }

      ul ::slotted(wc-navspace) {
        cursor: default;
        flex: 1 0;
      }

      /* Hover */
      ul.vertical ::slotted(wc-navitem) {
        border-right: var(--nav-border-item);
      }

      ul.vertical ::slotted(wc-navitem:hover) {
        border-right: var(--nav-border-item-hover);
      }

      ul:not(.vertical) ::slotted(wc-navitem) {
        border-bottom: var(--nav-border-item);
      }

      ul:not(.vertical) ::slotted(wc-navitem:hover) {
        border-bottom: var(--nav-border-item-hover);
      }

      /* Active */
      ul ::slotted(wc-navitem:active) {
        font-weight: var(--nav-font-weight-active);
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
