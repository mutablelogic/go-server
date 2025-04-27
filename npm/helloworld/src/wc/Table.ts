import { html, LitElement, PropertyValues } from 'lit';
import { customElement, property } from 'lit/decorators.js';

@customElement('wc-table')
export class Table extends LitElement {
  render() {
    return html`<span>TABLE</span>`;
  }
}
