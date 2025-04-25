import {html, css, LitElement} from 'lit';
import {customElement, property} from 'lit/decorators.js';

@customElement('wc-app')
export class App extends LitElement {
  static styles = css`h1 { color: blue }`;

  @property()
  name = 'Somebody';

  render() {
    return html`<h1>Hello, ${this.name}!</h1>`;
  }
}
