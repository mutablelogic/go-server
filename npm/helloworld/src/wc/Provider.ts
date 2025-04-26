import { html, LitElement } from 'lit';
import { customElement, property } from 'lit/decorators.js';

@customElement('wc-provider')
export class Provider extends LitElement {
  private _timer: number | null = null;
  private _debug: string = '';

  @property({ type: String }) set debug(value: string) {
    this._debug = value;
    this.requestUpdate();
  } get debug() {
    return this._debug;
  }

  @property({ type: String, reflect: true }) origin?: string;

  @property({ type: String, reflect: true }) path?: string;

  @property({ type: Number, reflect: true }) interval?: number;

  render() {
    return html`<span>[${this.debug}]</span>`;
  }

  /** On first update, set the origin and fetch if the path is not empty */
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
      this._timer = setInterval(() => {
        this.dofetch(url, request);
      }, interval * 1000);
    }
  }

  /**
   * Cancel any existing request interval timer.
   */
  cancel() {
    if (this._timer) {
      clearTimeout(this._timer);
      this._timer = null;
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
        throw new Error(`${response.status}`);
      }
      const contentType = response.headers ? response.headers.get('Content-Type') || '' : '';
      return this.fetchresponse(contentType.split(';')[0], response);
    }).then((data) => {
      this.fetchdata(data);
    }).catch((error) => {
      console.log('Error:', error);
      this.debug = `${error}`;
    });
  }

  private fetchresponse(contentType: string, response: Response) {
    switch (contentType.split(';')[0]) {
      case 'application/json':
      case 'text/json':
        return response.json();
      case 'text/plain':
      case 'text/html':
        return response.text();
      default:
        return response.blob();
    }
  }

  private fetchdata(data: any) {
    // Handle the fetched data
    if (typeof data === 'string') {
      this.fetchtext(data);
    } else if (Array.isArray(data)) {
      data.forEach((item) => {
        this.fetchobject(item);
      });
    } else if (data instanceof Object) {
      this.fetchobject(data);
    } else {
      this.fetchblob(data);
    }
  }

  private fetchtext(data: string) {
    this.debug = `Text: ${data}`;
    console.log('Text:', data);
  }

  private fetchobject(data: any) {
    this.debug = `Object: ${JSON.stringify(data)}`;
    console.log('Object:', data);
  }

  private fetchblob(data: Blob) {
    this.debug = `Blob: ${data}`;
    console.log('Blob:', data);
  }
}
