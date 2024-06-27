import { LitElement, html, nothing, css } from 'lit';

/**
 * @class FormControlElement
 *
 * This class is used as a base class for all form elements
 *
 * @property {String} name - The name of the switch
 * @property {String} value - The value of the control
 * @property {Boolean} disabled - Whether the form control is disabled
 * @property {Boolean} required - Whether the form control is required before submitting
 * @property {Boolean} autocomplete - Whether the form control allows autocomplete
 *
 * @example
 * <c-form-control>Power</c-form-control>
 */
export class FormControlElement extends LitElement {
  static get localName() {
    return 'c-form-control';
  }

  constructor() {
    super();

    // Attach with the form
    this.internals = this.attachInternals();

    // Default properties
    this.name = '';
    this.value = '';
    this.disabled = false;
    this.required = false;
    this.autocomplete = false;
  }

  static get formAssociated() {
    return true;
  }

  static get properties() {
    return {
      name: { type: String },
      value: { type: String },
      disabled: { type: Boolean },
      required: { type: Boolean },
      autocomplete: { type: Boolean },
    };
  }

  static get styles() {
    return css`
      :host {
        display: inline-block;
        flex-shrink: 0;
        padding: var(--form-control-padding-y) var(--form-control-padding-x);
      }
      label {
        cursor: pointer;
        user-select: none;
      }
      label.switch {        
        & input {
          width: var(--form-switch-width);
          height: var(--form-switch-height);
          border-radius: var(--form-switch-border-radius);
          vertical-align: middle;
          appearance: none;
          transition: background-position var(--form-switch-transition) ease-in-out;
          background-image: var(--form-switch-background-image);
          background-color: var(--form-switch-background-color);
          background-repeat: no-repeat;
          background-position: left center;
          background-size: contain;          
        }
        & input:checked {
          background-position: right center;
        }
      }

      label.select {
        cursor: inherit;

        & select {
          width: 100%;
          cursor: pointer;
          font-family: inherit;
          appearance: none;
          padding: var(--form-select-padding-y) var(--form-select-padding-x);
          background-image: var(--form-select-background-image);
          background-color: var(--form-select-background-color);
          background-repeat: no-repeat;
          background-position: right center;
          background-size: contain;          
          border: 1px solid var(--form-select-border-color);
          border-radius: var(--form-select-border-radius);

          &:focus {
            outline: 0;
          }
        }
      }
    `;
  }

  render() {
    return html`
      <label class=${this.classes.join(' ') || nothing}>
        <input 
          name=${this.name || nothing} 
          value=${this.value || nothing} 
          ?disabled=${this.disabled} 
          ?required=${this.required}
          ?autocomplete=${this.autocomplete}
          @input=${this.onInput}>
        <slot></slot>
      </label>
    `;
  }

  // Return classes for the form control
  get classes() {
    const classes = [];
    if (this.disabled) {
      classes.push('disabled');
    }
    if (this.required) {
      classes.push('required');
    }
    return classes;
  }

  // Form control properties
  get form() { return this.internals ? this.internals.form : null; }

  // Form control properties
  get type() { return this.localName; }

  // Form control properties
  get validity() { return this.internals ? this.internals.validity : null; }

  // Form control properties
  get validationMessage() { return this.internals ? this.internals.validationMessage : null; }

  // Form control properties
  get willValidate() { return this.internals ? this.internals.willValidate : null; }

  // Event hander for input event to update the value
  onInput(event) {
    if (!this.disabled) {
      this.value = event.target.value;
      this.internals.setFormValue(this.value);
      return true;
    }
    return false;
  }
}