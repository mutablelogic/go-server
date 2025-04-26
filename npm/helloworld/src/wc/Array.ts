import { html, LitElement, PropertyValues } from 'lit';
import { customElement, property } from 'lit/decorators.js';
import { Database } from '../model/Database';

@customElement('wc-array')
export class Array extends LitElement {
  private _count: number | null = null;
  private _body?: any[] = [];

  @property({ type: String }) provider?: string;

  @property({ type: Number }) get count() {
    return this._count || 0;
  }

  @property({ type: Number }) get length() {
    return this._body?.length
  }

  connectedCallback(): void {
    super.connectedCallback();
    const provider = document.querySelector(this.provider);
    if (provider) {
      provider.addEventListener('object', this._handleObject.bind(this));
    }
  }

  disconnectedCallback(): void {
    const provider = document.querySelector(this.provider);
    if (provider) {
      provider.removeEventListener('object', this._handleObject.bind(this));
    }
    super.disconnectedCallback();
  }

  render() {
    return html`<span>ARRAY count=${this.count} length=${this.length}</span>`;
  }

  private _handleObject(event: CustomEvent) {
    const data = event.detail;
    this._count = data.count;
    this._body = data.body.map((data: any) => {
      return new Database(data);
    });
    this.requestUpdate();
  }
}
