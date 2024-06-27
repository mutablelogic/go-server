import { html, nothing } from 'lit';
import { FormControlElement } from './FormControlElement';

/**
 * @class FormSwitchElement
 *
 * This class is used to create a binary switch
 *
 * @property {Boolean} selected - Whether the switch is checked
 *
 * @example
 * <c-form-switch name="power" selected>Power</c-form-switch>
 */
export class FormSwitchElement extends FormControlElement {
  static get localName() {
    return 'c-form-switch';
  }

  constructor() {
    super();

    // Default properties
    this.selected = false;
  }

  static get properties() {
    return {
      selected: { type: Boolean },
    };
  }

  render() {
    return html`
      <label class=${this.classes.join(' ') || nothing}>
        <nobr>
        <input type="checkbox" role="switch" 
          name=${this.name || nothing} 
          ?disabled=${this.disabled} 
          ?checked=${this.selected}
          @input=${this.onInput}>
        <slot></slot>
        </nobr>
      </label>
    `;
  }

  // Return classes for the switch control
  get classes() {
    const classes = super.classes;
    classes.push('switch');
    return classes;
  }

  // Change the selected state when the input is changed
  onInput(event) {
    if (super.onInput(event)) {
      this.selected = event.target.checked;
      this.dispatchEvent(new CustomEvent('change', {
        bubbles: true,
        composed: true,
        detail: this.name || this.textContent.trim(),
      }));
    }
  }
}