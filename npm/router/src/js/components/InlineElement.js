import { LitElement, html, css, nothing } from 'lit';

/**
 * @class InlineElement
 *
 * This class is used for content boxes which should be inline (ie,
 * one after another) and verically aligned.
 *
 * @property {Boolean} flex - Flex the element to fill the available space, default false
 *
 * @example
 * <c-inline>Inline content</c-inline>
 */
export class InlineElement extends LitElement {
  static get localName() {
    return 'c-inline';
  }

  static get properties() {
    return {
    };
  }

  static get styles() {
    return css`
        :host {
            display: inline-block;
        }
    `;
  }

  render() {
    return html`
        <div class=${this.className || nothing}><slot></slot></div>
    `;
  }

  get className() {
    const classes = [];
    return classes.join(' ');
  }
}