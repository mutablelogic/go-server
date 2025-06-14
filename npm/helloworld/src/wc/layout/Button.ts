import { LitElement, html, css, nothing } from 'lit';
import { customElement, property } from 'lit/decorators.js';

/**
 * @class Button
 *
 * This class is a button element, which can be used to trigger actions or events.
 *
 * @example
 * <wc-button>...</wc-button>
 */
@customElement('wc-button')
export class Button extends LitElement {
  @property({ type: String }) size: string = 'default';
  @property({ type: Boolean }) disabled: boolean;
  @property({ type: String }) color: string = 'primary';

  render() {
    return html`
      <button class=${this.className || nothing} ?disabled=${this.disabled} @click=${this.#onClick}>
        <slot></slot>
      </div>
    `;
  }

  static get styles() {
    return css`
      button {
        align-items: center;
        display: inline-flex;
        margin: var(--button-margin);
        padding: var(--button-padding);
        border: var(--button-border);
        border-radius: var(--button-border-radius);
      }

      button:not(:disabled) {
        cursor: pointer;
        user-select: none;
      }

      button:disabled {
        cursor: default;
        opacity: var(--button-opacity-disabled);
      }

      button:not(:disabled):hover {
        transform: scale(1.05);
      }

      button:not(:disabled):active {
        font-weight: var(--button-font-weight-active);
      }

      button.color-primary {
        background-color: var(--primary-color);
        color: var(--primary-opp-color);
      }
      button.color-secondary {
        background-color: var(--secondary-color);
        color: var(--secondary-opp-color);
      }
      button.color-light {
        background-color: var(--light-color);
        color: var(--light-opp-color);
      }
      button.color-dark {
        background-color: var(--dark-color);
        color: var(--dark-opp-color);
      }
      button.color-white {
        background-color: var(--white-color);
        color: var(--white-opp-color);
      }
      button.color-black {
        background-color: var(--black-color);
        color: var(--black-opp-color);
      }
      button.color-success {
        background-color: var(--success-color);
        color: var(--success-opp-color);
      }
      button.color-warning {
        background-color: var(--warning-color);
        color: var(--warning-opp-color);
      }
      button.color-error {
        background-color: var(--error-color);
        color: var(--error-opp-color);
      }
    `;
  }

  #onClick(event: MouseEvent): void {
    event.preventDefault();
  }

  get className() {
    const classes = [];
    if (this.size) {
      classes.push(`size-${this.size}`);
    }
    if (this.color) {
      classes.push(`color-${this.color}`);
    }
    return classes.join(' ');
  }
}
