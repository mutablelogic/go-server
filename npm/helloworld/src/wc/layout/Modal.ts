import { LitElement, html, css, nothing } from 'lit';
import { customElement, property } from 'lit/decorators.js';

/**
 * @class Modal
 *
 * This class is used to display a modal window overlaying the screen,
 * which can be dismissed by the user.
 *
 * @example
 * <wc-modal>
 *  <wc-nav>....</wc-nav>
 *  <wc-content>....</wc-content>
 *  <wc-content>....</wc-content>
 * </wc-modal>
 */
@customElement('wc-modal')
export class Modal extends LitElement {
  @property({ type: Boolean }) visible: boolean = false;

  render() {
    return html`
      <div class=${this.#canvasClassName || nothing}></div>
      <div class=${this.#contentClassName || nothing}><slot></slot></div>
    `;
  }

  static get styles() {
    return css`
      div {
        position: fixed; 
        left: 0; 
        top: 0;
        right: 0;
        bottom: 0;
        overflow-x: hidden;
        overflow-y: auto;
        transition: var(--modal-transition);

        &:not(.visible) {
          visibility: hidden;
          opacity: 0;
        }

        &.visible {
          visibility: visible;
          opacity: 1;
          &.canvas {
            opacity: var(--modal-opacity-canvas);
          }
        }
      }

      .canvas {
        background-color: var(--modal-background-color-canvas);
      }

      .content {
        background-color: var(--modal-background-color);
        margin: var(--modal-margin);
        border: var(--modal-border);
        border-radius: var(--modal-border-radius);
      }
    `;
  }

  get #canvasClassName() {
    const classes = [];
    classes.push('canvas');
    if (this.visible) {
      classes.push('visible');
    }
    return classes.join(' ');
  }

  get #contentClassName() {
    const classes = [];
    classes.push('content');
    if (this.visible) {
      classes.push('visible');
    }
    return classes.join(' ');
  }
}
