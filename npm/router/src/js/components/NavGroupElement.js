import { LitElement, html, css, nothing } from 'lit';

/**
 * @class NavGroupElement
 *
 * This class is used for a navigational group of elements
 *
 * @example
 * <c-nav-group>
 *   <c-nav-item>Nav 1</c-nav-item>
 *   <c-nav-item>Nav 2</c-nav-item>
 * </c-nav-group>
 */
export class NavGroupElement extends LitElement {
  static get localName() {
    return 'c-nav-group';
  }

  static get properties() {
    return {
    };
  }

  static get styles() {
    return css`
        :host {
            flex: 1 0;
            height: 100%;
        }
        ul {
            display: flex;
            list-style-type: none;
            padding: 0;
            margin: 0;
        }
        ul:has(*) {
          flex-direction: column;
        }
        ::slotted(c-nav-item) {
          flex: 0;
          margin: var(--nav-item-margin-y) var(--nav-item-margin-x);
          padding: var(--nav-item-padding-y) var(--nav-item-padding-x);

          color: var(--nav-item-color);
          background-color: var(--nav-item-background-color);
          border-radius: var(--nav-item-border-radius);          

          cursor: pointer;
          user-select: none;
        }

        ::slotted(c-nav-item:hover) {
          color: var(--nav-item-color-hover);
          background-color: var(--nav-item-background-color-hover);
        }

    `;
  }

  render() {
    return html`
        <ul class=${this.className || nothing}><slot></slot></ul>
    `;
  }

  get className() {
    const classes = [];
    return classes.join(' ');
  }
}