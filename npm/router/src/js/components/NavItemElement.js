import { LitElement, html, css, nothing } from 'lit';

/**
 * @class NavItemElement
 *
 * This class is used for a navigational group of elements
 *
 * @example
 * <c-nav-group>
 *   <c-nav-item>Nav 1</c-nav-item>
 *   <c-nav-item>Nav 2</c-nav-item>
 * </c-nav-group>
 */
export class NavItemElement extends LitElement {
  static get localName() {
    return 'c-nav-item';
  }

  static get properties() {
    return {
    };
  }

  render() {
    return html`
        <li class=${this.className || nothing}><slot></slot></li>
    `;
  }

  get className() {
    const classes = [];
    return classes.join(' ');
  }
}
