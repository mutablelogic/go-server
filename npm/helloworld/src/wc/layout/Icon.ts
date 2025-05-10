import { LitElement, svg, css, nothing } from 'lit';
import { customElement, property } from 'lit/decorators.js';

declare module '*.svg' {
    const content: string;
    export default content;
}

import icons from 'bootstrap-icons/bootstrap-icons.svg';

/**
 * @class Icon
 *
 * This class is a container for a bootstrap icon.
 *
 * @example
 * <wc-icon size="large">arrow</wc-icon>
 */
@customElement('wc-icon')
export class Icon extends LitElement {
  @property({ type: String }) size: string = 'default';

  render() {
    return svg`
        <div class=${this.className || nothing}>
            <svg class=${this.className || nothing}><use href="${icons}#${this.name}"/></svg>
        </div>
    `;
  }

  static get styles() {
    return css`
      :host {
          display: inline-block;
          vertical-align: middle;
      }
      .size-small {
          width: var(--icon-size-small);
          height: var(--icon-size-small);
      }
      .size-medium, .size-default {
          width: var(--icon-size-medium);
          height: var(--icon-size-medium);
      }
      .size-large {
          width: var(--icon-size-large);
          height: var(--icon-size-large);
      }
      .size-xlarge {
          width: var(--icon-size-xlarge);
          height: var(--icon-size-xlarge);
      }
      svg {
        width: 100%;
        height: 100%;
        fill: currentColor;
      }
    `;
  }

  get name() {
    return this.textContent.trim() || 'bootstrap-reboot';
  }

  get className() {
    const classes = [];
    classes.push(`size-${this.size}`);
    return classes.join(' ');
  }
}
