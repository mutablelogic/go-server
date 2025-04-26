/**
 * Options for setting a cookie.
 */
export interface CookieOptions {
  /** Expiration date as a Date object, number of days, or a UTC string. */
  expires?: Date;
  /** The URL path for which the cookie is valid. Defaults to '/'. */
  path?: string;
  /** The domain for which the cookie is valid. */
  domain?: string;
  /** Controls whether a cookie is sent with cross-site requests. */
  sameSite?: 'Strict' | 'Lax' | 'None';
}

export class Cookies {
  /**
   * Sets a cookie with a name, value, and optional settings.
   * @param name The name of the cookie.
   * @param value The value of the cookie.
   * @param options Optional settings like expires, path, domain, secure, sameSite.
   */
  set(name: string, value: string, options: CookieOptions = {}): void {
    let cookieString = encodeURIComponent(name) + "=" + encodeURIComponent(value);

    // Expiration
    if (options.expires instanceof Date) {
      cookieString += "; expires=" + options.expires.toUTCString();
    }

    // Path, defaulting to '/'
    cookieString += "; path=" + (options.path || '/');

    // Domain
    if (options.domain) {
      cookieString += "; domain=" + options.domain;
    }

    // Handle SameSite, ensuring Secure is set if SameSite=None
    if (options.sameSite) {
      cookieString += "; samesite=" + options.sameSite;
      if (options.sameSite === 'None') {
        // Secure attribute is required for SameSite=None
        cookieString += "; secure";
      }
    }

    // Set the cookie
    document.cookie = cookieString;
  }

  /**
   * Retrieves the value of a cookie by its name.
   * @param name The name of the cookie.
   * @returns The cookie value string, or null if the cookie is not found or document is unavailable.
   */
  get(name: string): string | null {
    const nameEQ = encodeURIComponent(name) + "=";
    const ca = document.cookie.split(';');
    for (let i = 0; i < ca.length; i++) {
      let c = ca[i];
      while (c.charAt(0) === ' ') {
        c = c.substring(1, c.length);
      }
      if (c.indexOf(nameEQ) === 0) {
        return decodeURIComponent(c.substring(nameEQ.length, c.length));
      }
    }
    return null;
  }
}
