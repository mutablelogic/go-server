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
@customElement('wc-forminput')
export class FormInput extends LitElement {
  #internals: ElementInternals;
  @property({ type: String }) name: string;
  @property({ type: String }) value: string;
  @property({ type: Boolean }) disabled: boolean;
  @property({ type: Boolean }) required: boolean;
  @property({ type: Boolean }) autocomplete: boolean;

  constructor() {
    super();

    // Attach with the form
    this.#internals = this.attachInternals();
  }

  render() {
    return html`
      <label class=${this.className || nothing}>
        <slot></slot><br>
        <input 
          name=${this.name || nothing} 
          value=${this.value || nothing} 
          ?disabled=${this.disabled} 
          ?required=${this.required}
          ?autocomplete=${this.autocomplete}
          @input=${this.#onInput}>
      </label>
    `;
  }

  static get styles() {
    return css`
      :host {
        margin: var(--form-control-margin);
      }
      label {
        cursor: pointer;
        user-select: none;
        font-size: var(--form-input-font-size-label);
      }
      input {
        margin: var(--form-input-margin);
        padding: var(--form-input-padding);
        border-width: var(--form-input-border-width);
        border-color: var(--form-input-border-color);
        background-color: var(--form-input-background-color);
        width: 100%;
      }
    `;
  }

  get className() {
    const classes = [];
    if (this.disabled) {
      classes.push('disabled');
    }
    if (this.required) {
      classes.push('required');
    }
    if (this.autocomplete) {
      classes.push('autocomplete');
    }
    return classes.join(' ');
  }

  static get formAssociated() {
    return true;
  }

  // Form control properties
  get form() { return this.#internals ? this.#internals.form : null; }

  // Form control properties
  get type() { return this.localName; }

  // Form control properties
  get validity() { return this.#internals ? this.#internals.validity : null; }

  // Form control properties
  get validationMessage() { return this.#internals ? this.#internals.validationMessage : null; }

  // Form control properties
  get willValidate() { return this.#internals ? this.#internals.willValidate : null; }


  // Event hander for input event to update the value
  #onInput(event: Event) {
    if (!this.disabled) {
      this.value = event.target.value;
      this.#internals.setFormValue(this.value);
      return true;
    }
    return false;
  }
}
