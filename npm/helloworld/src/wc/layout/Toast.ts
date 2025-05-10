import { LitElement, html, css, nothing } from 'lit';
import { customElement, property } from 'lit/decorators.js';

interface ShowOptions {
  /** The duration the toast is displayed for, in milliseconds. */
  duration?: number;

  /** The color of the button. */
  color?: string;
}

/**
 * @class Toast
 *
 * This class is a container for a toast message, which pops up from the bottom of the screen
 * and can be dismissed by the user, or automatically after a certain time.
 *
 * @example
 * <wc-toast visible>...</wc-toast>
 */
@customElement('wc-toast')
export class Toast extends LitElement {
  @property({ type: Boolean }) visible: boolean = false;
  @property({ type: Number }) duration: number = 3000;
  @property({ type: String }) color?: string;
  #timer: number | null = null;

  render() {
    return html`
      <div class=${this.className || nothing}>
        <slot></slot>
      </div>
    `;
  }

  show(message: string, opts = {} as ShowOptions) {
    this.visible = true;
    this.textContent = message;
    this.color = opts.color || 'primary';

    // Cancel any existing timer
    if (this.#timer) {
      clearTimeout(this.#timer);
    }
    // Set a new timer
    this.#timer = setTimeout(() => {
      this.visible = false;
      this.#timer = null;
    }, opts.duration || this.duration);
  }

  static get styles() {
    return css`
      :host {
        position: fixed;
        z-index: 1000;
        right: 0;
        bottom: 0;
      }

      :host div {
        display: inline-flex;
        margin: var(--toast-margin);
        padding: var(--toast-padding);
        border: var(--toast-border);
        border-radius: var(--toast-border-radius);
        transition: visibility 0.2s, opacity 0.2s ease-in-out;
        cursor: pointer;
        user-select: none;

        &.visible {
          visibility: visible;
          opacity: 1;
        }

        &:not(.visible) {
          visibility: hidden;
          opacity: 0;
        }
      }

      div.color-primary {
        background-color: var(--primary-color);
        color: var(--primary-opp-color);
      }
      div.color-secondary {
        background-color: var(--secondary-color);
        color: var(--secondary-opp-color);
      }
      div.color-light {
        background-color: var(--light-color);
        color: var(--light-opp-color);
      }
      div.color-dark {
        background-color: var(--dark-color);
        color: var(--dark-opp-color);
      }
      div.color-white {
        background-color: var(--white-color);
        color: var(--white-opp-color);
      }
      div.color-black {
        background-color: var(--black-color);
        color: var(--black-opp-color);
      }
      div.color-success {
        background-color: var(--success-color);
        color: var(--success-opp-color);
      }
      div.color-warning {
        background-color: var(--warning-color);
        color: var(--warning-opp-color);
      }
      div.color-error {
        background-color: var(--error-color);
        color: var(--error-opp-color);
      }      
    `;
  }

  get className() {
    const classes = [];
    if (this.visible) {
      classes.push('visible');
    }
    if (this.color) {
      classes.push(`color-${this.color}`);
    }
    return classes.join(' ');
  }
}
