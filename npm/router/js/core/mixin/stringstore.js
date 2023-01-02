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
const StringStore = (Base) => class extends Base {
  constructor() {
    super();
    this.$data = undefined;
  }

  /**
   * Set the object in the string store. Should trigger an add, change or delete event.
   *
   * @param {string} data
   */
  set string(data) {
    if (data !== this.$data) {
      if (!data && this.$data) {
        this.dispatchEvent(new CustomEvent(Event.DELETE, {
          detail: this.$data,
        }));
      } else if (data && !this.$data) {
        this.dispatchEvent(new CustomEvent(Event.ADD, {
          detail: data,
        }));
      } else {
        this.dispatchEvent(new CustomEvent(Event.CHANGE, {
          detail: data,
        }));
      }
      this.$data = data;
    }
  }
};

export default StringStore;
