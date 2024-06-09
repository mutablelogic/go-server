import { assertInstanceOf, assertNilOrInstanceOf } from "./assert.js";
import { Provider } from "./Provider.js";
import { Event } from "./event.js";

/**
 * @class
 * @name ModelArray 
 * @description Represents an array of objects, which can be bound to a provider
 * @param {Object} instancetype - The instance type of the object
 * @param {Provider} provider - The provider to bind to
 */
export class ModelArray extends EventTarget {
  #instancetype;
  #data;
  #mark;
  #equals;

  static get localName() {
    return 'c-model-array';
  }

  constructor(instancetype, provider) {
    super();
    assertInstanceOf(instancetype, Object);
    assertNilOrInstanceOf(provider, Provider);

    // Set defaults
    this.#instancetype = instancetype;
    this.#data = new Array();
    this.#mark = new Array();

    // Set the equals function to compare objects
    if (this.#instancetype.prototype.equals) {
      this.#equals = (a, b) => a.equals(b);
    } else {
      this.#equals = (a, b) => a.valueOf() === b.valueOf();
    }

    // Add listeners to the provider
    if (provider) {
      provider.addEventListener(Event.START, (e) => this.#onStart(e));
      provider.addEventListener(Event.OBJECT, (e) => this.#onObject(e));
      provider.addEventListener(Event.END, (e) => this.#onEnd(e));
    }
  }

  get length() {
    return this.#data.length;
  }

  indexOf(item) {
    assertInstanceOf(item, this.#instancetype);
    for (let i = 0; i < this.#data.length; i++) {
      if (this.#data[i].valueOf() === item.valueOf()) {
        return i;
      }
    }
    return -1;
  }

  #onStart(event) {
    /* set all marks to true */
    this.#mark.fill(true);
  }

  #onEnd(event) {
    /* check any existing marks, mark for deletion */
    this.#mark.forEach((mark, index) => {
      if (mark) {
        console.log('ModelArray#onEnd shouldBeDeleted', this.#data[index]);
        this.#data.splice(index, 1);
        this.#mark.splice(index, 1);
      }
    });
  }

  #onObject(event) {
    let obj = new this.#instancetype(event.detail);
    // Check for existing object
    const index = this.indexOf(obj);
    if (index >= 0) {
      // Object already exists
      if (this.#equals(this.#data[index], obj)) {
        this.#mark[index] = false;
      } else {
        console.log('ModelArray#object changed', obj);
        this.#data[index] = obj;
        this.#mark[index] = false;
      }
    } else {
      // Object is new, add it
      console.log('ModelArray#object new', obj);
      this.#data.push(obj);
      this.#mark.push(false);
    }
  }
}
