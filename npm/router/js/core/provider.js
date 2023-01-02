/*
    js-components
    Web Components Design System
    github.com/mutablelogic/js-components
*/
import Model from './model';
import Event from './event';
import StringStore from './mixin/stringstore';
import ModelStore from './mixin/modelstore';

/**
 * Provider of data. After creation, add provider to the controller
 * so that it is made available to any web components.
 * @class
 */
export default class Provider extends ModelStore(StringStore(EventTarget)) {
  constructor(model, origin) {
    super(model);

    // Set class properties
    this.$timer = null;

    // eslint-disable-next-line no-prototype-builtins
    if (model && !Model.prototype.isPrototypeOf(model.prototype)) {
      throw new Error('Provider: Invalid model');
    }

    this.$origin = origin || '/';
  }

  /**
   * Make a request on a periodic basis.
   */
  request(path, req, interval) {
    this.cancel();
    if (!this.$timer) {
      this.$fetch(path, req);
    }
    if (interval) {
      this.$timer = setInterval(this.$fetch.bind(this, path, req), interval);
    }
  }

  /**
  * Cancel any existing request interval timer.
  */
  cancel() {
    if (this.$timer) {
      clearTimeout(this.$timer);
      this.$timer = null;
    }
  }

  /**
   * Private method to fetch data
   */
  $fetch(path, req) {
    let status;
    const absurl = (this.$origin + (path || '').removePrefix('/')) || '/';
    this.dispatchEvent(new CustomEvent(Event.START, {
      detail: absurl,
    }));
    fetch(absurl, req).then((response) => {
      status = response;
      const contentType = response.headers ? response.headers.get('Content-Type') || '' : '';
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
    }).then((data) => {
      if (!status.ok) {
        throw new Error(`${absurl} ${status.statusText} ${status.status}`);
      } else if (data instanceof Array) {
        this.array = data;
      } else if (data instanceof Object) {
        this.object = data;
      } else if (data instanceof String || typeof data === 'string') {
        this.string = data;
      } else {
        throw Error(`Returned data is of type ${typeof (data)}`);
      }
    }).catch((error) => {
      this.dispatchEvent(new CustomEvent(Event.ERROR, {
        detail: error,
      }));
    })
      .finally(() => {
        this.dispatchEvent(new CustomEvent(Event.END, {
          detail: absurl,
        }));
      });
  }
}
