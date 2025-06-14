import { LitElement, html, nothing } from 'lit';
import { customElement, property } from 'lit/decorators.js';

/**
 * @class NavSpace
 *
 * This class is a container for a navigation spacer, within a nav element.
 *
 * @example
 * <wc-nav>
 *  <wc-navitem>....</wc-navitem>
 *  <wc-navspace>....</wc-navspace>
 *  <wc-navitem>....</wc-navitem>
 * </wc-nav>
 */
@customElement('wc-navspace')
export class NavSpace extends LitElement {
  render() {
    return html`
      <li></li>
    `;
  }
}
