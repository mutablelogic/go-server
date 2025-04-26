import { html, LitElement } from 'lit';
import { customElement, property } from 'lit/decorators.js';

@customElement('wc-provider')
export class Provider extends LitElement {
  private timer: number | null = null;

  @property({ type: String, attribute: false })
  debug: string = '';

  @property({ type: String, reflect: true })
  origin?: string;

  @property({ type: String, reflect: true })
  path?: string;

  @property({ type: Number, reflect: true })
  interval?: number;

  render() {
    return html`<span>[${this.debug}]</span>`;
  }

  /** On first update, fetch if path is not NULL */
  firstUpdated() {
    if (!this.origin) {
      this.origin = window.location.origin;
    }
    if (this.path) {
      this.fetch();
    }
  }

  /**
   * Fetch data from a remote source
   */
  fetch() {
    const path = this.path || '/';
    const request = {};
    const interval = this.interval;

    // Create an absolute URL
    let url: URL;
    try {
      url = new URL(path, this.origin);
    } catch (error) {
      this.debug = `${error}`;
      return;
    }

    // Cancel any existing requests
    this.cancel();

    // Fetch the data
    this.dofetch(url, request);

    // Set the interval for the next fetch
    if (interval) {
      this.timer = setInterval(() => {
        this.dofetch(url, request);
      }, interval * 1000);
    }
  }


  /**
   * Cancel any existing request interval timer.
   */
  cancel() {
    if (this.timer) {
      clearTimeout(this.timer);
      this.timer = null;
    }
  }

  /**
   * Perform the fetch operation.
   * @param url The URL to fetch.
   * @param request The request options.
   */
  private dofetch(url: URL, request: any) {
    this.debug = `Fetching ${url.toString()}`;

    // Perform the fetch operation
    fetch(url.toString(), {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
      ...request,
    }).then((response) => {
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      return response.json();
    }).then((data) => {
      console.log('Data fetched:', data);
      this.debug = JSON.stringify(data);
    }).catch((error) => {
      console.log('Error:', error);
      this.debug = `${error}`;
    });
  }
}
