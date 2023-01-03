/*
    js-components
    Web Components Design System
    github.com/mutablelogic/js-components
*/
import HostController from './hostcontroller';
import Provider from './provider';

/**
 * Controller acts as business logic between components
 * and providers.
 * @class
*/
class Controller {
  constructor() {
    this.$protos = new Map();
    this.$providers = new Map();
    this.$renderers = new Map();
  }

  /**
    * Define a data provider for the controller. The object is then
    * accessible as a property of the controller.
    *
    * @param {string} key - The name of the property
    * @param {Provider} object - The object
    */
  define(key, object) {
    if (!key || this.$providers.has(key)) {
      throw Error(`Controller: Duplicate or invalid key ${key}`);
    }

    // Set object
    if (object instanceof Provider) {
      this.$providers.set(key, object);
    } else {
      throw new Error(`Controller: Invalid object added to controller with key ${key}`);
    }

    // Set property
    Object.defineProperty(this, key, {
      value: object,
      writable: false,
    });

    // Return object
    return object;
  }

  /**
    * Bind a provider to a host (a web component) so that the host
    * can be updated by the controller.
    *
    * @param {string} key - The name of the provider
    * @param {LitElement} host - The web component
    * @param {function} renderer - The function used to render content before connecting to host
    */
  bind(key, host, renderer) {
    if (!this.$providers.has(key)) {
      throw Error(`Controller: Invalid key ${key}`);
    }
    if (!(host instanceof HTMLElement)) {
      throw Error(`Controller: host is not a HTMLElement for ${key.quote()}`);
    }
    if (typeof renderer !== 'function') {
      throw Error(`Controller: missing renderer for provider ${key.quote()}`);
    }
    return new HostController(host, this.$providers.get(key), renderer);
  }

  /**
    * Add properties to a newly constructed object.
    *
    * @param {Model} proto - The model to add the prototype to
    * @param {Class} model - The class
    */
  setPrototypeOf(proto, model) {
    let superproto = this.$protos.get(model);
    if (!superproto) {
      superproto = {};
      this.$protos.set(model, superproto);

      // $name property
      Object.defineProperty(superproto, '$name', {
        value: model.name,
        writable: false,
        enumerable: false,
      });

      // $properties property
      const properties = model.properties || {};
      Object.defineProperty(superproto, '$properties', {
        value: properties,
        writable: false,
        enumerable: false,
      });

      // Define each property with a getter and a setter
      Object.keys(properties).forEach((key) => {
        Object.defineProperty(superproto, key, {
          // eslint-disable-next-line object-shorthand, func-names
          get: function () { return this.get(key); },
          // eslint-disable-next-line object-shorthand, func-names
          set: function (value) { return this.set(key, value); },
          enumerable: true,
        });
      });
    }
    Object.setPrototypeOf(Object.getPrototypeOf(proto), superproto);
  }
}

// Export global controller
export default (new Controller());
