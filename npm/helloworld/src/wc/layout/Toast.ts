import { LitElement, html, css, nothing } from 'lit';
import { customElement, property } from 'lit/decorators.js';

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
  @property({ type: String }) color?: string;

  render() {
    return html`
      <div class=${this.className || nothing}>
        <slot></slot>
      </div>
    `;
  }

  static get styles() {
    return css`
      :host {
        position: fixed;
        z-index: 1000;
        right: 0;
        bottom: 0;

        margin: var(--toast-margin);
        background-color: var(--toast-background-color);
        text-color: var(--toast-text-color);
      }

      :host div {
        transition: visibility 0.2s, opacity 0.2s ease-in-out;

        &.visible {
          display: block;
          visibility: visible;
          opacity: 1;
        }

        &:not(.visible) {
          display: none;
          visibility: hidden;
          opacity: 0;
        }
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
