import { LitElement, html, css } from 'lit';
import Model from '../core/model';

/**
 * A navigation bar which can be made sticky
 *
 * @slot - This element has a slot for wc-nav-item elements
 */
window.customElements.define('wc-navbar', class extends LitElement {
  static get properties() {
    return {
      /**
       * Nav is displayed in a column
       * @type {boolean}
       */
      column: { type: Boolean },
    };
  }

  static get styles() {
    return css`
      :host {
        --navitem-color-hover: var(--white-color);
      }
      :host nav {
        position: relative;
        padding: var(--navbar-padding);
        background-color: var(--navbar-background-color);
        color: var(--navbar-color);
        border-bottom: var(--navbar-border-bottom);
      }
      :host nav slot {
        display: flex;
        flex-flow: row wrap;
        justify-content: space-between;
        align-items: center;
      }
      `;
  }

  // eslint-disable-next-line class-methods-use-this
  add(model) {
    const elementId = Model.toElementId(this, model.key);
    const existingElement = this.querySelector(`#${elementId}`);
    if (existingElement) {
      return existingElement;
    }

    const template = document.createElement('wc-nav-item');
    template.id = elementId;
    this.appendChild(template);
    return template;
  }

  // eslint-disable-next-line class-methods-use-this
  change(model) {
    return this.add(model);
  }

  // eslint-disable-next-line class-methods-use-this
  delete(model) {
    const elementId = Model.toElementId(this, model.key);
    const existingElement = this.querySelector(`#${elementId}`);
    if (existingElement) {
      this.removeChild(existingElement);
    }
    return existingElement;
  }

  // eslint-disable-next-line class-methods-use-this
  render() {
    return html`
        <nav><slot></slot></nav>
      `;
  }
});
