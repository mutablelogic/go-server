/*
    js-components
    Web Components Design System
    github.com/mutablelogic/js-components
*/
import Event from '../event';

/**
 * Mixin for storing string data.
 * @mixin
*/
const ModelStore = (Base) => class extends Base {
  constructor(model) {
    super();
    this.$model = model;
    this.$map = new Map();
  }

  /**
   * Set the object in the store. Should trigger an add, change or delete event.
   *
   * @param {Object} data
   */
  set object(data) {
    this.$set(data);
  }

  /**
   * Set the array of objects in the store. Should trigger multiple add, change or delete events.
   *
   * @param {array} data
   */
  set array(data) {
    // Add markers
    const keys = new Map();
    this.$map.forEach((_, key) => {
      keys[key] = true;
    });
    // Fire add and change events
    data.forEach((obj) => {
      keys.delete(this.$set(obj));
    });
    // Fire delete events
    keys.forEach((_, key) => {
      const model = this.$map.get(key);
      this.dispatchEvent(new CustomEvent(Event.DELETE, {
        detail: model,
      }));
      this.$map.delete(key);
    });
  }

  /**
   * Private method. Set the object in the store and
   * return the key. Fires an add or change event.
   *
   * @param {Object} data
   */
  $set(data) {
    const model = new this.$model(data);
    const { key } = model;
    if (this.$map.has(key)) {
      this.dispatchEvent(new CustomEvent(Event.CHANGE, {
        detail: model,
      }));
    } else {
      this.dispatchEvent(new CustomEvent(Event.ADD, {
        detail: model,
      }));
    }
    this.$map.set(key, model);
    return key;
  }
};

export default ModelStore;
