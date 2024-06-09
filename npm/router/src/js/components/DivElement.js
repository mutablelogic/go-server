import { LitElement, html, css, nothing } from 'lit';

/**
  * @class DivElement
  *
  * This class is used to contain text or other elements
  *
  * @example
  * <c-div>Text</c-div>
  */
export class DivElement extends LitElement {
  static get localName() {
    return 'c-div';
  }

  get classes() {
    const classes = [];
    return classes;
  }

  static get styles() {
    return css`
      :host {
        display: block;
        vertical-align: middle;
        padding: var(--div-padding-y) var(--div-padding-x);
      }
  `;
  }

  render() {
    return html`<div class=${this.classes.join(' ') || nothing}><slot></slot></div>`;
  }
}