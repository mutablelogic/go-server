import { LitElement, html, css, nothing } from 'lit';
import { customElement, property } from 'lit/decorators.js';

/**
 * @class NavItem
 *
 * This class is a container for a navigation item, within a nav element.
 *
 * @example
 * <wc-nav>
 *  <wc-navitem>....</wc-navitem>
 *  <wc-navitem>....</wc-navitem>
 * </wc-nav>
 */
@customElement('wc-navitem')
export class NavItem extends LitElement {
  @property({ type: Boolean }) selected: boolean = false;
  @property({ type: Boolean }) disabled: boolean = false;

  render() {
    return html`
      <li class=${this.className || nothing}>
        <slot></slot>
      </li>
    `;
  }

  static get styles() {
    return css`
      :host {
      }
    `;
  }

  get className() {
    const classes = [];
    if (this.selected) {
      classes.push('selected');
    }
    if (this.disabled) {
      classes.push('disabled');
    }
    return classes.join(' ');
  }
}
