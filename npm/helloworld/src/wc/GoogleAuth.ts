import { html, LitElement, PropertyValues } from 'lit';
import { customElement, property } from 'lit/decorators.js';
import { Cookies } from "./Cookies";

interface GoogleAuthOptions {
  /** The client ID for your Google application. */
  ClientId: string;

  /** The theme of the button. */
  Theme?: string;

  /** The size of the button. */
  Size?: string;
}

@customElement('wc-google-auth')
export class GoogleAuth extends LitElement {
  private cookies: Cookies = new Cookies();

  @property({ type: Object })
  opts: GoogleAuthOptions = { ClientId: '' };

  protected firstUpdated(_changedProperties: PropertyValues): void {
    super.firstUpdated(_changedProperties);
    this.load(() => {
      // Load the Google Identity Services library
      window.google?.accounts.id.initialize({
        client_id: this.opts?.ClientId,
        callback: this.credentialResponse.bind(this), // Bind the callback to the class instance
        use_fedcm_for_prompt: true,
        use_fedcm_for_button: true,
        button_auto_select: true
      });

      // Render the button
      const buttonContainer = this.shadowRoot?.getElementById('google-button-container');
      window.google?.accounts.id.renderButton(
        buttonContainer,
        {
          theme: this.opts?.Theme || 'outline', // Use theme from opts or default
          size: this.opts?.Size || 'large'      // Use size from opts or default
        }
      );
    });
  }

  render() {
    return html`<div id="google-button-container"></div>`;
  }

  /**
   * Return the authentication token.
   */
  get token() {
    return this.cookies.get('jwt');
  }

  private credentialResponse(response: any) {
    // Store JWT token in cookies
    this.cookies.set('jwt', response.credential, {
      sameSite: 'Strict',
    })
  }

  /**
   * Load the Google Identity Services library asynchronously, then
   * initialize it with the provided callback.
   */
  private load(callback: () => void) {
    const scriptTag = document.createElement('script');
    scriptTag.src = 'https://accounts.google.com/gsi/client';
    scriptTag.async = true;
    scriptTag.defer = true;
    scriptTag.onload = () => {
      // Initialize the Google Identity Services library
      callback && callback();
    };
    document.body.appendChild(scriptTag);
  }
}

// Define the type for the Google Identity Services library
declare global {
  interface Window {
    google: any;
  }
}
