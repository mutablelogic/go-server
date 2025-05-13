import { LitElement, html, css, nothing } from 'lit';
import { customElement, property } from 'lit/decorators.js';

/**
 * @class FormInput
 *
 * This class is a form input element, which should be wrapped in a form element.
 *
 * @example
 * <wc-forminput name="email" value="">...</wc-forminput>
 */
@customElement('wc-form')
export class Form extends LitElement {
  @property({ type: Boolean }) vertical: boolean;

  render() {
    return html`
      <form class=${this.className || nothing}>
        <slot></slot>
      </form>
    `;
  }


  static get styles() {
    return css`
      form {
        display: flex;
        flex-direction: row;
        &.vertical {
          flex-direction: column;
        }
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
