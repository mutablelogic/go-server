import { assertTypeOf, assertNilOrInstanceOf, assertNilOrTypeOf } from './assert.js';
import { Event } from './event.js';

/**
 * @class
 * @name Provider 
 * @description Fetch data from a remote endpoint
 */
export class Provider extends EventTarget {
    // Private properties
    #origin;
    #timer;
    #interval;

    constructor(origin) {
        super();
        assertNilOrTypeOf(origin, 'string');

        // The base URL is the current URL
        const base = window.location.href;
        this.#origin = origin ? new URL(origin, base) : new URL('/', base);

        // Set the timer and interval to undefined
        this.#timer = 0;
        this.#interval = 0;
    }

    /**
     * Get the origin of the provider
     * @returns {String}
     * @readonly
     * @memberof Provider
     */
    get Origin() {
        return this.#origin;
    }

    /**
     * Return a default request object, which is a GET request
     * @returns {Object}
     * @readonly
     * @memberof Provider
     */
    static get EmptyRequest() {
        return {
            method: 'GET',
            mode: 'cors',
            headers: new Headers(),
            redirect: 'follow',
            body: null,
            referrerPolicy: 'no-referrer',
        };
    }

    /**
    * Fetch data from a remote endpoint
    * @param {string} path - The path to the resource, relative to the origin
    * @param {Request} request - The request object, which is optional. If not provided, a default request is used.
    * @param {number} interval - The interval to fetch the data in milliseconds, which is optional. Requests will continue
    *   until the interval is cleared or a new request is made.
    * @memberof Provider
    */
    Fetch(path, request, interval) {
        assertTypeOf(path, 'string');
        assertNilOrInstanceOf(request, Request);
        assertNilOrTypeOf(interval, 'number');

        // Create a default request if not provided
        if (!request) {
            request = Provider.EmptyRequest;
        }

        // Make path relative
        if (path.startsWith('/')) {
            path = path.substring(1);
        }

        // Create an absolute URL
        let url = new URL(path, this.#origin);

        // Cancel any existing requests
        this.Cancel();

        // Fetch the data
        this.#fetch(url, request);

        // Set the interval for the next fetch
        if (interval) {
            this.#interval = interval;
            this.#timer = setTimeout(() => {
                this.#fetch(url, request);
            }, interval);
        } else {
            this.#timer = null;
        }
    }

    /**
     * Cancel any existing request interval timer.
     */
    Cancel() {
        if (this.#timer) {
            clearTimeout(this.#timer);
            this.#timer = null;
        }
    }

    /**
     * Perform a fetch request
     */
    #fetch(url, request) {
        this.dispatchEvent(new CustomEvent(Event.START, {
            detail: request,
        }));
        fetch(url, request).then((response) => {
            if (!response.ok) {
                throw new Error(`status: ${response.status}`);
            }
            const contentType = response.headers ? response.headers.get('Content-Type') || '' : '';
            switch (contentType.split(';')[0]) {
                case 'application/json':
                    return response.json();
                case 'text/plain':
                case 'text/html':
                    return response.text();
                default:
                    return response.blob();
            }
        }).then((data) => {
            if (typeof data == "string") {
                this.#string(data);
            } else if (data instanceof Array) {
                this.#array(data);
            } else if (data instanceof Object) {
                this.#object(data);
            } else {
                this.#blob(data);
            }
        }).catch((error) => {
            this.dispatchEvent(new ErrorEvent(Event.ERROR, {
                error: error,
                message: `${error}`
            }));
        }).finally(() => {
            this.dispatchEvent(new CustomEvent(Event.END, {
                detail: request,
            }));
            if (this.#timer && this.#interval) {
                this.Cancel();
                this.#timer = setTimeout(() => {
                    this.#fetch(url, request);
                }, this.#interval);
            }
        });
    }

    /**
     * Private method to process array of objects
     */
    #array(data) {
        data.forEach((item) => {
            this.#object(item);
        });
    }

    /**
     * Private method to process objects
     */
    #object(data) {
        console.log("Object: ", data);
    }

    /**
     * Private method to process string data
     */
    #string(data) {
        console.log("String: ", data);
    }

    /**
     * Private method to process blob data
     */
    #blob(data) {
        console.log("Blob: ", data);
    }
}
