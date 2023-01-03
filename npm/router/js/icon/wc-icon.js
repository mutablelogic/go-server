import { LitElement, svg, css } from 'lit';

// Icons
import icons from 'bootstrap-icons/bootstrap-icons.svg';

/**
 * An icon element
 *
 */
window.customElements.define('wc-icon', class extends LitElement {
  static get styles() {
    return css`
      :host {        
        display: inline-block;
        vertical-align: middle;
      }
      `;
  }

  static get properties() {
    return {
      /**
       * Name of the icon to display
       * @type {string}
       */
      name: { type: String },

      /**
       * Size of the icon to display
       * @type {string}
       */
      size: { type: String },
    };
  }

  constructor() {
    super();
    this.name = 'bootstrap-reboot';
    this.size = '1.5em';
  }

  render() {
    return svg`
        <svg style="width: ${this.size}; height: ${this.size};" fill="currentColor"><use href="${icons}#${this.name}"/></svg>
      `;
  }
});
